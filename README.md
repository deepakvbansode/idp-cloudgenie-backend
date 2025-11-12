# CloudGenie Backend Service

A backend-for-frontend (BFF) service written in Go that bridges frontend applications with AI language models and the CloudGenie MCP (Model Context Protocol) server for intelligent infrastructure management.

## Overview

This service provides a REST API that:

- Accepts user prompts from frontend applications
- Communicates with AI models (OpenAI GPT-4, Anthropic Claude)
- Orchestrates tool execution via the CloudGenie MCP server
- Returns structured responses with tool execution results

## Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────────┐
│             │  HTTP   │                  │  AI API │                 │
│  Frontend   │◄───────►│  Backend Service │◄────────│  OpenAI/Claude  │
│             │         │     (This)       │         │                 │
└─────────────┘         └──────────────────┘         └─────────────────┘
                               │
                               │ stdio/JSON-RPC
                               ▼
                        ┌──────────────────┐         ┌─────────────────┐
                        │                  │  HTTP   │                 │
                        │   MCP Server     │◄────────│ CloudGenie API  │
                        │                  │         │    Backend      │
                        └──────────────────┘         └─────────────────┘
```

## Features

- ✅ RESTful API for chat-based interactions
- ✅ Multi-provider AI support (OpenAI, Anthropic)
- ✅ Automatic tool discovery from MCP server
- ✅ Intelligent tool orchestration with iterative refinement
- ✅ CORS support for web frontends
- ✅ Health monitoring endpoints
- ✅ Graceful shutdown handling
- ✅ Environment-based configuration

## Prerequisites

- Go 1.21 or higher
- CloudGenie MCP Server (compiled binary)
- OpenAI API key OR Anthropic API key
- CloudGenie backend API (running or accessible)

## Installation

### 1. Clone the repository

```bash
cd /Users/deepakbansode/Documents/projects/idp-cloudgenie/idp-cloudgenie-backend
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Build the MCP Server

First, you need to build the CloudGenie MCP server:

```bash
# Clone and build the MCP server
cd ..
git clone https://github.com/deepakvbansode/idp-cloudgenie-mcp-server.git
cd idp-cloudgenie-mcp-server
go build -o cloudgenie-mcp-server .
```

### 4. Configure environment variables

Copy the example environment file and edit it:

```bash
cd ../idp-cloudgenie-backend
cp .env.example .env
```

Edit `.env` with your actual values:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8081

# AI Provider
DEFAULT_AI_PROVIDER=openai
OPENAI_API_KEY=sk-your-actual-openai-key
OPENAI_MODEL=gpt-4-turbo-preview

# MCP Server Path (update this to actual path)
MCP_SERVER_PATH=/path/to/cloudgenie-mcp-server

# CloudGenie Backend URL
CLOUDGENIE_BACKEND_URL=http://localhost:8080
```

### 5. Build the backend service

```bash
go build -o cloudgenie-backend .
```

## Usage

### Running the Server

```bash
./cloudgenie-backend
```

The server will start on `http://localhost:8081` (or the configured host/port).

### API Endpoints

#### POST /api/v1/chat

Send a chat prompt for processing.

**Request:**

```json
{
  "prompt": "Create a PostgreSQL database for production",
  "provider": "openai",
  "model": "gpt-4-turbo-preview"
}
```

**Response:**

```json
{
  "response": "I've created a PostgreSQL database for production...",
  "tool_calls": [
    {
      "id": "call_123",
      "name": "cloudgenie_create_resource",
      "arguments": {
        "name": "prod-postgres-db",
        "type": "database",
        "blueprint_id": "postgresql-blueprint"
      }
    }
  ],
  "tool_results": [
    {
      "tool_call_id": "call_123",
      "name": "cloudgenie_create_resource",
      "content": "{\"id\": \"res-456\", \"status\": \"created\"}",
      "is_error": false
    }
  ],
  "metadata": {
    "iterations": 2,
    "finish_reason": "stop",
    "provider": "openai",
    "tools_available": 8
  }
}
```

#### GET /api/v1/health

Check service health.

**Response:**

```json
{
  "status": "healthy",
  "mcp_server_ready": true,
  "services": {
    "mcp_client": "connected",
    "ai_provider": "openai",
    "tools_count": "8"
  }
}
```

#### GET /api/v1/tools

List available CloudGenie tools.

**Response:**

