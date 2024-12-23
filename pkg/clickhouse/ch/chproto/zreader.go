package chproto

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"io"
)

var errUnreadData = errors.New("ch: lz4 reader was closed with unread data")
var zstdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderLowmem(true))

type zReader struct {
	rd          *bufio.Reader
	compression Compression
	header      []byte
	data        []byte
	pos         int
	zbuf        []byte
}

func newZReader(r *bufio.Reader) *zReader { return &zReader{rd: r, header: make([]byte, headerSize)} }
func (r *zReader) Release() error {
	var err error
	if r.Buffered() > 0 {
		err = errUnreadData
	}
	r.data = r.data[:0]
	r.pos = 0
	r.zbuf = r.zbuf[:0]
	return err
}
func (r *zReader) Buffered() int { return len(r.data) - r.pos }
func (r *zReader) Read(buf []byte) (int, error) {
	var nread int
	if r.pos < len(r.data) {
		n := copy(buf, r.data[r.pos:])
		nread += n
		r.pos += n
	}
	for nread < len(buf) {
		if err := r.readBlock(); err != nil {
			return nread, err
		}
		n := copy(buf[nread:], r.data)
		nread += n
		r.pos = n
	}
	return nread, nil
}
func (r *zReader) ReadByte() (byte, error) {
	if r.pos == len(r.data) {
		if err := r.readBlock(); err != nil {
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
func (r *zReader) readBlock() error {
	if r.pos != len(r.data) {
		panic("not reached")
	}
	if _, err := io.ReadFull(r.rd, r.header); err != nil {
		return err
	}
	compressedSize := int(binary.LittleEndian.Uint32(r.header[17:])) - compressionHeaderSize
	uncompressedSize := int(binary.LittleEndian.Uint32(r.header[21:]))
	r.zbuf = grow(r.zbuf, compressedSize)
	r.data = grow(r.data, uncompressedSize)
	if _, err := io.ReadFull(r.rd, r.zbuf); err != nil {
		return err
	}
	switch r.header[16] {
	case CompressionLZ4:
		if _, err := lz4.UncompressBlock(r.zbuf, r.data); err != nil {
			return err
		}
	case CompressionZSTD:
		data, err := zstdDecoder.DecodeAll(r.zbuf, r.data)
		if err != nil {
			return err
		}
		if len(data) != uncompressedSize {
			return fmt.Errorf("ch: unexpected uncompressed data size: %d != %d", len(data), uncompressedSize)
		}
		r.data = data
	default:
		return fmt.Errorf("ch: unsupported compression method: 0x%02x", r.header[16])
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
