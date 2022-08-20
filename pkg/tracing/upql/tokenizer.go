package upql

import (
	"errors"
	"strconv"
	"time"

	"github.com/uptrace/uptrace/pkg/bunlex"
)

//go:generate parseme -struct=queryParser

///go:generate stringer -type=TokenID
type TokenID int8

const (
	EOF_TOKEN TokenID = iota
	BYTE_TOKEN
	IDENT_TOKEN
	NUMBER_TOKEN
	DURATION_TOKEN
	BYTES_TOKEN
	VALUE_TOKEN
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
		return t.Text
	}
	return t.ID.String()
}

//------------------------------------------------------------------------------

type tokenizer struct {
	lex bunlex.Lexer

	tokens []Token
	pos    int
}

func newTokenizer(s string) *tokenizer {
	t := &tokenizer{
		tokens: make([]Token, 0, 32),
	}
	t.Reset(s)
	return t
}

func (l *tokenizer) Reset(s string) {
	l.lex.Reset(s)
	l.tokens = l.tokens[:0]
	l.pos = 0
}

func (l *tokenizer) NextToken() (*Token, error) {
	tok, err := l.PeekToken()
	if err != nil {
		return nil, err
	}
	l.pos++
	return tok, nil
}

func (l *tokenizer) PeekToken() (*Token, error) {
	if l.pos < len(l.tokens) {
		return &l.tokens[l.pos], nil
	}
	return l.readToken()
}

func (l *tokenizer) Pos() int {
	return l.pos
}

func (l *tokenizer) ResetPos(pos int) {
	l.pos = pos
}

func (l *tokenizer) readToken() (*Token, error) {
	if !l.lex.Valid() {
		return eofToken, nil
	}

	c := l.lex.NextByte()

	switch c {
	case '\'', '"':
		return l.quotedValue(c)
	case '-', '+':
		return l.number()
	case '(', ')', '{', '}', ',', '|':
		start := l.lex.Pos() - 1
		return l.token(BYTE_TOKEN, l.lex.Slice(start, start+1), start), nil
	}

	if bunlex.IsWhitespace(c) {
		return l.readToken()
	}

	if bunlex.IsAlpha(c) {
		return l.ident(l.lex.Pos() - 1)
	}
	if bunlex.IsDigit(c) {
		return l.number()
	}

	return l.charToken(BYTE_TOKEN), nil
}

func (l *tokenizer) charToken(id TokenID) *Token {
	pos := l.lex.Pos()
	return l.token(id, l.lex.Slice(pos-1, pos), pos-1)
}

func (l *tokenizer) quotedValue(end byte) (*Token, error) {
	start := l.lex.Pos() - 1

	s, err := l.lex.ReadUnquoted(end)
	if err != nil {
		return nil, err
	}

	l.tokens = append(l.tokens, Token{
		ID:    VALUE_TOKEN,
		Text:  s,
		Start: start,
	})
	return &l.tokens[len(l.tokens)-1], nil
}

func (l *tokenizer) number() (*Token, error) {
	start := l.lex.Pos() - 1
	s, _ := l.lex.ReadSepFunc(l.lex.Pos()-1, l.isWordBoundary)

	if _, err := time.ParseDuration(s); err == nil {
		return l.token(DURATION_TOKEN, s, start), nil
	}

	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return l.token(VALUE_TOKEN, s, start), nil
	}
	return l.token(NUMBER_TOKEN, s, start), nil
}

func (l *tokenizer) ident(start int) (*Token, error) {
	for l.lex.Valid() {
		c := l.lex.PeekByte()
		if !isIdent(c) {
			break
		}
		l.lex.Advance()
	}

	s := l.lex.Slice(start, l.lex.Pos())
	return l.token(IDENT_TOKEN, s, start), nil
}

func (l *tokenizer) token(id TokenID, s string, start int) *Token {
	l.tokens = append(l.tokens, Token{
		ID:    id,
		Text:  s,
		Start: start,
	})
	return &l.tokens[len(l.tokens)-1]
}

func (l *tokenizer) isWordBoundary(c byte) bool {
	if bunlex.IsWhitespace(c) {
		return true
	}

	switch c {
	case '(', ')', '{', '}', ',', '|':
		return true
	case ':':
		return bunlex.IsWhitespace(l.lex.PeekByte())
	}

	return false
}

func isIdent(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_' || c == '.'
}
