package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/mcp"
	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
)

type GleanProvider struct {
	client   *glean.Glean
	instance string
}

func NewGleanProvider(apiKey, instance, model string) (*GleanProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Glean API key is required")
	}

	if instance == "" {
		return nil, fmt.Errorf("Glean instance is required (e.g., 'your-company')")
	}

	// Create Glean client using official SDK
	client := glean.New(
		glean.WithSecurity(apiKey),
		glean.WithInstance(instance),
	)

	return &GleanProvider{
		client:   client,
		instance: instance,
	}, nil
}

func (p *GleanProvider) GetProviderName() string {
	return "glean"
}

func (p *GleanProvider) Chat(ctx context.Context, prompt string, tools []*mcp.Tool, conversationHistory []Message) (*Response, error) {
	// Build system prompt with tools information
	systemPrompt := buildSystemPromptWithToolsGlean(tools)
	
	// Build messages using Glean SDK types
	messages := []components.ChatMessage{}
	
	// Add system prompt as first user message
	if systemPrompt != "" {
		messages = append(messages, components.ChatMessage{
			Fragments: []components.ChatMessageFragment{
				{Text: glean.String(systemPrompt)},
			},
		})
		messages = append(messages, components.ChatMessage{
			Fragments: []components.ChatMessageFragment{
				{Text: glean.String("Understood. I will use the TOOL_CALL format when I need to use tools.")},
			},
		})
	}
	
	// Add conversation history
	for _, msg := range conversationHistory {
		if msg.Role == "user" {
			messages = append(messages, components.ChatMessage{
				Fragments: []components.ChatMessageFragment{
					{Text: glean.String(msg.Content)},
				},
			})
		} else if msg.Role == "assistant" {
			content := msg.Content
			// Include tool results in the assistant message
			if len(msg.ToolResults) > 0 {
				content += "\n\nTool Results:\n"
				for _, tr := range msg.ToolResults {
					if tr.IsError {
						content += fmt.Sprintf(" Error: %s\n", tr.Content)
					} else {
						content += fmt.Sprintf("✓ %s\n", tr.Content)
					}
				}
			}
			messages = append(messages, components.ChatMessage{
				Fragments: []components.ChatMessageFragment{
					{Text: glean.String(content)},
				},
			})
		}
	}
	
	// Add current user prompt
	messages = append(messages, components.ChatMessage{
		Fragments: []components.ChatMessageFragment{
			{Text: glean.String(prompt)},
		},
	})

	// Create request using SDK
	chatReq := components.ChatRequest{
		Messages: messages,
	}

	// Call Glean API using SDK
	response, err := p.client.Client.Chat.Create(ctx, chatReq, nil)
	if err != nil {
		return nil, fmt.Errorf("Glean API error: %w", err)
	}

	// Extract content from response
	var content string
	if response.ChatResponse != nil && len(response.ChatResponse.Messages) > 0 {
		lastMessage := response.ChatResponse.Messages[len(response.ChatResponse.Messages)-1]
		for _, fragment := range lastMessage.Fragments {
			if fragment.Text != nil {
				content += *fragment.Text
			}
		}
	}

	// Extract tool calls from content using the same pattern as Gemini
	toolCalls := extractToolCallsGlean(content, tools)

	return &Response{
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: "stop",
		Usage:        nil, // Glean SDK doesn't expose token usage in simple response
	}, nil
}

