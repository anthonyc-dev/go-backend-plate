package database

import (
	"fmt"
	"log"
	"rest-api/configs"
	"rest-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var dsn string

	if configs.AppEnv.DatabaseURL != "" {
		dsn = configs.AppEnv.DatabaseURL
	} else {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s Timezone=UTC",
			getEnv(configs.AppEnv.DBHost, "localhost"),
			getEnv(configs.AppEnv.DBUser, "postgres"),
			getEnv(configs.AppEnv.DBPassword, "postgres"),
			getEnv(configs.AppEnv.DBName, "go-db"),
			getEnv(configs.AppEnv.DBPort, "5432"),
			"disable",
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.TokenBlacklist{})

	DB = db
}

func getEnv(key, fallback string) string {
	if key != "" {
		return key
	}
	return fallback
}
