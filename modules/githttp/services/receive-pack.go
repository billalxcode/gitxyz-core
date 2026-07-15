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

	options := helper.MakeOptionsFromContext(ctx)

	ctx.Header("Content-Type", "application/x-git-receive-pack-result")

	cmd := git.NewCommand()
	err := cmd.ReceivePackRPC(options.GetRepositoryStorage(), ctx.Request.Body, ctx.Writer)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
