package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	log "github.com/sirupsen/logrus"
)

type UserProviderFunc = func(apiKey string) (username string, err error)

func CreateApiKeyAuthMiddleware(userProvider UserProviderFunc) openapi.MiddlewareFunc {
	return func(c *gin.Context) {
		ApiKeyAuthMiddleware(c, userProvider)
	}
}

func ApiKeyAuthMiddleware(ctx *gin.Context, userProvider UserProviderFunc) {
	key := ctx.GetHeader("API-Key")
	if key == "" {
		ctx.Next()
		return
	}
	username, err := userProvider(key)
	if err != nil {
		log.WithError(err).Errorln("Failed to authenticate by API key")
		_ = ctx.Error(apperrors.Forbidden("Failed to authenticate by API key"))
		ctx.Abort()
		return
	}
	ctx.Set(authContextKey, &Authentication{Username: username, Token: key})
	log.Debugf("User %s authenticated by API key", username)

	ctx.Next()
}
