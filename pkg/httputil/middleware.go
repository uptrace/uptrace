package httputil

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

//------------------------------------------------------------------------------

type TraceparentHandler struct {
	next  http.Handler
	props propagation.TextMapPropagator
}

func NewTraceparentHandler(next http.Handler) *TraceparentHandler {
	return &TraceparentHandler{
		next:  next,
		props: otel.GetTextMapPropagator(),
	}
}

func (h *TraceparentHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.props.Inject(req.Context(), propagation.HeaderCarrier(w.Header()))
	h.next.ServeHTTP(w, req)
}

//------------------------------------------------------------------------------

type SubpathHandler struct {
	next    http.Handler
	subpath string
}

func NewSubpathHandler(next http.Handler, subpath string) *SubpathHandler {
	return &SubpathHandler{
		next:    next,
		subpath: strings.TrimSuffix(subpath, "/"),
	}
}

func (h *SubpathHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, h.subpath)
	req.URL.Path = strings.TrimPrefix(req.URL.Path, h.subpath)
	h.next.ServeHTTP(w, req)
}
