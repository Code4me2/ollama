# MCP Testing Plan - Detailed Roadmap

## Overview
This document outlines specific testing procedures for MCP servers and multi-model compatibility before open source release.

## Phase 1: MCP Server Testing (Week 1)

### 1. Filesystem Server ✅ (Completed)
**Server**: `@modelcontextprotocol/server-filesystem`
- [x] List directory operations
- [x] Read file operations
- [ ] Write file operations
- [ ] Search operations
- [ ] Permission testing (restricted paths)

### 2. Git Server 🔄 (Partially Tested)
**Server**: `@modelcontextprotocol/server-git`
```bash
# Installation
npm install -g @modelcontextprotocol/server-git

# Test commands
./ollama run qwen2.5:7b --mcp git:npx:@modelcontextprotocol/server-git:.
>>> Show me the recent commits
>>> What files were changed in the last commit?
>>> Search for TODO comments in the codebase
```

**Test Cases**:
- [ ] List commits
- [ ] Show diff
- [ ] Search code
- [ ] Blame operations
- [ ] Branch information

### 3. PostgreSQL Server ⏳ (Not Tested)
**Server**: `@modelcontextprotocol/server-postgres`
```bash
# Setup test database
docker run -d --name test-postgres \
  -e POSTGRES_PASSWORD=test \
  -p 5432:5432 \
  postgres:14

# Configure MCP
{
  "name": "postgres",
  "command": "npx",
  "args": ["@modelcontextprotocol/server-postgres", "postgresql://postgres:test@localhost/test"],
  "env": {}
}
```

**Test Cases**:
- [ ] List tables
- [ ] Query execution
- [ ] Schema information
- [ ] Data insertion
- [ ] Transaction handling

### 4. Python Server ⏳ (Not Tested)
**Server**: `mcp-server-python`
```bash
# Installation
pip install mcp-server-python

# Test configuration
{
  "name": "python",
  "command": "python",
  "args": ["-m", "mcp_server_python"],
  "env": {"PYTHONPATH": "/usr/local/lib/python3.9/site-packages"}
}
```

**Test Cases**:
- [ ] Execute Python code
- [ ] Import libraries
- [ ] Data manipulation
- [ ] File operations via Python
- [ ] Error handling

### 5. Web/Puppeteer Server ⏳ (Not Tested)
**Server**: `@modelcontextprotocol/server-puppeteer`
```bash
# Installation
npm install -g @modelcontextprotocol/server-puppeteer

# Test configuration
{
  "name": "browser",
  "command": "npx",
  "args": ["@modelcontextprotocol/server-puppeteer"],
  "env": {}
}
```

**Test Cases**:
- [ ] Navigate to URL
- [ ] Extract page content
- [ ] Take screenshots
- [ ] Fill forms
- [ ] Click elements

### 6. GitHub Server ⏳ (Not Tested)
**Server**: `@modelcontextprotocol/server-github`
```bash
# Installation
npm install -g @modelcontextprotocol/server-github

# Configuration with token
{
  "name": "github",
  "command": "npx",
  "args": ["@modelcontextprotocol/server-github"],
  "env": {"GITHUB_TOKEN": "ghp_..."}
}
```

**Test Cases**:
- [ ] List repositories
- [ ] Create issues
- [ ] Search code
- [ ] PR operations
- [ ] Gist management

### 7. Slack Server ⏳ (Not Tested)
**Server**: `@modelcontextprotocol/server-slack`
```bash
# Configuration
{
  "name": "slack",
  "command": "npx",
  "args": ["@modelcontextprotocol/server-slack"],
  "env": {"SLACK_TOKEN": "xoxb-..."}
}
```

**Test Cases**:
- [ ] List channels
- [ ] Send messages
- [ ] Read messages
- [ ] User information
- [ ] File uploads

## Phase 2: Multi-Model Compatibility Testing (Week 1-2)

### Test Matrix

