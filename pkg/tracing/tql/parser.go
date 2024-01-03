package tql

import (
	"errors"
)

var errAlias = errors.New("alias is required (AS alias)")

func (p *queryParser) trace(name, rule string) {}

func (p *queryParser) parseQuery() (AST, error) {
	// if-match: "where" filters EOF

	{
		var filters []Filter
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'w' || _tok.Text[0] == 'W') && (_tok.Text[1] == 'h' || _tok.Text[1] == 'H') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 'r' || _tok.Text[3] == 'R') && (_tok.Text[4] == 'e' || _tok.Text[4] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			var _err error
			filters, _err = p.filters()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				filters = nil
				goto i0_group_end
			}
		}
		return &Where{Filters: filters}, nil
	i0_group_end:
	}

	// if-match: "group" "by" columns EOF

	{
		var columns []Column
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			var _err error
			columns, _err = p.columns()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				columns = nil
				goto r1_i0_group_end
			}
		}
		return &Grouping{Columns: columns}, nil
	r1_i0_group_end:
	}

	// if-match: "select" columns EOF

	{
		var columns []Column
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 6 && (_tok.Text[0] == 's' || _tok.Text[0] == 'S') && (_tok.Text[1] == 'e' || _tok.Text[1] == 'E') && (_tok.Text[2] == 'l' || _tok.Text[2] == 'L') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E') && (_tok.Text[4] == 'c' || _tok.Text[4] == 'C') && (_tok.Text[5] == 't' || _tok.Text[5] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			var _err error
			columns, _err = p.columns()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				columns = nil
				goto r2_i0_group_end
			}
		}
		return &Selector{Columns: columns}, nil
	r2_i0_group_end:
	}

	// if-match: columns EOF

	{
		var columns []Column
		_pos1 := p.Pos()
		{
			var _err error
			columns, _err = p.columns()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				columns = nil
				goto r3_i0_group_end
			}
		}
		return &Selector{Columns: columns}, nil
	r3_i0_group_end:
	}

	// match: filters EOF
	var filters []Filter

	{
		var _err error
		filters, _err = p.filters()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.ID == EOF_TOKEN
		if !_match {
			return nil, errBacktrack
		}
	}
	return &Where{Filters: filters}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) filters() ([]Filter, error) {
	var filters []Filter

	// if-match: '{' strings '}' filterOp value

	{
		var filterOp FilterOp
		var strings []string
		var value Value
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "{"
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			var _err error
			strings, _err = p.strings()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				strings = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			filterOp, _err = p.filterOp()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				strings = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			value, _err = p.value()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				strings = nil
				filterOp = ""
				goto i0_group_end
			}
		}
		{
			for _, attrKey := range strings {
				filters = append(filters, Filter{
					BoolOp: BoolOr,
					LHS:    Attr{Name: attrKey},
					Op:     filterOp,
					RHS:    value,
				})
			}
			return filters, nil
		}
	i0_group_end:
	}

	// match: filter
	var filter Filter

	{
		var _err error
		filter, _err = p.filter()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		filters = append(filters, filter)
		p.cut()
	}

	// match-each: boolOp filter

	{
		var boolOp BoolOp
		var filter Filter
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				var _err error
				boolOp, _err = p.boolOp()
				if _err != nil && _err != errBacktrack {
					return nil, _err
				}
				_match := _err == nil
				if !_match {
					p.ResetPos(_pos1)
					goto r2_i0_no_match
				}
			}
			{
				var _err error
				filter, _err = p.filter()
				if _err != nil && _err != errBacktrack {
					return nil, _err
				}
				_match := _err == nil
				if !_match {
					p.ResetPos(_pos1)
					boolOp = ""
					goto r2_i0_no_match
				}
			}
			_matchCount = _matchCount + 1
			{
				filter.BoolOp = boolOp
				filters = append(filters, filter)
				p.cut()
			}
			continue
		r2_i0_no_match:
			p.ResetPos(_pos1)
			if _matchCount >= 0 {
				break
			}
			return nil, errBacktrack
		}
	}

	return filters, nil
}

