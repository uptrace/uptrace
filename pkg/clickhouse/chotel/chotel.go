package chotel

import (
	"context"
	"database/sql"
	"github.com/uptrace/pkg/clickhouse/ch"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"runtime"
	"strings"
)

var tracer = otel.Tracer("go-clickhouse")

type QueryHook struct{}

var _ ch.QueryHook = (*QueryHook)(nil)

func NewQueryHook() *QueryHook { return &QueryHook{} }
func (h *QueryHook) BeforeQuery(ctx context.Context, evt *ch.QueryEvent) (context.Context, error) {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil
	}
	ctx, _ = tracer.Start(ctx, "", trace.WithSpanKind(trace.SpanKindClient))
	return ctx, nil
}
func (h *QueryHook) AfterQuery(ctx context.Context, event *ch.QueryEvent) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	defer span.End()
	conf := event.DB.Config()
	operation := event.Operation()
	fn, file, line := funcFileLine("go-clickhouse")
	span.SetName(operation)
	attrs := make([]attribute.KeyValue, 0, 10)
	attrs = append(attrs, semconv.DBSystemKey.String("clickhouse"), semconv.DBNameKey.String(conf.Database), semconv.ServerAddressKey.String(conf.Addr), semconv.DBOperationKey.String(operation), semconv.DBStatementKey.String(event.Query), semconv.CodeFunctionKey.String(fn), semconv.CodeFilepathKey.String(file), semconv.CodeLineNumberKey.Int(line))
	if event.IQuery != nil {
		if tableName := event.IQuery.GetTableName(); tableName != "" {
			attrs = append(attrs, semconv.DBSQLTableKey.String(tableName))
		}
	}
	if event.Result != nil {
		numRow, err := event.Result.RowsAffected()
		if err == nil {
			attrs = append(attrs, attribute.Int64("db.rows_affected", numRow))
		}
	}
	span.SetAttributes(attrs...)
	switch event.Err {
	case nil, context.Canceled:
	case sql.ErrNoRows:
		span.RecordError(event.Err)
	default:
		span.SetStatus(codes.Error, "")
		span.RecordError(event.Err)
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
