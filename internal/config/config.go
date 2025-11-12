package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	ServerHost string
	ServerPort string

	// AI Provider configuration
	DefaultAIProvider string // "openai" or "anthropic"
	OpenAIAPIKey      string
	OpenAIModel       string
	AnthropicAPIKey   string
	AnthropicModel    string

	// MCP Server configuration
	MCPServerPath         string
	CloudGenieBackendURL  string

	// CORS configuration
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		ServerHost:            getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:            getEnv("SERVER_PORT", "8081"),
		DefaultAIProvider:     getEnv("DEFAULT_AI_PROVIDER", "openai"),
		OpenAIAPIKey:          getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:           getEnv("OPENAI_MODEL", "gpt-4-turbo-preview"),
		AnthropicAPIKey:       getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:        getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
		MCPServerPath:         getEnv("MCP_SERVER_PATH", ""),
		CloudGenieBackendURL:  getEnv("CLOUDGENIE_BACKEND_URL", "http://localhost:8080"),
		AllowedOrigins:        []string{getEnv("ALLOWED_ORIGINS", "*")},
	}

	// Validate required configuration
	if cfg.DefaultAIProvider == "openai" && cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when using OpenAI provider")
	}

	if cfg.DefaultAIProvider == "anthropic" && cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when using Anthropic provider")
	}

	if cfg.MCPServerPath == "" {
		return nil, fmt.Errorf("MCP_SERVER_PATH is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
