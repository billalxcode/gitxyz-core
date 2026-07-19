package services

import (
	"fmt"
	"log/slog"
	"net/http"

	"gitxyz/internal/logger"
	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) GetInfoRefs(ctx *gin.Context) {
	log := logger.FromGin(ctx)
	options := helper.MakeOptionsFromContext(ctx, s.db)

	repoID := ctx.GetString("repo_id")
	if repoID == "" {
		log.Warn("git info-refs: repository not found", slog.String("repo", options.RepoName))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	storagePath, err := options.EnsureRepositoryStorage(repoID)
	if err != nil {
		log.Error("git info-refs: storage error", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
		ctx.AbortWithError(500, err)
		return
	}

	authorized := s.Authorize(ctx, options)
	if !authorized {
		return
	}

	cmd := git.NewCommand()
	isBareRepository, err := cmd.IsBareRepository(storagePath)
	if err != nil {
		log.Error("git info-refs: bare check failed", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
		ctx.AbortWithError(500, err)
		return
	}

	if !isBareRepository {
		if _, err := cmd.InitBare(storagePath); err != nil {
			log.Error("git info-refs: init bare failed", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
			ctx.AbortWithError(500, err)
			return
		}
		log.Info("git info-refs: initialized bare repository", slog.String("repo", options.RepoName))
	}

	refs, err := cmd.ReceivePack(storagePath)
	if err != nil {
		log.Error("git info-refs: receive-pack failed", slog.String("repo", options.RepoName), slog.String("error", err.Error()))
		ctx.AbortWithError(500, err)
		return
	}

	log.Info("git info-refs: served advertisement",
		slog.String("repo", options.RepoName),
		slog.String("service", string(options.ServiceType)))

	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", options.ServiceType))
	ctx.Writer.Write(helper.PacketWrite(
		"# service=git-" + string(options.ServiceType) + "\n",
	))
	ctx.Writer.Write([]byte("0000"))
	ctx.Writer.Write(refs)
}
