package chschema

import (
	"fmt"
	"github.com/codemodus/kace"
	"github.com/jinzhu/inflection"
	"github.com/uptrace/pkg/clickhouse/ch/chtype"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"github.com/uptrace/pkg/tagparser"
	"github.com/uptrace/pkg/unsafeconv"
	"log/slog"
	"reflect"
	"slices"
)

const (
	discardUnknownColumnsFlag = internal.Flag(1) << iota
	columnarFlag
	afterScanRowHookFlag
)

var (
	chModelType        = reflect.TypeOf((*CHModel)(nil)).Elem()
	tableNameInflector = inflection.Plural
)

type CHModel struct{}

func SetTableNameInflector(fn func(string) string) { tableNameInflector = fn }

type Table struct {
	Type         reflect.Type
	ModelName    string
	Name         string
	CHName       string
	CHInsertName string
	CHAlias      string
	Engine       string
	Partition    string
	Fields       []*Field
	PKs          []*Field
	DataFields   []*Field
	FieldMap     map[string]*Field
	flags        internal.Flag
}

func newTable(typ reflect.Type, seen map[reflect.Type]*Table) *Table {
	t := new(Table)
	t.Type = typ
	t.ModelName = kace.Snake(t.Type.Name())
	tableName := tableNameInflector(t.ModelName)
	t.setName(tableName)
	t.CHAlias = t.ModelName
	t.processFields(typ, seen)
	typ = reflect.PtrTo(t.Type)
	if typ.Implements(afterScanRowHookType) {
		t.flags.Set(afterScanRowHookFlag)
	}
	return t
}
func (t *Table) String() string   { return "model=" + t.ModelName }
func (t *Table) IsColumnar() bool { return t.flags.Has(columnarFlag) }
func (t *Table) setName(name string) {
	t.Name = name
	t.CHName = name
	t.CHInsertName = name
	if t.CHAlias == "" {
		t.CHAlias = name
	}
}
func (t *Table) Field(name string) (*Field, error) {
	field, ok := t.FieldMap[name]
	if !ok {
		return nil, &UnknownColumnError{Table: t, Column: name}
	}
	return field, nil
}
func (t *Table) processFields(typ reflect.Type, seen map[reflect.Type]*Table) {
	type embeddedField struct {
		subtable *Table
		index    []int
		subfield *Field
	}
	names := make(map[string]struct{})
	embedded := make([]embeddedField, 0, 10)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		unexported := f.PkgPath != ""
		tagstr := f.Tag.Get("ch")
		if tagstr == "-" {
			names[f.Name] = struct{}{}
			continue
		}
		tag := tagparser.Parse(tagstr)
		if unexported && !f.Anonymous {
			continue
		}
		if !f.Anonymous {
			names[f.Name] = struct{}{}
			if field := t.newField(f, tag); field != nil {
				t.addField(field)
			}
			continue
		}
		if f.Name == "CHModel" && f.Type == chModelType {
			t.processCHModelField(f)
			continue
		}
		fieldType := indirectType(f.Type)
		if fieldType.Kind() != reflect.Struct {
			continue
		}
		subtable := newTable(fieldType, seen)
		for _, subfield := range subtable.Fields {
			embedded = append(embedded, embeddedField{index: f.Index, subfield: subfield})
		}
		if _, ok := tag.Options["extend"]; ok {
			t.Name = subtable.Name
			t.ModelName = subtable.ModelName
			t.CHName = subtable.CHName
			t.CHAlias = subtable.CHAlias
		}
	}
	ambiguousNames := make(map[string]int)
	ambiguousTags := make(map[string]int)
	for name := range names {
		ambiguousNames[name]++
		ambiguousTags[name]++
	}
	for _, f := range embedded {
		ambiguousNames[f.subfield.GoName]++
		if !f.subfield.Tag.IsZero() {
			ambiguousTags[f.subfield.GoName]++
		}
	}
	for _, embfield := range embedded {
		subfield := *embfield.subfield
		if ambiguousNames[subfield.GoName] > 1 && !(!subfield.Tag.IsZero() && ambiguousTags[subfield.GoName] == 1) {
			continue
		}
		subfield.Index = append(slices.Clone(embfield.index), subfield.Index...)
		t.addField(&subfield)
	}
	for _, f := range t.FieldMap {
		if t.IsColumnar() {
			f.Type = f.Type.Elem()
			if !f.hasFlag(customTypeFlag) {
				if s := chArrayElemType(f.CHType); s != "" {
					f.CHType = s
				}
			}
		}
		if f.NewColumn == nil {
			f.NewColumn = ColumnFactory(f.CHType, f.Type)
		}
	}
}
func (t *Table) processCHModelField(f reflect.StructField) {
	tag := tagparser.Parse(f.Tag.Get("ch"))
	if tag.Name != "" {
		t.setName(tag.Name)
	}
	if s, ok := tag.Option("table"); ok {
		t.setName(s)
	}
	if s, ok := tag.Option("alias"); ok {
		t.CHAlias = s
	}
	if s, ok := tag.Option("insert"); ok {
		t.CHInsertName = s
	}
	if s, ok := tag.Option("engine"); ok {
		t.Engine = s
	}
	if s, ok := tag.Option("partition"); ok {
		t.Partition = s
	}
	if tag.HasOption("columnar") {
		t.flags |= columnarFlag
	}
}
func (t *Table) newField(f reflect.StructField, tag tagparser.Tag) *Field {
	field := &Field{Field: f, Tag: tag, Type: f.Type, PtrType: ptrType(f.Type), GoName: f.Name, CHName: tag.Name, Index: f.Index}
	if field.CHName == "" {
		field.CHName = kace.Snake(field.GoName)
	}
	field.Column = quoteColumnName(field.CHName)
	field.NotNull = tag.HasOption("notnull")
	field.IsPK = tag.HasOption("pk")
	if s, ok := tag.Option("type"); ok {
		field.CHType = normEnumType(s)
		field.setFlag(customTypeFlag)
	} else {
		field.CHType = CHType(f.Type)
	}
	if tag.HasOption("lc") {
		if s := chSubType(field.CHType, "Array("); s != "" && s == chtype.String {
			field.CHType = "Array(LowCardinality(String))"
		} else if field.CHType == chtype.String {
			field.CHType = "LowCardinality(String)"
		} else {
			panic(fmt.Errorf("unsupported lc option on %s type", field.CHType))
		}
	}
	if tag.HasOption("nullable") {
		field.Nullable = true
	}
	if s, ok := tag.Option("default"); ok {
		field.CHDefault = Safe(s)
	}
	if _, ok := tag.Option("msgpack"); ok {
		field.NewColumn = NewMsgpackColumn
		field.CHType = "String"
		field.appendValue = msgpackAppender(f.Type)
	} else {
		field.appendValue = Appender(f.Type)
	}
	if tag.HasOption("scanonly") {
		t.FieldMap[field.CHName] = field
		return nil
	}
	return field
}
func (t *Table) addField(field *Field) {
	field.Offset = offsetForIndex(t.Type, field.Index)
	t.Fields = append(t.Fields, field)
	if field.IsPK {
		t.PKs = append(t.PKs, field)
	} else {
		t.DataFields = append(t.DataFields, field)
	}
	if t.FieldMap == nil {
		t.FieldMap = make(map[string]*Field)
	}
	t.FieldMap[field.CHName] = field
}
func (t *Table) NewColumn(colName, colType string) *Column {
	field, ok := t.FieldMap[colName]
	if !ok {
		slog.Error("unknown table column", slog.String("table", t.Name), slog.String("col_name", colName), slog.String("col_Table", colType))
		return &Column{Name: colName, Type: colType, Columnar: NewColumn(colType, nil)}
	}
	if colType != field.CHType {
		return &Column{Name: colName, Type: colType, Columnar: NewColumn(colType, field.Type)}
	}
	col := field.NewColumn()
	col.Init(field.CHType, field.Type)
	return &Column{Name: colName, Type: field.CHType, Columnar: col}
}
func (t *Table) HasAfterScanRowHook() bool { return t.flags.Has(afterScanRowHookFlag) }
func (t *Table) AppendNamedArg(fmter Formatter, b []byte, name string, strct reflect.Value) ([]byte, bool) {
	if field, ok := t.FieldMap[name]; ok {
		return field.AppendValue(fmter, b, strct), true
	}
	return b, false
}
func quoteColumnName(s string) Safe { return Safe(appendIdent(nil, unsafeconv.Bytes(s))) }
func ptrType(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return typ
	}
	return reflect.PtrTo(typ)
}
func normEnumType(chType string) string {
	name := chEnumType(chType)
	if name == "" {
		return chType
	}
	enum, ok := enumByNameMap[name]
	if !ok {
		return chType
	}
	return enum.chType
}

type UnknownColumnError struct {
	Table  *Table
	Column string
}

func (err *UnknownColumnError) Error() string {
	return fmt.Sprintf("ch: %s does not have column=%q", err.Table, err.Column)
}
