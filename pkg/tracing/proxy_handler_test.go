package tracing_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/testbed"
	"github.com/uptrace/uptrace/pkg/tracing"
)

func TestProxyHandler(t *testing.T) {
	_, app := testbed.StartApp(t)
	defer app.Stop()

	msg := "foo bar"

	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, msg)
	}))
	defer backendServer.Close()

	url := "/loki/api/v1/label"
	router := bunrouter.New()

	handler := tracing.NewLokiProxyHandler(app, backendServer.URL)
	router.GET(url, handler.Proxy)

	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(t, 200, resp.Code)
	require.Equal(t, msg, resp.Body.String())
}
