# CloudGenie Backend Service - Project Summary

## Overview

A production-ready Backend-for-Frontend (BFF) service written in Go that provides a REST API interface between frontend applications and the CloudGenie Infrastructure Development Platform (IDP). The service integrates AI language models (OpenAI GPT-4) with the CloudGenie MCP (Model Context Protocol) server to enable natural language infrastructure management.

## Architecture

```
┌─────────────────┐
│                 │
│    Frontend     │
│  (React/Vue/    │
│   Angular)      │
│                 │
└────────┬────────┘
         │ HTTP/REST
         │ POST /api/v1/chat
         ▼
┌─────────────────────────────────────┐
│                                     │
│  CloudGenie Backend Service (Go)   │
│  ┌─────────────────────────────┐   │
│  │  HTTP API Layer (Gin)       │   │
│  ├─────────────────────────────┤   │
│  │  Orchestration Logic        │   │
│  ├─────────────────────────────┤   │
│  │  AI Provider (OpenAI)       │◄──┼──── OpenAI API
│  ├─────────────────────────────┤   │
│  │  MCP Client (JSON-RPC)      │   │
│  └──────────┬──────────────────┘   │
│             │                       │
└─────────────┼───────────────────────┘
              │ stdio/JSON-RPC
              ▼
┌──────────────────────────┐
│  CloudGenie MCP Server   │
│  (Go)                    │
└──────────┬───────────────┘
           │ HTTP
           ▼
┌──────────────────────────┐
│  CloudGenie Backend API  │
│  (Infrastructure Mgmt)   │
└──────────────────────────┘
```

## Key Features

### 1. **Multi-Provider AI Integration**

- OpenAI GPT-4 integration (primary)
- Anthropic Claude support (framework ready)
- Easy to add more providers

### 2. **MCP Client Implementation**

- Full JSON-RPC 2.0 over stdio communication
- Tool discovery and execution
- Automatic initialization and handshaking
- Error handling and retries

### 3. **Intelligent Orchestration**

- Automatic tool discovery from MCP server
- Iterative tool execution (up to 5 rounds)
- Context-aware conversation handling
- Tool result formatting and aggregation

### 4. **RESTful API**

- Clean, well-documented endpoints
- JSON request/response format
- CORS support for web applications
- Health monitoring endpoints

### 5. **Production-Ready**

- Environment-based configuration
- Graceful shutdown handling
- Structured logging
- Error handling and recovery
- Docker support

## Project Structure

```
idp-cloudgenie-backend/
├── main.go                          # Application entry point
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── Makefile                         # Build and development commands
├── Dockerfile                       # Container image definition
├── .env.example                     # Environment variables template
├── .gitignore                       # Git ignore patterns
│
├── README.md                        # Comprehensive documentation
├── API.md                           # API documentation
├── QUICKSTART.md                    # Quick start guide
│
├── internal/
│   ├── ai/                          # AI provider implementations
│   │   ├── provider.go             # Provider interface definition
│   │   ├── openai.go               # OpenAI GPT integration
│   │   └── anthropic.go            # Anthropic Claude integration (stub)
│   │
│   ├── config/                      # Configuration management
│   │   └── config.go               # Environment variable loader
│   │
│   ├── handlers/                    # HTTP request handlers
│   │   ├── handlers.go             # API route handlers
│   │   └── orchestration.go        # AI-MCP orchestration logic
│   │
│   ├── mcp/                         # MCP client implementation
│   │   ├── client.go               # MCP JSON-RPC client
│   │   └── types.go                # MCP protocol type definitions
│   │
│   └── models/                      # Data models
│       └── api.go                  # API request/response structures
│
└── pkg/
    └── utils/                       # Utility functions
```

## API Endpoints

### POST /api/v1/chat

Send natural language prompts for infrastructure management.

**Example:**

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Create a PostgreSQL database for production"}'
```

### GET /api/v1/health

Check service health and component status.

### GET /api/v1/tools

List all available CloudGenie tools.

## Technologies Used

### Core

- **Go 1.20+**: Primary programming language
- **Gin**: HTTP web framework
- **godotenv**: Environment variable management

### AI Integration

- **go-openai**: OpenAI API client
- Support for multiple AI providers

### Infrastructure

- **MCP Protocol**: Model Context Protocol for tool communication
- **JSON-RPC 2.0**: Communication protocol with MCP server

### Development

- **Make**: Build automation
- **Docker**: Containerization
- **CORS**: Cross-origin resource sharing

## Configuration

All configuration is managed via environment variables:

| Variable                 | Description         | Default                 |
| ------------------------ | ------------------- | ----------------------- |
| `SERVER_HOST`            | Server bind address | `0.0.0.0`               |
| `SERVER_PORT`            | Server port         | `8081`                  |
| `DEFAULT_AI_PROVIDER`    | AI provider         | `openai`                |
| `OPENAI_API_KEY`         | OpenAI API key      | (required)              |
| `OPENAI_MODEL`           | OpenAI model        | `gpt-4-turbo-preview`   |
| `MCP_SERVER_PATH`        | Path to MCP server  | (required)              |
| `CLOUDGENIE_BACKEND_URL` | CloudGenie API URL  | `http://localhost:8080` |
| `ALLOWED_ORIGINS`        | CORS origins        | `*`                     |

