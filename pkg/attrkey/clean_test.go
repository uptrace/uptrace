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

func TestAWSMetricName(t *testing.T) {
	type Test struct {
		namespace string
		metric    string
		wanted    string
	}

	tests := []Test{
		{"AWS/SES", "Delivery", "amazonaws.com.aws.ses.delivery"},
		{"AWS/EBS", "VolumeTotalWriteTime", "amazonaws.com.aws.ebs.volume_total_write_time"},
		{"AWS", "XXX", "amazonaws.com.aws.xxx"},
		{"AWS", "DBName", "amazonaws.com.aws.db_name"},
		{"AWS", "XXXX1", "amazonaws.com.aws.xxxx1"},
		{"AWS", "X1XXX", "amazonaws.com.aws.x1xxx"},
		{"AWS", "XXXX1Foo", "amazonaws.com.aws.xxxx1_foo"},
		{"AWS", "XXXX123Foo", "amazonaws.com.aws.xxxx123_foo"},
		{"AWS", "123foo", "amazonaws.com.aws.123foo"},
		{"AWS", "123Foo", "amazonaws.com.aws.123_foo"},
		{"AWS", "StatusCheckFailed_System", "amazonaws.com.aws.status_check_failed_system"},
	}
	for _, test := range tests {
		got := attrkey.AWSMetricName(test.namespace, test.metric)
		require.Equal(t, test.wanted, got)
	}
}
