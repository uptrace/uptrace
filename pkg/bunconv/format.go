package bunconv

import "time"

func Format(v float64, unit string) string {
	switch unit {
	case UnitNone:
		return FormatFloat(v)
	case UnitPercents:
		return FormatPercents(v)
	case UnitUtilization:
		return FormatUtilization(v)

	case UnitNanoseconds:
		return FormatMicroseconds(v / 1000)
	case UnitMicroseconds:
		return FormatMicroseconds(v)
	case UnitMilliseconds:
		return FormatMicroseconds(v * 1000)
	case UnitSeconds:
		return FormatMicroseconds(v * 1e6)

	case UnitBytes:
		return FormatBytes(v)
	case UnitKilobytes:
		return FormatBytes(v * 1024)
	case UnitMegabytes:
		return FormatBytes(v * 1024 * 1024)
	case UnitGigabytes:
		return FormatBytes(v * 1024 * 1024 * 1024)
	case UnitTerabytes:
		return FormatBytes(v * 1024 * 1024 * 1024 * 1024)

	default:
		return FormatFloat(v) + " " + unit
	}
}

func FormatTime(tm time.Time) string {
	return tm.Format("02 Jan 2006 15:04:05")
}

func FormatDate(tm time.Time) string {
	return tm.Format("Jan 02 2006")
}
