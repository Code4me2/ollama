#!/bin/bash

# Test true multi-round execution where model needs to process tool results

echo "Testing true multi-round execution..."
echo ""
echo "This test will:"
echo "1. Create a file with initial content"
echo "2. Read it back"
echo "3. Model should then decide to modify it"
echo "4. Read the modified version"
echo ""

# Start server if not running
if ! pgrep -f "ollama serve" > /dev/null; then
    ./ollama serve > multi_round_server.log 2>&1 &
    SERVER_PID=$!
    echo "Started server with PID: $SERVER_PID"
    sleep 3
fi

echo "Running multi-round test..."
echo "Create a file called story.txt with 'Once upon a time', then read it, then append ' there was a dragon' to it, and read the final version." | ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files 2>&1 | tee multi_round_result.log

echo ""
echo "Final file content:"
cat /home/velvetm/Desktop/mcp-test-files/story.txt 2>/dev/null || echo "File not found"

echo ""
echo "Checking for multiple tool executions..."
grep -c "🔧 Executing tool" multi_round_result.log && echo "tool calls found"

# Cleanup
rm -f /home/velvetm/Desktop/mcp-test-files/story.txt 2>/dev/null

if [ ! -z "$SERVER_PID" ]; then
    kill $SERVER_PID 2>/dev/null
    echo "Server stopped"
fi