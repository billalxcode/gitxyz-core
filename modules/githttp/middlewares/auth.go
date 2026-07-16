package middlewares

import (
	"gitxyz/internal/helper"
	"gitxyz/internal/repository"
	githttpHelper "gitxyz/modules/githttp/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)
	patRepo := repository.NewPATRepository(db)
	repoRepo := repository.NewRepoRepository(db)

	return func(ctx *gin.Context) {
		username, secret, ok := ctx.Request.BasicAuth()
		if !ok {
			ctx.Set("username", "")
		} else if validateCredentials(userRepo, patRepo, username, secret) {
			ctx.Set("username", username)
		} else {
			ctx.Header("WWW-Authenticate", `Basic realm="Git"`)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		options := githttpHelper.MakeOptionsFromContext(ctx, db)
		repo, err := repoRepo.FindByName(options.RepoName)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "repository not found",
			})
			return
		}
		// The on-disk path is derived from the repo ID at runtime
		// (volume_path/<repoID>), so we only need to pass the ID forward.
		ctx.Set("repo_id", repo.ID.String())

		ctx.Next()
	}
}

func validateCredentials(
	userRepo *repository.UserRepositoryImpl,
	patRepo *repository.PATRepositoryImpl,
	username, secret string,
) bool {
	if user, err := userRepo.Authenticate(username, secret); err == nil && user.ID != [16]byte{} {
		return true
	}

	if _, err := patRepo.FindByTokenHash(helper.HashToken(secret)); err == nil {
		return true
	}

	return false
}
