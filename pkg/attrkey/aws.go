package attrkey

import "github.com/uptrace/pkg/unsafeconv"

func AWSMetricName(namespace, metric string) string {
	const prefix = "cloudwatch_"
	b := make([]byte, 0, len(prefix)+len(namespace)+len(metric)+5)
	b = append(b, prefix...)
	b = underscore(b, namespace)
	b = append(b, '_')
	b = underscore(b, metric)
	return unsafeconv.String(b)
}
