package unixtime

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTime(t *testing.T) {
	_, err := ParseTime("")
	require.Error(t, err)
	for _, s := range []string{"Mon Dec 05 19:11:00 2005", "Mon Dec 5 19:11:00 2005", "Mon Dec  5 19:11:00 2005", "Sat May 8 09:04:50 EDT 2004", "Mon Jan 02 15:04:05 -0700 2006"} {
		_, err := ParseTime(s)
		require.NoError(t, err)
		_, err = ParseTime(s + " foo")
		require.Error(t, err)
		_, err = ParseTime(" " + s)
		require.Error(t, err)
	}
	for _, s := range []string{"Jun 15 02:04:59", "Jun  5 02:04:59", "Jan 2 15:04:05.000000", "Jan 2 15:04:05.000000000"} {
		_, err := ParseTime(s)
		require.NoError(t, err)
		_, err = ParseTime(s + " foo")
		require.Error(t, err)
		_, err = ParseTime(" " + s)
		require.Error(t, err)
	}
	for _, s := range []string{"Jan-02 15:04:05.000", "03-17 16:13:39.006", "08 Sep 2022 07:17:48.087", "30/Jun/2022:14:30:59 +0000", "01/Jul/2022:10:03:03 +0000", "19/Jan/2023:11:08:59", "2023-01-11 13:42:34.504 EET", "2023-04-25 11:10:58.574", "2020-11-06 07:27:24.996372692 +0000 UTC", "2020-11-06 07:27:24.996372692 +0000 UTC m=+95.954853864", "2021-01-01 00:00:00+00:00", "2021-03-25T21:36:12Z", "2021-03-25T21:36:12.999999999Z", "2021-10-16T07:55:07.1234567-10:00", "2023/03/22 08:51:35"} {
		_, err := ParseTime(s)
		require.NoError(t, err)
		_, err = ParseTime(s + " foo")
		require.Error(t, err)
		_, err = ParseTime(" " + s)
		require.Error(t, err)
	}
}
