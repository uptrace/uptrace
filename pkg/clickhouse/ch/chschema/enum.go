package chschema

import (
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"strconv"
	"strings"
)

var enumByNameMap map[string]*EnumInfo

func LookupEnum(chType string) (*EnumInfo, error) {
	name := chEnumType(chType)
	if name == "" {
		return nil, fmt.Errorf("can't parse enum type: %q", chType)
	}
	if enum, ok := enumByNameMap[name]; ok {
		return enum, nil
	}
	return ParseEnum(chType)
}
func RegisterEnum(name string, dec []string) {
	enc := make(map[string]int16, len(dec))
	for i, s := range dec {
		enc[s] = int16(i)
	}
	registerEnum(name, enc, dec)
}
func registerEnum(name string, enc map[string]int16, dec []string) {
	ss := make([]string, len(dec))
	for k, v := range dec {
		ss[k] = fmt.Sprintf("'%s' = %d", v, k)
	}
	chType := enum8Type(ss)
	enum := &EnumInfo{chType: chType, enc: enc, dec: dec}
	if enumByNameMap == nil {
		enumByNameMap = make(map[string]*EnumInfo)
	}
	if _, ok := enumByNameMap[name]; ok {
		panic(fmt.Errorf("enum %s is already registered", name))
	}
	enumByNameMap[name] = enum
}
func enum8Type(values []string) string {
	enums := strings.Join(values, ", ")
	return "Enum8(" + enums + ")"
}

var enumByTypeMap = xsync.NewMapOf[string, *EnumInfo]()

type EnumInfo struct {
	chType string
	dec    []string
	enc    map[string]int16
}

func (e *EnumInfo) Encode(val string) int16 {
	if n, ok := e.enc[val]; ok {
		return n
	}
	return e.enc[""]
}
func (e *EnumInfo) Decode(i int16) string { return e.dec[i] }
func ParseEnum(chType string) (*EnumInfo, error) {
	if chType == "" {
		return nil, errors.New("enum type can't be empty")
	}
	if v, ok := enumByTypeMap.Load(chType); ok {
		return v, nil
	}
	enum, err := _parseEnum(chType)
	if err != nil {
		return nil, err
	}
	enum.chType = chType
	enumByTypeMap.Store(chType, enum)
	return enum, nil
}
func _parseEnum(chType string) (*EnumInfo, error) {
	s := chEnumType(chType)
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
	return &EnumInfo{chType: chType, dec: dec, enc: enc}, nil
}
func scanEnumKey(s string) (string, string, bool) {
loop:
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ':
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
		default:
			return s[i:], s[start:i]
		}
	}
	return "", s[start:]
}
