package pgq

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/uptrace/bun"
	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend"
	"github.com/vmihailenco/taskq/v4/backend/jobutil"
)

type Queue struct {
	opt *taskq.QueueConfig
	db  *bun.DB

	consumer *taskq.Consumer

	_closed uint32
}

type Job struct {
	bun.BaseModel `bun:"taskq_jobs,alias:m"`

	ID            ulid.ULID `bun:"type:uuid"`
	Queue         string
	RunAt         time.Time
	ReservedCount int
	ReservedAt    time.Time
	Data          []byte
}

var _ taskq.Queue = (*Queue)(nil)

func NewQueue(db *bun.DB, opt *taskq.QueueConfig) *Queue {
	if opt.WaitTimeout == 0 {
		opt.WaitTimeout = 3 * time.Second
	}
	opt.Init()

	q := &Queue{
		opt: opt,
		db:  db,
	}

	return q
}

func (q *Queue) Name() string {
	return q.opt.Name
}

func (q *Queue) String() string {
	return fmt.Sprintf("queue=%q", q.Name())
}

func (q *Queue) Options() *taskq.QueueConfig {
	return q.opt
}

func (q *Queue) Consumer() taskq.QueueConsumer {
	if q.consumer == nil {
		q.consumer = taskq.NewConsumer(q)
	}
	return q.consumer
}

func (q *Queue) Len(ctx context.Context) (int, error) {
	return q.db.NewSelect().
		Model((*Job)(nil)).
		Where("queue = ?", q.Name()).
		Count(ctx)
}

func (q *Queue) AddJob(ctx context.Context, job *taskq.Job) error {
	return q.add(ctx, job)
}

func (q *Queue) add(ctx context.Context, job *taskq.Job) error {
	if job.TaskName == "" {
		return backend.ErrTaskNameRequired
	}
	if job.Name != "" && q.isDuplicate(ctx, job) {
		job.Err = taskq.ErrDuplicate
		return nil
	}

	data, err := job.MarshalBinary()
	if err != nil {
		return err
	}

	model := &Job{
		Queue: q.Name(),
		Data:  data,
	}
	if job.ID != "" {
		model.ID = ulid.MustParse(job.ID)
	} else {
		model.ID = ulid.Make()
	}
	if job.Delay > 0 {
		model.RunAt = time.Now().Add(job.Delay)
	}

	if _, err := q.db.NewInsert().
		Model(model).
		Exec(ctx); err != nil {
		return err
	}

	job.ID = model.ID.String()
	return nil
}

func (q *Queue) isDuplicate(ctx context.Context, job *taskq.Job) bool {
	exists := q.opt.Storage.Exists(ctx, jobutil.FullJobName(q, job))
	return exists
}

func (q *Queue) ReserveN(
	ctx context.Context, n int, waitTimeout time.Duration,
) ([]taskq.Job, error) {
	if waitTimeout > 0 {
		if err := sleep(ctx, waitTimeout); err != nil {
			return nil, err
		}
	}

	subq := q.db.NewSelect().
		ColumnExpr("ctid").
		Model((*Job)(nil)).
		Where("queue = ?", q.Name()).
		Where("run_at <= ?", time.Now()).
		Where("reserved_at = ?", time.Time{}).
		OrderExpr("id ASC").
		Limit(n).
		For("UPDATE SKIP LOCKED")

	var src []Job

	if err := q.db.NewUpdate().
		Model(&src).
		With("todo", subq).
		TableExpr("todo").
		Set("reserved_count = reserved_count + 1").
		Set("reserved_at = now()").
		Where("m.ctid = todo.ctid").
		Returning("m.*").
		Scan(ctx); err != nil {
		return nil, err
	}

	jobs := make([]taskq.Job, len(src))

	for i := range jobs {
		job := &jobs[i]
		if err := unmarshalJob(job, &src[i]); err != nil {
			job.Err = err
		}
	}

	return jobs, nil
}

func (q *Queue) Release(ctx context.Context, job *taskq.Job) error {
	res, err := q.db.NewUpdate().
		Model((*Job)(nil)).
		Set("run_at = ?", time.Now().Add(job.Delay)).
		Set("reserved_at = ?", time.Time{}).
		Where("queue = ?", q.Name()).
		Where("id = ?", ulid.MustParse(job.ID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (q *Queue) Delete(ctx context.Context, job *taskq.Job) error {
	res, err := q.db.NewDelete().
		Model((*Job)(nil)).
		Where("queue = ?", q.Name()).
		Where("id = ?", ulid.MustParse(job.ID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (q *Queue) Purge(ctx context.Context) error {
	if _, err := q.db.NewDelete().
		Model((*Job)(nil)).
		Where("queue = ?", q.Name()).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

// Close is like CloseTimeout with 30 seconds timeout.
func (q *Queue) Close() error {
	return q.CloseTimeout(30 * time.Second)
}

// CloseTimeout closes the queue waiting for pending messages to be processed.
func (q *Queue) CloseTimeout(timeout time.Duration) error {
	if !atomic.CompareAndSwapUint32(&q._closed, 0, 1) {
		return nil
	}

	if q.consumer != nil {
		_ = q.consumer.StopTimeout(timeout)
	}

	return nil
}

func unmarshalJob(dest *taskq.Job, src *Job) error {
	if err := dest.UnmarshalBinary(src.Data); err != nil {
		return err
	}

	dest.ID = src.ID.String()
	dest.ReservedCount = src.ReservedCount

	return nil
}

func sleep(ctx context.Context, d time.Duration) error {
	done := ctx.Done()
	if done == nil {
		time.Sleep(d)
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-done:
		return ctx.Err()
	}
}
