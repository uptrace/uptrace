package chschema

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var enumMap sync.Map

type enumInfo struct {
	chType string
	dec    []string
	enc    map[string]int16
}

func (e *enumInfo) Encode(val string) (int16, bool) {
	i, ok := e.enc[val]
	return i, ok
}

func (e *enumInfo) Decode(i int16) string {
	return e.dec[i]
}

func parseEnum(s string) *enumInfo {
	if v, ok := enumMap.Load(s); ok {
		return v.(*enumInfo)
	}

	enumInfo, err := _parseEnum(s)
	if err != nil {
		panic(err)
	}
	enumInfo.chType = s

	enumMap.Store(s, enumInfo)
	return enumInfo
}

func _parseEnum(chType string) (*enumInfo, error) {
	s := enumType(chType)
	if s == "" {
		return nil, fmt.Errorf("can't parse enum type: %q", chType)
	}

	var dec []string
	for s != "" {
		var key, val string
		var ok bool

		s, key, ok = scanEnumKey(s)
		if !ok {
			return nil, fmt.Errorf("can't parse enum key: %q", s)
		}

		s, ok = scanEnumChar(s, '=')
		if !ok {
			return nil, fmt.Errorf("can't parse enum '=': %q", s)
		}

		s, val = scanEnumValue(s)
		if val == "" {
			return nil, fmt.Errorf("can't parse enum value: %q", s)
		}

		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, err
		}

		ln := int(n + 1)
		if len(dec) < ln {
			dec = append(dec, make([]string, ln-len(dec))...)
		}
		dec[n] = key

		s, _ = scanEnumChar(s, ',')
	}

	enc := make(map[string]int16, len(dec))
	for i, s := range dec {
		enc[s] = int16(i)
	}

	return &enumInfo{
		chType: chType,
		dec:    dec,
		enc:    enc,
	}, nil
}

func scanEnumKey(s string) (string, string, bool) {
loop:
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ':
			// ignore
		case '\'':
			s = s[i+1:]
			break loop
		default:
			return s, "", false
		}
	}

	i := strings.IndexByte(s, '\'')
	if i == -1 {
		return s, "", false
	}

	key := s[:i]
	s = s[i+1:]
	return s, key, true
}

func scanEnumChar(s string, ch byte) (string, bool) {
	var start int
loop:
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ':
			start = i + 1
		case ch:
			return s[i+1:], true
		default:
			break loop
		}
	}
	return s[start:], false
}

func scanEnumValue(s string) (string, string) {
	var start int
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ':
			start = i + 1
		case c >= '0' && c <= '9':
			// continue
		default:
			return s[i:], s[start:i]
		}
	}
	return "", s[start:]
}
