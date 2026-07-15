package helper

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ServiceType string

type Options struct {
	ServiceType ServiceType
	RepoName    string
	UserName    string
}

const (
	ServiceTypeUploadPack    = "upload-pack"
	ServiceTypeReceivePack   = "receive-pack"
	ServiceTypeUploadArchive = "upload-archive"
)

func MakeOptionsFromContext(ctx *gin.Context) Options {
	// initialize options
	options := &Options{}

	// get service type
	service := ctx.Query("service")
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

	username := ctx.Param("username")
	options.UserName = username

	reponame := ctx.Param("reponame")
	options.RepoName = strings.TrimSuffix(reponame, ".git")

	return *options
}

func (o *Options) GetRepositoryStorage() string {
	wd, _ := os.Getwd()
	volumePath := viper.GetString("volume_path")

	return filepath.Join(
		wd,
		volumePath,
		o.UserName,
		o.RepoName,
	)
}

func (o *Options) EnsureRepositoryStorage() (string, error) {
	path := o.GetRepositoryStorage()
	err := os.MkdirAll(path, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		return "", err
	}
	return path, nil
}
