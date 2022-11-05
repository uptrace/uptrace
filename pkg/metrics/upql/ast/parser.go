package ast

import (
	"errors"
	"fmt"
)

var errAlias = errors.New("alias is required (AS alias)")

func (p *queryParser) parseQuery() (any, error) {
	{
		var where []Filter
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == IDENT_TOKEN && len(_tok.Text) == 3 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 'l' || _tok.Text[1] == 'L') && (_tok.Text[2] == 'l' || _tok.Text[2] == 'L')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return &Grouping{GroupByAll: true}, nil
	r1_i0_group_end:
	}

	{
		var grouping []string
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
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
				goto r2_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				grouping = nil
				goto r2_i0_group_end
			}
		}
		return &Grouping{Names: grouping}, nil
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r3_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r3_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == IDENT_TOKEN && len(_tok.Text) == 3 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 'l' || _tok.Text[1] == 'L') && (_tok.Text[2] == 'l' || _tok.Text[2] == 'L')
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r3_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r3_i0_group_end
			}
		}
		return &Selector{
			Expr:       namedExpr,
			GroupByAll: true,
		}, nil
	r3_i0_group_end:
	}

	{
		var grouping []string
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
				goto r4_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'g' || _tok.Text[0] == 'G') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R') && (_tok.Text[2] == 'o' || _tok.Text[2] == 'O') && (_tok.Text[3] == 'u' || _tok.Text[3] == 'U') && (_tok.Text[4] == 'p' || _tok.Text[4] == 'P')
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r4_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'b' || _tok.Text[0] == 'B') && (_tok.Text[1] == 'y' || _tok.Text[1] == 'Y')
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r4_i0_group_end
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
				goto r4_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				grouping = nil
				goto r4_i0_group_end
			}
		}
		return &Selector{
			Expr:     namedExpr,
			Grouping: grouping,
		}, nil
	r4_i0_group_end:
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
				goto r5_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				namedExpr = NamedExpr{}
				goto r5_i0_group_end
			}
		}
		return &Selector{Expr: namedExpr}, nil
	r5_i0_group_end:
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
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return nil, _err
				}
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
				filter.Sep = BoolAnd
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
		switch f.Op {
		case FilterEqual, FilterNotEqual, FilterRegexp, FilterNotRegexp:
			// ok
		default:
			return nil, fmt.Errorf("only =, !=, ~, and !~ are allowed inside curly brackets")
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
				filter.Sep = boolOp
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

	return filters, nil
}

func (p *queryParser) filter() (Filter, error) {
	var filterOp FilterOp
	var lhs *Token
	var value Value

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return Filter{}, _err
		}
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
		LHS: lhs.Text,
		Op:  filterOp,
		RHS: value,
	}, nil
}

func (p *queryParser) filterOp() (FilterOp, error) {
	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		return FilterLike, nil
	i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'l' || _tok.Text[0] == 'L') && (_tok.Text[1] == 'i' || _tok.Text[1] == 'I') && (_tok.Text[2] == 'k' || _tok.Text[2] == 'K') && (_tok.Text[3] == 'e' || _tok.Text[3] == 'E')
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		return FilterNotLike, nil
	r1_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "!"
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		return FilterNotEqual, nil
	r2_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "!"
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		return FilterNotRegexp, nil
	r3_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		return FilterEqual, nil
	r4_i0_group_end:
	}

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
		_match := _tok.Text == "~"
		if !_match {
			return "", errBacktrack
		}
	}
	return FilterRegexp, nil
}

func (p *queryParser) value() (Value, error) {
	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
			_match := _tok.ID == DURATION_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			t = _tok
		}
		return Value{Text: t.Text, Kind: ValueDuration}, nil
	i0_group_end:
	}

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
			_match := _tok.ID == BYTES_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			t = _tok
		}
		return Value{Text: t.Text, Kind: ValueBytes}, nil
	r1_i0_group_end:
	}

	var t *Token

	// t=IDENT
	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
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
		_pos3 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				p.ResetPos(_pos3)
				goto r2_i0_alt2
			}
			t = _tok
		}
		goto r2_i0_has_match
	}

r2_i0_alt2:
	// t=NUMBER
	{
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
			_match := _tok.ID == NUMBER_TOKEN
			if !_match {
				return Value{}, errBacktrack
			}
			t = _tok
		}
	}

r2_i0_has_match:
	return Value{Text: t.Text}, nil
}

