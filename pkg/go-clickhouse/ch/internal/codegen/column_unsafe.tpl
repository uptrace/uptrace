//go:build amd64 || arm64

package chschema

import (
	"io"
	"reflect"
   "unsafe"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

{{- range . }}

{{ if eq .Size 0 }} {{ continue }} {{ end }}

func (c *{{ .Name }}Column) ReadFrom(rd *chproto.Reader, numRow int) error {
	const size = {{ .Size }} / 8

	if numRow == 0 {
		return nil
	}

	c.AllocForReading(numRow)

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	dest := *(*[]byte)(unsafe.Pointer(&slice))
	_, err := io.ReadFull(rd, dest)
	return err
}

func (c *{{ .Name }}Column) WriteTo(wr *chproto.Writer) error {
	const size = {{ .Size }} / 8

	if len(c.Column) == 0 {
	  return nil
	}

	slice := *(*reflect.SliceHeader)(unsafe.Pointer(&c.Column))
	slice.Len *= size
	slice.Cap *= size

	src := *(*[]byte)(unsafe.Pointer(&slice))
	wr.Write(src)
	return nil
}

{{- end }}