func (p *queryParser) boolOp() (BoolOp, error) {
	// if-match: "and"

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'd' || _tok.Text[2] == 'D')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return BoolAnd, nil
	i0_group_end:
	}

	// match: "or"

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'o' || _tok.Text[0] == 'O') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R')
		if !_match {
			return "", errBacktrack
		}
	}
	return BoolOr, nil
}

func (p *queryParser) filter() (Filter, error) {
	// if-match: expr=attrOrFunc "in" '(' values ')'

	{
		var expr Expr
		var values StringValues
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.attrOrFunc()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			values, _err = p.values()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				values = StringValues{}
				goto i0_group_end
			}
		}
		return Filter{
			LHS: expr,
			Op:  FilterIn,
			RHS: values,
		}, nil
	i0_group_end:
	}

	// if-match: expr=attrOrFunc "not" "in" '(' values ')'

	{
		var expr Expr
		var values StringValues
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.attrOrFunc()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto r1_i0_group_end
			}
		}
		{
			var _err error
			values, _err = p.values()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				values = StringValues{}
				goto r1_i0_group_end
			}
		}
		return Filter{
			LHS: expr,
			Op:  FilterNotIn,
			RHS: values,
		}, nil
	r1_i0_group_end:
	}

	// if-match: expr=attrOrFunc filterOp value

	{
		var expr Expr
		var filterOp FilterOp
		var value Value
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.attrOrFunc()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			var _err error
			filterOp, _err = p.filterOp()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto r2_i0_group_end
			}
		}
		{
			var _err error
			value, _err = p.value()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				filterOp = ""
				goto r2_i0_group_end
			}
		}
		return Filter{
			LHS: expr,
			Op:  filterOp,
			RHS: value,
		}, nil
	r2_i0_group_end:
	}

	// if-match: attr "does"? "not" ("exist" | "exists")

	{
		var attr Attr
		_pos1 := p.Pos()
		{
			var _err error
			attr, _err = p.attr()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		{
			_pos3 := p.Pos()
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
			if _match {
			} else {
				p.ResetPos(_pos3)
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				attr = Attr{}
				goto r3_i0_group_end
			}
		}
		// "exist"
		{
			_pos5 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos5)
					goto r3_i0_i3_alt1
				}
			}
			goto r3_i0_i3_has_match
		}

	r3_i0_i3_alt1:
		// "exists"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					attr = Attr{}
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_i3_has_match:
		return Filter{
			LHS: attr,
			Op:  FilterNotExists,
		}, nil
	r3_i0_group_end:
	}

	// if-match: attr ("exist" | "exists")

	{
		var attr Attr
		_pos1 := p.Pos()
		{
			var _err error
			attr, _err = p.attr()
			if _err != nil && _err != errBacktrack {
				return Filter{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		// "exist"
		{
			_pos3 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos3)
					goto r4_i0_i1_alt1
				}
			}
			goto r4_i0_i1_has_match
		}

	r4_i0_i1_alt1:
		// "exists"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					attr = Attr{}
					goto r4_i0_group_end
				}
			}
		}

	r4_i0_i1_has_match:
		return Filter{
			LHS: attr,
			Op:  FilterExists,
		}, nil
	r4_i0_group_end:
	}

	// match: attr
	var attr Attr

	{
		var _err error
		attr, _err = p.attr()
		if _err != nil && _err != errBacktrack {
			return Filter{}, _err
		}
		_match := _err == nil
		if !_match {
			return Filter{}, errBacktrack
		}
	}
	return Filter{
		LHS: attr,
		Op:  FilterEqual,
		RHS: NumberValue{Text: "1"},
	}, nil
}

