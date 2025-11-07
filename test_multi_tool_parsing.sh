#!/bin/bash

echo "Testing multi-tool parsing with debug logs..."

# Kill any existing server
pkill ollama 2>/dev/null
sleep 2

# Start server with debug logging
echo "Starting server with debug logging..."
OLLAMA_DEBUG=1 ./ollama serve 2>&1 | grep -E "TOOL_PARSE|TOOL_ACCUMULATED|TOOL_FINAL|EXECUTION_DEBUG" > debug_tools.log &
SERVER_PID=$!
sleep 3

# Test with a simple multi-tool request
echo "Sending multi-tool request..."
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:0.5b",
    "messages": [
      {
        "role": "user",
        "content": "Call two tools: first list files, then read test.txt"
      }
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "list_files",
          "description": "List files in directory",
          "parameters": {
            "type": "object",
            "properties": {
              "path": {"type": "string"}
            }
          }
        }
      },
      {
        "type": "function", 
        "function": {
          "name": "read_file",
          "description": "Read a file",
          "parameters": {
            "type": "object",
            "properties": {
              "path": {"type": "string"}
            }
          }
        }
      }
    ],
    "stream": false
  }' | jq .

# Give server time to process
sleep 2

# Kill server
kill $SERVER_PID 2>/dev/null

echo "Debug logs:"
cat debug_tools.log

echo -e "\nTest complete!"