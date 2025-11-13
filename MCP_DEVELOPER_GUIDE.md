# MCP Server Developer Guide

This guide explains how to create custom Model Context Protocol (MCP) servers for Ollama.

## Table of Contents
- [What is an MCP Server?](#what-is-an-mcp-server)
- [Protocol Basics](#protocol-basics)
- [Required Methods](#required-methods)
- [Minimal Python Example](#minimal-python-example)
- [Minimal Node.js Example](#minimal-nodejs-example)
- [Registering Your Server](#registering-your-server)
- [Testing Your Server](#testing-your-server)
- [Debugging Tips](#debugging-tips)
- [Full Example: Calculator Server](#full-example-calculator-server)

## What is an MCP Server?

An MCP server is a standalone program that:
1. Communicates via JSON-RPC 2.0 over stdin/stdout
2. Provides tools that AI models can discover and use
3. Handles tool execution requests from the model
4. Returns structured results

## Protocol Basics

MCP uses JSON-RPC 2.0 protocol with these key concepts:

### Communication Flow
```
Ollama <---> MCP Manager <---> Your MCP Server
        JSON-RPC over stdin/stdout
```

### Message Format
```json
{
  "jsonrpc": "2.0",
  "method": "methodName",
  "params": { ... },
  "id": 1
}
```

### Response Format
```json
{
  "jsonrpc": "2.0",
  "result": { ... },
  "id": 1
}
```

## Required Methods

Your MCP server MUST implement these three methods:

### 1. `initialize`
Called when server starts. Returns server capabilities.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "initialize",
  "params": {
    "protocolVersion": "0.1.0",
    "capabilities": {},
    "clientInfo": { "name": "ollama", "version": "0.0.0" }
  },
  "id": 1
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "protocolVersion": "0.1.0",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "your-server-name",
      "version": "1.0.0"
    }
  },
  "id": 1
}
```

### 2. `tools/list`
Returns available tools and their schemas.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "params": {},
  "id": 2
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "tool_name",
        "description": "What this tool does",
        "inputSchema": {
          "type": "object",
          "properties": {
            "param1": {
              "type": "string",
              "description": "Parameter description"
            }
          },
          "required": ["param1"]
        }
      }
    ]
  },
  "id": 2
}
```

### 3. `tools/call`
Executes a tool and returns results.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      "param1": "value"
    }
  },
  "id": 3
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Tool execution result"
      }
    ]
  },
  "id": 3
}
```

## Minimal Python Example

```python
#!/usr/bin/env python3
import json
import sys

class MinimalMCPServer:
    def handle_request(self, request):
        method = request.get("method")
        request_id = request.get("id")
        
        if method == "initialize":
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "protocolVersion": "0.1.0",
                    "capabilities": {"tools": {}},
                    "serverInfo": {
                        "name": "minimal-server",
                        "version": "1.0.0"
                    }
                }
            }
        
        elif method == "tools/list":
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "tools": [{
                        "name": "hello",
                        "description": "Say hello",
                        "inputSchema": {
                            "type": "object",
                            "properties": {
                                "name": {
                                    "type": "string",
                                    "description": "Name to greet"
                                }
                            },
                            "required": ["name"]
                        }
                    }]
                }
            }
        
        elif method == "tools/call":
            name = request["params"]["arguments"].get("name", "World")
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "content": [{
                        "type": "text",
                        "text": f"Hello, {name}!"
                    }]
                }
            }

if __name__ == "__main__":
    server = MinimalMCPServer()
    while True:
        line = sys.stdin.readline()
        if not line:
            break
        request = json.loads(line)
        response = server.handle_request(request)
        sys.stdout.write(json.dumps(response) + "\n")
        sys.stdout.flush()
```

## Minimal Node.js Example

```javascript
#!/usr/bin/env node
const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

function handleRequest(request) {
  const { method, id, params } = request;
  
  switch(method) {
    case 'initialize':
      return {
        jsonrpc: '2.0',
        id,
        result: {
          protocolVersion: '0.1.0',
          capabilities: { tools: {} },
          serverInfo: {
            name: 'minimal-server',
            version: '1.0.0'
          }
        }
      };
    
    case 'tools/list':
      return {
        jsonrpc: '2.0',
        id,
        result: {
          tools: [{
            name: 'hello',
            description: 'Say hello',
            inputSchema: {
              type: 'object',
              properties: {
                name: {
                  type: 'string',
                  description: 'Name to greet'
                }
              },
              required: ['name']
            }
          }]
        }
      };
    
    case 'tools/call':
      const name = params.arguments.name || 'World';
      return {
        jsonrpc: '2.0',
        id,
        result: {
          content: [{
            type: 'text',
            text: `Hello, ${name}!`
          }]
        }
      };
  }
}

