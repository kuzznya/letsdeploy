package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/storage"
)

func RegisterAllRoutes(r gin.IRouter, s *storage.Storage) {
	RegisterProjectRoutes(r, s)
}
