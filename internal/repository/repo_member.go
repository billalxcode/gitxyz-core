package repository

import (
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type RepoMemberRepository interface {
	Add(m *models.RepositoryMember) error
	FindByUserAndRepo(userID, repoID string) (models.RepositoryMember, error)
	FindByRepo(repoID string) ([]models.RepositoryMember, error)
	UpdateRole(userID, repoID, role string) error
	Remove(userID, repoID string) error
}

type RepoMemberRepositoryImpl struct {
	db *gorm.DB
}

func NewRepoMemberRepository(db *gorm.DB) *RepoMemberRepositoryImpl {
	return &RepoMemberRepositoryImpl{db: db}
}

func (r *RepoMemberRepositoryImpl) Add(m *models.RepositoryMember) error {
	return r.db.Create(m).Error
}

func (r *RepoMemberRepositoryImpl) FindByUserAndRepo(userID, repoID string) (models.RepositoryMember, error) {
	var m models.RepositoryMember
	err := r.db.Where("user_id = ? AND repo_id = ?", userID, repoID).First(&m).Error
	return m, err
}

func (r *RepoMemberRepositoryImpl) FindByRepo(repoID string) ([]models.RepositoryMember, error) {
	var list []models.RepositoryMember
	err := r.db.Where("repo_id = ?", repoID).Find(&list).Error
	return list, err
}

func (r *RepoMemberRepositoryImpl) UpdateRole(userID, repoID, role string) error {
	return r.db.Model(&models.RepositoryMember{}).
		Where("user_id = ? AND repo_id = ?", userID, repoID).
		Update("role", role).Error
}

func (r *RepoMemberRepositoryImpl) Remove(userID, repoID string) error {
	return r.db.Where("user_id = ? AND repo_id = ?", userID, repoID).
		Delete(&models.RepositoryMember{}).Error
}
