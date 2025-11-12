package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if model == "" {
		model = openai.GPT4TurboPreview // Default to GPT-4 Turbo
	}

	client := openai.NewClient(apiKey)

	return &OpenAIProvider{
		client: client,
		model:  model,
	}, nil
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

func (p *OpenAIProvider) Chat(ctx context.Context, prompt string, tools []mcp.Tool, conversationHistory []Message) (*Response, error) {
	// Build messages from conversation history
	messages := []openai.ChatCompletionMessage{}

	// Add system message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful AI assistant that can interact with CloudGenie infrastructure management platform. You have access to various tools to help manage cloud resources. When asked to perform operations, use the available tools to accomplish the task.",
	})

	// Add conversation history
	for _, msg := range conversationHistory {
		role := msg.Role
		if role == "user" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})
		} else if role == "assistant" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Content,
			})
		}
	}

	// Add current prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// Convert MCP tools to OpenAI function format
	var openaiTools []openai.Tool
	if len(tools) > 0 {
		for _, tool := range tools {
			openaiTools = append(openaiTools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
	}

	// Create completion request
	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
	}

	if len(openaiTools) > 0 {
		req.Tools = openaiTools
		req.ToolChoice = "auto"
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	response := &Response{
		Content:      choice.Message.Content,
		FinishReason: string(choice.FinishReason),
		Usage: &Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Handle tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		for _, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}
	}

	return response, nil
}
