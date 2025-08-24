package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	R2       R2Config
	Gemini   GeminiConfig
	OpenAI   OpenAIConfig
	Firebase FirebaseConfig
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

type OpenAIConfig struct {
	APIKey              string
	Model               string
	MaxCompletionTokens int
}

type FirebaseConfig struct {
	ProjectID           string
	PrivateKeyID        string
	PrivateKey          string
	ClientEmail         string
	ClientID            string
	AuthURI             string
	TokenURI            string
	AuthProviderCertURL string
	ClientCertURL       string
	UniverseDomain      string
}

func Load() *Config {
	return &Config{}
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
