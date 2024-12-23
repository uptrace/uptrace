package pgquery

import (
	"github.com/uptrace/pkg/unsafeconv"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type TSBuilder struct {
	title []byte
	body  []byte
	attrs []string
}

func NewTSBuilder() *TSBuilder {
	return &TSBuilder{}
}

func (b *TSBuilder) Title() string {
	return unsafeconv.String(b.title)
}

func (b *TSBuilder) Body() string {
	return unsafeconv.String(b.body)
}

func (b *TSBuilder) Attrs() []string {
	return b.attrs
}

func (b *TSBuilder) AddTitle(s string) {
	if len(s) == 0 {
		return
	}

	if len(b.title) > 0 {
		b.title = append(b.title, ' ')
	}
	b.title = append(b.title, s...)
}

func (b *TSBuilder) AddBody(s string) {
	if len(s) == 0 {
		return
	}

	if len(b.body) > 0 {
		b.body = append(b.body, ' ')
	}
	b.body = append(b.body, s...)
}

func (b *TSBuilder) AddAttr(key, value string) {
	b.attrs = append(b.attrs, BuildAttr(key, value))
	b.AddBody(value)
}

func BuildAttr(key, value string) string {
	return "~~" + key + "~~" + utf8util.TruncLC(value)
}

func EscapeWord(s string) string {
	dst := make([]byte, 0, len(s))
	dst = appendWord(dst, s)
	return unsafeconv.String(dst)
}

func appendWord(b []byte, word string) []byte {
	for i := 0; i < len(word); i++ {
		ch := word[i]
		switch ch {
		case ':', '&', '|', '!', '(', ')', ' ':
			b = append(b, '\\', ch)
		case '\'', '<', '>':
			// Do nothing.
		default:
			b = append(b, ch)
		}
	}
	return b
}
