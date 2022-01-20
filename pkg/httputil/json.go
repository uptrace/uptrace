package httputil

import (
	"bytes"
	"io"
	"net/http"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
)

func JSON(w http.ResponseWriter, res any) error {
	if res == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	switch res := res.(type) {
	case string:
		_, _ = io.WriteString(w, res)
		return nil
	case []byte:
		_, _ = w.Write(res)
		return nil
	}

	enc := newEncoder(w)
	if err := enc.Encode(res); err != nil {
		return err
	}

	return nil
}

func Must(m bunrouter.H, err error) any {
	if err != nil {
		return err
	}
	return m
}

func MarshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := newEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func newEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetStringifyLargeInts(true)
	return enc
}
