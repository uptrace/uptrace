package ast

import (
	"errors"
	"fmt"
)

var errAlias = errors.New("alias is required (AS alias)")

func (p *queryParser) parseQuery() (QueryPart, error) {

	{
		var where []Filter
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
			where, _err = p.where()
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
				where = nil
				goto i0_group_end
			}
		}
		return &Where{Filters: where}, nil
	i0_group_end:
	}

	{
		var grouping []GroupingElem
		_pos1 := p.Pos()
		{
			var _err error
			grouping, _err = p.grouping()
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
				grouping = nil
				goto r1_i0_group_end
			}
		}
		return &Grouping{Elems: grouping}, nil
	r1_i0_group_end:
	}

	{
		var grouping []GroupingElem
		var namedExpr NamedExpr
		_pos1 := p.Pos()
		{
			var _err error
			namedExpr, _err = p.namedExpr()
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
			var _err error
			grouping, _err = p.grouping()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r2_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				grouping = nil
				goto r2_i0_group_end
			}
		}
		{
			applyGrouping(namedExpr.Expr, grouping)
			return &Selector{Expr: namedExpr}, nil
		}
	r2_i0_group_end:
	}

	{
		var namedExpr NamedExpr
		_pos1 := p.Pos()
		{
			var _err error
			namedExpr, _err = p.namedExpr()
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
				namedExpr = NamedExpr{}
				goto r3_i0_group_end
			}
		}
		return &Selector{Expr: namedExpr}, nil
	r3_i0_group_end:
	}

	var where []Filter

	{
		var _err error
		where, _err = p.where()
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
	return &Where{Filters: where}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) filters() ([]Filter, error) {
	var filters []Filter

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

	{
		var filter Filter
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
				filter, _err = p.filter()
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
				filter.BoolOp = BoolAnd
				filters = append(filters, filter)
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

	for i := range filters {
		f := &filters[i]

		if alias, _ := SplitAliasName(f.LHS); alias != "" {
			return nil, fmt.Errorf("filters can't reference other metrics: %s", alias)
		}

		switch f.Op {
		case FilterEqual, FilterNotEqual, FilterRegexp, FilterNotRegexp, FilterIn:
			// ok
		default:
			return nil, fmt.Errorf(`only =, !=, ~, !~, and "in" are allowed inside curly brackets`)
		}
	}

	return filters, nil
}

func (p *queryParser) where() ([]Filter, error) {
	var filters []Filter

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
					goto r1_i0_no_match
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
					goto r1_i0_no_match
				}
			}
			_matchCount = _matchCount + 1
			{
				filter.BoolOp = boolOp
				filters = append(filters, filter)
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

	for i := range filters {
		f := &filters[i]

		if alias, _ := SplitAliasName(f.LHS); alias != "" {
			return nil, fmt.Errorf("where can't reference other metrics: %s", alias)
		}
	}

	return filters, nil
}

func (p *queryParser) filter() (Filter, error) {

	{
		var lhs *Token
		var values StringValues
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			lhs = _tok
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
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
				lhs = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
				values = StringValues{}
				goto i0_group_end
			}
		}
		return Filter{
			LHS: clean(lhs.Text),
			Op:  FilterIn,
			RHS: values,
		}, nil
	i0_group_end:
	}

	{
		var lhs *Token
		var values StringValues
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			lhs = _tok
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
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
				lhs = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				lhs = nil
				values = StringValues{}
				goto r1_i0_group_end
			}
		}
		return Filter{
			LHS: clean(lhs.Text),
			Op:  FilterNotIn,
			RHS: values,
		}, nil
	r1_i0_group_end:
	}

	{
		var lhs *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
			lhs = _tok
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
				lhs = nil
				goto r2_i0_group_end
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
					goto r2_i0_i3_alt1
				}
			}
			goto r2_i0_i3_has_match
		}

	r2_i0_i3_alt1:
		// "exists"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					lhs = nil
					goto r2_i0_group_end
				}
			}
		}

	r2_i0_i3_has_match:
		return Filter{
			LHS: clean(lhs.Text),
			Op:  FilterNotExists,
		}, nil
	r2_i0_group_end:
	}

	{
		var lhs *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
			lhs = _tok
		}
		// "exists"
		{
			_pos3 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos3)
					goto r3_i0_i1_alt1
				}
			}
			goto r3_i0_i1_has_match
		}

	r3_i0_i1_alt1:
		// "exist"
		{
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos1)
					lhs = nil
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_i1_has_match:
		return Filter{
			LHS: clean(lhs.Text),
			Op:  FilterExists,
		}, nil
	r3_i0_group_end:
	}

	var filterOp FilterOp
	var lhs *Token
	var value Value

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return Filter{}, errBacktrack
		}
		lhs = _tok
	}
	{
		var _err error
		filterOp, _err = p.filterOp()
		if _err != nil && _err != errBacktrack {
			return Filter{}, _err
		}
		_match := _err == nil
		if !_match {
			return Filter{}, errBacktrack
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
			return Filter{}, errBacktrack
		}
	}
	return Filter{
		LHS: clean(lhs.Text),
		Op:  filterOp,
		RHS: value,
	}, nil
}

