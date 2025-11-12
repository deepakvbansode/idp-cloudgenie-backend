package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client represents a connection to an MCP server
type Client struct {
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	scanner       *bufio.Scanner
	requestID     atomic.Int64
	pendingCalls  map[int64]chan *JSONRPCResponse
	callsMutex    sync.RWMutex
	initialized   bool
	serverInfo    *ServerInfo
	tools         []Tool
	mu            sync.Mutex
}

// NewClient creates a new MCP client that communicates with the server via stdio
func NewClient(mcpServerPath string, env []string) (*Client, error) {
	cmd := exec.Command(mcpServerPath)
	cmd.Env = append(cmd.Env, env...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	client := &Client{
		cmd:          cmd,
		stdin:        stdin,
		stdout:       stdout,
		stderr:       stderr,
		scanner:      bufio.NewScanner(stdout),
		pendingCalls: make(map[int64]chan *JSONRPCResponse),
	}

	// Start reading responses in a goroutine
	go client.readResponses()

	// Read stderr for debugging
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[MCP Server stderr]: %s\n", scanner.Text())
		}
	}()

	return client, nil
}

// Initialize performs the MCP initialization handshake
func (c *Client) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ClientInfo: ClientInfo{
			Name:    "idp-cloudgenie-backend",
			Version: "1.0.0",
		},
	}

	result, err := c.call("initialize", params)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Parse initialization result
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal init result: %w", err)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(resultBytes, &initResult); err != nil {
		return fmt.Errorf("failed to unmarshal init result: %w", err)
	}

	c.serverInfo = &initResult.ServerInfo
	c.initialized = true

	// Call initialized notification
	if err := c.notify("notifications/initialized", nil); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// ListTools retrieves the list of available tools from the MCP server
func (c *Client) ListTools() ([]Tool, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	result, err := c.call("tools/list", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tools result: %w", err)
	}

	var toolsResult ToolsListResult
	if err := json.Unmarshal(resultBytes, &toolsResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools result: %w", err)
	}

	c.mu.Lock()
	c.tools = toolsResult.Tools
	c.mu.Unlock()

	return toolsResult.Tools, nil
}

// CallTool executes a tool on the MCP server
func (c *Client) CallTool(name string, arguments map[string]interface{}) (*ToolCallResult, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	params := ToolCallParams{
		Name:      name,
		Arguments: arguments,
	}

	result, err := c.call("tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool call result: %w", err)
	}

	var toolResult ToolCallResult
	if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call result: %w", err)
	}

	return &toolResult, nil
}

// GetTools returns the cached list of tools
func (c *Client) GetTools() []Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tools
}

// call sends a JSON-RPC request and waits for the response
func (c *Client) call(method string, params interface{}) (interface{}, error) {
	id := c.requestID.Add(1)
	responseChan := make(chan *JSONRPCResponse, 1)

	c.callsMutex.Lock()
	c.pendingCalls[id] = responseChan
	c.callsMutex.Unlock()

	defer func() {
		c.callsMutex.Lock()
		delete(c.pendingCalls, id)
		c.callsMutex.Unlock()
	}()

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := c.stdin.Write(append(requestBytes, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	response := <-responseChan

	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// notify sends a JSON-RPC notification (no response expected)
func (c *Client) notify(method string, params interface{}) error {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if _, err := c.stdin.Write(append(requestBytes, '\n')); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// readResponses reads JSON-RPC responses from the server
func (c *Client) readResponses() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		
		var response JSONRPCResponse
		if err := json.Unmarshal(line, &response); err != nil {
			fmt.Printf("Failed to unmarshal response: %v\n", err)
			continue
		}

		// Handle response
		if response.ID != nil {
			var id int64
			switch v := response.ID.(type) {
			case float64:
				id = int64(v)
			case int64:
				id = v
			case int:
				id = int64(v)
			default:
				fmt.Printf("Unknown ID type: %T\n", response.ID)
				continue
			}

			c.callsMutex.RLock()
			responseChan, ok := c.pendingCalls[id]
			c.callsMutex.RUnlock()

			if ok {
				responseChan <- &response
			}
		}
	}

	if err := c.scanner.Err(); err != nil {
		fmt.Printf("Scanner error: %v\n", err)
	}
}

// Close closes the connection to the MCP server
func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}
