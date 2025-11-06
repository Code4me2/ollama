# MCP Tool Context Injection System

## Problem Statement

When MCP tools are initialized with specific constraints (e.g., filesystem tools restricted to certain directories), the language model is unaware of these restrictions. This leads to the model attempting invalid operations that fail at execution time.

## Current Behavior

- Model receives tool definitions but not their runtime constraints
- Path restrictions, authentication requirements, and other constraints are enforced at execution but not communicated to the model
- Results in failed tool calls and poor user experience

## Proposed Solution: Tool Context Metadata

### 1. Tool Metadata Enhancement

Each MCP tool should provide metadata about its constraints and usage context:

```go
type ToolMetadata struct {
    Constraints  map[string]string `json:"constraints,omitempty"`
    Examples     []string          `json:"examples,omitempty"`
    DefaultArgs  map[string]interface{} `json:"default_args,omitempty"`
    UsageHints   string            `json:"usage_hints,omitempty"`
}
```

### 2. MCP Server Context Provider

MCP servers should provide context about their configuration:

```go
type MCPServerContext struct {
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Constraints  map[string]string `json:"constraints"`
    Capabilities []string          `json:"capabilities"`
}
```

### 3. Automatic Context Injection

When tools are loaded, their context should be automatically injected into the conversation:

```go
func InjectToolContext(tools []api.Tool, serverContexts []MCPServerContext) string {
    var context strings.Builder
    
    context.WriteString("Available tools and their constraints:\n\n")
    
    for _, ctx := range serverContexts {
        context.WriteString(fmt.Sprintf("- %s: %s\n", ctx.Name, ctx.Description))
        if len(ctx.Constraints) > 0 {
            context.WriteString("  Constraints:\n")
            for key, value := range ctx.Constraints {
                context.WriteString(fmt.Sprintf("    %s: %s\n", key, value))
            }
        }
    }
    
    return context.String()
}
```

## Implementation Strategy

### Phase 1: MCP Protocol Extension
- Extend MCP protocol to include `context/describe` method
- Returns server-specific constraints and usage information
- Backward compatible - servers without this method continue to work

### Phase 2: Context Collection
- During MCP server initialization, query for context
- Store context alongside tool definitions
- Cache context for performance

### Phase 3: Smart Injection
- Inject context based on model capabilities
- For models with large context windows, include full context
- For limited models, include only essential constraints
- Use token counting to ensure context doesn't overwhelm the model

### Phase 4: Dynamic Adaptation
- Monitor tool execution failures
- Learn from failures and adjust context injection
- Build a feedback loop for context improvement

## Example Implementation

### Filesystem MCP Server Context
```json
{
  "name": "filesystem",
  "description": "File system operations",
  "constraints": {
    "allowed_paths": ["/home/user/documents"],
    "forbidden_operations": ["delete", "execute"],
    "max_file_size": "10MB"
  },
  "usage_hints": "When accessing files, use paths within the allowed directories. Paths outside these directories will be rejected."
}
```

### Database MCP Server Context
```json
{
  "name": "postgres",
  "description": "PostgreSQL database access",
  "constraints": {
    "database": "analytics",
    "read_only": true,
    "allowed_tables": ["users", "events", "metrics"]
  },
  "usage_hints": "Only SELECT queries are allowed. Modifications will be rejected."
}
```

## Benefits

1. **Better Tool Usage**: Models understand constraints before attempting operations
2. **Reduced Failures**: Fewer invalid tool calls
3. **Improved UX**: Users get better responses on first attempt
4. **Extensible**: New constraints can be added without code changes
5. **Protocol-Level**: Works with any MCP-compliant server

## Migration Path

1. **Stage 1**: Implement basic context injection for known servers (filesystem, git, etc.)
2. **Stage 2**: Add protocol extension for dynamic context discovery
3. **Stage 3**: Machine learning optimization of context injection
4. **Stage 4**: Full automation with minimal configuration

## Configuration Example

```yaml
mcp:
  context_injection:
    enabled: true
    mode: "smart"  # full, minimal, smart
    max_tokens: 500
    include_examples: true
    
  servers:
    filesystem:
      context:
        constraints:
          allowed_path: "${TOOLS_PATH}"
        examples:
          - "List files in the current directory"
          - "Read the README.md file"
```

## Testing Strategy

1. **Unit Tests**: Context generation and injection
2. **Integration Tests**: End-to-end tool execution with context
3. **Benchmark Tests**: Performance impact of context injection
4. **User Tests**: Improved success rate metrics

## Future Enhancements

1. **Context Templates**: Customizable templates for different model families
2. **Multi-Language Support**: Context in user's preferred language
3. **Visual Context**: For multimodal models, include visual representations
4. **Interactive Context**: Allow models to query for more context as needed
5. **Context Compression**: Optimize context size using compression techniques

## Conclusion

This generalized approach to tool context injection solves the immediate problem while providing a foundation for future enhancements. It maintains backward compatibility while enabling smarter tool usage across all MCP servers.