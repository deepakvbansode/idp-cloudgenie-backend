package ai

import (
	"context"
	"fmt"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
)

type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(apiKey, model string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default to Claude 3.5 Sonnet
	}

	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}, nil
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) Chat(ctx context.Context, prompt string, tools []*mcp.Tool, conversationHistory []Message) (*Response, error) {
	// TODO: Implement Anthropic API integration
	// For now, return a basic response
	return &Response{
		Content:      "Anthropic integration coming soon. Please use OpenAI provider for now.",
		FinishReason: "stop",
	}, nil
}
