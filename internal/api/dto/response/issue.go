package dto

import (
	"gitxyz/internal/models"
	"time"
)

type IssueResponse struct {
	ID        string          `json:"id"`
	Number    int             `json:"number"`
	RepoID    string          `json:"repo_id"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	State     string          `json:"state"`
	Author    UserResponse    `json:"author"`
	Assignee  *UserResponse   `json:"assignee,omitempty"`
	Labels    []LabelResponse `json:"labels"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	ClosedAt  *time.Time      `json:"closed_at"`
}

func ToIssueResponse(issue *models.Issue) IssueResponse {
	labels := make([]LabelResponse, 0, len(issue.Labels))
	for i := range issue.Labels {
		labels = append(labels, ToLabelResponse(&issue.Labels[i]))
	}

	var assignee *UserResponse
	if issue.Assignee != nil && issue.Assignee.ID != [16]byte{} {
		a := ToUserResponse(issue.Assignee)
		assignee = &a
	}

	var author UserResponse
	if issue.Author != nil {
		author = ToUserResponse(issue.Author)
	}

	return IssueResponse{
		ID:        issue.ID.String(),
		Number:    issue.Number,
		RepoID:    issue.RepoID,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     issue.State,
		Author:    author,
		Assignee:  assignee,
		Labels:    labels,
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
		ClosedAt:  issue.ClosedAt,
	}
}

func ToIssueResponseSlice(list []models.Issue) []IssueResponse {
	out := make([]IssueResponse, 0, len(list))
	for i := range list {
		out = append(out, ToIssueResponse(&list[i]))
	}
	return out
}

type LabelResponse struct {
	ID          string `json:"id"`
	RepoID      string `json:"repo_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func ToLabelResponse(label *models.Label) LabelResponse {
	return LabelResponse{
		ID:          label.ID.String(),
		RepoID:      label.RepoID,
		Name:        label.Name,
		Color:       label.Color,
		Description: label.Description,
	}
}

func ToLabelResponseSlice(list []models.Label) []LabelResponse {
	out := make([]LabelResponse, 0, len(list))
	for i := range list {
		out = append(out, ToLabelResponse(&list[i]))
	}
	return out
}

type IssueCommentResponse struct {
	ID        string       `json:"id"`
	IssueID   string       `json:"issue_id"`
	Author    UserResponse `json:"author"`
	Body      string       `json:"body"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func ToIssueCommentResponse(comment *models.IssueComment) IssueCommentResponse {
	var author UserResponse
	if comment.Author != nil {
		author = ToUserResponse(comment.Author)
	}
	return IssueCommentResponse{
		ID:        comment.ID.String(),
		IssueID:   comment.IssueID,
		Author:    author,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

func ToIssueCommentResponseSlice(list []models.IssueComment) []IssueCommentResponse {
	out := make([]IssueCommentResponse, 0, len(list))
	for i := range list {
		out = append(out, ToIssueCommentResponse(&list[i]))
	}
	return out
}

type AssigneeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func ToAssigneeResponseSlice(list []models.User) []AssigneeResponse {
	out := make([]AssigneeResponse, 0, len(list))
	for i := range list {
		out = append(out, AssigneeResponse{
			ID:       list[i].ID.String(),
			Username: list[i].Username,
			Email:    list[i].Email,
		})
	}
	return out
}
