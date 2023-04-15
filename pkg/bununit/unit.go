package bununit

import (
	"fmt"
	"strings"
)

const (
	Invalid string = "_"

	None        string = ""
	Percents    string = "percents"
	Utilization string = "utilization"

	Nanoseconds  string = "nanoseconds"
	Microseconds string = "microseconds"
	Milliseconds string = "milliseconds"
	Seconds      string = "seconds"
	Duration     string = "duration"

	Bytes     string = "bytes"
	Kilobytes string = "kilobytes"
	Megabytes string = "megabytes"
	Gigabytes string = "gigabytes"
	Terabytes string = "terabytes"
)

func FromString(s string) string {
	s = strings.ToLower(s)
	return fromString(s)
}

func fromString(s string) string {
	switch s {
	case "", "1", "0":
		return None
	case Percents, "%":
		return Percents

	case Nanoseconds, "ns", "nanosecond":
		return Nanoseconds
	case Microseconds, "us", "microsecond":
		return Microseconds
	case Milliseconds, "ms", "millisecond":
		return Milliseconds
	case Seconds, "s", "sec", "second":
		return Seconds

	case Bytes, "by", "byte":
		return Bytes
	case Kilobytes, "kb", "kib", "kbyte":
		return Kilobytes
	case Megabytes, "mb", "mib", "mbyte":
		return Megabytes
	case Gigabytes, "gb", "gib", "gbyte":
		return Gigabytes
	case Terabytes, "tb", "tib", "tbyte":
		return Terabytes

	default:
		return s
	}
}

// ConvertValue converts the value between units.
func ConvertValue(n float64, from, to string) (float64, error) {
	from = FromString(from)
	to = FromString(to)

	switch to {
	case Bytes:
		switch from {
		case Bytes:
			return n, nil
		case Kilobytes:
			return n * 1024, nil
		case Megabytes:
			return n * 1024 * 1024, nil
		case Gigabytes:
			return n * 1024 * 1024 * 1024, nil
		case Terabytes:
			return n * 1024 * 1024 * 1024 * 1024, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case Nanoseconds:
		switch from {
		case Nanoseconds:
			return n, nil
		case Microseconds:
			return n * 1e3, nil
		case Milliseconds:
			return n * 1e6, nil
		case Seconds:
			return n * 1e9, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case Microseconds:
		switch from {
		case Nanoseconds:
			return n / 1e3, nil
		case Microseconds:
			return n, nil
		case Milliseconds:
			return n * 1e3, nil
		case Seconds:
			return n * 1e6, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case Milliseconds:
		switch from {
		case Nanoseconds:
			return n / 1e6, nil
		case Microseconds:
			return n / 1e3, nil
		case Milliseconds:
			return n, nil
		case Seconds:
			return n * 1e3, nil
		default:
			return 0, convertValueError(n, from, to)
		}

	case Seconds:
		switch from {
		case Nanoseconds:
			return n / 1e9, nil
		case Microseconds:
			return n / 1e6, nil
		case Milliseconds:
			return n / 1e3, nil
		case Seconds:
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
