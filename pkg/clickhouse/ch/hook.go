package ch

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type QueryEvent struct {
	DB        *DB
	Model     Model
	IQuery    Query
	Query     string
	QueryArgs []any
	StartTime time.Time
	Result    sql.Result
	Err       error
	Stash     map[any]any
}

func (e *QueryEvent) Operation() string {
	if e.IQuery != nil {
		return e.IQuery.Operation()
	}
	return queryOperation(e.Query)
}
func queryOperation(query string) string {
	if idx := strings.IndexByte(query, ' '); idx > 0 {
		query = query[:idx]
	}
	if len(query) > 16 {
		query = query[:16]
	}
	return query
}

type QueryHook interface {
	BeforeQuery(context.Context, *QueryEvent) (context.Context, error)
	AfterQuery(context.Context, *QueryEvent)
}

func (db *DB) AddQueryHook(hook QueryHook) { db.queryHooks = append(db.queryHooks, hook) }
func (db *DB) beforeQuery(ctx context.Context, iquery Query, query string, params []any, model Model) (context.Context, *QueryEvent, error) {
	if len(db.queryHooks) == 0 {
		return ctx, nil, nil
	}
	evt := &QueryEvent{StartTime: time.Now(), DB: db, Model: model, IQuery: iquery, Query: query, QueryArgs: params}
	for i, hook := range db.queryHooks {
		var err error
		ctx, err = hook.BeforeQuery(ctx, evt)
		if err != nil {
			db._afterQuery(ctx, evt, nil, err, i-1)
			return nil, nil, err
		}
	}
	return ctx, evt, nil
}
func (db *DB) afterQuery(ctx context.Context, evt *QueryEvent, res *Result, err error) {
	if evt == nil {
		return
	}
	db._afterQuery(ctx, evt, res, err, len(db.queryHooks)-1)
}
func (db *DB) _afterQuery(ctx context.Context, evt *QueryEvent, res *Result, err error, hookIndex int) {
	evt.Err = err
	if res != nil {
		evt.Result = res
	}
	for i := hookIndex; i >= 0; i-- {
		db.queryHooks[i].AfterQuery(ctx, evt)
	}
}
func callAfterScanRowHook(ctx context.Context, v reflect.Value) error {
	return v.Interface().(AfterScanRowHook).AfterScanRow(ctx)
}
func callAfterScanRowHookSlice(ctx context.Context, slice reflect.Value) error {
	return callHookSlice(ctx, slice, callAfterScanRowHook)
}
func callHookSlice(ctx context.Context, slice reflect.Value, hook func(context.Context, reflect.Value) error) error {
	var ptr bool
	switch slice.Type().Elem().Kind() {
	case reflect.Ptr, reflect.Interface:
		ptr = true
	}
	var firstErr error
	sliceLen := slice.Len()
	for i := 0; i < sliceLen; i++ {
		v := slice.Index(i)
		if !ptr {
			v = v.Addr()
		}
		err := hook(ctx, v)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