func (p *queryParser) filterOp() (FilterOp, error) {

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return FilterNotLike, nil
	i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return FilterLike, nil
	r1_i0_group_end:
	}

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
		return FilterNotRegexp, nil
	r4_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "<"
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
		}
		return FilterLTE, nil
	r5_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == ">"
			if !_match {
				p.ResetPos(_pos1)
				goto r6_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r6_i0_group_end
			}
		}
		return FilterGTE, nil
	r6_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r7_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r7_i0_group_end
			}
		}
		return FilterRegexp, nil
	r7_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r8_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r8_i0_group_end
			}
		}
		return FilterRegexp, nil
	r8_i0_group_end:
	}

	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.Text == "<" || _tok.Text == ">" || _tok.Text == "=" || _tok.Text == "~"
		if !_match {
			return "", errBacktrack
		}
		t = _tok
	}
	return FilterOp(t.Text), nil
}

func (p *queryParser) values() (StringValues, error) {
	var ss []string

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

	return StringValues{ss}, nil
}

func (p *queryParser) value() (Value, error) {

	{
		var number Number
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

func (p *queryParser) boolOp() (BoolOp, error) {

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

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'o' || _tok.Text[0] == 'O') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R')
		if !_match {
			return "", errBacktrack
		}
	}
	return BoolOr, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) namedExprs() ([]NamedExpr, error) {
	var exprs []NamedExpr

	var namedExpr NamedExpr

	{
		var _err error
		namedExpr, _err = p.namedExpr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		exprs = append(exprs, namedExpr)
		p.cut()
	}

	{
		var namedExpr NamedExpr
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
				namedExpr, _err = p.namedExpr()
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
				exprs = append(exprs, namedExpr)
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

	return exprs, nil
}

func (p *queryParser) namedExpr() (NamedExpr, error) {

	{
		var alias string
		var expr Expr
		_pos1 := p.Pos()
		{
			var _err error
			expr, _err = p.expr()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
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
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				expr = nil
				goto i0_group_end
			}
		}
		return NamedExpr{
			Expr:     expr,
			HasAlias: true,
			Alias:    alias,
		}, nil
	i0_group_end:
	}

	var expr Expr

	{
		var _err error
		expr, _err = p.expr()
		if _err != nil && _err != errBacktrack {
			return NamedExpr{}, _err
		}
		_match := _err == nil
		if !_match {
			return NamedExpr{}, errBacktrack
		}
	}
	return NamedExpr{
		Expr:  expr,
		Alias: defaultAliasForExpr(expr),
	}, nil
}