rl.on('line', (line) => {
  try {
    const request = JSON.parse(line);
    const response = handleRequest(request);
    process.stdout.write(JSON.stringify(response) + '\n');
  } catch (error) {
    // Send error response
    process.stdout.write(JSON.stringify({
      jsonrpc: '2.0',
      error: {
        code: -32603,
        message: error.message
      }
    }) + '\n');
  }
});
```

## Registering Your Server

### Option 1: User Configuration
Create or edit `~/.ollama/mcp-servers.json`:

```json
{
  "servers": [
    {
      "name": "my-custom-server",
      "description": "My custom MCP server",
      "command": "python3",
      "args": ["/path/to/my_server.py"],
      "requires_path": false
    }
  ]
}
```

### Option 2: Local Project Configuration
Create `mcp-servers.json` in your project directory:

```json
{
  "servers": [
    {
      "name": "project-tools",
      "command": "./my-server.py",
      "args": [],
      "requires_path": false
    }
  ]
}
```

### Option 3: Environment Variable
```bash
export OLLAMA_MCP_SERVERS='{"servers":[{"name":"my-server","command":"python3","args":["server.py"]}]}'
```

## Testing Your Server

### 1. Manual Testing
Test your server directly with JSON input:

```bash
# Test initialize
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | python3 my_server.py

# Test tools/list
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | python3 my_server.py

# Test tools/call
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"hello","arguments":{"name":"Alice"}},"id":3}' | python3 my_server.py
```

### 2. Integration Testing with Ollama

```bash
# Start your server manually first to see debug output
python3 my_server.py

# In another terminal, configure and test
echo '{"servers":[{"name":"test","command":"python3","args":["my_server.py"]}]}' > mcp-servers.json
./ollama run qwen2.5:7b --tools .

# Test the tool
>>> Use the hello tool to greet Alice
```

## Debugging Tips

### 1. Enable Debug Logging
Add logging to your server:

```python
import sys
import logging

# Log to stderr so it doesn't interfere with JSON-RPC
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(message)s',
    stream=sys.stderr
)

# In your handler:
logging.debug(f"Received request: {request}")
logging.debug(f"Sending response: {response}")
```

### 2. Common Issues and Solutions

**Server doesn't start:**
- Check file permissions: `chmod +x my_server.py`
- Verify shebang line: `#!/usr/bin/env python3`
- Test manually: `python3 my_server.py < test_input.json`

**Tools don't appear:**
- Ensure `tools/list` returns valid JSON
- Check for JSON syntax errors
- Verify tool names are unique

**Tool execution fails:**
- Log the incoming arguments
- Handle missing/invalid parameters gracefully
- Return errors in the correct format

**Server exits unexpectedly:**
- Catch all exceptions in the main loop
- Don't exit on invalid input
- Use try/except around JSON parsing

### 3. View Ollama Logs
```bash
# Run Ollama with debug logging
OLLAMA_DEBUG=1 ./ollama serve

# Check server logs
tail -f server_*.log | grep MCP
```

## Full Example: Calculator Server

Here's a complete example with error handling and multiple tools:

