package app

import (
	"codnect.io/chrono"
	"fmt"
	certManagerClientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/core"
	"github.com/kuzznya/letsdeploy/app/infrastructure/database"
	"github.com/kuzznya/letsdeploy/app/infrastructure/k8s"
	"github.com/kuzznya/letsdeploy/app/infrastructure/redisclient"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/server"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	oapiMiddleware "github.com/oapi-codegen/gin-middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"strings"
	"time"
)

func Start() {
	cfg := setupConfig()
	configureLogging(cfg)

	db := database.New(cfg)
	store := storage.New(db)
	rdb := redisclient.New(cfg)
	clientset, cmClientset := setupK8sClientsets(cfg)

	c := core.New(store, rdb, clientset, cmClientset, chrono.NewDefaultTaskScheduler(), cfg)
	s := server.New(c)

	r := gin.Default()
	r.Use(openApiValidatorMiddleware([]string{"/api/v1"}, []string{"/api/v1/services/:id/logs"}))
	r.Use(middleware.ErrorHandler)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://letsdeploy.space", "http://localhost:5173"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		MaxAge:           12 * time.Hour,
	}))

	handler := openapi.NewStrictHandler(s, make([]openapi.StrictMiddlewareFunc, 0))
	openapi.RegisterHandlersWithOptions(r, handler, openapi.GinServerOptions{
		Middlewares: []openapi.MiddlewareFunc{
			middleware.CreateAuthMiddleware(cfg),
			middleware.CreateApiKeyAuthMiddleware(func(apiKey string) (string, error) {
				return c.ApiKeys.GetUsernameByApiKey(apiKey)
			}),
			middleware.Authz,
		},
		ErrorHandler: func(ctx *gin.Context, err error, code int) {
			ctx.JSON(code, gin.H{"error": err.Error()})
		},
	})

	server.ServiceLogStreamEndpoint(r, c, rdb)

	r.Use(func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(200)
		}
	})

	r.GET("/v3/api-docs", func(ctx *gin.Context) {
		docs, err := openapi.GetSwagger()
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		ctx.JSON(200, docs)
	})
	r.StaticFile("swagger-ui.html", "./static/swagger-ui.html")
	r.StaticFile("oauth2-redirect.html", "./static/oauth2-redirect.html")

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

	profile := cfg.GetString("profile")
	if profile != "" {
		log.Infof("Run profile %s, loading config.%s.yaml if exists", profile, profile)
		cfg.SetConfigName("config." + profile)
		err = cfg.MergeInConfig()
		if err != nil {
			log.WithError(err).Warnf("Failed to read profile-specific config")
		}
	}

	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	return cfg
}

func configureLogging(cfg *viper.Viper) {
	cfg.SetDefault("log.level", "info")
	logLevel := cfg.GetString("log.level")
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Panicf("Failed to parse log level: %s", err)
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{FullTimestamp: true}
	log.SetFormatter(formatter)
}

func setupK8sClientsets(cfg *viper.Viper) (*kubernetes.Clientset, *certManagerClientset.Clientset) {
	clienset, cmClientset := k8s.Setup(cfg)
	version, err := clienset.ServerVersion()
	if err != nil {
		log.WithError(err).Panicln("Kubernetes server version retrieval failed")
	}
	log.Infof("Kubernetes server version: %s\n", version.String())
	return clienset, cmClientset
}

func openApiValidatorMiddleware(includePaths []string, excludePaths []string) gin.HandlerFunc {
	apiDocs, err := openapi.GetSwagger()
	if err != nil {
		log.WithError(err).Panicln("Failed to get OpenAPI docs")
	}

	options := oapiMiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			MultiError:         true,
		},
		MultiErrorHandler: func(me openapi3.MultiError) error {
			return me
		},
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			log.Infof("Bad request: %s", message)
			c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
		},
	}
	validator := oapiMiddleware.OapiRequestValidatorWithOptions(apiDocs, &options)
	if err != nil {
		log.WithError(err).Panicln("Failed to create OpenAPI validator middleware")
	}

	return func(ctx *gin.Context) {
		for _, path := range excludePaths {
			if strings.HasPrefix(ctx.FullPath(), path) {
				ctx.Next()
				return
			}
		}

		for _, path := range includePaths {
			if strings.HasPrefix(ctx.FullPath(), path) {
				validator(ctx)
				return
			}
		}
		ctx.Next()
	}
}

func healthcheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
