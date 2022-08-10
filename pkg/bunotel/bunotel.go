package bunotel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	Meter  = global.Meter("github.com/uptrace/uptrace")
	Tracer = otel.Tracer("github.com/uptrace/uptrace")
)

var ProjectID = attribute.Key("project_id")
