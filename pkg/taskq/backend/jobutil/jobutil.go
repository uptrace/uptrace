package jobutil

import (
	"encoding/binary"
	"fmt"

	"github.com/dgryski/go-farm"

	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend/unsafeconv"
)

func WrapJob(msg *taskq.Job) *taskq.Job {
	msg0 := taskq.NewJob(msg)
	msg0.Name = msg.Name
	return msg0
}

func UnwrapJob(msg *taskq.Job) (*taskq.Job, error) {
	if len(msg.Args) != 1 {
		err := fmt.Errorf("UnwrapJob: got %d args, wanted 1", len(msg.Args))
		return nil, err
	}

	msg, ok := msg.Args[0].(*taskq.Job)
	if !ok {
		err := fmt.Errorf("UnwrapJob: got %v, wanted *taskq.Job", msg.Args)
		return nil, err
	}
	return msg, nil
}

func UnwrapJobHandler(fn interface{}) taskq.HandlerFunc {
	if fn == nil {
		return nil
	}
	h := fn.(func(*taskq.Job) error)
	return taskq.HandlerFunc(func(msg *taskq.Job) error {
		msg, err := UnwrapJob(msg)
		if err != nil {
			return err
		}
		return h(msg)
	})
}

func FullJobName(q taskq.Queue, msg *taskq.Job) string {
	ln := len(q.Name()) + len(msg.TaskName)
	data := make([]byte, 0, ln+len(msg.Name))
	data = append(data, q.Name()...)
	data = append(data, msg.TaskName...)
	data = append(data, msg.Name...)

	b := make([]byte, 3+8+8)
	copy(b, "tq:")

	// Hash message name.
	h := farm.Hash64(data[ln:])
	binary.BigEndian.PutUint64(b[3:11], h)

	// Hash queue name and use it as a seed.
	seed := farm.Hash64(data[:ln])

	// Hash everything using the seed.
	h = farm.Hash64WithSeed(data, seed)
	binary.BigEndian.PutUint64(b[11:19], h)

	return unsafeconv.String(b)
}
