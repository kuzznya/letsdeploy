package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/app/infrastructure/database"
	"github.com/kuzznya/letsdeploy/app/infrastructure/k8s"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/routers"
	"github.com/kuzznya/letsdeploy/app/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

func Start() {
	cfg := viper.New()

	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath("/etc/letsdeploy/")
	cfg.AddConfigPath("./configs")

	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := cfg.ReadInConfig()
	if err != nil {
		msg := fmt.Errorf("fatal error reading config: %w", err)
		log.Fatalln(msg)
	}

	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	db := database.Setup(cfg)
	store := storage.New(db)

	clienset := k8s.Setup(cfg)

	version, err := clienset.ServerVersion()
	if err != nil {
		log.WithError(err).Panicln("Kubernetes server version retrieval failed")
	}
	log.Infof("Kubernetes server version: %s\n", version.String())

	c := core.New(store, clienset)

	r := gin.Default()

	r.Use(middleware.ErrorHandler)

	r.GET("/health", healthcheck)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.Auth)
	routers.RegisterAllRoutes(v1, c)

	err = r.Run()
	if err != nil {
		log.Panicln(fmt.Errorf("cannot start server: %w", err))
	}
}

func healthcheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
