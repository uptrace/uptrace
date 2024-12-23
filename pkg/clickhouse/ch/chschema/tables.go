package chschema

import (
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"reflect"
)

var globalTables = newTablesMap()

func TableForType(typ reflect.Type) *Table { return globalTables.Get(typ) }

type tablesMap struct {
	m *xsync.MapOf[reflect.Type, *Table]
}

func newTablesMap() *tablesMap { return &tablesMap{m: xsync.NewMapOf[reflect.Type, *Table]()} }
func (t *tablesMap) Get(typ reflect.Type) *Table {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("got %s, wanted %s", typ.Kind(), reflect.Struct))
	}
	if v, ok := t.m.Load(typ); ok {
		return v
	}
	table := newTable(typ, make(map[reflect.Type]*Table))
	if v, loaded := t.m.LoadOrStore(typ, table); loaded {
		return v
	}
	return table
}
func (t *tablesMap) getByName(name string) *Table {
	var found *Table
	t.m.Range(func(key reflect.Type, table *Table) bool {
		if table.Name == name || table.ModelName == name {
			found = table
			return false
		}
		return true
	})
	return found
}
