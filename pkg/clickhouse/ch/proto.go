package ch

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/asm/bswap"
	"github.com/uptrace/pkg/clickhouse/ch/chpool"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	osUser      = os.Getenv("USER")
	hostname, _ = os.Hostname()
)

func (db *DB) hello(ctx context.Context, cn *chpool.Conn) error {
	if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
		wr.WriteByte(chproto.ClientHello)
		writeClientInfo(wr)
		wr.String(db.conf.Database)
		wr.String(db.conf.User)
		wr.String(db.conf.Password)
	}); err != nil {
		return err
	}
	return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
		packet, err := rd.Uvarint()
		if err != nil {
			return err
		}
		switch packet {
		case chproto.ServerHello:
			return cn.ServerInfo.ReadFrom(rd)
		case chproto.ServerException:
			return readException(rd, nil)
		default:
			return fmt.Errorf("ch: hello: unexpected packet: %d", packet)
		}
	})
}

const (
	clientName     = "go-clickhouse"
	chVersionMajor = 1
	chVersionMinor = 1
	chProtoVersion = chproto.DBMS_TCP_PROTOCOL_VERSION
)

func writeClientInfo(wr *chproto.Writer) {
	wr.String(clientName)
	wr.Uvarint(chVersionMajor)
	wr.Uvarint(chVersionMinor)
	wr.Uvarint(chProtoVersion)
}
func readException(rd *chproto.Reader, result *Result) (err error) {
	exc := Error{Result: result}
	if exc.Code, err = rd.Int32(); err != nil {
		return err
	}
	if exc.Name, err = rd.String(); err != nil {
		return err
	}
	if exc.Message, err = rd.String(); err != nil {
		return err
	}
	exc.Message = strings.TrimSpace(strings.TrimPrefix(exc.Message, exc.Name+":"))
	if exc.StackTrace, err = rd.String(); err != nil {
		return err
	}
	hasNested, err := rd.Bool()
	if err != nil {
		return err
	}
	if hasNested {
		exc.Nested = readException(rd, nil)
	}
	return &exc
}
func readProfileInfo(rd *chproto.Reader) error {
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Bool(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Bool(); err != nil {
		return err
	}
	return nil
}

type Progress struct {
	Rows       uint64        `json:"rows"`
	Bytes      uint64        `json:"bytes"`
	TotalRows  uint64        `json:"totalRows"`
	WroteRows  uint64        `json:"wroteRows"`
	WroteBytes uint64        `json:"wroteBytes"`
	Elapsed    time.Duration `json:"elapsed"`
}

func (p *Progress) readFrom(cn *chpool.Conn, rd *chproto.Reader) error {
	rows, err := rd.Uvarint()
	if err != nil {
		return err
	}
	p.Rows += rows
	bytes, err := rd.Uvarint()
	if err != nil {
		return err
	}
	p.Bytes += bytes
	totalRows, err := rd.Uvarint()
	if err != nil {
		return err
	}
	if totalRows != 0 {
		p.TotalRows = totalRows
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_CLIENT_WRITE_INFO {
		wroteRows, err := rd.Uvarint()
		if err != nil {
			return err
		}
		p.WroteRows += wroteRows
		wroteBytes, err := rd.Uvarint()
		if err != nil {
			return err
		}
		p.WroteBytes += wroteBytes
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_SERVER_QUERY_TIME_IN_PROGRESS {
		n, err := rd.Uvarint()
		if err != nil {
			return err
		}
		p.Elapsed += time.Duration(n) * time.Nanosecond
	}
	return nil
}
func writePing(wr *chproto.Writer) { wr.WriteByte(chproto.ClientPing) }
func readPong(rd *chproto.Reader) error {
	for {
		packet, err := rd.Uvarint()
		if err != nil {
			return err
		}
		switch packet {
		case chproto.ServerPong:
			return nil
		case chproto.ServerException:
			return readException(rd, nil)
		case chproto.ServerEndOfStream:
			return nil
		default:
			return fmt.Errorf("ch: readPong: unexpected packet: %d", packet)
		}
	}
}
func (db *DB) writeQuery(ctx context.Context, cn *chpool.Conn, wr *chproto.Writer, query string) {
	wr.WriteByte(chproto.ClientQuery)
	wr.String("")
	wr.WriteByte(chproto.QueryInitial)
	wr.String("")
	wr.String("")
	wr.String(cn.LocalAddr().String())
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_PROTOCOL_VERSION_WITH_INITIAL_QUERY_START_TIME {
		wr.Int64(0)
	}
	wr.WriteByte(1)
	wr.String(osUser)
	wr.String(hostname)
	writeClientInfo(wr)
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_QUOTA_KEY_IN_CLIENT_INFO {
		wr.String("")
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_PROTOCOL_VERSION_WITH_DISTRIBUTED_DEPTH {
		wr.Uvarint(0)
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_VERSION_PATCH {
		wr.Uvarint(0)
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_OPENTELEMETRY {
		if spanCtx := trace.SpanContextFromContext(ctx); false && spanCtx.IsValid() {
			wr.WriteByte(1)
			{
				traceID := spanCtx.TraceID()
				wr.UUID(traceID[:])
			}
			{
				spanID := spanCtx.SpanID()
				bs := spanID[:]
				bswap.Swap64(bs)
				wr.Write(bs)
			}
			wr.String(spanCtx.TraceState().String())
			wr.WriteByte(byte(spanCtx.TraceFlags()))
		} else {
			wr.WriteByte(0)
		}
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS {
		wr.Uvarint(0)
		wr.Uvarint(0)
		wr.Uvarint(0)
	}
	db.writeSettings(cn, wr)
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_INTERSERVER_SECRET {
		wr.String("")
	}
	wr.Uvarint(2)
	wr.Bool(db.conf.Compression)
	wr.String(query)
}
func (db *DB) writeSettings(cn *chpool.Conn, wr *chproto.Writer) {
	for key, value := range db.conf.QuerySettings {
		wr.String(key)
		if cn.ServerInfo.Revision > chproto.DBMS_MIN_REVISION_WITH_SETTINGS_SERIALIZED_AS_STRINGS {
			wr.Bool(true)
			wr.String(fmt.Sprint(value))
			continue
		}
		switch value := value.(type) {
		case string:
			wr.String(value)
		case int:
			wr.Uvarint(uint64(value))
		case int64:
			wr.Uvarint(uint64(value))
		case uint64:
			wr.Uvarint(value)
		case bool:
			wr.Bool(value)
		default:
			panic(fmt.Errorf("%s setting has unsupported type: %T", key, value))
		}
	}
	wr.String("")
}

var emptyBlock Block

func (db *DB) writeBlock(ctx context.Context, wr *chproto.Writer, block *Block) {
	if block == nil {
		block = &emptyBlock
	}
	wr.WriteByte(chproto.ClientData)
	wr.String("")
	wr.WithCompression(db.conf.Compression, func() error {
		writeBlockInfo(wr)
		return block.WriteTo(wr)
	})
}
func writeBlockInfo(wr *chproto.Writer) {
	wr.Uvarint(1)
	wr.Bool(false)
	wr.Uvarint(2)
	wr.Int32(-1)
	wr.Uvarint(0)
}
func (db *DB) readSampleBlock(rd *chproto.Reader) error {
	for {
		packet, err := rd.Uvarint()
		if err != nil {
			return err
		}
		switch packet {
		case chproto.ServerData:
			return db.readBlock(rd, nil, true)
		case chproto.ServerTableColumns:
			if err := readServerTableColumns(rd); err != nil {
				return err
			}
		case chproto.ServerException:
			return readException(rd, nil)
		default:
			return fmt.Errorf("ch: readSampleBlock: unexpected packet: %d", packet)
		}
	}
}
func (db *DB) readDataBlocks(cn *chpool.Conn, rd *chproto.Reader) (*Result, error) {
	var block *Block
	res := NewResult()
	for {
		packet, err := rd.Uvarint()
		if err != nil {
			return nil, err
		}
		switch packet {
		case chproto.ServerData, chproto.ServerTotals, chproto.ServerExtremes:
			if block == nil {
				block = NewBlock()
			}
			if err := db.readBlock(rd, block, true); err != nil {
				return nil, err
			}
			res.affected += int64(block.NumRow)
		case chproto.ServerException:
			return nil, readException(rd, res)
		case chproto.ServerProgress:
			if err := res.progress.readFrom(cn, rd); err != nil {
				return nil, err
			}
		case chproto.ServerProfileInfo:
			if err := readProfileInfo(rd); err != nil {
				return nil, err
			}
		case chproto.ServerTableColumns:
			if err := readServerTableColumns(rd); err != nil {
				return nil, err
			}
		case chproto.ServerProfileEvents:
			if err := withProfileEventsBlock(func(block *Block) error { return db.readBlock(rd, block, false) }); err != nil {
				return nil, err
			}
		case chproto.ServerEndOfStream:
			return res, nil
		default:
			return nil, fmt.Errorf("ch: readDataBlocks: unexpected packet: %d", packet)
		}
	}
}
func readInsertResult(cn *chpool.Conn, rd *chproto.Reader) (*Result, error) {
	packet, err := rd.Uvarint()
	if err != nil {
		return nil, err
	}
	res := NewResult()
	switch packet {
	case chproto.ServerException:
		return nil, readException(rd, res)
	case chproto.ServerProgress:
		if err := res.progress.readFrom(cn, rd); err != nil {
			return nil, err
		}
		return res, nil
	case chproto.ServerProfileInfo:
		if err := readProfileInfo(rd); err != nil {
			return nil, err
		}
		return res, nil
	case chproto.ServerTableColumns:
		if err := readServerTableColumns(rd); err != nil {
			return nil, err
		}
		return res, nil
	case chproto.ServerEndOfStream:
		return res, nil
	default:
		return nil, fmt.Errorf("ch: readInsertResult: unexpected packet: %d", packet)
	}
}
func (db *DB) readBlock(rd *chproto.Reader, block *Block, compressible bool) error {
	if _, err := rd.String(); err != nil {
		return err
	}
	return rd.WithCompression(compressible && db.conf.Compression, func() error {
		if err := readBlockInfo(rd); err != nil {
			return err
		}
		numColumn, err := rd.Uvarint()
		if err != nil {
			return err
		}
		numRow, err := rd.Uvarint()
		if err != nil {
			return err
		}
		if block != nil {
			block.NumColumn = int(numColumn)
			block.NumRow = int(numRow)
			if numColumn == 0 && numRow == 0 {
				for _, col := range block.Columns {
					col.Grow(0)
				}
			}
		}
		for i := 0; i < int(numColumn); i++ {
			colName, err := rd.String()
			if err != nil {
				return err
			}
			if colName == "" {
				return errors.New("ch: column has empty name")
			}
			colType, err := rd.String()
			if err != nil {
				return err
			}
			if colType == "" {
				return fmt.Errorf("ch: column=%s has empty type", colName)
			}
			if block == nil {
				continue
			}
			colName = strings.TrimSuffix(colName, "_")
			col := block.Column(colName, colType)
			col.Grow(int(numRow))
			if numRow > 0 {
				if err := col.ReadPrefix(rd); err != nil {
					return err
				}
				if err := col.ReadData(rd, int(numRow)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func readBlockInfo(rd *chproto.Reader) error {
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Bool(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Int32(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	return nil
}
func writeCancel(wr *chproto.Writer) { wr.WriteByte(chproto.ClientCancel) }
func readServerTableColumns(rd *chproto.Reader) error {
	if _, err := rd.String(); err != nil {
		return err
	}
	if _, err := rd.String(); err != nil {
		return err
	}
	return nil
}

type BlockIter struct {
	db        *DB
	cn        *chpool.Conn
	rd        *chproto.Reader
	release   func(error)
	packet    uint64
	stickyErr error
	result    Result
}

func NewEmptyBlockIter() *BlockIter { return new(BlockIter) }
func newBlockIter(ctx context.Context, db *DB, cn *chpool.Conn, release func(error)) (*BlockIter, error) {
	rd := cn.Reader(ctx, db.conf.ReadTimeout)
	packet, err := rd.Uvarint()
	if err != nil {
		return nil, err
	}
	if packet == chproto.ServerException {
		return nil, readException(rd, nil)
	}
	return &BlockIter{db: db, cn: cn, rd: rd, release: release, packet: packet}, nil
}

var errClosed = errors.New("ch: closed before reading full data")

func (it *BlockIter) Close() error {
	if it.closed() {
		return nil
	}
	it.close(errClosed)
	return nil
}
func (it *BlockIter) closed() bool { return it.cn == nil }
func (it *BlockIter) close(err error) {
	it.release(err)
	it.cn = nil
}
func (it *BlockIter) Err() error { return it.stickyErr }
func (it *BlockIter) Next(block *Block) bool {
	if it.closed() {
		return false
	}
	ok, err := it.read(it.rd, block)
	if err != nil {
		it.stickyErr = err
		it.close(err)
		return false
	}
	if !ok {
		it.close(nil)
		return false
	}
	it.result.affected += int64(block.NumRow)
	return true
}
func (it *BlockIter) read(rd *chproto.Reader, block *Block) (bool, error) {
	for {
		packet, err := it.readPacket(it.rd)
		if err != nil {
			return false, err
		}
		switch packet {
		case chproto.ServerData:
			if err := it.db.readBlock(rd, block, true); err != nil {
				return false, err
			}
			return true, nil
		case chproto.ServerException:
			return false, readException(rd, &it.result)
		case chproto.ServerProgress:
			if err := it.result.progress.readFrom(it.cn, rd); err != nil {
				return false, err
			}
		case chproto.ServerProfileInfo:
			if err := readProfileInfo(rd); err != nil {
				return false, err
			}
		case chproto.ServerTableColumns:
			if err := readServerTableColumns(rd); err != nil {
				return false, err
			}
		case chproto.ServerProfileEvents:
			if err := withProfileEventsBlock(func(block *Block) error { return it.db.readBlock(rd, block, false) }); err != nil {
				return false, err
			}
		case chproto.ServerEndOfStream:
			return false, nil
		default:
			return false, fmt.Errorf("ch: BlockIter.Next: unexpected packet: %d", packet)
		}
	}
}
func (it *BlockIter) readPacket(rd *chproto.Reader) (uint64, error) {
	if it.packet != 0 {
		packet := it.packet
		it.packet = 0
		return packet, nil
	}
	return rd.Uvarint()
}
func (it *BlockIter) Result() *Result { return &it.result }

type Block struct {
	Table     *chschema.Table
	NumColumn int
	NumRow    int
	Columns   []*chschema.Column
	columnMap map[string]*chschema.Column
}

func NewBlock() *Block { return &Block{} }
func (b *Block) initTable(table *chschema.Table, numCol, numRow int) {
	b.Table = table
	b.NumColumn = numCol
	b.NumRow = numRow
}
func (b *Block) Clear() {
	for _, col := range b.Columns {
		col.Clear()
	}
}
func (b *Block) ColumnForField(field *chschema.Field) *chschema.Column {
	col := b.Column(field.CHName, field.CHType)
	col.Field = field
	return col
}
func (b *Block) Column(colName, colType string) *chschema.Column {
	if col, ok := b.columnMap[colName]; ok {
		if col.Type == colType {
			return col
		}
		slog.Error("block column has incorrect type (replacing)", slog.String("col_name", col.Name), slog.String("col_type", col.Type), slog.String("wanted", colType))
	}
	var col *chschema.Column
	if b.Table != nil {
		col = b.Table.NewColumn(colName, colType)
	} else {
		col = &chschema.Column{Name: colName, Type: colType, Columnar: chschema.NewColumn(colType, nil)}
	}
	b.AddColumn(colName, col)
	return col
}
func (b *Block) AddColumnar(colName, colType string, columnar chschema.Columnar) {
	if err := columnar.Init(colType, nil); err != nil {
		panic(err)
	}
	b.AddColumn(colName, &chschema.Column{Name: colName, Type: colType, Columnar: columnar})
}
func (b *Block) AddColumn(colName string, col *chschema.Column) {
	if b.Columns == nil && b.columnMap == nil {
		b.Columns = make([]*chschema.Column, 0, b.NumColumn)
		b.columnMap = make(map[string]*chschema.Column, b.NumColumn)
	}
	b.Columns = append(b.Columns, col)
	b.columnMap[colName] = col
}
func (b *Block) WriteTo(wr *chproto.Writer) error {
	var numRow int
	if len(b.Columns) > 0 {
		numRow = b.Columns[0].Len()
	}
	wr.Uvarint(uint64(len(b.Columns)))
	wr.Uvarint(uint64(numRow))
	for _, col := range b.Columns {
		if col.Len() != numRow {
			err := fmt.Errorf("%s does not have expected number of rows: got %d, wanted %d", col, col.Len(), numRow)
			panic(err)
		}
		wr.String(col.Name)
		wr.String(col.Type)
		if err := col.WritePrefix(wr); err != nil {
			return err
		}
		if err := col.WriteData(wr); err != nil {
			return err
		}
	}
	return nil
}
func (b *Block) Scan(dest ...any) error {
	if b.NumRow == 0 {
		for _, dest := range dest {
			v := reflect.ValueOf(dest).Elem()
			v.Set(v.Slice(0, 0))
		}
		return nil
	}
	if b.NumColumn != len(dest) {
		return fmt.Errorf("ch: got %d columns, but Scan has %d values", b.NumColumn, len(dest))
	}
	for i, dest := range dest {
		col := b.Columns[i]
		v := reflect.ValueOf(dest).Elem()
		v.Set(reflect.ValueOf(col.Value()))
	}
	return nil
}

var eventsBlockPool = sync.Pool{New: func() any { return NewBlock() }}

func withProfileEventsBlock(fn func(block *Block) error) error {
	block := eventsBlockPool.Get().(*Block)
	err := fn(block)
	eventsBlockPool.Put(block)
	return err
}
