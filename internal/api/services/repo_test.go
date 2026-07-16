package services

import (
	"gitxyz/internal/models"
	"gitxyz/internal/repository"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateRepositoryValidation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.Repository{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	service := &RepoServiceImpl{Repository: repository.NewRepoRepository(db)}

	repo := &models.Repository{Name: "", UserID: "user-1"}
	if err := service.CreateRepository(repo); err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestCreateRepositorySuccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.Repository{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	service := &RepoServiceImpl{Repository: repository.NewRepoRepository(db)}

	repo := &models.Repository{Name: "test-repo", UserID: "user-1"}
	if err := service.CreateRepository(repo); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.ID == [16]byte{} {
		t.Fatal("repo ID should be assigned by the model hook")
	}
}
