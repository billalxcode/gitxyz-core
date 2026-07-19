package services

import (
	"log/slog"
	"net/http"

	"gitxyz/internal/logger"
	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) UploadPack(ctx *gin.Context) {
	log := logger.FromGin(ctx)
	options := helper.MakeOptionsFromContext(ctx, s.db)

	repoID := ctx.GetString("repo_id")
	if repoID == "" {
		log.Warn("git upload-pack: repository not found", slog.String("repo", options.RepoName))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	storagePath, err := options.EnsureRepositoryStorage(repoID)
	if err != nil {
		log.Error("git upload-pack: storage error", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Authorize the fetch/clone (read permission). Anonymous clients receive a
	// 401 Basic challenge so Git prompts for credentials on private repos.
	if !s.Authorize(ctx, options) {
		return
	}

	log.Info("git upload-pack: fetch started",
		slog.String("repo", options.RepoName),
		slog.String("username", ctx.GetString("username")))

	ctx.Header("Content-Type", "application/x-git-upload-pack-result")

	cmd := git.NewCommand()
	if err := cmd.UploadPackRPC(storagePath, ctx.Request.Body, ctx.Writer); err != nil {
		log.Error("git upload-pack: fetch failed",
			slog.String("repo", options.RepoName),
			slog.String("username", ctx.GetString("username")),
			slog.String("error", err.Error()))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Info("git upload-pack: fetch completed",
		slog.String("repo", options.RepoName),
		slog.String("username", ctx.GetString("username")))
}
