package org

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type AnnotationHandler struct {
	pg *bun.DB
}

func NewAnnotationHandler(pg *bun.DB) *AnnotationHandler {
	return &AnnotationHandler{pg: pg}
}

func (h *AnnotationHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(AnnotationFilter)
	if err := decodeAnnotationFilter(req, f); err != nil {
		return err
	}

	anns := make([]*Annotation, 0)

	count, err := h.pg.NewSelect().
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

	fakeApp := &bunapp.App{PG: h.pg}
	project, err := SelectProjectByDSN(ctx, fakeApp, dsn)
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

	if _, err := h.pg.NewInsert().
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

	if err := h.pg.NewUpdate().
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

	if _, err := h.pg.NewDelete().
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

		if err := h.pg.NewSelect().
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
