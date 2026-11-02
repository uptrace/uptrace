package chschema

import (
	"reflect"
	"testing"
	"time"

	"github.com/uptrace/pkg/unixtime"
)

func TestColumnFactoryDateTimeArrays(t *testing.T) {
	tests := []struct {
		chType string
		want   reflect.Type
		mul    int64 // DateTime64 multiplier to nanoseconds, 0 when not applicable
	}{
		// Previous behaviour: timezone-less element types.
		{"Array(DateTime)", reflect.TypeFor[[]unixtime.Nano](), 0},
		{"Array(DateTime64)", reflect.TypeFor[[]int64](), 1},
		{"Array(Array(DateTime))", reflect.TypeFor[[][]time.Time](), 0},
		// ClickHouse >= 26 keeps the source timezone in CAST results; explicit
		// DateTime64 precision was unsupported in arrays before as well.
		{"Array(DateTime64(3))", reflect.TypeFor[[]int64](), 1e6},
		{"Array(DateTime('UTC'))", reflect.TypeFor[[]unixtime.Nano](), 0},
		{"Array(DateTime64(6, 'UTC'))", reflect.TypeFor[[]int64](), 1e3},
		{"Array(DateTime64(9, 'UTC'))", reflect.TypeFor[[]int64](), 1},
		{"Array(Array(DateTime('UTC')))", reflect.TypeFor[[][]time.Time](), 0},
		{"SimpleAggregateFunction(max, Array(DateTime('UTC')))", reflect.TypeFor[[]unixtime.Nano](), 0},
	}
	for _, tt := range tests {
		t.Run(tt.chType, func(t *testing.T) {
			factory := ColumnFactory(tt.chType, nil)
			if factory == nil {
				t.Fatalf("no column factory for %s", tt.chType)
			}
			col := factory()
			if err := col.Init(tt.chType, nil); err != nil {
				t.Fatalf("Init(%s): %v", tt.chType, err)
			}
			if got := col.Type(); got != tt.want {
				t.Fatalf("Type() = %s, want %s", got, tt.want)
			}
			if tt.mul != 0 {
				if c, ok := col.(*ArrayDateTime64Column); !ok {
					t.Fatalf("got %T, want *ArrayDateTime64Column", col)
				} else if c.elem.Mul != tt.mul {
					t.Fatalf("DateTime64 multiplier = %d, want %d", c.elem.Mul, tt.mul)
				}
			}
		})
	}
}

func TestColumnFactoryDateTimeArraysMatchPlainTypes(t *testing.T) {
	// A timezone-tagged array must produce the same column implementation as the
	// plain type did on older ClickHouse versions.
	pairs := [][2]string{
		{"Array(DateTime)", "Array(DateTime('UTC'))"},
		{"Array(DateTime64)", "Array(DateTime64(9, 'UTC'))"},
		{"Array(Array(DateTime))", "Array(Array(DateTime('UTC')))"},
	}
	for _, p := range pairs {
		plain := reflect.TypeOf(ColumnFactory(p[0], nil)())
		tagged := reflect.TypeOf(ColumnFactory(p[1], nil)())
		if plain != tagged {
			t.Fatalf("%s -> %s but %s -> %s", p[0], plain, p[1], tagged)
		}
	}
}
