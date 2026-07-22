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

func TestEvaluatePolicyDenyWins(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Policy{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	polRepo := repository.NewPolicyRepository(db)
	if err := polRepo.Add(&models.Policy{
		SubjectType: "user", SubjectID: "u1", Action: "repo:read",
		ResourceType: "repository", ResourceID: "r1", Effect: "deny",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	ok, err := EvaluatePolicy(db, "user", "u1", "repo:read", "repository", "r1")
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if ok {
		t.Error("expected deny to win")
	}
}

func TestEvaluatePolicyAdminBypass(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Policy{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ok, err := EvaluatePolicy(db, "role", "admin", "repo:write", "repository", "r1")
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !ok {
		t.Error("expected admin role to bypass")
	}
}

func TestEvaluatePolicyUserAdminBypass(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Policy{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(&models.User{Username: "admin1", Email: "a@e.com", Role: "admin"}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	u, _ := userRepo.FindByUsername("admin1")
	ok, err := EvaluatePolicy(db, "user", u.ID.String(), "repo:write", "repository", "r1")
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !ok {
		t.Error("expected admin user to bypass")
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
