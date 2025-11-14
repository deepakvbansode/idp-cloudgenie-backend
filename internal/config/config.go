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
	DefaultAIProvider string // "openai", "anthropic", "gemini", or "glean"
	OpenAIAPIKey      string
	OpenAIModel       string
	AnthropicAPIKey   string
	AnthropicModel    string
	GeminiAPIKey      string
	GeminiModel       string
	GleanAPIKey       string
	GleanInstance     string // Company instance name (e.g., "your-company")
	GleanModel        string

	// MCP Server configuration
	MCPServerURL          string
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
		GeminiAPIKey:          getEnv("GEMINI_API_KEY", ""),
		GeminiModel:           getEnv("GEMINI_MODEL", "gemini-1.5-pro"),
		GleanAPIKey:           getEnv("GLEAN_API_KEY", ""),
		GleanInstance:         getEnv("GLEAN_INSTANCE", ""),
		GleanModel:            getEnv("GLEAN_MODEL", "glean-default"),
		MCPServerURL:          getEnv("MCP_SERVER_URL", "http://localhost:3000"),
		CloudGenieBackendURL:  getEnv("CLOUDGENIE_BACKEND_URL", "http://localhost:8080"),
		AllowedOrigins:        []string{getEnv("ALLOWED_ORIGINS", "*")},
	}

	

	// Validate required fields based on AI provider
	if cfg.DefaultAIProvider == "openai" && cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when using openai provider")
	}
	if cfg.DefaultAIProvider == "anthropic" && cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when using anthropic provider")
	}
	if cfg.DefaultAIProvider == "gemini" && cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when using gemini provider")
	}
	if cfg.DefaultAIProvider == "glean" && cfg.GleanAPIKey == "" {
		return nil, fmt.Errorf("GLEAN_API_KEY is required when using glean provider")
	}
	if cfg.MCPServerURL == "" {
		return nil, fmt.Errorf("MCP_SERVER_URL is required")
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
