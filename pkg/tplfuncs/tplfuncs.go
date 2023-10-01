package tplfunc

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

var funcs = template.FuncMap{
	"safe": func(s string) template.HTML {
		return template.HTML(s)
	},
	"safeAttr": func(s string) template.HTMLAttr {
		return template.HTMLAttr(s)
	},
	"safeURL": func(s string) template.URL {
		return template.URL(s)
	},

	"trunc": func(n int, s string) string {
		if n == 0 {
			n = 100
		}
		return utf8util.Trunc(s, n)
	},

	"lastIndex": func(v any, idx int) bool {
		return idx+1 == reflect.ValueOf(v).Len()
	},

	"format": func(unit, v any) string {
		return bununit.Format(toFloat64(v), fmt.Sprint(unit))
	},
	"diff": func(a uint64, b uint64) string {
		return bununit.FormatFloat(float64(b) - float64(a))
	},
	"percents": func(v any) string {
		return bununit.FormatPercents(toFloat64(v))
	},
	"percentsSign": func(v any) string {
		return bununit.FormatPercents(toFloat64(v))
	},
	"microseconds": func(v any) string {
		return bununit.FormatMicroseconds(toFloat64(v))
	},
	"microsecondsSign": func(v any) string {
		return bununit.FormatMicroseconds(toFloat64(v))
	},
	"bytes": func(v any) string {
		return bununit.FormatBytes(toFloat64(v))
	},
	"bytesSign": func(v any) string {
		return bununit.FormatBytes(toFloat64(v))
	},
	"number": func(v any) string {
		return bununit.FormatFloat(toFloat64(v))
	},
	"numberSign": func(v any) string {
		return bununit.FormatFloat(toFloat64(v))
	},
	"date": func(v any) string {
		return bununit.FormatDate(toTime(v))
	},
	"time": func(v any) string {
		return bununit.FormatTime(toTime(v))
	},
	"timeSince": func(v any) time.Duration {
		return time.Since(v.(time.Time))
	},

	"float64": toFloat64,
	"round": func(mantissa int, value float64) float64 {
		return bununit.Round(value, mantissa)
	},
	"fixed": func(mantissa int, value float64) string {
		format := fmt.Sprintf("%%.%df", mantissa)
		return fmt.Sprintf(format, value)
	},

	"changeStyle": func(v any) template.HTMLAttr {
		return changeStyle(toFloat64(v))
	},

	"formatAttr": formatAttr,
}

func Funcs() template.FuncMap {
	return funcs
}

func toFloat64(v any) float64 {
	switch v := v.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case json.Number:
		n, err := v.Float64()
		if err != nil {
			panic(err)
		}
		return n
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	case int:
		return float64(v)
	default:
		logrus.Errorf("unsupported type: %T", v)
		return 0
	}
}

func toTime(v any) time.Time {
	switch v := v.(type) {
	case time.Time:
		return v
	case uint64:
		return time.Unix(0, int64(v))
	case int64:
		return time.Unix(0, v)
	default:
		return time.Time{}
	}
}

func changeStyle(n float64) template.HTMLAttr {
	if n > 0 {
		return `style="color: #E53935"`
	}
	return `style="color: #43A047"`
}

func formatAttr(attrs map[string]any, key string) template.HTML {
	switch v := attrs[key].(type) {
	case string:
		if m, ok := bunutil.IsJSON(v); ok {
			prettyJSON, _ := json.MarshalIndent(m, "", "  ")
			return htmlTag("pre", string(prettyJSON))
		}

		if strings.IndexByte(v, '\n') >= 0 {
			return htmlTag("pre", v)
		}

		return template.HTML(v)
	case []string:
		data := strings.Join(v, ", ")
		return htmlTag("span", data)
	case []any:
		ss := make([]string, len(v))
		for i, v := range v {
			ss[i] = fmt.Sprint(v)
		}
		return htmlTag("span", strings.Join(ss, ", "))
	default:
		return htmlTag("span", fmt.Sprintf("%v", v))
	}
}

func htmlTag(tag, val string) template.HTML {
	html := fmt.Sprintf("<%s>%s</%s>", tag, val, tag)
	return template.HTML(html)
}
