package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string

	ConsistencyCronEnabled  bool
	ConsistencyCronInterval time.Duration
	ConsistencyCronTimeout  time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "banking"),
		JWTSecret:  getEnv("JWT_SECRET", "bank"),
		Port:       getEnv("PORT", "8080"),

		ConsistencyCronEnabled:  getEnvBool("CONSISTENCY_CRON_ENABLED", false),
		ConsistencyCronInterval: getEnvDurationSeconds("CONSISTENCY_CRON_INTERVAL_SECONDS", 300),
		ConsistencyCronTimeout:  getEnvDurationSeconds("CONSISTENCY_CRON_TIMEOUT_SECONDS", 30),
	}

	if config.JWTSecret == "bank" {
		fmt.Println("WARNING: Using default JWT_SECRET")
	}

	return config, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvDurationSeconds(key string, defaultSeconds int) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return time.Duration(defaultSeconds) * time.Second
	}
	sec, err := strconv.Atoi(raw)
	if err != nil || sec <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(sec) * time.Second
}
