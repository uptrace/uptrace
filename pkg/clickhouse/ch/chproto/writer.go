package chproto

import (
	"bufio"
	"encoding/binary"
	"github.com/segmentio/asm/bswap"
	"github.com/uptrace/pkg/unixtime"
	"github.com/uptrace/pkg/unsafeconv"
	"io"
	"math"
	"time"
)

type Compression int

const (
	CompressionNone = 0x02
	CompressionLZ4  = 0x82
	CompressionZSTD = 0x90
)
const (
	checksumSize          = 16
	compressionHeaderSize = 1 + 4 + 4
	headerSize            = checksumSize + compressionHeaderSize
	blockSize             = 1 << 20
)
const (
	uuidLen          = 16
	nanosecondsInDay = 24 * 60 * 60 * 1e9
)

type writer interface {
	io.Writer
	io.ByteWriter
	Flush() error
}
type Writer struct {
	bw  *bufio.Writer
	zw  *zWriter
	wr  writer
	err error
	buf []byte
	lc  *LowCard
}

func NewWriter(w io.Writer) *Writer {
	bw := bufio.NewWriter(w)
	return &Writer{bw: bw, zw: newZWriter(bw, CompressionLZ4), wr: bw, buf: make([]byte, uuidLen)}
}
func (w *Writer) WithCompression(enabled bool, fn func() error) {
	if w.err != nil {
		return
	}
	if enabled {
		w.wr = w.zw
	}
	w.err = fn()
	if enabled {
		if err := w.zw.Close(); err != nil && w.err == nil {
			w.err = err
		}
		w.wr = w.bw
	}
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
func (w *Writer) WriteByte(c byte) {
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
	w.UInt8(num)
}
func (w *Writer) Uvarint(num uint64)  { n := binary.PutUvarint(w.buf, num); w.Write(w.buf[:n]) }
func (w *Writer) UInt8(num uint8)     { w.WriteByte(num) }
func (w *Writer) UInt16(num uint16)   { binary.LittleEndian.PutUint16(w.buf, num); w.Write(w.buf[:2]) }
func (w *Writer) UInt32(num uint32)   { binary.LittleEndian.PutUint32(w.buf, num); w.Write(w.buf[:4]) }
func (w *Writer) UInt64(num uint64)   { binary.LittleEndian.PutUint64(w.buf, num); w.Write(w.buf[:8]) }
func (w *Writer) Int8(num int8)       { w.UInt8(uint8(num)) }
func (w *Writer) Int16(num int16)     { w.UInt16(uint16(num)) }
func (w *Writer) Int32(num int32)     { w.UInt32(uint32(num)) }
func (w *Writer) Int64(num int64)     { w.UInt64(uint64(num)) }
func (w *Writer) Float32(num float32) { w.UInt32(math.Float32bits(num)) }
func (w *Writer) Float64(num float64) { w.UInt64(math.Float64bits(num)) }
func (w *Writer) String(s string) {
	w.Uvarint(uint64(len(s)))
	if s != "" {
		w.Write(unsafeconv.Bytes(s))
	}
}
func (w *Writer) Bytes(b []byte) {
	w.Uvarint(uint64(len(b)))
	if len(b) > 0 {
		w.Write(b)
	}
}
func (w *Writer) UUID(b []byte) {
	if len(b) != uuidLen {
		panic("not reached")
	}
	buf := w.buf[:uuidLen]
	copy(buf, b)
	bswap.Swap64(buf)
	w.Write(buf)
}
func (w *Writer) GoDateTime(t time.Time) {
	if t.IsZero() {
		w.DateTime(0)
	} else {
		w.DateTime(unixtime.ToNano(t))
	}
}
func (w *Writer) DateTime(tm unixtime.Nano) { w.UInt32(uint32(tm / unixtime.Second)) }
func (w *Writer) Date(tm unixtime.Nano)     { w.UInt16(uint16(tm / nanosecondsInDay)) }
func (w *Writer) LowCard() *LowCard {
	if w.lc == nil {
		w.lc = NewLowCard()
	} else {
		w.lc.Reset()
	}
	return w.lc
}