// buildSystemPromptWithToolsGlean creates a system prompt that includes tool information
func buildSystemPromptWithToolsGlean(tools []*mcp.Tool) string {
	if len(tools) == 0 {
		return "You are a helpful AI assistant for infrastructure and DevOps tasks."
	}

	prompt := `You are CloudGenie AI, an intelligent assistant that helps users create, deploy, and manage infrastructure services.

Your Capabilities:
- Answer questions about infrastructure, DevOps, and cloud services
- Help users understand and design their infrastructure architecture
- Create and deploy infrastructure resources using available tools
- Retrieve information about existing resources and blueprints
- Guide users through infrastructure deployment processes

CRITICAL INSTRUCTIONS - When to Use Tools:

ALWAYS call tools for these requests (call tool ONLY ONCE):
✓ "Show blueprints" / "List blueprints" / "Get blueprints" → Call get_blueprints ONCE
✓ "Show resources" / "List resources" / "Get resources" → Call get_resources ONCE
✓ "Create/Deploy [specific resource]" (e.g., "Create a web server") → Call create_resource ONCE
✓ "Get details about [resource_name]" → Call get_resource_by_name ONCE

Capability Questions - CRITICAL RESPONSE FORMAT:
When user asks "Can you deploy [X]?" or "Do you support [X]?":
1. Call get_blueprints ONCE to check available blueprints
2. Search for a blueprint matching X (e.g., if X="database", look for "database", "db", "postgres", "mysql", etc.)
3. Give a CLEAR YES or NO answer first:
   
   If blueprint DOES NOT exist for X:
   "No, I cannot deploy a [X] at this time. The DevOps engineers haven't created a blueprint for [X] deployment yet. 
   
   I can currently deploy:
   - [blueprint-1]: [description]
   - [blueprint-2]: [description]
   
   If you need [X] deployment, please contact the DevOps team to create the appropriate blueprint."
   
   If blueprint EXISTS for X:
   "Yes, I can deploy a [X] using the [blueprint-name] blueprint. Would you like me to create one for you?"

NEVER call tools for these requests:
✗ "How can you help?" / "What can you do?" → Answer with your capabilities directly
✗ "What is Kubernetes?" / General knowledge questions → Answer from your knowledge
✗ Conversational questions or greetings → Respond naturally
✗ NEVER call the same tool multiple times in a single response

Tool Calling Format:
When you need to use a tool, use this exact format:

TOOL_CALL: tool_name({"param1": "value1", "param2": "value2"})

If a tool requires no parameters, use:

TOOL_CALL: tool_name({})

Available Tools:

`

	for _, tool := range tools {
		prompt += fmt.Sprintf("🔧 %s\n", tool.Name)
		prompt += fmt.Sprintf("   Description: %s\n", tool.Description)
		
		if tool.InputSchema != nil {
			if schema, ok := tool.InputSchema.(map[string]interface{}); ok {
				if properties, ok := schema["properties"].(map[string]interface{}); ok {
					if len(properties) > 0 {
						prompt += "   Parameters:\n"
						
						// Get required fields
						requiredFields := []string{}
						if required, ok := schema["required"].([]interface{}); ok {
							for _, req := range required {
								if reqStr, ok := req.(string); ok {
									requiredFields = append(requiredFields, reqStr)
								}
							}
						}
						
						for paramName, paramInfo := range properties {
							if paramMap, ok := paramInfo.(map[string]interface{}); ok {
								paramType := "any"
								if t, ok := paramMap["type"].(string); ok {
									paramType = t
								}
								paramDesc := ""
								if d, ok := paramMap["description"].(string); ok {
									paramDesc = d
								}
								
								// Check if required
								isRequired := false
								for _, req := range requiredFields {
									if req == paramName {
										isRequired = true
										break
									}
								}
								
								requiredMark := ""
								if isRequired {
									requiredMark = " [REQUIRED]"
								}
								
								prompt += fmt.Sprintf("      • %s (%s)%s: %s\n", paramName, paramType, requiredMark, paramDesc)
							}
						}
					} else {
						prompt += "   Parameters: None required\n"
					}
				}
			}
		}
		prompt += "\n"
	}

	prompt += `
Important Guidelines:
1. Be conversational and helpful in your responses
2. Call each tool ONLY ONCE per response - NEVER call the same tool multiple times
3. For capability questions ("Can you...?"), START your answer with a clear YES or NO
4. For capability questions, check blueprints and match the requested service name exactly
5. When using tools, use the TOOL_CALL format exactly as shown above
6. Provide all REQUIRED parameters when calling tools
7. After receiving tool results, analyze them and provide a clear, helpful response
8. If the user asks for something outside your capabilities, politely explain what you can and cannot do
9. Guide users through multi-step processes by breaking them down into clear steps
10. Always confirm destructive actions before executing them

Example Interactions:

User: "What is Kubernetes?"
You: Kubernetes is an open-source container orchestration platform... [Answer from knowledge, NO TOOL CALL]

User: "How can you help me?"
You: I can assist you with various tasks related to cloud infrastructure management as per Gruve's policies. I can create projects, setup CI/CD pipelines, Provide you details about available resources, projects, pipelines, and more. I can also help you understand infrastructure concepts and best practices followed in Gruve. I provision infrastructure resources and tools as per blueprints and guidelines defined by devops engineers in gruve. [NO TOOL CALL - answer directly]

User: "Can you deploy a database?" or "Do you support database deployment?"
You: Let me check what blueprints are available.
TOOL_CALL: get_blueprints({})

[After receiving results showing only "git-repo" blueprint:]
You: No, I cannot deploy a database at this time. The DevOps engineers haven't created a blueprint for database deployment yet.

I can currently deploy:
- git-repo: Creates a simple repository with a readme file

If you need database deployment, please contact the DevOps team to create the appropriate blueprint.

[IMPORTANT: Start with clear NO, explain why, list what IS available, provide guidance]

User: "Show me available blueprints" or "List all blueprints"
You: Let me fetch the available blueprints for you.
TOOL_CALL: get_blueprints({})
[Call ONCE and show results!]

User: "Create a web server resource called my-app"
You: I'll create that web server resource for you.
TOOL_CALL: create_resource({"name": "my-app", "blueprint": "web-server"})
[Call ONCE to perform the action!]

User: "What resources do I have?" or "Show my resources"
You: Let me retrieve your resources.
TOOL_CALL: get_resources({})
[Call ONCE and show results!]

Now, help the user with their request!
`

	return prompt
}

