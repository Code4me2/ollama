#!/bin/bash

# MCP Stability Monitor Script
# This script periodically tests the ollama server with MCP functionality
# Run via cron every 30 minutes: */30 * * * * /path/to/mcp_stability_monitor.sh

LOG_DIR="/home/velvetm/Desktop/ollama/stability_logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/monitor_${TIMESTAMP}.log"

echo "=== MCP Stability Monitor ===" >> "$LOG_FILE"
echo "Timestamp: $(date)" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Test 1: Check if server is responding
echo "Test 1: Server Health Check" >> "$LOG_FILE"
if curl -s http://localhost:11434/api/version > /dev/null 2>&1; then
    echo "✓ Server is responding" >> "$LOG_FILE"
    VERSION=$(curl -s http://localhost:11434/api/version)
    echo "  Version: $VERSION" >> "$LOG_FILE"
else
    echo "✗ Server not responding - attempting restart" >> "$LOG_FILE"
    pkill ollama
    sleep 2
    nohup /home/velvetm/Desktop/ollama/ollama serve > /home/velvetm/Desktop/ollama/stability_test_restart_${TIMESTAMP}.log 2>&1 &
    sleep 5
    if curl -s http://localhost:11434/api/version > /dev/null 2>&1; then
        echo "  Server restarted successfully" >> "$LOG_FILE"
    else
        echo "  CRITICAL: Server restart failed" >> "$LOG_FILE"
        exit 1
    fi
fi

# Test 2: Memory usage check
echo "" >> "$LOG_FILE"
echo "Test 2: Memory Usage" >> "$LOG_FILE"
OLLAMA_PID=$(pgrep -f "ollama serve")
if [ ! -z "$OLLAMA_PID" ]; then
    MEM_INFO=$(ps -o pid,vsz,rss,comm -p $OLLAMA_PID)
    echo "$MEM_INFO" >> "$LOG_FILE"
    
    # Check for memory leak (RSS > 2GB warning)
    RSS=$(ps -o rss= -p $OLLAMA_PID)
    RSS_GB=$((RSS / 1024 / 1024))
    if [ "$RSS_GB" -gt 2 ]; then
        echo "  WARNING: High memory usage detected (${RSS_GB}GB)" >> "$LOG_FILE"
    fi
fi

# Test 3: Simple completion test (no tools)
echo "" >> "$LOG_FILE"
echo "Test 3: Basic Completion Test" >> "$LOG_FILE"
RESPONSE=$(curl -s -X POST http://localhost:11434/api/chat \
    -H "Content-Type: application/json" \
    -d '{
        "model": "qwen2.5:7b",
        "messages": [{"role": "user", "content": "Say hello"}],
        "stream": false
    }' 2>&1)

if echo "$RESPONSE" | grep -q "message"; then
    echo "✓ Basic completion working" >> "$LOG_FILE"
else
    echo "✗ Basic completion failed" >> "$LOG_FILE"
    echo "  Response: $RESPONSE" >> "$LOG_FILE"
fi

# Test 4: MCP tools endpoint check
echo "" >> "$LOG_FILE"
echo "Test 4: MCP Tools Endpoint" >> "$LOG_FILE"
TOOLS_RESPONSE=$(curl -s -X GET http://localhost:11434/api/tools 2>&1)
if [ ! -z "$TOOLS_RESPONSE" ]; then
    echo "✓ Tools endpoint accessible" >> "$LOG_FILE"
else
    echo "✗ Tools endpoint not responding" >> "$LOG_FILE"
fi

# Test 5: Check for error patterns in server log
echo "" >> "$LOG_FILE"
echo "Test 5: Error Pattern Check" >> "$LOG_FILE"
LATEST_LOG=$(ls -t /home/velvetm/Desktop/ollama/stability_test_*.log 2>/dev/null | head -1)
if [ -f "$LATEST_LOG" ]; then
    ERROR_COUNT=$(tail -n 1000 "$LATEST_LOG" | grep -c -E "ERROR|PANIC|FATAL" || true)
    WARN_COUNT=$(tail -n 1000 "$LATEST_LOG" | grep -c "WARN" || true)
    echo "  Errors in last 1000 lines: $ERROR_COUNT" >> "$LOG_FILE"
    echo "  Warnings in last 1000 lines: $WARN_COUNT" >> "$LOG_FILE"
    
    if [ "$ERROR_COUNT" -gt 10 ]; then
        echo "  WARNING: High error count detected" >> "$LOG_FILE"
        tail -n 20 "$LATEST_LOG" | grep -E "ERROR|PANIC|FATAL" >> "$LOG_FILE" 2>&1 || true
    fi
fi

# Summary
echo "" >> "$LOG_FILE"
echo "=== Monitor Complete ===" >> "$LOG_FILE"
echo "End: $(date)" >> "$LOG_FILE"

# Keep only last 100 monitor logs
ls -t "$LOG_DIR"/monitor_*.log | tail -n +101 | xargs -r rm

exit 0