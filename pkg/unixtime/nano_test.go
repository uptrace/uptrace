package unixtime

import (
	"github.com/stretchr/testify/require"
	"github.com/uptrace/pkg/msgp"
	"testing"
	"time"
)

func TestNow(t *testing.T) { require.InDelta(t, Now().UnixNano(), time.Now().UnixNano(), 5000) }
func TestMsgpack(t *testing.T) {
	for _, src := range []Nano{1, 0} {
		buf, err := src.AppendMsgpack(nil, 0)
		require.NoError(t, err)
		buf2, err := msgp.Append(nil, src, 0)
		require.NoError(t, err)
		require.Equal(t, buf, buf2)
		var dest Nano
		buf, err = dest.ParseMsgpack(buf, 0)
		require.NoError(t, err)
		require.Empty(t, buf)
		require.Equal(t, src, dest)
		var dest2 Nano
		err = msgp.Unmarshal(buf2, &dest2, 0)
		require.NoError(t, err)
		require.Empty(t, buf)
		require.Equal(t, dest, dest2)
	}
}
func TestCeilTime(t *testing.T) {
	type Test struct {
		in   string
		prec time.Duration
		out  string
	}
	tests := []Test{{in: "2021-08-06T09:01:00.000Z", prec: 5 * time.Minute, out: "2021-08-06T09:05:00Z"}, {in: "2021-08-06T09:00:00.000Z", prec: 3 * time.Hour, out: "2021-08-06T09:00:00Z"}, {in: "2021-08-06T09:56:00.000Z", prec: 3 * time.Hour, out: "2021-08-06T12:00:00Z"}}
	for _, test := range tests {
		tm, err := Parse(time.RFC3339, test.in)
		require.NoError(t, err)
		tm = tm.Ceil(test.prec)
		require.Equal(t, test.out, tm.Format(time.RFC3339))
	}
}
