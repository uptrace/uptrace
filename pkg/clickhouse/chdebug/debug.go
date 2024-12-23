package chdebug

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fatih/color"
	"github.com/uptrace/pkg/clickhouse/ch"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Option func(*QueryHook)

func WithEnabled(on bool) Option    { return func(h *QueryHook) { h.enabled = on } }
func WithVerbose(on bool) Option    { return func(h *QueryHook) { h.verbose = on } }
func WithWriter(w io.Writer) Option { return func(h *QueryHook) { h.writer = w } }
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"CHDEBUG"}
	}
	return func(h *QueryHook) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				h.enabled = env != "" && env != "0"
				h.verbose = env == "2"
				break
			}
		}
	}
}

type QueryHook struct {
	enabled bool
	verbose bool
	writer  io.Writer
}

var _ ch.QueryHook = (*QueryHook)(nil)

func NewQueryHook(opts ...Option) *QueryHook {
	h := &QueryHook{enabled: true, writer: os.Stderr}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
func (h *QueryHook) BeforeQuery(ctx context.Context, evt *ch.QueryEvent) (context.Context, error) {
	return ctx, nil
}
func (h *QueryHook) AfterQuery(ctx context.Context, event *ch.QueryEvent) {
	if !h.enabled {
		return
	}
	if !h.verbose {
		switch event.Err {
		case nil, sql.ErrNoRows:
			return
		}
	}
	now := time.Now()
	dur := now.Sub(event.StartTime)
	args := []any{"[ch]", now.Format(" 15:04:05.000 "), formatOperation(event), formatFileLine(), fmt.Sprintf(" %10s ", dur.Round(time.Microsecond)), event.Query}
	if event.Err != nil {
		typ := reflect.TypeOf(event.Err).String()
		args = append(args, "\t", color.New(color.BgRed).Sprintf(" %s ", typ+": "+event.Err.Error()))
	}
	fmt.Fprintln(h.writer, args...)
}
func formatOperation(event *ch.QueryEvent) string {
	operation := event.Operation()
	return operationColor(operation).Sprintf(" %-16s ", operation)
}
func formatFileLine() string {
	_, file, line := funcFileLine("pkg/clickhouse")
	return fmt.Sprintf("%s:%d", file, line)
}
func operationColor(operation string) *color.Color {
	switch operation {
	case "SELECT":
		return color.New(color.BgGreen, color.FgHiWhite)
	case "INSERT":
		return color.New(color.BgBlue, color.FgHiWhite)
	case "UPDATE":
		return color.New(color.BgYellow, color.FgHiBlack)
	case "DELETE":
		return color.New(color.BgMagenta, color.FgHiWhite)
	default:
		return color.New(color.BgWhite, color.FgHiBlack)
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
