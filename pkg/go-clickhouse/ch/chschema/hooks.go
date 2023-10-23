package chschema

import (
	"context"
	"reflect"
)

type Query interface {
	QueryAppender
	Operation() string
	GetModel() Model
	GetTableName() string
}

type Model interface {
	ScanBlock(*Block) error
}

type AfterScanRowHook interface {
	AfterScanRow(context.Context) error
}

var afterScanBlockHookType = reflect.TypeOf((*AfterScanRowHook)(nil)).Elem()
