package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/app/models"
	"net/http"
	"strconv"
)

func RegisterProjectRoutes(r gin.IRouter, c *core.Core) {
	projects := r.Group("/projects")
	projects.GET("", GetUserProjectsRoute(c))
	projects.POST("", CreateProjectRoute(c))
	projects.GET("/:id", GetProjectRoute(c))
	projects.PUT("/:id", UpdateProjectRoute(c))
	projects.DELETE("/:id", DeleteProjectRoute(c))

	projects.GET("/:id/participants", GetParticipantsRoute(c))
	projects.PUT("/:id/participants/:username", AddParticipantRoute(c))
	projects.DELETE("/:id/participants/:username", RemoveParticipantRoute(c))
}

func GetUserProjectsRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		username := ctx.GetString("username")
		projects, err := c.Projects.GetUserProjects(username)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, projects)
	}
}

func CreateProjectRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		username := ctx.GetString("username")
		project := models.Project{}
		err := ctx.BindJSON(&project)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		createdProject, err := c.Projects.CreateProject(ctx, project, username)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, createdProject)
	}
}

func GetProjectRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		project, err := c.Projects.GetProject(id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, project)
	}
}

func UpdateProjectRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		project := models.Project{Id: id}
		err = ctx.BindJSON(&project)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		err = c.Projects.UpdateProject(project, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func DeleteProjectRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		err = c.Projects.DeleteProject(ctx, id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func GetParticipantsRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		participants, err := c.Projects.GetParticipants(id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, participants)
	}
}

func AddParticipantRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		username := ctx.Param("username")
		if username == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "new participant username not defined"})
			return
		}
		err = c.Projects.AddParticipant(id, username, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func RemoveParticipantRoute(c *core.Core) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.Error(apperrors.BadRequest("failed to parse project id"))
			return
		}
		username := ctx.Param("username")
		if username == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "new participant username not defined"})
			return
		}
		err = c.Projects.RemoveParticipant(id, username, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}
