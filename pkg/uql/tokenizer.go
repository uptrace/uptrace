package uql

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/uptrace/uptrace/pkg/sqlparser"
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
	lex sqlparser.Lexer

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
	case '(', ')', ',', '|':
		start := l.lex.Pos() - 1
		l.tokens = append(l.tokens, Token{
			ID:    BYTE_TOKEN,
			Text:  l.lex.Slice(start, start+1),
			Start: start,
		})
		return &l.tokens[len(l.tokens)-1], nil
	}

	if isWhitespace(c) {
		return l.readToken()
	}
	if isDigit(c) {
		return l.number()
	}

	return l.value()
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
	s, _ := l.lex.ReadSepFunc(l.lex.Pos()-1, l.isWordSep)

	if _, err := time.ParseDuration(s); err == nil {
		return l.token(DURATION_TOKEN, s, start), nil
	}

	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return l.token(VALUE_TOKEN, s, start), nil
	}
	return l.token(NUMBER_TOKEN, s, start), nil
}

func (l *tokenizer) value() (*Token, error) {
	start := l.lex.Pos() - 1
	s, _ := l.lex.ReadSepFunc(start, l.isWordSep)
	id, s := valueToken(s)
	return l.token(id, s, start), nil
}

func (l *tokenizer) simple() (*Token, error) {
	start := l.lex.Pos()
	s, _ := l.lex.ReadSepFunc(start, l.isWordSep)
	if !isIdent(s) {
		return nil, fmt.Errorf("invalid indentifier: %q", s)
	}
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

func valueToken(s string) (TokenID, string) {
	if isIdent(s) {
		return IDENT_TOKEN, s
	}
	return VALUE_TOKEN, s
}

func (l *tokenizer) isWordSep(c byte) bool {
	if isWhitespace(c) {
		return true
	}

	switch c {
	case '(', ')', ',', '|':
		return true
	case ':':
		return isWhitespace(l.lex.PeekByte())
	}

	return false
}

//------------------------------------------------------------------------------

func isWhitespace(c byte) bool {
	switch c {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

var identRE = regexp.MustCompile(`^[[:alnum:]]+([._][[:alnum:]]+)*$`)

func isIdent(s string) bool {
	return identRE.MatchString(s)
}
