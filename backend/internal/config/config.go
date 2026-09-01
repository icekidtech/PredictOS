package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	JWTSecret             string
	EncryptionKey         string
	OpenAIAPIKey          string
	AnthropicKey          string
	SomniaTestnetRPCURL   string
	SomniaMainnetRPCURL   string
	Env                   string
}

func Load() *Config {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:                  getEnv("PORT", ""),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		EncryptionKey:         getEnv("ENCRYPTION_KEY", ""), // 32 bytes for AES-256
		OpenAIAPIKey:          getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:          getEnv("ANTHROPIC_API_KEY", ""),
		SomniaTestnetRPCURL:   getEnv("SOMNIA_TESTNET_RPC_URL", ""),
		SomniaMainnetRPCURL:   getEnv("SOMNIA_MAINNET_RPC_URL", ""),
		Env:                   getEnv("ENV", "development"),
	}
}

// HasAIProvider returns true if at least one AI provider key is configured.
func (c *Config) HasAIProvider() bool {
	return c.OpenAIAPIKey != "" || c.AnthropicKey != ""
}

// RPCForNetwork returns the RPC URL for the given network ("testnet" or "mainnet").
func (c *Config) RPCForNetwork(network string) string {
	switch network {
	case "mainnet":
		return c.SomniaMainnetRPCURL
	default:
		return c.SomniaTestnetRPCURL
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
