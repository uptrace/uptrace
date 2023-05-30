package tql

import (
	"fmt"
	"strings"
)

func ParsePart(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	p := &queryParser{
		lexer: newLexer(s),
	}

	expr, err := p.parseQuery()
	if err == errBacktrack {
		return nil, p.errorWithHint(s)
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

func (p *queryParser) errorWithHint(str string) error {
	const distance = 50

	if len(p.tokens) <= 1 {
		return fmt.Errorf("can't parse %q", str)
	}

	ltqlTokPos := p.cutPos
	if ltqlTokPos >= len(p.tokens) {
		ltqlTokPos = len(p.tokens) - 1
	}
	tok := &p.tokens[ltqlTokPos]

	pos := tok.Start + len(tok.Text)
	s := pos - distance
	if s < 0 {
		s = 0
	}

	e := pos + distance
	if e > len(str) {
		e = len(str)
	}

	const arrow = "<-"
	text := make([]byte, 0, e-s+len(arrow))
	text = append(text, str[s:pos]...)
	text = append(text, arrow...)
	text = append(text, str[pos:e]...)

	return fmt.Errorf("unexpected %q in %q", tok.Text, text)
}
