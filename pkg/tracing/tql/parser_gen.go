package tql

import (
	"errors"
)

var errAlias = errors.New("alias is required (AS alias)")

func (p *queryParser) trace(name, rule string) {}

func (p *queryParser) parseQuery() (any, error) {

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

	{
		var names []Name
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
			names, _err = p.names()
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
				names = nil
				goto r1_i0_group_end
			}
		}
		return &Grouping{Names: names}, nil
	r1_i0_group_end:
	}

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

	{
		var filterOp FilterOp
		var simples []string
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
			simples, _err = p.simples()
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
				simples = nil
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
				simples = nil
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
				simples = nil
				filterOp = ""
				goto i0_group_end
			}
		}
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
	i0_group_end:
	}

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

func (p *queryParser) filter() (Filter, error) {

	{
		var name Name
		var values StringValues
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
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
				name = Name{}
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
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
				name = Name{}
				goto i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				values = StringValues{}
				goto i0_group_end
			}
		}
		return Filter{
			LHS: name,
			Op:  FilterIn,
			RHS: values,
		}, nil
	i0_group_end:
	}

	{
		var name Name
		var values StringValues
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
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
				name = Name{}
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
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
				name = Name{}
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				values = StringValues{}
				goto r1_i0_group_end
			}
		}
		return Filter{
			LHS: name,
			Op:  FilterNotIn,
			RHS: values,
		}, nil
	r1_i0_group_end:
	}

	{
		var filterOp FilterOp
		var name Name
		var value Value
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
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
				name = Name{}
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
				name = Name{}
				filterOp = ""
				goto r2_i0_group_end
			}
		}
		return Filter{
			LHS: name,
			Op:  filterOp,
			RHS: value,
		}, nil
	r2_i0_group_end:
	}

	{
		var key *Token
		_pos1 := p.Pos()
		// key=IDENT
		{
			{
				_tok := p.NextToken()
				_match := _tok.ID == IDENT_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_i0_alt1
				}
				key = _tok
			}
			goto r3_i0_i0_has_match
		}

	r3_i0_i0_alt1:
		// key=VALUE
		{
			{
				_tok := p.NextToken()
				_match := _tok.ID == VALUE_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					key = nil
					goto r3_i0_group_end
				}
				key = _tok
			}
		}

	r3_i0_i0_has_match:
		{
			_pos6 := p.Pos()
			_tok := p.NextToken()
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
			if _match {
			} else {
				p.ResetPos(_pos6)
			}
		}
		{
			_tok := p.NextToken()
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				key = nil
				goto r3_i0_group_end
			}
		}
		// "exist"
		{
			_pos8 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos8)
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
					key = nil
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_i3_has_match:
		return Filter{
			LHS: Name{AttrKey: clean(key.Text)},
			Op:  FilterNotExists,
		}, nil
	r3_i0_group_end:
	}

	{
		var key *Token
		_pos1 := p.Pos()
		// key=IDENT
		{
			{
				_tok := p.NextToken()
				_match := _tok.ID == IDENT_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					goto r4_i0_i0_alt1
				}
				key = _tok
			}
			goto r4_i0_i0_has_match
		}

	r4_i0_i0_alt1:
		// key=VALUE
		{
			{
				_tok := p.NextToken()
				_match := _tok.ID == VALUE_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					key = nil
					goto r4_i0_group_end
				}
				key = _tok
			}
		}

	r4_i0_i0_has_match:
		// "exist"
		{
			_pos6 := p.Pos()
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos6)
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
					key = nil
					goto r4_i0_group_end
				}
			}
		}

	r4_i0_i1_has_match:
		return Filter{
			LHS: Name{AttrKey: clean(key.Text)},
			Op:  FilterExists,
		}, nil
	r4_i0_group_end:
	}

	var key *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return Filter{}, errBacktrack
		}
		key = _tok
	}
	return Filter{
		LHS: Name{AttrKey: clean(key.Text)},
		Op:  FilterEqual,
		RHS: &Number{Text: "1"},
	}, nil
}

func (p *queryParser) filterOp() (FilterOp, error) {

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
		return FilterNotContains, nil
	r4_i0_group_end:
	}

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
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
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
			{
				_tok := p.NextToken()
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
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

	{
		var number *Number
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

func (p *queryParser) number() (*Number, error) {

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
		return &Number{Text: t.Text, Kind: NumberDuration}, nil
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
		return &Number{Text: t.Text, Kind: NumberBytes}, nil
	r1_i0_group_end:
	}

	var t *Token

	{
		_tok := p.NextToken()
		_match := _tok.ID == NUMBER_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		t = _tok
	}
	return &Number{Text: t.Text}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) columns() ([]Column, error) {
	var columns []Column

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

	{
		var alias string
		var name Name
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
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
				name = Name{}
				goto i0_group_end
			}
		}
		return []Column{{
			Name:  name,
			Alias: alias,
		}}, nil
	i0_group_end:
	}

	{
		var attr *Token
		var simples []string
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.Text == "{"
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			var _err error
			simples, _err = p.simples()
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
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto r1_i0_group_end
			}
			attr = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				attr = nil
				goto r1_i0_group_end
			}
		}
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
	r1_i0_group_end:
	}

	var name Name

	{
		var _err error
		name, _err = p.name()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	return []Column{{Name: name}}, nil
}

func (p *queryParser) name() (Name, error) {

	{
		var attr *Token
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
		// attr=IDENT
		{
			_pos4 := p.Pos()
			{
				_tok := p.NextToken()
				_match := _tok.ID == IDENT_TOKEN
				if !_match {
					p.ResetPos(_pos4)
					goto i0_i2_alt1
				}
				attr = _tok
			}
			goto i0_i2_has_match
		}

	i0_i2_alt1:
		// attr=VALUE
		{
			{
				_tok := p.NextToken()
				_match := _tok.ID == VALUE_TOKEN
				if !_match {
					p.ResetPos(_pos1)
					fn = nil
					attr = nil
					goto i0_group_end
				}
				attr = _tok
			}
		}

	i0_i2_has_match:
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				attr = nil
				goto i0_group_end
			}
		}
		return Name{
			FuncName: fn.Text,
			AttrKey:  clean(attr.Text),
		}, nil
	i0_group_end:
	}

	{
		var fn *Token
		_pos1 := p.Pos()
		{
			_tok := p.NextToken()
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			fn = _tok
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto r1_i0_group_end
			}
		}
		{
			_tok := p.NextToken()
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto r1_i0_group_end
			}
		}
		return Name{
			FuncName: fn.Text,
		}, nil
	r1_i0_group_end:
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
				goto r2_i0_alt1
			}
			t = _tok
		}
		goto r2_i0_has_match
	}

r2_i0_alt1:
	// t=VALUE
	{
		{
			_tok := p.NextToken()
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				return Name{}, errBacktrack
			}
			t = _tok
		}
	}

r2_i0_has_match:
	return Name{
		AttrKey: clean(t.Text),
	}, nil
}

func (p *queryParser) binaryOp() (BinaryOp, error) {

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

//------------------------------------------------------------------------------

func (p *queryParser) simples() ([]string, error) {
	var ss []string

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

func (p *queryParser) names() ([]Name, error) {
	var names []Name

	var name Name

	{
		var _err error
		name, _err = p.name()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		names = append(names, name)
		p.cut()
	}

	{
		var name Name
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
				name, _err = p.name()
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
				names = append(names, name)
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
