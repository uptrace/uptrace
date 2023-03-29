package ast

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/bununit"
)

type TokenID int8

const (
	EOF_TOKEN TokenID = iota
	BYTE_TOKEN
	IDENT_TOKEN
	VALUE_TOKEN
	NUMBER_TOKEN
	DURATION_TOKEN
	BYTES_TOKEN
)

var eofToken = &Token{ID: EOF_TOKEN}

var errBacktrack = errors.New("backtrack")

type Token struct {
	ID    TokenID
	Text  string
	Start int
}

func (t *Token) String() string {
	if t.Text != "" {
		return fmt.Sprintf("%s(%s)", t.ID.String(), t.Text)
	}
	return t.ID.String()
}

//------------------------------------------------------------------------------

type lexer struct {
	s   string
	lex bunlex.Lexer

	ignoreSpaces bool

	tokens []Token
	pos    int
}

func newLexer(s string) *lexer {
	lex := &lexer{
		tokens: make([]Token, 0, 32),
	}
	lex.Reset(s)
	return lex
}

func (l *lexer) Reset(s string) {
	l.ignoreSpaces = false

	l.tokens = l.tokens[:0]
	l.pos = 0

	l.s = s
	l.lex.Reset(s)
}

func (l *lexer) IgnoreSpaces() *lexer {
	l.ignoreSpaces = true
	return l
}

func (l *lexer) NextToken() (*Token, error) {
	tok, err := l.PeekToken()
	if err != nil {
		return nil, err
	}
	l.pos++
	return tok, nil
}

func (l *lexer) PeekToken() (*Token, error) {
	if l.pos < len(l.tokens) {
		return &l.tokens[l.pos], nil
	}
	return l.readToken()
}

func (l *lexer) Pos() int {
	return l.pos
}

func (l *lexer) ResetPos(pos int) {
	l.pos = pos
}

func (l *lexer) readToken() (*Token, error) {
	if !l.lex.Valid() {
		return eofToken, nil
	}

	c := l.lex.NextByte()

	switch c {
	case '\'', '"':
		return l.quotedValue(c)
	case '(', ')', ',':
		return l.charToken(BYTE_TOKEN), nil
	case '_':
		return l.ident(l.lex.Pos() - 1)
	case '$':
		return l.ident(l.lex.Pos() - 1)
	}

	if l.ignoreSpaces && bunlex.IsWhitespace(c) {
		return l.readToken()
	}

	if bunlex.IsDigit(c) {
		l.lex.Rewind()
		return l.number(), nil
	}
	if bunlex.IsAlpha(c) {
		return l.ident(l.lex.Pos() - 1)
	}

	return l.charToken(BYTE_TOKEN), nil
}

func (l *lexer) charToken(id TokenID) *Token {
	pos := l.lex.Pos()
	return l.token(id, l.s[pos-1:pos], pos-1)
}

func (l *lexer) quotedValue(end byte) (*Token, error) {
	start := l.lex.Pos() - 1
	s, err := l.lex.ReadUnquoted(end)
	if err != nil {
		return nil, err
	}
	return l.token(VALUE_TOKEN, s, start), nil
}

func (l *lexer) number() *Token {
	start := l.lex.Pos()
	s, _ := l.lex.ReadSepFunc(start, l.isWordBoundary)
	if _, err := time.ParseDuration(s); err == nil {
		return l.token(DURATION_TOKEN, s, start)
	}
	if _, err := bununit.ParseBytes(s); err == nil {
		return l.token(BYTES_TOKEN, s, start)
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return l.token(NUMBER_TOKEN, s, start)
	}
	return l.token(VALUE_TOKEN, s, start)
}

func (l *lexer) isWordBoundary(c byte) bool {
	if bunlex.IsWhitespace(c) {
		return true
	}

	switch c {
	case '+', '-', '/', '%', '*', '=', '<', '>', '!',
		'(', ')', '{', '}', ',', '|':
		return true
	default:
		return false
	}
}

func (l *lexer) ident(start int) (*Token, error) {
	for l.lex.Valid() {
		c := l.lex.PeekByte()
		if !isIdentChar(c) {
			break
		}
		l.lex.Advance()
	}

	s := l.s[start:l.lex.Pos()]
	return l.token(IDENT_TOKEN, s, start), nil
}

func (l *lexer) token(id TokenID, s string, start int) *Token {
	l.tokens = append(l.tokens, Token{
		ID:    id,
		Text:  s,
		Start: start,
	})
	return &l.tokens[len(l.tokens)-1]
}

func IsIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range []byte(s) {
		if !isIdentChar(c) {
			return false
		}
	}
	return true
}

func isIdentChar(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_' || c == '.'
}
