package tql

import (
	"errors"
	"fmt"
	"strings"
)

var errAlias = errors.New("alias is required: expr AS alias")

func (p *queryParser) parseQuery() (any, error) {
	{
		var conds []Cond
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
			conds, _err = p.conds()
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
				conds = nil
				goto i0_group_end
			}
		}
		return &Where{Conds: conds}, nil
	i0_group_end:
	}

	{
		var names []Name
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				names = nil
				goto r1_i0_group_end
			}
		}
		return &Group{Names: names}, nil
	r1_i0_group_end:
	}

	{
		var columns []Name
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == EOF_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				columns = nil
				goto r2_i0_group_end
			}
		}
		return &Columns{Names: columns}, nil
	r2_i0_group_end:
	}

	var columns []Name

	{
		var _err error
		columns, _err = p.columns()
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
	return &Columns{Names: columns}, nil
}

//------------------------------------------------------------------------------

func (p *queryParser) conds() ([]Cond, error) {
	var conds []Cond

	{
		var compOp string
		var simples []string
		var value Value
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto i0_group_end
			}
		}
		{
			var _err error
			compOp, _err = p.compOp()
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
				compOp = ""
				goto i0_group_end
			}
		}
		{
			for _, attrKey := range simples {
				conds = append(conds, Cond{
					Sep:   CondSep{Op: OrOp},
					Left:  Name{AttrKey: attrKey},
					Op:    compOp,
					Right: value,
				})
			}
			return conds, nil
		}
	i0_group_end:
	}

	var cond Cond
	var not *Token

	{
		_pos1 := p.Pos()
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
		_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
		if _match {
			not = _tok
		} else {
			p.ResetPos(_pos1)
		}
	}
	{
		var _err error
		cond, _err = p.cond()
		if _err != nil && _err != errBacktrack {
			return nil, _err
		}
		_match := _err == nil
		if !_match {
			return nil, errBacktrack
		}
	}
	{
		if not != nil {
			cond.Sep.Negate = true
		}
		conds = append(conds, cond)
		p.cut()
	}

	{
		var cond Cond
		var condSep CondSep
		var _matchCount int
		for {
			_pos1 := p.Pos()
			{
				var _err error
				condSep, _err = p.condSep()
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
				cond, _err = p.cond()
				if _err != nil && _err != errBacktrack {
					return nil, _err
				}
				_match := _err == nil
				if !_match {
					p.ResetPos(_pos1)
					condSep = CondSep{}
					goto r2_i0_no_match
				}
			}
			_matchCount = _matchCount + 1
			{
				cond.Sep = condSep
				conds = append(conds, cond)
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

	return conds, nil
}

func (p *queryParser) condSep() (CondSep, error) {
	var sep CondSep

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return CondSep{}, _err
			}
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'a' || _tok.Text[0] == 'A') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N') && (_tok.Text[2] == 'd' || _tok.Text[2] == 'D')
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
		}
		sep.Op = AndOp
	i0_group_end:
	}

	if sep.Op == "" {

		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return CondSep{}, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'o' || _tok.Text[0] == 'O') && (_tok.Text[1] == 'r' || _tok.Text[1] == 'R')
			if !_match {
				return CondSep{}, errBacktrack
			}
		}
		sep.Op = OrOp
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return CondSep{}, _err
			}
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
		}
		sep.Negate = true
	r2_i0_group_end:
	}

	return sep, nil
}

