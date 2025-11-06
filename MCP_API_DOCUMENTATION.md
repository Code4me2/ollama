# MCP (Model Context Protocol) API Documentation

## Overview

The MCP integration in Ollama enables autonomous tool execution during model inference. This allows language models to interact with external systems and tools via the Model Context Protocol without requiring client-side tool handling.

## API Endpoints

### Chat API with MCP Support

**Endpoint:** `POST /api/chat`

The standard chat endpoint now supports MCP server configuration for autonomous tool execution.

#### Request Format

```json
{
  "model": "qwen2.5:0.5b",
  "messages": [
    {
      "role": "user",
      "content": "List the files in the current directory"
    }
  ],
  "mcp_servers": [
    {
      "name": "filesystem",
      "command": "python",
      "args": ["-m", "mcp_server_filesystem", "/safe/path"],
      "env": {
        "PYTHONPATH": "/usr/local/lib/python3.9/site-packages"
      }
    }
  ],
  "max_tool_rounds": 10,
  "tool_timeout": 30000,
  "stream": false
}
```

#### Response Format

```json
{
  "message": {
    "role": "assistant",
    "content": "Here are the files in the current directory:\n- file1.txt\n- file2.md\n- directory1/",
    "tool_calls": [
      {
        "function": {
          "name": "filesystem:list_directory",
          "arguments": {
            "path": "."
          }
        }
      }
    ]
  },
  "done": true,
  "total_duration": 2345678901,
  "eval_count": 150,
  "eval_duration": 1234567890
}
```

### CLI Usage

#### Basic Tool Execution

```bash
ollama run qwen2.5:0.5b --tools /path/to/allowed/directory
```

This enables filesystem tools with access restricted to the specified directory.

#### Multiple MCP Servers

```bash
ollama run llama3.1 \
  --mcp filesystem:python:-m:mcp_server_filesystem:/safe/path \
  --mcp weather:python:-m:weather_mcp_server
```

## Security Features

### 1. Environment Variable Filtering

The MCP integration automatically filters sensitive environment variables to prevent credential leakage:

**Filtered Variables:**
- AWS credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- API keys (`*_API_KEY`, `*_TOKEN`)
- Authentication tokens (`GITHUB_TOKEN`, `GITLAB_TOKEN`)
- SSH/GPG agent information
- Database credentials
- Any variable containing `SECRET`, `PASSWORD`, `TOKEN`, or `KEY`

**Allowed Variables:**
- `PATH` (sanitized)
- `HOME`, `USER`
- `LANG`, `LC_ALL`, `TZ`
- `TMPDIR`, `TEMP`, `TMP`
- `PYTHONPATH`, `NODE_PATH`

### 2. Command Validation

Dangerous commands are blocked for security:

**Blocked Commands:**
- Shell interpreters: `sh`, `bash`, `zsh`, etc.
- Privilege escalation: `sudo`, `su`, `doas`
- Destructive commands: `rm`, `dd`, `mkfs`
- Network tools: `curl`, `wget`, `nc`
- System management: `systemctl`, `service`

### 3. Path Sanitization

PATH environment variable is sanitized to remove:
- Root directories (`/root/*`)
- Relative paths (`./`, `../`)
- User home shortcuts (`~`)

### 4. Argument Validation

