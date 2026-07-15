package main

import (
	"gitxyz/internal/bootstrap"
	"os"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("gitxyz")
	viper.AddConfigPath("config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	if !viper.IsSet("jwt_secret") {
		viper.Set("jwt_secret", "gitxyz-dev-secret-change-me")
	}
	if !viper.IsSet("jwt_expiry_hours") {
		viper.Set("jwt_expiry_hours", 24)
	}

	viper.SafeWriteConfig()

	volumePath := viper.GetString("volume_path")
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		panic(err)
	}

}

func main() {
	bootstrap.Initialize()
}
