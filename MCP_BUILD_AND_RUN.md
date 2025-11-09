# Building and Running Ollama with MCP Integration

## Quick Start

### Prerequisites

1. **Go 1.22 or later**
   ```bash
   go version  # Should show 1.22+
   ```

2. **GCC/G++ Compiler**
   ```bash
   gcc --version  # Required for CGO
   ```

3. **CUDA Toolkit (for NVIDIA GPU support)**
   ```bash
   nvcc --version  # Optional but recommended
   ```

4. **Node.js 18+ (for MCP servers)**
   ```bash
   node --version  # Should show v18.0.0+
   npx --version   # Required for MCP servers
   ```

### Build Instructions

```bash
# Clone the repository with MCP integration
git clone https://github.com/Code4me2/ollama.git
cd ollama

# Checkout the stable MCP branch
git checkout mcp-multiturn-fix

# Clean any previous builds
go clean -cache

# Build the project
go build -v

# Verify the build
./ollama --version
```

### Running the Server

```bash
# Start the Ollama server
./ollama serve

# Or run with logging
./ollama serve 2>&1 | tee server.log

# Run in background
nohup ./ollama serve > server.log 2>&1 &
```

### Testing MCP Functionality

```bash
# Basic test without MCP
./ollama run qwen2.5:7b
>>> Hello
# Should respond normally

# Test with MCP filesystem tools
./ollama run qwen2.5:7b --tools /tmp
>>> List the files in /tmp
# Should execute filesystem:list_directory tool

# Test via API
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:7b",
    "messages": [{"role": "user", "content": "List files in /home"}],
    "mcp_servers": [{
      "name": "filesystem",
      "command": "npx",
      "args": ["@modelcontextprotocol/server-filesystem", "/home"],
      "env": {}
    }],
    "stream": false
  }'
```

## Detailed Build Options

### Linux Build

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y build-essential cmake git golang-go

# For NVIDIA GPU support
# Install CUDA from https://developer.nvidia.com/cuda-downloads

# Build with optimizations
CGO_ENABLED=1 go build -v -tags cuda

# Build without GPU support
CGO_ENABLED=1 go build -v
```

### macOS Build

```bash
# Install dependencies via Homebrew
brew install go cmake

# Apple Silicon (M1/M2/M3)
go build -v  # Metal support is built-in

# Intel Mac
cmake -B build
cmake --build build
go build -v
```

### Windows Build

```bash
# Install prerequisites:
# - Go from https://go.dev
# - Visual Studio 2022
# - CMake from https://cmake.org
# - CUDA SDK (optional) from NVIDIA

# Build
cmake -B build
cmake --build build --config Release
go build -v
```

## Docker Build

### Basic Dockerfile

```dockerfile
# Dockerfile.mcp
FROM golang:1.22 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy source code
WORKDIR /build
COPY . .

# Build Ollama with MCP
RUN go clean -cache && \
    go build -v -o ollama

# Runtime stage
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    nodejs \
    npm \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Copy built binary
COPY --from=builder /build/ollama /usr/local/bin/ollama

# Create directories
RUN mkdir -p /root/.ollama/models

# Expose port
EXPOSE 11434

# Run Ollama
CMD ["ollama", "serve"]
```

### Build and Run Docker Container

```bash
# Build the Docker image
docker build -f Dockerfile.mcp -t ollama-mcp:latest .

# Run the container
docker run -d \
  --name ollama-mcp \
  -p 11434:11434 \
  -v ~/.ollama:/root/.ollama \
  --gpus all \
  ollama-mcp:latest

# Test the container
docker exec ollama-mcp ollama run qwen2.5:0.5b --tools /tmp
```

### Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  ollama:
    build:
      context: .
      dockerfile: Dockerfile.mcp
    container_name: ollama-mcp
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
      - ./mcp-servers.json:/etc/ollama/mcp-servers.json
    environment:
      - OLLAMA_HOST=0.0.0.0:11434
      - OLLAMA_MODELS=/root/.ollama/models
      - OLLAMA_DEBUG=INFO
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
    restart: unless-stopped

volumes:
  ollama_data:
```

## MCP Server Configuration

### Default MCP Servers Configuration

Create `mcp-servers.json`:

```json
{
  "servers": [
    {
      "name": "filesystem",
      "command": "npx",
      "args": ["@modelcontextprotocol/server-filesystem", "/safe/path"],
      "env": {}
    },
    {
      "name": "git",
      "command": "npx",
      "args": ["@modelcontextprotocol/server-git"],
      "env": {}
    },
    {
      "name": "postgres",
      "command": "npx",
      "args": ["@modelcontextprotocol/server-postgres", "postgresql://localhost/db"],
      "env": {}
    }
  ]
}
```

