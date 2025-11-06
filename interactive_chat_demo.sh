#!/bin/bash

# Interactive Chat Demo with MCP Tools
# Shows real-time interaction with tool execution

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

clear

echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║       Ollama Interactive Chat with MCP Tools Demo        ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo
echo -e "${GREEN}Starting interactive chat session with filesystem tools...${NC}"
echo -e "${YELLOW}Working directory: /home/velvetm/Desktop/mcp-test-files${NC}"
echo
echo -e "${BLUE}Available commands:${NC}"
echo "  - Ask to list, read, create, edit, or delete files"
echo "  - Type '/bye' or Ctrl+D to exit"
echo "  - Type '/help' for more examples"
echo
echo -e "${GREEN}Tip: Try these examples:${NC}"
echo "  • 'List all files in the current directory'"
echo "  • 'Read the content of test.txt'"
echo "  • 'Create a new file called hello.txt with some content'"
echo "  • 'Search for files containing the word test'"
echo
echo "═══════════════════════════════════════════════════════════"
echo

# Check if server is running
if ! pgrep -f "ollama serve" > /dev/null; then
    ./ollama serve > /dev/null 2>&1 &
    sleep 3
fi

# Start interactive session
./ollama run qwen2.5:0.5b --tools /home/velvetm/Desktop/mcp-test-files

echo
echo -e "${CYAN}Chat session ended. Thank you!${NC}"