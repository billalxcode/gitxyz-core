package middlewares

import (
	"log/slog"
	"net/http"

	"gitxyz/internal/helper"
	"gitxyz/internal/logger"
	"gitxyz/internal/repository"
	githttpHelper "gitxyz/modules/githttp/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)
	patRepo := repository.NewPATRepository(db)
	repoRepo := repository.NewRepoRepository(db)

	return func(ctx *gin.Context) {
		log := logger.FromGin(ctx)
		options := githttpHelper.MakeOptionsFromContext(ctx, db)
		username, secret, ok := ctx.Request.BasicAuth()
		if !ok {
			log.Debug("git auth: anonymous request",
				slog.String("repo", options.RepoName))
			ctx.Set("username", "")
			ctx.Set("role", "")
			ctx.Set("token_scopes", "")
			ctx.Set("token_type", "")
		} else if res, authed := authenticate(userRepo, patRepo, username, secret); authed {
			log.Info("git auth: authenticated",
				slog.String("username", username),
				slog.String("token_type", res.TokenType),
				slog.String("role", res.Role))
			ctx.Set("username", username)
			ctx.Set("role", res.Role)
			ctx.Set("token_scopes", res.Scopes)
			ctx.Set("token_type", res.TokenType)
		} else {
			log.Warn("git auth: failed",
				slog.String("username", username))
			ctx.Header("WWW-Authenticate", `Basic realm="Git"`)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		repo, err := repoRepo.FindByName(options.RepoName)
		if err != nil {
			log.Warn("git auth: repository not found",
				slog.String("repo", options.RepoName))
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

// authResult carries the resolved identity metadata.
type authResult struct {
	Role      string
	Scopes    string
	TokenType string
}

func authenticate(
	userRepo *repository.UserRepositoryImpl,
	patRepo *repository.PATRepositoryImpl,
	username, secret string,
) (authResult, bool) {
	if user, err := userRepo.Authenticate(username, secret); err == nil && user.ID != [16]byte{} {
		return authResult{Role: user.Role, TokenType: "jwt"}, true
	}
	if pat, err := patRepo.FindByTokenHash(helper.HashToken(secret)); err == nil {
		return authResult{Role: "", Scopes: pat.Scopes, TokenType: "pat"}, true
	}
	return authResult{}, false
}