func (p *queryParser) filterOp() (FilterOp, error) {
	// if-match: '>' '='

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == ">"
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return FilterOp(">="), nil
	i0_group_end:
	}

	// if-match: '<' '='

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "<"
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return FilterOp("<="), nil
	r1_i0_group_end:
	}

	// if-match: '=' '='

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		return FilterEqual, nil
	r2_i0_group_end:
	}

	// if-match: '!' '=' | '<' '>'

	{
		_pos1 := p.Pos()
		// '!' '='
		{
			{
				_tok := p.NextToken()
				_match := _tok.Text == "!"
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_alt1
				}
			}
			{
				_tok := p.NextToken()
				_match := _tok.Text == "="
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_alt1
				}
			}
			goto r3_i0_has_match
		}

	r3_i0_alt1:
		// '<' '>'
		{
			{
				_tok := p.NextToken()
				_match := _tok.Text == "<"
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_group_end
				}
			}
			{
				_tok := p.NextToken()
				_match := _tok.Text == ">"
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_has_match:
		return FilterNotEqual, nil
	r3_i0_group_end:
	}

	// if-match: '!' '~'

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "!"
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		return FilterNotContains, nil
	r4_i0_group_end:
	}

	// if-match: t=[<>=~]

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "<" || _tok.Text == ">" || _tok.Text == "=" || _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
			t = _tok
		}
		return FilterOp(t.Text), nil
	r5_i0_group_end:
	}

	// if-match: "does"? "not" ("contain" | "contains" | "include" | "includes")

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
			if _match {
			} else {
				p.ResetPos(_pos1)
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto r6_i0_group_end
			}
		}
		// "contain"
		{
			_pos4 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N')
				if !_match {
					p.ResetPos(_pos4)
					goto r6_i0_i2_alt1
				}
			}
			goto r6_i0_i2_has_match
		}

	r6_i0_i2_alt1:
		// "contains"
		{
			_pos6 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos6)
					goto r6_i0_i2_alt2
				}
			}
			goto r6_i0_i2_has_match
		}

	r6_i0_i2_alt2:
		// "include"
		{
			_pos8 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'c' || _tok.Text[2] == 'C') && (_tok.Text[3] == 'l' || _tok.Text[3] == 'L') && (_tok.Text[4] == 'u' || _tok.Text[4] == 'U') && (_tok.Text[5] == 'd' || _tok.Text[5] == 'D') && (_tok.Text[6] == 'e' || _tok.Text[6] == 'E')
				if !_match {
					p.ResetPos(_pos8)
					goto r6_i0_i2_alt3
				}
			}
			goto r6_i0_i2_has_match
		}

	r6_i0_i2_alt3:
		// "includes"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'c' || _tok.Text[2] == 'C') && (_tok.Text[3] == 'l' || _tok.Text[3] == 'L') && (_tok.Text[4] == 'u' || _tok.Text[4] == 'U') && (_tok.Text[5] == 'd' || _tok.Text[5] == 'D') && (_tok.Text[6] == 'e' || _tok.Text[6] == 'E') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					goto r6_i0_group_end
				}
			}
		}

	r6_i0_i2_has_match:
		return FilterNotContains, nil
	r6_i0_group_end:
	}

	// if-match: "contain" | "contains" | "include" | "includes"

	{
		_pos1 := p.Pos()
		// "contain"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N')
				if !_match {
					p.ResetPos(_pos1)
					goto r7_i0_alt1
				}
			}
			goto r7_i0_has_match
		}

	r7_i0_alt1:
		// "contains"
		{
			_pos4 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos4)
					goto r7_i0_alt2
				}
			}
			goto r7_i0_has_match
		}

	r7_i0_alt2:
		// "include"
		{
			_pos6 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'c' || _tok.Text[2] == 'C') && (_tok.Text[3] == 'l' || _tok.Text[3] == 'L') && (_tok.Text[4] == 'u' || _tok.Text[4] == 'U') && (_tok.Text[5] == 'd' || _tok.Text[5] == 'D') && (_tok.Text[6] == 'e' || _tok.Text[6] == 'E')
				if !_match {
					p.ResetPos(_pos6)
					goto r7_i0_alt3
				}
			}
			goto r7_i0_has_match
		}

	r7_i0_alt3:
		// "includes"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'c' || _tok.Text[2] == 'C') && (_tok.Text[3] == 'l' || _tok.Text[3] == 'L') && (_tok.Text[4] == 'u' || _tok.Text[4] == 'U') && (_tok.Text[5] == 'd' || _tok.Text[5] == 'D') && (_tok.Text[6] == 'e' || _tok.Text[6] == 'E') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					goto r7_i0_group_end
				}
			}
		}

	r7_i0_has_match:
		return FilterContains, nil
	r7_i0_group_end:
	}

	// if-match: "not" "like"

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto r8_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto r8_i0_group_end
			}
		}
		return FilterNotLike, nil
	r8_i0_group_end:
	}

	// if-match: "like"

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto r9_i0_group_end
			}
		}
		return FilterLike, nil
	r9_i0_group_end:
	}

	// if-match: '!' '~'

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "!"
			if !_match {
				p.ResetPos(_pos1)
				goto r10_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r10_i0_group_end
			}
		}
		return FilterNotRegexp, nil
	r10_i0_group_end:
	}

	// match: '~'

	{
		_tok := p.NextToken()
		_match := _tok.Text == "~"
		if !_match {
			return "", errBacktrack
		}
	}
	return FilterRegexp, nil
}

