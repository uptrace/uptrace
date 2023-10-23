package bunconv

import (
	"fmt"
	"strings"
)

const (
	UnitInvalid string = "_"

	UnitNone        string = ""
	UnitPercents    string = "percents"
	UnitUtilization string = "utilization"

	UnitNanoseconds  string = "nanoseconds"
	UnitMicroseconds string = "microseconds"
	UnitMilliseconds string = "milliseconds"
	UnitSeconds      string = "seconds"
	UnitDuration     string = "duration"

	UnitBytes     string = "bytes"
	UnitKilobytes string = "kilobytes"
	UnitMegabytes string = "megabytes"
	UnitGigabytes string = "gigabytes"
	UnitTerabytes string = "terabytes"

	UnitTime string = "time"
)

func NormUnit(s string) string {
	switch s {
	case "", "1", "0", "none", "None":
		return UnitNone
	}

	if norm := fromString(s); norm != "" {
		return norm
	}

	s = strings.ToLower(s)

	if norm := fromString(s); norm != "" {
		return norm
	}

	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return s
	}
	return "{" + s + "}"
}

func fromString(s string) string {
	switch s {
	case "percents", "%", "Percent", "percent":
		return UnitPercents
	case "utilization":
		return UnitUtilization

	case "nanoseconds", "ns", "nanosecond":
		return UnitNanoseconds
	case "microseconds", "us", "microsecond":
		return UnitMicroseconds
	case "milliseconds", "ms", "millisecond":
		return UnitMilliseconds
	case "seconds", "s", "sec", "second":
		return UnitSeconds

	case "bytes", "Bytes", "by", "byte", "Byte":
		return UnitBytes
	case "kilobytes", "kb", "kib", "kbyte":
		return UnitKilobytes
	case "megabytes", "mb", "mib", "mbyte":
		return UnitMegabytes
	case "gigabytes", "gb", "gib", "gbyte":
		return UnitGigabytes
	case "terabytes", "tb", "tib", "tbyte":
		return UnitTerabytes

	case "count", "Count":
		return "{count}"

	default:
		return ""
	}
}

func ConvertValue(n float64, from, to string) (float64, error) {
	from = NormUnit(from)
	to = NormUnit(to)

	switch to {
	case UnitBytes:
		switch from {
		case UnitBytes:
			return n, nil
		case UnitKilobytes:
			return n * 1024, nil
		case UnitMegabytes:
			return n * 1024 * 1024, nil
		case UnitGigabytes:
			return n * 1024 * 1024 * 1024, nil
		case UnitTerabytes:
			return n * 1024 * 1024 * 1024 * 1024, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case UnitNanoseconds:
		switch from {
		case UnitNanoseconds:
			return n, nil
		case UnitMicroseconds:
			return n * 1e3, nil
		case UnitMilliseconds:
			return n * 1e6, nil
		case UnitSeconds:
			return n * 1e9, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case UnitMicroseconds:
		switch from {
		case UnitNanoseconds:
			return n / 1e3, nil
		case UnitMicroseconds:
			return n, nil
		case UnitMilliseconds:
			return n * 1e3, nil
		case UnitSeconds:
			return n * 1e6, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case UnitMilliseconds:
		switch from {
		case UnitNanoseconds:
			return n / 1e6, nil
		case UnitMicroseconds:
			return n / 1e3, nil
		case UnitMilliseconds:
			return n, nil
		case UnitSeconds:
			return n * 1e3, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case UnitSeconds:
		switch from {
		case UnitNanoseconds:
			return n / 1e9, nil
		case UnitMicroseconds:
			return n / 1e6, nil
		case UnitMilliseconds:
			return n / 1e3, nil
		case UnitSeconds:
			return n, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	default:
		return 0, convertValueError(n, from, to)
	}
}

func convertValueError(n float64, from, to string) error {
	return fmt.Errorf("can't convert %g from %q to %q", n, from, to)
}
