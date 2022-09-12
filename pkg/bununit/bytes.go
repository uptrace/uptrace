package bununit

import (
	"fmt"
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
