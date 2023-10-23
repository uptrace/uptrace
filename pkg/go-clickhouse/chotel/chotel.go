package chotel

import (
	"context"
	"database/sql"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/uptrace/go-clickhouse/ch"
)

var tracer = otel.Tracer("go-clickhouse")

type QueryHook struct{}

var _ ch.QueryHook = (*QueryHook)(nil)

func NewQueryHook() *QueryHook {
	return &QueryHook{}
}

func (h *QueryHook) BeforeQuery(
	ctx context.Context, evt *ch.QueryEvent,
) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	ctx, _ = tracer.Start(ctx, "", trace.WithSpanKind(trace.SpanKindClient))
	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *ch.QueryEvent) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	defer span.End()

	operation := event.Operation()
	fn, file, line := funcFileLine("go-clickhouse")
	span.SetName(operation)

	attrs := []attribute.KeyValue{
		semconv.CodeFunctionKey.String(fn),
		semconv.CodeFilepathKey.String(file),
		semconv.CodeLineNumberKey.Int(line),
		semconv.DBSystemKey.String("clickhouse"),
		semconv.DBOperationKey.String(operation),
		semconv.DBStatementKey.String(event.Query),
	}

	if event.IQuery != nil {
		if tableName := event.IQuery.GetTableName(); tableName != "" {
			attrs = append(attrs, semconv.DBSQLTableKey.String(tableName))
		}
	}

	span.SetAttributes(attrs...)

	switch event.Err {
	case nil, sql.ErrNoRows:
	default:
		span.SetStatus(codes.Error, "")
		span.RecordError(event.Err)
	}

	if event.Result != nil {
		numRow, err := event.Result.RowsAffected()
		if err == nil {
			span.SetAttributes(attribute.Int64("db.rows_affected", numRow))
		}
	}
}

func funcFileLine(pkg string) (string, string, int) {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	var fn, file string
	var line int
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		fn, file, line = f.Function, f.File, f.Line
		if !strings.Contains(fn, pkg) {
			break
		}
	}

	if ind := strings.LastIndexByte(fn, '/'); ind != -1 {
		fn = fn[ind+1:]
	}

	return fn, file, line
}
