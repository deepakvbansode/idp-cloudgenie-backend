# Using Google Gemini with CloudGenie Backend

## Overview

The CloudGenie Backend Service now supports **Google Gemini AI models** in addition to OpenAI and Anthropic. Gemini provides powerful language understanding and generation capabilities for infrastructure management tasks.

## Supported Gemini Models

- `gemini-1.5-pro` (default) - Best for complex reasoning and multi-turn conversations
- `gemini-1.5-flash` - Faster responses for simpler tasks
- `gemini-pro` - Previous generation, still capable

## Setup

### 1. Get Your Gemini API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Get API Key" or "Create API Key"
4. Copy your API key

### 2. Configure the Backend

Edit your `.env` file:

```env
# AI Provider Configuration
DEFAULT_AI_PROVIDER=gemini

# Google Gemini Configuration
GEMINI_API_KEY=your-gemini-api-key-here
GEMINI_MODEL=gemini-1.5-pro

# Other configuration...
MCP_SERVER_PATH=/path/to/cloudgenie-mcp-server
CLOUDGENIE_BACKEND_URL=http://localhost:8080
```

### 3. Run the Service

```bash
# Build
go build -o cloudgenie-backend .

# Run
./cloudgenie-backend
```

You should see:

```
Initializing AI provider: gemini
...
Server starting on 0.0.0.0:8081
```

## Usage Examples

### Example 1: List Blueprints

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Show me all available infrastructure blueprints",
    "provider": "gemini"
  }'
```

### Example 2: Create Resources

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a PostgreSQL database for my production environment",
    "provider": "gemini",
    "model": "gemini-1.5-pro"
  }'
```

### Example 3: Check Resource Status

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is the status of all my cloud resources?",
    "provider": "gemini"
  }'
```

## API Request Format

When using Gemini, you can optionally specify the provider and model in your request:

```json
{
  "prompt": "Your infrastructure management request",
  "provider": "gemini",
  "model": "gemini-1.5-pro",
  "context": {}
}
```

If `provider` is not specified, the `DEFAULT_AI_PROVIDER` from `.env` will be used.

## Features

### ✅ Supported

- Natural language understanding
- Multi-turn conversations
- Context-aware responses
- Integration with CloudGenie MCP tools
- Conversation history

### ⚠️ Current Limitations

- Tool/function calling not yet implemented (basic text responses only)
- Token usage metrics may not be available
- Streaming responses not supported yet

### 🔜 Coming Soon

- Full function calling support for automatic tool execution
- Enhanced tool discovery and usage
- Streaming API support
- Better token usage tracking

## Comparison with Other Providers

| Feature              | OpenAI  | Anthropic | Gemini     |
| -------------------- | ------- | --------- | ---------- |
| Tool Calling         | ✅ Full | ⚠️ Stub   | ⚠️ Planned |
| Conversation History | ✅      | ✅        | ✅         |
| Streaming            | ✅      | ❌        | ❌         |
| Token Metrics        | ✅      | ✅        | ⚠️ Limited |
| Cost                 | $$$     | $$$       | $          |

## Switching Between Providers

You can easily switch between AI providers by changing the environment variable:

### Use OpenAI

```env
DEFAULT_AI_PROVIDER=openai
OPENAI_API_KEY=sk-...
```

### Use Gemini

```env
DEFAULT_AI_PROVIDER=gemini
GEMINI_API_KEY=AIza...
```

### Use Anthropic (when available)

```env
DEFAULT_AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...
```

Or specify per-request:

```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Your query",
    "provider": "gemini"
  }'
```

## Pricing

Google Gemini offers competitive pricing:

- **Gemini 1.5 Pro**: $0.00125 per 1K characters (input), $0.00375 per 1K characters (output)
- **Gemini 1.5 Flash**: $0.00001875 per 1K characters (input), $0.00005625 per 1K characters (output)

Free tier includes:

- 15 requests per minute
- 1 million tokens per minute
- 1,500 requests per day

See [Google AI Pricing](https://ai.google.dev/pricing) for current rates.

## Troubleshooting

### Error: "Gemini API key is required"

Make sure your `.env` file has:

```env
GEMINI_API_KEY=your-actual-key-here
```

### Error: "Failed to create Gemini client"

1. Check your API key is valid
2. Ensure you have internet connectivity
3. Verify the API key has not been restricted or revoked

### Slow Responses

- Try using `gemini-1.5-flash` for faster responses
- Check your internet connection
- Verify you're not hitting rate limits

### Different Results from OpenAI

Gemini and OpenAI have different architectures and training data, so responses may vary. Both are capable, but may approach problems differently.

## Best Practices

1. **Start with gemini-1.5-pro** for complex infrastructure tasks
2. **Use gemini-1.5-flash** for simple queries or when speed matters
3. **Monitor API usage** through Google Cloud Console
4. **Test prompts** - Gemini may respond differently than OpenAI
5. **Provide context** - Include relevant details in your prompts

## Environment Variables Reference

```env
# Required for Gemini
DEFAULT_AI_PROVIDER=gemini
GEMINI_API_KEY=your-api-key

# Optional
GEMINI_MODEL=gemini-1.5-pro  # or gemini-1.5-flash
```

## Example .env File

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8081

# AI Provider - Using Gemini
DEFAULT_AI_PROVIDER=gemini
GEMINI_API_KEY=AIzaSyD...your...key...here
GEMINI_MODEL=gemini-1.5-pro

# MCP Server
MCP_SERVER_PATH=/Users/yourusername/cloudgenie-mcp-server/cloudgenie-mcp-server
CLOUDGENIE_BACKEND_URL=http://localhost:8080

# CORS
ALLOWED_ORIGINS=*
```

## Support

For Gemini-specific issues:

- [Google AI Documentation](https://ai.google.dev/docs)
- [API Reference](https://ai.google.dev/api)
- [Community Forum](https://discuss.ai.google.dev/)

For CloudGenie Backend issues:

- Check the main [README.md](README.md)
- Open an issue on GitHub

---

**Happy infrastructure management with Gemini! 🚀**
