package middlewares

import (
	"net/http"
	"strings"

	"gitxyz/internal/api/auth"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InjectDB stores the database handle in the request context so other
// middlewares (e.g. AuthRequired) can resolve users without re-capturing db.
func InjectDB(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if authorizationHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization format must be Bearer <token>"})
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		userID := claims.UserID
		// Backwards-compat: tokens issued before the user_id claim existed
		// carry an empty UserID but a valid Username. Resolve the ID from the
		// database so downstream code (e.g. issue author) works correctly.
		if userID == "" && claims.Username != "" {
			if db, ok := ctx.MustGet("db").(*gorm.DB); ok {
				if user, err := repository.NewUserRepository(db).FindByUsername(claims.Username); err == nil {
					userID = user.ID.String()
				}
			}
		}

		ctx.Set("user_id", userID)
		ctx.Set("username", claims.Username)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}
