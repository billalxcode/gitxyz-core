package controllers

import (
	"net/http"
	"strconv"

	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"
	"gitxyz/internal/api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PatchController interface {
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Get(ctx *gin.Context)
	Update(ctx *gin.Context)
	Refresh(ctx *gin.Context)
	Merge(ctx *gin.Context)

	ListCommits(ctx *gin.Context)
	ListFiles(ctx *gin.Context)

	AssignReviewer(ctx *gin.Context)
	UnassignReviewer(ctx *gin.Context)
	ListReviewers(ctx *gin.Context)

	SubmitReview(ctx *gin.Context)
	ListReviews(ctx *gin.Context)

	CreateComment(ctx *gin.Context)
	ListComments(ctx *gin.Context)
}

type PatchControllerImpl struct {
	service services.PatchService
}

func NewPatchController(db *gorm.DB) PatchController {
	return &PatchControllerImpl{
		service: services.NewPatchService(db),
	}
}

func (c *PatchControllerImpl) ownerRepo(ctx *gin.Context) (string, string) {
	return ctx.Param("owner"), ctx.Param("reponame")
}

func (c *PatchControllerImpl) number(ctx *gin.Context) (int, bool) {
	n, err := strconv.Atoi(ctx.Param("number"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid patch number"})
		return 0, false
	}
	return n, true
}

func (c *PatchControllerImpl) List(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	patches, err := c.service.ListPatches(owner, repo)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "patches listed", response.ToPatchResponseSlice(patches))
}

func (c *PatchControllerImpl) Create(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	var req dto.CreatePatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID := ctx.GetString("user_id")
	patch, err := c.service.CreatePatch(owner, repo, authorID, req.Title, req.Body, req.SourceBranch, req.TargetBranch)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteCreated(ctx, "patch created", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) Get(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	patch, err := c.service.GetPatch(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "patch found", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) Update(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	var req dto.UpdatePatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title, body, state := "", "", ""
	if req.Title != nil {
		title = *req.Title
	}
	if req.Body != nil {
		body = *req.Body
	}
	if req.State != nil {
		state = *req.State
	}
	patch, err := c.service.UpdatePatch(owner, repo, n, title, body, state)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteOK(ctx, "patch updated", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) Refresh(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	patch, err := c.service.RefreshPatch(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteOK(ctx, "patch refreshed", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) Merge(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	mergerID := ctx.GetString("user_id")
	patch, err := c.service.MergePatch(owner, repo, n, mergerID)
	if err != nil {
		response.WriteError(ctx, http.StatusConflict, err.Error())
		return
	}
	response.WriteOK(ctx, "patch merged", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) ListCommits(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	commits, err := c.service.ListCommits(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "commits listed", response.ToPatchCommitResponseSlice(commits))
}

func (c *PatchControllerImpl) ListFiles(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	files, err := c.service.ListFiles(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "files listed", response.ToPatchFileResponseSlice(files))
}

func (c *PatchControllerImpl) AssignReviewer(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	var req dto.AssignPatchReviewerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	patch, err := c.service.AssignReviewer(owner, repo, n, req.Username)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteOK(ctx, "reviewer assigned", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) UnassignReviewer(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	username := ctx.Param("username")
	patch, err := c.service.UnassignReviewer(owner, repo, n, username)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteOK(ctx, "reviewer unassigned", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) ListReviewers(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	reviewers, err := c.service.ListReviewers(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "reviewers listed", response.ToUserResponseSlice(reviewers))
}

func (c *PatchControllerImpl) SubmitReview(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	var req dto.SubmitPatchReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	authorID := ctx.GetString("user_id")
	patch, err := c.service.SubmitReview(owner, repo, n, authorID, req.State, req.Body)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteOK(ctx, "review submitted", response.ToPatchResponse(patch))
}

func (c *PatchControllerImpl) ListReviews(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	reviews, err := c.service.ListReviews(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "reviews listed", response.ToPatchReviewResponseSlice(reviews))
}

func (c *PatchControllerImpl) CreateComment(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	var req dto.CreatePatchCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authorID := ctx.GetString("user_id")
	comment, err := c.service.CreateComment(owner, repo, n, authorID, req.Body, req.FilePath, req.Line)
	if err != nil {
		response.WriteError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteCreated(ctx, "comment created", response.ToPatchCommentResponse(comment))
}

func (c *PatchControllerImpl) ListComments(ctx *gin.Context) {
	owner, repo := c.ownerRepo(ctx)
	n, ok := c.number(ctx)
	if !ok {
		return
	}
	comments, err := c.service.ListComments(owner, repo, n)
	if err != nil {
		response.WriteError(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.WriteOK(ctx, "comments listed", response.ToPatchCommentResponseSlice(comments))
}
