# 72-Hour Stability Test Tracker

## Current Test Session
**Started**: Wednesday, November 12, 2025 at 5:16 PM PST  
**Target End**: Saturday, November 15, 2025 at 5:16 PM PST  
**PID**: 1124291  
**Log File**: stability_test_20251112_1716.log  
**Branch**: mcp-multiturn-fix (cleaned, rebased)  
**Commit**: 8b3d64dd (latest after cleanup)  

## Test Parameters
- **Server Command**: `./ollama serve`
- **Monitoring**: Every 30 minutes via cron job
- **Monitor Script**: mcp_stability_monitor.sh
- **Monitor Logs**: stability_logs/monitor_*.log

## Metrics to Track
- [ ] Memory usage over time
- [ ] CPU usage patterns
- [ ] Error count in logs
- [ ] Warning count in logs
- [ ] Tool execution success rate
- [ ] Response time consistency
- [ ] Process stability (no crashes)
- [ ] Resource leaks

## Checkpoints
- [ ] 1 hour (6:16 PM Nov 12)
- [ ] 6 hours (11:16 PM Nov 12)
- [ ] 12 hours (5:16 AM Nov 13)
- [ ] 24 hours (5:16 PM Nov 13)
- [ ] 36 hours (5:16 AM Nov 14)
- [ ] 48 hours (5:16 PM Nov 14)
- [ ] 60 hours (5:16 AM Nov 15)
- [ ] 72 hours (5:16 PM Nov 15)

## Test Commands to Run Periodically
```bash
# Basic health check
curl -s http://localhost:11434/api/version

# Simple completion
echo "Hello" | ./ollama run qwen2.5:0.5b

# MCP tool test
echo "List files in /tmp" | ./ollama run qwen2.5:0.5b --tools /tmp

# Memory check
ps aux | grep ollama | grep -v grep

# Log error check
tail -1000 stability_test_20251112_1716.log | grep -c ERROR

# Log warning check  
tail -1000 stability_test_20251112_1716.log | grep -c WARN
```

## Notes
- Server restarted after successful commit cleanup and rebase
- Previous test ran for ~1.5 hours without issues
- Monitoring script running via cron every 30 minutes
- This is the official 72-hour stability test before release consideration

## Incidents Log
(Record any issues, restarts, or anomalies here)

---
*Update this document as the test progresses*