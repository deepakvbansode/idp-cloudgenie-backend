# CloudGenie Backend API Documentation

## Base URL

```
http://localhost:8081/api/v1
```

## Authentication

Currently, the API does not require authentication. This will be added in a future version.

## Endpoints

### 1. Chat Endpoint

Send a natural language prompt to interact with CloudGenie infrastructure.

**Endpoint:** `POST /api/v1/chat`

**Request Body:**

```json
{
  "prompt": "string (required) - The user's natural language prompt",
  "provider": "string (optional) - AI provider: 'openai' or 'anthropic'. Defaults to configured provider",
  "model": "string (optional) - Specific model to use. Defaults to configured model",
  "context": "object (optional) - Additional context for the conversation"
}
```

**Example Request:**

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a PostgreSQL database named prod-db with 100GB storage in us-east-1"
  }'
```

**Response:**

```json
{
  "response": "I've created a PostgreSQL database named prod-db with 100GB storage in the us-east-1 region...",
  "tool_calls": [
    {
      "id": "call_abc123",
      "name": "cloudgenie_create_resource",
      "arguments": {
        "name": "prod-db",
        "type": "database",
        "blueprint_id": "postgresql-blueprint",
        "properties": {
          "storage": "100GB",
          "region": "us-east-1"
        }
      }
    }
  ],
  "tool_results": [
    {
      "tool_call_id": "call_abc123",
      "name": "cloudgenie_create_resource",
      "content": "{\"id\": \"res-789\", \"name\": \"prod-db\", \"status\": \"creating\"}",
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

**Response Fields:**

- `response` (string): The AI's final response to the user
- `tool_calls` (array): List of tools that were called during processing
  - `id` (string): Unique identifier for this tool call
  - `name` (string): Name of the tool that was called
  - `arguments` (object): Arguments passed to the tool
- `tool_results` (array): Results from tool executions
  - `tool_call_id` (string): ID of the corresponding tool call
  - `name` (string): Name of the tool
  - `content` (string): Result content from the tool
  - `is_error` (boolean): Whether the tool execution resulted in an error
- `metadata` (object): Additional information about the request processing
  - `iterations` (number): Number of AI-tool interaction cycles
  - `finish_reason` (string): Why the AI stopped generating
  - `provider` (string): AI provider used
  - `tools_available` (number): Number of tools available to the AI

**Status Codes:**

- `200 OK`: Request processed successfully
- `400 Bad Request`: Invalid request format
- `500 Internal Server Error`: Server error during processing

---

### 2. Health Check

Check the health status of the backend service and its dependencies.

**Endpoint:** `GET /api/v1/health`

**Example Request:**

```bash
curl http://localhost:8081/api/v1/health
```

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

**Response Fields:**

- `status` (string): Overall health status: "healthy" or "degraded"
- `mcp_server_ready` (boolean): Whether MCP server connection is active
- `services` (object): Status of individual service components
  - `mcp_client` (string): MCP client connection status
  - `ai_provider` (string): Active AI provider name
  - `tools_count` (string): Number of available tools

**Status Codes:**

- `200 OK`: Service is operational (may be degraded)

---

### 3. List Available Tools

Get the list of available CloudGenie tools that the AI can use.

**Endpoint:** `GET /api/v1/tools`

**Example Request:**

```bash
curl http://localhost:8081/api/v1/tools
```

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
      "description": "Retrieves all available blueprints from CloudGenie",
      "parameters": {}
    },
    {
      "name": "cloudgenie_get_blueprint",
      "description": "Retrieves detailed information about a specific blueprint",
      "parameters": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "The unique identifier of the blueprint"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "cloudgenie_get_resources",
      "description": "Retrieves all resources from CloudGenie",
      "parameters": {}
    },
    {
      "name": "cloudgenie_get_resource",
      "description": "Retrieves detailed information about a specific resource",
      "parameters": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "The unique identifier of the resource"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "cloudgenie_create_resource",
      "description": "Creates a new resource in CloudGenie",
      "parameters": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "description": "Name of the resource"
          },
          "type": {
            "type": "string",
            "description": "Type of the resource (e.g., 'vm', 'database', 'storage')"
          },
          "blueprint_id": {
            "type": "string",
            "description": "ID of the blueprint to use"
          },
          "properties": {
            "type": "object",
            "description": "Additional properties for the resource"
          }
        },
        "required": ["name", "type", "blueprint_id"]
      }
    },
    {
      "name": "cloudgenie_delete_resource",
      "description": "Deletes a resource from CloudGenie",
      "parameters": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "The unique identifier of the resource to delete"
          }
        },
        "required": ["id"]
      }
    },
    {
      "name": "cloudgenie_update_resource_status",
      "description": "Updates the status of a resource",
      "parameters": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "The unique identifier of the resource"
          },
          "status": {
            "type": "string",
            "description": "New status (e.g., 'running', 'stopped', 'error')"
          }
        },
        "required": ["id", "status"]
      }
    }
  ]
}
```

**Response Fields:**

- `tools` (array): List of available tools
  - `name` (string): Tool identifier
  - `description` (string): Human-readable description of what the tool does
  - `parameters` (object): JSON Schema defining the tool's parameters

**Status Codes:**

- `200 OK`: Successfully retrieved tools list

---

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "code": 500
}
```

**Common Error Codes:**

- `invalid_request`: The request format is invalid
- `processing_error`: An error occurred while processing the request
- `mcp_error`: Error communicating with MCP server
- `ai_error`: Error from the AI provider

---

## Example Use Cases

### 1. List Available Infrastructure Blueprints

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Show me all available infrastructure blueprints"
  }'
```

### 2. Create a Virtual Machine

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a virtual machine named web-server-01 with 4 CPUs and 8GB RAM"
  }'
```

### 3. Check Resource Status

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is the status of all my resources?"
  }'
```

### 4. Delete a Resource

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Delete the resource with ID res-123"
  }'
```

### 5. Complex Multi-Step Operation

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a complete web application stack with a load balancer, 3 web servers, and a PostgreSQL database"
  }'
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. This may be added in future versions.

---

## WebSocket Support

WebSocket support for streaming responses is planned for a future release.

---

## Versioning

The API is currently at version 1 (`/api/v1`). Breaking changes will result in a new API version.
