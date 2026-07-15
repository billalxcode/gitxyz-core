package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
)

func (r *RoutesImpl) RegisterAuth() {
	controller := controllers.NewAuthController(r.db)

	routes := r.engine.Group("/api/auth")
	publicRoutes := routes.Group("/")
	protectedRoutes := routes.Group("/")
	protectedRoutes.Use(middlewares.AuthRequired())

	publicRoutes.POST("/register", controller.Register)
	publicRoutes.POST("/login", controller.Login)
	publicRoutes.POST("/refresh-token", controller.RefreshToken)
	publicRoutes.POST("/send-verification-email", controller.SendVerificationEmail)
	publicRoutes.POST("/verify-email", controller.VerifyEmail)
	publicRoutes.POST("/send-reset-password", controller.SendPasswordReset)
	publicRoutes.POST("/reset-password", controller.ResetPassword)

	protectedRoutes.POST("/logout", controller.Logout)
	protectedRoutes.GET("/me", controller.Profile)
}
