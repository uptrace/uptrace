package bunlex

import (
	"fmt"
	"strings"
)

type Lexer struct {
	s string
	i int
}

func (l *Lexer) Reset(s string) {
	l.s = s
	l.i = 0
}

func (l *Lexer) Valid() bool {
	return l.i < len(l.s)
}

func (l *Lexer) Pos() int {
	return l.i
}

func (l *Lexer) SetPos(pos int) {
	if pos > len(l.s) {
		panic("not reached")
	}
	l.i = pos
}

func (l *Lexer) Slice(s, e int) string {
	return l.s[s:e]
}

func (l *Lexer) Advance() {
	if l.Valid() {
		l.i++
	}
}

func (l *Lexer) Rewind() {
	if l.i > 0 {
		l.i--
	}
}

func (l *Lexer) NextByte() byte {
	var c byte
	if l.i < len(l.s) {
		c = l.s[l.i]
		l.i++
	}
	return c
}

func (l *Lexer) PeekByte() byte {
	if l.i < len(l.s) {
		return l.s[l.i]
	}
	return 0
}

func (l *Lexer) ReadSep(sep byte) (_ string, ok bool) {
	pos := l.Pos()

	for l.Valid() {
		c := l.NextByte()
		if c == sep {
			ok = true
			l.Rewind()
			break
		}
	}

	return l.s[pos:l.Pos()], ok
}

func (l *Lexer) ReadSepFunc(start int, isSep func(byte) bool) (_ string, ok bool) {
	for l.Valid() {
		c := l.NextByte()
		if isSep(c) {
			ok = true
			l.Rewind()
			break
		}
	}
	return l.s[start:l.Pos()], ok
}

func (l *Lexer) ReadUnquoted(quote byte) (string, error) {
	pos := l.Pos()
	var buf []byte

loop:
	for l.Valid() {
		c := l.NextByte()

		switch c {
		case '\\':
			switch next := l.PeekByte(); next {
			case '\\':
				l.i++
				buf = append(buf, '\\')
				continue loop
			case quote:
				l.i++
				buf = append(buf, quote)
				continue loop
			case 'n':
				l.i++
				buf = append(buf, '\n')
				continue loop
			case 'r':
				l.i++
				buf = append(buf, '\r')
				continue loop
			case 't':
				l.i++
				buf = append(buf, '\t')
				continue loop
			default:
				l.i++
				continue loop
			}
		case quote:
			return string(buf), nil
		}

		buf = append(buf, c)
	}

	if quote == '`' {
		l.SetPos(pos)
		return l.ReadUnquoted('\'')
	}

	return string(buf), syntaxError(l.s[pos:], "missing %q at the end of a string", quote)
}

func (l *Lexer) ReadQuoted(quote byte) (string, error) {
	i := strings.IndexByte(l.s[l.i:], quote) + 1
	if i == -1 {
		return "", syntaxError(l.s[l.i-1:], "missing %q at the end of a string", quote)
	}

	s := l.s[l.i-1 : l.i+i]
	if strings.IndexByte(s, '\\') == -1 {
		l.i += i
		return s, nil
	}

	return l.readQuoted(quote)
}

func (l *Lexer) readQuoted(quote byte) (string, error) {
	pos := l.i - 1

	for l.Valid() {
		c := l.NextByte()

		switch c {
		case '\\':
			l.Advance()
		case quote:
			return string(l.s[pos:l.i]), nil
		}
	}

	err := syntaxError(l.s[pos:], "missing %q at the end of a string", quote)
	return string(l.s[pos:]), err
}

func (l *Lexer) ReadQuotedSQL(quote byte) (string, error) {
	pos := l.i - 1

	for l.Valid() {
		c := l.NextByte()

		switch c {
		case '\\':
			l.Advance()
		case quote:
			if l.PeekByte() != quote {
				return string(l.s[pos:l.i]), nil
			}
			l.i++
		}
	}

	err := syntaxError(l.s[pos:], "missing %q at the end of a string", quote)
	return string(l.s[pos:]), err
}

func (l *Lexer) Number() string {
	var hasPunct bool
	var hasExp bool

	start := l.Pos()
	for l.Valid() {
		c := l.NextByte()

		switch c {
		case '.':
			if hasPunct {
				goto break_loop
			}
			hasPunct = true
		case 'e', 'E':
			if hasExp {
				goto break_loop
			}
			hasExp = true
			if l.PeekByte() == '-' {
				l.Advance()
			}
		default:
			if !IsDigit(c) {
				goto break_loop
			}
		}

		continue

	break_loop:
		l.Rewind()
		break
	}

	s := l.s[start:l.Pos()]
	return s
}

func (l *Lexer) Group(start, end byte) string {
	startPos := l.Pos()

	var level int
	for l.Valid() {
		c := l.NextByte()
		switch c {
		case '"', '\'':
			_, _ = l.ReadQuoted(c)
		case start:
			level++
		case end:
			if level == 0 {
				return l.s[startPos-1 : l.Pos()]
			}
			level--
		}
	}

	l.i = startPos
	return ""
}

//------------------------------------------------------------------------------

type SyntaxError struct {
	s   string
	msg string
}

func (e SyntaxError) Error() string {
	return e.msg + ": " + e.s
}

func syntaxError(s string, msg string, args ...any) SyntaxError {
	return SyntaxError{
		s:   s,
		msg: fmt.Sprintf(msg, args...),
	}
}
