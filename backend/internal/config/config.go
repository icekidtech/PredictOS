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
	DreamDEXTestnetAPIURL string
	DreamDEXMainnetAPIURL string
	DreamDEXTestnetWSURL  string
	DreamDEXMainnetWSURL  string
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	FrontendURL           string
	Env                   string
}

func Load() *Config {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:                getEnv("PORT", ""),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", ""), // 32 bytes for AES-256
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:        getEnv("ANTHROPIC_API_KEY", ""),
		SomniaTestnetRPCURL:   getEnv("SOMNIA_TESTNET_RPC_URL", "https://dream-rpc.somnia.network"),
		SomniaMainnetRPCURL:   getEnv("SOMNIA_MAINNET_RPC_URL", "https://api.infra.mainnet.somnia.network"),
		DreamDEXTestnetAPIURL: getEnv("DREAMDEX_TESTNET_API_URL", "https://stg.api.dreamdex.io/v0"),
		DreamDEXMainnetAPIURL: getEnv("DREAMDEX_MAINNET_API_URL", "https://api.dreamdex.io/v0"),
		DreamDEXTestnetWSURL:  getEnv("DREAMDEX_TESTNET_WS_URL", "wss://stg.api.dreamdex.io/v0/ws/public"),
		DreamDEXMainnetWSURL:  getEnv("DREAMDEX_MAINNET_WS_URL", "wss://api.dreamdex.io/v0/ws/public"),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:     getEnv("GOOGLE_REDIRECT_URL", ""),
		FrontendURL:           getEnv("FRONTEND_URL", ""),
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

func (c *Config) DreamDEXAPIForNetwork(network string) string {
	if network == "mainnet" {
		return c.DreamDEXMainnetAPIURL
	}
	return c.DreamDEXTestnetAPIURL
}

func (c *Config) DreamDEXWSForNetwork(network string) string {
	if network == "mainnet" {
		return c.DreamDEXMainnetWSURL
	}
	return c.DreamDEXTestnetWSURL
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
