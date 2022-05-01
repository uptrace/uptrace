package tracing

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type ClokiSample struct {
	ch.CHModel `ch:"table:samples_read_v2_2,alias:s"`

	String string         `json:"string"`
	Time   time.Time      `json:"time" ch:"type:DateTime64"`
	Labels map[string]any `json:"labels" ch:",json"`
}

func SelectClokiSamples(
	ctx context.Context, app *bunapp.App, f *ClokiFilter,
) ([]*ClokiSample, error) {
	s1 := f.TraceID                       // UUID v4
	s2 := strings.ReplaceAll(s1, "-", "") // UUID v4 without '-'

	samples := make([]*ClokiSample, 0)

	if err := app.ClokiDB().NewSelect().
		ColumnExpr("s.string").
		ColumnExpr("toDateTime64(s.timestamp_ns / 1e9, 6) AS time").
		ColumnExpr("ts.labels").
		Model(&samples).
		Join("LEFT JOIN time_series AS ts ON ts.fingerprint = s.fingerprint").
		Where("timestamp_ns >= ?", f.TimeGTE.UnixNano()).
		Where("timestamp_ns < ?", f.TimeLT.UnixNano()).
		Where("multiSearchAny(string, [?, ?])", s1, s2).
		Limit(10000).
		Scan(ctx); err != nil {
		return nil, err
	}

	return samples, nil
}
