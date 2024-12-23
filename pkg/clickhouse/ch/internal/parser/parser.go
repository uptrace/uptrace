package parser

import (
	"bytes"
	"github.com/uptrace/pkg/unsafeconv"
	"strconv"
)

var (
	digit [256]bool
	alpha [256]bool
)

func init() {
	for c := '0'; c <= '9'; c++ {
		digit[c] = true
	}
	for c := 'a'; c <= 'z'; c++ {
		alpha[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		alpha[c] = true
	}
}

type Parser struct {
	b []byte
	i int
}

func New(b []byte) *Parser       { return &Parser{b: b} }
func NewString(s string) *Parser { return New(unsafeconv.Bytes(s)) }
func (p *Parser) Valid() bool    { return p.i < len(p.b) }
func (p *Parser) Bytes() []byte  { return p.b[p.i:] }
func (p *Parser) ReadByte() byte {
	if p.Valid() {
		c := p.b[p.i]
		p.Advance()
		return c
	}
	return 0
}
func (p *Parser) PeekByte() byte {
	if p.Valid() {
		return p.b[p.i]
	}
	return 0
}
func (p *Parser) Advance() { p.i++ }
func (p *Parser) Skip(skip byte) bool {
	if p.PeekByte() == skip {
		p.Advance()
		return true
	}
	return false
}
func (p *Parser) SkipBytes(skip []byte) bool {
	if len(skip) > len(p.b[p.i:]) {
		return false
	}
	if !bytes.Equal(p.b[p.i:p.i+len(skip)], skip) {
		return false
	}
	p.i += len(skip)
	return true
}
func (p *Parser) ReadSep(sep byte) ([]byte, bool) {
	ind := bytes.IndexByte(p.b[p.i:], sep)
	if ind == -1 {
		b := p.b[p.i:]
		p.i = len(p.b)
		return b, false
	}
	b := p.b[p.i : p.i+ind]
	p.i += ind + 1
	return b, true
}
func (p *Parser) ReadIdent() (string, bool) {
	switch start, endCh := p.i, byte('}'); p.PeekByte() {
	case '(':
		endCh = ')'
		fallthrough
	case '{':
		p.Advance()
		name, numeric := p.readIdent()
		if p.ReadByte() != endCh {
			p.i = start
			return "", false
		}
		return name, numeric
	}
	return p.readIdent()
}
func (p *Parser) readIdent() (string, bool) {
	var alnum bool
	end := len(p.b) - p.i
	for i, ch := range p.b[p.i:] {
		if isDigit(ch) {
			continue
		}
		if isAlpha(ch) || (i > 0 && alnum && ch == '_') {
			alnum = true
			continue
		}
		end = i
		break
	}
	if end == 0 {
		return "", false
	}
	b := p.b[p.i : p.i+end]
	p.i += end
	return unsafeconv.String(b), !alnum
}
func (p *Parser) ReadNumber() int {
	ind := len(p.b) - p.i
	for i, c := range p.b[p.i:] {
		if !isDigit(c) {
			ind = i
			break
		}
	}
	if ind == 0 {
		return 0
	}
	n, err := strconv.Atoi(string(p.b[p.i : p.i+ind]))
	if err != nil {
		panic(err)
	}
	p.i += ind
	return n
}
func isDigit(c byte) bool { return digit[c] }
func isAlpha(c byte) bool { return alpha[c] }
