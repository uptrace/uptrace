package upql

import (
	"strconv"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type Where struct {
	Conds []Cond
}

type Group struct {
	Names []Name
}

type Columns struct {
	Names []Name
}

type Name struct {
	FuncName string
	AttrKey  string
}

func (n Name) String() string {
	return string(n.Append(nil))
}

func (n *Name) Append(b []byte) []byte {
	hasFunc := n.FuncName != "" && n.FuncName != "any"

	if hasFunc {
		b = append(b, n.FuncName...)
		b = append(b, '(')
	}

	b = append(b, n.AttrKey...)

	if hasFunc {
		b = append(b, ')')
	}

	return b
}

type Cond struct {
	Sep   CondSep
	Left  Name
	Op    string
	Right Value
}

type CondSep struct {
	Op     string
	Negate bool
}

type Value struct {
	Kind ValueKind
	Text string
}

func (v *Value) Append(b []byte) []byte {
	switch v.Kind {
	case StringValue:
		return chschema.AppendString(b, v.Text)
	case NumberValue:
		return append(b, v.Text...)
	case DurationValue:
		d, err := time.ParseDuration(v.Text)
		if err != nil {
			panic("err") // should not happen
		}
		return strconv.AppendInt(b, int64(d), 10)
	default:
		panic("not reached")
	}
}

//------------------------------------------------------------------------------

const (
	InvalidValue ValueKind = iota
	StringValue
	NumberValue
	DurationValue
)

type ValueKind int

func (k ValueKind) String() string {
	switch k {
	case StringValue:
		return "string"
	case NumberValue:
		return "number"
	case DurationValue:
		return "duration"
	default:
		return "invalid"
	}
}

func (k ValueKind) IsNum() bool {
	switch k {
	case NumberValue, DurationValue:
		return true
	default:
		return false
	}
}

//------------------------------------------------------------------------------

const (
	AndOp string = " AND "
	OrOp  string = " OR "
)

const (
	EqualOp    string = "="
	NotEqualOp string = "!="

	ContainsOp       string = "contains"
	DoesNotContainOp string = "does not contain"

	LikeOp    string = "like"
	NotLikeOp string = "not like"

	ExistsOp       string = "exists"
	DoesNotExistOp string = "does not exist"

	MatchesOp      string = "~"
	DoesNotMatchOp string = "!~"
)
