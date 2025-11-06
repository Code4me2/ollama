#!/bin/bash

# Test with explicit session ID and MCPServers config
SESSION_ID="test-session-$(date +%s)"
echo "Using session ID: $SESSION_ID"

# First request - list directory
echo "=== Request 1: List directory ==="
curl -s http://localhost:11434/api/chat -d "{
  \"model\": \"qwen2.5:7b\",
  \"session_id\": \"$SESSION_ID\",
  \"messages\": [{
    \"role\": \"user\",
    \"content\": \"List the files in /home/velvetm/Desktop/mcp-test-files\"
  }],
  \"stream\": false,
  \"mcp_servers\": [{
    \"name\": \"filesystem\",
    \"command\": \"npx\",
    \"args\": [\"-y\", \"@modelcontextprotocol/server-filesystem\", \"/home/velvetm/Desktop/mcp-test-files\"]
  }]
}" | python3 -m json.tool | head -20

sleep 2

# Second request - should reuse same MCP manager
echo -e "\n=== Request 2: Read a file (should reuse MCP session) ==="
curl -s http://localhost:11434/api/chat -d "{
  \"model\": \"qwen2.5:7b\",
  \"session_id\": \"$SESSION_ID\",
  \"messages\": [{
    \"role\": \"user\",
    \"content\": \"Read the file test_multi.txt from /home/velvetm/Desktop/mcp-test-files\"
  }],
  \"stream\": false,
  \"mcp_servers\": [{
    \"name\": \"filesystem\",
    \"command\": \"npx\",
    \"args\": [\"-y\", \"@modelcontextprotocol/server-filesystem\", \"/home/velvetm/Desktop/mcp-test-files\"]
  }]
}" | python3 -m json.tool | head -20
