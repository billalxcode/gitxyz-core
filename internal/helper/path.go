package helper

import (
	"crypto/sha512"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

func GenerateRepositoryPath() string {
	hash := sha512.Sum512(fmt.Appendf(nil, "%d", time.Now().UnixNano()))
	hashStr := fmt.Sprintf("%x", hash)[:16]

	volumePath := viper.GetString("volume_path")
	return filepath.Join(volumePath, hashStr)
}
