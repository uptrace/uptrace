package ch_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/extra/chdebug"
)

func chDB(opts ...ch.Option) *ch.DB {
	db := ch.Connect(opts...)
	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithEnabled(false),
		chdebug.FromEnv("CHDEBUG"),
	))
	return db
}

func TestCHError(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	err := db.Ping(ctx)
	require.NoError(t, err)

	res, err := db.ExecContext(ctx, "hi")
	require.Error(t, err)
	require.Nil(t, res)

	exc := err.(*ch.Error)
	require.Equal(t, int32(62), exc.Code)
	require.Equal(t, "DB::Exception", exc.Name)
}

func TestCHTimeout(t *testing.T) {
	ctx := context.Background()

	db := chDB(ch.WithTimeout(time.Second), ch.WithMaxRetries(0))
	defer db.Close()

	_, err := db.ExecContext(
		ctx, "SELECT sleepEachRow(0.01) from numbers(10000) settings max_block_size=10")
	require.Error(t, err)
	require.Contains(t, err.Error(), "i/o timeout")

	require.Eventually(t, func() bool {
		var num int
		err := db.NewSelect().ColumnExpr("count()").TableExpr("system.processes").Scan(ctx, &num)
		require.NoError(t, err)
		return num == 1
	}, time.Second, 100*time.Millisecond)
}

func TestPlaceholder(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	params := struct {
		A     int
		B     int
		Alias ch.Ident
	}{
		A:     1,
		B:     2,
		Alias: "sum",
	}

	t.Run("raw", func(t *testing.T) {
		var sum int
		err := db.QueryRow("SELECT ?a + ?b AS ?alias", params).Scan(&sum)
		require.NoError(t, err)
		require.Equal(t, 3, sum)

		res, err := db.Exec("SELECT ?a + ?b AS ?alias", params)
		require.NoError(t, err)

		n, err := res.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), n)
	})

	t.Run("query builder", func(t *testing.T) {
		var sum int
		err := db.NewSelect().ColumnExpr("?a + ?b AS ?alias", params).Scan(ctx, &sum)
		require.NoError(t, err)
		require.Equal(t, 3, sum)
	})
}

func TestScanArray(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	t.Run("uint64", func(t *testing.T) {
		var nums []uint64
		err := db.NewSelect().
			ColumnExpr("groupArray(number)").
			TableExpr("numbers(3)").
			Scan(ctx, &nums)
		require.NoError(t, err)
		require.Equal(t, []uint64{0, 1, 2}, nums)
	})

	t.Run("float64", func(t *testing.T) {
		var nums []float64
		var str string
		err := db.NewSelect().ColumnExpr("[1., 2, 3], 'hello'").Scan(ctx, &nums, &str)
		require.NoError(t, err)
		require.Equal(t, []float64{1, 2, 3}, nums)
		require.Equal(t, "hello", str)
	})
}

func TestScanEmptyResult(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	var m map[string]any
	err := db.NewSelect().TableExpr("numbers(0)").Scan(ctx, &m)
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"number": uint64(0),
	}, m)
}

func TestScanNaN(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	t.Run("uint32", func(t *testing.T) {
		var num uint32
		err := db.QueryRowContext(ctx, "SELECT NaN").Scan(&num)
		require.NoError(t, err)
		require.Equal(t, uint32(0), num)
	})

	t.Run("int32", func(t *testing.T) {
		var num int32
		err := db.QueryRowContext(ctx, "SELECT NaN").Scan(&num)
		require.NoError(t, err)
		require.Equal(t, int32(0), num)
	})
}

type Event struct {
	ch.BaseModel `ch:"goch_events,partition:toYYYYMM(created_at)"`

	ID        uint64
	Name      string `ch:",lc"`
	Count     uint32
	Keys      []string `ch:",lc"`
	Values    [][]string
	Kind      string    `ch:"type:Enum8('invalid' = 0, 'hello' = 1, 'world' = 2)"`
	CreatedAt time.Time `ch:",pk"`
}

type EventColumnar struct {
	ch.BaseModel `ch:"goch_events,columnar"`

	ID        []uint64
	Name      []string `ch:",lc"`
	Count     []uint32
	Keys      [][]string `ch:"type:Array(LowCardinality(String))"`
	Values    [][][]string
	Kind      []string `ch:"type:Enum8('invalid' = 0, 'hello' = 1, 'world' = 2)"`
	CreatedAt []time.Time
}

func TestORM(t *testing.T) {
	ctx := context.Background()

	db := chDB()
	defer db.Close()

	err := db.ResetModel(ctx, (*Event)(nil))
	require.NoError(t, err)

	tests := []func(t *testing.T, db *ch.DB){
		testORMStruct,
		testORMSlice,
		testORMColumnarStruct,
		testORMInvalidEnumValue,
	}
	for _, fn := range tests {
		_, err := db.NewTruncateTable().Model((*Event)(nil)).Exec(ctx)
		require.NoError(t, err)

		t.Run("", func(t *testing.T) {
			fn(t, db)
		})
	}
}

