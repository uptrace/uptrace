package uql_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/uptrace/uptrace/pkg/uql"
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
	}

	snapshotsDir := filepath.Join("testdata", "snapshots")
	snapshot := cupaloy.New(cupaloy.SnapshotSubdirectory(snapshotsDir))

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			expr, err := uql.ParsePart(test.in)
			snapshot.SnapshotT(t, map[string]any{
				"uql": expr,
				"err": fmt.Sprint(err),
			})
		})
	}
}
