package httputil

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

type PanicHandler struct {
	Next http.Handler
}

func (h PanicHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 10<<10)
			n := runtime.Stack(buf, false)
			fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", err, buf[:n])
			os.Exit(1)
		}
	}()

	h.Next.ServeHTTP(w, req)
}

//------------------------------------------------------------------------------

type DecompressHandler struct {
	Next http.Handler
}

func (h DecompressHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	newBody, err := newBodyReader(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if newBody != nil {
		defer newBody.Close()
		req.Header.Del("Content-Encoding")
		req.Header.Del("Content-Length")
		req.ContentLength = -1
		req.Body = newBody
	}
	h.Next.ServeHTTP(w, req)
}

func newBodyReader(req *http.Request) (io.ReadCloser, error) {
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		gr, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		return gr, nil
	case "deflate", "zlib":
		zr, err := zlib.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		return zr, nil
	}
	return nil, nil
}