func (p *queryParser) cond() (Cond, error) {
	{
		var name Name
		var values []string
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
			if _err != nil && _err != errBacktrack {
				return Cond{}, _err
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
				return Cond{}, _err
			}
			_match := len(_tok.Text) == 2 && (_tok.Text[0] == 'i' || _tok.Text[0] == 'I') && (_tok.Text[1] == 'n' || _tok.Text[1] == 'N')
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
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
				return Cond{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				values = nil
				goto i0_group_end
			}
		}
		return Cond{
			Left: name,
			Op:   InOp,
			Right: Value{
				Kind:   ArrayValue,
				Values: values,
			},
		}, nil
	i0_group_end:
	}

	{
		var compOp string
		var name Name
		var value Value
		_pos1 := p.Pos()
		{
			var _err error
			name, _err = p.name()
			if _err != nil && _err != errBacktrack {
				return Cond{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
		}
		{
			var _err error
			compOp, _err = p.compOp()
			if _err != nil && _err != errBacktrack {
				return Cond{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				goto r1_i0_group_end
			}
		}
		{
			var _err error
			value, _err = p.value()
			if _err != nil && _err != errBacktrack {
				return Cond{}, _err
			}
			_match := _err == nil
			if !_match {
				p.ResetPos(_pos1)
				name = Name{}
				compOp = ""
				goto r1_i0_group_end
			}
		}
		return Cond{
			Left:  name,
			Op:    compOp,
			Right: value,
		}, nil
	r1_i0_group_end:
	}

	{
		var key *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r2_i0_group_end
			}
			key = _tok
		}
		{
			_pos3 := p.Pos()
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
			if _match {
			} else {
				p.ResetPos(_pos3)
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
			_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
			if !_match {
				p.ResetPos(_pos1)
				key = nil
				goto r2_i0_group_end
			}
		}
		// "exist"
		{
			_pos5 := p.Pos()
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return Cond{}, _err
				}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return Cond{}, _err
				}
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					key = nil
					goto r2_i0_group_end
				}
			}
		}

	r2_i0_i3_has_match:
		return Cond{
			Left: Name{AttrKey: key.Text},
			Op:   DoesNotExistOp,
		}, nil
	r2_i0_group_end:
	}

	{
		var key *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Cond{}, _err
			}
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r3_i0_group_end
			}
			key = _tok
		}
		// "exist"
		{
			_pos3 := p.Pos()
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return Cond{}, _err
				}
				_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T')
				if !_match {
					p.ResetPos(_pos3)
					goto r3_i0_i1_alt1
				}
			}
			goto r3_i0_i1_has_match
		}

	r3_i0_i1_alt1:
		// "exists"
		{
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return Cond{}, _err
				}
				_match := len(_tok.Text) == 6 && (_tok.Text[0] == 'e' || _tok.Text[0] == 'E') && (_tok.Text[1] == 'x' || _tok.Text[1] == 'X') && (_tok.Text[2] == 'i' || _tok.Text[2] == 'I') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S') && (_tok.Text[4] == 't' || _tok.Text[4] == 'T') && (_tok.Text[5] == 's' || _tok.Text[5] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					key = nil
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_i1_has_match:
		return Cond{
			Left: Name{AttrKey: key.Text},
			Op:   ExistsOp,
		}, nil
	r3_i0_group_end:
	}

	var key *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return Cond{}, _err
		}
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return Cond{}, errBacktrack
		}
		key = _tok
	}
	return Cond{
		Left: Name{AttrKey: key.Text},
		Op:   EqualOp,
		Right: Value{
			Kind: NumberValue,
			Text: "1",
		},
	}, nil
}

