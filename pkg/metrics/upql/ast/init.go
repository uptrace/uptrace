package ast

import (
	"errors"
	"fmt"
)

func Parse(s string) (any, error) {
	if s == "" {
		return nil, errors.New("query is empty")
	}

	p := &queryParser{
		lexer: newLexer(s),
	}

	expr, err := p.parseQuery()
	if err == errBacktrack {
		err = p.errorWithHint()
	}
	return expr, err
}

type queryParser struct {
	*lexer
	cutPos int
}

func (p *queryParser) cut() {
	p.cutPos = p.pos + 1
}

func (p *queryParser) errorWithHint() error {
	const distance = 50

	if len(p.tokens) <= 1 {
		return fmt.Errorf("can't parse %q", p.s)
	}

	lastTokPos := p.cutPos
	if lastTokPos >= len(p.tokens) {
		lastTokPos = len(p.tokens) - 1
	}
	tok := &p.tokens[lastTokPos]

	pos := tok.Start + len(tok.Text)
	s := pos - distance
	if s < 0 {
		s = 0
	}

	e := pos + distance
	if e > len(p.s) {
		e = len(p.s)
	}

	const arrow = "<-"
	text := make([]byte, 0, e-s+len(arrow))
	text = append(text, p.s[s:pos]...)
	text = append(text, arrow...)
	text = append(text, p.s[pos:e]...)

	return fmt.Errorf("unexpected %q in %q", tok.Text, text)
}
