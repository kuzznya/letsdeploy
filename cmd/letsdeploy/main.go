package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/infrastructure/database"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/projects"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func main() {
	cfg := viper.New()

	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath("/etc/letsdeploy/")
	cfg.AddConfigPath(".")
	err := cfg.ReadInConfig()
	if err != nil {
		msg := fmt.Errorf("fatal error reading config: %w", err)
		logrus.Fatalln(msg)
		panic(msg)
	}

	db := database.Setup()
	store := storage.New(db)
	store.ProjectRepository.FindUserProjects("kuzznya")

	r := gin.Default()
	r.Use(middleware.AuthMiddleware)
	r.GET("/health", healthcheck)
	v1 := r.Group("/api/v1")
	projects.RegisterRoutes(v1)
	r.Group("/projects")

	err = r.Run()
	if err != nil {
		msg := fmt.Errorf("cannot start server: %w", err)
		logrus.Fatalln(msg)
		panic(msg)
	}
}

func healthcheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
