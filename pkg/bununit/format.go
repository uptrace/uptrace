package bununit

import "time"

func Format(v float64, unit string) string {
	switch unit {
	case None:
		return FormatNumber(v)
	case Percents:
		return FormatPercents(v)

	case Nanoseconds:
		return FormatMicroseconds(v / 1000)
	case Microseconds:
		return FormatMicroseconds(v)
	case Milliseconds:
		return FormatMicroseconds(v * 1000)
	case Seconds:
		return FormatMicroseconds(v * 1e6)

	case Bytes:
		return FormatBytes(v)
	case Kilobytes:
		return FormatBytes(v * 1024)
	case Megabytes:
		return FormatBytes(v * 1024 * 1024)
	case Gigabytes:
		return FormatBytes(v * 1024 * 1024 * 1024)
	case Terabytes:
		return FormatBytes(v * 1024 * 1024 * 1024 * 1024)

	default:
		return FormatNumber(v) + " " + unit
	}
}

func FormatTime(tm time.Time) string {
	return tm.Format("02 Jan 2006 15:04:05")
}

func FormatDate(tm time.Time) string {
	return tm.Format("Jan 02 2006")
}
