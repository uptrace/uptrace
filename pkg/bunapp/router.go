package bunapp

import (
	"github.com/uptrace/bunrouter"
)

type Router struct {
	Router      *bunrouter.Router
	RouterGroup *bunrouter.Group
	InternalV1  *bunrouter.Group
	PublicV1    *bunrouter.Group
}
