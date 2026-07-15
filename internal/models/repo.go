package models

type Repository struct {
	Base

	Name         string `json:"name" gorm:"size:255;not null;uniqueIndex"`
	Description  string `json:"description" gorm:"type:text"`
	PhysicalPath string `json:"physical_path" gorm:"not null;uniqueIndex"`
	IsPrivate    bool   `json:"is_private" gorm:"not null;default:false"`
	IsActive     bool   `json:"is_active" gorm:"not null;default:true"`

	UserID string `json:"user_id" gorm:"not null"`
	User   User   `json:"user" gorm:"foreignKey:UserID;references:ID"`
}
