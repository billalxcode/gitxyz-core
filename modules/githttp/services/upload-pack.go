package services

import (
	"gitxyz/internal/logger"

	"github.com/gin-gonic/gin"
)

func (s *GitServiceImpl) UploadPack(ctx *gin.Context) {
	log := logger.FromGin(ctx)
	log.Warn("git upload-pack: not implemented")
	ctx.AbortWithStatusJSON(501, gin.H{"error": "upload-pack not implemented"})
}
