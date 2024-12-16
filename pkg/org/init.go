package org

import (
	"github.com/vmihailenco/taskq/v4"
	"go.uber.org/fx"

	"github.com/uptrace/uptrace/pkg/httperror"
)

var (
	ErrUnauthorized    = httperror.Unauthorized("please log in")
	ErrAccessDenied    = httperror.Forbidden("access denied")
	ErrProjectNotFound = httperror.NotFound("project not found")
)

var CreateErrorAlertTask = taskq.NewTask("create-error-alert")

type TrackableModel string

const (
	ModelUser      TrackableModel = "User"
	ModelProject   TrackableModel = "Project"
	ModelSpan      TrackableModel = "Span"
	ModelSpanGroup TrackableModel = "SpanGroup"
)

var Module = fx.Module("org",
	fx.Provide(
		NewProjectStore,
	),
	fx.Provide(
		fx.Private,
		NewMiddleware,
		NewUserHandler,
		NewSSOHandler,
		NewProjectHandler,
		NewUsageHandler,
		NewPinnedFacetHandler,
		NewAchievementHandler,
		NewAnnotationHandler,
	),
	fx.Invoke(
		registerUserHandler,
		registerSSOHandler,
		registerProjectHandler,
		registerUsageHandler,
		registerPinnedFacetHandler,
		registerAchievementHandler,
		registerAnnotationHandler,
	),
)
