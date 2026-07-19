package repository

import (
	"testing"

	"gitxyz/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{}, &models.Repository{},
		&models.RepositoryMember{}, &models.Policy{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestRepoMemberAddAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := NewRepoMemberRepository(db)

	m := models.RepositoryMember{UserID: "u1", RepoID: "r1", Role: "maintainer"}
	if err := repo.Add(&m); err != nil {
		t.Fatalf("add: %v", err)
	}
	got, err := repo.FindByUserAndRepo("u1", "r1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Role != "maintainer" {
		t.Errorf("want maintainer got %q", got.Role)
	}
}

func TestPolicyRepositoryAddFindRemove(t *testing.T) {
	db := newTestDB(t)
	repo := NewPolicyRepository(db)

	p := models.Policy{
		SubjectType: "user", SubjectID: "u1", Action: "repo:read",
		ResourceType: "repository", ResourceID: "r1", Effect: "allow",
	}
	if err := repo.Add(&p); err != nil {
		t.Fatalf("add: %v", err)
	}
	if p.ID.String() == "" {
		t.Fatal("expected generated ID")
	}

	got, err := repo.FindApplicable("user", "u1", "repo:read", "repository", "r1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(got) != 1 || got[0].Effect != "allow" {
		t.Fatalf("want 1 allow policy, got %+v", got)
	}

	// Wildcard resource match: a policy with resource_id='*' applies to any repo.
	if err := repo.Add(&models.Policy{
		SubjectType: "user", SubjectID: "u1", Action: "repo:read",
		ResourceType: "repository", ResourceID: "*", Effect: "allow",
	}); err != nil {
		t.Fatalf("add wildcard: %v", err)
	}
	wild, err := repo.FindApplicable("user", "u1", "repo:read", "repository", "any-repo")
	if err != nil {
		t.Fatalf("find wildcard: %v", err)
	}
	if len(wild) != 1 {
		t.Fatalf("want 1 wildcard match, got %d", len(wild))
	}

	if err := repo.Remove(p.ID.String()); err != nil {
		t.Fatalf("remove: %v", err)
	}
	after, err := repo.FindApplicable("user", "u1", "repo:read", "repository", "r1")
	if err != nil {
		t.Fatalf("find after remove: %v", err)
	}
	// The specific policy is gone; the wildcard ('*') policy remains.
	if len(after) != 1 || after[0].ResourceID != "*" {
		t.Fatalf("want only wildcard remaining, got %+v", after)
	}
}
