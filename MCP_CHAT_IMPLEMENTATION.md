# MCP Interactive Chat Implementation

## Overview
The MCP (Model Context Protocol) tools integration in Ollama enables interactive chat sessions where models can execute tools in real-time, handle errors gracefully, and maintain conversation context across multiple tool calls.

## Current Status: ✅ READY FOR USE

The implementation is fully functional in commit `98855356` with:
- Interactive chat mode with `--tools` flag
- Multi-turn tool execution
- Comprehensive error handling
- Real-time streaming feedback

## Command Syntax

```bash
# Basic usage
ollama run <model> --tools <path-or-config>

# Examples
ollama run qwen2.5:0.5b --tools /path/to/directory     # Filesystem tools
ollama run llama3:8b --tools filesystem:/home/docs     # Explicit type
ollama run codellama --tools git:.                     # Git tools
ollama run mixtral --tools python                      # Python execution
ollama run model --tools custom-config.json            # Custom config
```

## Supported Tool Types

1. **filesystem** - File operations within specified directory
   - list_directory, read_file, write_file, edit_file
   - create_directory, delete_file, move_file
   - search_files, directory_tree, get_file_info

2. **git** - Git repository operations (planned)
3. **python** - Python code execution (planned)
4. **custom** - User-defined MCP servers

## Error Handling

### 1. Invalid Tool Parameters
```
>>> Edit test.txt with wrong parameters
🔧 Executing tool 'edit_file' with arguments: {"file":"test.txt"}
❌ Tool 'edit_file' failed: Invalid tool parameters - expected 'path' and 'edits'
[Model automatically retries with correct parameters]
```

### 2. Access Denied
```
>>> Read /etc/passwd
🔧 Executing tool 'read_file' with arguments: {"path":"/etc/passwd"}
❌ Tool 'read_file' failed: Access denied - path outside allowed directories
```

### 3. File Not Found
```
>>> Read nonexistent.txt
🔧 Executing tool 'read_file' with arguments: {"path":"nonexistent.txt"}
❌ Tool 'read_file' failed: File not found
```

### 4. Tool Execution Failure
```
>>> Delete protected file
🔧 Executing tool 'delete_file' with arguments: {"path":"protected.txt"}
❌ Tool 'delete_file' failed: Permission denied
```

## Key Features

### Multi-Turn Execution
Models can execute multiple tools in a single response:
```
User: Read test.txt and create a summary
Model: [Executes read_file, then write_file with summary]
```

### Error Recovery
Models interpret errors and suggest alternatives:
```
User: Access /root/secret
Model: ❌ Access denied. I can only work within /allowed/path.
       Would you like me to list available files there instead?
```

### Streaming Feedback
Real-time indicators during tool execution:
- `🔧 Executing tool...` - Tool in progress
- `✅ Tool result:` - Successful execution
- `❌ Tool failed:` - Error occurred

## Implementation Details

### Routes (server/routes.go)
- ChatHandler initializes MCPManager from request
- ExecuteToolsParallel for concurrent execution
- Error results passed as tool_result messages
- Maintains conversation history with tool calls

### MCP Manager (server/mcp_manager.go)
- Manages multiple MCP server connections
- Routes tools to appropriate servers
- Caches results for 5 minutes
- Validates security boundaries

### Command Interface (cmd/cmd.go)
- Parses --tools flag
- Creates MCPServerConfig
- Passes to chat/generate endpoints

## Test Scripts

### 1. test_chat_tools.sh
Automated test scenarios for error handling and multi-tool execution

### 2. interactive_chat_demo.sh
Interactive demonstration with colored output and examples

## Usage Examples

### Basic File Operations
```bash
$ ./ollama run qwen2.5:0.5b --tools /home/user/project

>>> List all Python files
🔧 Executing tool 'search_files' with arguments: {"pattern":"*.py"}
✅ Found: main.py, utils.py, test.py

>>> Read main.py and explain what it does
🔧 Executing tool 'read_file' with arguments: {"path":"main.py"}
✅ [Shows file content and explanation]
```

### Error Handling Demo
```bash
>>> Try to read /etc/shadow
❌ Access denied - outside allowed directory

>>> Create ../outside/file.txt
❌ Access denied - path traversal not allowed

>>> Edit nonexistent.txt
❌ File not found. Would you like me to create it?
```

### Multi-Tool Workflow
```bash
>>> Analyze all .md files and create a summary
🔧 Executing tool 'search_files' with arguments: {"pattern":"*.md"}
🔧 Executing tool 'read_file' with arguments: {"path":"README.md"}
🔧 Executing tool 'read_file' with arguments: {"path":"DOCS.md"}
🔧 Executing tool 'write_file' with arguments: {"path":"summary.md", ...}
✅ Created summary.md with analysis of 2 markdown files
```

## Security Considerations

1. **Path Validation**: All paths validated within allowed directory
2. **Command Injection**: Arguments properly escaped
3. **Resource Limits**: Tool execution timeouts
4. **Result Caching**: Prevents abuse through repeated calls

## Next Steps

1. **Additional Tool Types**:
   - Git integration for repository operations
   - Python execution sandbox
   - Web fetch capabilities
   - Database queries

2. **Enhanced Error Handling**:
   - Retry logic with backoff
   - Error categorization
   - Suggested fixes

3. **UI Improvements**:
   - Progress bars for long operations
   - Syntax highlighting for code
   - Interactive tool selection

## Testing

Run the test script to verify all scenarios:
```bash
chmod +x test_chat_tools.sh
./test_chat_tools.sh
```

For interactive testing:
```bash
chmod +x interactive_chat_demo.sh
./interactive_chat_demo.sh
```

## Conclusion

The MCP tools integration is production-ready with comprehensive error handling that allows models to:
- Recover from errors gracefully
- Suggest alternatives when operations fail
- Maintain conversation context
- Execute multiple tools efficiently

The implementation follows security best practices and provides a robust foundation for AI-assisted file operations and automation.