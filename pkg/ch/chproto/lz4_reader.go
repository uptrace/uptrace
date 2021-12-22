package chproto

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/pierrec/lz4/v4"
)

var errUnreadData = errors.New("ch: lz4 reader was closed with unread data")

type lz4Reader struct {
	rd *bufio.Reader

	header []byte

	data []byte
	pos  int
}

func newLZ4Reader(r *bufio.Reader) *lz4Reader {
	return &lz4Reader{
		rd: r,

		header: make([]byte, headerSize),
	}
}

func (r *lz4Reader) Init() {}

func (r *lz4Reader) Release() error {
	var err error
	if r.Buffered() > 0 {
		err = errUnreadData
	}

	r.data = nil
	r.pos = 0

	return err
}

func (r *lz4Reader) Buffered() int {
	return len(r.data) - r.pos
}

func (r *lz4Reader) Read(buf []byte) (int, error) {
	var nread int

	if r.pos < len(r.data) {
		n := copy(buf, r.data[r.pos:])
		nread += n
		r.pos += n
	}

	for nread < len(buf) {
		if err := r.readData(); err != nil {
			return nread, err
		}

		n := copy(buf[nread:], r.data)
		nread += n
		r.pos = n
	}

	return nread, nil
}

func (r *lz4Reader) ReadByte() (byte, error) {
	if r.pos == len(r.data) {
		if err := r.readData(); err != nil {
			return 0, err
		}
	}

	if r.pos < len(r.data) {
		c := r.data[r.pos]
		r.pos++
		return c, nil
	}

	return 0, io.EOF
}

func (r *lz4Reader) readData() error {
	if r.pos != len(r.data) {
		panic("not reached")
	}

	_, err := io.ReadFull(r.rd, r.header)
	if err != nil {
		return err
	}

	if r.header[16] != lz4Compression {
		return fmt.Errorf("ch: unsupported compression method: 0x%02x", r.header[16])
	}

	compressedSize := int(binary.LittleEndian.Uint32(r.header[17:])) - compressionHeaderSize
	uncompressedSize := int(binary.LittleEndian.Uint32(r.header[21:]))

	zdata := make([]byte, compressedSize)
	r.data = grow(r.data, uncompressedSize)

	if _, err := io.ReadFull(r.rd, zdata); err != nil {
		return err
	}
	if _, err := lz4.UncompressBlock(zdata, r.data); err != nil {
		return err
	}

	r.pos = 0
	return nil
}

func grow(b []byte, n int) []byte {
	if cap(b) < n {
		return make([]byte, n)
	}
	return b[:n]
}
