# Working SSE Examples - Reference Implementation

These examples demonstrate how Server-Sent Events (SSE) **should work** - providing a baseline for comparison with the broken MCP implementations.

## Purpose

When debugging MCP sampling issues, it's crucial to understand what working SSE communication looks like. These examples prove that:
- ✅ SSE connections establish cleanly
- ✅ Events stream in real-time without timeouts
- ✅ Bidirectional patterns work reliably
- ✅ The problem is in mcp-go, not the underlying concepts

## Quick Test

```bash
# Terminal 1: Start reference SSE server
go run cmd/basic_sse_server/main.go

# Terminal 2: Connect reference client
go run cmd/basic_sse_client/main.go

# You should see immediate, clean event streaming
```

## What You Should Observe

### ✅ Working SSE Connection
```
2024/08/30 14:30:15 SSE client starting...
2024/08/30 14:30:15 Connected to SSE stream
2024/08/30 14:30:16 Received: Hello from server! Time: 2024-08-30 14:30:16
2024/08/30 14:30:17 Received: Hello from server! Time: 2024-08-30 14:30:17
2024/08/30 14:30:18 Received: Hello from server! Time: 2024-08-30 14:30:18
```

### ❌ Broken MCP Connection (what you see with mcp-go)
```
2024/08/30 14:30:15 Starting enhanced MCP client...
2024/08/30 14:30:15 Connecting to server...
2024/08/30 14:30:45 Error: context deadline exceeded
2024/08/30 14:30:45 SSE connection failed
```

## Key Differences

| Aspect             | Working SSE                 | Broken mcp-go                |
| ------------------ | --------------------------- | ---------------------------- |
| **Connection**     | Instant HTTP 200            | Timeout after 30s            |
| **Event Delivery** | Immediate streaming         | No events received           |
| **Error Messages** | Clean success logs          | "context deadline exceeded"  |
| **Browser Test**   | Events visible in real-time | Connection never established |

## Implementation Highlights

### Clean SSE Headers
```go
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
```

### Immediate Event Delivery
```go
fmt.Fprintf(w, "data: %s\n\n", message)
if flusher, ok := w.(http.Flusher); ok {
    flusher.Flush()  // Critical: flush immediately
}
```

### Proper Client Parsing
```go
scanner := bufio.NewScanner(resp.Body)
for scanner.Scan() {
    line := scanner.Text()
    if strings.HasPrefix(line, "data: ") {
        data := strings.TrimPrefix(line, "data: ")
        // Process event immediately
    }
}
```

## Debugging Workflow

### Step 1: Establish Baseline
Run these working examples first to confirm your environment supports SSE properly.

### Step 2: Compare with MCP
Run the broken MCP examples in `../mcp-implementations/cmd/sampling-http-*` to see the difference.

### Step 3: Analyze the Gap
The difference between working SSE and broken MCP identifies the specific bugs in mcp-go.

## Browser Testing

Open `http://localhost:8080/web` while running `basic-sse-server` to see:
- Real-time events appearing in browser
- Clean Network tab showing SSE connection
- No timeout or connection errors

This proves the server-side SSE implementation works perfectly.

## What This Proves

1. **SSE Protocol Works**: The underlying technology is solid
1. **Go Implementation Works**: Our server/client code is correct
1. **Network is Fine**: No firewall or proxy issues
1. **MCP Library is Broken**: mcp-go has specific SSE bugs

## Using as Reference

When contributing fixes to mcp-go:
1. **Compare Behavior**: How does mcp-go differ from these examples?
1. **Copy Patterns**: Use these working implementations as templates
1. **Test Against Baseline**: Verify fixes work as well as these examples
1. **Document Differences**: Explain what mcp-go needs to change

## Files Included

- `cmd/basic_sse_server/main.go` - Minimal working SSE server
- `cmd/basic_sse_client/main.go` - Clean SSE client implementation

These are intentionally simple to isolate the core SSE functionality from MCP complexity.

---

**Key Insight**: The problem isn't SSE or bidirectional communication - it's the specific implementation
in mcp-go's StreamableHTTP transport! 🎯