func (p *queryParser) value() (Value, error) {
	// if-match: number

	{
		var number NumberValue
		_pos1 := p.Pos()
		{
			var _err error
			number, _err = p.number()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return number, nil
	i0_group_end:
	}

	// match: t=(IDENT | VALUE)
	var t *Token

	// t=IDENT
	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_alt1
			}
			t = _tok
		}
		goto r1_i0_has_match
	}

r1_i0_alt1:
	// t=VALUE
	{
		{
			_tok := p.NextToken()
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				return nil, errBacktrack
			}
			t = _tok
		}
	}

r1_i0_has_match:
	return StringValue{Text: t.Text}, nil
}

func (p *queryParser) number() (NumberValue, error) {
	// if-match: t=DURATION

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == DURATION_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			t = _tok
		}
		return NumberValue{Text: t.Text, Kind: NumberDuration}, nil
	i0_group_end:
	}

	// if-match: t=BYTES

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == BYTES_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			t = _tok
		}
		return NumberValue{Text: t.Text, Kind: NumberBytes}, nil
	r1_i0_group_end:
	}

	// match: t=NUMBER
	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == NUMBER_TOKEN
		if !_match {
			return NumberValue{}, errBacktrack
		}
		t = _tok
	}
	return NumberValue{Text: t.Text}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) columns() ([]Column, error) {
	var columns []Column

	// match: column
	var column []Column

	{
		var _err error
		column, _err = p.column()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		columns = append(columns, column...)
		p.cut()
	}

	// match-each: ',' column

	{
		var column []Column
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				_tok := p.NextToken()
				_match := _tok.Text == ","
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			{
				var _err error
				column, _err = p.column()
				if _err != nil && _err != errBacktrack {
					return nil, _err
				}
				_match := _err == nil
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			_matchCount = _matchCount + 1
			{
				columns = append(columns, column...)
				p.cut()
			}
			continue
		r1_i0_no_match:
			p.ResetPos(_pos1)
			if _matchCount >= 0 {
				break
			}
			return nil, errBacktrack
		}
	}

	return columns, nil
}