## How It Works

1. **Frontend sends prompt** → Backend receives natural language request
2. **Load tools** → Backend queries MCP server for available tools
3. **AI processing** → OpenAI analyzes prompt with tool definitions
4. **Tool execution** → If needed, backend calls MCP tools via JSON-RPC
5. **Result processing** → MCP executes against CloudGenie API
6. **Iteration** → AI analyzes results, may call more tools
7. **Final response** → Backend returns structured response to frontend

## Available CloudGenie Tools

The MCP server provides these tools (automatically discovered):

1. **cloudgenie_health_check** - Check API health
2. **cloudgenie_get_blueprints** - List infrastructure blueprints
3. **cloudgenie_get_blueprint** - Get specific blueprint details
4. **cloudgenie_get_resources** - List all resources
5. **cloudgenie_get_resource** - Get specific resource details
6. **cloudgenie_create_resource** - Create new infrastructure resource
7. **cloudgenie_delete_resource** - Delete a resource
8. **cloudgenie_update_resource_status** - Update resource status

## Example Use Cases

### 1. Infrastructure Discovery

```
User: "What blueprints are available?"
AI: Calls cloudgenie_get_blueprints
Response: Lists all available infrastructure templates
```

### 2. Resource Creation

```
User: "Create a PostgreSQL database for production"
AI:
  1. Calls cloudgenie_get_blueprints (find PostgreSQL)
  2. Calls cloudgenie_create_resource (create database)
Response: "Created PostgreSQL database 'prod-db-123'"
```

### 3. Status Monitoring

```
User: "What's the status of my resources?"
AI: Calls cloudgenie_get_resources
Response: Lists all resources with current status
```

### 4. Complex Operations

```
User: "Set up a complete web stack with load balancer, 3 web servers, and database"
AI: (Multiple iterations)
  1. Get available blueprints
  2. Create load balancer
  3. Create 3 web server instances
  4. Create database
  5. Update configurations
Response: "Created complete web stack with 5 resources"
```

## Development Commands

```bash
# Initialize project
make init

# Install dependencies
make install

# Build binary
make build

# Run in development mode
make dev

# Run tests
make test

# Clean build artifacts
make clean

# View all commands
make help
```

## Deployment

### Local Development

```bash
go run main.go
```

### Production Binary

```bash
make build-prod
./cloudgenie-backend
```

### Docker

```bash
docker build -t cloudgenie-backend:latest .
docker run -p 8081:8081 --env-file .env cloudgenie-backend:latest
```

## Security Considerations

⚠️ **For Production Use:**

1. **Add Authentication**: Implement JWT or OAuth2
2. **Rate Limiting**: Prevent API abuse
3. **Input Validation**: Sanitize all inputs
4. **HTTPS**: Use TLS for all communications
5. **API Keys**: Secure storage (vault/secrets manager)
6. **CORS**: Restrict allowed origins
7. **Logging**: Implement proper audit logging
8. **Monitoring**: Add metrics and alerting

## Future Enhancements

- [ ] WebSocket support for streaming responses
- [ ] Conversation history persistence
- [ ] Multi-user support with authentication
- [ ] Support for multiple MCP servers
- [ ] Caching layer for tool responses
- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Kubernetes manifests
- [ ] Helm charts
- [ ] CI/CD pipeline

## Performance Characteristics

- **Latency**: 2-5 seconds (depends on AI provider and tool complexity)
- **Throughput**: Handles concurrent requests via Go's goroutines
- **Memory**: ~50-100MB base (scales with concurrent requests)
- **CPU**: Minimal (most time spent waiting for I/O)

## Testing Strategy

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./... -tags=integration

# Coverage report
make test-coverage
```

## Monitoring

Health check endpoint provides:

- MCP server connectivity status
- AI provider status
- Available tools count
- Overall service health

## Troubleshooting

See [QUICKSTART.md](QUICKSTART.md) for common issues and solutions.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Related Projects

- [CloudGenie MCP Server](https://github.com/deepakvbansode/idp-cloudgenie-mcp-server)
- [Model Context Protocol](https://modelcontextprotocol.io/)

## Support

For issues and questions:

- GitHub Issues: Report bugs and feature requests
- Documentation: Check README.md and API.md
- Examples: See QUICKSTART.md for usage examples

---

**Built with ❤️ for CloudGenie IDP**
