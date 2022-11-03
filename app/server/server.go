package server

import (
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/internal/openapi"
)

type Server struct {
	core *core.Core
}

// assert that Server implements openapi.StrictServerInterface
var _ openapi.StrictServerInterface = (*Server)(nil)

func New(core *core.Core) Server {
	return Server{core: core}
}
