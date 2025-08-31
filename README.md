# Learn MCP Sampling

Understanding Model Context Protocol (MCP) sampling through hands-on examples and debugging real library issues.

## What You'll Learn

1. **MCP Protocol** - How sampling should work in theory
2. **Library Analysis** - What's broken in mcp-go implementation  
3. **Reference Implementation** - Working examples for comparison
4. **Debugging Techniques** - Tools and methods for protocol analysis

## Prerequisites

**SSE Fundamentals Required**: This project assumes you understand Server-Sent Events and bidirectional communication patterns. If you're new to SSE, start with: [learn-sse-bidirectional](https://github.com/pavelanni/learn-sse-bidirectional)

## Quick Start

### Environment Setup
```bash
# Required: Anthropic API key for real LLM integration
export ANTHROPIC_API_KEY="your-api-key-here"

# Optional: Enable verbose logging
export MCP_DEBUG=1
```

### Test the Working Examples
```bash
# Terminal 1: Start enhanced server with file analysis
go run mcp-implementations/cmd/enhanced-server/main.go

# Terminal 2: Connect enhanced client with real LLM
go run mcp-implementations/cmd/enhanced-client/main.go

# Terminal 3: Run workflow test
go run debugging-tools/cmd/test-workflow/main.go
```

### Compare with SSE Reference
```bash
# See how SSE should work (from working-examples/)
go run working-examples/cmd/basic-sse-server/main.go
```

## Project Goals

