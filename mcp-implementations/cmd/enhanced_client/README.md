# Enhanced MCP Client with Anthropic API Integration

This enhanced MCP client demonstrates real LLM integration by connecting to Anthropic's Claude API to handle sampling requests from MCP servers.

## Features

- **Real LLM Integration**: Uses Anthropic's Claude 3.5 Sonnet model
- **Multi-Modal Support**: Handles text, image, and binary file analysis
- **Token Usage Tracking**: Reports input/output token consumption
- **Proper Error Handling**: Comprehensive error handling and logging
- **Production Ready**: Includes timeouts, proper headers, and API best practices

## Prerequisites

1. **Anthropic API Key**: Get an API key from [Anthropic Console](https://console.anthropic.com/)
2. **Environment Variable**: Set `ANTHROPIC_API_KEY` environment variable
3. **Go Dependencies**: Run `go mod tidy` to install dependencies

## Usage

1. **Set API Key**:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

2. **Start the Client**:
   ```bash
   go run cmd/enhanced_client/main.go
   ```

3. **Connect to Server**: The client will connect to the MCP server at `http://localhost:8080/mcp`

## How It Works

### Sampling Handler Implementation
The `AnthropicSamplingHandler` implements the `client.SamplingHandler` interface:

1. **Receives MCP Request**: Gets sampling request from the MCP server
2. **Converts Format**: Transforms MCP message format to Anthropic API format
3. **Handles Content Types**:
   - Text content: Sent as text blocks
   - Image content: Sent as base64-encoded image blocks
   - Binary content: Described and sent as base64 text
4. **Calls Anthropic API**: Makes HTTP request to Claude API
5. **Returns Results**: Converts response back to MCP format

### Content Type Handling

- **Text Files**: Sent directly as text content
- **Images**: Converted to base64 and sent as image blocks with proper MIME types
- **Binary Files**: Encoded as base64 with descriptive context

### API Configuration

- **Model**: Claude 3.5 Sonnet (latest available)
- **Temperature**: 0.3 (focused analysis)
- **Max Tokens**: 2000 (configurable per request)
- **Timeout**: 2 minutes per request

## Real-World Usage

This client emulates how real MCP clients like Claude Desktop, Claude Code, or VS Code extensions would integrate with LLM services:

1. **Bidirectional Communication**: Maintains persistent connection to MCP server
2. **Sampling Request Handling**: Processes incoming sampling requests asynchronously  
3. **LLM Integration**: Connects to actual language model API
4. **Response Processing**: Formats and returns LLM responses appropriately

## Integration Example

For production use, you might integrate this pattern into existing applications:

```go
// Create sampling handler with your preferred LLM service
samplingHandler := NewCustomLLMHandler(apiKey, modelConfig)

// Create MCP client with sampling support
mcpClient := client.NewClient(
    transport,
    client.WithSamplingHandler(samplingHandler),
)
```

## Monitoring and Logging

The client provides comprehensive logging:
- Request/response timing
- Token usage statistics
- Error conditions and recovery
- Model and API version information

This enables monitoring of LLM usage and costs in production environments.