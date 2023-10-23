package chproto

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/uptrace/go-clickhouse/ch/internal"
)

type reader interface {
	io.Reader
	io.ByteReader
	Buffered() int
}

type Reader struct {
	br *bufio.Reader
	zr *lz4Reader
	rd reader // points to br or zr

	buf []byte
}

func NewReader(r io.Reader) *Reader {
	br := bufio.NewReader(r)
	return &Reader{
		br: br,
		zr: newLZ4Reader(br),
		rd: br,

		buf: make([]byte, uuidLen),
	}
}

func (r *Reader) WithCompression(enabled bool, fn func() error) error {
	if enabled {
		r.rd = r.zr
	}

	firstErr := fn()

	if enabled {
		r.rd = r.br
		if err := r.zr.Release(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (r *Reader) Read(buf []byte) (int, error) {
	return r.rd.Read(buf)
}

func (r *Reader) Buffered() int {
	return r.rd.Buffered()
}

func (r *Reader) Bool() (bool, error) {
	c, err := r.rd.ReadByte()
	if err != nil {
		return false, err
	}
	return c == 1, nil
}

func (r *Reader) Uvarint() (uint64, error) {
	return binary.ReadUvarint(r.rd)
}

func (r *Reader) UInt8() (uint8, error) {
	c, err := r.rd.ReadByte()
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (r *Reader) UInt16() (uint16, error) {
	b, err := r.readNTemp(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (r *Reader) UInt32() (uint32, error) {
	b, err := r.readNTemp(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *Reader) UInt64() (uint64, error) {
	b, err := r.readNTemp(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (r *Reader) Int8() (int8, error) {
	num, err := r.UInt8()
	return int8(num), err
}

func (r *Reader) Int16() (int16, error) {
	num, err := r.UInt16()
	return int16(num), err
}

func (r *Reader) Int32() (int32, error) {
	num, err := r.UInt32()
	return int32(num), err
}

func (r *Reader) Int64() (int64, error) {
	num, err := r.UInt64()
	return int64(num), err
}

func (r *Reader) Float32() (float32, error) {
	num, err := r.UInt32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(num), nil
}

func (r *Reader) Float64() (float64, error) {
	num, err := r.UInt64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(num), nil
}

func (r *Reader) Bytes() ([]byte, error) {
	num, err := r.Uvarint()
	if err != nil {
		return nil, err
	}

	b := make([]byte, int(num))
	_, err = io.ReadFull(r.rd, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Reader) String() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return internal.String(b), nil
}

func (r *Reader) UUID(b []byte) error {
	if len(b) != uuidLen {
		return fmt.Errorf("got %d bytes, wanted %d", len(b), uuidLen)
	}

	_, err := io.ReadFull(r.rd, b)
	if err != nil {
		return err
	}
	packUUID(b)
	return nil
}

func (r *Reader) readNTemp(n int) ([]byte, error) {
	buf := r.buf[:n]
	_, err := io.ReadFull(r.rd, buf)
	return buf, err
}

func (r *Reader) DateTime() (time.Time, error) {
	sec, err := r.UInt32()
	if err != nil {
		return time.Time{}, err
	}
	if sec == 0 {
		return time.Time{}, nil
	}
	return time.Unix(int64(sec), 0), nil
}

func (r *Reader) Date() (time.Time, error) {
	days, err := r.UInt16()
	if err != nil {
		return time.Time{}, err
	}
	if days == 0 {
		return time.Time{}, nil
	}
	return time.Unix(int64(days)*secsInDay, 0), nil
}
