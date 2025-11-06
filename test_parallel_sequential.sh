#!/bin/bash

# Test script for parallel/sequential execution with interactive mode
# This demonstrates the new execution planning feature

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

clear

echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   Parallel/Sequential Execution Test - Interactive Mode   ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo

# Test directory
TEST_DIR="/home/velvetm/Desktop/mcp-test-files"

echo -e "${YELLOW}Starting interactive tests...${NC}"
echo -e "${BLUE}Working directory: $TEST_DIR${NC}"
echo

# Test 1: Sequential execution (write then read)
echo -e "${GREEN}Test 1: Sequential Execution (Create → Read)${NC}"
echo "Prompt: 'Create a file test_seq.txt with content \"Sequential test\" then read it back'"
echo
read -p "Press Enter to run test 1..."

echo "First, create a file called test_seq.txt with the content 'Sequential execution works!' and then immediately read the file to confirm it was created." | \
./ollama run qwen2.5:7b --tools "$TEST_DIR" 2>&1 | tee test1_output.log

echo
echo -e "${YELLOW}Check logs for: 'sequential=true reason=\"Tool names suggest sequential dependency\"'${NC}"
grep -q "sequential=true" server_parallel_sequential.log && echo -e "${GREEN}✓ Sequential execution detected!${NC}" || echo -e "${RED}✗ Sequential execution not detected${NC}"
echo

# Test 2: Parallel execution (read multiple files)
echo -e "${GREEN}Test 2: Parallel Execution (Read multiple files)${NC}"
echo "Prompt: 'Read three different files simultaneously'"
echo
read -p "Press Enter to run test 2..."

echo "Read the following three files and tell me their contents: test.txt, output.txt, and example.md" | \
./ollama run qwen2.5:7b --tools "$TEST_DIR" 2>&1 | tee test2_output.log

echo
echo -e "${YELLOW}Check logs for: 'sequential=false reason=\"Can execute in parallel\"'${NC}"
grep -q "sequential=false" server_parallel_sequential.log && echo -e "${GREEN}✓ Parallel execution detected!${NC}" || echo -e "${RED}✗ Parallel execution not detected${NC}"
echo

# Test 3: Mixed operations requiring sequential
echo -e "${GREEN}Test 3: Mixed Operations (Write multiple, then read all)${NC}"
echo "Prompt: 'Create files, modify them, then read all'"
echo
read -p "Press Enter to run test 3..."

echo "Create file1.txt with 'First file', create file2.txt with 'Second file', then read both files to verify their contents" | \
./ollama run qwen2.5:7b --tools "$TEST_DIR" 2>&1 | tee test3_output.log

echo
echo -e "${YELLOW}Checking execution strategy in logs...${NC}"
tail -20 server_parallel_sequential.log | grep -E "(Execution plan analyzed|Tool execution strategy)"
echo

# Interactive mode
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Now entering interactive mode for manual testing...${NC}"
echo -e "${BLUE}Try commands like:${NC}"
echo "  • 'Create a file and then read it' (should be sequential)"
echo "  • 'Read multiple existing files' (should be parallel)"
echo "  • 'List files, then create one, then list again' (should be sequential)"
echo
echo -e "${YELLOW}Type '/bye' or Ctrl+D to exit${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo

./ollama run qwen2.5:7b --tools "$TEST_DIR"

echo
echo -e "${CYAN}Testing complete! Check server_parallel_sequential.log for execution details.${NC}"