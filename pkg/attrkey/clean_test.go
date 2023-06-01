package attrkey_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/uptrace/pkg/attrkey"
)

func TestClean(t *testing.T) {
	type Test struct {
		in     string
		wanted string
	}

	tests := []Test{
		{"foo", "foo"},
		{`"foo"`, "foo"},
		{`'foo'`, "foo"},
		{"foo_bar", "foo_bar"},
		{"foo_bar123", "foo_bar123"},
		{"FOO_bar", "foo_bar"},
		{"foo/bar", "foo.bar"},
		{"foo.bar", "foo.bar"},
		{"foo-bar", "foo_bar"},
		{"привет", ""},
		{"exception.param.<group>", "exception.param.group"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := attrkey.Clean(test.in)
			require.Equal(t, test.wanted, got, "in=%q", test.in)
		})
	}
}
