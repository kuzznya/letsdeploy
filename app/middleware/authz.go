package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kuzznya/letsdeploy/app/apperrors"
)

func Authz(c *gin.Context) {
	if c.Value(authContextKey) == nil {
		_ = c.Error(apperrors.Unauthorized("Either Bearer authentication or API key authentication should be provided"))
		c.Abort()
		return
	}
	c.Next()
}
