# MCP (Model Context Protocol) System Requirements

## Overview

The MCP integration in Ollama allows language models to execute tools and interact with external systems. To use MCP features, certain system requirements must be met.

## Core Requirements

### Ollama Base Requirements
- Ollama v0.5.0 or later
- Supported operating system (Linux, macOS, Windows)
- Sufficient RAM for model execution (varies by model)

## Optional MCP Server Requirements

MCP servers are external tools that provide specific capabilities. Each type has its own requirements:

### JavaScript/TypeScript MCP Servers

For servers like `@modelcontextprotocol/server-filesystem`, `@modelcontextprotocol/server-git`:

**Requirements:**
- Node.js 18.0+ or compatible runtime
- Package manager (one of):
  - npm with npx
  - pnpm with dlx support
  - yarn 2+ with dlx support
  - bun with bunx

**Installation:**
```bash
# Node.js (if not installed)
# Ubuntu/Debian:
sudo apt-get install nodejs npm

# macOS:
brew install node

# Windows:
# Download from https://nodejs.org

# Verify installation
node --version
npx --version
```

### Python MCP Servers

For servers like `mcp_server_python`:

**Requirements:**
- Python 3.8 or later
- pip package manager

**Installation:**
```bash
# Ubuntu/Debian:
sudo apt-get install python3 python3-pip

# macOS:
brew install python3

# Windows:
# Download from https://python.org

# Verify installation
python3 --version
pip3 --version

# Install MCP server package (example)
pip install mcp-server-python
```

## Environment Configuration

### Command Resolution

Ollama automatically detects available commands. You can override detection with environment variables:

```bash
# Override Python command
export OLLAMA_PYTHON_COMMAND="python3.11"

# Override package manager
export OLLAMA_NPX_COMMAND="pnpm dlx"

# Override Node.js
export OLLAMA_NODE_COMMAND="/usr/local/bin/node"
```

### MCP Server Configuration

MCP servers are configured via JSON files. Location priority:

1. `~/.ollama/mcp-servers.json` (user configuration)
2. `/etc/ollama/mcp-servers.json` (system configuration)
3. `./mcp-servers.json` (local configuration)
4. `./examples/mcp-servers.json` (example configuration)

Example configuration:
```json
{
  "servers": [
    {
      "name": "filesystem",
      "description": "File system operations",
      "command": "npx",
      "args": ["@modelcontextprotocol/server-filesystem"],
      "requires_path": true,
      "capabilities": ["read", "write", "list", "search"]
    }
  ]
}
```

You can also set configuration via environment variable:
```bash
export OLLAMA_MCP_SERVERS='{"servers":[...]}'
```

## Security Considerations

### Process Isolation
- MCP servers run in separate process groups
- Sensitive environment variables are filtered
- Dangerous commands are blocked

### Restricted Commands
The following commands cannot be used as MCP servers:
- Shell interpreters (bash, sh, zsh, cmd, powershell)
- System utilities (sudo, su, rm, dd, format)
- Network tools (curl, wget, nc, telnet)
- Script interpreters (eval, exec)

### Environment Filtering
Sensitive environment variables are automatically removed:
- AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- API tokens (GITHUB_TOKEN, GITLAB_TOKEN, etc.)
- Database credentials
- SSH keys and passwords

## Verification

### Check System Requirements

Run this command to check if your system meets MCP requirements:

```bash
# Check Node.js
node --version || echo "Node.js not found"

# Check Python
python3 --version || python --version || echo "Python not found"

# Check package managers
npx --version || echo "npx not found"
pnpm --version || echo "pnpm not found"
yarn --version || echo "yarn not found"

# Check Ollama MCP support
ollama --version | grep -q "MCP" && echo "MCP support available" || echo "MCP support not available"
```

### Test MCP Server

To test if MCP is working correctly:

```bash
# List available MCP servers
curl http://localhost:11434/api/tools

# Test with a model that supports tools
ollama run qwen2.5:latest --tools filesystem:/tmp
```

## Troubleshooting

### Common Issues

1. **"npx not found"**
   - Install Node.js and npm
   - Or set `OLLAMA_NPX_COMMAND` to an alternative

2. **"python not found"**
   - Install Python 3.8+
   - Or set `OLLAMA_PYTHON_COMMAND` to your Python path

3. **"MCP server failed to start"**
   - Check the server command exists
   - Verify required packages are installed
   - Check file permissions

4. **"No tools available"**
   - Ensure MCP servers are configured
   - Verify the model supports tool calling
   - Check server logs for errors

### Debug Mode

Enable debug logging for MCP:
```bash
export OLLAMA_MCP_DEBUG=1
ollama serve
```

## Platform-Specific Notes

### Linux
- Most distributions include Python 3
- Node.js may need to be installed separately
- Use package manager for dependencies

### macOS
- Python 3 included in recent versions
- Use Homebrew for Node.js
- Gatekeeper may block unsigned MCP servers

### Windows
- Use official installers for Python and Node.js
- Run in Administrator mode if needed
- Path separators use backslash

## Further Resources

- [MCP Protocol Specification](https://github.com/modelcontextprotocol/specification)
- [Ollama Documentation](https://github.com/ollama/ollama/blob/main/docs/README.md)
- [Example MCP Servers](./examples/mcp-servers.json)