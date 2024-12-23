package urlstruct

import (
	"context"
	"net/url"
	"reflect"
)

func Unmarshal(ctx context.Context, values url.Values, strct any) error {
	d := newStructDecoder(reflect.ValueOf(strct))
	return d.Decode(ctx, values)
}
