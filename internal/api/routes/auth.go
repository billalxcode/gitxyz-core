package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
)

func (r *RoutesImpl) RegisterAuth() {
	authController := controllers.NewAuthController(r.db)
	userController := controllers.NewUserController(r.db)

	// --- Auth routes (ENDPOINT.md §2) ---
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

	// --- User routes (ENDPOINT.md §2) ---
	user := r.engine.Group("/api/user")
	user.Use(middlewares.AuthRequired())

	user.GET("", authController.Profile)         // GET /user — profil sendiri
	user.PATCH("", authController.UpdateProfile) // PATCH /user
	user.POST("/change-password", authController.ChangePassword)

	user.GET("/keys", userController.ListSSHKeys) // GET /user/keys
	user.POST("/keys", userController.AddSSHKey)  // POST /user/keys
	user.DELETE("/keys/:id", userController.DeleteSSHKey)

	user.GET("/tokens", userController.ListTokens)   // GET /user/tokens
	user.POST("/tokens", userController.CreateToken) // POST /user/tokens
	user.DELETE("/tokens/:id", userController.DeleteToken)

	// Public user lookup
	r.engine.GET("/api/users/:username", authController.GetUserByUsername)
}
