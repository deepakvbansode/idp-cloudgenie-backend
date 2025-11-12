# Quick Start Guide

Get the CloudGenie Backend Service up and running in 5 minutes!

## Prerequisites

Before you begin, ensure you have:

- ✅ Go 1.20 or higher installed
- ✅ OpenAI API key (get one at https://platform.openai.com/)
- ✅ CloudGenie MCP Server built and ready

## Step 1: Build the MCP Server

```bash
# Navigate to your projects directory
cd /Users/deepakbansode/Documents/projects/idp-cloudgenie

# Clone the MCP server repository
git clone https://github.com/deepakvbansode/idp-cloudgenie-mcp-server.git

# Build the MCP server
cd idp-cloudgenie-mcp-server
go build -o cloudgenie-mcp-server .

# Verify it was built successfully
ls -lh cloudgenie-mcp-server
```

## Step 2: Configure the Backend

```bash
# Navigate to backend directory
cd /Users/deepakbansode/Documents/projects/idp-cloudgenie/idp-cloudgenie-backend

# Create .env file from template
cp .env.example .env

# Edit .env file with your settings
# Replace the following values:
# - OPENAI_API_KEY with your actual OpenAI API key
# - MCP_SERVER_PATH with the full path to cloudgenie-mcp-server binary
```

**Example .env:**

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8081

DEFAULT_AI_PROVIDER=openai
OPENAI_API_KEY=sk-proj-your-actual-key-here
OPENAI_MODEL=gpt-4-turbo-preview

MCP_SERVER_PATH=/Users/deepakbansode/Documents/projects/idp-cloudgenie/idp-cloudgenie-mcp-server/cloudgenie-mcp-server

CLOUDGENIE_BACKEND_URL=http://localhost:8080

ALLOWED_ORIGINS=*
```

## Step 3: Start the CloudGenie Backend API (if not already running)

The MCP server needs to connect to a CloudGenie backend API. Make sure it's running:

```bash
# This depends on your CloudGenie setup
# Example:
# cd /path/to/cloudgenie-api
# ./cloudgenie-api
```

## Step 4: Build and Run the Backend Service

```bash
# Build the backend
go build -o cloudgenie-backend .

# Run the service
./cloudgenie-backend
```

You should see output like:

```
Starting CloudGenie Backend Service...
AI Provider: openai
MCP Server Path: /Users/deepakbansode/Documents/projects/idp-cloudgenie/idp-cloudgenie-mcp-server/cloudgenie-mcp-server
CloudGenie Backend URL: http://localhost:8080
Initializing MCP client...
Initializing AI provider: openai
Initializing orchestration service...
Server starting on 0.0.0.0:8081
Available endpoints:
  POST /api/v1/chat       - Send chat prompts
  GET  /api/v1/health     - Health check
  GET  /api/v1/tools      - List available tools
```

## Step 5: Test the Service

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Expected response:

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

### List Available Tools

```bash
curl http://localhost:8081/api/v1/tools
```

### Send a Chat Request

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Show me all available blueprints"
  }'
```

## Step 6: Use with Frontend

Now you can integrate this backend with any frontend application!

**Example JavaScript:**

```javascript
async function chat(prompt) {
  const response = await fetch("http://localhost:8081/api/v1/chat", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ prompt }),
  });

  const data = await response.json();
  return data;
}

// Usage
const result = await chat("Create a PostgreSQL database for production");
console.log(result.response);
```

## Using Makefile (Alternative)

For easier management, you can use the provided Makefile:

```bash
# Initialize the project
make init

# Edit .env file with your configuration
# Then:

# Install dependencies
make install

# Build
make build

# Run
make run

# Or run in development mode
make dev
```

## Troubleshooting

### "MCP_SERVER_PATH is required"

Make sure your `.env` file has the correct path to the MCP server binary:

```bash
# Find the MCP server
ls -la /Users/deepakbansode/Documents/projects/idp-cloudgenie/idp-cloudgenie-mcp-server/cloudgenie-mcp-server

# Update .env with the correct path
```

### "OPENAI_API_KEY is required"

Get an API key from https://platform.openai.com/ and add it to your `.env` file.

### "Failed to create MCP client"

- Ensure the MCP server binary exists and has execute permissions:
  ```bash
  chmod +x /path/to/cloudgenie-mcp-server
  ```
- Check that the path in `.env` is absolute, not relative

### MCP Server Can't Connect to CloudGenie API

Make sure the CloudGenie backend API is running on the configured URL (default: http://localhost:8080).

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Check [API.md](API.md) for complete API documentation
- Explore the code in the `internal/` directory
- Add authentication and rate limiting for production use
- Deploy to your cloud provider

## Quick Commands Reference

```bash
# Build
make build

# Run
make run

# Development mode
make dev

# Run tests
make test

# Clean build artifacts
make clean

# View all commands
make help
```

## Getting Help

If you encounter issues:

1. Check the server logs for detailed error messages
2. Verify all prerequisites are met
3. Ensure environment variables are set correctly
4. Check that the CloudGenie API is accessible
5. Open an issue on GitHub with detailed information

Happy coding! 🚀
