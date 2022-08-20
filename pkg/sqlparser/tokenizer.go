package sqlparser

import (
	"io"

	"github.com/uptrace/uptrace/pkg/bunlex"
)

type TokenType int8

const (
	InvalidToken TokenType = iota
	SpaceToken
	CharToken
	IdentToken
	QuotedIdentToken
	ValueToken
	NumberToken
)

type Token struct {
	Type TokenType
	Text string
}

func newToken(typ TokenType, text string) Token {
	return Token{Type: typ, Text: text}
}

type Tokenizer struct {
	s   string
	lex bunlex.Lexer
}

func NewTokenizer(s string) *Tokenizer {
	t := &Tokenizer{
		s: s,
	}
	t.lex.Reset(s)
	return t
}

func (t *Tokenizer) NextToken() (Token, error) {
	c := t.lex.NextByte()
	if c == 0 {
		return Token{}, io.EOF
	}

	switch c {
	case '\'':
		return t.value(c)
	case '"', '`':
		return t.quotedIdent(c)
	case '_', '?':
		if isIdent(t.lex.PeekByte()) {
			return t.ident(), nil
		}
	}

	if bunlex.IsDigit(c) {
		t.lex.Rewind()
		return t.number(), nil
	}
	if bunlex.IsAlpha(c) {
		return t.ident(), nil
	}
	if bunlex.IsWhitespace(c) {
		return t.byteToken(SpaceToken), nil
	}

	return t.byteToken(CharToken), nil
}

func (t *Tokenizer) ident() Token {
	start := t.lex.Pos() - 1
	for t.lex.Valid() {
		c := t.lex.PeekByte()
		if !isIdent(c) {
			break
		}
		t.lex.Advance()
	}

	s := t.s[start:t.lex.Pos()]
	return newToken(IdentToken, s)
}

func (t *Tokenizer) byteToken(typ TokenType) Token {
	pos := t.lex.Pos()
	return newToken(typ, t.s[pos-1:pos])
}

func (t *Tokenizer) value(end byte) (Token, error) {
	s, err := t.lex.ReadQuotedSQL(end)
	if err != nil {
		return Token{}, err
	}
	return newToken(ValueToken, s), nil
}

func (t *Tokenizer) number() Token {
	return newToken(NumberToken, t.lex.Number())
}

func (t *Tokenizer) quotedIdent(end byte) (Token, error) {
	s, err := t.lex.ReadQuotedSQL(end)
	if err != nil {
		return Token{}, err
	}
	return newToken(QuotedIdentToken, s), nil
}

func isIdent(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_'
}
