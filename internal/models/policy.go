package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Policy struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	SubjectType  string    `json:"subject_type" gorm:"size:20;not null"`
	SubjectID    string    `json:"subject_id" gorm:"size:255; not null"`
	Action       string    `json:"action" gorm:"size:50; not null"`
	ResourceType string    `json:"resource_type" gorm:"size:20; not null"`
	ResourceID   string    `json:"resource_id" gorm:"size:255; not null;default:'*'"`
	Effect       string    `json:"effect" gorm:"size:10; not null;default:'allow'"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (Policy) TableName() string { return "policies" }