func (p *queryParser) column() ([]Column, error) {
	// if-match: expr alias

	{
		var alias string
		var expr Expr
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.expr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			var _err error
			alias, _err = p.alias()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto i0_group_end
			}
		}
		return []Column{{
			Value: expr,
			Alias: alias,
		}}, nil
	i0_group_end:
	}

	// if-match: expr

	{
		var expr Expr
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.expr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return []Column{{
			Value: expr,
		}}, nil
	r1_i0_group_end:
	}

	// match: '{' strings '}' '(' attr ')'
	var attr Attr
	var strings []string

	{
		_tok := p.NextToken()
		_match := _tok.Text == "{"
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		strings, _err = p.strings()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == "}"
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == "("
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		attr, _err = p.attr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == ")"
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		columns := make([]Column, len(strings))
		for i, funcName := range strings {
			columns[i] = Column{
				Value: &FuncCall{
					Func: funcName,
					Arg:  attr,
				},
			}
		}
		return columns, nil
	}
}

func (p *queryParser) expr() (Expr, error) {
	// if-match: binaryExpr

	{
		var binaryExpr *BinaryExpr
		_pos1 := p.Pos()
		{
			var _err error
			binaryExpr, _err = p.binaryExpr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return binaryExpr, nil
	i0_group_end:
	}

	return p.term()
}

func (p *queryParser) binaryExpr() (*BinaryExpr, error) {
	// match: lhs=term binaryOp rhs=expr
	var binaryOp BinaryOp
	var lhs Expr
	var rhs Expr

	{
		var _err error
		lhs, _err = p.term()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		binaryOp, _err = p.binaryOp()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		rhs, _err = p.expr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	expr := &BinaryExpr{
		Op:  binaryOp,
		LHS: lhs,
		RHS: rhs,
	}
	return binaryExprPrecedence(expr), nil
}

func (p *queryParser) term() (Expr, error) {
	// if-match: funcCall

	{
		var funcCall *FuncCall
		_pos1 := p.Pos()
		{
			var _err error
			funcCall, _err = p.funcCall()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return funcCall, nil
	i0_group_end:
	}

	// if-match: attr

	{
		var attr Attr
		_pos1 := p.Pos()
		{
			var _err error
			attr, _err = p.attr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return attr, nil
	r1_i0_group_end:
	}

	// if-match: number

	{
		var number NumberValue
		_pos1 := p.Pos()
		{
			var _err error
			number, _err = p.number()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			if number.Kind != NumberUnitless {
				return nil, errors.New("numbers with units aren't allowed in exprs")
			}
			return number, nil
		}
	r2_i0_group_end:
	}

	// match: '(' binaryExpr ')'
	var binaryExpr *BinaryExpr

	{
		_tok := p.NextToken()
		_match := _tok.Text == "("
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		binaryExpr, _err = p.binaryExpr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == ")"
		if !_match {
			return nil, errBacktrack
		}
	}
	return ParenExpr{Expr: binaryExpr}, nil
}

func (p *queryParser) binaryOp() (BinaryOp, error) {
	// match: t=[=+-/*%]
	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.Text == "=" || _tok.Text == "+" || _tok.Text == "-" || _tok.Text == "/" || _tok.Text == "*" || _tok.Text == "%"
		if !_match {
			return "", errBacktrack
		}
		t = _tok
	}
	return BinaryOp(t.Text), nil
}

func (p *queryParser) attrOrFunc() (Expr, error) {
	// if-match: funcCall

	{
		var funcCall *FuncCall
		_pos1 := p.Pos()
		{
			var _err error
			funcCall, _err = p.funcCall()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return funcCall, nil
	i0_group_end:
	}

	// match: attr
	var attr Attr

	{
		var _err error
		attr, _err = p.attr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	return attr, nil
}

func (p *queryParser) attr() (Attr, error) {
	// match: t=(IDENT | VALUE)
	var t *Token

	// t=IDENT
	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_alt1
			}
			t = _tok
		}
		goto i0_has_match
	}

i0_alt1:
	// t=VALUE
	{
		{
			_tok := p.NextToken()
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				return Attr{}, errBacktrack
			}
			t = _tok
		}
	}

i0_has_match:
	return Attr{
		Name: clean(t.Text),
	}, nil
}

func (p *queryParser) funcCall() (*FuncCall, error) {
	// if-match: fn=IDENT '(' expr ')'

	{
		var expr Expr
		var fn *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			fn = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			expr, _err = p.expr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				expr = nil
				goto i0_group_end
			}
		}
		return &FuncCall{
			Func: fn.Text,
			Arg:  expr,
		}, nil
	i0_group_end:
	}

	// match: fn=IDENT '(' ')'
	var fn *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		fn = _tok
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == "("
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := _tok.Text == ")"
		if !_match {
			return nil, errBacktrack
		}
	}
	return &FuncCall{
		Func: fn.Text,
	}, nil

}

