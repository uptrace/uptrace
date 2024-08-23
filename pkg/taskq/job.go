package taskq

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/zstd"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/taskq/v4/backend"
	"github.com/vmihailenco/taskq/v4/backend/unsafeconv"
)

// ErrDuplicate is returned when adding duplicate message to the queue.
var ErrDuplicate = errors.New("taskq: message with such name already exists")

// Job is used to create and retrieve messages from a queue.
type Job struct {
	// SQS/IronMQ message id.
	ID string `msgpack:"1,omitempty,alias:ID"`

	// Optional name for the message. Jobs with the same name
	// are processed only once.
	Name string `msgpack:"-"`

	// Delay specifies the duration the queue must wait
	// before executing the message.
	Delay time.Duration `msgpack:"-"`

	// Args passed to the handler.
	Args []interface{} `msgpack:"-"`

	// Binary representation of the args.
	ArgsCompression string `msgpack:"2,omitempty,alias:ArgsCompression"`
	ArgsBin         []byte `msgpack:"3,alias:ArgsBin"`

	// SQS/IronMQ reservation id that is used to release/delete the message.
	ReservationID string `msgpack:"-"`

	// The number of times the message has been reserved or released.
	ReservedCount int `msgpack:"4,omitempty,alias:ReservedCount"`

	TaskName string `msgpack:"5,alias:TaskName"`
	Err      error  `msgpack:"-"`

	evt                *ProcessJobEvent
	marshalBinaryCache []byte
}

func NewJob(args ...interface{}) *Job {
	return &Job{
		Args: args,
	}
}

func (m *Job) String() string {
	return fmt.Sprintf("Job<ID=%q Name=%q ReservedCount=%d>",
		m.ID, m.Name, m.ReservedCount)
}

// SetDelay sets the message delay.
func (m *Job) SetDelay(delay time.Duration) {
	m.Delay = delay
}

// OnceInPeriod uses the period and the args to generate such a message name
// that message with such args is added to the queue once in a given period.
// If args are not provided then message args are used instead.
func (m *Job) OnceInPeriod(period time.Duration, args ...interface{}) {
	m.setNameFromArgs(period, args...)
	m.SetDelay(period)
}

func (m *Job) OnceWithDelay(delay time.Duration) {
	m.setNameFromArgs(0)
	if delay > 0 {
		m.SetDelay(delay)
	}
}

func (m *Job) OnceWithSchedule(tm time.Time) {
	m.OnceWithDelay(time.Until(tm))
}

func (m *Job) setNameFromArgs(period time.Duration, args ...interface{}) {
	var b []byte
	if len(args) > 0 {
		b, _ = msgpack.Marshal(args)
	} else {
		b, _ = m.MarshalArgs()
	}
	if period > 0 {
		b = appendTimeSlot(b, period)
	}
	m.Name = unsafeconv.String(b)
}

var zdec, _ = zstd.NewReader(nil)

func (m *Job) decompress() ([]byte, error) {
	switch m.ArgsCompression {
	case "":
		return m.ArgsBin, nil
	case "zstd":
		return zdec.DecodeAll(m.ArgsBin, nil)
	case "s2":
		return s2.Decode(nil, m.ArgsBin)
	default:
		return nil, fmt.Errorf("taskq: unsupported compression=%s", m.ArgsCompression)
	}
}

func (m *Job) MarshalArgs() ([]byte, error) {
	if m.ArgsBin != nil {
		if m.ArgsCompression == "" {
			return m.ArgsBin, nil
		}
		if m.Args == nil {
			return m.decompress()
		}
	}

	b, err := msgpack.Marshal(m.Args)
	if err != nil {
		return nil, err
	}
	m.ArgsBin = b

	return b, nil
}

type messageRaw Job

var _ encoding.BinaryMarshaler = (*Job)(nil)

func (m *Job) MarshalBinary() ([]byte, error) {
	if m.TaskName == "" {
		return nil, backend.ErrTaskNameRequired
	}
	if m.marshalBinaryCache != nil {
		return m.marshalBinaryCache, nil
	}

	_, err := m.MarshalArgs()
	if err != nil {
		return nil, err
	}

	if m.ArgsCompression == "" && len(m.ArgsBin) >= 512 {
		compressed := s2.Encode(nil, m.ArgsBin)
		if len(compressed) < len(m.ArgsBin) {
			m.ArgsCompression = "s2"
			m.ArgsBin = compressed
		}
	}

	b, err := msgpack.Marshal((*messageRaw)(m))
	if err != nil {
		return nil, err
	}

	m.marshalBinaryCache = b
	return b, nil
}

var _ encoding.BinaryUnmarshaler = (*Job)(nil)

func (m *Job) UnmarshalBinary(b []byte) error {
	if err := msgpack.Unmarshal(b, (*messageRaw)(m)); err != nil {
		return err
	}

	b, err := m.decompress()
	if err != nil {
		return err
	}

	m.ArgsCompression = ""
	m.ArgsBin = b

	return nil
}

func appendTimeSlot(b []byte, period time.Duration) []byte {
	l := len(b)
	b = append(b, make([]byte, 16)...)
	binary.LittleEndian.PutUint64(b[l:], uint64(period))
	binary.LittleEndian.PutUint64(b[l+8:], uint64(timeSlot(period)))
	return b
}

func timeSlot(period time.Duration) int64 {
	if period <= 0 {
		return 0
	}
	return time.Now().UnixNano() / int64(period)
}
