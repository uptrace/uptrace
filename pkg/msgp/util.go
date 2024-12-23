package msgp

import "github.com/uptrace/pkg/msgp/msgpcode"

func ParseIntOrBytes(b []byte) (int, []byte, []byte, error) {
	if num, b, ok := parseInt(b); ok {
		return num, nil, b, nil
	}
	bs, b, err := ParseBytes(b, ZeroCopyBytes)
	if err != nil {
		return 0, nil, nil, err
	}
	return 0, bs, b, nil
}
func parseInt(b []byte) (int, []byte, bool) {
	c, b, err := readByte(b)
	if err != nil {
		return 0, nil, false
	}
	if msgpcode.IsFixedNum(c) {
		return int(int8(c)), b, true
	}
	switch c {
	case msgpcode.Uint8:
		n, b, err := readByte(b)
		if err != nil {
			return 0, nil, false
		}
		return int(n), b, true
	case msgpcode.Uint16:
		n, b, err := parseUint16(b)
		if err != nil {
			return 0, nil, false
		}
		return int(n), b, true
	default:
		return 0, nil, false
	}
}
