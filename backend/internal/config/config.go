package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort           string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	AuthSecret        string
	AuthTokenTTLHours int
}

func Load() Config {
	return Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "rsl_user"),
		DBPassword:        getEnv("DB_PASSWORD", "rsl_password"),
		DBName:            getEnv("DB_NAME", "rsl_learning"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		AuthSecret:        getEnv("AUTH_SECRET", "rsl-learning-generator-demo-secret"),
		AuthTokenTTLHours: getEnvInt("AUTH_TOKEN_TTL_HOURS", 12),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
