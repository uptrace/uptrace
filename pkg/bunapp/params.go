package bunapp

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"go.uber.org/fx"
)

type RouterBundle struct {
	Router           *bunrouter.Router
	RouterGroup      *bunrouter.Group `name:"router_group"`
	RouterPublicV1   *bunrouter.Group `name:"router_public_v1"`
	RouterInternalV1 *bunrouter.Group `name:"router_internal_v1"`
}

type RouterParams struct {
	fx.In

	RouterBundle
}

type RouterResults struct {
	fx.Out

	RouterBundle
}

type PostgresParams struct {
	fx.In

	PG *bun.DB
}

type ClickhouseParams struct {
	fx.In

	Conf    *bunconf.Config
	CHSuper *ch.DB
}
