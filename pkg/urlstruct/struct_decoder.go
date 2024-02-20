package urlstruct

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

const endOfValues = "\000"

type structDecoder struct {
	v     reflect.Value
	sinfo *StructInfo

	decMap           map[string]*structDecoder
	paramUnmarshaler ParamUnmarshaler
}

func newStructDecoder(v reflect.Value) *structDecoder {
	v = reflect.Indirect(v)
	return &structDecoder{
		v:     v,
		sinfo: DescribeStruct(v.Type()),
	}
}

func (d *structDecoder) Decode(ctx context.Context, values url.Values) error {
	var maps map[string][]string

	for name, values := range values {
		name = strings.TrimPrefix(name, ":")
		name = strings.TrimSuffix(name, "[]")

		if name, key, ok := mapKey(name); ok {
			if maps == nil {
				maps = make(map[string][]string)
			}
			mapValues := maps[name]
			if mapValues == nil {
				mapValues = make([]string, 0, 1+len(values)+1)
			}
			mapValues = append(mapValues, key)
			mapValues = append(mapValues, values...)
			mapValues = append(mapValues, endOfValues)
			maps[name] = mapValues
			continue
		}

		if err := d.decodeParam(ctx, name, values); err != nil {
			return err
		}
	}

	for name, kvs := range maps {
		if err := d.decodeParam(ctx, name, kvs); err != nil {
			return nil
		}
	}

	if d.sinfo.isValuesUnmarshaler {
		return d.v.Addr().Interface().(ValuesUnmarshaler).UnmarshalValues(ctx, values)
	}

	return nil
}

func (d *structDecoder) decodeParam(ctx context.Context, name string, values []string) error {
	if err := d._decodeParam(ctx, name, values); err != nil {
		return fmt.Errorf("urlstruct: can't decode %q: %w", name, err)
	}
	return nil
}

func (d *structDecoder) _decodeParam(ctx context.Context, name string, values []string) error {
	if field := d.sinfo.Field(name); field != nil && !field.noDecode {
		return field.scanValue(field.Value(d.v), values)
	}

	if d.sinfo.isParamUnmarshaler {
		if d.paramUnmarshaler == nil {
			d.paramUnmarshaler = d.v.Addr().Interface().(ParamUnmarshaler)
		}
		return d.paramUnmarshaler.UnmarshalParam(ctx, name, values)
	}

	return nil
}

func mapKey(s string) (name string, key string, ok bool) {
	ind := strings.IndexByte(s, '[')
	if ind == -1 || s[len(s)-1] != ']' {
		return "", "", false
	}
	key = s[ind+1 : len(s)-1]
	if key == "" {
		return "", "", false
	}
	name = s[:ind]
	return name, key, true
}
