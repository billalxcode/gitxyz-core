package models

type Repository struct {
	Base

	Hash         string `json:"hash" gorm:"size:255;not null"`
	RepoName     string `json:"reponame" gorm:"size:255;not null"`
	PhysicalPath string `json:"physical_path" gorm:"not null"`
	IsPrivate    bool   `json:"is_private" gorm:"not null;default:false"`

	UserID string `json:"user_id" gorm:"not null"`
	User   User   `json:"user" gorm:"foreignKey:UserID;references:ID"`
}
