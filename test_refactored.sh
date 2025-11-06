#!/bin/bash

# Test script for refactored multi-turn tool execution

echo "Starting Ollama server with refactored code..."

# Kill any existing ollama processes
pkill -f "ollama serve" 2>/dev/null

# Start the refactored server in background
./ollama serve > refactored_server.log 2>&1 &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
echo "Waiting for server to be ready..."
sleep 3

# Test 1: Simple tool call
echo ""
echo "Test 1: Simple single tool call"
echo "================================"
echo "List files in the test directory" | ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files 2>&1 | tee test1_result.log

# Test 2: Multi-turn tool execution
echo ""
echo "Test 2: Multi-turn tool execution"
echo "=================================="
echo "Create a file called test_multi.txt with content 'Multi-turn test successful!', then read it back to confirm." | ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files 2>&1 | tee test2_result.log

# Test 3: Simple conversation without tools
echo ""
echo "Test 3: Simple conversation (no tools)"
echo "======================================="
echo "Hello, how are you?" | ./ollama run qwen2.5:7b 2>&1 | tee test3_result.log

echo ""
echo "Checking server logs for issues..."
tail -50 refactored_server.log | grep -E "ERROR|panic|fatal" || echo "No errors found in server logs"

echo ""
echo "Tests complete. Cleaning up..."
kill $SERVER_PID 2>/dev/null
echo "Server stopped."

echo ""
echo "Summary:"
echo "- Check test1_result.log for single tool execution"
echo "- Check test2_result.log for multi-turn execution"
echo "- Check test3_result.log for simple conversation"
echo "- Check refactored_server.log for detailed server logs"