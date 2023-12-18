package ast

import (
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
)

var groupingFuncs = make(map[string]*Func)

var (
	FuncLower = NewGroupingFunc("lower", "lowerUTF8(?)")
	FuncUpper = NewGroupingFunc("upper", "upperUTF8(?)")
)

type Func struct {
	Name string
	Expr string
}

func NewGroupingFunc(name, expr string) *Func {
	fn := &Func{
		Name: name,
		Expr: expr,
	}
	if _, ok := groupingFuncs[fn.Name]; ok {
		panic("not reached")
	}
	groupingFuncs[fn.Name] = fn
	return fn
}

func (f *Func) AppendQuery(b []byte, arg ch.Safe) (_ []byte, err error) {
	return chschema.AppendQuery(b, f.Expr, arg), nil
}
