package bununit

import "strings"

const (
	Invalid string = "_"

	None     string = ""
	Percents string = "percents"

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
	switch strings.ToLower(s) {
	case "", "1":
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
		return string(s)
	}
}