// extractToolCallsGlean extracts tool calls from the model's response
func extractToolCallsGlean(content string, tools []*mcp.Tool) []ToolCall {
	toolCalls := []ToolCall{}
	
	// Pattern 1: TOOL_CALL: tool_name({"param": "value"})
	pattern1 := regexp.MustCompile(`TOOL_CALL:\s*([a-zA-Z0-9_-]+)\s*\((.*?)\)`)
	matches1 := pattern1.FindAllStringSubmatch(content, -1)
	
	// Pattern 2: TOOL_CALL: tool_name (without parentheses)
	pattern2 := regexp.MustCompile(`TOOL_CALL:\s*([a-zA-Z0-9_-]+)\s*(?:\n|$)`)
	matches2 := pattern2.FindAllStringSubmatch(content, -1)
	
	// Process pattern 1 matches (with arguments)
	for i, match := range matches1 {
		if len(match) < 3 {
			continue
		}
		
		toolName := match[1]
		argsJSON := match[2]
		
		// Validate tool exists
		toolExists := false
		for _, tool := range tools {
			if tool.Name == toolName {
				toolExists = true
				break
			}
		}
		
		if !toolExists {
			continue
		}
		
		// Parse arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			// If parsing fails, try with empty args
			args = make(map[string]interface{})
		}
		
		toolCalls = append(toolCalls, ToolCall{
			ID:        fmt.Sprintf("call_%d", i+1),
			Name:      toolName,
			Arguments: args,
		})
	}
	
	// Process pattern 2 matches (without arguments) - only if pattern 1 didn't match
	if len(toolCalls) == 0 {
		for i, match := range matches2 {
			if len(match) < 2 {
				continue
			}
			
			toolName := match[1]
			
			// Validate tool exists
			toolExists := false
			for _, tool := range tools {
				if tool.Name == toolName {
					toolExists = true
					break
				}
			}
			
			if !toolExists {
				continue
			}
			
			// Use empty args for tools without parameters
			args := make(map[string]interface{})
			
			toolCalls = append(toolCalls, ToolCall{
				ID:        fmt.Sprintf("call_%d", i+1),
				Name:      toolName,
				Arguments: args,
			})
		}
	}
	
	return toolCalls
}
