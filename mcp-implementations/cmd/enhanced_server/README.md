# Enhanced MCP Server with File Analysis

This enhanced MCP server demonstrates real-world sampling capabilities by providing file analysis services through LLM integration.

## Features

- **File Analysis**: Analyze text, images, and binary files using LLM sampling
- **Multiple Analysis Types**: Summarize, explain, analyze, or extract key points
- **Secure File Access**: Files must be within the designated `files/` directory
- **Support for Various File Types**:
  - Text files (`.txt`, `.md`, `.json`, `.xml`, `.csv`)
  - Images (`.jpg`, `.png`, `.gif`, etc.)
  - Binary files (PDFs, documents, etc.)

## Tools Available

### `analyze_file`
Analyzes a file using LLM sampling with the following parameters:
- `filename` (required): Name of the file to analyze
- `analysis_type` (optional): Type of analysis - "summarize", "explain", "analyze", "extract_key_points"
- `custom_prompt` (optional): Custom prompt for the analysis

### `list_files`
Lists all available files in the `files/` directory with their sizes and MIME types.

### `echo`
Simple echo tool for testing (no sampling required).

## Usage

1. **Prepare Files**: Place files to analyze in the `files/` directory
2. **Start Server**:
   ```bash
   go run cmd/enhanced_server/main.go
   ```
3. **Start Enhanced Client**: Run the enhanced client with Anthropic API integration
4. **Connect and Analyze**: Use any MCP client to call the analysis tools

## File Processing

The server handles different file types appropriately:
- **Text files**: Sent as plain text content
- **Images**: Encoded as base64 with proper MIME type for image analysis
- **Binary files**: Encoded as base64 with descriptive context

## Security

- Path traversal protection ensures files must be within the `files/` directory
- File existence validation before processing
- MIME type detection for appropriate content handling

## Example Workflow

1. Client calls `list_files` to see available files
2. Client calls `analyze_file` with specific filename and analysis type
3. Server reads the file and creates a sampling request
4. Server sends sampling request to the enhanced client via SSE
5. Enhanced client processes the request using Anthropic API
6. Enhanced client sends response back to server
7. Server returns analysis results to the original client

This demonstrates the complete bidirectional MCP sampling workflow with real LLM integration.