Tool arguments are validated to prevent:
- Shell injection (blocks `;`, `|`, `&`, `` ` ``, `$()`)
- Path traversal (blocks `..`)
- Command substitution

### 5. Process Isolation

MCP servers run with:
- Separate process groups
- Resource limits
- Timeout enforcement (30 seconds default)

## Configuration Examples

### Example 1: Filesystem Access

```json
{
  "mcp_servers": [
    {
      "name": "documents",
      "command": "python",
      "args": ["-m", "mcp_server_filesystem", "/home/user/documents"],
      "env": {
        "READONLY": "true"
      }
    }
  ]
}
```

### Example 2: Database Query Tool

```json
{
  "mcp_servers": [
    {
      "name": "database",
      "command": "node",
      "args": ["/usr/local/lib/mcp/database_server.js"],
      "env": {
        "DB_HOST": "localhost",
        "DB_NAME": "analytics",
        "DB_USER": "readonly_user"
      }
    }
  ]
}
```

### Example 3: Weather API

```json
{
  "mcp_servers": [
    {
      "name": "weather",
      "command": "/usr/local/bin/weather_mcp",
      "args": ["--cache", "60"],
      "env": {
        "WEATHER_CACHE_DIR": "/tmp/weather_cache"
      }
    }
  ]
}
```

## Tool Execution Flow

1. **Model generates tool call** in JSON format
2. **Parser detects tool call** during streaming
3. **Execution pauses** while tool runs
4. **MCP server executes** tool via JSON-RPC
5. **Results injected** back into context
6. **Model continues** generation with results
7. **Process repeats** for multiple tool calls

## Performance Considerations

### Caching

Tool results are cached with:
- **Default TTL:** 5 minutes
- **Max entries:** 1000
- **Cache key:** Tool name + arguments

### Limits

- **Max MCP servers:** 10 (configurable)
- **Max tool rounds:** 15 per request
- **Tool timeout:** 30 seconds
- **Parallel execution:** Up to 10 concurrent tools

### Latency

- **Tool execution:** ~100-300ms average
- **Cache hit:** <1ms
- **MCP initialization:** ~500ms per server

## Error Handling

### Tool Execution Errors

When a tool fails, the error is returned to the model:

```json
{
  "tool_calls": [
    {
      "function": {
        "name": "filesystem:read_file",
        "arguments": {"path": "/etc/shadow"}
      }
    }
  ],
  "tool_results": [
    {
      "error": "Access denied - path outside allowed directories"
    }
  ]
}
```

### MCP Server Failures

If an MCP server crashes:
1. Error returned to model
2. Server marked as failed
3. Automatic restart attempted on next request

### Timeout Handling

Tools exceeding timeout (30s default) are terminated:
- Process killed with SIGTERM
- 5-second grace period
- SIGKILL if still running
- Error returned to model

## Testing Tools

### Mock MCP Server

For testing, create a simple MCP server:

```python
#!/usr/bin/env python3
import json
import sys

def handle_request(request):
    if request["method"] == "initialize":
        return {"protocolVersion": "1.0", "serverInfo": {"name": "test"}}
    elif request["method"] == "tools/list":
        return {"tools": [
            {
                "name": "echo",
                "description": "Echo input",
                "inputSchema": {"type": "object"}
            }
        ]}
    elif request["method"] == "tools/call":
        return {"content": [{"type": "text", "text": f"Echo: {request['params']}"}]}

while True:
    line = sys.stdin.readline()
    if not line:
        break
    request = json.loads(line)
    response = {
        "jsonrpc": "2.0",
        "id": request.get("id"),
        "result": handle_request(request)
    }
    print(json.dumps(response))
    sys.stdout.flush()
```

### Testing Security

Test that dangerous commands are blocked:

```bash
# This should fail
curl -X POST http://localhost:11434/api/chat -d '{
  "model": "qwen2.5:0.5b",
  "messages": [{"role": "user", "content": "test"}],
  "mcp_servers": [{
    "name": "test",
    "command": "bash",
    "args": ["-c", "echo hacked"]
  }]
}'
```

Expected: Error about dangerous command

## Troubleshooting

### Debug Logging

Enable debug logging to see MCP interactions:

```bash
OLLAMA_DEBUG=INFO ollama serve
```

### Common Issues

1. **"Tool not found"**
   - Check MCP server initialized correctly
   - Verify tool name includes namespace prefix

2. **"MCP server failed to initialize"**
   - Check command path is correct
   - Verify Python/Node environment
   - Check server implements MCP protocol

3. **"Access denied"**
   - Path outside allowed directories
   - Insufficient permissions
   - Security policy violation

4. **"Tool execution timeout"**
   - Tool taking >30 seconds
   - Increase timeout in request
   - Check for infinite loops

## Best Practices

1. **Use absolute paths** for MCP server commands
2. **Restrict filesystem access** to specific directories
3. **Set appropriate timeouts** for long-running tools
4. **Cache frequently used** tool results
5. **Validate tool inputs** in MCP server implementation
6. **Log tool execution** for audit purposes
7. **Test security boundaries** regularly
8. **Monitor resource usage** of MCP servers

## Migration Guide

### From Client-Side Tools

Before (client handles tools):
```python
response = ollama.chat(model="llama3.1", messages=messages, tools=tools)
if response.tool_calls:
    for tool_call in response.tool_calls:
        result = execute_tool(tool_call)
        messages.append({"role": "tool", "content": result})
    response = ollama.chat(model="llama3.1", messages=messages)
```

After (server handles tools):
```python
response = ollama.chat(
    model="llama3.1",
    messages=messages,
    mcp_servers=[{"name": "tools", "command": "mcp_server"}]
)
# Tools executed automatically, final response returned
```

## Limitations

1. **Platform Support:** Linux/macOS (Windows support pending)
2. **Protocol:** MCP 1.0 only
3. **Transport:** stdio only (no HTTP/WebSocket yet)
4. **Models:** Best with tool-capable models (Qwen, Llama 3.1+)
5. **Concurrency:** Max 10 parallel MCP servers

## Future Enhancements

- [ ] Windows support
- [ ] HTTP/WebSocket transports
- [ ] MCP 2.0 protocol support
- [ ] Dynamic tool loading
- [ ] Tool usage analytics
- [ ] Rate limiting per tool
- [ ] Tool result streaming
- [ ] Conditional tool execution