func (p *queryParser) compOp() (string, error) {
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
		return ">=", nil
	i0_group_end:
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
		return "<=", nil
	r1_i0_group_end:
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
		return EqualOp, nil
	r2_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		// '!' '='
		{
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return "", _err
				}
				_match := _tok.Text == "!"
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_alt1
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
					goto r3_i0_alt1
				}
			}
			goto r3_i0_has_match
		}

	r3_i0_alt1:
		// '<' '>'
		{
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
				_match := _tok.Text == ">"
				if !_match {
					p.ResetPos(_pos1)
					goto r3_i0_group_end
				}
			}
		}

	r3_i0_has_match:
		return NotEqualOp, nil
	r3_i0_group_end:
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
				goto r4_i0_group_end
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
				goto r4_i0_group_end
			}
		}
		return DoesNotMatchOp, nil
	r4_i0_group_end:
	}

	{
		var t *Token
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := _tok.Text == "<" || _tok.Text == ">" || _tok.Text == "=" || _tok.Text == "~"
			if !_match {
				p.ResetPos(_pos1)
				goto r5_i0_group_end
			}
			t = _tok
		}
		return t.Text, nil
	r5_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
			if _match {
			} else {
				p.ResetPos(_pos1)
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return "", _err
				}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return "", _err
				}
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					goto r6_i0_group_end
				}
			}
		}

	r6_i0_i2_has_match:
		return DoesNotContainOp, nil
	r6_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		// "contain"
		{
			{
				_tok, _err := p.NextToken()
				if _err != nil {
					return "", _err
				}
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
				_tok, _err := p.NextToken()
				if _err != nil {
					return "", _err
				}
				_match := len(_tok.Text) == 8 && (_tok.Text[0] == 'c' || _tok.Text[0] == 'C') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'n' || _tok.Text[2] == 'N') && (_tok.Text[3] == 't' || _tok.Text[3] == 'T') && (_tok.Text[4] == 'a' || _tok.Text[4] == 'A') && (_tok.Text[5] == 'i' || _tok.Text[5] == 'I') && (_tok.Text[6] == 'n' || _tok.Text[6] == 'N') && (_tok.Text[7] == 's' || _tok.Text[7] == 'S')
				if !_match {
					p.ResetPos(_pos1)
					goto r7_i0_group_end
				}
			}
		}

	r7_i0_has_match:
		return ContainsOp, nil
	r7_i0_group_end:
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
				goto r8_i0_group_end
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
				goto r8_i0_group_end
			}
		}
		return NotLikeOp, nil
	r8_i0_group_end:
	}

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
				goto r9_i0_group_end
			}
		}
		return LikeOp, nil
	r9_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'm' || _tok.Text[0] == 'M') && (_tok.Text[1] == 'a' || _tok.Text[1] == 'A') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T') && (_tok.Text[3] == 'c' || _tok.Text[3] == 'C') && (_tok.Text[4] == 'h' || _tok.Text[4] == 'H') && (_tok.Text[5] == 'e' || _tok.Text[5] == 'E') && (_tok.Text[6] == 's' || _tok.Text[6] == 'S')
			if !_match {
				p.ResetPos(_pos1)
				goto r10_i0_group_end
			}
		}
		return MatchesOp, nil
	r10_i0_group_end:
	}

	{
		_pos1 := p.Pos()
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
		_match := len(_tok.Text) == 4 && (_tok.Text[0] == 'd' || _tok.Text[0] == 'D') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 'e' || _tok.Text[2] == 'E') && (_tok.Text[3] == 's' || _tok.Text[3] == 'S')
		if _match {
		} else {
			p.ResetPos(_pos1)
		}
	}
	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return "", _err
		}
		_match := len(_tok.Text) == 3 && (_tok.Text[0] == 'n' || _tok.Text[0] == 'N') && (_tok.Text[1] == 'o' || _tok.Text[1] == 'O') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T')
		if !_match {
			return "", errBacktrack
		}
	}
	// "match"
	{
		_pos3 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 5 && (_tok.Text[0] == 'm' || _tok.Text[0] == 'M') && (_tok.Text[1] == 'a' || _tok.Text[1] == 'A') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T') && (_tok.Text[3] == 'c' || _tok.Text[3] == 'C') && (_tok.Text[4] == 'h' || _tok.Text[4] == 'H')
			if !_match {
				p.ResetPos(_pos3)
				goto r11_i2_alt1
			}
		}
		goto r11_i2_has_match
	}

r11_i2_alt1:
	// "matches"
	{
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return "", _err
			}
			_match := len(_tok.Text) == 7 && (_tok.Text[0] == 'm' || _tok.Text[0] == 'M') && (_tok.Text[1] == 'a' || _tok.Text[1] == 'A') && (_tok.Text[2] == 't' || _tok.Text[2] == 'T') && (_tok.Text[3] == 'c' || _tok.Text[3] == 'C') && (_tok.Text[4] == 'h' || _tok.Text[4] == 'H') && (_tok.Text[5] == 'e' || _tok.Text[5] == 'E') && (_tok.Text[6] == 's' || _tok.Text[6] == 'S')
			if !_match {
				return "", errBacktrack
			}
		}
	}

