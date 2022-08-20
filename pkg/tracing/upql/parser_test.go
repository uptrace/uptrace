package upql_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/uptrace/uptrace/pkg/tracing/upql"
)

func TestParse(t *testing.T) {
	type Test struct {
		in string
	}

	tests := []Test{
		{in: ``},
		{in: `where a.foo = 'bar'`},
		{in: `avg(a)`},
		{in: "where foo"},
		{in: "where not foo"},
		{in: "where span.duration > 100ms"},
		{in: `{p50,p90,p99}(span.duration)`},
		{in: `where {foo,bar} contains something`},
		{in: "where span.duration>100ms"},
		{in: "where span.duration >= 100ms"},
	}

	snapshotsDir := filepath.Join("testdata", "snapshots")
	snapshot := cupaloy.New(cupaloy.SnapshotSubdirectory(snapshotsDir))

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			expr, err := upql.ParsePart(test.in)
			snapshot.SnapshotT(t, map[string]any{
				"upql": expr,
				"err":  fmt.Sprint(err),
			})
		})
	}
}
