package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Use the same AnthropicSamplingHandler from enhanced_client
// (Copy-pasted to avoid import issues)

func main() {
	fmt.Println("All-in-One MCP Sampling Test")
	fmt.Println("============================")
	fmt.Println("This test combines both sampling handler AND tool calls in one client")
	fmt.Println("")

	// Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable is required")
		fmt.Println("Run: export ANTHROPIC_API_KEY=\"your-key\"")
		return
	}
	fmt.Println("✅ ANTHROPIC_API_KEY is set")

	// Create sampling handler
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

	// Create client with BOTH sampling handler AND tool calling capability
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
				Name:    "all-in-one-test-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	fmt.Printf("✅ Connected to: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)

	// Test 1: Echo (no sampling)
	fmt.Println("\n1. Testing echo tool...")
	_, err = mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]any{
				"message": "All-in-one test",
			},
		},
	})
	if err != nil {
		fmt.Printf("❌ Echo failed: %v\n", err)
	} else {
		fmt.Println("✅ Echo works")
	}

	// Test 2: File analysis (requires sampling)
	fmt.Println("\n2. Testing file analysis with sampling...")
	fmt.Println("   This should work since we have both sampling handler and tool calling in same session!")

	analysisCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	result, err := mcpClient.CallTool(analysisCtx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "analyze_file",
			Arguments: map[string]any{
				"filename":      "sample_document.md",
				"analysis_type": "summarize",
			},
		},
	})

	if err != nil {
		fmt.Printf("❌ File analysis failed: %v\n", err)
	} else {
		fmt.Println("✅ File analysis successful!")
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				// Truncate long responses for display
				text := textContent.Text
				if len(text) > 500 {
					text = text[:500] + "..."
				}
				fmt.Printf("📄 Analysis result:\n%s\n", text)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	if err == nil {
		fmt.Println("🎉 SUCCESS: All-in-one test completed successfully!")
		fmt.Println("This proves the sampling workflow works when client and handler are in same session.")
	} else {
		fmt.Println("❌ Test failed - check server logs for details")
	}
}

// Simplified version of AnthropicSamplingHandler for this test
type SimpleAnthropicSamplingHandler struct {
	APIKey string
}

func NewAnthropicSamplingHandler(apiKey string) *SimpleAnthropicSamplingHandler {
	return &SimpleAnthropicSamplingHandler{APIKey: apiKey}
}

func (h *SimpleAnthropicSamplingHandler) CreateMessage(ctx context.Context, request mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	log.Printf("📨 All-in-one client received sampling request!")
	
	// For this test, return a simple mock response to prove the flow works
	// In real usage, you'd call the Anthropic API here
	
	responseText := "MOCK RESPONSE: This is a summary of the requested file. The sampling workflow is working correctly!"
	
	result := &mcp.CreateMessageResult{
		SamplingMessage: mcp.SamplingMessage{
			Role: mcp.RoleAssistant,
			Content: mcp.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
		Model:      "mock-test-model",
		StopReason: "endTurn",
	}

	log.Printf("📤 All-in-one client sending response back to server")
	return result, nil
}