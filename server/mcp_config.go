package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ollama/ollama/api"
)

// MCPServerRegistry holds available MCP server configurations
type MCPServerRegistry struct {
	Servers map[string]MCPServerDefinition `json:"servers"`
}

// MCPServerDefinition defines an available MCP server type
type MCPServerDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args,omitempty"`
	RequiresPath bool                  `json:"requires_path,omitempty"`
	PathArgIndex int                   `json:"path_arg_index,omitempty"`
	Env         map[string]string      `json:"env,omitempty"`
	Capabilities []string              `json:"capabilities,omitempty"`
}

// DefaultMCPServers returns minimal built-in MCP server definitions
// Full examples are provided in examples/mcp-servers.json
func DefaultMCPServers() map[string]MCPServerDefinition {
	// Return empty map by default - users should configure their own servers
	// This ensures the open source distribution doesn't make assumptions
	// about available MCP servers in the user's environment
	return map[string]MCPServerDefinition{}
}

// LoadMCPRegistry loads MCP server configurations from various sources
func LoadMCPRegistry() (*MCPServerRegistry, error) {
	registry := &MCPServerRegistry{
		Servers: DefaultMCPServers(),
	}

	// Load from user config if exists
	// Priority order: user config > system config > example > defaults
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".ollama", "mcp-servers.json"),
		"/etc/ollama/mcp-servers.json",
		"./mcp-servers.json",
		"./examples/mcp-servers.json", // Example configuration as fallback
	}

	for _, path := range configPaths {
		if err := registry.LoadFromFile(path); err == nil {
			break // Successfully loaded custom config
		}
	}

	// Load from environment variable if set
	if mcpConfig := os.Getenv("OLLAMA_MCP_SERVERS"); mcpConfig != "" {
		if err := registry.LoadFromJSON([]byte(mcpConfig)); err != nil {
			return nil, fmt.Errorf("failed to parse OLLAMA_MCP_SERVERS: %w", err)
		}
	}

	return registry, nil
}

// LoadFromFile loads additional MCP server definitions from a JSON file
func (r *MCPServerRegistry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return r.LoadFromJSON(data)
}

// LoadFromJSON loads MCP server definitions from JSON data
func (r *MCPServerRegistry) LoadFromJSON(data []byte) error {
	var config struct {
		Servers []MCPServerDefinition `json:"servers"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for _, server := range config.Servers {
		r.Servers[server.Name] = server
	}

	return nil
}

// GetServer returns a configured MCP server for use
func (r *MCPServerRegistry) GetServer(name string, options map[string]string) (*api.MCPServerConfig, error) {
	def, exists := r.Servers[name]
	if !exists {
		return nil, fmt.Errorf("MCP server '%s' not found in registry", name)
	}

	// Resolve the command using the command resolver
	resolvedCommand := globalCommandResolver.ResolveForEnvironment(def.Command)
	
	config := &api.MCPServerConfig{
		Name:    def.Name,
		Command: resolvedCommand,
		Args:    append([]string{}, def.Args...), // Copy args
		Env:     make(map[string]string),
	}

	// Copy environment variables
	for k, v := range def.Env {
		config.Env[k] = v
	}

	// Apply options (like paths or connection strings)
	if def.RequiresPath {
		path, ok := options["path"]
		if !ok {
			return nil, fmt.Errorf("MCP server '%s' requires a path", name)
		}
		
		// Validate path exists and is accessible
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("invalid path for MCP server '%s': %w", name, err)
		}

		// Add path to args at specified position
		if def.PathArgIndex < 0 {
			config.Args = append(config.Args, path)
		} else {
			// Insert at specific position
			config.Args = append(config.Args[:def.PathArgIndex], 
				append([]string{path}, config.Args[def.PathArgIndex:]...)...)
		}
	}

	// Apply any additional options to environment
	for k, v := range options {
		if k != "path" && strings.HasPrefix(k, "env_") {
			config.Env[strings.TrimPrefix(k, "env_")] = v
		}
	}

	return config, nil
}

// ListAvailableServers returns information about all registered MCP servers
func (r *MCPServerRegistry) ListAvailableServers() []MCPServerInfo {
	var servers []MCPServerInfo
	for _, def := range r.Servers {
		servers = append(servers, MCPServerInfo{
			Name:         def.Name,
			Description:  def.Description,
			RequiresPath: def.RequiresPath,
			Capabilities: def.Capabilities,
		})
	}
	return servers
}

// MCPServerInfo provides information about an available MCP server
type MCPServerInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RequiresPath bool     `json:"requires_path"`
	Capabilities []string `json:"capabilities,omitempty"`
}