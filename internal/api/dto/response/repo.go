package dto

import (
	"gitxyz/internal/models"
	"time"
)

// RepositoryResponse is the safe, filtered projection of a Repository that is
// returned to API clients. It intentionally omits the nested User association
// and GORM internals (e.g. DeletedAt) to avoid leaking zero-value structs.
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

// ToRepositoryResponse builds the response. owner is the repository path
// prefix (typically the owner's username), e.g. "gitxyz" for "gitxyz/gitxyz".
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
