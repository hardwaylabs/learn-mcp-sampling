package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	fmt.Println("Testing Basic Sampling (Original Examples)")
	fmt.Println("==========================================")
	fmt.Println("This tests the original ask_llm tool from sampling_http_server")
	fmt.Println("Make sure sampling_http_server and sampling_http_client are running!")
	fmt.Println("")

	// Create HTTP transport
	httpTransport, err := transport.NewStreamableHTTP("http://localhost:8080/mcp")
	if err != nil {
		log.Fatalf("Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()

	// Create client (without sampling handler - this is just to call tools)
	mcpClient := client.NewClient(httpTransport)

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
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "basic-sampling-test-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	fmt.Printf("✅ Connected to: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)

	// List available tools
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Println("Available tools:")
	for _, tool := range toolsResult.Tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}

	// Test the ask_llm tool (should be available if basic server is running)
	fmt.Println("\nTesting ask_llm tool...")
	
	askCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := mcpClient.CallTool(askCtx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "ask_llm",
			Arguments: map[string]any{
				"question": "What is the capital of France?",
			},
		},
	})

	if err != nil {
		fmt.Printf("❌ ask_llm tool failed: %v\n", err)
		fmt.Println("This means the basic sampling workflow is also broken")
	} else {
		fmt.Println("✅ ask_llm tool works!")
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				fmt.Printf("Response: %s\n", textContent.Text)
			}
		}
	}
}