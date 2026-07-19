package repository

import (
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

type PolicyRepository interface {
	Add(p *models.Policy) error
	Remove(id string) error
	// FindApplicable returns policies matching the subject/action/resource triple.
	FindApplicable(subjectType, subjectID, action, resourceType, resourceID string) ([]models.Policy, error)
}

type PolicyRepositoryImpl struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) *PolicyRepositoryImpl {
	return &PolicyRepositoryImpl{db: db}
}

func (p *PolicyRepositoryImpl) Add(pol *models.Policy) error {
	return p.db.Create(pol).Error
}

func (p *PolicyRepositoryImpl) Remove(id string) error {
	return p.db.Where("id = ?", id).Delete(&models.Policy{}).Error
}

func (p *PolicyRepositoryImpl) FindApplicable(subjectType, subjectID, action, resourceType, resourceID string) ([]models.Policy, error) {
	var list []models.Policy
	err := p.db.
		Where("subject_type = ? AND subject_id = ?", subjectType, subjectID).
		Where("action = ? AND resource_type = ?", action, resourceType).
		Where("resource_id = ? OR resource_id = '*'", resourceID).
		Find(&list).Error
	return list, err
}
