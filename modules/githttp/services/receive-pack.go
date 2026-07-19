package services

import (
	"log/slog"
	"net/http"

	"gitxyz/internal/logger"
	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) ReceivePack(ctx *gin.Context) {
	log := logger.FromGin(ctx)
	options := helper.MakeOptionsFromContext(ctx, s.db)

	repoID := ctx.GetString("repo_id")
	if repoID == "" {
		log.Warn("git receive-pack: repository not found", slog.String("repo", options.RepoName))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	storagePath, err := options.EnsureRepositoryStorage(repoID)
	if err != nil {
		log.Error("git receive-pack: storage error", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Authorize the push (write permission). Anonymous clients receive a 401
	// Basic challenge so Git prompts for credentials.
	if !s.Authorize(ctx, options) {
		return
	}

	log.Info("git receive-pack: push started",
		slog.String("repo", options.RepoName),
		slog.String("username", ctx.GetString("username")))

	ctx.Header("Content-Type", "application/x-git-receive-pack-result")

	cmd := git.NewCommand()
	err = cmd.ReceivePackRPC(storagePath, ctx.Request.Body, ctx.Writer)
	if err != nil {
		log.Error("git receive-pack: push failed",
			slog.String("repo", options.RepoName),
			slog.String("username", ctx.GetString("username")),
			slog.String("error", err.Error()))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Info("git receive-pack: push completed",
		slog.String("repo", options.RepoName),
		slog.String("username", ctx.GetString("username")))
}
