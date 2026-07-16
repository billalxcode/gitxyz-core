package repository

import (
	"errors"
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type SSHKeyRepository interface {
	Create(key *models.SSHKey) error
	FindByID(id string) (models.SSHKey, error)
	FindByUserID(userID string) ([]models.SSHKey, error)
	FindByFingerprint(fingerprint string) (models.SSHKey, error)
	Delete(id string) error
	ExistsByFingerprint(fingerprint string) bool
}

type SSHKeyRepositoryImpl struct {
	db *gorm.DB
}

func NewSSHKeyRepository(db *gorm.DB) *SSHKeyRepositoryImpl {
	return &SSHKeyRepositoryImpl{db: db}
}

func (r *SSHKeyRepositoryImpl) Create(key *models.SSHKey) error {
	return r.db.Create(key).Error
}

func (r *SSHKeyRepositoryImpl) FindByID(id string) (models.SSHKey, error) {
	var key models.SSHKey
	result := r.db.First(&key, "id = ?", id)
	if result.Error != nil {
		return models.SSHKey{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.SSHKey{}, gorm.ErrRecordNotFound
	}
	return key, nil
}

func (r *SSHKeyRepositoryImpl) FindByUserID(userID string) ([]models.SSHKey, error) {
	var keys []models.SSHKey
	result := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&keys)
	if result.Error != nil {
		return nil, result.Error
	}
	return keys, nil
}

func (r *SSHKeyRepositoryImpl) FindByFingerprint(fingerprint string) (models.SSHKey, error) {
	var key models.SSHKey
	result := r.db.Where("fingerprint = ?", fingerprint).First(&key)
	if result.Error != nil {
		return models.SSHKey{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.SSHKey{}, gorm.ErrRecordNotFound
	}
	return key, nil
}

func (r *SSHKeyRepositoryImpl) Delete(id string) error {
	result := r.db.Delete(&models.SSHKey{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SSHKeyRepositoryImpl) ExistsByFingerprint(fingerprint string) bool {
	var count int64
	r.db.Model(&models.SSHKey{}).Where("fingerprint = ?", fingerprint).Count(&count)
	return count > 0
}

type PATRepository interface {
	Create(token *models.PersonalAccessToken) error
	FindByID(id string) (models.PersonalAccessToken, error)
	FindByUserID(userID string) ([]models.PersonalAccessToken, error)
	FindByTokenHash(hash string) (models.PersonalAccessToken, error)
	Update(token *models.PersonalAccessToken) error
	Delete(id string) error
}

type PATRepositoryImpl struct {
	db *gorm.DB
}

func NewPATRepository(db *gorm.DB) *PATRepositoryImpl {
	return &PATRepositoryImpl{db: db}
}

func (r *PATRepositoryImpl) Create(token *models.PersonalAccessToken) error {
	return r.db.Create(token).Error
}

func (r *PATRepositoryImpl) FindByID(id string) (models.PersonalAccessToken, error) {
	var token models.PersonalAccessToken
	result := r.db.First(&token, "id = ?", id)
	if result.Error != nil {
		return models.PersonalAccessToken{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.PersonalAccessToken{}, gorm.ErrRecordNotFound
	}
	return token, nil
}

func (r *PATRepositoryImpl) FindByUserID(userID string) ([]models.PersonalAccessToken, error) {
	var tokens []models.PersonalAccessToken
	result := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

func (r *PATRepositoryImpl) FindByTokenHash(hash string) (models.PersonalAccessToken, error) {
	var token models.PersonalAccessToken
	result := r.db.Where("token_hash = ?", hash).First(&token)
	if result.Error != nil {
		return models.PersonalAccessToken{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.PersonalAccessToken{}, gorm.ErrRecordNotFound
	}
	return token, nil
}

func (r *PATRepositoryImpl) Update(token *models.PersonalAccessToken) error {
	return r.db.Save(token).Error
}

func (r *PATRepositoryImpl) Delete(id string) error {
	result := r.db.Delete(&models.PersonalAccessToken{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

var _ = errors.New
