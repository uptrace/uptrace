package ch

import (
	"bytes"
	"testing"

	"github.com/uptrace/pkg/clickhouse/ch/chpool"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
)

func TestReadInsertResultReadsUntilEndOfStream(t *testing.T) {
	var buf bytes.Buffer
	wr := chproto.NewWriter(&buf)

	wr.Uvarint(chproto.ServerTableColumns)
	wr.String("")
	wr.String("columns format version: 1\n")

	wr.Uvarint(chproto.ServerProgress)
	wr.Uvarint(10)  // rows
	wr.Uvarint(100) // bytes
	wr.Uvarint(0)   // total rows

	wr.Uvarint(chproto.ServerProfileInfo)
	wr.Uvarint(0)
	wr.Uvarint(0)
	wr.Uvarint(0)
	wr.Bool(false)
	wr.Uvarint(0)
	wr.Bool(false)

	wr.Uvarint(chproto.ServerProgress)
	wr.Uvarint(5)
	wr.Uvarint(50)
	wr.Uvarint(0)

	wr.Uvarint(chproto.ServerEndOfStream)

	// Bytes that belong to the next query on the same connection.
	wr.Uvarint(chproto.ServerPong)
	if err := wr.Flush(); err != nil {
		t.Fatal(err)
	}

	db := &DB{conf: new(Config)}
	cn := new(chpool.Conn)
	rd := chproto.NewReader(&buf)

	res, err := db.readInsertResult(cn, rd)
	if err != nil {
		t.Fatalf("readInsertResult: %v", err)
	}
	if got := res.Progress(); got.Rows != 15 || got.Bytes != 150 {
		t.Fatalf("progress = %+v, want rows=15 bytes=150", *got)
	}

	packet, err := rd.Uvarint()
	if err != nil {
		t.Fatalf("reading next packet: %v", err)
	}
	if packet != chproto.ServerPong {
		t.Fatalf("next packet = %d, want ServerPong (%d): EndOfStream was left unread", packet, chproto.ServerPong)
	}
}

func TestReadInsertResultException(t *testing.T) {
	var buf bytes.Buffer
	wr := chproto.NewWriter(&buf)

	wr.Uvarint(chproto.ServerProgress)
	wr.Uvarint(1)
	wr.Uvarint(1)
	wr.Uvarint(0)

	wr.Uvarint(chproto.ServerException)
	wr.Int32(241)
	wr.String("DB::Exception")
	wr.String("Memory limit exceeded")
	wr.String("")
	wr.Bool(false)
	if err := wr.Flush(); err != nil {
		t.Fatal(err)
	}

	db := &DB{conf: new(Config)}
	_, err := db.readInsertResult(new(chpool.Conn), chproto.NewReader(&buf))
	if err == nil {
		t.Fatal("expected an error")
	}
	if exc, ok := err.(*Error); !ok || exc.Code != 241 {
		t.Fatalf("err = %#v, want *Error with code 241", err)
	}
}
