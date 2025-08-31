package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// This simulates the sampling workflow by calling the Anthropic API directly
// instead of going through the broken HTTP sampling transport

func main() {
	fmt.Println("MCP Sampling Workflow Simulation")
	fmt.Println("================================")
	fmt.Println("Since HTTP sampling is broken in mcp-go, this simulates the workflow")
	fmt.Println("")

	// Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable is required")
		return
	}
	fmt.Println("✅ ANTHROPIC_API_KEY is set")

	// Connect to enhanced server (for file operations)
	httpTransport, err := transport.NewStreamableHTTP("http://localhost:8080/mcp")
	if err != nil {
		log.Fatalf("Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()

	mcpClient := client.NewClient(httpTransport)
	ctx := context.Background()
	err = mcpClient.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "sampling-simulation-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	fmt.Printf("✅ Connected to: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)

	// Step 1: List files (this works)
	fmt.Println("\n📁 Step 1: Listing available files...")
	listResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_files",
			Arguments: map[string]any{},
		},
	})

	if err != nil {
		fmt.Printf("❌ Failed to list files: %v\n", err)
		return
	}

	if textContent, ok := listResult.Content[0].(mcp.TextContent); ok {
		fmt.Printf("%s\n", textContent.Text)
	}

	// Step 2: Manually get file content and simulate the sampling workflow
	fmt.Println("\n🔍 Step 2: Simulating the sampling workflow...")
	
	filename := "sample_document.md"
	analysisType := "summarize"
	
	fmt.Printf("📖 Reading file: %s\n", filename)
	
	// Simulate reading the file (in real implementation, server would do this)
	content, err := os.ReadFile("files/" + filename)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %v\n", err)
		return
	}
	
	fmt.Printf("📄 File content (%d bytes): %s...\n", len(content), string(content)[:100])
	
	// Step 3: Call Anthropic API directly (simulating what the sampling handler would do)
	fmt.Println("\n🤖 Step 3: Calling Anthropic API (simulating sampling)...")
	
	handler := NewAnthropicSamplingHandler(apiKey)
	
	// Create the sampling request that would be sent
	samplingRequest := mcp.CreateMessageRequest{
		CreateMessageParams: mcp.CreateMessageParams{
			Messages: []mcp.SamplingMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: string(content),
					},
				},
			},
			SystemPrompt: fmt.Sprintf("Please %s this content. The content is a markdown file named '%s'.", analysisType, filename),
			MaxTokens:    2000,
			Temperature:  0.3,
		},
	}
	
	// Call the handler directly (simulating the sampling)
	result, err := handler.CreateMessage(ctx, samplingRequest)
	if err != nil {
		fmt.Printf("❌ Anthropic API call failed: %v\n", err)
		return
	}
	
	fmt.Println("✅ Analysis completed!")
	
	// Step 4: Display results
	fmt.Println("\n📋 Step 4: Analysis Results")
	fmt.Println(strings.Repeat("=", 50))
	
	if textContent, ok := result.Content.(mcp.TextContent); ok {
		fmt.Printf("File: %s\n", filename)
		fmt.Printf("Analysis: %s\n", analysisType) 
		fmt.Printf("Model: %s\n", result.Model)
		fmt.Printf("\nResult:\n%s\n", textContent.Text)
	}
	
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("🎉 Sampling workflow simulation completed successfully!")
	fmt.Println("")
	fmt.Println("This demonstrates what SHOULD happen:")
	fmt.Println("1. ✅ Client connects to MCP server")
	fmt.Println("2. ✅ Client lists available files") 
	fmt.Println("3. ✅ Server reads file content")
	fmt.Println("4. ✅ Server sends content to LLM via sampling")
	fmt.Println("5. ✅ LLM analyzes content and returns result")
	fmt.Println("6. ✅ Result is returned to client")
	fmt.Println("")
	fmt.Println("The only broken part is step 4 (HTTP sampling transport)")
}

// Simplified Anthropic handler for simulation
type AnthropicSamplingHandler struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewAnthropicSamplingHandler(apiKey string) *AnthropicSamplingHandler {
	return &AnthropicSamplingHandler{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

type AnthropicRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []AnthropicContent `json:"content"`
	Model   string             `json:"model"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (h *AnthropicSamplingHandler) CreateMessage(ctx context.Context, request mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	// Convert MCP to Anthropic format
	var messages []Message
	for _, mcpMsg := range request.Messages {
		var content string
		if textContent, ok := mcpMsg.Content.(mcp.TextContent); ok {
			content = textContent.Text
		} else {
			content = fmt.Sprintf("%v", mcpMsg.Content)
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

	anthropicReq := AnthropicRequest{
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   request.MaxTokens,
		Messages:    messages,
		System:      request.SystemPrompt,
		Temperature: request.Temperature,
	}

	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	fmt.Printf("🔄 Calling Anthropic API...\n")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", h.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := h.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	var responseText string
	if len(anthropicResp.Content) > 0 {
		responseText = anthropicResp.Content[0].Text
	}

	return &mcp.CreateMessageResult{
		SamplingMessage: mcp.SamplingMessage{
			Role: mcp.RoleAssistant,
			Content: mcp.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
		Model:      anthropicResp.Model,
		StopReason: "endTurn",
	}, nil
}