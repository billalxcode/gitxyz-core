package controllers

import (
	"gitxyz/internal/api/services"

	"gorm.io/gorm"
)

type RepoController interface{}
type RepoControllerImpl struct {
	service services.RepoService
	db      *gorm.DB
}

func NewRepoController(db *gorm.DB) RepoController {
	service := services.NewRepoService()

	return &RepoControllerImpl{
		service: service,
		db:      db,
	}
}
