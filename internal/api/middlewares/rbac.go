package middlewares

import (
	"net/http"

	"gitxyz/internal/api/services"
	"gitxyz/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireRole aborts with 403 unless the context role is in allowed.
func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetString("role")
		for _, a := range allowed {
			if role == a {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
	}
}

// RequireScope aborts with 403 unless the credential covers scope.
//   - PAT auth: must carry the scope (or admin:*).
//   - JWT auth: allowed only if the user's system role is equivalent
//     (admin/owner grant everything; maintainer grants user-scoped scopes).
func RequireScope(db *gorm.DB, scope string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetString("token_type") == "pat" {
			scopes := models.ParseScopes(ctx.GetString("token_scopes"))
			for _, s := range scopes {
				if s == scope || s == models.ScopeAdmin {
					ctx.Next()
					return
				}
			}
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient token scope"})
			return
		}

		// JWT user auth: role-equivalence check.
		role := ctx.GetString("role")
		if role == models.RoleAdmin || role == models.RoleOwner {
			ctx.Next()
			return
		}
		// maintainer may perform user-scoped actions.
		if role == models.RoleMaintainer &&
			(scope == models.ScopeUserRead || scope == models.ScopeUserWrite) {
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role for scope"})
	}
}

// CheckPolicy evaluates ABAC for a specific resource action.
func CheckPolicy(db *gorm.DB, action, resourceType, resourceID string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetString("role")
		userID := ctx.GetString("user_id")

		if ok, err := services.EvaluatePolicy(db, "role", role, action, resourceType, resourceID); err == nil && ok {
			ctx.Next()
			return
		}
		if userID != "" {
			if ok, err := services.EvaluatePolicy(db, "user", userID, action, resourceType, resourceID); err == nil && ok {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "policy denied"})
	}
}
