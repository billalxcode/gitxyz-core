package repository

import (
	"errors"
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type RepoRepository interface {
	Create(repo *models.Repository) error
	FindById(id string) (repo models.Repository, err error)
	FindByPhysicalPath(path string) (repo models.Repository, err error)
	FindByName(name string) (repo models.Repository, err error)
	ExistsByName(name string) bool
}

type RepoRepositoryImpl struct {
	db *gorm.DB
}

func NewRepoRepository(db *gorm.DB) *RepoRepositoryImpl {
	return &RepoRepositoryImpl{
		db: db,
	}
}

func check(result *gorm.DB, repo *models.Repository) (models.Repository, error) {
	if result.Error != nil {
		return models.Repository{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.Repository{}, errors.New("no repository found")
	}

	if repo == nil {
		return models.Repository{}, nil
	}

	return *repo, nil
}

func (r *RepoRepositoryImpl) Create(repo *models.Repository) error {
	return r.db.Create(repo).Error
}

func (r *RepoRepositoryImpl) FindById(id string) (repo models.Repository, err error) {
	result := r.db.First(&repo, "id = ?", id)

	return check(result, &repo)
}

func (r *RepoRepositoryImpl) FindByPhysicalPath(path string) (repo models.Repository, err error) {
	result := r.db.Where("physical_path = ?", path).First(&repo)

	return check(result, &repo)
}

func (r *RepoRepositoryImpl) FindByName(name string) (repo models.Repository, err error) {
	result := r.db.Where("name = ?", name).First(&repo)

	return check(result, &repo)
}

func (r *RepoRepositoryImpl) ExistsByName(name string) bool {
	var count int64
	r.db.Model(&models.Repository{}).Where("name = ?", name).Count(&count)
	return count > 0
}
