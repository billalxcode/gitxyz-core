package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
	"gitxyz/internal/models"
)

func (r *RoutesImpl) RegisterAuth() {
	authController := controllers.NewAuthController(r.db)
	userController := controllers.NewUserController(r.db)

	auth := r.engine.Group("/api/auth")
	public := auth.Group("/")
	protected := auth.Group("/")
	protected.Use(middlewares.AuthRequired())

	public.POST("/register", authController.Register)
	public.POST("/login", authController.Login)
	public.POST("/token/refresh", authController.RefreshToken)
	public.POST("/send-verification-email", authController.SendVerificationEmail)
	public.POST("/verify-email", authController.VerifyEmail)
	public.POST("/send-reset-password", authController.SendPasswordReset)
	public.POST("/reset-password", authController.ResetPassword)

	protected.POST("/logout", authController.Logout)
	protected.GET("/me", authController.Profile)

	user := r.engine.Group("/api/user")
	user.Use(middlewares.AuthRequired())

	user.GET("", authController.Profile)
	user.PATCH("", authController.UpdateProfile)
	user.POST("/change-password", authController.ChangePassword)

	user.GET("/keys", userController.ListSSHKeys)
	user.POST("/keys", userController.AddSSHKey)
	user.DELETE("/keys/:id", userController.DeleteSSHKey)

	// Token management requires a PAT-capable credential (or admin/owner role).
	user.Use(middlewares.RequireScope(r.db, models.ScopeUserWrite))
	user.GET("/tokens", userController.ListTokens)
	user.POST("/tokens", userController.CreateToken)
	user.DELETE("/tokens/:id", userController.DeleteToken)

	// Admin-only user lookup.
	admin := r.engine.Group("/api/users")
	admin.Use(middlewares.AuthRequired(), middlewares.RequireRole(models.RoleAdmin, models.RoleOwner))
	admin.GET("/:username", authController.GetUserByUsername)
}
