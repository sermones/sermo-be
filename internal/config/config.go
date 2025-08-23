package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	R2       R2Config
	Gemini   GeminiConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type R2Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	Bucket          string
}

type GeminiConfig struct {
	APIKey     string
	ImageSize  string
	ImageStyle string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "3000"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "sermo"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		R2: R2Config{
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", "7d72013b97878d91c8d4b44b7515c9f9"),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", "0cef3b206ee353b43f988558ae07808a3ad46a12a22463c2b17bbbfde7908cdb"),
			Endpoint:        getEnv("R2_ENDPOINT", "https://035ea6d0a6159a3e4c6f41ded546d78d.r2.cloudflarestorage.com"),
			Bucket:          getEnv("R2_BUCKET", "sermo-be"),
		},
		Gemini: GeminiConfig{
			APIKey:     getEnv("GEMINI_API_KEY", "AIzaSyCtP88vApuqrnsfj9HGlG1NWRI6QxPJDfk"),
			ImageSize:  getEnv("GEMINI_IMAGE_SIZE", "1024x1024"),
			ImageStyle: getEnv("GEMINI_IMAGE_STYLE", "anime profile pricture style, 1:1 aspect ratio, high quality"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
