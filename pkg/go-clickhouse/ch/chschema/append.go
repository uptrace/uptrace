package chschema

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"
)

func Append(fmter Formatter, b []byte, v any) []byte {
	switch v := v.(type) {
	case nil:
		return AppendNull(b)
	case bool:
		return AppendBool(b, v)
	case int8:
		return strconv.AppendInt(b, int64(v), 10)
	case int16:
		return strconv.AppendInt(b, int64(v), 10)
	case int32:
		return strconv.AppendInt(b, int64(v), 10)
	case int64:
		return strconv.AppendInt(b, v, 10)
	case int:
		return strconv.AppendInt(b, int64(v), 10)
	case uint8:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(b, v, 10)
	case uint:
		return strconv.AppendUint(b, uint64(v), 10)
	case float32:
		return appendFloat(b, float64(v), 32)
	case float64:
		return appendFloat(b, v, 64)
	case string:
		return AppendString(b, v)
	case time.Time:
		return AppendTime(b, v)
	case []byte:
		return AppendBytes(b, v)
	case QueryAppender:
		return AppendQueryAppender(fmter, b, v)
	case driver.Valuer:
		return appendDriverValue(fmter, b, v)
	default:
		return AppendError(b, fmt.Errorf("ch: can't append %T", v))
	}
}

func AppendError(b []byte, err error) []byte {
	b = append(b, "?!("...)
	b = append(b, err.Error()...)
	b = append(b, ')')
	return b
}

func AppendNull(b []byte) []byte {
	return append(b, "NULL"...)
}

func AppendBool(dst []byte, v bool) []byte {
	var c byte
	if v {
		c = 1
	}
	return append(dst, c)
}

func AppendFloat(dst []byte, v float64) []byte {
	return appendFloat(dst, v, 64)
}

func appendFloat(dst []byte, v float64, bitSize int) []byte {
	switch {
	case math.IsNaN(v):
		return append(dst, "nan"...)
	case math.IsInf(v, 1):
		return append(dst, "inf"...)
	case math.IsInf(v, -1):
		return append(dst, "-inf"...)
	default:
		return strconv.AppendFloat(dst, v, 'f', -1, bitSize)
	}
}

func AppendString(b []byte, s string) []byte {
	b = append(b, '\'')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\'':
			b = append(b, '\\', '\'')
		case '\\':
			b = append(b, '\\', '\\')
		default:
			b = append(b, c)
		}
	}
	b = append(b, '\'')
	return b
}

func AppendTime(b []byte, tm time.Time) []byte {
	b = append(b, "toDateTime('"...)
	b = tm.UTC().AppendFormat(b, "2006-01-02 15:04:05")
	b = append(b, "', 'UTC')"...)
	return b
}

func AppendBytes(b []byte, bytes []byte) []byte {
	if bytes == nil {
		return AppendNull(b)
	}

	tmp := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(tmp, bytes)

	b = append(b, "unhex('"...)
	b = append(b, tmp...)
	b = append(b, "')"...)

	return b
}
