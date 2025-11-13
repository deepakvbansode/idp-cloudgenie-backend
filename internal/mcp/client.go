package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client wraps the official MCP SDK client
type Client struct {
	mcpClient   *mcp.Client
	session     *mcp.ClientSession
	serverURL   string
	httpClient  *http.Client
	tools       []*mcp.Tool
	mu          sync.RWMutex
	initialized bool
}

// NewClient creates a new MCP client using the official SDK with HTTP transport
func NewClient(mcpServerURL string, env []string) (*Client, error) {
	// Create the official MCP client
	impl := &mcp.Implementation{
		Name:    "idp-cloudgenie-backend",
		Version: "1.0.0",
	}

	mcpClient := mcp.NewClient(impl, nil)

	client := &Client{
		mcpClient:  mcpClient,
		serverURL:  mcpServerURL,
		httpClient: &http.Client{},
		tools:      []*mcp.Tool{},
	}

	return client, nil
}

// Initialize performs the MCP initialization handshake over HTTP
func (c *Client) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Create StreamableClientTransport for HTTP communication
	transport := &mcp.StreamableClientTransport{
		Endpoint:   c.serverURL,
		HTTPClient: c.httpClient,
	}

	// Connect to the MCP server over HTTP
	ctx := context.Background()
	session, err := c.mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server via HTTP: %w", err)
	}

	c.session = session
	c.initialized = true

	return nil
}

// ListTools retrieves the list of available tools from the MCP server
func (c *Client) ListTools() ([]*mcp.Tool, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	ctx := context.Background()
	result, err := c.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	c.mu.Lock()
	c.tools = result.Tools
	c.mu.Unlock()

	return result.Tools, nil
}

// CallTool executes a tool on the MCP server
func (c *Client) CallTool(name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	ctx := context.Background()
	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: arguments,
	}

	result, err := c.session.CallTool(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	return result, nil
}

// GetTools returns the cached list of tools
func (c *Client) GetTools() []*mcp.Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tools
}

// Close closes the connection to the MCP server
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		return c.session.Close()
	}
	return nil
}