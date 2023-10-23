package chproto

import (
	"bufio"
	"encoding/binary"

	"github.com/pierrec/lz4/v4"

	"github.com/uptrace/go-clickhouse/ch/internal"
	"github.com/uptrace/go-clickhouse/ch/internal/cityhash102"
)

const (
	noCompression   = 0x02
	lz4Compression  = 0x82
	zstdCompression = 0x90
)

const (
	checksumSize          = 16        // city hash 128
	compressionHeaderSize = 1 + 4 + 4 // method + compressed + uncompressed

	headerSize = checksumSize + compressionHeaderSize
	blockSize  = 1 << 20 // 1 MB
)

//------------------------------------------------------------------------------

type lz4Writer struct {
	wr *bufio.Writer

	data  []byte
	pos   int
	zdata []byte
}

func newLZ4Writer(w *bufio.Writer) *lz4Writer {
	return &lz4Writer{
		wr:   w,
		data: make([]byte, blockSize),
	}
}

func (w *lz4Writer) Close() error {
	err := w.flush()
	w.pos = 0
	return err
}

func (w *lz4Writer) Flush() error {
	return w.Close()
}

func (w *lz4Writer) WriteByte(c byte) error {
	w.data[w.pos] = c
	w.pos++
	return w.checkFlush()
}

func (w *lz4Writer) WriteString(s string) (int, error) {
	return w.Write(internal.Bytes(s))
}

func (w *lz4Writer) Write(data []byte) (int, error) {
	var written int
	for len(data) > 0 {
		n := copy(w.data[w.pos:], data)
		data = data[n:]
		w.pos += n
		if err := w.checkFlush(); err != nil {
			return written, err
		}
		written += n
	}
	return written, nil
}

func (w *lz4Writer) checkFlush() error {
	if w.pos < len(w.data) {
		return nil
	}
	return w.flush()
}

func (w *lz4Writer) flush() error {
	if w.pos == 0 {
		return nil
	}

	zlen := headerSize + lz4.CompressBlockBound(w.pos)
	w.zdata = grow(w.zdata, zlen)

	compressedSize, err := compress(w.zdata[headerSize:], w.data[:w.pos])
	if err != nil {
		return err
	}
	compressedSize += compressionHeaderSize

	w.zdata[16] = lz4Compression
	binary.LittleEndian.PutUint32(w.zdata[17:], uint32(compressedSize))
	binary.LittleEndian.PutUint32(w.zdata[21:], uint32(w.pos))

	checkSum := cityhash102.CityHash128(w.zdata[16:], uint32(compressedSize))
	binary.LittleEndian.PutUint64(w.zdata[0:], checkSum.Lower64())
	binary.LittleEndian.PutUint64(w.zdata[8:], checkSum.Higher64())

	w.wr.Write(w.zdata[:checksumSize+compressedSize])
	w.pos = 0

	return nil
}

//------------------------------------------------------------------------------

func compress(dest, src []byte) (int, error) {
	if len(src) < 16 {
		return uncompressable(dest, src), nil
	}
	var c lz4.Compressor
	return c.CompressBlock(src, dest)
}

func uncompressable(dest, src []byte) int {
	dest[0] = byte(len(src)) << 4
	copy(dest[1:], src)
	return len(src) + 1
}
