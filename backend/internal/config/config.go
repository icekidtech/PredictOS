package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	EncryptionKey string
	OpenAIAPIKey  string
	AnthropicKey  string
	SomniaRPCURL  string
	Env           string
}

func Load() *Config {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:          getEnv("PORT", ""),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		EncryptionKey: getEnv("ENCRYPTION_KEY", ""), // 32 bytes for AES-256
		OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:  getEnv("ANTHROPIC_API_KEY", ""),
		SomniaRPCURL:  getEnv("SOMNIA_RPC_URL", ""),
		Env:           getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
