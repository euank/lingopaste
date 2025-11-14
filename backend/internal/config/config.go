package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// AWS
	AWSRegion               string
	AWSAccessKeyID          string
	AWSSecretAccessKey      string
	S3BucketName            string
	DynamoDBAccountsTable   string
	DynamoDBPastesTable     string
	DynamoDBRateLimitsTable string

	// OpenAI
	OpenAIAPIKey string
	OpenAIModel  string

	// Auth
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	AppleClientID      string
	AppleClientSecret  string
	FrontendURL        string

	// Stripe
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePriceID       string

	// Server
	Port           string
	CacheSize      int
	MaxPasteLength int
}

func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		AWSRegion:               getEnv("AWS_REGION", "ap-northeast-1"),
		AWSAccessKeyID:          getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:            getEnv("S3_BUCKET_NAME", "lingopaste-data"),
		DynamoDBAccountsTable:   getEnv("DYNAMODB_ACCOUNTS_TABLE", "lingopaste-accounts"),
		DynamoDBPastesTable:     getEnv("DYNAMODB_PASTES_TABLE", "lingopaste-pastes"),
		DynamoDBRateLimitsTable: getEnv("DYNAMODB_RATE_LIMITS_TABLE", "lingopaste-rate-limits"),
		OpenAIAPIKey:            getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:             getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		AppleClientID:           getEnv("APPLE_CLIENT_ID", ""),
		AppleClientSecret:       getEnv("APPLE_CLIENT_SECRET", ""),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:5173"),
		StripeSecretKey:         getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:     getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePriceID:           getEnv("STRIPE_PRICE_ID", ""),
		Port:                    getEnv("PORT", "8080"),
		CacheSize:               getEnvInt("CACHE_SIZE", 100000),
		MaxPasteLength:          getEnvInt("MAX_PASTE_LENGTH", 20000),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.S3BucketName == "" {
		return fmt.Errorf("S3_BUCKET_NAME is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
