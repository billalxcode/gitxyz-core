package services

import (
	"fmt"
	"gitxyz/modules/githttp/helper"
	"gitxyz/pkg/git"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) GetInfoRefs(ctx *gin.Context) {
	options := helper.MakeOptionsFromContext(ctx)

	physical_path, err := options.EnsureRepositoryStorage()
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	authorized := s.Authorize(ctx, options)
	if !authorized {
		return
	}

	cmd := git.NewCommand()
	isBareRepository, err := cmd.IsBareRepository(physical_path)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	if !isBareRepository {
		if _, err := cmd.InitBare(physical_path); err != nil {
			ctx.AbortWithError(500, err)
			return
		}
	}

	refs, err := cmd.ReceivePack(physical_path)
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
