package bootstrap

import "github.com/gin-gonic/gin"

func Initialize() {
	gin.SetMode(gin.DebugMode)

	db := NewDatabase()

	server := gin.New()
	server.Use(gin.Logger())

	MakeRouter(server, db)

	server.Run()
}
