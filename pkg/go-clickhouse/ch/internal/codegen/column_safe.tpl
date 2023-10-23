//go:build !amd64 && !arm64

package chschema

import (
	"github.com/uptrace/go-clickhouse/ch/chproto"
)

{{- range . }}

{{ if eq .Size 0 }} {{ continue }} {{ end }}

func (c *{{ .Name }}Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	c.AllocForReading(numRow)

	for i := range c.Column {
		n, err := rd.{{ .Name }}()
		if err != nil {
			return err
		}
		c.Column[i] = n
	}

	return nil
}

func (c *{{ .Name }}Column) WriteTo(wr *chproto.Writer) error {
	for _, n := range c.Column {
		wr.{{ .Name }}(n)
	}
	return nil
}

{{- end }}
