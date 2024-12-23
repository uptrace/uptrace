package msgp

import (
	"reflect"
)

type UnsupportedTypeError struct{ Type reflect.Type }

func (e *UnsupportedTypeError) Error() string { return "msgp: unsupported type: " + e.Type.String() }
