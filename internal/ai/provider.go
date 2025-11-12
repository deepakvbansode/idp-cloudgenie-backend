package ai

import (
	"context"
	"fmt"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
)

// Provider defines the interface for AI providers
type Provider interface {
	Chat(ctx context.Context, prompt string, tools []mcp.Tool, conversationHistory []Message) (*Response, error)
	GetProviderName() string
}

// Message represents a conversation message
type Message struct {
	Role    string                 `json:"role"`    // "user", "assistant", "system"
	Content string                 `json:"content"`
	ToolCalls []ToolCall           `json:"tool_calls,omitempty"`
	ToolResults []ToolResult       `json:"tool_results,omitempty"`
}

// ToolCall represents a tool call request from the AI
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// Response represents the AI response
type Response struct {
	Content     string       `json:"content"`
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	FinishReason string      `json:"finish_reason"`
	Usage       *Usage       `json:"usage,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewProvider creates a new AI provider based on the provider name
func NewProvider(providerName, apiKey, model string) (Provider, error) {
	switch providerName {
	case "openai", "":
		return NewOpenAIProvider(apiKey, model)
	case "anthropic":
		return NewAnthropicProvider(apiKey, model)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
