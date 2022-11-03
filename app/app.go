package app

import (
	"fmt"
	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/app/infrastructure/database"
	"github.com/kuzznya/letsdeploy/app/infrastructure/k8s"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/server"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/procyon-projects/chrono"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"strings"
)

var logLevels = map[string]log.Level{
	"trace": log.TraceLevel,
	"debug": log.DebugLevel,
	"info":  log.InfoLevel,
	"warn":  log.WarnLevel,
	"error": log.ErrorLevel,
	"fatal": log.FatalLevel,
	"panic": log.PanicLevel,
}

func Start() {
	cfg := setupConfig()
	configureLogging(cfg)

	db := database.New(cfg)
	store := storage.New(db)
	clientset := setupK8sClientset(cfg)

	c := core.New(store, clientset, chrono.NewDefaultTaskScheduler())
	s := server.New(c)

	r := gin.Default()
	r.Use(openApiValidatorMiddleware(cfg))
	r.Use(middleware.ErrorHandler)

	handler := openapi.NewStrictHandler(s, make([]openapi.StrictMiddlewareFunc, 0))
	openapi.RegisterHandlersWithOptions(r, handler, openapi.GinServerOptions{
		Middlewares: []openapi.MiddlewareFunc{middleware.AuthMiddleware},
		ErrorHandler: func(ctx *gin.Context, err error, code int) {
			ctx.JSON(code, gin.H{"error": err.Error()})
		},
	})

	r.GET("/health", healthcheck)

	err := r.Run()
	if err != nil {
		log.Panicln(fmt.Errorf("cannot start server: %w", err))
	}
}

func setupConfig() *viper.Viper {
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

	return cfg
}

func configureLogging(cfg *viper.Viper) {
	cfg.SetDefault("log.level", "info")
	logLevel := cfg.GetString("log.level")
	if level, ok := logLevels[strings.ToLower(logLevel)]; ok {
		log.SetLevel(level)
	} else {
		log.Panicf("Unknown log level %s\n", logLevel)
	}

	formatter := &log.TextFormatter{FullTimestamp: true}
	log.SetFormatter(formatter)
}

func setupK8sClientset(cfg *viper.Viper) *kubernetes.Clientset {
	clienset := k8s.Setup(cfg)
	version, err := clienset.ServerVersion()
	if err != nil {
		log.WithError(err).Panicln("Kubernetes server version retrieval failed")
	}
	log.Infof("Kubernetes server version: %s\n", version.String())
	return clienset
}

func openApiValidatorMiddleware(cfg *viper.Viper) gin.HandlerFunc {
	openapiPath := cfg.GetString("openapi.path")
	if openapiPath == "" {
		log.Panicln("openapi.path parameter unset")
	}
	validator, err := oapiMiddleware.OapiValidatorFromYamlFile(openapiPath)
	if err != nil {
		log.WithError(err).Panicln("Failed to create OpenAPI validator middleware")
	}
	return validator
}

func healthcheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
