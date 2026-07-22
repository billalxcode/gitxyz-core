package controllers

import (
	"net/http"
	"strconv"

	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"
	"gitxyz/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IssueController interface {
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Get(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)

	ListLabels(ctx *gin.Context)
	CreateLabel(ctx *gin.Context)

	ListComments(ctx *gin.Context)
	CreateComment(ctx *gin.Context)

	ListAssignees(ctx *gin.Context)
	AddAssignee(ctx *gin.Context)
	RemoveAssignee(ctx *gin.Context)
}

type IssueControllerImpl struct {
	service services.IssueService
}

func NewIssueController(db *gorm.DB) IssueController {
	return &IssueControllerImpl{
		service: services.NewIssueService(db),
	}
}

func (c *IssueControllerImpl) ownerRepo(ctx *gin.Context) (string, string) {
	return ctx.Param("owner"), ctx.Param("reponame")
}

func (c *IssueControllerImpl) List(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	issues, err := c.service.ListIssues(owner, repo)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToIssueResponseSlice(issues))
}

func (c *IssueControllerImpl) Create(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	var req dto.CreateIssueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID := ctx.GetString("user_id")
	issue, err := c.service.CreateIssue(owner, repo, authorID, req.Title, req.Body, req.Labels)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, response.ToIssueResponse(issue))
}

func (c *IssueControllerImpl) Get(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	issue, err := c.service.GetIssue(owner, repo, number)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToIssueResponse(issue))
}

func (c *IssueControllerImpl) Update(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	var req dto.UpdateIssueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patch := &models.Issue{}
	if req.Title != nil {
		patch.Title = *req.Title
	}
	if req.Body != nil {
		patch.Body = *req.Body
	}
	if req.State != nil {
		patch.State = *req.State
	}

	issue, err := c.service.UpdateIssue(owner, repo, number, patch, req.Labels)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToIssueResponse(issue))
}

func (c *IssueControllerImpl) Delete(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	if err := c.service.DeleteIssue(owner, repo, number); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IssueControllerImpl) ListLabels(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	labels, err := c.service.ListLabels(owner, repo)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToLabelResponseSlice(labels))
}

func (c *IssueControllerImpl) CreateLabel(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	var req dto.CreateLabelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label, err := c.service.CreateLabel(owner, repo, req.Name, req.Color, req.Description)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, response.ToLabelResponse(label))
}

func (c *IssueControllerImpl) ListComments(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	comments, err := c.service.ListComments(owner, repo, number)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToIssueCommentResponseSlice(comments))
}

func (c *IssueControllerImpl) CreateComment(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	var req dto.CreateIssueCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID := ctx.GetString("user_id")
	comment, err := c.service.CreateComment(owner, repo, number, authorID, req.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, response.ToIssueCommentResponse(comment))
}

func (c *IssueControllerImpl) ListAssignees(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	users, err := c.service.ListAssignees(owner, repo, number)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response.ToAssigneeResponseSlice(users))
}

func (c *IssueControllerImpl) AddAssignee(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	var req dto.AssignIssueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.service.AddAssignee(owner, repo, number, req.Username); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IssueControllerImpl) RemoveAssignee(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	number, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue number"})
		return
	}
	username := ctx.Param("username")
	if err := c.service.RemoveAssignee(owner, repo, number, username); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}
