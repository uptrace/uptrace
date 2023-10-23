package ch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/uptrace/go-clickhouse/ch/chpool"
	"github.com/uptrace/go-clickhouse/ch/chproto"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"go.opentelemetry.io/otel/trace"
)

const (
	clientName     = "go-clickhouse"
	chVersionMajor = 1
	chVersionMinor = 1
	chProtoVersion = chproto.DBMS_TCP_PROTOCOL_VERSION
)

var (
	osUser      = os.Getenv("USER")
	hostname, _ = os.Hostname()
)

type blockIter struct {
	db *DB
	cn *chpool.Conn

	stickyErr error
}

func newBlockIter(db *DB, cn *chpool.Conn) *blockIter {
	return &blockIter{
		db: db,
		cn: cn,
	}
}

func (it *blockIter) Close() error {
	if it.cn != nil {
		it.close()
	}
	return nil
}

func (it *blockIter) close() {
	it.db.releaseConn(it.cn, it.stickyErr)
	it.cn = nil
}

func (it *blockIter) Err() error {
	return it.stickyErr
}

func (it *blockIter) Next(ctx context.Context, block *chschema.Block) bool {
	if it.cn == nil {
		return false
	}

	ok, err := it.read(ctx, block)
	if err != nil {
		it.stickyErr = err
		it.close()
		return false
	}

	if !ok {
		it.close()
		return false
	}
	return true
}

func (it *blockIter) read(ctx context.Context, block *chschema.Block) (bool, error) {
	rd := it.cn.Reader(ctx, it.db.conf.ReadTimeout)
	for {
		packet, err := rd.Uvarint()
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
			return false, readException(rd)
		case chproto.ServerProgress:
			if err := readProgress(it.cn, rd); err != nil {
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
			block := new(chschema.Block)
			if err := it.db.readBlock(rd, block, false); err != nil {
				return false, err
			}
		case chproto.ServerEndOfStream:
			return false, nil
		default:
			return false, fmt.Errorf("ch: blockIter.Next: unexpected packet: %d", packet)
		}
	}
}

func (db *DB) hello(ctx context.Context, cn *chpool.Conn) error {
	err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
		wr.WriteByte(chproto.ClientHello)
		writeClientInfo(wr)

		wr.String(db.conf.Database)
		wr.String(db.conf.User)
		wr.String(db.conf.Password)
	})
	if err != nil {
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
			return readException(rd)
		default:
			return fmt.Errorf("ch: hello: unexpected packet: %d", packet)
		}
	})
}

func writeClientInfo(wr *chproto.Writer) {
	wr.String(clientName)
	wr.Uvarint(chVersionMajor)
	wr.Uvarint(chVersionMinor)
	wr.Uvarint(chProtoVersion)
}

func readException(rd *chproto.Reader) (err error) {
	var exc Error

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
		exc.Nested = readException(rd)
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

func readProgress(cn *chpool.Conn, rd *chproto.Reader) error {
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if _, err := rd.Uvarint(); err != nil {
		return err
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_CLIENT_WRITE_INFO {
		if _, err := rd.Uvarint(); err != nil {
			return err
		}
		if _, err := rd.Uvarint(); err != nil {
			return err
		}
	}
	return nil
}

func writePing(wr *chproto.Writer) {
	wr.WriteByte(chproto.ClientPing)
}

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
			return readException(rd)
		case chproto.ServerEndOfStream:
			return nil
		default:
			return fmt.Errorf("ch: readPong: unexpected packet: %d", packet)
		}
	}
}