| Model | Size | MCP Support | Test Status | Notes |
|-------|------|-------------|-------------|-------|
| **Qwen Family** |
| qwen2.5:0.5b | 494M | ✅ Full | ✅ Tested | Primary small model |
| qwen2.5:7b | 7.6B | ✅ Full | ✅ Tested | Primary test model |
| qwen2.5:14b | 14B | ✅ Full | ⏳ Pending | |
| qwen2.5:32b | 32B | ✅ Full | ⏳ Pending | |
| qwen2.5:72b | 72B | ✅ Full | ⏳ Pending | Memory intensive |
| **Llama Family** |
| llama3.2:1b | 1.3B | ❓ Unknown | ⏳ Pending | |
| llama3.2:3b | 2.0B | ❓ Unknown | ⏳ Pending | |
| llama3.1:8b | 4.7B | ❓ Unknown | ⏳ Pending | |
| llama3.1:70b | 43GB | ❓ Unknown | ⏳ Pending | |
| codellama:7b | 3.8B | ❓ Unknown | ⏳ Pending | Code-specific |
| **Mistral Family** |
| mistral:7b | 4.1B | ❓ Unknown | ⏳ Pending | |
| mixtral:8x7b | 26GB | ❓ Unknown | ⏳ Pending | MoE model |
| **Other Models** |
| phi3:3.8b | 2.3B | ❓ Unknown | ⏳ Pending | Microsoft model |
| gemma2:9b | 5.5B | ❓ Unknown | ⏳ Pending | Google model |
| deepseek-coder:6.7b | 4.0B | ❓ Unknown | ⏳ Pending | Code model |
| neural-chat:7b | 4.1B | ❓ Unknown | ⏳ Pending | Intel model |

### Testing Procedure for Each Model

```bash
#!/bin/bash
MODEL=$1

echo "Testing MCP support for $MODEL"

# 1. Pull the model
./ollama pull $MODEL

# 2. Test basic completion (no tools)
echo "Test 1: Basic completion"
echo "Hello, how are you?" | ./ollama run $MODEL

# 3. Test with filesystem tools
echo "Test 2: Filesystem tools"
echo "List files in /tmp" | ./ollama run $MODEL --tools /tmp

# 4. Test with multiple tools via API
echo "Test 3: API with multiple tools"
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"messages\": [{\"role\": \"user\", \"content\": \"What files are in the current directory?\"}],
    \"mcp_servers\": [{
      \"name\": \"filesystem\",
      \"command\": \"npx\",
      \"args\": [\"@modelcontextprotocol/server-filesystem\", \".\"],
      \"env\": {}
    }],
    \"stream\": false
  }"

# 5. Check for tool call detection
echo "Test 4: Tool call parsing"
tail -100 server.log | grep -E "tool_calls|MCP"
```

### Success Criteria for Model Support

**Full Support** ✅:
- Correctly generates tool calls in expected format
- Parser detects and executes tools
- Handles tool results appropriately
- No errors in server logs

**Partial Support** ⚠️:
- May generate tool calls but format issues
- Parser needs adjustments
- Some tools work, others fail
- Minor errors that can be fixed

**No Support** ❌:
- Doesn't generate tool calls
- Parser cannot detect calls
- Causes server errors
- Would require major changes

## Phase 3: Performance Testing (Week 2)

### Load Testing Scenarios

#### 1. Concurrent Tool Execution
```bash
# Test 10 concurrent requests with tools
for i in {1..10}; do
  curl -X POST http://localhost:11434/api/generate \
    -d "{
      \"model\": \"qwen2.5:0.5b\",
      \"prompt\": \"List files in /tmp\",
      \"mcp_servers\": [{
        \"name\": \"filesystem\",
        \"command\": \"npx\",
        \"args\": [\"@modelcontextprotocol/server-filesystem\", \"/tmp\"]
      }]
    }" &
done
wait
```

#### 2. Memory Leak Testing
```bash
# Monitor memory over 1000 requests
./test_memory_leak.sh

#!/bin/bash
# test_memory_leak.sh
for i in {1..1000}; do
  echo "Request $i"
  echo "List files" | ./ollama run qwen2.5:0.5b --tools /tmp
  
  # Check memory every 100 requests
  if [ $((i % 100)) -eq 0 ]; then
    ps aux | grep ollama | grep -v grep
  fi
done
```

