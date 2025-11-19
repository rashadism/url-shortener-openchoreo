package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	Port              string
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://urlshortener:password123@localhost:5432/urlshortener?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		Port:              getEnv("PORT", "7543"),
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 5),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
