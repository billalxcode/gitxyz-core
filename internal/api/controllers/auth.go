package controllers

import (
	"fmt"
	"gitxyz/internal/api/auth"
	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"
	"gitxyz/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	AuthController interface {
		Register(ctx *gin.Context)
		Login(ctx *gin.Context)
		Profile(ctx *gin.Context)
		RefreshToken(ctx *gin.Context)
		Logout(ctx *gin.Context)
		SendVerificationEmail(ctx *gin.Context)
		VerifyEmail(ctx *gin.Context)
		SendPasswordReset(ctx *gin.Context)
		ResetPassword(ctx *gin.Context)
	}

	AuthControllerImpl struct {
		service services.AuthService
		tx      *gorm.DB
	}
)

func NewAuthController(tx *gorm.DB) AuthController {
	service := services.NewAuthService(tx)

	return &AuthControllerImpl{
		service: service,
		tx:      tx,
	}
}

func (c *AuthControllerImpl) Register(ctx *gin.Context) {
	var request dto.RegisterRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("mapping to models")
	user := &models.User{
		FullName: request.FullName,
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
		IsActive: true,
		Avatar:   "default.png",
		Bio:      "",
		Location: "",
	}

	fmt.Println("register using service")
	if err := c.service.Register(user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "data": user})
}

func (c *AuthControllerImpl) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.Login(request.UsernameOrEmail, request.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateToken(&user)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Login successful", "data": response.ToUserResponse(&user), "token": token})
}

func (c *AuthControllerImpl) RefreshToken(ctx *gin.Context) {

}

func (c *AuthControllerImpl) Logout(ctx *gin.Context) {

}

func (c *AuthControllerImpl) SendVerificationEmail(ctx *gin.Context) {

}

func (c *AuthControllerImpl) VerifyEmail(ctx *gin.Context) {

}

func (c *AuthControllerImpl) SendPasswordReset(ctx *gin.Context) {

}

func (c *AuthControllerImpl) ResetPassword(ctx *gin.Context) {

}
