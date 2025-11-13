package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Use official SDK types as aliases for backward compatibility
type (
	// Tool represents an MCP tool (alias for SDK type)
	Tool = mcp.Tool

	// CallToolParams represents tool call parameters (alias for SDK type)
	CallToolParams = mcp.CallToolParams

	// CallToolResult represents tool call result (alias for SDK type)
	CallToolResult = mcp.CallToolResult

	// Content represents MCP content (alias for SDK type)
	Content = mcp.Content

	// TextContent represents text content (alias for SDK type)
	TextContent = mcp.TextContent
)

// ToolContent is a helper to extract text from Content interface
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// FormatToolResult converts CallToolResult to a readable string
func FormatToolResult(result *mcp.CallToolResult) string {
	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}
	return text
}

