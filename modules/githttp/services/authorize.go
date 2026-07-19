package services

import (
	"errors"
	"log/slog"
	"net/http"

	"gitxyz/internal/logger"
	"gitxyz/internal/repository"
	"gitxyz/modules/githttp/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func authorizeRepo(ctx *gin.Context, db *gorm.DB, reponame string) {
	log := logger.FromGin(ctx)
	repoRepo := repository.NewRepoRepository(db)
	repo, err := repoRepo.FindByName(reponame)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("git authorize: repository not found", slog.String("repo", reponame))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
			return
		}
		log.Error("git authorize: database error", slog.String("repo", reponame), slog.String("error", err.Error()))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	_ = repo
}

func (s *GitServiceImpl) Authorize(ctx *gin.Context, options helper.Options) bool {
	log := logger.FromGin(ctx)
	authorizeRepo(ctx, s.db, options.RepoName)
	if ctx.IsAborted() {
		return false
	}

	permission := NewPermission(s.db)
	username := ctx.GetString("username")
	anonymous := username == ""

	switch options.ServiceType {
	case helper.ServiceTypeReceivePack:
		// Push requires write permission.
		if !permission.CanWrite(ctx, options.RepoName) {
			// Anonymous clients must receive a Basic auth challenge so Git
			// prompts for credentials instead of failing with 403.
			if anonymous {
				log.Info("git authorize: write requires authentication",
					slog.String("repo", options.RepoName))
				ctx.Header("WWW-Authenticate", `Basic realm="Git"`)
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return false
			}
			log.Warn("git authorize: write denied",
				slog.String("repo", options.RepoName),
				slog.String("username", username))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "write access denied"})
			return false
		}
		log.Info("git authorize: write granted",
			slog.String("repo", options.RepoName),
			slog.String("username", username))
	case helper.ServiceTypeUploadPack:
		// Clone/fetch requires read permission.
		if !permission.CanRead(ctx, options.RepoName) {
			// Anonymous clients must receive a Basic auth challenge so Git
			// prompts for credentials instead of failing with 403.
			if anonymous {
				log.Info("git authorize: read requires authentication",
					slog.String("repo", options.RepoName))
				ctx.Header("WWW-Authenticate", `Basic realm="Git"`)
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return false
			}
			log.Warn("git authorize: read denied",
				slog.String("repo", options.RepoName),
				slog.String("username", username))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "read access denied"})
			return false
		}
		log.Info("git authorize: read granted",
			slog.String("repo", options.RepoName),
			slog.String("username", username))
	default:
		log.Warn("git authorize: invalid service type", slog.String("service", string(options.ServiceType)))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid service type"})
		return false
	}

	return true
}
