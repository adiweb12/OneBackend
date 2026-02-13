package config

import "os"

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	GeminiAPIKey   string
	CloudinaryURL  string
	ServerPort     string
	RefreshSecret  string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost:5432/onechat?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		RefreshSecret: getEnv("REFRESH_SECRET", "your-refresh-secret-change-in-production"),
		GeminiAPIKey:  getEnv("GEMINI_API_KEY", ""),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
		ServerPort:    getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
