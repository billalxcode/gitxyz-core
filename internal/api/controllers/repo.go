package controllers

import (
	"net/http"

	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"
	"gitxyz/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RepoController interface {
	Create(ctx *gin.Context)
	Get(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)

	ListCollaborators(ctx *gin.Context)
	AddCollaborator(ctx *gin.Context)
	UpdateCollaborator(ctx *gin.Context)
	RemoveCollaborator(ctx *gin.Context)

	ListPolicies(ctx *gin.Context)
	AddPolicy(ctx *gin.Context)
	RemovePolicy(ctx *gin.Context)
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
	owner, _ := ctx.Get("username")

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

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "repository created",
		"data":    response.ToRepositoryResponse(repo, owner.(string)),
	})
}

func (c *RepoControllerImpl) Get(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	repo, err := c.service.GetRepository(owner, name)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "repository fetched",
		"data":    response.ToRepositoryResponse(repo, owner),
	})
}

func (c *RepoControllerImpl) Update(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	var request struct {
		Description string `json:"description"`
		IsPrivate   *bool  `json:"is_private"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patch := &models.Repository{
		Description: request.Description,
	}
	if request.IsPrivate != nil {
		patch.IsPrivate = *request.IsPrivate
	}
	if request.IsActive != nil {
		patch.IsActive = *request.IsActive
	}

	repo, err := c.service.UpdateRepository(owner, name, patch)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "repository updated",
		"data":    response.ToRepositoryResponse(repo, owner),
	})
}

func (c *RepoControllerImpl) Delete(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	if err := c.service.DeleteRepository(owner, name); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "repository deleted"})
}

func (c *RepoControllerImpl) ListCollaborators(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	members, err := c.service.ListCollaborators(owner, name)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "collaborators fetched",
		"data":    response.ToCollaboratorResponseSlice(members),
	})
}

func (c *RepoControllerImpl) AddCollaborator(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	var request dto.CollaboratorRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := c.service.AddCollaborator(owner, name, request.Username, request.Role)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "collaborator added",
		"data":    response.ToCollaboratorResponse(member),
	})
}

func (c *RepoControllerImpl) UpdateCollaborator(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")
	username := ctx.Param("username")

	var request struct {
		Role string `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := c.service.UpdateCollaborator(owner, name, username, request.Role)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "collaborator updated",
		"data":    response.ToCollaboratorResponse(member),
	})
}

func (c *RepoControllerImpl) RemoveCollaborator(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")
	username := ctx.Param("username")

	if err := c.service.RemoveCollaborator(owner, name, username); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "collaborator removed"})
}

func (c *RepoControllerImpl) ListPolicies(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	policies, err := c.service.ListPolicies(owner, name)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "policies fetched",
		"data":    response.ToPolicyResponseSlice(policies),
	})
}

func (c *RepoControllerImpl) AddPolicy(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")

	var request dto.PolicyRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resourceID := request.ResourceID
	if resourceID == "" {
		resourceID = "*"
	}

	pol, err := c.service.AddPolicy(owner, name, request.SubjectType, request.SubjectID, request.Action, resourceID, request.Effect)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "policy added",
		"data":    response.ToPolicyResponse(pol),
	})
}

func (c *RepoControllerImpl) RemovePolicy(ctx *gin.Context) {
	owner := ctx.Param("owner")
	name := ctx.Param("reponame")
	policyID := ctx.Param("id")

	if err := c.service.RemovePolicy(owner, name, policyID); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "policy removed"})
}
