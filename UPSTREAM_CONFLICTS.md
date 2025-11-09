# Upstream Conflict Analysis - MCP Integration Branch

## Summary
Analysis of conflicts between our MCP (Model Context Protocol) integration branch and the official ollama repository (upstream/main).

## Divergence Point
- Common ancestor commit: `f67a6df` 
- Upstream has 20+ new commits since divergence
- Our branch adds 416 files (mostly MCP-related + test logs)

## Key Modified Core Files

### 1. api/types.go
**Our changes:**
- Added MCP tool-related types (`ToolCall`, `ToolCallFunction`, `Tool`)
- Modified `Message` struct to include tool calls
- Added support for tool execution results in messages

**Upstream changes:**
- Recent commit `30fcc719`: Added omitempty to required tool function parameters
- Commit `565b802a`: Tool call ID mapping fixes

**Conflict potential:** HIGH - Both branches modify tool-related structures

### 2. cmd/cmd.go  
**Our changes:**
- Implemented `StreamingToolDetector` for better chunk handling
- Added configurable buffer delay (`OLLAMA_TOOL_BUFFER_DELAY`)
- Improved tool output formatting with separate lines for name/args/results
- Added MCP server connection status display

**Upstream changes:**
- Commit `1ca608bc`: Added embeddings command for CLI
- Commit `f89fc1ca`: Fixed connection string for interactive usage

**Conflict potential:** MEDIUM - Different areas of the file, but needs careful merging

### 3. model/parsers/qwen3vl.go
**Our changes:**
- Fixed control token contamination (`<|im_start|>`, `<|im_end|>`)
- Enhanced `fixIncompleteJSON` to strip ChatML tokens
- Improved JSON extraction from XML-wrapped tool calls
- Added stateful accumulation of tool calls during streaming

**Upstream changes:** None detected in this file

**Conflict potential:** LOW - No upstream changes to this file

### 4. server/routes.go
**Our changes:**
- Added MCP manager integration
- Modified tool execution flow to use MCP servers
- Enhanced prompt construction between tool rounds
- Added MCP-specific error handling

**Upstream changes:**
- Tool-related fixes that may interact with our changes
- Stream handling improvements

**Conflict potential:** HIGH - Core routing logic modified by both branches

## New MCP Files (No Conflicts)
These are entirely new files added by our branch:
- `server/mcp_*.go` (8 files) - MCP implementation
- `docs/MCP_*.md` - Documentation
- `examples/mcp-*.json` - Configuration examples
- Various test scripts and logs

## Upstream Changes We Should Incorporate

### Critical Fixes
1. **Tool Call ID Mapping** (`565b802a`) - Important fix for tool call tracking
2. **Tool Parameter Omitempty** (`30fcc719`) - API compatibility improvement
3. **GGML Updates** (`544b6739`) - Backend performance improvements
4. **macOS VRAM Fix** (`6aa72830`) - Memory management fix

### Nice to Have
1. WebP image support (`bddfa210`)
2. Embeddings CLI command (`1ca608bc`)
3. Documentation updates (multiple commits)

## Merge Strategy Recommendation

### Phase 1: Rebase Preparation
1. Create a backup branch: `git checkout -b mcp-backup`
2. Fetch latest upstream: `git fetch upstream`

### Phase 2: Cherry-pick Critical Fixes
Before attempting full merge, cherry-pick critical upstream fixes:
```bash
# Tool call ID mapping
git cherry-pick 565b802a

# Tool parameter omitempty  
git cherry-pick 30fcc719

# GGML updates (may require conflict resolution)
git cherry-pick 544b6739
```

### Phase 3: Merge Resolution
1. **api/types.go**: Merge tool-related changes carefully, ensuring:
   - Our MCP tool types remain intact
   - Upstream's omitempty and ID mapping incorporated
   
2. **cmd/cmd.go**: Keep both sets of changes:
   - Our StreamingToolDetector and formatting
   - Upstream's embeddings command
   
3. **server/routes.go**: Most complex merge:
   - Preserve MCP manager integration
   - Incorporate upstream's tool ID fixes
   - Test thoroughly after merge

### Phase 4: Testing
1. Run existing MCP tests
2. Test tool execution with multiple rounds
3. Verify streaming output still works
4. Check control token stripping in Qwen models

## Potential Issues

1. **Tool Call ID Conflicts**: Upstream added tool call ID mapping which may conflict with our MCP implementation's ID handling

2. **GGML Backend Changes**: The GGML update (`544b6739`) is substantial and may affect our parser modifications

3. **API Compatibility**: Need to ensure our MCP additions don't break Ollama's OpenAI-compatible API

## Next Steps

1. **Create PR-ready branch**: Clean up test logs and temporary files
2. **Write comprehensive tests**: Ensure MCP integration doesn't break existing functionality  
3. **Documentation**: Update README with MCP usage instructions
4. **Consider splitting PR**: Core fixes (parser) vs. MCP feature addition

## Files to Clean Before PR
Remove these test/debug files:
- All `*.log` files (30+ server logs)
- Test scripts (`test_*.sh`, `demo_*.sh`)
- Temporary analysis files (`*_ANALYSIS.md`, `*_STATUS.md`)

## Conclusion

The conflicts are manageable but require careful attention, especially in:
- `api/types.go` - Tool type definitions
- `server/routes.go` - Core routing logic

Our qwen3vl parser fixes are isolated and should merge cleanly. The MCP implementation is largely additive but needs to incorporate upstream's tool improvements for compatibility.