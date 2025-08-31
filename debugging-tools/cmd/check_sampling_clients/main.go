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
	fmt.Println("Checking for sampling-capable clients connected to the server...")
	
	// Create HTTP transport
	httpTransport, err := transport.NewStreamableHTTP("http://localhost:8080/mcp")
	if err != nil {
		log.Fatalf("Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()

	// Create client
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
				Name:    "sampling-check-client",
				Version: "1.0.0",
			},
		},
	}

	initResponse, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	fmt.Printf("Connected to: %s v%s\n", initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)
	
	// Check server capabilities
	if initResponse.Capabilities.Sampling != nil {
		fmt.Println("✅ Server has sampling capability")
	} else {
		fmt.Println("❌ Server does not have sampling capability")
		return
	}

	// Try calling analyze_file with a short timeout to see what happens
	fmt.Println("\nTesting sampling request with 10 second timeout...")
	
	analysisCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = mcpClient.CallTool(analysisCtx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "analyze_file",
			Arguments: map[string]any{
				"filename":      "sample_document.md",
				"analysis_type": "summarize",
			},
		},
	})

	if err != nil {
		if analysisCtx.Err() == context.DeadlineExceeded {
			fmt.Println("❌ Sampling request timed out - no enhanced client is handling sampling requests")
			fmt.Println("\n💡 To fix this, run in another terminal:")
			fmt.Println("export ANTHROPIC_API_KEY=\"your-key\"")
			fmt.Println("go run cmd/enhanced_client/main.go")
		} else {
			fmt.Printf("❌ Sampling request failed with error: %v\n", err)
		}
	} else {
		fmt.Println("✅ Sampling request succeeded - enhanced client is working!")
	}
}