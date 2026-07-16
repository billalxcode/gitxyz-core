package controllers

import (
	"net/http"

	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController interface {
	ListSSHKeys(ctx *gin.Context)
	AddSSHKey(ctx *gin.Context)
	DeleteSSHKey(ctx *gin.Context)
	ListTokens(ctx *gin.Context)
	CreateToken(ctx *gin.Context)
	DeleteToken(ctx *gin.Context)
}

type UserControllerImpl struct {
	service services.UserService
	db      *gorm.DB
}

func NewUserController(db *gorm.DB) UserController {
	return &UserControllerImpl{
		service: services.NewUserService(db),
		db:      db,
	}
}

func (c *UserControllerImpl) ListSSHKeys(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	keys, err := c.service.ListSSHKeys(userID.(string))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ssh keys fetched",
		"data":    response.ToSSHKeyResponseSlice(keys),
	})
}

func (c *UserControllerImpl) AddSSHKey(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	var request dto.AddSSHKeyRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := c.service.AddSSHKey(userID.(string), request.Title, request.PublicKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "ssh key added",
		"data":    response.ToSSHKeyResponse(&key),
	})
}

func (c *UserControllerImpl) DeleteSSHKey(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	keyID := ctx.Param("id")
	if keyID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "key id is required"})
		return
	}

	if err := c.service.DeleteSSHKey(userID.(string), keyID); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ssh key deleted"})
}

func (c *UserControllerImpl) ListTokens(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	tokens, err := c.service.ListTokens(userID.(string))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "tokens fetched",
		"data":    response.ToTokenResponseSlice(tokens),
	})
}

func (c *UserControllerImpl) CreateToken(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	var request dto.CreateTokenRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, plain, err := c.service.CreateToken(userID.(string), request.Name, request.Scopes)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":      "token created",
		"data":         response.ToTokenResponse(&token),
		"access_token": plain,
	})
}

func (c *UserControllerImpl) DeleteToken(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	tokenID := ctx.Param("id")
	if tokenID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "token id is required"})
		return
	}

	if err := c.service.DeleteToken(userID.(string), tokenID); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "token deleted"})
}
