package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRequireRoleAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "admin"); c.Next() })
	r.Use(RequireRole("admin", "owner"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("want 200 got %d", w.Code)
	}
}

func TestRequireRoleDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "user"); c.Next() })
	r.Use(RequireRole("admin", "owner"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403 got %d", w.Code)
	}
}

func newScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Policy{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestRequireScopePATAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("token_type", "pat")
		c.Set("token_scopes", "repo:write")
		c.Next()
	})
	r.Use(RequireScope(db, models.ScopeRepoWrite))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("want 200 got %d", w.Code)
	}
}

func TestRequireScopePATDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("token_type", "pat")
		c.Set("token_scopes", "repo:read")
		c.Next()
	})
	r.Use(RequireScope(db, models.ScopeRepoWrite))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403 got %d", w.Code)
	}
}

func TestRequireScopeJWTUserDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("token_type", "jwt")
		c.Set("role", "user")
		c.Next()
	})
	r.Use(RequireScope(db, models.ScopeRepoWrite))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403 got %d", w.Code)
	}
}

func TestRequireScopeJWTAdminAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("token_type", "jwt")
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(RequireScope(db, models.ScopeRepoWrite))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("want 200 got %d", w.Code)
	}
}

func TestCheckPolicyDenyWins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	polRepo := repository.NewPolicyRepository(db)
	if err := polRepo.Add(&models.Policy{
		SubjectType: "user", SubjectID: "u1", Action: "repo:read",
		ResourceType: "repository", ResourceID: "r1", Effect: "deny",
	}); err != nil {
		t.Fatalf("add policy: %v", err)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Set("user_id", "u1")
		c.Next()
	})
	r.Use(CheckPolicy(db, "repo:read", "repository", "r1"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("want 403 got %d", w.Code)
	}
}

func TestCheckPolicyUserAllow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newScopeTestDB(t)
	polRepo := repository.NewPolicyRepository(db)
	if err := polRepo.Add(&models.Policy{
		SubjectType: "user", SubjectID: "u1", Action: "repo:read",
		ResourceType: "repository", ResourceID: "r1", Effect: "allow",
	}); err != nil {
		t.Fatalf("add policy: %v", err)
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Set("user_id", "u1")
		c.Next()
	})
	r.Use(CheckPolicy(db, "repo:read", "repository", "r1"))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("want 200 got %d", w.Code)
	}
}
