package org

import (
	"go.uber.org/fx"

	"github.com/uptrace/uptrace/pkg/httperror"
)

var (
	ErrUnauthorized    = httperror.Unauthorized("please log in")
	ErrAccessDenied    = httperror.Forbidden("access denied")
	ErrProjectNotFound = httperror.NotFound("project not found")
)

type TrackableModel string

const (
	ModelUser    TrackableModel = "User"
	ModelProject TrackableModel = "Project"
	ModelSpan    TrackableModel = "Span"
)

var Module = fx.Module("org",
	fx.Provide(
		NewProjectGateway,
		NewUserGateway,
		fx.Annotate(
			NewJWTProvider,
			fx.As(new(UserProvider)),
			fx.ResultTags(`group:"user_providers"`),
		),
	),
	fx.Provide(
		fx.Private,
		NewMiddleware,
		NewUserHandler,
		NewProjectHandler,
		NewUsageHandler,
	),
	fx.Invoke(
		registerUserHandler,
		registerProjectHandler,
		registerUsageHandler,
	),
)
