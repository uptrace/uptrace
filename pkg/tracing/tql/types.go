package tql

import (
	"fmt"
	"strings"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type Node interface {
	fmt.Stringer
	AppendString([]byte) []byte
}

type Where struct {
	Filters []Filter
}

type Selector struct {
	Columns []Column
}

type Grouping struct {
	Names []Name
}

type Column struct {
	Name  Name
	Alias string
}

type Name struct {
	FuncName string
	AttrKey  string
}

func (n Name) IsNum() bool {
	switch n.FuncName {
	case "":
		switch n.AttrKey {
		case attrkey.SpanCount,
			attrkey.SpanCountPerMin,
			attrkey.SpanErrorCount,
			attrkey.SpanErrorPct,
			attrkey.SpanErrorRate:
			return true
		}
	case "sum", "avg", "min", "max", "p50", "p75", "p90", "p99":
		return true
	}
	return false
}

func (n Name) String() string {
	b := n.AppendString(nil)
	return unsafeconv.String(b)
}

func (n *Name) AppendString(b []byte) []byte {
	switch n.FuncName {
	case "", "any":
		return append(b, n.AttrKey...)
	}

	b = append(b, n.FuncName...)
	b = append(b, '(')
	b = append(b, n.AttrKey...)
	b = append(b, ')')
	return b
}

type Expr struct {
	LHS Node
	Ops []ExprOp
}

func (e *Expr) String() string {
	b := e.AppendString(nil)
	return unsafeconv.String(b)
}

func (e *Expr) AppendString(b []byte) []byte {
	b = e.LHS.AppendString(b)
	for _, v := range e.Ops {
		b = append(b, ' ')
		b = append(b, v.Op...)
		b = append(b, ' ')
		b = v.RHS.AppendString(b)
	}
	return b
}

type ExprOp struct {
	Op  BinaryOp
	RHS Node
}

type BinaryOp string

//------------------------------------------------------------------------------

type Filter struct {
	BoolOp BoolOp
	LHS    Name
	Op     FilterOp
	RHS    Value
}

type BoolOp string

const (
	BoolAnd BoolOp = "AND"
	BoolOr  BoolOp = "OR"
)

type FilterOp string

const (
	FilterEqual    FilterOp = "="
	FilterNotEqual FilterOp = "!="

	FilterIn    FilterOp = "in"
	FilterNotIn FilterOp = "not in"

	FilterLike    FilterOp = "like"
	FilterNotLike FilterOp = "not like"

	FilterContains    FilterOp = "contains"
	FilterNotContains FilterOp = "not contains"

	FilterExists    FilterOp = "exists"
	FilterNotExists FilterOp = "not exists"

	// For compatibility with metrics.
	FilterRegexp    FilterOp = "~"
	FilterNotRegexp FilterOp = "!~"
)

type Value interface {
	String() string
}

type StringValue struct {
	Text string
}

func (v StringValue) String() string {
	return v.Text
}

type StringValues struct {
	Values []string
}

func (v StringValues) String() string {
	return strings.Join(v.Values, "|")
}

type NumberKind int

const (
	NumberUnitless NumberKind = iota
	NumberDuration
	NumberBytes
)

type Number struct {
	Kind NumberKind
	Text string
}

func (n *Number) String() string {
	return n.Text
}

func clean(attrKey string) string {
	if strings.HasPrefix(attrKey, "span.") {
		return strings.TrimPrefix(attrKey, "span")
	}
	return attrKey
}
