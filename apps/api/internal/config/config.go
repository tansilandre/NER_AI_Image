package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string
	Env  string

	// Supabase
	SupabaseURL             string
	SupabaseAnonKey         string
	SupabaseServiceRoleKey  string
	DatabaseURL             string

	// Provider API Keys (can also be in DB)
	KieAIAPIKey     string
	OpenAIAPIKey    string
	GeminiAPIKey    string

	// R2 Storage
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string

	// App
	CallbackBaseURL            string
	ProviderKeyEncryptionSecret string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),

		SupabaseURL:                getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:            getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceRoleKey:     getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		DatabaseURL:                getEnv("DATABASE_URL", ""),

		KieAIAPIKey:  getEnv("KIE_AI_API_KEY", ""),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),

		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "ner-storage"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),

		CallbackBaseURL:             getEnv("CALLBACK_BASE_URL", "http://localhost:8080"),
		ProviderKeyEncryptionSecret: getEnv("PROVIDER_KEY_ENCRYPTION_SECRET", ""),
	}

	// Validate required config
	if cfg.SupabaseURL == "" {
		log.Println("Warning: SUPABASE_URL not set")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// IsDevelopment returns true if running in development
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
