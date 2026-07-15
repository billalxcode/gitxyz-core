package githttp

import (
	"gitxyz/modules/githttp/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	GitHTTPController interface {
		InfoRefs(ctx *gin.Context)
		ReceivePack(ctx *gin.Context)
	}

	gitHTTPController struct {
		service services.GitService
		tx      *gorm.DB
	}
)

func NewController(tx *gorm.DB) GitHTTPController {
	service := services.NewService(tx)

	return &gitHTTPController{
		service: service,
		tx:      tx,
	}
}

func (c *gitHTTPController) InfoRefs(ctx *gin.Context) {
	c.service.GetInfoRefs(ctx)
}

func (c *gitHTTPController) ReceivePack(ctx *gin.Context) {
	c.service.ReceivePack(ctx)
}
