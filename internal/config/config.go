package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	R2       R2Config
	Gemini   GeminiConfig
	OpenAI   OpenAIConfig
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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
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

type OpenAIConfig struct {
	APIKey              string
	Model               string
	MaxCompletionTokens int
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
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
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
		OpenAI: OpenAIConfig{
			APIKey:              getEnv("OPENAI_API_KEY", "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"),
			Model:               getEnv("OPENAI_MODEL", "gpt-5-nano-2025-08-07"),
			MaxCompletionTokens: getEnvAsInt("OPENAI_MAX_COMPLETION_TOKENS", 2048),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