```json
{
  "tools": [
    {
      "name": "cloudgenie_health_check",
      "description": "Checks the health status of the CloudGenie backend API",
      "parameters": {}
    },
    {
      "name": "cloudgenie_get_blueprints",
      "description": "Retrieves all available blueprints",
      "parameters": {}
    }
  ]
}
```

## Example Usage with curl

### Send a chat message

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Show me all available blueprints"
  }'
```

### Check health

```bash
curl http://localhost:8081/api/v1/health
```

### List tools

```bash
curl http://localhost:8081/api/v1/tools
```

## Configuration Options

| Environment Variable     | Description               | Default                       |
| ------------------------ | ------------------------- | ----------------------------- |
| `SERVER_HOST`            | Server bind address       | `0.0.0.0`                     |
| `SERVER_PORT`            | Server port               | `8081`                        |
| `DEFAULT_AI_PROVIDER`    | AI provider to use        | `openai`                      |
| `OPENAI_API_KEY`         | OpenAI API key            | (required if using OpenAI)    |
| `OPENAI_MODEL`           | OpenAI model name         | `gpt-4-turbo-preview`         |
| `ANTHROPIC_API_KEY`      | Anthropic API key         | (required if using Anthropic) |
| `ANTHROPIC_MODEL`        | Anthropic model name      | `claude-3-5-sonnet-20241022`  |
| `MCP_SERVER_PATH`        | Path to MCP server binary | (required)                    |
| `CLOUDGENIE_BACKEND_URL` | CloudGenie API URL        | `http://localhost:8080`       |
| `ALLOWED_ORIGINS`        | CORS allowed origins      | `*`                           |

## Project Structure

```
idp-cloudgenie-backend/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── .env.example                 # Environment variables template
├── internal/
│   ├── ai/                      # AI provider integrations
│   │   ├── provider.go         # Provider interface
│   │   ├── openai.go           # OpenAI implementation
│   │   └── anthropic.go        # Anthropic implementation
│   ├── config/                  # Configuration management
│   │   └── config.go           # Config loader
│   ├── handlers/                # HTTP handlers
│   │   ├── handlers.go         # API route handlers
│   │   └── orchestration.go    # AI-MCP orchestration logic
│   ├── mcp/                     # MCP client
│   │   ├── client.go           # MCP client implementation
│   │   └── types.go            # MCP protocol types
│   └── models/                  # Data models
│       └── api.go              # API request/response models
└── pkg/
    └── utils/                   # Utility functions
```

## How It Works

1. **Frontend sends a prompt** to `/api/v1/chat`
2. **Backend loads available tools** from the MCP server
3. **AI model processes the prompt** with tool definitions
4. **If AI requests tool execution**, backend calls MCP server via stdio/JSON-RPC
5. **MCP server executes tools** against CloudGenie API
6. **Results are sent back to AI** for analysis
7. **AI generates final response** based on tool results
8. **Backend returns structured response** to frontend

This process can iterate up to 5 times for complex multi-step operations.

## Development

### Running in development mode

```bash
go run main.go
```

### Running tests

```bash
go test ./...
```

### Building for production

```bash
CGO_ENABLED=0 go build -ldflags="-w -s" -o cloudgenie-backend .
```

## Troubleshooting

### MCP Server Connection Issues

- Ensure the MCP server binary path is correct
- Check that the MCP server has execute permissions: `chmod +x /path/to/cloudgenie-mcp-server`
- Verify CloudGenie backend API is accessible

### AI Provider Errors

- Verify API keys are correct and have sufficient credits
- Check API key permissions and rate limits
- Ensure network connectivity to AI provider APIs

### Tool Execution Failures

- Check CloudGenie backend API is running and accessible
- Verify MCP server can reach CloudGenie API
- Review MCP server logs for detailed error messages

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Related Projects

- [CloudGenie MCP Server](https://github.com/deepakvbansode/idp-cloudgenie-mcp-server) - MCP server for CloudGenie
- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP specification

## Support

For issues and questions:

- Open an issue on GitHub
- Check existing issues for solutions

## Roadmap

- [ ] Add streaming responses support
- [ ] Implement conversation history persistence
- [ ] Add authentication and authorization
- [ ] Support for multiple MCP servers
- [ ] WebSocket support for real-time updates
- [ ] Metrics and observability
- [ ] Docker containerization
- [ ] Kubernetes deployment manifests
