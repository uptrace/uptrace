package logparser

type TokenType int8

const (
	InvalidToken TokenType = iota
	WordToken
	QuotedToken
	ParamToken
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
	lex Lexer
}

func NewTokenizer(s string) *Tokenizer {
	t := &Tokenizer{
		s: s,
	}
	t.lex.Reset(s)
	return t
}

func (t *Tokenizer) NextToken() Token {
	c := t.lex.NextByte()
	if c == 0 {
		return Token{}
	}

	switch c {
	case '"', '`', '\'':
		s, _ := t.lex.ReadQuoted(c)
		return Token{Type: QuotedToken, Text: s}
	}

	part := t.readPart(t.lex.Pos() - 1)
	if isWord(part) {
		return Token{Type: WordToken, Text: part}
	}
	return Token{Type: ParamToken, Text: part}
}

func (t *Tokenizer) readPart(start int) string {
loop:
	for t.lex.Valid() {
		c := t.lex.NextByte()

		switch c {
		case '"':
			_, _ = t.lex.ReadQuoted('"')
			continue loop
		case '\'':
			_, _ = t.lex.ReadQuoted('\'')
			continue loop
		case '{':
			_, _ = t.lex.Group('{', '}')
			continue loop
		case '<':
			_, _ = t.lex.Group('<', '>')
			continue loop
		}

		if t.isWordSep(c) {
			t.lex.Rewind()
			break
		}
	}
	s := t.s[start:t.lex.Pos()]
	return s
}

func (t *Tokenizer) isWordSep(c byte) bool {
	if isWhitespace(c) {
		return true
	}

	switch c {
	case ',', ';':
		return true
	case '.', ':':
		return isWhitespace(t.lex.PeekByte())
	}

	return false
}

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

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isWord(s string) bool {
	for _, c := range []byte(s) {
		if !(isAlpha(c) || c == '-') {
			return false
		}
	}
	return true
}
