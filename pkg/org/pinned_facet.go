package org

import (
	"cmp"
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"golang.org/x/exp/slices"
)

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
	attrkey.DBSqlTable,
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

type PinnedFacet struct {
	bun.BaseModel `bun:"pinned_facets,alias:f"`

	UserID   uint64 `json:"userId"`
	Attr     string `json:"attr"`
	Unpinned bool   `json:"-"`
}

func SelectPinnedFacets(ctx context.Context, pg *bun.DB, userID uint64) ([]string, error) {
	var facets []*PinnedFacet

	if err := pg.NewSelect().
		Model(&facets).
		Where("user_id = ?", userID).
		Scan(ctx); err != nil {
		return nil, err
	}

	attrs := make([]string, 0)

	unpinnedAttrs := make(map[string]bool, 0)
	for _, facet := range facets {
		if facet.Unpinned {
			unpinnedAttrs[facet.Attr] = true
		} else {
			attrs = append(attrs, facet.Attr)
		}
	}

	for _, facet := range coreAttrs {
		if !unpinnedAttrs[facet] {
			attrs = append(attrs, facet)
		}
	}

	return attrs, nil
}

func SelectPinnedFacetMap(ctx context.Context, pg *bun.DB, userID uint64) (map[string]bool, error) {
	pinnedAttrs, err := SelectPinnedFacets(ctx, pg, userID)
	if err != nil {
		return nil, err
	}

	pinnedAttrMap := make(map[string]bool, len(pinnedAttrs))
	for _, attrKey := range pinnedAttrs {
		pinnedAttrMap[attrKey] = true
	}
	return pinnedAttrMap, nil
}
