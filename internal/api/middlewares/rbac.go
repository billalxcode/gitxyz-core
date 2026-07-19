package middlewares

import (
	"net/http"

	"gitxyz/internal/api/services"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

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
		// The repository owner (creator) is always allowed on their own repo,
		// regardless of their system role.
		if isRepoOwner(ctx, db) {
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role for scope"})
	}
}

// isRepoOwner reports whether the authenticated user owns the repository
// referenced by the :owner/:reponame path params.
func isRepoOwner(ctx *gin.Context, db *gorm.DB) bool {
	owner := ctx.Param("owner")
	reponame := ctx.Param("reponame")
	if owner == "" || reponame == "" {
		return false
	}
	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.FindByUsername(owner)
	if err != nil {
		return false
	}
	currentUserID := ctx.GetString("user_id")
	if currentUserID == "" {
		return false
	}
	if user.ID.String() != currentUserID {
		return false
	}
	repoRepo := repository.NewRepoRepository(db)
	_, err = repoRepo.FindByUserAndName(user.ID.String(), reponame)
	return err == nil
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
