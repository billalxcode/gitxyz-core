package repository

import (
	"errors"
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type RepoRepository interface {
	Create(repo *models.Repository) error
	FindById(id string) (repo models.Repository, err error)
	FindByName(name string) (repo models.Repository, err error)
	FindByUserAndName(userID, name string) (repo models.Repository, err error)
	ListByUser(userID string, dest *[]models.Repository) error
	Update(repo *models.Repository) error
	Delete(id string) error
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

func (r *RepoRepositoryImpl) FindByName(name string) (repo models.Repository, err error) {
	result := r.db.Where("name = ?", name).First(&repo)

	return check(result, &repo)
}

func (r *RepoRepositoryImpl) ExistsByName(name string) bool {
	var count int64
	r.db.Model(&models.Repository{}).Where("name = ?", name).Count(&count)
	return count > 0
}

func (r *RepoRepositoryImpl) FindByUserAndName(userID, name string) (repo models.Repository, err error) {
	result := r.db.Where("user_id = ? AND name = ?", userID, name).First(&repo)
	return check(result, &repo)
}

func (r *RepoRepositoryImpl) ListByUser(userID string, dest *[]models.Repository) error {
	return r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(dest).Error
}

func (r *RepoRepositoryImpl) Update(repo *models.Repository) error {
	return r.db.Model(repo).Updates(map[string]interface{}{
		"description": repo.Description,
		"is_private":  repo.IsPrivate,
		"is_active":   repo.IsActive,
	}).Error
}

func (r *RepoRepositoryImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Repository{}).Error
}