func (p *queryParser) alias() (string, error) {
	// match: "as"

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 's' || _tok.Text[1] == 'S')
		if !_match {
			return "", errBacktrack
		}
	}
	tok := p.NextToken()
	if tok.ID != IDENT_TOKEN {
		return "", errAlias
	}
	return tok.Text, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) strings() ([]string, error) {
	var ss []string

	// match: t=IDENT
	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		t = _tok
	}
	ss = append(ss, t.Text)

	// match-each: ',' t=IDENT

	{
		var t *Token
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				_tok := p.NextToken()
				_match := _tok.Text == ","
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			{
				_tok := p.NextToken()
				_match := _tok.ID == IDENT_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
				t = _tok
			}
			_matchCount = _matchCount + 1
			ss = append(ss, t.Text)
			continue
		r1_i0_no_match:
			p.ResetPos(_pos1)
			if _matchCount >= 0 {
				break
			}
			return nil, errBacktrack
		}
	}

	return ss, nil
}

func (p *queryParser) values() (StringValues, error) {
	var ss []string

	// match: t=(IDENT|VALUE|NUMBER)
	var t *Token

	// t=IDENT
	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_alt1
			}
			t = _tok
		}
		goto i0_has_match
	}

i0_alt1:
	// t=VALUE
	{
		_pos3 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				p.ResetPos(_pos3)
				goto i0_alt2
			}
			t = _tok
		}
		goto i0_has_match
	}

i0_alt2:
	// t=NUMBER
	{
		{
			_tok := p.NextToken()
			_match := _tok.ID == NUMBER_TOKEN
			if !_match {
				return StringValues{}, errBacktrack
			}
			t = _tok
		}
	}

i0_has_match:
	ss = append(ss, t.Text)

	// match-each: ',' t=(IDENT|VALUE|NUMBER)

	{
		var t *Token
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				_tok := p.NextToken()
				_match := _tok.Text == ","
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			// t=IDENT
			{
				_pos3 := p.Pos()
				{
					_tok := p.NextToken()
					_match := _tok.ID == IDENT_TOKEN
					if !_match {
						p.ResetPos(_pos3)
						goto r1_i0_i1_alt1
					}
					t = _tok
				}
				goto r1_i0_i1_has_match
			}

		r1_i0_i1_alt1:
			// t=VALUE
			{
				_pos5 := p.Pos()
				{
					_tok := p.NextToken()
					_match := _tok.ID == VALUE_TOKEN
					if !_match {
						p.ResetPos(_pos5)
						goto r1_i0_i1_alt2
					}
					t = _tok
				}
				goto r1_i0_i1_has_match
			}

		r1_i0_i1_alt2:
			// t=NUMBER
			{
				{
					_tok := p.NextToken()
					_match := _tok.ID == NUMBER_TOKEN
					if !_match {
						p.ResetPos(_pos1)
						t = nil
						goto r1_i0_no_match
					}
					t = _tok
				}
			}

		r1_i0_i1_has_match:
			_matchCount = _matchCount + 1
			ss = append(ss, t.Text)
			continue
		r1_i0_no_match:
			p.ResetPos(_pos1)
			if _matchCount >= 0 {
				break
			}
			return StringValues{}, errBacktrack
		}
	}

	return StringValues{Strings: ss}, nil
}
