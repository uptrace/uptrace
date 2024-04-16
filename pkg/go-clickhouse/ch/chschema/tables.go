package chschema

import (
	"fmt"
	"reflect"
	"sync"
)

var globalTables = newTablesMap()

func TableForType(typ reflect.Type) *Table {
	return globalTables.Get(typ)
}

type tablesMap struct {
	m sync.Map
}

func newTablesMap() *tablesMap {
	return new(tablesMap)
}

func (t *tablesMap) Get(typ reflect.Type) *Table {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("got %s, wanted %s", typ.Kind(), reflect.Struct))
	}

	if v, ok := t.m.Load(typ); ok {
		return v.(*Table)
	}

	table := newTable(typ, make(map[reflect.Type]*Table))
	if v, loaded := t.m.LoadOrStore(typ, table); loaded {
		return v.(*Table)
	}

	return table
}

func (t *tablesMap) getByName(name string) *Table {
	var found *Table
	t.m.Range(func(key, value any) bool {
		t := value.(*Table)
		if t.Name == name || t.ModelName == name {
			found = t
			return false
		}
		return true
	})
	return found
}
