#!/bin/bash

# Interactive Demo: Parallel vs Sequential Execution Planning
# This script demonstrates the intelligent execution planning feature

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

clear

echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║    Intelligent Tool Execution Planning Demo              ║${NC}"
echo -e "${CYAN}║    Sequential vs Parallel Execution Detection            ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo

echo -e "${YELLOW}This demo shows how Ollama intelligently decides whether to${NC}"
echo -e "${YELLOW}execute tools in parallel or sequentially based on dependencies.${NC}"
echo
echo -e "${GREEN}Watch the server logs to see execution decisions!${NC}"
echo -e "${BLUE}Logs location: ./server_parallel_sequential.log${NC}"
echo

# Monitor logs in background (optional)
if command -v tail &> /dev/null; then
    echo -e "${MAGENTA}Starting log monitor in background...${NC}"
    tail -f server_parallel_sequential.log | grep -E "(Execution plan analyzed|Tool execution strategy)" &
    LOG_PID=$!
    echo
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo

# Demo 1: Sequential execution
echo -e "${GREEN}Demo 1: Sequential Execution${NC}"
echo -e "${BLUE}Task: Create a file, then read it${NC}"
echo -e "This will execute ${RED}SEQUENTIALLY${NC} because reading depends on creation"
echo
echo "Create a file test_seq.txt with 'Sequential test' then read it back" | \
    ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files
echo

sleep 2

# Demo 2: Parallel execution
echo -e "${GREEN}Demo 2: Parallel Execution${NC}"
echo -e "${BLUE}Task: Read multiple independent files${NC}"
echo -e "This will execute ${GREEN}IN PARALLEL${NC} because operations are independent"
echo
echo "Read test.txt, output.txt, and example.md simultaneously" | \
    ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files
echo

sleep 2

# Demo 3: Complex scenario
echo -e "${GREEN}Demo 3: Mixed Operations${NC}"
echo -e "${BLUE}Task: List files, create new file, list again${NC}"
echo -e "This will execute ${RED}SEQUENTIALLY${NC} due to ordering dependencies"
echo
echo "List all files, then create hello_world.txt with 'Hello!', then list files again" | \
    ./ollama run qwen2.5:7b --tools /home/velvetm/Desktop/mcp-test-files
echo

echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Demo complete!${NC}"
echo
echo -e "${YELLOW}Review the execution strategies:${NC}"
echo -e "  • ${RED}Sequential${NC}: When operations depend on each other"
echo -e "  • ${GREEN}Parallel${NC}: When operations are independent"
echo
echo -e "${BLUE}Check the logs for detailed execution planning:${NC}"
echo "  grep 'Execution plan' server_parallel_sequential.log"
echo

# Kill log monitor if started
if [ ! -z "$LOG_PID" ]; then
    kill $LOG_PID 2>/dev/null
fi