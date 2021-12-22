package chproto

import (
	"bufio"
	"encoding/binary"
	"sync"

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

type writeBuffer struct {
	buf []byte
}

var writeBufferPool = sync.Pool{
	New: func() any {
		return &writeBuffer{
			buf: make([]byte, blockSize),
		}
	},
}

func getWriterBuffer() *writeBuffer {
	return writeBufferPool.Get().(*writeBuffer)
}

func putWriterBuffer(db *writeBuffer) {
	writeBufferPool.Put(db)
}

//------------------------------------------------------------------------------

type lz4Writer struct {
	wr *bufio.Writer

	data *writeBuffer
	pos  int
}

func newLZ4Writer(w *bufio.Writer) *lz4Writer {
	return &lz4Writer{
		wr: w,
	}
}

func (w *lz4Writer) Init() {
	w.data = getWriterBuffer()
	w.pos = 0
}

func (w *lz4Writer) Close() error {
	err := w.flush()
	putWriterBuffer(w.data)
	w.data = nil
	return err
}

func (w *lz4Writer) Flush() error {
	return w.Close()
}

func (w *lz4Writer) WriteByte(c byte) error {
	w.data.buf[w.pos] = c
	w.pos++
	return w.checkFlush()
}

func (w *lz4Writer) WriteString(s string) (int, error) {
	return w.Write(internal.Bytes(s))
}

func (w *lz4Writer) Write(data []byte) (int, error) {
	var written int
	for len(data) > 0 {
		n := copy(w.data.buf[w.pos:], data)
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
	if w.pos < len(w.data.buf) {
		return nil
	}
	return w.flush()
}

func (w *lz4Writer) flush() error {
	if w.pos == 0 {
		return nil
	}

	zlen := headerSize + lz4.CompressBlockBound(w.pos)
	zdata := make([]byte, zlen)

	compressedSize, err := compress(zdata[headerSize:], w.data.buf[:w.pos])
	if err != nil {
		return err
	}
	compressedSize += compressionHeaderSize

	zdata[16] = lz4Compression
	binary.LittleEndian.PutUint32(zdata[17:], uint32(compressedSize))
	binary.LittleEndian.PutUint32(zdata[21:], uint32(w.pos))

	checkSum := cityhash102.CityHash128(zdata[16:], uint32(compressedSize))
	binary.LittleEndian.PutUint64(zdata[0:], checkSum.Lower64())
	binary.LittleEndian.PutUint64(zdata[8:], checkSum.Higher64())

	w.wr.Write(zdata[:checksumSize+compressedSize])
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
