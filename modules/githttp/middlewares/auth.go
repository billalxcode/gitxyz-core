package middlewares

import (
	"github.com/gin-gonic/gin"
)

func validateUser(username string, password string) bool {
	// TODO: Implement validate user on auth middleware
	return true // placeholder
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username, password, ok := ctx.Request.BasicAuth()
		if !ok || validateUser(username, password) {
			ctx.Header("WWW-Authenticate", "Basic realm=Git")
			ctx.AbortWithStatus(401)
			return
		}
		ctx.Set("username", username)
		ctx.Next()
	}
}
