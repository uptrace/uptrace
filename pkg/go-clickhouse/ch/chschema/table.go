package chschema

import (
	"fmt"
	"reflect"

	"github.com/codemodus/kace"
	"github.com/jinzhu/inflection"

	"github.com/uptrace/go-clickhouse/ch/chtype"
	"github.com/uptrace/go-clickhouse/ch/internal"
	"github.com/uptrace/go-clickhouse/ch/internal/tagparser"
)

const (
	columnarFlag = internal.Flag(1) << iota
	afterScanBlockHookFlag
)

var (
	chModelType        = reflect.TypeOf((*CHModel)(nil)).Elem()
	tableNameInflector = inflection.Plural
)

type CHModel struct{}

// SetTableNameInflector overrides the default func that pluralizes
// model name to get table name, e.g. my_article becomes my_articles.
func SetTableNameInflector(fn func(string) string) {
	tableNameInflector = fn
}

type Table struct {
	Type reflect.Type

	ModelName string

	Name         string
	CHName       Safe
	CHInsertName Safe
	CHAlias      Safe
	CHEngine     string
	CHPartition  string

	Fields     []*Field // PKs + DataFields
	PKs        []*Field
	DataFields []*Field
	FieldMap   map[string]*Field

	flags internal.Flag
}

func newTable(typ reflect.Type) *Table {
	t := new(Table)
	t.Type = typ
	t.ModelName = kace.Snake(t.Type.Name())
	tableName := tableNameInflector(t.ModelName)
	t.setName(tableName)
	t.CHAlias = quoteColumnName(t.ModelName)
	t.initFields()

	typ = reflect.PtrTo(t.Type)
	if typ.Implements(afterScanBlockHookType) {
		t.flags.Set(afterScanBlockHookFlag)
	}

	return t
}

func (t *Table) String() string {
	return "model=" + t.ModelName
}

func (t *Table) IsColumnar() bool {
	return t.flags.Has(columnarFlag)
}

func (t *Table) setName(name string) {
	quoted := quoteTableName(name)
	t.Name = name
	t.CHName = quoted
	t.CHInsertName = quoted
	if t.CHAlias == "" {
		t.CHAlias = quoted
	}
}

func (t *Table) Field(name string) (*Field, error) {
	field, ok := t.FieldMap[name]
	if !ok {
		return nil, &UnknownColumnError{
			Table:  t,
			Column: name,
		}
	}
	return field, nil
}

func (t *Table) initFields() {
	t.Fields = make([]*Field, 0, t.Type.NumField())
	t.FieldMap = make(map[string]*Field, t.Type.NumField())
	t.addFields(t.Type, nil)
}

func (t *Table) addFields(typ reflect.Type, baseIndex []int) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		tag := tagparser.Parse(f.Tag.Get("ch"))
		if tag.Name == "-" {
			continue
		}

		// Make a copy so slice is not shared between fields.
		index := make([]int, len(baseIndex))
		copy(index, baseIndex)

		if f.Anonymous {
			if f.Name == "CHModel" && f.Type == chModelType {
				if len(index) == 0 {
					t.processCHModelField(f)
				}
				continue
			}

			fieldType := indirectType(f.Type)
			if fieldType.Kind() != reflect.Struct {
				continue
			}
			t.addFields(fieldType, append(index, f.Index...))

			if _, ok := tag.Options["inherit"]; ok {
				embeddedTable := globalTables.Get(fieldType)
				t.ModelName = embeddedTable.ModelName
				t.CHName = embeddedTable.CHName
				t.CHAlias = embeddedTable.CHAlias
			}

			continue
		}

		if field := t.newField(f, index, tag); field != nil {
			t.addField(field)
		}
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
		t.CHAlias = quoteColumnName(s)
	}
	if s, ok := tag.Option("insert"); ok {
		t.CHInsertName = quoteTableName(s)
	}
	if s, ok := tag.Option("engine"); ok {
		t.CHEngine = s
	}
	if s, ok := tag.Option("partition"); ok {
		t.CHPartition = s
	}
	if tag.HasOption("columnar") {
		t.flags |= columnarFlag
	}
}

func (t *Table) newField(f reflect.StructField, index []int, tag tagparser.Tag) *Field {
	if f.PkgPath != "" {
		return nil
	}

	if tag.Name == "" {
		tag.Name = kace.Snake(f.Name)
	}

	field := &Field{
		Field: f,
		Type:  f.Type,

		GoName: f.Name,
		CHName: tag.Name,
		Column: quoteColumnName(tag.Name),

		Index: append(index, f.Index...),
	}
	field.NotNull = tag.HasOption("notnull")
	field.IsPK = tag.HasOption("pk")

	if s, ok := tag.Option("type"); ok {
		field.CHType = s
		field.setFlag(customTypeFlag)
	} else {
		field.CHType = chType(f.Type)
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

	if s, ok := tag.Option("default"); ok {
		field.CHDefault = Safe(s)
	}
	field.appendValue = Appender(f.Type)

	if s, ok := tag.Option("alt"); ok {
		t.FieldMap[s] = field
	}

	if tag.HasOption("scanonly") {
		t.FieldMap[field.CHName] = field
		return nil
	}

	return field
}

func (t *Table) addField(field *Field) {
	t.Fields = append(t.Fields, field)
	if field.IsPK {
		t.PKs = append(t.PKs, field)
	} else {
		t.DataFields = append(t.DataFields, field)
	}
	t.FieldMap[field.CHName] = field
}

func (t *Table) NewColumn(colName, colType string) *Column {
	field, ok := t.FieldMap[colName]
	if !ok {
		internal.Logger.Printf("ch: %s has no column=%q", t, colName)
		return nil
	}

	if colType != field.CHType {
		return &Column{
			Name:     colName,
			Type:     colType,
			Columnar: NewColumn(colType, field.Type),
		}
	}

	col := field.NewColumn()
	col.Init(field.CHType)

	return &Column{
		Name:     colName,
		Type:     field.CHType,
		Columnar: col,
	}
}

func (t *Table) HasAfterScanRowHook() bool { return t.flags.Has(afterScanBlockHookFlag) }

func (t *Table) AppendNamedArg(
	fmter Formatter, b []byte, name string, strct reflect.Value,
) ([]byte, bool) {
	if field, ok := t.FieldMap[name]; ok {
		return field.AppendValue(fmter, b, strct), true
	}
	return b, false
}

func quoteTableName(s string) Safe {
	return Safe(appendIdent(nil, internal.Bytes(s)))
}

func quoteColumnName(s string) Safe {
	return Safe(appendIdent(nil, internal.Bytes(s)))
}

//------------------------------------------------------------------------------

type UnknownColumnError struct {
	Table  *Table
	Column string
}

func (err *UnknownColumnError) Error() string {
	return fmt.Sprintf("ch: %s does not have column=%q",
		err.Table, err.Column)
}
