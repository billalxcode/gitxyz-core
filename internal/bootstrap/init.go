package bootstrap

import (
	"gitxyz/internal/api/middlewares"
	"gitxyz/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Initialize() {
	gin.SetMode(gin.DebugMode)

	// Configure structured JSON logging from config (log_level).
	logger.Configure(viper.GetString("log_level"))

	db := NewDatabase()

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middlewares.RequestID())
	server.Use(middlewares.InjectDB(db))

	MakeRouter(server, db)

	server.Run()
}
