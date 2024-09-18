package tracing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/vmihailenco/msgpack/v5"
)

type LogData struct {
	ch.CHModel `ch:"table:logs_data_buffer,insert:logs_data_buffer,alias:s"`

	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	ParentID  idgen.SpanID
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

func (sd *LogData) FilledLog() (*Span, error) {
	log := new(Span)
	if err := sd.Decode(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (sd *LogData) Decode(log *Span) error {
	if err := msgpack.Unmarshal(sd.Data, log); err != nil {
		return fmt.Errorf("msgpack.Unmarshal failed: %w", err)
	}

	log.ProjectID = sd.ProjectID
	log.TraceID = sd.TraceID
	log.ID = sd.ID
	log.ParentID = sd.ParentID
	log.Time = sd.Time

	log.Type = log.System
	if i := strings.IndexByte(log.Type, ':'); i >= 0 {
		log.Type = log.Type[:i]
	}

	return nil
}

func initLogData(data *LogData, log *Span) {
	data.Type = log.Type
	data.ProjectID = log.ProjectID
	data.TraceID = log.TraceID
	data.ID = log.ID
	data.ParentID = log.ParentID
	data.Time = log.Time
	data.Data = marshalLogData(log)

	fmt.Printf("LogData: %+v\n", data)
}

func SelectLog(
	ctx context.Context,
	app *bunapp.App,
	projectID uint32,
	traceID idgen.TraceID,
	spanID idgen.SpanID,
) (*Span, error) {
	var logs []*LogData

	selq := app.CH.NewSelect().
		ColumnExpr("project_id, trace_id, id, parent_id, time, data").
		Model(&logs).
		Where("project_id = ?", projectID).
		Where("trace_id = ?", traceID)

	if spanID != 0 {
		selq.Where("id = ? OR parent_id = ?", spanID, spanID)
	} else {
		selq.Where("id = 0").Limit(1)
	}

	if err := selq.Scan(ctx); err != nil {
		return nil, err
	}

	var found *Span

	for i := len(logs) - 1; i >= 0; i-- {
		sd := logs[i]
		if sd.ID == spanID {
			span, err := sd.FilledLog()
			if err != nil {
				return nil, err
			}
			found = span
			logs = append(logs[:i], logs[i+1:]...)
			break
		}
	}

	if found == nil {
		return nil, sql.ErrNoRows
	}

	for _, sd := range logs {
		span, err := sd.FilledLog()
		if err != nil {
			return nil, err
		}
		if span.IsEvent() {
			found.Events = append(found.Events, span.Event())
		}
	}

	return found, nil
}

func SelectTraceLogs(
	ctx context.Context, app *bunapp.App, traceID idgen.TraceID,
) ([]*Span, bool, error) {
	const limit = 10000

	var data []SpanData

	if err := app.CH.NewSelect().
		DistinctOn("id").
		ColumnExpr("project_id, trace_id, id, parent_id, time, data").
		Model(&data).
		Column("data").
		Where("trace_id = ?", traceID).
		OrderExpr("time ASC").
		Limit(limit).
		Scan(ctx); err != nil {
		return nil, false, err
	}

	logs := make([]*Span, len(data))

	for i := range logs {
		log := new(Span)
		logs[i] = log
		if err := data[i].Decode(log); err != nil {
			return nil, false, err
		}
	}

	return logs, len(logs) == limit, nil
}

func marshalLogData(log *Span) []byte {
	b, err := msgpack.Marshal(log)
	if err != nil {
		panic(err)
	}
	return b
}
