package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// Check if API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable is not set")
		fmt.Println("Please set it with: export ANTHROPIC_API_KEY=\"your-api-key\"")
		return
	}
	fmt.Println("✅ ANTHROPIC_API_KEY is set")

	// Test connection to server
	httpTransport, err := transport.NewStreamableHTTP("http://localhost:8080/mcp")
	if err != nil {
		log.Fatalf("❌ Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()
	fmt.Println("✅ HTTP transport created")

	// Create client without sampling handler first (just to test connection)
	mcpClient := client.NewClient(httpTransport)

	// Start the client
	ctx := context.Background()
	err = mcpClient.Start(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}
	fmt.Println("✅ Client started")

	// Initialize the MCP session
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "connection-test-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("❌ Failed to initialize MCP session: %v", err)
	}
	
	fmt.Printf("✅ Connected to server: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)
	
	// Test echo tool (no sampling)
	_, err = mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]any{
				"message": "Connection test",
			},
		},
	})
	if err != nil {
		fmt.Printf("❌ Echo tool failed: %v\n", err)
	} else {
		fmt.Println("✅ Echo tool works")
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Connection test completed!")
	fmt.Println("If all tests passed, the connection is working.")
	fmt.Println("Now you need to start the enhanced_client to handle sampling requests.")
	fmt.Println("\nRun in another terminal:")
	fmt.Println("go run cmd/enhanced_client/main.go")
}