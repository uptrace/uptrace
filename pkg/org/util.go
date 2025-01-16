package org

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/uptrace/uptrace/pkg/attrkey"
)

type AttrMatcherOp string

const (
	AttrEqual    = "="
	AttrNotEqual = "!="
)

type AttrMatcher struct {
	Attr  string        `json:"attr"`
	Op    AttrMatcherOp `json:"op"`
	Value string        `json:"value"`
}

func (m *AttrMatcher) Matches(attrs map[string]any) bool {
	valueAny, ok := attrs[m.Attr]
	if !ok {
		return false
	}

	value := fmt.Sprint(valueAny)

	switch m.Op {
	case "=":
		return value == m.Value
	case "!=":
		return value != m.Value
	default:
		return false
	}
}

//------------------------------------------------------------------------------

var coreAttrs = []string{
	attrkey.SpanStatusCode,
	attrkey.DeploymentEnvironment,
	attrkey.ServiceName,
	attrkey.ServiceVersion,
	attrkey.ServiceNamespace,
	attrkey.HostName,
	attrkey.RPCMethod,
	attrkey.RPCService,
	attrkey.HTTPRequestMethod,
	attrkey.HTTPResponseStatusCode,
	attrkey.DBName,
	attrkey.DBOperation,
	attrkey.DBSqlTables,
	attrkey.LogSeverity,
	attrkey.LogSource,
	attrkey.LogFilePath,
	attrkey.LogFileName,
	attrkey.ExceptionType,
	attrkey.CodeFilepath,
	attrkey.CodeFunction,
}

func CompareAttrs(a, b string) int {
	i0 := slices.Index(coreAttrs, a)
	i1 := slices.Index(coreAttrs, b)

	if i0 == -1 && i1 == -1 {
		return strings.Compare(a, b)
	}
	if i0 == -1 {
		return -1
	}
	if i1 == -1 {
		return 1
	}
	return cmp.Compare(i0, i1)
}
