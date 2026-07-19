package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:uniq_repo_member"`
	RepoID    string    `json:"repo_id" gorm:"type:uuid;not null;uniqueIndex:uniq_repo_member"`
	Role      string    `json:"role" gorm:"size:20;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (m *RepositoryMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (RepositoryMember) TableName() string { return "repository_members" }

// Repository (per-repo collaborator) role constants.
const (
	RepoRoleOwner      = "owner"
	RepoRoleMaintainer = "maintainer"
	RepoRoleTriager    = "triager"
	RepoRoleReader     = "reader"
	RepoRoleGuest      = "guest"
)

// ValidRepoRole reports whether role is a recognized repository role.
func ValidRepoRole(role string) bool {
	switch role {
	case RepoRoleOwner, RepoRoleMaintainer, RepoRoleTriager, RepoRoleReader, RepoRoleGuest:
		return true
	}
	return false
}
