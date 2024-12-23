package chschema

import (
	"context"
	"reflect"
)

type AfterScanRowHook interface{ AfterScanRow(context.Context) error }

var afterScanRowHookType = reflect.TypeOf((*AfterScanRowHook)(nil)).Elem()
