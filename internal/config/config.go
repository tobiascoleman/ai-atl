package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI     string
	DBName       string
	JWTSecret    string
	GeminiAPIKey string
	RedisURL     string
	Environment  string
	Port         string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		MongoURI:     getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:       getEnv("DB_NAME", "nfl_platform"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		Port:         getEnv("PORT", "8080"),
	}

	// Validate critical config
	if cfg.GeminiAPIKey == "" {
		log.Println("WARNING: GEMINI_API_KEY not set - AI features will not work")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
