package services

import (
	"gitxyz/modules/githttp/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	GitService interface {
		Authorize(ctx *gin.Context, options helper.Options) bool
		GetInfoRefs(ctx *gin.Context)
		ReceivePack(ctx *gin.Context)
		UploadPack(ctx *gin.Context)
	}

	GitServiceImpl struct {
		db *gorm.DB
	}
)

func NewService(db *gorm.DB) GitService {
	return &GitServiceImpl{
		db: db,
	}
}
