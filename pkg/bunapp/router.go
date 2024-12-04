package bunapp

import (
	"github.com/uptrace/bunrouter"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Router           *bunrouter.Router
	RouterGroup      *bunrouter.Group `name:"router_group"`
	RouterInternalV1 *bunrouter.Group `name:"router_internal_apiv1"`
	RouterPublicV1   *bunrouter.Group `name:"router_public_apiv1"`
}
