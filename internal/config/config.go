package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	AppPort   int
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	appPort, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "avito"),
		DBPass:    getEnv("DB_PASSWORD", "avito"),
		DBName:    getEnv("DB_NAME", "avito"),
		AppPort:   appPort,
		JWTSecret: getEnv("JWT_SECRET", "super-secret-key"),
	}
	return cfg, nil
}

func (c *Config) Address() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
