//go:build ignore

package tql

import (
	"errors"
)

var errAlias = errors.New("alias is required (AS alias)")

func (p *queryParser) trace(name, rule string) {}

func (p *queryParser) parseQuery() (any, error) {
	// if-match: "where" filters EOF
	return &Where{Filters: filters}, nil

	// if-match: "group" "by" names EOF
	return &Grouping{Names: names}, nil

	// if-match: "select" columns EOF
	return &Selector{Columns: columns}, nil

	// if-match: columns EOF
	return &Selector{Columns: columns}, nil

	// match: filters EOF
	return &Where{Filters: filters}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) filters() ([]Filter, error) {
	var filters []Filter

	// if-match: '{' simples '}' filterOp value
	{
		for _, attrKey := range simples {
			filters = append(filters, Filter{
				BoolOp: BoolOr,
				LHS:    Name{AttrKey: clean(attrKey)},
				Op:     filterOp,
				RHS:    value,
			})
		}
		return filters, nil
	}

	// match: filter
	{
		filters = append(filters, filter)
		p.cut()
	}

	// match-each: boolOp filter
	{
		filter.BoolOp = boolOp
		filters = append(filters, filter)
		p.cut()
	}

	return filters, nil
}

func (p *queryParser) boolOp() (BoolOp, error) {
	// if-match: "and"
	return BoolAnd, nil

	// match: "or"
	return BoolOr, nil
}

func (p *queryParser) filter() (Filter, error) {
	// if-match: name "in" '(' values ')'
	return Filter{
		LHS: name,
		Op:  FilterIn,
		RHS: values,
	}, nil

	// if-match: name "not" "in" '(' values ')'
	return Filter{
		LHS: name,
		Op:  FilterNotIn,
		RHS: values,
	}, nil

	// if-match: name filterOp value
	return Filter{
		LHS: name,
		Op:  filterOp,
		RHS: value,
	}, nil

	// if-match: key=(IDENT | VALUE) "does"? "not" ("exist" | "exists")
	return Filter{
		LHS: Name{AttrKey: clean(key.Text)},
		Op:  FilterNotExists,
	}, nil

	// if-match: key=(IDENT | VALUE) ("exist" | "exists")
	return Filter{
		LHS: Name{AttrKey: clean(key.Text)},
		Op:  FilterExists,
	}, nil

	// match: key=IDENT
	return Filter{
		LHS: Name{AttrKey: clean(key.Text)},
		Op:  FilterEqual,
		RHS: &Number{Text: "1"},
	}, nil
}

func (p *queryParser) filterOp() (FilterOp, error) {
	// if-match: '>' '='
	return FilterOp(">="), nil

	// if-match: '<' '='
	return FilterOp("<="), nil

	// if-match: '=' '='
	return FilterEqual, nil

	// if-match: '!' '=' | '<' '>'
	return FilterNotEqual, nil

	// if-match: '!' '~'
	return FilterNotContains, nil

	// if-match: t=[<>=~]
	return FilterOp(t.Text), nil

	// if-match: "does"? "not" ("contain" | "contains")
	return FilterNotContains, nil

	// if-match: "contain" | "contains"
	return FilterContains, nil

	// if-match: "not" "like"
	return FilterNotLike, nil

	// if-match: "like"
	return FilterLike, nil

	// if-match: '!' '~'
	return FilterNotRegexp, nil

	// match: '~'
	return FilterRegexp, nil
}

func (p *queryParser) value() (Value, error) {
	// if-match: number
	return number, nil

	// match: t=(IDENT | VALUE)
	return StringValue{Text: t.Text}, nil
}

func (p *queryParser) number() (*Number, error) {
	// if-match: t=DURATION
	return &Number{Text: t.Text, Kind: NumberDuration}, nil

	// if-match: t=BYTES
	return &Number{Text: t.Text, Kind: NumberBytes}, nil

	// match: t=NUMBER
	return &Number{Text: t.Text}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) columns() ([]Column, error) {
	var columns []Column

	// match: column
	{
		columns = append(columns, column...)
		p.cut()
	}

	// match-each: ',' column
	{
		columns = append(columns, column...)
		p.cut()
	}

	return columns, nil
}

func (p *queryParser) column() ([]Column, error) {
	// if-match: name alias
	return []Column{{
		Name:  name,
		Alias: alias,
	}}, nil

	// if-match: '{' simples '}' '(' attr=IDENT ')'
	{
		columns := make([]Column, len(simples))
		for i, funcName := range simples {
			columns[i] = Column{
				Name: Name{
					FuncName: funcName,
					AttrKey:  clean(attr.Text),
				},
			}
		}
		return columns, nil
	}

	// match: name
	return []Column{{Name: name}}, nil
}

func (p *queryParser) name() (Name, error) {
	// if-match: fn=IDENT '(' attr=(IDENT | VALUE) ')'
	return Name{
		FuncName: fn.Text,
		AttrKey:  clean(attr.Text),
	}, nil

	// if-match: fn=IDENT '(' ')'
	return Name{
		FuncName: fn.Text,
	}, nil

	// match: t=(IDENT | VALUE)
	return Name{
		AttrKey: clean(t.Text),
	}, nil
}

func (p *queryParser) binaryOp() (BinaryOp, error) {
	// match: t=[=+-/*%]
	return BinaryOp(t.Text), nil
}

func (p *queryParser) alias() (string, error) {
	// match: "as"

	tok := p.NextToken()
	if tok.ID != IDENT_TOKEN {
		return "", errAlias
	}
	return tok.Text, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) simples() ([]string, error) {
	var ss []string

	// match: t=IDENT
	ss = append(ss, t.Text)

	// match-each: ',' t=IDENT
	ss = append(ss, t.Text)

	return ss, nil
}

func (p *queryParser) names() ([]Name, error) {
	var names []Name

	// match: name
	{
		names = append(names, name)
		p.cut()
	}

	// match-each: ',' name
	{
		names = append(names, name)
		p.cut()
	}

	return names, nil
}

func (p *queryParser) values() (StringValues, error) {
	var ss []string

	// match: t=(IDENT|VALUE|NUMBER)
	ss = append(ss, t.Text)

	// match-each: ',' t=(IDENT|VALUE|NUMBER)
	ss = append(ss, t.Text)

	return StringValues{ss}, nil
}
