package services

import (
	"errors"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

type RepoService interface {
	CreateRepository(repo *models.Repository) error
}

type RepoServiceImpl struct {
	Repository repository.RepoRepository
}

func NewRepoService(db *gorm.DB) RepoService {
	return &RepoServiceImpl{Repository: repository.NewRepoRepository(db)}
}

func (s *RepoServiceImpl) CreateRepository(repo *models.Repository) error {
	if repo.Name == "" {
		return errors.New("repository name is required")
	}
	if s.Repository.ExistsByName(repo.Name) {
		return errors.New("repository already exists")
	}
	if repo.UserID == "" {
		return errors.New("user id is required")
	}

	// The on-disk path is derived from repo.ID at runtime (volume_path/<repoID>),
	// so no physical_path column is stored.
	return s.Repository.Create(repo)
}
