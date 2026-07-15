package bootstrap

import (
	"gitxyz/internal/api/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MakeRouter(engine *gin.Engine, db *gorm.DB) {
	router := routes.NewRoutes(engine, db)
	router.RegisterAuth()
	router.RegisterRepositories()
}
