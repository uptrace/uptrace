package ch

import (
	"errors"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"reflect"
	"time"
)

var errNilModel = errors.New("ch: Model(nil)")
var (
	timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
	mapType  = reflect.TypeOf((*map[string]any)(nil)).Elem()
)

type Query interface {
	chschema.QueryAppender
	Operation() string
	GetTableName() string
	GetModel() Model
}

func newModel(db *DB, values ...any) (Model, error) {
	if len(values) > 1 {
		return scan(values...), nil
	}
	v0 := values[0]
	switch v0 := v0.(type) {
	case Model:
		return v0, nil
	}
	v := reflect.ValueOf(v0)
	if !v.IsValid() {
		return nil, errNilModel
	}
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("ch: Model(non-pointer %T)", v0)
	}
	v = v.Elem()
	switch v.Kind() {
	case reflect.Struct:
		if v.Type() != timeType {
			return newStructTableModelValue(db, v), nil
		}
	case reflect.Slice:
		typ := v.Type()
		elemType := indirectType(typ.Elem())
		if elemType == mapType {
			return newSliceMapModel(v), nil
		}
		if elemType.Kind() == reflect.Struct && elemType != timeType {
			return newSliceTableModel(db, v, elemType), nil
		}
	case reflect.Map:
		if v.Type() == mapType {
			return newMapModel(v), nil
		}
	}
	return scan(v0), nil
}
func scan(values ...any) Model { return &scanModel{values: values} }

type Model interface{ ScanBlock(block *Block) error }
type TableModel interface {
	Model
	Table() *chschema.Table
	WriteToBlock(block *Block, fields []*chschema.Field)
}

var (
	_ TableModel = (*sliceTableModel)(nil)
	_ TableModel = (*structTableModel)(nil)
)

func newTableModel(db *DB, value any) (TableModel, error) {
	if value, ok := value.(TableModel); ok {
		return value, nil
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil, errNilModel
	}
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("ch: Model(non-pointer %T)", value)
	}
	if v.IsNil() {
		typ := v.Type().Elem()
		if typ.Kind() == reflect.Struct {
			return newStructTableModel(db, chschema.TableForType(typ)), nil
		}
		return nil, errNilModel
	}
	v = v.Elem()
	switch v.Kind() {
	case reflect.Struct:
		return newStructTableModelValue(db, v), nil
	case reflect.Slice:
		elemType := sliceElemType(v)
		if elemType.Kind() == reflect.Struct {
			return newSliceTableModel(db, v, elemType), nil
		}
	}
	return nil, fmt.Errorf("ch: Model(unsupported %s)", v.Type())
}
