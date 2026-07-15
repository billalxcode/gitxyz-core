package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
)

func (r *RoutesImpl) RegisterAuth() {
	controller := controllers.NewAuthController(r.db)

	routes := r.engine.Group("/api/auth")
	routes.POST("/register", controller.Register)
	routes.POST("/login", controller.Login)
	routes.POST("/logout", middlewares.AuthRequired(), controller.Register)
	routes.GET("/me", middlewares.AuthRequired(), controller.Profile)
	routes.POST("/send-verification-email", controller.Register)
	routes.POST("/verify-email", controller.Register)
	routes.POST("/send-reset-password", controller.Register)
	routes.POST("/reset-password", controller.Register)
}
