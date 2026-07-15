package controllers

import (
	"fmt"
	"gitxyz/internal/api/auth"
	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"
	"gitxyz/internal/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	user, err := c.service.Login(request.Username, request.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := auth.GenerateToken(&user, "access")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	refreshToken, err := auth.GenerateToken(&user, "refresh")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not generate refresh token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"data":          response.ToUserResponse(&user),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (c *AuthControllerImpl) RefreshToken(ctx *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ParseToken(request.RefreshToken)
	if err != nil || claims.TokenType != "refresh" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	user := models.User{Base: models.Base{ID: userID}, Username: claims.Username, Email: claims.Email}
	accessToken, err := auth.GenerateToken(&user, "access")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (c *AuthControllerImpl) Logout(ctx *gin.Context) {
	authorizationHeader := ctx.GetHeader("Authorization")
	parts := strings.Split(authorizationHeader, " ")
	if len(parts) == 2 {
		auth.RevokeToken(parts[1])
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (c *AuthControllerImpl) SendVerificationEmail(ctx *gin.Context) {

}

func (c *AuthControllerImpl) VerifyEmail(ctx *gin.Context) {

}

func (c *AuthControllerImpl) SendPasswordReset(ctx *gin.Context) {

}

func (c *AuthControllerImpl) ResetPassword(ctx *gin.Context) {

}
