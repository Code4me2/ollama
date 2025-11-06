#!/bin/bash

# Test script for interactive chat with MCP tools
# This script demonstrates various scenarios including error handling

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Ollama MCP Tools Interactive Chat Test ===${NC}"
echo

# Check if server is running
if ! pgrep -f "ollama serve" > /dev/null; then
    echo -e "${YELLOW}Starting ollama server...${NC}"
    ./ollama serve > /dev/null 2>&1 &
    SERVER_PID=$!
    sleep 3
else
    echo -e "${GREEN}Ollama server already running${NC}"
    SERVER_PID=""
fi

# Function to send a message to ollama and capture response
test_scenario() {
    local scenario_name="$1"
    local input_text="$2"
    local expected_behavior="$3"
    
    echo -e "${YELLOW}Test: $scenario_name${NC}"
    echo -e "Input: $input_text"
    echo -e "Expected: $expected_behavior"
    echo
    
    # Use timeout to prevent hanging
    echo "$input_text" | timeout 15 ./ollama run qwen2.5:0.5b --tools /home/velvetm/Desktop/mcp-test-files 2>&1
    
    echo
    echo "---"
    echo
}

# Test 1: Valid tool call - list directory
test_scenario \
    "Valid Tool Call - List Directory" \
    "List all files in /home/velvetm/Desktop/mcp-test-files" \
    "Should successfully list files"

# Test 2: Invalid parameters
test_scenario \
    "Invalid Parameters" \
    "Use the edit_file tool to change test.txt but pass a string instead of an array for edits parameter" \
    "Should show 'Invalid tool parameters' error and retry"

# Test 3: Access denied - outside directory
test_scenario \
    "Access Denied - Outside Directory" \
    "Read the file /etc/passwd" \
    "Should show 'Access denied - path outside allowed directories' error"

# Test 4: File not found
test_scenario \
    "File Not Found" \
    "Read the file /home/velvetm/Desktop/mcp-test-files/nonexistent.txt" \
    "Should show 'File not found' error"

# Test 5: Multiple tools in sequence
test_scenario \
    "Multiple Tools in Sequence" \
    "First list the files, then read test.txt, then create a new file called demo.txt with content 'Hello from test script'" \
    "Should execute multiple tools successfully"

# Test 6: Tool recovery after error
test_scenario \
    "Error Recovery" \
    "Try to read /invalid/path.txt, then when that fails, list the valid files instead" \
    "Should recover from error and execute valid tool"

# Cleanup
if [ ! -z "$SERVER_PID" ]; then
    echo -e "${YELLOW}Stopping ollama server...${NC}"
    kill $SERVER_PID 2>/dev/null
fi

echo -e "${GREEN}Test completed!${NC}"