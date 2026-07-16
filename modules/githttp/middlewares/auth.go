package middlewares

import (
	"gitxyz/internal/helper"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware validates Git HTTP Basic credentials. The credentials may be
// either a username + password, or a username + personal access token (PAT).
// On success it stores the resolved username in the context.
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)
	patRepo := repository.NewPATRepository(db)

	return func(ctx *gin.Context) {
		username, secret, ok := ctx.Request.BasicAuth()
		if !ok {
			// No credentials: leave username empty so downstream
			// authorization can still allow public, read-only access.
			ctx.Set("username", "")
			ctx.Next()
			return
		}

		if validateCredentials(userRepo, patRepo, username, secret) {
			ctx.Set("username", username)
			ctx.Next()
			return
		}

		ctx.Header("WWW-Authenticate", `Basic realm="Git"`)
		ctx.AbortWithStatus(401)
	}
}

// validateCredentials checks a username/secret pair against either the user's
// password or a personal access token.
func validateCredentials(
	userRepo *repository.UserRepositoryImpl,
	patRepo *repository.PATRepositoryImpl,
	username, secret string,
) bool {
	// 1. Try password login.
	if user, err := userRepo.Authenticate(username, secret); err == nil && user.ID != [16]byte{} {
		return true
	}

	// 2. Try personal access token (hashed lookup).
	if _, err := patRepo.FindByTokenHash(helper.HashToken(secret)); err == nil {
		return true
	}

	return false
}
