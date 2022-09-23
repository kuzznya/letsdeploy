package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/appErrors"
	"github.com/kuzznya/letsdeploy/app/handlers"
	"github.com/kuzznya/letsdeploy/app/models"
	"github.com/kuzznya/letsdeploy/app/storage"
	"net/http"
	"strconv"
)

func RegisterProjectRoutes(r gin.IRouter, s *storage.Storage) {
	projects := r.Group("/projects")
	projects.GET("", GetUserProjectsRoute(s))
	projects.POST("", CreateProjectRoute(s))
	projects.GET("/:id", GetProjectRoute(s))
	projects.PUT("/:id", UpdateProjectRoute(s))
	projects.DELETE("/:id", DeleteProjectRoute(s))

	projects.GET("/:id/participants", GetParticipantsRoute(s))
	projects.PUT("/:id/participants/:username", AddParticipantRoute(s))
	projects.DELETE("/:id/participants/:username", RemoveParticipantRoute(s))
}

func GetUserProjectsRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		username := ctx.GetString("username")
		projects, err := handlers.GetUserProjects(s, username)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, projects)
	}
}

func CreateProjectRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		username := ctx.GetString("username")
		project := models.Project{}
		err := ctx.BindJSON(&project)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		createdProject, err := handlers.CreateProject(ctx, s, project, username)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, createdProject)
	}
}

func GetProjectRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		project, err := handlers.GetProject(s, id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, project)
	}
}

func UpdateProjectRoute(s *storage.Storage) func(ctx *gin.Context) {
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
		err = handlers.UpdateProject(s, project, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func DeleteProjectRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		err = handlers.DeleteProject(s, id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func GetParticipantsRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse project id"})
			return
		}
		participants, err := handlers.GetParticipants(s, id, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, participants)
	}
}

func AddParticipantRoute(s *storage.Storage) func(ctx *gin.Context) {
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
		err = handlers.AddParticipant(s, id, username, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func RemoveParticipantRoute(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		requester := ctx.GetString("username")
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.Error(appErrors.BadRequest("failed to parse project id"))
			return
		}
		username := ctx.Param("username")
		if username == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "new participant username not defined"})
			return
		}
		err = handlers.RemoveParticipant(s, id, username, requester)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.Status(http.StatusOK)
	}
}
