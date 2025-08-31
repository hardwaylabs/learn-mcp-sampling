package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	fmt.Println("MCP Sampling Workflow Test")
	fmt.Println("==========================")
	fmt.Println("This test assumes:")
	fmt.Println("1. Enhanced server is running on :8080")
	fmt.Println("2. Enhanced client is running with ANTHROPIC_API_KEY")
	fmt.Println()

	// Create HTTP transport
	httpTransport, err := transport.NewStreamableHTTP(
		"http://localhost:8080/mcp",
	)
	if err != nil {
		log.Fatalf("Failed to create HTTP transport: %v", err)
	}
	defer httpTransport.Close()

	// Create client (without sampling handler - this client makes requests)
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
				Name:    "workflow-test-client",
				Version: "1.0.0",
			},
		},
	}

	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP session: %v", err)
	}

	fmt.Println("✓ Connected to MCP server")

	// Test 1: Echo tool (no sampling)
	fmt.Println("\n1. Testing echo tool (no sampling)...")
	echoResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]any{
				"message": "Hello from workflow test!",
			},
		},
	})

	if err != nil {
		fmt.Printf("✗ Echo test failed: %v\n", err)
	} else {
		if textContent, ok := echoResult.Content[0].(mcp.TextContent); ok {
			fmt.Printf("✓ Echo test passed: %s\n", textContent.Text)
		}
	}

	// Test 2: List files
	fmt.Println("\n2. Listing available files...")
	listResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_files",
			Arguments: map[string]any{},
		},
	})

	if err != nil {
		fmt.Printf("✗ List files failed: %v\n", err)
	} else {
		if textContent, ok := listResult.Content[0].(mcp.TextContent); ok {
			fmt.Printf("✓ Available files:\n%s\n", textContent.Text)
		}
	}

	// Test 3: File analysis (requires sampling)
	fmt.Println("\n3. Testing file analysis with sampling...")
	fmt.Println("   This will only work if the enhanced client with Anthropic API is running!")

	// Create a context with timeout for the sampling operation
	analysisCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	analysisResult, err := mcpClient.CallTool(analysisCtx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "analyze_file",
			Arguments: map[string]any{
				"filename":      "sample_document.md",
				"analysis_type": "summarize",
			},
		},
	})

	if err != nil {
		fmt.Printf("✗ File analysis failed: %v\n", err)
		fmt.Println("   This likely means:")
		fmt.Println("   - Enhanced client is not running, OR")
		fmt.Println("   - Enhanced client doesn't have valid ANTHROPIC_API_KEY, OR")
		fmt.Println("   - There's a network/connection issue")
	} else {
		if textContent, ok := analysisResult.Content[0].(mcp.TextContent); ok {
			fmt.Printf("✓ File analysis successful!\n")
			fmt.Printf("Analysis result:\n%s\n", textContent.Text)
		}
	}

	// Test 4: Custom prompt analysis
	if err == nil { // Only test if previous analysis worked
		fmt.Println("\n4. Testing custom prompt analysis...")
		
		customCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()

		customResult, err := mcpClient.CallTool(customCtx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "analyze_file",
				Arguments: map[string]any{
					"filename":      "code_example.py",
					"custom_prompt": "What machine learning techniques are demonstrated in this code? List the key components.",
				},
			},
		})

		if err != nil {
			fmt.Printf("✗ Custom prompt analysis failed: %v\n", err)
		} else {
			if textContent, ok := customResult.Content[0].(mcp.TextContent); ok {
				fmt.Printf("✓ Custom prompt analysis successful!\n")
				fmt.Printf("Analysis result:\n%s\n", textContent.Text)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Workflow test completed!")
	fmt.Println("If all tests passed, your enhanced MCP sampling setup is working correctly.")
	fmt.Println("If the sampling tests failed, ensure the enhanced client is running with a valid API key.")
}