func (p *queryParser) expr() (Expr, error) {

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

	{
		var number Number
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

	{
		var uniq *UniqExpr
		_pos1 := p.Pos()
		{
			var _err error
			uniq, _err = p.uniq()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return uniq, nil
	r1_i0_group_end:
	}

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
				goto r2_i0_group_end
			}
		}
		return funcCall, nil
	r2_i0_group_end:
	}

	{
		var metricExpr MetricExpr
		_pos1 := p.Pos()
		{
			var _err error
			metricExpr, _err = p.metricExpr()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		return &metricExpr, nil
	r3_i0_group_end:
	}

	var expr Expr

	{
		_tok := p.NextToken()
		_match := _tok.Text == "("
		if !_match {
			return nil, errBacktrack
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
	return ParenExpr{Expr: expr}, nil
}

func (p *queryParser) metricExpr() (MetricExpr, error) {

	var metricExpr1 MetricExpr

	{
		var _err error
		metricExpr1, _err = p.metricExpr1()
		if _err != nil && _err != errBacktrack {
			return MetricExpr{}, _err
		}
		_match := _err == nil
		if !_match {
			return MetricExpr{}, errBacktrack
		}
	}
	{
	}

	{
		var inlineGrouping []GroupingElem
		_pos1 := p.Pos()
		{
			var _err error
			inlineGrouping, _err = p.inlineGrouping()
			if _err != nil && _err != errBacktrack {
				return MetricExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		metricExpr1.Grouping = inlineGrouping
	r1_i0_group_end:
	}

	return metricExpr1, nil
}

func (p *queryParser) metricExpr1() (MetricExpr, error) {

	{
		var filters []Filter
		var name *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			name = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "{"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			filters, _err = p.filters()
			if _err != nil && _err != errBacktrack {
				return MetricExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				filters = nil
				goto i0_group_end
			}
		}
		return MetricExpr{
			Name:    name.Text,
			Filters: filters,
		}, nil
	i0_group_end:
	}

	{
		var name *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			name = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "{"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto r1_i0_group_end
			}
		}
		return MetricExpr{
			Name: name.Text,
		}, nil
	r1_i0_group_end:
	}

	var name *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return MetricExpr{}, errBacktrack
		}
		name = _tok
	}
	return MetricExpr{
		Name: name.Text,
	}, nil
}

func (p *queryParser) number() (Number, error) {

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
		return Number{Text: t.Text, Kind: NumberDuration}, nil
	i0_group_end:
	}

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
		return Number{Text: t.Text, Kind: NumberBytes}, nil
	r1_i0_group_end:
	}

	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == NUMBER_TOKEN
		if !_match {
			return Number{}, errBacktrack
		}
		t = _tok
	}
	return Number{Text: t.Text}, nil
}

func (p *queryParser) binaryOp() (BinaryOp, error) {

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "="
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
		return BinaryOp("=="), nil
	i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "!"
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
		return BinaryOp("!="), nil
	r1_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == ">"
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
		return BinaryOp(">="), nil
	r2_i0_group_end:
	}

	{
		_pos1 := p.Pos()
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
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		return BinaryOp("<="), nil
	r3_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'd' || _tok.Text[2] == 'D')
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		return BinaryOp("and"), nil
	r4_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'o' || _tok.Text[0] == 'O') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R')
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
		}
		return BinaryOp("or"), nil
	r5_i0_group_end:
	}

	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.Text == "+" || _tok.Text == "-" || _tok.Text == "/" || _tok.Text == "*" || _tok.Text == "%" || _tok.Text == "<" || _tok.Text == ">"
		if !_match {
			return "", errBacktrack
		}
		t = _tok
	}
	return BinaryOp(t.Text), nil
}

func (p *queryParser) alias() (string, error) {

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

func (p *queryParser) uniq() (*UniqExpr, error) {

	{
		var metricExpr MetricExpr
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'u' || _tok.Text[0] == 'U') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 'q' || _tok.Text[3] == 'Q')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			var _err error
			metricExpr, _err = p.metricExpr()
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
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				metricExpr = MetricExpr{}
				goto i0_group_end
			}
		}
		{
			alias, attr := SplitAliasName(metricExpr.Name)
			metricExpr.Name = alias
			uq := &UniqExpr{Name: metricExpr}
			if attr != "" {
				uq.Attrs = []string{attr}
			}
			return uq, nil
		}
	i0_group_end:
	}

	var idents []string
	var metricExpr MetricExpr

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'u' || _tok.Text[0] == 'U') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 'q' || _tok.Text[3] == 'Q')
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
		metricExpr, _err = p.metricExpr()
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
		_match := _tok.Text == ","
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		idents, _err = p.idents()
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
	return &UniqExpr{
		Name:  metricExpr,
		Attrs: idents,
	}, nil
}

