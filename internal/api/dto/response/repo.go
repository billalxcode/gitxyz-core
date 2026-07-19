package dto

import (
	"gitxyz/internal/models"
	"time"
)

type RepositoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	IsActive    bool      `json:"is_active"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToRepositoryResponse(repo *models.Repository, owner string) RepositoryResponse {
	fullName := repo.Name
	if owner != "" {
		fullName = owner + "/" + repo.Name
	}
	return RepositoryResponse{
		ID:          repo.ID.String(),
		Name:        repo.Name,
		FullName:    fullName,
		Description: repo.Description,
		IsPrivate:   repo.IsPrivate,
		IsActive:    repo.IsActive,
		UserID:      repo.UserID,
		CreatedAt:   repo.CreatedAt,
		UpdatedAt:   repo.UpdatedAt,
	}
}

func ToRepositoryResponseSlice(repos []models.Repository, owner string) []RepositoryResponse {
	out := make([]RepositoryResponse, 0, len(repos))
	for i := range repos {
		out = append(out, ToRepositoryResponse(&repos[i], owner))
	}
	return out
}

type CollaboratorResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	RepoID string `json:"repo_id"`
	Role   string `json:"role"`
}

func ToCollaboratorResponse(m *models.RepositoryMember) CollaboratorResponse {
	return CollaboratorResponse{
		ID:     m.ID.String(),
		UserID: m.UserID,
		RepoID: m.RepoID,
		Role:   m.Role,
	}
}

func ToCollaboratorResponseSlice(list []models.RepositoryMember) []CollaboratorResponse {
	out := make([]CollaboratorResponse, 0, len(list))
	for i := range list {
		out = append(out, ToCollaboratorResponse(&list[i]))
	}
	return out
}

type PolicyResponse struct {
	ID           string `json:"id"`
	SubjectType  string `json:"subject_type"`
	SubjectID    string `json:"subject_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Effect       string `json:"effect"`
}

func ToPolicyResponse(p *models.Policy) PolicyResponse {
	return PolicyResponse{
		ID:           p.ID.String(),
		SubjectType:  p.SubjectType,
		SubjectID:    p.SubjectID,
		Action:       p.Action,
		ResourceType: p.ResourceType,
		ResourceID:   p.ResourceID,
		Effect:       p.Effect,
	}
}

func ToPolicyResponseSlice(list []models.Policy) []PolicyResponse {
	out := make([]PolicyResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPolicyResponse(&list[i]))
	}
	return out
}
