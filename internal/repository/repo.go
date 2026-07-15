package repository

import (
	"errors"
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type RepoRepository interface {
	FindById(id string) (repo models.Repository, err error)
	FindByPhysicalPath(path string) (repo models.Repository, err error)
}

type RepoRepositoryImpl struct {
	db *gorm.DB
}

func NewRepoRepository(db *gorm.DB) *RepoRepositoryImpl {
	return &RepoRepositoryImpl{
		db: db,
	}
}

func check(result *gorm.DB) (repo models.Repository, err error) {
	if result.Error != nil {
		return models.Repository{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.Repository{}, errors.New("no repository found")
	}

	return repo, nil
}

func (r *RepoRepositoryImpl) FindById(id string) (repo models.Repository, err error) {
	result := r.db.Find(&repo, id)

	return check(result)
}

func (r *RepoRepositoryImpl) FindByPhysicalPath(path string) (repo models.Repository, err error) {
	result := r.db.Find(&repo, "physical_path = ?", path)

	return check(result)
}

func (r *RepoRepositoryImpl) FindByName(name string) (repo models.Repository, err error) {
	result := r.db.Find(&repo, "name = ?", name)

	return check(result)
}

func (r *RepoRepositoryImpl) ExistsByName(name string) bool {
	// TODO: Implement check exists repository
	return true // placeholder
}
