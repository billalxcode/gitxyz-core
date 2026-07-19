package githttp

import (
	apimw "gitxyz/internal/api/middlewares"
	githttpmw "gitxyz/modules/githttp/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, tx *gorm.DB) {
	controller := NewController(tx)

	routes := server.Group("/:username/:reponame")
	routes.Use(apimw.RequestID())
	routes.Use(githttpmw.AuthMiddleware(tx))

	routes.Match([]string{"GET"}, "/info/refs", controller.InfoRefs)
	routes.Match([]string{"POST", "OPTIONS"}, "/git-receive-pack", controller.ReceivePack)
}
