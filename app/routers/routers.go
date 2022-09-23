package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/core"
)

func RegisterAllRoutes(r gin.IRouter, c *core.Core) {
	RegisterProjectRoutes(r, c)
}