```python
#!/usr/bin/env python3
"""
Calculator MCP Server - A complete example
Provides basic math operations as tools
"""

import json
import sys
import logging
import traceback
from typing import Any, Dict

# Configure logging to stderr
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(levelname)s - %(message)s',
    stream=sys.stderr
)

class CalculatorMCPServer:
    def __init__(self):
        self.name = "calculator-mcp"
        self.version = "1.0.0"
        
    def handle_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle JSON-RPC 2.0 requests"""
        method = request.get("method", "")
        params = request.get("params", {})
        request_id = request.get("id")
        
        logging.debug(f"Handling request: {method}")
        
        try:
            if method == "initialize":
                return self.handle_initialize(request_id)
            elif method == "tools/list":
                return self.handle_tools_list(request_id)
            elif method == "tools/call":
                return self.handle_tool_call(params, request_id)
            else:
                return self.error_response(request_id, -32601, f"Method not found: {method}")
        except Exception as e:
            logging.error(f"Error handling request: {e}")
            return self.error_response(request_id, -32603, str(e))
    
    def handle_initialize(self, request_id: Any) -> Dict[str, Any]:
        """Handle initialization request"""
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {
                "protocolVersion": "0.1.0",
                "capabilities": {
                    "tools": {}
                },
                "serverInfo": {
                    "name": self.name,
                    "version": self.version
                }
            }
        }
    
    def handle_tools_list(self, request_id: Any) -> Dict[str, Any]:
        """List available calculator tools"""
        tools = [
            {
                "name": "add",
                "description": "Add two numbers",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number", "description": "First number"},
                        "b": {"type": "number", "description": "Second number"}
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "subtract",
                "description": "Subtract two numbers",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number", "description": "First number"},
                        "b": {"type": "number", "description": "Second number"}
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "multiply",
                "description": "Multiply two numbers",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number", "description": "First number"},
                        "b": {"type": "number", "description": "Second number"}
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "divide",
                "description": "Divide two numbers",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number", "description": "Dividend"},
                        "b": {"type": "number", "description": "Divisor (cannot be zero)"}
                    },
                    "required": ["a", "b"]
                }
            }
        ]
        
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {
                "tools": tools
            }
        }
    
    def handle_tool_call(self, params: Dict[str, Any], request_id: Any) -> Dict[str, Any]:
        """Execute a calculator tool"""
        tool_name = params.get("name", "")
        arguments = params.get("arguments", {})
        
        logging.debug(f"Executing tool: {tool_name} with arguments: {arguments}")
        
        try:
            # Extract numbers
            a = float(arguments.get("a", 0))
            b = float(arguments.get("b", 0))
            
            # Perform calculation
            if tool_name == "add":
                result = a + b
                operation = f"{a} + {b}"
            elif tool_name == "subtract":
                result = a - b
                operation = f"{a} - {b}"
            elif tool_name == "multiply":
                result = a * b
                operation = f"{a} × {b}"
            elif tool_name == "divide":
                if b == 0:
                    return self.tool_error(request_id, "Division by zero")
                result = a / b
                operation = f"{a} ÷ {b}"
            else:
                return self.tool_error(request_id, f"Unknown tool: {tool_name}")
            
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "content": [
                        {
                            "type": "text",
                            "text": f"{operation} = {result}"
                        }
                    ]
                }
            }
            
        except (KeyError, ValueError, TypeError) as e:
            return self.tool_error(request_id, f"Invalid arguments: {e}")
        except Exception as e:
            return self.tool_error(request_id, f"Calculation error: {e}")
    
    def tool_error(self, request_id: Any, message: str) -> Dict[str, Any]:
        """Return a tool execution error"""
        logging.error(f"Tool error: {message}")
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {
                "content": [
                    {
                        "type": "text",
                        "text": f"Error: {message}"
                    }
                ],
                "isError": True
            }
        }
    
    def error_response(self, request_id: Any, code: int, message: str) -> Dict[str, Any]:
        """Return a JSON-RPC error response"""
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": code,
                "message": message
            }
        }
    
    def run(self):
        """Main loop - read JSON-RPC from stdin, write to stdout"""
        logging.info(f"Starting {self.name} v{self.version}")
        
        while True:
            try:
                line = sys.stdin.readline()
                if not line:
                    logging.info("EOF received, shutting down")
                    break
                
                # Parse JSON request
                request = json.loads(line.strip())
                logging.debug(f"Received: {request}")
                
                # Handle the request
                response = self.handle_request(request)
                logging.debug(f"Sending: {response}")
                
                # Write response
                sys.stdout.write(json.dumps(response) + "\n")
                sys.stdout.flush()
                
            except json.JSONDecodeError as e:
                logging.error(f"JSON parse error: {e}")
                error_response = self.error_response(None, -32700, f"Parse error: {e}")
                sys.stdout.write(json.dumps(error_response) + "\n")
                sys.stdout.flush()
            except KeyboardInterrupt:
                logging.info("Interrupted, shutting down")
                break
            except Exception as e:
                logging.error(f"Unexpected error: {e}\n{traceback.format_exc()}")
                error_response = self.error_response(None, -32603, f"Internal error: {e}")
                sys.stdout.write(json.dumps(error_response) + "\n")
                sys.stdout.flush()

if __name__ == "__main__":
    server = CalculatorMCPServer()
    server.run()
```

## Additional Resources

- [MCP Specification](https://modelcontextprotocol.io/docs) - Official protocol documentation
- [MCP TypeScript SDK](https://github.com/anthropics/mcp) - Reference implementation
- [Example MCP Servers](https://github.com/modelcontextprotocol/servers) - Official server examples
- [Ollama MCP Integration](./MCP_INTEGRATION.md) - How MCP works in Ollama

## Contributing Your Server

If you've created a useful MCP server:

1. Test it thoroughly with multiple models
2. Document its capabilities and requirements
3. Add it to `examples/mcp-servers.json` as an example
4. Submit a pull request with:
   - Your server code (if open source)
   - Documentation
   - Configuration example
   - Test cases

## Questions?

For help with MCP server development:
- Check existing servers in `examples/`
- Review test files in `server/mcp_test.go`
- Open an issue with the `mcp` label
- Join the Ollama Discord #development channel