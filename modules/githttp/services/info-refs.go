package services

import (
	"fmt"
	"net/http"

	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) GetInfoRefs(ctx *gin.Context) {
	options := helper.MakeOptionsFromContext(ctx, s.db)

	repoID := ctx.GetString("repo_id")
	if repoID == "" {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	storagePath, err := options.EnsureRepositoryStorage(repoID)
	if err != nil {
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
		ctx.AbortWithError(500, err)
		return
	}

	if !isBareRepository {
		if _, err := cmd.InitBare(storagePath); err != nil {
			ctx.AbortWithError(500, err)
			return
		}
	}

	refs, err := cmd.ReceivePack(storagePath)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", options.ServiceType))
	ctx.Writer.Write(helper.PacketWrite(
		"# service=git-" + string(options.ServiceType) + "\n",
	))
	ctx.Writer.Write([]byte("0000"))
	ctx.Writer.Write(refs)
}