func (db *DB) writeQuery(ctx context.Context, cn *chpool.Conn, wr *chproto.Writer, query string) {
	wr.WriteByte(chproto.ClientQuery)
	wr.String("") // query id

	// TODO: use QuerySecondary - https://github.com/ClickHouse/ClickHouse/blob/master/dbms/src/Client/Connection.cpp#L388-L404
	wr.WriteByte(chproto.QueryInitial)
	wr.String("") // initial user
	wr.String("") // initial query id
	wr.String(cn.LocalAddr().String())
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_PROTOCOL_VERSION_WITH_INITIAL_QUERY_START_TIME {
		wr.Int64(0) // initial_query_start_time_microseconds
	}
	wr.WriteByte(1) // interface [tcp - 1, http - 2]
	wr.String(osUser)
	wr.String(hostname)
	writeClientInfo(wr)
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_QUOTA_KEY_IN_CLIENT_INFO {
		wr.String("") // quota key
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_PROTOCOL_VERSION_WITH_DISTRIBUTED_DEPTH {
		wr.Uvarint(0)
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_VERSION_PATCH {
		wr.Uvarint(0) // client version patch
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_OPENTELEMETRY {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			wr.WriteByte(1)
			{
				v := spanCtx.TraceID()
				wr.UUID(v[:])
			}
			{
				v := spanCtx.SpanID()
				wr.Write(reverseBytes(v[:]))
			}
			wr.String(spanCtx.TraceState().String())
			wr.WriteByte(byte(spanCtx.TraceFlags()))
		} else {
			wr.WriteByte(0)
		}
	}
	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS {
		wr.Uvarint(0) // collaborate_with_initiator
		wr.Uvarint(0) // count_participating_replicas
		wr.Uvarint(0) // number_of_current_replica
	}

	db.writeSettings(cn, wr)

	if cn.ServerInfo.Revision >= chproto.DBMS_MIN_REVISION_WITH_INTERSERVER_SECRET {
		wr.String("")
	}
	wr.Uvarint(2) // state complete
	wr.Bool(db.conf.Compression)
	wr.String(query)
}

func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

func (db *DB) writeSettings(cn *chpool.Conn, wr *chproto.Writer) {
	for key, value := range db.conf.QuerySettings {
		wr.String(key)

		if cn.ServerInfo.Revision > chproto.DBMS_MIN_REVISION_WITH_SETTINGS_SERIALIZED_AS_STRINGS {
			wr.Bool(true) // is_important
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

	wr.String("") // end of settings
}

var emptyBlock chschema.Block

func (db *DB) writeBlock(ctx context.Context, wr *chproto.Writer, block *chschema.Block) {
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

func (db *DB) readSampleBlock(rd *chproto.Reader) (*chschema.Block, error) {
	for {
		packet, err := rd.Uvarint()
		if err != nil {
			return nil, err
		}

		switch packet {
		case chproto.ServerData:
			block := new(chschema.Block)
			if err := db.readBlock(rd, block, true); err != nil {
				return nil, err
			}
			return block, nil
		case chproto.ServerTableColumns:
			if err := readServerTableColumns(rd); err != nil {
				return nil, err
			}
		case chproto.ServerException:
			return nil, readException(rd)
		default:
			return nil, fmt.Errorf("ch: readSampleBlock: unexpected packet: %d", packet)
		}
	}
}

func (db *DB) readDataBlocks(cn *chpool.Conn, rd *chproto.Reader) (*result, error) {
	var res *result
	block := new(chschema.Block)
	for {
		packet, err := rd.Uvarint()
		if err != nil {
			return nil, err
		}

		switch packet {
		case chproto.ServerData, chproto.ServerTotals, chproto.ServerExtremes:
			if err := db.readBlock(rd, block, true); err != nil {
				return nil, err
			}

			if res == nil {
				res = new(result)
			}
			res.affected += block.NumRow
		case chproto.ServerException:
			return nil, readException(rd)
		case chproto.ServerProgress:
			if err := readProgress(cn, rd); err != nil {
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
			block := new(chschema.Block)
			if err := db.readBlock(rd, block, false); err != nil {
				return nil, err
			}
		case chproto.ServerEndOfStream:
			return res, nil
		default:
			return nil, fmt.Errorf("ch: readDataBlocks: unexpected packet: %d", packet)
		}
	}
}

func readPacket(cn *chpool.Conn, rd *chproto.Reader) (*result, error) {
	packet, err := rd.Uvarint()
	if err != nil {
		return nil, err
	}

	res := new(result)
	switch packet {
	case chproto.ServerException:
		return nil, readException(rd)
	case chproto.ServerProgress:
		if err := readProgress(cn, rd); err != nil {
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
		return nil, fmt.Errorf("ch: readPacket: unexpected packet: %d", packet)
	}
}

func (db *DB) readBlock(rd *chproto.Reader, block *chschema.Block, compressible bool) error {
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

		block.NumColumn = int(numColumn)
		block.NumRow = int(numRow)

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

			col := block.Column(colName, colType)
			if err := col.ReadFrom(rd, int(numRow)); err != nil {
				return err
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

func readServerTableColumns(rd *chproto.Reader) error {
	_, err := rd.String()
	if err != nil {
		return err
	}
	_, err = rd.String()
	if err != nil {
		return err
	}
	return nil
}
