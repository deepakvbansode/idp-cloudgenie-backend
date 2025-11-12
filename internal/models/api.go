package models

// Request and Response types for the API
type ChatRequest struct {
	Prompt   string                 `json:"prompt" binding:"required"`
	Provider string                 `json:"provider,omitempty"` // "openai" or "anthropic", defaults to openai
	Model    string                 `json:"model,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

type ChatResponse struct {
	Response    string                 `json:"response"`
	ToolCalls   []ToolCall             `json:"tool_calls,omitempty"`
	ToolResults []ToolResult           `json:"tool_results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

type HealthResponse struct {
	Status         string            `json:"status"`
	MCPServerReady bool              `json:"mcp_server_ready"`
	Services       map[string]string `json:"services"`
}

type ToolsResponse struct {
	Tools []ToolInfo `json:"tools"`
}

type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}
