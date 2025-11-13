package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/ai"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/models"
)

const MaxToolIterations = 5

// OrchestrationService coordinates between AI and MCP server
type OrchestrationService struct {
	mcpClient  *mcp.Client
	aiProvider ai.Provider
	tools      []*mcp.Tool
}

func NewOrchestrationService(mcpClient *mcp.Client, aiProvider ai.Provider) (*OrchestrationService, error) {
	// Initialize MCP client and get tools
	if err := mcpClient.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	tools, err := mcpClient.ListTools()
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return &OrchestrationService{
		mcpClient:  mcpClient,
		aiProvider: aiProvider,
		tools:      tools,
	}, nil
}

// ProcessPrompt processes a user prompt and coordinates with AI and MCP
func (s *OrchestrationService) ProcessPrompt(ctx context.Context, request *models.ChatRequest) (*models.ChatResponse, error) {
	conversationHistory := []ai.Message{}
	allToolCalls := []models.ToolCall{}
	allToolResults := []models.ToolResult{}

	currentPrompt := request.Prompt
	iteration := 0

	for iteration < MaxToolIterations {
		iteration++

		// Call AI with current prompt and tools
		aiResponse, err := s.aiProvider.Chat(ctx, currentPrompt, s.tools, conversationHistory)
		if err != nil {
			return nil, fmt.Errorf("AI provider error: %w", err)
		}

		// Add assistant response to history
		conversationHistory = append(conversationHistory, ai.Message{
			Role:    "assistant",
			Content: aiResponse.Content,
		})

		// If no tool calls, we're done
		if len(aiResponse.ToolCalls) == 0 {
			return &models.ChatResponse{
				Response:    aiResponse.Content,
				ToolCalls:   allToolCalls,
				ToolResults: allToolResults,
				Metadata: map[string]interface{}{
					"iterations":     iteration,
					"finish_reason":  aiResponse.FinishReason,
					"provider":       s.aiProvider.GetProviderName(),
					"tools_available": len(s.tools),
				},
			}, nil
		}

		// Execute tool calls
		toolResults := []ai.ToolResult{}
		for _, toolCall := range aiResponse.ToolCalls {
			log.Printf("Executing tool: %s with args: %v", toolCall.Name, toolCall.Arguments)

			// Call MCP tool
			mcpResult, err := s.mcpClient.CallTool(toolCall.Name, toolCall.Arguments)
			if err != nil {
				errMsg := fmt.Sprintf("Error calling tool %s: %v", toolCall.Name, err)
				log.Printf(errMsg)
				
				toolResults = append(toolResults, ai.ToolResult{
					ToolCallID: toolCall.ID,
					Content:    errMsg,
					IsError:    true,
				})

				allToolResults = append(allToolResults, models.ToolResult{
					ToolCallID: toolCall.ID,
					Name:       toolCall.Name,
					Content:    errMsg,
					IsError:    true,
				})
				continue
			}

			// Format tool result
			resultContent := formatToolResult(mcpResult)
			toolResults = append(toolResults, ai.ToolResult{
				ToolCallID: toolCall.ID,
				Content:    resultContent,
				IsError:    mcpResult.IsError,
			})

			// Track for response
			allToolCalls = append(allToolCalls, models.ToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Name,
				Arguments: toolCall.Arguments,
			})

			allToolResults = append(allToolResults, models.ToolResult{
				ToolCallID: toolCall.ID,
				Name:       toolCall.Name,
				Content:    resultContent,
				IsError:    mcpResult.IsError,
			})
		}

		// Add tool results to conversation history
		conversationHistory = append(conversationHistory, ai.Message{
			Role:        "assistant",
			ToolResults: toolResults,
		})

		// Prepare next prompt with tool results
		currentPrompt = formatToolResultsForPrompt(toolResults)
	}

	// If we hit max iterations, return what we have
	return &models.ChatResponse{
		Response:    "Maximum tool execution iterations reached. Please try breaking down your request.",
		ToolCalls:   allToolCalls,
		ToolResults: allToolResults,
		Metadata: map[string]interface{}{
			"iterations":    iteration,
			"max_reached":   true,
			"provider":      s.aiProvider.GetProviderName(),
			"tools_available": len(s.tools),
		},
	}, nil
}

// GetAvailableTools returns the list of available MCP tools
func (s *OrchestrationService) GetAvailableTools() []models.ToolInfo {
	toolInfos := make([]models.ToolInfo, len(s.tools))
	for i, tool := range s.tools {
		// Type assert InputSchema to map[string]interface{}
		var params map[string]interface{}
		if tool.InputSchema != nil {
			if schema, ok := tool.InputSchema.(map[string]interface{}); ok {
				params = schema
			}
		}
		
		toolInfos[i] = models.ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  params,
		}
	}
	return toolInfos
}

// HealthCheck checks the health of MCP and AI services
func (s *OrchestrationService) HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)

	// Check MCP client
	if s.mcpClient != nil {
		status["mcp_client"] = "connected"
	} else {
		status["mcp_client"] = "disconnected"
	}

	// Check AI provider
	if s.aiProvider != nil {
		status["ai_provider"] = s.aiProvider.GetProviderName()
	} else {
		status["ai_provider"] = "not configured"
	}

	// Check tools
	status["tools_count"] = fmt.Sprintf("%d", len(s.tools))

	return status
}

// formatToolResult formats the MCP tool result into a string
func formatToolResult(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return "Tool executed successfully with no output"
	}

	var textParts []string
	for _, content := range result.Content {
		// Use type assertion to extract text from Content interface
		if textContent, ok := content.(*mcp.TextContent); ok {
			textParts = append(textParts, textContent.Text)
		}
	}

	if len(textParts) == 0 {
		// Try to marshal the whole result as JSON
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return "Tool result could not be formatted"
		}
		return string(jsonBytes)
	}

	if len(textParts) == 1 {
		return textParts[0]
	}

	// Multiple text parts, join them
	resultBytes, _ := json.MarshalIndent(textParts, "", "  ")
	return string(resultBytes)
}

// formatToolResultsForPrompt formats tool results for the next AI prompt
func formatToolResultsForPrompt(results []ai.ToolResult) string {
	if len(results) == 0 {
		return ""
	}

	prompt := "Tool execution results:\n\n"
	for _, result := range results {
		if result.IsError {
			prompt += fmt.Sprintf("❌ Error: %s\n\n", result.Content)
		} else {
			prompt += fmt.Sprintf("✓ Success: %s\n\n", result.Content)
		}
	}

	prompt += "Please analyze these results and provide a response to the user."
	return prompt
}