### 1. Document the MCP Sampling Problem
The mcp-go library has fundamental issues with SSE implementation that prevent MCP sampling from working correctly. This project:
- Identifies specific bugs (GitHub issue #530)
- Provides working reference implementations
- Offers debugging tools and techniques

### 2. Create Educational Resources
Learn MCP through progressively complex examples:
- Working SSE patterns (baseline)
- Mock MCP implementations (learning)
- Real LLM integration (production-ready)
- Debugging tools (troubleshooting)

### 3. Enable Community Contribution
Help fix the mcp-go library by providing:
- Clear problem reproduction
- Working alternative implementations  
- Detailed analysis documentation
- Test workflows for validation

## Directory Structure

```
learn-mcp-sampling/
├── README.md                      # This overview
├── MCP_PROTOCOL_EXPLAINED.md      # MCP-specific concepts  
├── working-examples/               # SSE reference implementations
│   ├── cmd/
│   │   ├── basic-sse-server/      # How SSE should work
│   │   └── basic-sse-client/      # Clean connection example
│   └── README.md                  # SSE comparison guide
├── mcp-implementations/            # MCP protocol examples
│   ├── cmd/
│   │   ├── enhanced-server/       # Real file analysis server
│   │   ├── enhanced-client/       # Anthropic API integration
│   │   ├── sampling-http-server/  # Basic MCP server
│   │   ├── sampling-http-client/  # Mock sampling client
│   │   └── simulate-sampling/     # Working simulation
│   └── files/                     # Sample files for analysis
├── debugging-tools/                # Analysis and testing tools
│   ├── cmd/
│   │   ├── test-workflow/         # End-to-end testing
│   │   ├── check-sampling-clients/# Connection diagnostics
│   │   ├── debug-server/          # SSE debugging server
│   │   └── all-in-one-client/     # Session testing
│   └── analysis/
│       ├── SAMPLING_ISSUE_ANALYSIS.md # Bug documentation
│       └── LIBRARY_BUGS.md           # GitHub issues summary
├── go.mod                         # Go module definition
└── go.sum                         # Dependency checksums
```

## Learning Path

### Phase 1: Understand the Problem
1. **Read the Analysis**: Start with `debugging-tools/analysis/SAMPLING_ISSUE_ANALYSIS.md`
2. **See Working SSE**: Run examples in `working-examples/` 
3. **Try Broken MCP**: Attempt `mcp-implementations/cmd/sampling-http-*`
4. **Compare Results**: Identify the difference

### Phase 2: Study Working Solutions
1. **Enhanced Server**: Real file analysis with `enhanced-server/main.go`
2. **Anthropic Integration**: Live LLM calls with `enhanced-client/main.go`
3. **Simulation**: See intended behavior with `simulate-sampling/main.go`

### Phase 3: Debug and Contribute
1. **Use Debug Tools**: Test connections with tools in `debugging-tools/cmd/`
2. **Analyze Library Code**: Understand mcp-go StreamableHTTP issues
3. **Propose Fixes**: Contribute to mcp-go project with evidence

## Key Technical Insights

### How MCP Sampling Should Work
```
1. MCP Client declares sampling capability
2. Someone calls a tool that needs LLM analysis
3. MCP Server sends sampling request via SSE  
4. MCP Client processes request using LLM API
5. MCP Client returns results via HTTP POST
6. MCP Server provides results to tool caller
```

### What's Actually Broken
- **SSE Connection Issues**: "context deadline exceeded" errors
- **Missing Transport Options**: Need `WithContinuousListening()`
- **Session Management**: Improper header handling
- **Event Parsing**: SSE stream processing bugs

### Working Alternatives  
Our enhanced implementations prove the pattern works:
- ✅ Clean SSE connections with instant delivery
- ✅ Proper bidirectional communication
- ✅ Real Anthropic API integration
- ✅ Multi-modal content support (text, images, PDFs)

## Real-World Applications

### File Analysis System (Enhanced Implementation)
- **Server**: Serves files from local directory for analysis
- **Client**: Uses Claude 3.5 Sonnet for content analysis  
- **Capabilities**: Text, image, and binary file support
- **Features**: Summarization, code analysis, visual understanding

### Production Patterns Demonstrated
- **Authentication**: API key management
- **Error Handling**: Graceful failure modes
- **Logging**: Comprehensive request/response tracking
- **Token Usage**: Cost monitoring and optimization
- **Security**: Path traversal protection

## Debug Workflow

### Connection Testing
```bash
# Test basic connectivity
go run debugging-tools/cmd/check-sampling-clients/main.go

# Debug SSE streams
go run debugging-tools/cmd/debug-server/main.go

# Full workflow test  
go run debugging-tools/cmd/test-workflow/main.go
```

### Issue Identification
1. **SSE Stream Analysis**: Monitor connection establishment
2. **Event Delivery Testing**: Verify message transmission
3. **Response Correlation**: Check request/response matching
4. **Session Management**: Validate MCP headers

### Library Comparison
Compare mcp-go behavior with our reference implementations:
- Connection establishment patterns
- Event streaming reliability  
- Error handling approaches
- Session lifecycle management

## Contributing to mcp-go

### Current Known Issues
- **GitHub Issue #530**: SSE connection timeouts
- **Missing Features**: Continuous listening support
- **Documentation**: Incomplete sampling examples
- **Testing**: Limited real-world scenarios

### How This Project Helps
- **Clear Reproduction**: Demonstrates exact failure conditions
- **Working Reference**: Shows how it should behave
- **Test Cases**: Provides validation scenarios
- **Documentation**: Explains complex concepts clearly

### Contribution Workflow
1. Use our debugging tools to isolate specific issues
2. Reference our working implementations for correct behavior
3. Submit PRs to mcp-go with evidence from this project
4. Test fixes using our comprehensive test suite

## Advanced Features

### Multi-Modal Analysis
- **Text Files**: Code, documentation, configuration
- **Images**: Screenshots, diagrams, photos
- **Binary Files**: PDFs, archives, executables
- **Custom Prompts**: Flexible analysis workflows

### Production Considerations
- **Rate Limiting**: API call management
- **Cost Control**: Token usage monitoring
- **Error Recovery**: Robust failure handling
- **Scaling**: Multiple client support

## Related Projects

- [learn-sse-bidirectional](https://github.com/pavelanni/learn-sse-bidirectional) - SSE fundamentals (prerequisite)
- [mcp-go](https://github.com/mark3labs/mcp-go) - The library we're helping to fix
- [Model Context Protocol](https://modelcontextprotocol.io/) - Official MCP specification
- [Caret Labs](https://caretlabs.dev) - Hands-on learning methodology

## Getting Help

### Common Issues
- **"context deadline exceeded"**: Use `WithContinuousListening()` or our enhanced examples
- **No sampling requests**: Check server tool registration and client capability declaration  
- **Authentication errors**: Verify `ANTHROPIC_API_KEY` environment variable
- **Connection failures**: Compare with working SSE examples in this project

### Debug Process
1. Start with working examples to establish baseline
2. Use debugging tools to identify specific failure points
3. Consult analysis documentation for known issues
4. Compare behavior with reference implementations

## License

MIT License - Use freely for learning, debugging, and contributing!

---

**Debug Responsibly!** 🔍

*Part of the Caret Labs learning methodology - hands-on, progressive, multi-component education for developers.*

*Contributing to open source by making complex protocols understandable and fixable.*