func (p *queryParser) boolOp() (BoolOp, error) {
	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return nil, _err
				}
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
		var funcCall *FuncCall
		_pos1 := p.Pos()
		{
			var _err error
			funcCall, _err = p.funcCall()
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
				funcCall = nil
				goto i0_group_end
			}
		}
		return NamedExpr{
			Expr:  funcCall,
			Alias: alias,
		}, nil
	i0_group_end:
	}

	{
		var alias string
		var name Name
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
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
				name = Name{}
				goto r1_i0_group_end
			}
		}
		return NamedExpr{
			Expr:  &name,
			Alias: alias,
		}, nil
	r1_i0_group_end:
	}

	{
		var alias string
		var filteredName *Name
		_pos1 := p.Pos()
		{
			var _err error
			filteredName, _err = p.filteredName()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
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
				filteredName = nil
				goto r2_i0_group_end
			}
		}
		return NamedExpr{
			Expr:  filteredName,
			Alias: alias,
		}, nil
	r2_i0_group_end:
	}

	{
		var alias string
		var binaryExpr *BinaryExpr
		_pos1 := p.Pos()
		{
			var _err error
			binaryExpr, _err = p.binaryExpr()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
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
				binaryExpr = nil
				goto r3_i0_group_end
			}
		}
		return NamedExpr{
			Expr:  binaryOpPrecedence(binaryExpr),
			Alias: alias,
		}, nil
	r3_i0_group_end:
	}

	{
		var binaryExpr *BinaryExpr
		_pos1 := p.Pos()
		{
			var _err error
			binaryExpr, _err = p.binaryExpr()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		return NamedExpr{
			Expr: binaryOpPrecedence(binaryExpr),
		}, nil
	r4_i0_group_end:
	}

	{
		var funcCall *FuncCall
		_pos1 := p.Pos()
		{
			var _err error
			funcCall, _err = p.funcCall()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
		}
		return NamedExpr{
			Expr: funcCall,
		}, nil
	r5_i0_group_end:
	}

	{
		var filteredName *Name
		_pos1 := p.Pos()
		{
			var _err error
			filteredName, _err = p.filteredName()
			if _err != nil && _err != errBacktrack {
				return NamedExpr{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r6_i0_group_end
			}
		}
		return NamedExpr{
			Expr: filteredName,
		}, nil
	r6_i0_group_end:
	}

	var name Name

	{
		var _err error
		name, _err = p.name()
		if _err != nil && _err != errBacktrack {
			return NamedExpr{}, _err
		}
		_match := _err == nil
		if !_match {
			return NamedExpr{}, errBacktrack
		}
	}
	return NamedExpr{
		Expr: &name,
	}, nil
}

func (p *queryParser) name() (Name, error) {
	{
		var name *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			name = _tok
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.Text == "{"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				name = nil
				goto i0_group_end
			}
		}
		return Name{
			Name: name.Text,
		}, nil
	i0_group_end:
	}

	var name *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return Name{}, _err
		}
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return Name{}, errBacktrack
		}
		name = _tok
	}
	return Name{
		Name: name.Text,
	}, nil
}

func (p *queryParser) filteredName() (*Name, error) {
	var filters []Filter
	var name *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		name = _tok
	}
	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.Text == "{"
		if !_match {
			return nil, errBacktrack
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
			return nil, errBacktrack
		}
	}
	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.Text == "}"
		if !_match {
			return nil, errBacktrack
		}
	}
	return &Name{
		Name:    name.Text,
		Filters: filters,
	}, nil
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

	return expr, nil
}

func (p *queryParser) term() (Expr, error) {
	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == DURATION_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			t = _tok
		}
		return &Number{Text: t.Text, Kind: ValueDuration}, nil
	i0_group_end:
	}

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == BYTES_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			t = _tok
		}
		return &Number{Text: t.Text, Kind: ValueBytes}, nil
	r1_i0_group_end:
	}

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == NUMBER_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
			t = _tok
		}
		return &Number{Text: t.Text}, nil
	r2_i0_group_end:
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
				goto r3_i0_group_end
			}
		}
		return funcCall, nil
	r3_i0_group_end:
	}

	{
		var filteredName *Name
		_pos1 := p.Pos()
		{
			var _err error
			filteredName, _err = p.filteredName()
			if _err != nil && _err != errBacktrack {
				return nil, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r4_i0_group_end
			}
		}
		return filteredName, nil
	r4_i0_group_end:
	}

	{
		var binaryExpr *BinaryExpr
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
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
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				binaryExpr = nil
				goto r5_i0_group_end
			}
		}
		return ParenExpr{Expr: binaryExpr}, nil
	r5_i0_group_end:
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
	return &name, nil
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

func (p *queryParser) binaryOp() (BinaryOp, error) {
	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "="
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "!"
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == ">"
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "<"
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
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
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
		_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 's' || _tok.Text[1] == 'S')
		if !_match {
			return "", errBacktrack
		}
	}
	tok, err := p.NextToken()
	if err != nil {
		return "", err
	}
	if tok.ID != IDENT_TOKEN {
		return "", errAlias
	}
	return tok.Text, nil
}

func (p *queryParser) funcCall() (*FuncCall, error) {
	var args []Expr
	var fn *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return nil, errBacktrack
		}
		fn = _tok
	}
	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.Text == "("
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		var _err error
		args, _err = p.args()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := _tok.Text == ")"
		if !_match {
			return nil, errBacktrack
		}
	}
	return &FuncCall{
		Func: fn.Text,
		Args: args,
	}, nil
}

func (p *queryParser) args() ([]Expr, error) {
	var args []Expr

	var term Expr

	{
		var _err error
		term, _err = p.term()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		args = append(args, term)
		p.cut()
	}

	{
		var term Expr
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return nil, _err
				}
				_match := _tok.Text == ","
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			{
				var _err error
				term, _err = p.term()
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
				args = append(args, term)
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

func (p *queryParser) grouping() ([]string, error) {
	var names []string

	var name *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return nil, _err
				}
				_match := _tok.Text == ","
				if !_match {
					p.ResetPos(_pos1)
					goto r1_i0_no_match
				}
			}
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return nil, _err
				}
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
