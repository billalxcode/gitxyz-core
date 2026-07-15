package controllers

import (
	"net/http"

	"gitxyz/internal/api/services"
	"gitxyz/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RepoController interface {
	Create(ctx *gin.Context)
}

type RepoControllerImpl struct {
	service services.RepoService
	db      *gorm.DB
}

func NewRepoController(db *gorm.DB) RepoController {
	service := services.NewRepoService(db)

	return &RepoControllerImpl{
		service: service,
		db:      db,
	}
}

func (c *RepoControllerImpl) Create(ctx *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	repo := &models.Repository{
		Name:        request.Name,
		Description: request.Description,
		IsPrivate:   request.IsPrivate,
		IsActive:    true,
		UserID:      userID.(string),
	}

	if err := c.service.CreateRepository(repo); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "repository created", "data": repo})
}