#### 3. Tool Timeout Testing
```bash
# Test with slow/hanging tools
{
  "name": "slow_tool",
  "command": "bash",
  "args": ["-c", "sleep 60"],
  "timeout": 5000
}
```

#### 4. Error Recovery Testing
- Kill MCP server mid-execution
- Send malformed tool responses
- Exceed resource limits
- Network interruptions

### Performance Metrics to Track

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Tool execution latency | <100ms | ~100ms | ✅ |
| Memory per request | <10MB | Unknown | ⏳ |
| Concurrent tool limit | 10+ | Unknown | ⏳ |
| Cache hit rate | >80% | ~85% | ✅ |
| Error recovery time | <5s | Unknown | ⏳ |
| Max requests/sec | 100+ | Unknown | ⏳ |

## Phase 4: Security Testing (Week 2)

### Security Test Cases

1. **Path Traversal**
```bash
# Should be blocked
./ollama run qwen2.5:7b --tools /
>>> Delete ../../../etc/passwd
```

2. **Command Injection**
```bash
# Test shell injection attempts
{
  "name": "test",
  "command": "echo",
  "args": ["test; rm -rf /"]
}
```

3. **Resource Exhaustion**
```bash
# Fork bomb attempt
./ollama run qwen2.5:7b
>>> Create 1000000 files in /tmp
```

4. **Privilege Escalation**
```bash
# Should be blocked
{
  "name": "test",
  "command": "sudo",
  "args": ["ls"]
}
```

## Testing Schedule

### Week 1 (Current)
- [x] Day 1-3: 72-hour stability test (started)
- [ ] Day 4: Test Git, PostgreSQL servers
- [ ] Day 5: Test Python, Web/Puppeteer servers
- [ ] Day 6: Test GitHub, Slack servers
- [ ] Day 7: Document results, fix issues

### Week 2
- [ ] Day 1-2: Llama family testing
- [ ] Day 3: Mistral family testing
- [ ] Day 4: Other models testing
- [ ] Day 5: Performance testing
- [ ] Day 6: Security testing
- [ ] Day 7: Compile test report

### Week 3
- [ ] Fix compatibility issues
- [ ] Update documentation
- [ ] Create model compatibility matrix
- [ ] Prepare beta release

## Test Automation

### Automated Test Suite
```bash
#!/bin/bash
# run_mcp_tests.sh

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Starting MCP Test Suite"

# Test each MCP server
for server in filesystem git postgres python web github slack; do
  echo "Testing $server server..."
  ./test_mcp_server.sh $server
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ $server passed${NC}"
  else
    echo -e "${RED}✗ $server failed${NC}"
  fi
done

# Test each model
for model in qwen2.5:7b llama3.1:8b mistral:7b; do
  echo "Testing $model..."
  ./test_model_compatibility.sh $model
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ $model passed${NC}"
  else
    echo -e "${RED}✗ $model failed${NC}"
  fi
done

echo "Test suite complete"
```

## Success Metrics for Release

### Minimum Requirements (Before Beta)
- [ ] 5+ MCP servers tested and working
- [ ] 3+ model families with MCP support
- [ ] 72 hours stable runtime
- [ ] <100ms tool execution latency
- [ ] Zero critical security issues
- [ ] Basic documentation complete

### Target Goals (Before Production)
- [ ] 10+ MCP servers tested
- [ ] 5+ model families supported
- [ ] 168 hours (1 week) stable runtime
- [ ] Performance benchmarks documented
- [ ] Security audit passed
- [ ] Comprehensive documentation
- [ ] Integration test suite
- [ ] CI/CD pipeline setup

## Reporting

Test results will be documented in:
- `MCP_TEST_RESULTS.md` - Summary of all tests
- `MCP_COMPATIBILITY_MATRIX.md` - Model/server compatibility
- `MCP_PERFORMANCE_REPORT.md` - Performance metrics
- `MCP_SECURITY_AUDIT.md` - Security test results

---
*Last Updated: November 12, 2025*
*Status: Testing Phase Beginning*