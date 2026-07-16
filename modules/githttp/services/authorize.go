package services

import (
	"errors"
	"gitxyz/internal/repository"
	"gitxyz/modules/githttp/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func authorizeRepo(ctx *gin.Context, db *gorm.DB, reponame string) {
	repoRepo := repository.NewRepoRepository(db)
	repo, err := repoRepo.FindByName(reponame)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "repository not found"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	_ = repo
}

func (s *GitServiceImpl) Authorize(ctx *gin.Context, options helper.Options) bool {
	authorizeRepo(ctx, s.db, options.RepoName)

	permission := NewPermission(s.db)

	// Use the authenticated username (set by AuthMiddleware), not the URL owner.
	username := ctx.GetString("username")

	switch options.ServiceType {
	case helper.ServiceTypeReceivePack:
		// Push requires write permission.
		if !permission.CanWrite(username, options.RepoName) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "write access denied"})
			return false
		}
	case helper.ServiceTypeUploadPack:
		// Clone/fetch requires read permission.
		if !permission.CanRead(username, options.RepoName) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "read access denied"})
			return false
		}
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid service type"})
		return false
	}

	return true
}
