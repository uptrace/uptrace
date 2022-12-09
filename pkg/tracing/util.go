package tracing

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
)

func asString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprint(v)
	}
}

func listToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func quantileLevel(fn string) float64 {
	n, err := strconv.ParseInt(fn[1:], 10, 64)
	if err != nil {
		panic(err)
	}
	return float64(n) / 100
}

func dsnFromRequest(req bunrouter.Request) string {
	if dsn := req.Header.Get("uptrace-dsn"); dsn != "" {
		return dsn
	}

	if auth := req.Header.Get("Authorization"); auth != "" {
		const bearer = "Bearer "
		return strings.TrimPrefix(auth, bearer)
	}

	if dsn := req.URL.Query().Get("dsn"); dsn != "" {
		return dsn
	}

	return ""
}

//------------------------------------------------------------------------------

func formatSQL(query string) string {
	cmd := exec.Command("clickhouse-format")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ""
	}

	if _, err := stdin.Write([]byte(query)); err != nil {
		stdin.Close()
		return ""
	}
	stdin.Close()

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(out)
}
