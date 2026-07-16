package services

import (
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

type Permission interface {
	CanRead(username string, reponame string) bool
	CanWrite(username string, reponame string) bool
}

type PermissionImpl struct {
	db *gorm.DB
}

func NewPermission(db *gorm.DB) Permission {
	return &PermissionImpl{
		db: db,
	}
}

// CanRead reports whether username may fetch/clone reponame.
// Public repositories are readable by anyone, including anonymous users
// (username == ""). Private repositories require the owner.
func (a *PermissionImpl) CanRead(username string, reponame string) bool {
	repoRepo := repository.NewRepoRepository(a.db)
	repo, err := repoRepo.FindByName(reponame)
	if err != nil {
		return false
	}

	if !repo.IsPrivate {
		return true
	}

	// Private: only the owner may read.
	return username != "" && repo.UserID != "" && ownerMatches(a.db, username, repo.UserID)
}

// CanWrite reports whether username may push to reponame.
// Per project rules, public repositories accept pushes from anyone
// (including anonymous users). Private repositories require the owner.
func (a *PermissionImpl) CanWrite(username string, reponame string) bool {
	repoRepo := repository.NewRepoRepository(a.db)
	repo, err := repoRepo.FindByName(reponame)
	if err != nil {
		return false
	}

	if !repo.IsPrivate {
		return true
	}

	return username != "" && repo.UserID != "" && ownerMatches(a.db, username, repo.UserID)
}

// ownerMatches resolves username to its user ID and compares with repo.UserID.
func ownerMatches(db *gorm.DB, username, repoUserID string) bool {
	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.FindByUsername(username)
	if err != nil {
		return false
	}
	return user.ID.String() == repoUserID
}
