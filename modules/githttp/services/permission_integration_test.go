package services

import (
	"testing"

	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPermissionMatrix(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.Repository{}, &models.RepositoryMember{})
	ur := repository.NewUserRepository(db)
	_ = ur.Create(&models.User{Username: "m1", Email: "m1@example.com", Role: "user"})
	mu, _ := ur.FindByUsername("m1")
	rr := repository.NewRepoRepository(db)
	rr.Create(&models.Repository{Name: "pubrepo", IsPrivate: false, UserID: "o1"})
	rr.Create(&models.Repository{Name: "privrepo", IsPrivate: true, UserID: "o1"})
	privRepo, _ := rr.FindByName("privrepo")
	mr := repository.NewRepoMemberRepository(db)
	if err := mr.Add(&models.RepositoryMember{UserID: mu.ID.String(), RepoID: privRepo.ID.String(), Role: "maintainer"}); err != nil {
		t.Fatalf("add member: %v", err)
	}

	perm := NewPermission(db)

	cases := []struct {
		name  string
		ctx   func() *gin.Context
		repo  string
		wantR bool
		wantW bool
	}{
		{"anon public", ctxAnon, "pubrepo", true, false},
		{"login public", ctxLogin("user"), "pubrepo", true, true},
		{"anon private", ctxAnon, "privrepo", false, false},
		{"member maintainer private", ctxLogin("m1"), "privrepo", true, true},
		{"stranger private", ctxLogin("stranger"), "privrepo", false, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := c.ctx()
			if got := perm.CanRead(ctx, c.repo); got != c.wantR {
				t.Errorf("CanRead: want %v got %v", c.wantR, got)
			}
			if got := perm.CanWrite(ctx, c.repo); got != c.wantW {
				t.Errorf("CanWrite: want %v got %v", c.wantW, got)
			}
		})
	}
}

func ctxAnon() *gin.Context {
	return withCtx("", "", "jwt", "")
}

func ctxLogin(user string) func() *gin.Context {
	return func() *gin.Context { return withCtx(user, "user", "jwt", "") }
}

func withCtx(user, role, tokType, scopes string) *gin.Context {
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("username", user)
	ctx.Set("role", role)
	ctx.Set("token_type", tokType)
	ctx.Set("token_scopes", scopes)
	return ctx
}
