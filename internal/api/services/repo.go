package services

import (
	"errors"
	"gitxyz/internal/helper"
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

	repo.PhysicalPath = helper.GenerateRepositoryPath()

	return s.Repository.Create(repo)
}
