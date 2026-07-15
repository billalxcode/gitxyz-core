package services

import "gorm.io/gorm"

type Permission interface {
	CanRead(username string, reponame string) bool
	CanWrite(username string, reponame string) bool
}

type PermissionImpl struct {
	db *gorm.DB
}

func NewPermission(db *gorm.DB) Permission {
	return &PermissionImpl{
		db: db,
	}
}

func (a *PermissionImpl) CanRead(username string, reponame string) bool {
	// TODO: Implement function for authorization repository
	return true // placeholder
}

func (a *PermissionImpl) CanWrite(username string, reponame string) bool {
	// TODO: Implement function for authorization repository
	return true // placeholder
}
