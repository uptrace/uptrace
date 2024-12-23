package chproto

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/go-faster/city"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/uptrace/pkg/unsafeconv"
)

var zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest), zstd.WithEncoderConcurrency(1), zstd.WithLowerEncoderMem(true))

type zWriter struct {
	wr          *bufio.Writer
	lz4         lz4.Compressor
	compression Compression
	data        []byte
	pos         int
	zbuf        []byte
}

func newZWriter(wr *bufio.Writer, compression Compression) *zWriter {
	return &zWriter{wr: wr, compression: compression, data: make([]byte, blockSize)}
}
func (w *zWriter) Init() {}
func (w *zWriter) Close() error {
	err := w.flush()
	w.pos = 0
	return err
}
func (w *zWriter) Flush() error { return w.Close() }
func (w *zWriter) WriteByte(c byte) error {
	w.data[w.pos] = c
	w.pos++
	return w.flushIfFull()
}
func (w *zWriter) WriteString(s string) (int, error) { return w.Write(unsafeconv.Bytes(s)) }
func (w *zWriter) Write(data []byte) (int, error) {
	var written int
	for len(data) > 0 {
		n := copy(w.data[w.pos:], data)
		data = data[n:]
		w.pos += n
		if err := w.flushIfFull(); err != nil {
			return written, err
		}
		written += n
	}
	return written, nil
}
func (w *zWriter) flushIfFull() error {
	if w.pos < len(w.data) {
		return nil
	}
	return w.flush()
}
func (w *zWriter) flush() error {
	if w.pos == 0 {
		return nil
	}
	err := w.writeBlock(w.data[:w.pos])
	w.pos = 0
	return err
}
func (w *zWriter) writeBlock(data []byte) error {
	zlen := headerSize + lz4.CompressBlockBound(len(data))
	w.zbuf = grow(w.zbuf, zlen)
	var compressedSize int
	switch w.compression {
	case CompressionLZ4:
		var err error
		compressedSize, err = w.lz4.CompressBlock(data, w.zbuf[headerSize:])
		if err != nil {
			return err
		}
		w.zbuf[16] = CompressionLZ4
	case CompressionZSTD:
		w.zbuf = zstdEncoder.EncodeAll(data, w.zbuf[:headerSize])
		compressedSize = len(w.zbuf) - headerSize
		w.zbuf[16] = CompressionZSTD
	default:
		return fmt.Errorf("ch: unsupported compression method: 0x%02x", w.compression)
	}
	compressedSize += compressionHeaderSize
	w.zbuf = w.zbuf[:checksumSize+compressedSize]
	binary.LittleEndian.PutUint32(w.zbuf[17:], uint32(compressedSize))
	binary.LittleEndian.PutUint32(w.zbuf[21:], uint32(w.pos))
	checkSum := city.CH128(w.zbuf[16:])
	binary.LittleEndian.PutUint64(w.zbuf[0:8], checkSum.Low)
	binary.LittleEndian.PutUint64(w.zbuf[8:16], checkSum.High)
	w.wr.Write(w.zbuf)
	return nil
}
