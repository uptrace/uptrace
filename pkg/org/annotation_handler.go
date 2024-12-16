package org

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"go.uber.org/fx"
)

type AnnotationHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
	PS     *ProjectStore
}

type AnnotationHandler struct {
	*AnnotationHandlerParams
}

func NewAnnotationHandler(p AnnotationHandlerParams) *AnnotationHandler {
	return &AnnotationHandler{&p}
}

func registerAnnotationHandler(h *AnnotationHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.POST("/annotations", h.CreatePublic)

	p.RouterInternalV1.Use(m.UserAndProject).
		WithGroup("/projects/:project_id/annotations", func(g *bunrouter.Group) {
			g.GET("", h.List)
			g.POST("", h.Create)

			g = g.Use(h.AnnotationMiddleware)
			g.GET("/:annotation_id", h.Show)
			g.PUT("/:annotation_id", h.Update)
			g.DELETE("/:annotation_id", h.Delete)
		})
}

func (h *AnnotationHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(AnnotationFilter)
	if err := decodeAnnotationFilter(req, f); err != nil {
		return err
	}

	anns := make([]*Annotation, 0)

	count, err := h.PG.NewSelect().
		Model(&anns).
		Apply(f.WhereClause).
		Apply(f.PGOrder).
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		ScanAndCount(ctx)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"annotations": anns,
		"count":       count,
	})
}

func (h *AnnotationHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	ann := AnnotationFromContext(ctx)

	return httputil.JSON(w, bunrouter.H{
		"annotation": ann,
	})
}

type AnnotationIn struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fingerprint string            `json:"fingerprint"`
	Color       string            `json:"color"`
	Attrs       map[string]string `json:"attrs"`
	Time        time.Time         `json:"time"`
}

func (in *AnnotationIn) Validate(ann *Annotation) error {
	if in.Name == "" {
		return errors.New("name is required")
	}
	if in.Color == "" {
		in.Color = "red"
	}

	ann.Name = utf8util.Trunc(in.Name, 100)
	ann.Description = utf8util.Trunc(in.Description, 5000)
	ann.Color = utf8util.Trunc(in.Color, 100)
	ann.Attrs = in.Attrs
	ann.CreatedAt = in.Time

	ann.Attrs = make(map[string]string, len(in.Attrs))
	for k, v := range in.Attrs {
		k = utf8util.TruncLC(attrkey.Clean(k))
		v = utf8util.TruncLarge(v)
		ann.Attrs[k] = v
	}

	if in.Fingerprint != "" {
		ann.Hash = xxhash.Sum64([]byte(in.Fingerprint))
	}

	return nil
}

func (h *AnnotationHandler) CreatePublic(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := h.PS.SelectProjectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	if _, err := h.createAnnotation(w, req, project); err != nil {
		return err
	}

	return nil
}

func (h *AnnotationHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := ProjectFromContext(ctx)

	if _, err := h.createAnnotation(w, req, project); err != nil {
		return err
	}

	return nil
}

func (h *AnnotationHandler) createAnnotation(
	w http.ResponseWriter, req bunrouter.Request, project *Project,
) (*Annotation, error) {
	var in AnnotationIn

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return nil, err
	}

	ann := &Annotation{
		ProjectID: project.ID,
	}
	if err := in.Validate(ann); err != nil {
		return nil, err
	}

	if _, err := h.PG.NewInsert().
		Model(ann).
		On("CONFLICT (project_id, hash) DO NOTHING").
		Exec(req.Context()); err != nil {
		return nil, err
	}

	return ann, nil
}

func (h *AnnotationHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	ann := AnnotationFromContext(ctx)

	var in AnnotationIn
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}
	if err := in.Validate(ann); err != nil {
		return err
	}

	if err := h.PG.NewUpdate().
		Model(ann).
		Set("name = ?", ann.Name).
		Set("description = ?", ann.Description).
		Set("color = ?", ann.Color).
		Set("attrs = ?", ann.Attrs).
		Where("id = ?", ann.ID).
		Returning("*").
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"annotation": ann,
	})
}

func (h *AnnotationHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	ann := AnnotationFromContext(ctx)

	if _, err := h.PG.NewDelete().
		Model(ann).
		Where("id = ?", ann.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *AnnotationHandler) AnnotationMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		project := ProjectFromContext(ctx)

		annID, err := req.Params().Uint64("annotation_id")
		if err != nil {
			return err
		}

		ann := new(Annotation)

		if err := h.PG.NewSelect().
			Model(ann).
			Where("id = ?", annID).
			Where("project_id = ?", project.ID).
			Scan(ctx); err != nil {
			return err
		}

		ctx = ContextWithAnnotation(ctx, ann)
		return next(w, req.WithContext(ctx))
	}
}

type annCtxKey struct{}

func AnnotationFromContext(ctx context.Context) *Annotation {
	return ctx.Value(annCtxKey{}).(*Annotation)
}

func ContextWithAnnotation(ctx context.Context, ann *Annotation) context.Context {
	return context.WithValue(ctx, annCtxKey{}, ann)
}
