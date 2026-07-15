package bootstrap

import (
	"gitxyz/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func runAutoMigrate(database *gorm.DB) {
	currentMode := gin.Mode()

	if currentMode != gin.ReleaseMode {
		database.AutoMigrate(&models.User{})
		database.AutoMigrate(&models.Repository{})
	}
}

func NewDatabase() *gorm.DB {
	DSN := viper.GetString("database_url")
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN: DSN,
	}))

	if err != nil {
		panic(err)
	}

	runAutoMigrate(database)

	return database
}
