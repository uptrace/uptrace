package tracing

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

type ServiceGraphHandler struct {
	*bunapp.App
}

func NewServiceGraphHandler(app *bunapp.App) *ServiceGraphHandler {
	return &ServiceGraphHandler{
		App: app,
	}
}

type ServiceGraphLink struct {
	Type       string `json:"type"`
	ClientName string `json:"clientName"`
	ServerName string `json:"serverName"`
	ServerAttr string `json:"serverAttr"`
	ServiceGraphStats
}

type ServiceGraphStats struct {
	DurationMin float32 `json:"durationMin"`
	DurationMax float32 `json:"durationMax"`
	DurationSum float64 `json:"durationSum"`
	DurationAvg float64 `json:"durationAvg"`
	Count       uint64  `json:"count"`
	Rate        float64 `json:"rate"`
	ErrorCount  uint64  `json:"errorCount"`
	ErrorRate   float64 `json:"errorRate"`
}

func (h *ServiceGraphHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	var envs []string
	var serviceNamespaces []string
	var serviceNames []string

	for _, part := range f.parts {
		if part.Disabled || part.Error.Wrapped != nil {
			continue
		}

		where, ok := part.AST.(*tql.Where)
		if !ok {
			continue
		}

		if len(where.Filters) != 1 {
			continue
		}

		filter := &where.Filters[0]
		if filter.Op != tql.FilterEqual && filter.Op != tql.FilterIn {
			continue
		}

		switch tql.String(filter.LHS) {
		case attrkey.DeploymentEnvironment:
			envs = filter.RHS.Values()
		case attrkey.ServiceNamespace:
			serviceNamespaces = filter.RHS.Values()
		case attrkey.ServiceName:
			serviceNames = filter.RHS.Values()
		}
	}

	minutes := f.Duration().Minutes()
	q := h.CH.NewSelect().
		Model((*ServiceGraphEdge)(nil)).
		ColumnExpr("e.type").
		ColumnExpr("e.client_name").
		ColumnExpr("e.server_name").
		ColumnExpr("any(e.server_attr) AS server_attr").
		// Use client duration since it is always present.
		ColumnExpr("min(e.client_duration_min) AS duration_min").
		ColumnExpr("max(e.client_duration_max) AS duration_max").
		ColumnExpr("sum(e.client_duration_sum) AS duration_sum").
		ColumnExpr("sum(e.count) AS count").
		ColumnExpr("sum(e.count) / ? AS rate", minutes).
		ColumnExpr("sum(e.error_count) AS error_count").
		Where("e.project_id = ?", f.ProjectID).
		Where("e.time >= ?", f.TimeGTE).
		Where("e.time < ?", f.TimeLT).
		GroupExpr("e.type, e.client_name, e.server_name")

	// if len(f.System) > 0 {
	// 	q = q.Where("e.client_name IN ?0 OR e.server_name IN ?0", ch.In(f.System))
	// }
	if len(envs) > 0 {
		q = q.Where("e.deployment_environment IN ?0", ch.In(envs))
	}
	if len(serviceNamespaces) > 0 {
		q = q.Where("e.service_namespace IN ?0", ch.In(serviceNamespaces))
	}
	if len(serviceNames) > 0 {
		q = q.Where("e.client_name IN ?0 OR e.server_name IN ?0", ch.In(serviceNames))
	}

	edges := make([]*ServiceGraphLink, 0)

	if err := q.Scan(ctx, &edges); err != nil {
		return err
	}

	for _, edge := range edges {
		if edge.Count > 0 {
			edge.ErrorRate = float64(edge.ErrorCount) / float64(edge.Count)
			edge.DurationAvg = float64(edge.DurationSum) / float64(edge.Count)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"edges": edges,
	})
}
