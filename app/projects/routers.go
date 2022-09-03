package projects

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter) {
	projects := r.Group("/projects")
	projects.GET("", GetUserProjectsHandler)
}

func GetUserProjectsHandler(ctx *gin.Context) {
	ctx.Get("claims")
}
