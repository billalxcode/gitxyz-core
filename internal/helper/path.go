package helper

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// RepositoryStoragePath returns the on-disk location of a repository relative
// to the project root, derived from its ID: "volume_path/<repoID>".
// The path is deterministic (no random hash) so it can be rebuilt from the
// repo ID alone — no physical_path column is stored.
func RepositoryStoragePath(repoID string) string {
	volumePath := viper.GetString("volume_path")
	return filepath.Join(volumePath, repoID)
}
