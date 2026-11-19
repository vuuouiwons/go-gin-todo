package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config encapsulates runtime application configurations.
type Config struct {
	Port              string
	JWTSecret         string
	JWTExpiresInHours int
	DBPath            string
	GinMode           string
}

// LoadConfig reads configuration from .env and environment variables.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from system environment variables")
	}

	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-please-change-in-production")
	dbPath := getEnv("DB_PATH", "todo.db")
	ginMode := getEnv("GIN_MODE", "debug")

	expiresHoursStr := getEnv("JWT_EXPIRES_IN_HOURS", "24")
	expiresHours, err := strconv.Atoi(expiresHoursStr)
	if err != nil {
		expiresHours = 24
	}

	return &Config{
		Port:              port,
		JWTSecret:         jwtSecret,
		JWTExpiresInHours: expiresHours,
		DBPath:            dbPath,
		GinMode:           ginMode,
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}
