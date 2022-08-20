package httputil

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/httperror"
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

//------------------------------------------------------------------------------

func UnmarshalJSON(
	w http.ResponseWriter,
	req bunrouter.Request,
	dst any,
	maxBytes int64,
) error {
	req.Body = http.MaxBytesReader(w, req.Body, maxBytes)

	dec := json.NewDecoder(req.Body)
	dec.DontMatchCaseInsensitiveStructFields()

	if err := dec.Decode(dst); err != nil {
		if isTimeout(err) {
			return httperror.ErrRequestTimeout
		}
		return err
	}

	return nil
}

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() {
		return true
	}

	if oe, ok := err.(*net.OpError); ok {
		if se, ok := oe.Err.(*os.SyscallError); ok {
			if se.Err == syscall.ECONNRESET {
				return true
			}
		}
	}

	return false
}
