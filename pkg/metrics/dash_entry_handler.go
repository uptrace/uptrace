package metrics

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/uptrace/bunrouter"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type DashEntryHandler struct {
	App *bunapp.App
}

func NewDashEntryHandler(app *bunapp.App) *DashEntryHandler {
	return &DashEntryHandler{
		App: app,
	}
}

func (h *DashEntryHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	if dash.TemplateID != "" {
		return errPrebuiltDashboard
	}

	entries, err := SelectDashEntries(ctx, h.App, dash)
	if err != nil {
		return err
	}
	if len(entries) >= numEntryLimit {
		return fmt.Errorf("dashboards can't have more than 20 entries")
	}

	entry := new(DashEntry)
	if err := httputil.UnmarshalJSON(w, req, entry, 100<<10); err != nil {
		return err
	}

	entry.ProjectID = project.ID
	entry.DashID = dash.ID
	if entry.Columns == nil {
		entry.Columns = make(map[string]*MetricColumn)
	}

	if _, err := h.App.DB.NewInsert().
		Model(entry).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"entry": entry,
	})
}

func (h *DashEntryHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	if dash.TemplateID != "" {
		return errPrebuiltDashboard
	}

	entry := new(DashEntry)
	if err := httputil.UnmarshalJSON(w, req, entry, 100<<10); err != nil {
		return err
	}

	if len(entry.Metrics) == 0 {
		return errors.New("at least one query is required")
	}
	if len(entry.Metrics) > numMetricLimit {
		return errors.New("you can't have more than 6 queries on a single chart")
	}

	entryID, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}
	entry.ID = entryID

	if _, err := h.App.DB.NewUpdate().
		Model(entry).
		Column("name", "description", "chart_type", "query", "metrics", "columns").
		Where("id = ?", entryID).
		Where("project_id = ?", project.ID).
		Where("dash_id = ?", dash.ID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"entry": entry,
	})
}

func (h *DashEntryHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	if dash.TemplateID != "" {
		return errPrebuiltDashboard
	}

	entryID, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	entry := new(DashEntry)

	if _, err := h.App.DB.
		NewDelete().
		Model(entry).
		Where("id = ?", entryID).
		Where("project_id = ?", project.ID).
		Where("dash_id = ?", dash.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"entry": entry,
	})
}

func (h *DashEntryHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	entries, err := SelectDashEntries(ctx, h.App, dash)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"list": nil,
		"grid": entries,
	})
}

func (h *DashEntryHandler) UpdateOrder(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	var entries []DashEntry
	if err := httputil.UnmarshalJSON(w, req, &entries, 10<<10); err != nil {
		return err
	}

	if _, err := h.App.DB.NewUpdate().
		With("_data", h.App.DB.NewValues(&entries)).
		Model(&entries).
		TableExpr("_data").
		Set("weight = _data.weight").
		Where("e.id = _data.id").
		Where("e.dash_id = ?", dash.ID).
		Where("e.project_id = ?", project.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"entries": entries,
	})
}
