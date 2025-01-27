package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	DatabasePath  string
	ServerName    string
	BaseURL       string
}

func Load() (*Config, error) {
	godotenv.Load()

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":3000"),
		DatabasePath:  getEnv("DATABASE_PATH", "data.db"),
		ServerName:    getEnv("SERVER_NAME", "My ActivityPub Server"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:3000"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
