package services

import (
	"log/slog"

	"gitxyz/internal/api/services"
	"gitxyz/internal/logger"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Permission interface {
	CanRead(ctx *gin.Context, reponame string) bool
	CanWrite(ctx *gin.Context, reponame string) bool
}

type PermissionImpl struct {
	db *gorm.DB
}

func NewPermission(db *gorm.DB) Permission {
	return &PermissionImpl{db: db}
}

// CanRead reports whether the request context may fetch/clone reponame.
func (a *PermissionImpl) CanRead(ctx *gin.Context, reponame string) bool {
	log := logger.FromGin(ctx)
	repo, err := repository.NewRepoRepository(a.db).FindByName(reponame)
	if err != nil {
		log.Debug("permission: repo lookup failed", slog.String("repo", reponame), slog.String("error", err.Error()))
		return false
	}

	// Public repos: readable by anyone (incl. anonymous).
	if !repo.IsPrivate {
		log.Debug("permission: public repo read allowed", slog.String("repo", reponame))
		return true
	}

	// PAT scope check.
	if ctx.GetString("token_type") == "pat" {
		scopes := models.ParseScopes(ctx.GetString("token_scopes"))
		allowed := hasScope(scopes, models.ScopeRepoRead)
		log.Debug("permission: PAT read check", slog.String("repo", reponame), slog.Bool("allowed", allowed))
		return allowed
	}

	// System admin/owner bypass.
	if isSystemAdmin(ctx) {
		log.Debug("permission: admin/owner bypass read", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return true
	}

	// Explicit ABAC policy check (deny wins). PAT subject uses token scopes
	// already validated above; here we evaluate user/role policies.
	if denied, err := a.policyDenies(ctx, "repo:read", repo.ID.String()); err == nil && denied {
		log.Debug("permission: ABAC deny read", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return false
	}
	if ok, err := services.EvaluatePolicy(a.db, "user", ctx.GetString("user_id"), "repo:read", "repository", repo.ID.String()); err == nil && ok {
		log.Debug("permission: ABAC allow read", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return true
	}

	// Private repo: must be a member with at least reader role, or the creator.
	userID, ok := a.resolveUserID(ctx.GetString("username"))
	if !ok {
		log.Debug("permission: read denied (no user)", slog.String("repo", reponame))
		return false
	}
	if ctx.GetString("user_id") == repo.UserID {
		log.Debug("permission: creator read allowed", slog.String("repo", reponame))
		return true
	}
	allowed := a.memberCan(repo.ID.String(), userID, readRoles)
	log.Debug("permission: membership read check", slog.String("repo", reponame), slog.Bool("allowed", allowed))
	return allowed
}

// CanWrite reports whether the request context may push to reponame.
func (a *PermissionImpl) CanWrite(ctx *gin.Context, reponame string) bool {
	log := logger.FromGin(ctx)
	repo, err := repository.NewRepoRepository(a.db).FindByName(reponame)
	if err != nil {
		log.Debug("permission: repo lookup failed", slog.String("repo", reponame), slog.String("error", err.Error()))
		return false
	}

	// Public repos: writable by any logged-in user.
	if !repo.IsPrivate {
		allowed := ctx.GetString("username") != ""
		log.Debug("permission: public repo write check", slog.String("repo", reponame), slog.Bool("allowed", allowed))
		return allowed
	}

	// PAT scope check.
	if ctx.GetString("token_type") == "pat" {
		scopes := models.ParseScopes(ctx.GetString("token_scopes"))
		allowed := hasScope(scopes, models.ScopeRepoWrite)
		log.Debug("permission: PAT write check", slog.String("repo", reponame), slog.Bool("allowed", allowed))
		return allowed
	}

	// System admin/owner bypass.
	if isSystemAdmin(ctx) {
		log.Debug("permission: admin/owner bypass write", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return true
	}

	// Explicit ABAC policy check (deny wins).
	if denied, err := a.policyDenies(ctx, "repo:write", repo.ID.String()); err == nil && denied {
		log.Debug("permission: ABAC deny write", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return false
	}
	if ok, err := services.EvaluatePolicy(a.db, "user", ctx.GetString("user_id"), "repo:write", "repository", repo.ID.String()); err == nil && ok {
		log.Debug("permission: ABAC allow write", slog.String("repo", reponame), slog.String("username", ctx.GetString("username")))
		return true
	}

	// Private repo: member with write-capable role, or the creator.
	userID, ok := a.resolveUserID(ctx.GetString("username"))
	if !ok {
		log.Debug("permission: write denied (no user)", slog.String("repo", reponame))
		return false
	}
	if ctx.GetString("user_id") == repo.UserID {
		log.Debug("permission: creator write allowed", slog.String("repo", reponame))
		return true
	}
	allowed := a.memberCan(repo.ID.String(), userID, writeRoles)
	log.Debug("permission: membership write check", slog.String("repo", reponame), slog.Bool("allowed", allowed))
	return allowed
}

var (
	readRoles  = []string{"owner", "maintainer", "triager", "reader", "guest"}
	writeRoles = []string{"owner", "maintainer", "triager"}
)

func (a *PermissionImpl) resolveUserID(username string) (string, bool) {
	if username == "" {
		return "", false
	}
	user, err := repository.NewUserRepository(a.db).FindByUsername(username)
	if err != nil {
		return "", false
	}
	return user.ID.String(), true
}

// policyDenies reports whether an explicit deny policy blocks the action for
// the current user. A deny always wins over membership.
func (a *PermissionImpl) policyDenies(ctx *gin.Context, action, repoID string) (bool, error) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		return false, nil
	}
	polRepo := repository.NewPolicyRepository(a.db)
	ps, err := polRepo.FindApplicable("user", userID, action, "repository", repoID)
	if err != nil {
		return false, err
	}
	for _, p := range ps {
		if p.Effect == "deny" {
			return true, nil
		}
	}
	return false, nil
}

func (a *PermissionImpl) memberCan(repoID, userID string, allowed []string) bool {
	if userID == "" {
		return false
	}
	mem, err := repository.NewRepoMemberRepository(a.db).FindByUserAndRepo(userID, repoID)
	if err != nil {
		return false
	}
	for _, r := range allowed {
		if mem.Role == r {
			return true
		}
	}
	return false
}

func isSystemAdmin(ctx *gin.Context) bool {
	role := ctx.GetString("role")
	return role == models.RoleAdmin || role == models.RoleOwner
}

func hasScope(scopes []string, want string) bool {
	for _, s := range scopes {
		if s == want || s == models.ScopeAdmin {
			return true
		}
	}
	return false
}
