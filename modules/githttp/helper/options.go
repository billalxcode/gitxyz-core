package helper

import (
	"os"
	"path/filepath"
	"strings"

	"gitxyz/internal/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServiceType string

type Options struct {
	ServiceType ServiceType
	RepoName    string
	UserName    string

	db *gorm.DB
}

const (
	ServiceTypeUploadPack    = "upload-pack"
	ServiceTypeReceivePack   = "receive-pack"
	ServiceTypeUploadArchive = "upload-archive"
)

func MakeOptionsFromContext(ctx *gin.Context, db *gorm.DB) Options {
	// initialize options
	options := &Options{
		db: db,
	}

	username := ctx.Param("username")
	options.UserName = username

	reponame := ctx.Param("reponame")
	options.RepoName = strings.TrimSuffix(reponame, ".git")

	// get service type
	service := ctx.Query("service")
	if service == "" {
		// On the actual RPC endpoints (POST /git-receive-pack,
		// POST /git-upload-pack) the service is in the path, not the query.
		path := ctx.Request.URL.Path
		service = strings.TrimPrefix(path, "/"+username+"/"+reponame)
		service = strings.TrimPrefix(service, "/")
		service = strings.TrimSuffix(service, "/")
	}
	switch service {
	case "git-receive-pack":
		options.ServiceType = ServiceTypeReceivePack

	case "git-upload-pack":
		options.ServiceType = ServiceTypeUploadPack

	case "git-upload-archive":
		options.ServiceType = ServiceTypeUploadArchive

	default:
		options.ServiceType = ""
	}

	return *options
}

// GetRepositoryStorage resolves the absolute on-disk location of a repository
// from its ID: "wd/volume_path/<repoID>". The path is deterministic and derived
// from the repo ID alone — no physical_path column is stored.
func (o *Options) GetRepositoryStorage(repoID string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, helper.RepositoryStoragePath(repoID))
}

func (o *Options) EnsureRepositoryStorage(repoID string) (string, error) {
	path := o.GetRepositoryStorage(repoID)
	err := os.MkdirAll(path, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		return "", err
	}
	return path, nil
}
