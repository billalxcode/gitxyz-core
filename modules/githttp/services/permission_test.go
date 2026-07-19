package services

import (
	"testing"

	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPermDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&models.User{}, &models.Repository{}, &models.RepositoryMember{}, &models.Policy{})
	return db
}

func TestCanWritePublicLoggedIn(t *testing.T) {
	db := setupPermDB(t)
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "pub", IsPrivate: false, UserID: "owner1"})

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "anyone")
	ctx.Set("role", "user")
	ctx.Set("token_type", "jwt")
	ctx.Set("token_scopes", "")

	if !perm.CanWrite(ctx, "pub") {
		t.Error("logged-in user should be able to write public repo")
	}
}

func TestCanWritePrivateNonMember(t *testing.T) {
	db := setupPermDB(t)
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "priv", IsPrivate: true, UserID: "owner1"})

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "stranger")
	ctx.Set("role", "user")
	ctx.Set("token_type", "jwt")
	ctx.Set("token_scopes", "")

	if perm.CanWrite(ctx, "priv") {
		t.Error("non-member should NOT write private repo")
	}
}

func TestCanReadPrivateMemberReader(t *testing.T) {
	db := setupPermDB(t)
	userRepo := repository.NewUserRepository(db)
	_ = userRepo.Create(&models.User{Username: "reader1", Email: "reader1@example.com", Role: "user"})
	u, _ := userRepo.FindByUsername("reader1")
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "priv", IsPrivate: true, UserID: "owner1"})
	privRepo, _ := repoRepo.FindByName("priv")
	memRepo := repository.NewRepoMemberRepository(db)
	if err := memRepo.Add(&models.RepositoryMember{UserID: u.ID.String(), RepoID: privRepo.ID.String(), Role: "reader"}); err != nil {
		t.Fatalf("add member: %v", err)
	}

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "reader1")
	ctx.Set("role", "user")
	ctx.Set("token_type", "jwt")
	ctx.Set("token_scopes", "")

	if !perm.CanRead(ctx, "priv") {
		t.Error("reader member should read private repo")
	}
	if perm.CanWrite(ctx, "priv") {
		t.Error("reader member should NOT write private repo")
	}
}

func TestPATScopeDeny(t *testing.T) {
	db := setupPermDB(t)
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "priv", IsPrivate: true, UserID: "owner1"})

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "patuser")
	ctx.Set("role", "")
	ctx.Set("token_type", "pat")
	ctx.Set("token_scopes", "repo:read")

	if perm.CanWrite(ctx, "priv") {
		t.Error("PAT with repo:read only must NOT write")
	}
	if !perm.CanRead(ctx, "priv") {
		t.Error("PAT with repo:read should read")
	}
}

func TestABACPolicyDeniesMember(t *testing.T) {
	db := setupPermDB(t)
	userRepo := repository.NewUserRepository(db)
	_ = userRepo.Create(&models.User{Username: "reader1", Email: "reader1@example.com", Role: "user"})
	u, _ := userRepo.FindByUsername("reader1")
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "priv", IsPrivate: true, UserID: "owner1"})
	privRepo, _ := repoRepo.FindByName("priv")
	memRepo := repository.NewRepoMemberRepository(db)
	if err := memRepo.Add(&models.RepositoryMember{UserID: u.ID.String(), RepoID: privRepo.ID.String(), Role: "reader"}); err != nil {
		t.Fatalf("add member: %v", err)
	}
	// Explicit deny policy overrides membership.
	polRepo := repository.NewPolicyRepository(db)
	if err := polRepo.Add(&models.Policy{
		SubjectType: "user", SubjectID: u.ID.String(), Action: "repo:read",
		ResourceType: "repository", ResourceID: privRepo.ID.String(), Effect: "deny",
	}); err != nil {
		t.Fatalf("add policy: %v", err)
	}

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "reader1")
	ctx.Set("user_id", u.ID.String())
	ctx.Set("role", "user")
	ctx.Set("token_type", "jwt")
	ctx.Set("token_scopes", "")

	if perm.CanRead(ctx, "priv") {
		t.Error("explicit deny policy must override reader membership")
	}
}

func TestCreatorCanAccessPrivateRepo(t *testing.T) {
	db := setupPermDB(t)
	userRepo := repository.NewUserRepository(db)
	_ = userRepo.Create(&models.User{Username: "owner1", Email: "owner1@example.com", Role: "user"})
	repoRepo := repository.NewRepoRepository(db)
	_ = repoRepo.Create(&models.Repository{Name: "priv", IsPrivate: true, UserID: "owner1"})

	perm := NewPermission(db)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", "owner1")
	ctx.Set("role", "user")
	ctx.Set("token_type", "jwt")
	ctx.Set("token_scopes", "")

	if !perm.CanRead(ctx, "priv") {
		t.Error("repo creator should read own private repo")
	}
	if !perm.CanWrite(ctx, "priv") {
		t.Error("repo creator should write own private repo")
	}
}
