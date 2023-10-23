package bunconv

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type byteUnit struct {
	suffix string
	bytes  int64
}

var byteUnits = []byteUnit{
	{"kb", 1 << 10},
	{"KB", 1 << 10},
	{"KiB", 1 << 10},

	{"mb", 1 << 20},
	{"MB", 1 << 20},
	{"MiB", 1 << 20},

	{"gb", 1 << 30},
	{"GB", 1 << 30},
	{"GiB", 1 << 30},

	{"tb", 1 << 40},
	{"TB", 1 << 40},
	{"TiB", 1 << 40},
}

func ParseBytes(s string) (int64, error) {
	for _, unit := range byteUnits {
		if !strings.HasSuffix(s, unit.suffix) {
			continue
		}

		s = strings.TrimSuffix(s, unit.suffix)

		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}

		return int64(float64(unit.bytes) * n), nil
	}

	return 0, fmt.Errorf("can't parse bytes: %q", s)
}

func FormatBytes(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) || n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 1024 {
		return format(n, 0)
	}

	for _, suffix := range []string{"KB", "MB", "GB", "TB", "PB"} {
		n /= 1024
		abs /= 1024

		if abs < 1 {
			return format(n, 2) + suffix
		}
		if abs < 10 {
			return format(n, 1) + suffix
		}
		if abs < 1000 {
			return format(n, 0) + suffix
		}
	}

	return format(n, 0) + "PB"
}
