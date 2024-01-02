package chquery

import (
	"regexp"
	"strconv"

	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

//go:generate stringer -type=TokenID
type TokenID int8

const (
	EOF_TOKEN TokenID = iota
	INCLUDE_TOKEN
	EXCLUDE_TOKEN
	REGEXP_TOKEN
)

var eofToken = &Token{ID: EOF_TOKEN}

type Tokens []Token

func (tokens Tokens) String() string {
	var b []byte
	for i := range tokens {
		if i > 0 {
			b = append(b, ' ')
		}
		b = tokens[i].AppendString(b)
	}
	return unsafeconv.String(b)
}

type Token struct {
	ID     TokenID
	Values []string
}

func (t *Token) AppendString(b []byte) []byte {
	switch t.ID {
	case INCLUDE_TOKEN:
		b = appendValues(b, t.Values)
		return b
	case EXCLUDE_TOKEN:
		b = append(b, '-')
		b = appendValues(b, t.Values)
		return b
	case REGEXP_TOKEN:
		b = append(b, '~')
		b = strconv.AppendQuote(b, t.Values[0])
		return b
	default:
		b = append(b, t.ID.String()...)
		return b
	}
}

func appendValues(b []byte, values []string) []byte {
	for i, value := range values {
		if i > 0 {
			b = append(b, '|')
		}
		b = strconv.AppendQuote(b, value)
	}
	return b
}

//------------------------------------------------------------------------------

type lexer struct {
	s   string
	lex bunlex.Lexer

	tokens []Token
}

func newLexer(s string) *lexer {
	lex := &lexer{
		tokens: make([]Token, 0, 32),
	}
	lex.Reset(s)
	return lex
}

func (l *lexer) Reset(s string) {
	l.tokens = l.tokens[:0]

	l.s = s
	l.lex.Reset(s)
}

func (l *lexer) NextToken() (*Token, error) {
	tok, err := l.readToken()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return eofToken, nil
	}
	return tok, nil
}

func (l *lexer) readToken() (*Token, error) {
	ch := l.lex.NextByte()
	if ch == 0 {
		return nil, nil
	}

	if bunlex.IsWhitespace(ch) {
		return l.readToken()
	}

	switch ch {
	case '-':
		values := l.alts(l.lex.NextByte())
		return l.token(EXCLUDE_TOKEN, values), nil
	case '~':
		pattern := l.regexp(l.lex.NextByte())
		if _, err := regexp.Compile(pattern); err != nil {
			return nil, err
		}
		return l.token(REGEXP_TOKEN, []string{pattern}), nil
	default:
		values := l.alts(ch)
		return l.token(INCLUDE_TOKEN, values), nil
	}
}

func (l *lexer) alts(ch byte) []string {
	var values []string
	for {
		value := l.wordOrPhrase(ch)
		values = append(values, value)

		if !l.lex.Valid() {
			return values
		}

		if l.lex.NextByte() == '|' {
			ch = l.lex.NextByte()
			continue
		}

		l.lex.Rewind()
		return values
	}
}

func (l *lexer) wordOrPhrase(ch byte) string {
	switch ch {
	case '\'', '"':
		s, _ := l.lex.ReadUnquoted(ch)
		return s
	}

	start := l.lex.Pos() - 1
	for l.lex.Valid() {
		ch := l.lex.PeekByte()
		if isWordBoundary(ch) {
			break
		}
		l.lex.Advance()
	}
	return l.s[start:l.lex.Pos()]
}

func (l *lexer) regexp(ch byte) string {
	switch ch {
	case '\'', '"':
		s, _ := l.lex.ReadUnquoted(ch)
		return s
	}

	start := l.lex.Pos() - 1
	for l.lex.Valid() {
		ch := l.lex.PeekByte()
		if bunlex.IsWhitespace(ch) {
			break
		}
		l.lex.Advance()
	}
	return l.s[start:l.lex.Pos()]
}

func (l *lexer) token(id TokenID, values []string) *Token {
	l.tokens = append(l.tokens, Token{
		ID:     id,
		Values: values,
	})
	return &l.tokens[len(l.tokens)-1]
}

func isWordBoundary(ch byte) bool {
	return bunlex.IsWhitespace(ch) || ch == '|'
}
