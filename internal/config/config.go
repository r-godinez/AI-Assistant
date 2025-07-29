package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	OllamaURL   string
	Environment string
	DBPath      string
}

func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DBPath:      getEnv("DB_PATH", "./conversations.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
