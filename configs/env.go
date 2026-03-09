package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	Port              string
	DBConnection      string
	DBHost            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBPort            string
	DatabaseURL       string
	JWTSecret         string
	RedisHost         string
	RedisPort         string
	MaxLoginAttempts  int
	LoginBlockMinutes int
	SMTPEmail         string
	SMTPPassword      string
}

var AppEnv *Env

func LoadEnv() {
	env := os.Getenv("APP_ENV")
	switch env {
	case "production":
		godotenv.Load(".env.prod")
	case "development":
		godotenv.Load(".env.dev")
	default:
		godotenv.Load(".env")
	}

	maxAttempts, _ := strconv.Atoi(os.Getenv("MAX_LOGIN_ATTEMPTS"))
	if maxAttempts == 0 {
		maxAttempts = 3
	}

	blockMinutes, _ := strconv.Atoi(os.Getenv("LOGIN_BLOCK_MINUTES"))
	if blockMinutes == 0 {
		blockMinutes = 5
	}

	AppEnv = &Env{
		Port:              os.Getenv("PORT"),
		DBConnection:      os.Getenv("DB_CONNECTION"),
		DBHost:            os.Getenv("DB_HOST"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
		DBPort:            os.Getenv("DB_PORT"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		RedisHost:         os.Getenv("REDIS_HOST"),
		RedisPort:         os.Getenv("REDIS_PORT"),
		MaxLoginAttempts:  maxAttempts,
		LoginBlockMinutes: blockMinutes,
		SMTPEmail:         os.Getenv("SMTP_EMAIL"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
	}
}
