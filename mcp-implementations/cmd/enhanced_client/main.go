package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// AnthropicSamplingHandler implements client.SamplingHandler using the Anthropic API
type AnthropicSamplingHandler struct {
	APIKey     string
	HTTPClient *http.Client
}

// AnthropicRequest represents the structure for Anthropic API requests
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

type Content interface{}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type   string `json:"type"`
	Source Source `json:"source"`
}

type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// AnthropicResponse represents the structure for Anthropic API responses
type AnthropicResponse struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Role         string                   `json:"role"`
	Content      []AnthropicTextContent   `json:"content"`
	Model        string                   `json:"model"`
	StopReason   string                   `json:"stop_reason"`
	StopSequence string                   `json:"stop_sequence"`
	Usage        AnthropicUsage           `json:"usage"`
}

type AnthropicTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func NewAnthropicSamplingHandler(apiKey string) *AnthropicSamplingHandler {
	return &AnthropicSamplingHandler{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (h *AnthropicSamplingHandler) CreateMessage(ctx context.Context, request mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	log.Printf("📨 Received sampling request with %d messages", len(request.Messages))
	
	if len(request.Messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// Convert MCP messages to Anthropic format
	var messages []Message
	for _, mcpMsg := range request.Messages {
		var content Content

		switch mcpContent := mcpMsg.Content.(type) {
		case mcp.TextContent:
			content = []TextContent{{
				Type: "text",
				Text: mcpContent.Text,
			}}
		case mcp.ImageContent:
			// For image content, create image block
			content = []interface{}{
				ImageContent{
					Type: "image",
					Source: Source{
						Type:      "base64",
						MediaType: mcpContent.MIMEType,
						Data:      mcpContent.Data,
					},
				},
			}
		default:
			// Fallback to text
			content = []TextContent{{
				Type: "text",
				Text: fmt.Sprintf("%v", mcpContent),
			}}
		}

		role := "user"
		if mcpMsg.Role == mcp.RoleAssistant {
			role = "assistant"
		}

		messages = append(messages, Message{
			Role:    role,
			Content: content,
		})
	}

	// Create Anthropic API request
	anthropicReq := AnthropicRequest{
		Model:       "claude-3-5-sonnet-20241022", // Use latest Sonnet model
		MaxTokens:   request.MaxTokens,
		Messages:    messages,
		System:      request.SystemPrompt,
		Temperature: request.Temperature,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("Sending request to Anthropic API (model: %s, tokens: %d)", anthropicReq.Model, anthropicReq.MaxTokens)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", h.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := h.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract text content
	var responseText string
	if len(anthropicResp.Content) > 0 {
		responseText = anthropicResp.Content[0].Text
	}

	log.Printf("Received response from Anthropic API (model: %s, input tokens: %d, output tokens: %d)", 
		anthropicResp.Model, anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)

	// Convert back to MCP format
	result := &mcp.CreateMessageResult{
		SamplingMessage: mcp.SamplingMessage{
			Role: mcp.RoleAssistant,
			Content: mcp.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
		Model:      anthropicResp.Model,
		StopReason: anthropicResp.StopReason,
	}

	return result, nil
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create sampling handler with Anthropic API integration
	samplingHandler := NewAnthropicSamplingHandler(apiKey)

	// Create HTTP transport with continuous listening for sampling
	httpTransport, err := transport.NewStreamableHTTP(
		"http://localhost:8080/mcp",
		transport.WithContinuousListening(),
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()

	// Create client with sampling support
	mcpClient := client.NewClient(
		httpTransport,
		client.WithSamplingHandler(samplingHandler),
	)

	// Start the client
	ctx := context.Background()
	err = mcpClient.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	// Initialize the MCP session
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities: mcp.ClientCapabilities{
				// Sampling capability will be automatically added by the client
			},
			ClientInfo: mcp.Implementation{
				Name:    "enhanced-anthropic-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	log.Println("✅ Enhanced HTTP MCP Client with Anthropic API integration started successfully!")
	log.Println("")
	log.Printf("🔗 Connected to MCP Server: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)
	log.Println("🤖 Connected to Anthropic API (Claude 3.5 Sonnet)")
	log.Println("📡 Continuous listening enabled for server notifications")
	log.Println("")
	log.Println("Features:")
	log.Println("- Supports text, image, and binary file analysis")
	log.Println("- Handles sampling requests from MCP server")
	log.Println("- Real LLM processing with token usage tracking")
	log.Println("- Long-lived connection for server-to-client notifications")
	log.Println("")
	log.Println("The client is now ready to:")
	log.Println("1. Receive file content from the MCP server")
	log.Println("2. Send it to Claude for analysis/summarization") 
	log.Println("3. Return the results back to the server")
	log.Println("")
	log.Println("🎧 Waiting for sampling requests from the server...")
	log.Println("💡 You can now run 'go run test_workflow.go' in another terminal")

	// Keep the client running
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Client context cancelled")
	case <-sigChan:
		log.Println("Received shutdown signal")
	}

	log.Println("Shutting down client...")
}