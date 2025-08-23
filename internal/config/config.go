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
		Firebase: FirebaseConfig{
			ProjectID:           getEnv("FIREBASE_PROJECT_ID", "sermo-a64b0"),
			PrivateKeyID:        getEnv("FIREBASE_PRIVATE_KEY_ID", "745c28c113c54c165f63f00fed23a9f23f978fd8"),
			PrivateKey:          getEnv("FIREBASE_PRIVATE_KEY", "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC/tGUUJyJUJ9oB\nxOe7Wg7bf6n9LH8rZT7bAUQ4XUUSm5dozHFNwbaTNHEC6FpNXVBJlFayo4z1jd9n\npACu/vEmCRkEVqoLfNbtaQuGrQA3/wNLJRwyHDun5IEd37rBYPqdaHoWS4GxUD5Y\nnBfhRegMtIevXcPjZzu9HhAE3kSZA/aP09TL1OAu8u7A0edfpbp9aUuZBy8+Uja/\nmRIJI/FJfw/EIntZvIDMnDaQSZs1fReaqb2zm7NNfg1K6tDDI9QAJc6lCNxqIKs2\n3d/4q0s7CETzNUmWc2fpTVSoaW+ULUxfGOlfKHVWgh9Dr8C6lFnI1yGb/8trBTZi\nLBpnAy5/AgMBAAECggEABtTnaz3vx7750mlRdshBUkos5SRSTdoYUNatNzL/tKeN\n0ELE4Of/2SQbyOY93ZbvNwgNxOT2L4QES525FoNoVZoqZZCvE6S5xhEhw4VjxhsU\n43cpl5GRAK0wer+P2LdbhwODokrDju2UozTA1xrWJ6nBUhsN0YtKwjURjhLbr/Y3\nwRj9GflWP1d72bQQP1vv+GlQq/r1C6RsRE7gNUL6PaiwT99PQU8t0/QEyvMI5v62\n13YMnBpMks+AnJvlZKjAalTYqKT8N8FcF+qMbWvSDmq7/e+AD2Hhx7a04DMp2Lnx\n2CL1Kept0n5NecMhyHfysozL97sOboFr6S2GDmIdfQKBgQD+SOjB6OFr6K/peu7G\nqQNms3ow1zQ3iPJpIBbsCmnJr+dNLTlf9pC9geDPq/O+q9xtPNJ/Vztwj970xwq/\nQbcie1m2AgA6sgzKYFuyIf+g14n/NzH+PHms8x2w4peq+mXOz9SyCX+Tw8OjUP40\nFbma5zakE+49FlDAgLaUhuuREwKBgQDA/2ydKrL6T5u7GYDrnVkf9Ws7R5eOB/iA\n3ncTZpCOdNNWczfnhxOGAQjQ8g5M/yyakCo2ZbP8P3L8ORBaOjcjudIuSUzTZuau\nAJWK9SukAU2rU5uZ9N1En94iDnz6RKIa2Dw8gBHPHK1/ZMnS1voTnK5owcbkXrIc\n61TVMLGGZQKBgCeebn1/7ldkyrvDBp73SGtg/WHMtfsNIE/Wyxt9x9u/x3cT28Oi\n5AxSxxc0QGbt2gs/FcD3c3BnSiKzPG5uK714oJKmHykGAs4pU0Ae4fhKfNrB280z\n2PVkb7TWqTDfkKs3YHlY14LLVpkEjobI98E10yKfZFgqOOy1YT0lBGD3AoGAEtpg\nv9Gl/jG8osBRCbMrO6X7vaS2t1cr/Vq+AxUn1eKvqmhC88kMLTD4rYCXyQm8T7T5\niqrQtDl2gBEK+eVp8YF7eK4MZTJOIn1IHnTouHKwJaZbMuTqoIOFbYpAxynhNAIf\nkEFqe/LvN9yeoowBjdzmZLFZPoHJoG2Usea50MkCgYAGOT4TVG7sYTdOGfTIUi3N\nWg+2V8/vgx84BqltZTZ0/tbaERjKJlEafokGstcCExgTsjiCfqtlUXtY2K2hPMFo\n9WrCeP/my39/Viut18zTLkfhiv8FsPF5xWCLm64pAmbmFv7N4S/kB38HUEKYciXA\nX4olDQIpCN8ps4v7mVq73Q==\n-----END PRIVATE KEY-----\n"),
			ClientEmail:         getEnv("FIREBASE_CLIENT_EMAIL", "firebase-adminsdk-fbsvc@sermo-a64b0.iam.gserviceaccount.com"),
			ClientID:            getEnv("FIREBASE_CLIENT_ID", "109682621878570509138"),
			AuthURI:             getEnv("FIREBASE_AUTH_URI", "https://accounts.google.com/o/oauth2/auth"),
			TokenURI:            getEnv("FIREBASE_TOKEN_URI", "https://oauth2.googleapis.com/token"),
			AuthProviderCertURL: getEnv("FIREBASE_AUTH_PROVIDER_CERT_URL", "https://www.googleapis.com/oauth2/v1/certs"),
			ClientCertURL:       getEnv("FIREBASE_CLIENT_CERT_URL", "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-fbsvc%40sermo-a64b0.iam.gserviceaccount.com"),
			UniverseDomain:      getEnv("FIREBASE_UNIVERSE_DOMAIN", "googleapis.com"),
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
