package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const DEFAULT_FILES_DIR = "./files"

func main() {
	// Create MCP server with sampling capability
	mcpServer := server.NewMCPServer("enhanced-sampling-server", "1.0.0")

	// Enable sampling capability
	mcpServer.EnableSampling()

	// Ensure files directory exists
	if err := os.MkdirAll(DEFAULT_FILES_DIR, 0755); err != nil {
		log.Printf("Warning: Could not create files directory: %v", err)
	}

	// Add tool to analyze a file using LLM sampling
	mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_file",
		Description: "Analyze a file from the local directory using LLM sampling",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"filename": map[string]any{
					"type":        "string",
					"description": "The name of the file to analyze (relative to files directory)",
				},
				"analysis_type": map[string]any{
					"type":        "string",
					"description": "Type of analysis to perform",
					"enum":        []string{"summarize", "explain", "analyze", "extract_key_points"},
				},
				"custom_prompt": map[string]any{
					"type":        "string",
					"description": "Optional custom prompt for the analysis",
				},
			},
			Required: []string{"filename"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		filename, err := request.RequireString("filename")
		if err != nil {
			return nil, err
		}

		analysisType := request.GetString("analysis_type", "summarize")
		customPrompt := request.GetString("custom_prompt", "")

		// Construct file path
		filePath := filepath.Join(DEFAULT_FILES_DIR, filename)
		
		// Security check - ensure file is within the files directory
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error resolving file path: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		absDirPath, err := filepath.Abs(DEFAULT_FILES_DIR)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error resolving directory path: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		if !strings.HasPrefix(absFilePath, absDirPath) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Access denied: File must be within the files directory",
					},
				},
				IsError: true,
			}, nil
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("File not found: %s", filename),
					},
				},
				IsError: true,
			}, nil
		}

		// Read file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error reading file: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		// Determine file type
		ext := strings.ToLower(filepath.Ext(filename))
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Prepare content for LLM based on file type
		var contentForLLM mcp.Content
		var systemPrompt string

		// Create appropriate prompt based on analysis type
		var basePrompt string
		switch analysisType {
		case "summarize":
			basePrompt = "Please provide a clear and concise summary of this content."
		case "explain":
			basePrompt = "Please explain what this content is about and its main purpose."
		case "analyze":
			basePrompt = "Please provide a detailed analysis of this content, including its structure, key components, and any notable patterns."
		case "extract_key_points":
			basePrompt = "Please extract the key points and main ideas from this content."
		default:
			basePrompt = "Please analyze this content and provide insights."
		}

		if customPrompt != "" {
			basePrompt = customPrompt
		}

		if strings.HasPrefix(mimeType, "text/") || ext == ".md" || ext == ".txt" || ext == ".json" || ext == ".xml" || ext == ".csv" {
			// Text file - send as text content
			contentForLLM = mcp.TextContent{
				Type: "text",
				Text: string(fileContent),
			}
			systemPrompt = fmt.Sprintf("%s The content is a %s file named '%s'.", basePrompt, mimeType, filename)
		} else if strings.HasPrefix(mimeType, "image/") {
			// Image file - send as base64 encoded image
			base64Content := base64.StdEncoding.EncodeToString(fileContent)
			contentForLLM = mcp.ImageContent{
				Type: "image",
				Data: base64Content,
				MIMEType: mimeType,
			}
			systemPrompt = fmt.Sprintf("%s The content is an image file named '%s' of type %s.", basePrompt, filename, mimeType)
		} else {
			// Binary file - send as base64 with description
			base64Content := base64.StdEncoding.EncodeToString(fileContent)
			contentForLLM = mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("This is a binary file (%s) encoded in base64:\n\n%s", mimeType, base64Content),
			}
			systemPrompt = fmt.Sprintf("%s The content is a binary file named '%s' of type %s, provided as base64-encoded data.", basePrompt, filename, mimeType)
		}

		// Create sampling request
		samplingRequest := mcp.CreateMessageRequest{
			CreateMessageParams: mcp.CreateMessageParams{
				Messages: []mcp.SamplingMessage{
					{
						Role:    mcp.RoleUser,
						Content: contentForLLM,
					},
				},
				SystemPrompt: systemPrompt,
				MaxTokens:    2000,
				Temperature:  0.3, // Lower temperature for more focused analysis
			},
		}

		// Request sampling from the client with timeout
		log.Printf("📤 Sending sampling request for file: %s (analysis: %s)", filename, analysisType)
		samplingCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		serverFromCtx := server.ServerFromContext(ctx)
		result, err := serverFromCtx.RequestSampling(samplingCtx, samplingRequest)
		if err != nil {
			log.Printf("❌ Sampling request failed: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error requesting sampling: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		log.Printf("✅ Sampling request successful! Model: %s", result.Model)
		
		// Extract response text safely
		var responseText string
		if textContent, ok := result.Content.(mcp.TextContent); ok {
			responseText = textContent.Text
		} else {
			responseText = fmt.Sprintf("%v", result.Content)
		}

		// Return the analysis result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("File Analysis Results\n" +
						"=====================\n" +
						"File: %s\n" +
						"Type: %s\n" +
						"Analysis: %s\n" +
						"Model: %s\n\n" +
						"%s", filename, mimeType, analysisType, result.Model, responseText),
				},
			},
		}, nil
	})

	// Add tool to list available files
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_files",
		Description: "List all files available for analysis in the files directory",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entries, err := os.ReadDir(DEFAULT_FILES_DIR)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error reading files directory: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		var fileList []string
		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				size := info.Size()
				mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(entry.Name())))
				if mimeType == "" {
					mimeType = "application/octet-stream"
				}
				fileList = append(fileList, fmt.Sprintf("- %s (%d bytes, %s)", entry.Name(), size, mimeType))
			}
		}

		if len(fileList) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("No files found in %s directory", DEFAULT_FILES_DIR),
					},
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Available files in %s:\n\n%s", DEFAULT_FILES_DIR, strings.Join(fileList, "\n")),
				},
			},
		}, nil
	})

	// Add the original echo tool for testing
	mcpServer.AddTool(mcp.Tool{
		Name:        "echo",
		Description: "Echo back the input message (no sampling required)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "The message to echo back",
				},
			},
			Required: []string{"message"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		message := request.GetString("message", "")

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Echo: %s", message),
				},
			},
		}, nil
	})

	// Create HTTP server
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	log.Println("Starting Enhanced HTTP MCP Server with File Analysis on :8080")
	log.Println("Endpoint: http://localhost:8080/mcp")
	log.Printf("Files directory: %s", DEFAULT_FILES_DIR)
	log.Println("")
	log.Println("This server supports file analysis using LLM sampling over HTTP transport.")
	log.Println("")
	log.Println("Available tools:")
	log.Println("- analyze_file: Analyze files using LLM sampling (text, images, PDFs)")
	log.Println("- list_files: List available files for analysis")
	log.Println("- echo: Simple echo tool (no sampling required)")
	log.Println("")
	log.Println("To test:")
	log.Printf("1. Place files to analyze in the %s directory", DEFAULT_FILES_DIR)
	log.Println("2. Start the enhanced client with your Anthropic API key")
	log.Println("3. The client will connect and handle sampling requests")

	// Start the server
	if err := httpServer.Start(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}