func testORMStruct(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	err := db.NewSelect().Model(new(Event)).Scan(ctx)
	require.Equal(t, sql.ErrNoRows, err)

	src := &Event{
		ID:        1,
		Name:      "hello",
		Count:     42,
		Keys:      []string{"foo", "bar"},
		Values:    [][]string{{}, {"hello", "world"}},
		Kind:      "hello",
		CreatedAt: time.Time{},
	}
	_, err = db.NewInsert().Model(src).Exec(ctx)
	require.NoError(t, err)

	dest := new(Event)
	err = db.NewSelect().Model(dest).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, src, dest)

	n, err := db.NewSelect().Model((*Event)(nil)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	names := make([]string, 0)
	counts := make([]uint32, 0)
	err = db.NewSelect().
		Model((*Event)(nil)).
		Column("name", "count").
		ScanColumns(ctx, &names, &counts)
	require.NoError(t, err)
	require.Equal(t, []string{"hello"}, names)
	require.Equal(t, []uint32{42}, counts)

	var m map[string]any
	err = db.NewSelect().Model((*Event)(nil)).ScanColumns(ctx, &m)
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"id":         []uint64{1},
		"name":       []string{"hello"},
		"count":      []uint32{42},
		"keys":       [][]string{{"foo", "bar"}},
		"values":     [][][]string{{{}, {"hello", "world"}}},
		"kind":       []string{"hello"},
		"created_at": []time.Time{{}},
	}, m)
}

func testORMSlice(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	var events []*Event
	err := db.NewSelect().Model(&events).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(events))

	src := []*Event{{
		ID:        1,
		Name:      "hello",
		Count:     42,
		Keys:      []string{"foo", "bar"},
		Values:    [][]string{{}, {"hello", "world"}},
		Kind:      "hello",
		CreatedAt: time.Time{},
	}, {

		ID:        2,
		Name:      "world",
		Count:     84,
		Keys:      []string{"1", "2", "3"},
		Values:    [][]string{{}, {"hello", "world"}, {}},
		Kind:      "world",
		CreatedAt: time.Unix(1000, 0),
	}}
	_, err = db.NewInsert().Model(&src).Exec(ctx)
	require.NoError(t, err)

	var dest []*Event
	err = db.NewSelect().Model(&dest).OrderExpr("id ASC").Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, src, dest)

	n, err := db.NewSelect().Model((*Event)(nil)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	var temp []struct {
		Name  string `ch:"type:LowCardinality(String)"`
		Count uint64
	}
	err = db.NewSelect().
		Model((*Event)(nil)).
		ColumnExpr("name, count(*) as count").
		GroupExpr("name").
		OrderExpr("name asc").
		Scan(ctx, &temp)
	require.NoError(t, err)
	require.Equal(t, 2, len(temp))
	require.Equal(t, "hello", temp[0].Name)
	require.Equal(t, uint64(1), temp[0].Count)
	require.Equal(t, "world", temp[1].Name)
	require.Equal(t, uint64(1), temp[1].Count)

	names := make([]string, 0)
	counts := make([]uint32, 0)
	err = db.NewSelect().
		Model((*Event)(nil)).
		Column("name", "count").
		ScanColumns(ctx, &names, &counts)
	require.NoError(t, err)
	require.Equal(t, []string{"hello", "world"}, names)
	require.Equal(t, []uint32{42, 84}, counts)

	var values []map[string]any
	err = db.NewSelect().Model((*Event)(nil)).Scan(ctx, &values)
	require.NoError(t, err)
	require.Equal(t, []map[string]any{{
		"id":         uint64(1),
		"name":       "hello",
		"count":      uint32(42),
		"keys":       []string{"foo", "bar"},
		"values":     [][]string{{}, {"hello", "world"}},
		"kind":       "hello",
		"created_at": time.Time{},
	}, {
		"id":         uint64(2),
		"name":       "world",
		"count":      uint32(84),
		"keys":       []string{"1", "2", "3"},
		"values":     [][]string{{}, {"hello", "world"}, {}},
		"kind":       "world",
		"created_at": time.Unix(1000, 0),
	}}, values)
}

func testORMColumnarStruct(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	err := db.NewSelect().Model(new(EventColumnar)).Scan(ctx)
	require.NoError(t, err)

	src := &EventColumnar{
		ID:        []uint64{1, 2},
		Name:      []string{"hello", "world"},
		Count:     []uint32{42, 84},
		Keys:      [][]string{{"foo", "bar"}, {"1", "2", "3"}},
		Values:    [][][]string{{{}, {"hello", "world"}}, {{}, {}, {}}},
		Kind:      []string{"hello", "world"},
		CreatedAt: []time.Time{{}, time.Unix(1000, 0)},
	}
	_, err = db.NewInsert().Model(src).Exec(ctx)
	require.NoError(t, err)

	dest := new(EventColumnar)
	err = db.NewSelect().Model(dest).OrderExpr("id ASC").Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, src, dest)
}

func testORMInvalidEnumValue(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	src := &Event{
		Kind: "foobar",
	}
	_, err := db.NewInsert().Model(src).Exec(ctx)
	require.NoError(t, err)

	dest := new(Event)
	err = db.NewSelect().Model(dest).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, "invalid", dest.Kind)
}
