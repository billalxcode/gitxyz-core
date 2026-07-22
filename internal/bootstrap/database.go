package bootstrap

import (
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase() *gorm.DB {
	DSN := viper.GetString("database_url")
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN: DSN,
	}))

	if err != nil {
		panic(err)
	}

	if err := runMigrations(DSN); err != nil {
		panic(err)
	}

	return database
}
