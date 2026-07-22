package dto

import (
	"gitxyz/internal/models"
	"time"
)

type PatchResponse struct {
	ID             string         `json:"id"`
	RepoID         string         `json:"repo_id"`
	Number         int            `json:"number"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	SourceBranch   string         `json:"source_branch"`
	TargetBranch   string         `json:"target_branch"`
	Author         UserResponse   `json:"author"`
	State          string         `json:"state"`
	BaseSHA        string         `json:"base_sha"`
	HeadSHA        string         `json:"head_sha"`
	MergeCommitSHA *string        `json:"merge_commit_sha"`
	IsMergeable    *bool          `json:"is_mergeable"`
	MergedAt       *time.Time     `json:"merged_at"`
	ClosedAt       *time.Time     `json:"closed_at"`
	Reviewers      []UserResponse `json:"reviewers"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func ToPatchResponse(patch *models.PatchRequest) PatchResponse {
	reviewers := make([]UserResponse, 0, len(patch.Reviewers))
	for i := range patch.Reviewers {
		if patch.Reviewers[i].User != nil && patch.Reviewers[i].User.ID != [16]byte{} {
			reviewers = append(reviewers, ToUserResponse(patch.Reviewers[i].User))
		}
	}

	var author UserResponse
	if patch.Author != nil {
		author = ToUserResponse(patch.Author)
	}

	return PatchResponse{
		ID:             patch.ID.String(),
		RepoID:         patch.RepoID,
		Number:         patch.Number,
		Title:          patch.Title,
		Body:           patch.Body,
		SourceBranch:   patch.SourceBranch,
		TargetBranch:   patch.TargetBranch,
		Author:         author,
		State:          patch.State,
		BaseSHA:        patch.BaseSHA,
		HeadSHA:        patch.HeadSHA,
		MergeCommitSHA: patch.MergeCommitSHA,
		IsMergeable:    patch.IsMergeable,
		MergedAt:       patch.MergedAt,
		ClosedAt:       patch.ClosedAt,
		Reviewers:      reviewers,
		CreatedAt:      patch.CreatedAt,
		UpdatedAt:      patch.UpdatedAt,
	}
}

func ToPatchResponseSlice(list []models.PatchRequest) []PatchResponse {
	out := make([]PatchResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPatchResponse(&list[i]))
	}
	return out
}

type PatchCommitResponse struct {
	SHA     string    `json:"sha"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Email   string    `json:"email"`
	Date    time.Time `json:"date"`
}

func ToPatchCommitResponse(c *models.PatchCommit) PatchCommitResponse {
	return PatchCommitResponse{
		SHA:     c.SHA,
		Message: c.Message,
		Author:  c.AuthorName,
		Email:   c.AuthorEmail,
		Date:    c.AuthorDate,
	}
}

func ToPatchCommitResponseSlice(list []models.PatchCommit) []PatchCommitResponse {
	out := make([]PatchCommitResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPatchCommitResponse(&list[i]))
	}
	return out
}

type PatchFileResponse struct {
	FilePath string `json:"file_path"`
	Status   string `json:"status"`
	Diff     string `json:"diff"`
}

func ToPatchFileResponse(f *models.PatchFile) PatchFileResponse {
	return PatchFileResponse{
		FilePath: f.FilePath,
		Status:   f.Status,
		Diff:     f.Diff,
	}
}

func ToPatchFileResponseSlice(list []models.PatchFile) []PatchFileResponse {
	out := make([]PatchFileResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPatchFileResponse(&list[i]))
	}
	return out
}

type PatchReviewResponse struct {
	ID        string       `json:"id"`
	PatchID   string       `json:"patch_id"`
	Author    UserResponse `json:"author"`
	State     string       `json:"state"`
	Body      string       `json:"body"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func ToPatchReviewResponse(r *models.PatchReview) PatchReviewResponse {
	var author UserResponse
	if r.Author != nil {
		author = ToUserResponse(r.Author)
	}
	return PatchReviewResponse{
		ID:        r.ID.String(),
		PatchID:   r.PatchID,
		Author:    author,
		State:     r.State,
		Body:      r.Body,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func ToPatchReviewResponseSlice(list []models.PatchReview) []PatchReviewResponse {
	out := make([]PatchReviewResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPatchReviewResponse(&list[i]))
	}
	return out
}

type PatchCommentResponse struct {
	ID        string       `json:"id"`
	PatchID   string       `json:"patch_id"`
	Author    UserResponse `json:"author"`
	Body      string       `json:"body"`
	FilePath  *string      `json:"file_path"`
	Line      *int         `json:"line"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func ToPatchCommentResponse(c *models.PatchComment) PatchCommentResponse {
	var author UserResponse
	if c.Author != nil {
		author = ToUserResponse(c.Author)
	}
	return PatchCommentResponse{
		ID:        c.ID.String(),
		PatchID:   c.PatchID,
		Author:    author,
		Body:      c.Body,
		FilePath:  c.FilePath,
		Line:      c.Line,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ToPatchCommentResponseSlice(list []models.PatchComment) []PatchCommentResponse {
	out := make([]PatchCommentResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPatchCommentResponse(&list[i]))
	}
	return out
}