### Installing MCP Servers

```bash
# Install Node.js MCP servers globally
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-git

# Or use npx (no installation needed)
npx @modelcontextprotocol/server-filesystem --help

# Python MCP servers
pip install mcp-server-python
```

## Environment Variables

```bash
# Server configuration
export OLLAMA_HOST="0.0.0.0:11434"
export OLLAMA_MODELS="/path/to/models"
export OLLAMA_DEBUG="INFO"

# MCP configuration
export OLLAMA_MCP_DEBUG="1"
export OLLAMA_MCP_SERVERS='{"servers":[...]}'
export OLLAMA_MCP_TIMEOUT="30000"

# Command overrides
export OLLAMA_NPX_COMMAND="pnpm dlx"
export OLLAMA_PYTHON_COMMAND="python3.11"
```

## Troubleshooting

### Common Build Issues

1. **CGO errors**
   ```bash
   # Force rebuild native code
   go clean -cache
   go build -v
   ```

2. **Missing dependencies**
   ```bash
   # Check Go version
   go version  # Must be 1.22+
   
   # Check compiler
   gcc --version
   ```

3. **GPU not detected**
   ```bash
   # Check CUDA installation
   nvidia-smi
   nvcc --version
   
   # Rebuild with CUDA support
   go build -v -tags cuda
   ```

### Runtime Issues

1. **MCP server not starting**
   ```bash
   # Check Node.js
   node --version
   npx --version
   
   # Test MCP server directly
   npx @modelcontextprotocol/server-filesystem --help
   ```

2. **Port already in use**
   ```bash
   # Check what's using port 11434
   lsof -i :11434
   
   # Use different port
   OLLAMA_HOST=127.0.0.1:11435 ./ollama serve
   ```

3. **Memory issues**
   ```bash
   # Check available memory
   free -h
   
   # Limit model size
   ./ollama run qwen2.5:0.5b  # Use smaller model
   ```

## Testing the Build

### Unit Tests

```bash
# Run all tests
go test ./...

# Run MCP-specific tests
go test ./server -run TestMCP -v

# Run with race detection
go test -race ./server
```

### Integration Tests

```bash
# Start server for testing
./ollama serve &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Run test commands
./test_mcp_integration.sh

# Stop server
kill $SERVER_PID
```

### Performance Testing

```bash
# Basic benchmark
go test -bench=. ./server

# Load testing with concurrent requests
for i in {1..10}; do
  curl -X POST http://localhost:11434/api/generate \
    -d '{"model":"qwen2.5:0.5b","prompt":"Hello"}' &
done
wait
```

## Deployment Considerations

### System Requirements

- **Minimum**: 8GB RAM, 4 CPU cores
- **Recommended**: 16GB RAM, 8 CPU cores, NVIDIA GPU
- **For large models**: 32GB+ RAM, high-end GPU

### Security Recommendations

1. **Run as non-root user**
   ```bash
   useradd -m -s /bin/bash ollama
   chown -R ollama:ollama /path/to/ollama
   su - ollama -c "./ollama serve"
   ```

2. **Restrict network access**
   ```bash
   # Bind to localhost only
   OLLAMA_HOST=127.0.0.1:11434 ./ollama serve
   ```

3. **Configure MCP security**
   - Use absolute paths for MCP servers
   - Limit filesystem access paths
   - Review mcp-security.json configuration

### Monitoring

```bash
# Check server health
curl http://localhost:11434/api/version

# Monitor logs
tail -f server.log | grep -E "ERROR|WARN"

# Check memory usage
ps aux | grep ollama

# Monitor MCP execution
tail -f server.log | grep "MCP"
```

## Migration from Standard Ollama

If migrating from standard Ollama to MCP-enabled version:

1. **Backup models**
   ```bash
   cp -r ~/.ollama ~/.ollama.backup
   ```

2. **Stop existing Ollama**
   ```bash
   systemctl stop ollama  # If running as service
   pkill ollama           # Or kill process
   ```

3. **Replace binary**
   ```bash
   # Build new version
   go build -v -o ollama-mcp
   
   # Replace existing
   sudo mv ollama-mcp /usr/local/bin/ollama
   ```

4. **Test compatibility**
   ```bash
   ollama list          # Should show existing models
   ollama run llama2    # Test existing model
   ollama run qwen2.5:7b --tools /tmp  # Test MCP
   ```

## Support and Resources

- **Documentation**: See MCP_*.md files in repository
- **Issues**: https://github.com/Code4me2/ollama/issues
- **Original Ollama**: https://github.com/ollama/ollama
- **MCP Specification**: Model Context Protocol documentation

---
*Last updated: November 12, 2025*
*Version: mcp-multiturn-fix branch*