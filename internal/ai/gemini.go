package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(apiKey, model string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Google Gemini API key is required")
	}

	if model == "" {
		model = "gemini-1.5-pro" // Default to Gemini 1.5 Pro
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

func (p *GeminiProvider) GetProviderName() string {
	return "gemini"
}

func (p *GeminiProvider) Chat(ctx context.Context, prompt string, tools []*mcp.Tool, conversationHistory []Message) (*Response, error) {
	model := p.client.GenerativeModel(p.model)

	// Configure model
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetTopK(40)

	// Build the system instruction with tools information
	systemPrompt := buildSystemPromptWithTools(tools)
	
	// Build the complete prompt with context
	fullPrompt := systemPrompt + "\n\n"
	
	// Add conversation history
	for _, msg := range conversationHistory {
		if msg.Role == "user" {
			fullPrompt += fmt.Sprintf("User: %s\n", msg.Content)
		} else if msg.Role == "assistant" {
			fullPrompt += fmt.Sprintf("Assistant: %s\n", msg.Content)
		}
	}
	
	// Add current prompt with instructions for tool usage
	fullPrompt += fmt.Sprintf("\nUser: %s\n\n", prompt)
	fullPrompt += "Assistant: Let me help you with that. "
	
	// If tools are available, add instruction to use them
	if len(tools) > 0 {
		fullPrompt += "I'll use the available tools to accomplish this task. "
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	candidate := resp.Candidates[0]
	
	// Extract response content
	var responseContent string
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseContent += string(text)
		}
	}

	response := &Response{
		Content:      responseContent,
		FinishReason: fmt.Sprintf("%v", candidate.FinishReason),
		Usage: &Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	// Parse tool calls from the response
	// Look for tool call patterns in the format: TOOL_CALL: tool_name({"arg": "value"})
	toolCalls := extractToolCalls(responseContent, tools)
	if len(toolCalls) > 0 {
		response.ToolCalls = toolCalls
	}

	return response, nil
}

// buildSystemPromptWithTools creates a system prompt that includes tool information
func buildSystemPromptWithTools(tools []*mcp.Tool) string {
	prompt := `You are a helpful AI assistant that can interact with CloudGenie infrastructure management platform.

You have access to the following tools to help manage cloud resources. When you need to perform an action, you should call the appropriate tool by responding in this EXACT format:

TOOL_CALL: tool_name({"arg1": "value1", "arg2": "value2"})

For example:
TOOL_CALL: cloudgenie_get_blueprints({})
TOOL_CALL: cloudgenie_create_resource({"name": "my-db", "type": "database", "blueprint_id": "postgres-123"})

Available tools:
`

	for _, tool := range tools {
		prompt += fmt.Sprintf("\n%s: %s\n", tool.Name, tool.Description)
		
		// Add parameter information
		if tool.InputSchema != nil {
			// Type assert InputSchema to map[string]interface{}
			if schema, ok := tool.InputSchema.(map[string]interface{}); ok {
				if props, ok := schema["properties"].(map[string]interface{}); ok {
					prompt += "  Parameters:\n"
					for paramName, paramInfo := range props {
						if paramMap, ok := paramInfo.(map[string]interface{}); ok {
							paramType := "string"
							if t, ok := paramMap["type"].(string); ok {
								paramType = t
							}
							paramDesc := ""
							if d, ok := paramMap["description"].(string); ok {
								paramDesc = d
							}
							prompt += fmt.Sprintf("    - %s (%s): %s\n", paramName, paramType, paramDesc)
						}
					}
				}
				
				// Add required fields
				if required, ok := schema["required"].([]interface{}); ok && len(required) > 0 {
					reqFields := []string{}
					for _, r := range required {
						if rs, ok := r.(string); ok {
							reqFields = append(reqFields, rs)
						}
					}
					if len(reqFields) > 0 {
						prompt += fmt.Sprintf("  Required: %s\n", strings.Join(reqFields, ", "))
					}
				}
			}
		}
	}

	prompt += `
IMPORTANT RULES:
1. When you need to use a tool, output EXACTLY in the format: TOOL_CALL: tool_name({json_args})
2. You can call multiple tools by outputting multiple TOOL_CALL lines
3. After calling tools, explain what you're doing
4. Use proper JSON format for arguments
5. Don't make up tool names - only use the tools listed above

When the user asks you to do something:
1. First, determine if you need to use any tools
2. If yes, output the TOOL_CALL lines
3. Then provide a natural language explanation

Example response:
TOOL_CALL: cloudgenie_get_blueprints({})
I'm fetching all available infrastructure blueprints for you.
`

	return prompt
}

// extractToolCalls parses the response to find tool call requests
func extractToolCalls(content string, tools []*mcp.Tool) []ToolCall {
	var toolCalls []ToolCall
	
	// Create a map of valid tool names for quick lookup
	validTools := make(map[string]bool)
	for _, tool := range tools {
		validTools[tool.Name] = true
	}
	
	// Split by lines and look for TOOL_CALL patterns
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for TOOL_CALL: pattern
		if strings.HasPrefix(line, "TOOL_CALL:") {
			// Extract the tool call: tool_name(json_args)
			callPart := strings.TrimSpace(strings.TrimPrefix(line, "TOOL_CALL:"))
			
			// Find the opening parenthesis
			parenIdx := strings.Index(callPart, "(")
			if parenIdx == -1 {
				continue
			}
			
			toolName := strings.TrimSpace(callPart[:parenIdx])
			
			// Validate tool name
			if !validTools[toolName] {
				continue
			}
			
			// Extract JSON arguments
			argsStr := callPart[parenIdx+1:]
			// Find the closing parenthesis
			closeParenIdx := strings.LastIndex(argsStr, ")")
			if closeParenIdx != -1 {
				argsStr = argsStr[:closeParenIdx]
			}
			
			// Parse JSON arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
				// If parsing fails, try with empty args
				args = make(map[string]interface{})
			}
			
			// Create tool call with unique ID
			toolCalls = append(toolCalls, ToolCall{
				ID:        fmt.Sprintf("gemini_call_%d", i),
				Name:      toolName,
				Arguments: args,
			})
		}
	}
	
	return toolCalls
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
