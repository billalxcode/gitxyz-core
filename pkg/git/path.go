package git

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func MakeRepositoryDirectory(path string) (err error) {
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func JoinRepositoryPath(reponame string) string {
	dataPath := viper.GetString("volume_path")

	return filepath.Join(
		dataPath,
		reponame,
	)
}
