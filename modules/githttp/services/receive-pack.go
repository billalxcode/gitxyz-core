package services

import (
	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) ReceivePack(ctx *gin.Context) {
	log.Println(">>> ReceivePack handler")

	options := helper.MakeOptionsFromContext(ctx, s.db)

	repoID := ctx.GetString("repo_id")
	if repoID == "" {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
		return
	}

	ctx.Header("Content-Type", "application/x-git-receive-pack-result")

	cmd := git.NewCommand()
	err := cmd.ReceivePackRPC(options.GetRepositoryStorage(repoID), ctx.Request.Body, ctx.Writer)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
