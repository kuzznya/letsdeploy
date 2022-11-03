package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
)

const authContextKey = "authentication"

type Authentication struct {
	Username string
}

func AuthMiddleware(ctx *gin.Context) {
	//headerValue := ctx.GetHeader("Authorization")
	//if headerValue == "" || len(headerValue) < 8 || headerValue[:7] != "Bearer" {
	//	ctx.JSON(http.StatusUnauthorized,
	//		gin.H{"error": "Bearer token should be provided in Authorization header"})
	//	ctx.Abort()
	//	return
	//}
	//if headerValue[:7] != "Bearer" {
	//	ctx.JSON(http.StatusUnauthorized,
	//		gin.H{"error": "Authorization header value should start with 'Bearer'"})
	//	ctx.Abort()
	//	return
	//}
	////token := headerValue[7:]
	//// TODO
	//panic("TODO")
	ctx.Set(authContextKey, &Authentication{Username: "kuzznya"})
}

func GetAuth(ctx context.Context) *Authentication {
	return ctx.Value(authContextKey).(*Authentication)
}