func (p *queryParser) funcCall() (*FuncCall, error) {

	{
		var expr Expr
		var fn *Token
		var inlineGrouping []GroupingElem
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
			var _err error
			inlineGrouping, _err = p.inlineGrouping()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				expr = nil
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
				inlineGrouping = nil
				goto i0_group_end
			}
		}
		return &FuncCall{
			Func:     fn.Text,
			Arg:      expr,
			Grouping: inlineGrouping,
		}, nil
	i0_group_end:
	}

	var expr Expr
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
		var _err error
		expr, _err = p.expr()
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
	return &FuncCall{
		Func: fn.Text,
		Arg:  expr,
	}, nil
}

func (p *queryParser) args() ([]Expr, error) {
	var args []Expr

	var expr Expr

	{
		var _err error
		expr, _err = p.expr()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		args = append(args, expr)
		p.cut()
	}

	{
		var expr Expr
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
				expr, _err = p.expr()
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
				args = append(args, expr)
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

	return args, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) grouping() ([]GroupingElem, error) {

	var grouping1 []GroupingElem

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		grouping1, _err = p.grouping1()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	return grouping1, nil
}

func (p *queryParser) inlineGrouping() ([]GroupingElem, error) {

	var grouping1 []GroupingElem

	{
		_tok := p.NextToken()
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
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
		grouping1, _err = p.grouping1()
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
	return grouping1, nil
}

func (p *queryParser) grouping1() ([]GroupingElem, error) {
	var grouping []GroupingElem

	var groupingElem GroupingElem

	{
		var _err error
		groupingElem, _err = p.groupingElem()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		grouping = append(grouping, groupingElem)
		p.cut()
	}

	{
		var groupingElem GroupingElem
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
				groupingElem, _err = p.groupingElem()
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
				grouping = append(grouping, groupingElem)
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

	for i := range grouping {
		el := &grouping[i]

		if alias, _ := SplitAliasName(el.Name); alias != "" {
			return nil, fmt.Errorf("group by can't reference other metrics: %s", alias)
		}
	}

	return grouping, nil
}

func (p *queryParser) groupingElem() (GroupingElem, error) {

	{
		var alias string
		var groupingExpr GroupingElem
		_pos1 := p.Pos()
		{
			var _err error
			groupingExpr, _err = p.groupingExpr()
			if _err != nil && _err != errBacktrack {
				return GroupingElem{}, _err
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
				return GroupingElem{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				groupingExpr = GroupingElem{}
				goto i0_group_end
			}
		}
		{
			groupingExpr.HasAlias = true
			groupingExpr.Alias = alias
			return groupingExpr, nil
		}
	i0_group_end:
	}

	var groupingExpr GroupingElem

	{
		var _err error
		groupingExpr, _err = p.groupingExpr()
		if _err != nil && _err != errBacktrack {
			return GroupingElem{}, _err
		}
		_match := _err == nil
		if !_match {
			return GroupingElem{}, errBacktrack
		}
	}
	groupingExpr.Alias = defaultAliasForGrouping(&groupingExpr)
	return groupingExpr, nil
}

func (p *queryParser) groupingExpr() (GroupingElem, error) {

	{
		var arg *Token
		var funcName *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			funcName = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				funcName = nil
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				funcName = nil
				goto i0_group_end
			}
			arg = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				funcName = nil
				arg = nil
				goto i0_group_end
			}
		}
		{
			return GroupingElem{
				Func: funcName.Text,
				Name: arg.Text,
			}, nil
		}
	i0_group_end:
	}

	var name *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return GroupingElem{}, errBacktrack
		}
		name = _tok
	}
	return GroupingElem{Name: name.Text}, nil
}

func (p *queryParser) idents() ([]string, error) {
	var names []string

	var name *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		name = _tok
	}
	{
		names = append(names, name.Text)
		p.cut()
	}

	{
		var name *Token
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
				name = _tok
			}
			_matchCount = _matchCount + 1
			{
				names = append(names, name.Text)
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

	return names, nil
}
