package chproto

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/uptrace/go-clickhouse/ch/internal"
)

const uuidLen = 16

type writer interface {
	io.Writer
	io.ByteWriter
	Flush() error
}

type Writer struct {
	bw *bufio.Writer
	zw *lz4Writer
	wr writer // points to bw or zw

	err error

	buf []byte
}

func NewWriter(w io.Writer) *Writer {
	bw := bufio.NewWriter(w)
	return &Writer{
		bw: bw,
		zw: newLZ4Writer(bw),
		wr: bw,

		buf: make([]byte, uuidLen),
	}
}

func (w *Writer) WithCompression(fn func() error) {
	if w.err != nil {
		return
	}

	w.zw.Init()
	w.wr = w.zw

	w.err = fn()

	if err := w.zw.Close(); err != nil && w.err == nil {
		w.err = err
	}
	w.wr = w.bw
}

func (w *Writer) Flush() (err error) {
	if w.err != nil {
		err, w.err = w.err, nil
		return err
	}

	return w.wr.Flush()
}

func (w *Writer) Write(b []byte) {
	if w.err != nil {
		return
	}

	_, err := w.wr.Write(b)
	w.err = err
}

func (w *Writer) writeByte(c byte) {
	if w.err != nil {
		return
	}
	w.err = w.wr.WriteByte(c)
}

func (w *Writer) Bool(flag bool) {
	var num uint8
	if flag {
		num = 1
	}
	w.Uint8(num)
}

func (w *Writer) Uvarint(num uint64) {
	n := binary.PutUvarint(w.buf, num)
	w.Write(w.buf[:n])
}

func (w *Writer) Uint8(num uint8) {
	w.writeByte(num)
}

func (w *Writer) Uint16(num uint16) {
	binary.LittleEndian.PutUint16(w.buf, num)
	w.Write(w.buf[:2])
}

func (w *Writer) Uint32(num uint32) {
	binary.LittleEndian.PutUint32(w.buf, num)
	w.Write(w.buf[:4])
}

func (w *Writer) Uint64(num uint64) {
	binary.LittleEndian.PutUint64(w.buf, num)
	w.Write(w.buf[:8])
}

func (w *Writer) Int8(num int8) {
	w.Uint8(uint8(num))
}

func (w *Writer) Int16(num int16) {
	w.Uint16(uint16(num))
}

func (w *Writer) Int32(num int32) {
	w.Uint32(uint32(num))
}

func (w *Writer) Int64(num int64) {
	w.Uint64(uint64(num))
}

func (w *Writer) Float32(num float32) {
	w.Uint32(math.Float32bits(num))
}

func (w *Writer) Float64(num float64) {
	w.Uint64(math.Float64bits(num))
}

func (w *Writer) String(s string) {
	w.Uvarint(uint64(len(s)))
	w.Write(internal.Bytes(s))
}

func (w *Writer) Bytes(b []byte) {
	w.Uvarint(uint64(len(b)))
	w.Write(b)
}

func (w *Writer) UUID(b []byte) {
	if len(b) != uuidLen {
		panic("not reached")
	}

	buf := w.buf[:uuidLen]
	copy(buf, b)
	packUUID(buf)
	w.Write(buf)
}

// 2 int64 in little endian order?
func packUUID(b []byte) []byte {
	_ = b[15]
	b[0], b[7] = b[7], b[0]
	b[1], b[6] = b[6], b[1]
	b[2], b[5] = b[5], b[2]
	b[3], b[4] = b[4], b[3]
	b[8], b[15] = b[15], b[8]
	b[9], b[14] = b[14], b[9]
	b[10], b[13] = b[13], b[10]
	b[11], b[12] = b[12], b[11]
	return b
}

func (w *Writer) DateTime(tm time.Time) {
	w.Uint32(uint32(unixTime(tm)))
}

const secsInDay = 24 * 3600

func (w *Writer) Date(tm time.Time) {
	w.Uint16(uint16(unixTime(tm) / secsInDay))
}

func unixTime(tm time.Time) int64 {
	if tm.IsZero() {
		return 0
	}
	return tm.Unix()
}
