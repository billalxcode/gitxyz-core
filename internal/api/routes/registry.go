package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoutesImpl struct {
	engine *gin.Engine
	db     *gorm.DB
}

func NewRoutes(engine *gin.Engine, db *gorm.DB) *RoutesImpl {
	return &RoutesImpl{
		engine: engine,
		db:     db,
	}
}