r11_i2_has_match:
	return DoesNotMatchOp, nil
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
			_match := _tok.ID == NUMBER_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto i0_group_end
			}
			t = _tok
		}
		return Value{
			Kind: NumberValue,
			Text: t.Text,
		}, nil
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
			_match := _tok.ID == DURATION_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				goto r1_i0_group_end
			}
			t = _tok
		}
		return Value{
			Kind: DurationValue,
			Text: t.Text,
		}, nil
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
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Value{}, _err
			}
			_match := _tok.ID == VALUE_TOKEN
			if !_match {
				return Value{}, errBacktrack
			}
			t = _tok
		}
	}

r2_i0_has_match:
	return Value{
		Kind: StringValue,
		Text: t.Text,
	}, nil
}

//------------------------------------------------------------------------------

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

func (p *queryParser) name() (Name, error) {
	{
		var attr *Token
		var fn *Token
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
			fn = _tok
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.ID == IDENT_TOKEN
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				goto i0_group_end
			}
			attr = _tok
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return Name{}, _err
			}
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				fn = nil
				attr = nil
				goto i0_group_end
			}
		}
		{
			funcName := strings.ToLower(fn.Text)
			switch funcName {
			case "p50", "p75", "p90", "p95", "p99",
				"min", "max", "sum", "avg",
				"top3", "top10",
				"any", "uniq":
				return Name{
					FuncName: funcName,
					AttrKey:  attr.Text,
				}, nil
			default:
				return Name{}, fmt.Errorf("unknown function: %q", fn.Text)
			}
		}
	i0_group_end:
	}

	var t *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return Name{}, _err
		}
		_match := _tok.ID == IDENT_TOKEN
		if !_match {
			return Name{}, errBacktrack
		}
		t = _tok
	}
	return Name{
		AttrKey: t.Text,
	}, nil
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

//------------------------------------------------------------------------------

func (p *queryParser) simples() ([]string, error) {
	var ss []string

	var t *Token

	{
		_tok, _err := p.NextToken()
		if _err != nil {
			return nil, _err
		}
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

func (p *queryParser) columns() ([]Name, error) {
	var columns []Name

	var column []Name

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
		var column []Name
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

func (p *queryParser) column() ([]Name, error) {
	{
		var attr *Token
		var simples []string
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == "}"
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto i0_group_end
			}
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == "("
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				goto i0_group_end
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
				simples = nil
				goto i0_group_end
			}
			attr = _tok
		}
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.Text == ")"
			if !_match {
				p.ResetPos(_pos1)
				simples = nil
				attr = nil
				goto i0_group_end
			}
		}
		{
			columns := make([]Name, len(simples))
			for i, funcName := range simples {
				columns[i] = Name{
					FuncName: funcName,
					AttrKey:  attr.Text,
				}
			}
			return columns, nil
		}
	i0_group_end:
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
	return []Name{name}, nil
}

func (p *queryParser) values() ([]string, error) {
	var ss []string

	var t *Token

	// t=IDENT
	{
		_pos1 := p.Pos()
		{
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
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
			_tok, _err := p.NextToken()
			if _err != nil {
				return nil, _err
			}
			_match := _tok.ID == NUMBER_TOKEN
			if !_match {
				return nil, errBacktrack
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
			// t=IDENT
			{
				_pos3 := p.Pos()
				{
					_tok, _err := p.NextToken()
					if _err != nil {
						return nil, _err
					}
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
					_tok, _err := p.NextToken()
					if _err != nil {
						return nil, _err
					}
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
					_tok, _err := p.NextToken()
					if _err != nil {
						return nil, _err
					}
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
			return nil, errBacktrack
		}
	}

	return ss, nil
}
