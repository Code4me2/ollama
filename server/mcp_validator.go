package server

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// MCPValidator validates MCP installation and requirements
type MCPValidator struct {
	resolver *CommandResolver
	warnings []string
	errors   []string
}

// NewMCPValidator creates a new MCP validator
func NewMCPValidator() *MCPValidator {
	return &MCPValidator{
		resolver: globalCommandResolver,
		warnings: []string{},
		errors:   []string{},
	}
}

// ValidateOnStartup performs validation checks for MCP functionality
func (v *MCPValidator) ValidateOnStartup() {
	// Only validate if MCP is being used
	if os.Getenv("OLLAMA_MCP_DISABLE") == "1" {
		return
	}

	// Check for MCP server configurations
	v.checkMCPConfiguration()

	// Check system requirements
	v.checkSystemRequirements()

	// Log results
	v.logValidationResults()
}

// checkMCPConfiguration validates MCP server configuration
func (v *MCPValidator) checkMCPConfiguration() {
	registry, err := LoadMCPRegistry()
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("Failed to load MCP registry: %v", err))
		return
	}

	if len(registry.Servers) == 0 {
		v.warnings = append(v.warnings, 
			"No MCP servers configured. To use MCP features, configure servers in ~/.ollama/mcp-servers.json or see examples/mcp-servers.json")
	}
}

// checkSystemRequirements validates system dependencies
func (v *MCPValidator) checkSystemRequirements() {
	requirements := v.resolver.GetSystemRequirements()

	// Check Node.js (optional but recommended)
	if nodeReq, ok := requirements["node"].(map[string]interface{}); ok {
		if nodeReq["status"] == "not found" {
			v.warnings = append(v.warnings, 
				"Node.js not found. Install Node.js to use JavaScript-based MCP servers")
		}
	}

	// Check Python (optional but recommended)
	if pythonReq, ok := requirements["python"].(map[string]interface{}); ok {
		if pythonReq["status"] == "not found" {
			v.warnings = append(v.warnings,
				"Python not found. Install Python 3.8+ to use Python-based MCP servers")
		}
	}

	// Check package manager
	if pmReq, ok := requirements["package_manager"].(map[string]interface{}); ok {
		if pmReq["status"] == "not found" {
			v.warnings = append(v.warnings,
				"No Node.js package manager found (npx, pnpm, yarn, bunx). Install one to use npm-based MCP servers")
		}
	}

	// Check for environment variable overrides
	v.checkEnvironmentOverrides()
}

// checkEnvironmentOverrides validates environment variable settings
func (v *MCPValidator) checkEnvironmentOverrides() {
	// Check Python override
	if pythonCmd := os.Getenv("OLLAMA_PYTHON_COMMAND"); pythonCmd != "" {
		if _, err := v.resolver.checkCommand(pythonCmd); err != nil {
			v.warnings = append(v.warnings,
				fmt.Sprintf("OLLAMA_PYTHON_COMMAND set to '%s' but command not found", pythonCmd))
		}
	}

	// Check NPX override
	if npxCmd := os.Getenv("OLLAMA_NPX_COMMAND"); npxCmd != "" {
		// Split command in case it's like "pnpm dlx"
		parts := strings.Fields(npxCmd)
		if _, err := v.resolver.checkCommand(parts[0]); err != nil {
			v.warnings = append(v.warnings,
				fmt.Sprintf("OLLAMA_NPX_COMMAND set to '%s' but command not found", npxCmd))
		}
	}

	// Check Node override
	if nodeCmd := os.Getenv("OLLAMA_NODE_COMMAND"); nodeCmd != "" {
		if _, err := v.resolver.checkCommand(nodeCmd); err != nil {
			v.warnings = append(v.warnings,
				fmt.Sprintf("OLLAMA_NODE_COMMAND set to '%s' but command not found", nodeCmd))
		}
	}
}

// logValidationResults logs validation warnings and errors
func (v *MCPValidator) logValidationResults() {
	// Log errors (critical issues)
	for _, err := range v.errors {
		slog.Error("MCP validation error", "error", err)
	}

	// Log warnings (non-critical issues)
	for _, warning := range v.warnings {
		slog.Warn("MCP validation warning", "warning", warning)
	}

	// Log summary
	if len(v.errors) == 0 && len(v.warnings) == 0 {
		slog.Info("MCP validation completed successfully")
	} else if len(v.errors) > 0 {
		slog.Error("MCP validation failed", 
			"errors", len(v.errors),
			"warnings", len(v.warnings),
			"note", "MCP features may not work correctly")
	} else {
		slog.Info("MCP validation completed with warnings",
			"warnings", len(v.warnings),
			"note", "Some MCP features may be limited")
	}
}

// GetValidationReport returns a detailed validation report
func (v *MCPValidator) GetValidationReport() map[string]interface{} {
	requirements := v.resolver.GetSystemRequirements()
	registry, _ := LoadMCPRegistry()

	return map[string]interface{}{
		"status": map[string]interface{}{
			"errors":   v.errors,
			"warnings": v.warnings,
			"valid":    len(v.errors) == 0,
		},
		"requirements": requirements,
		"configuration": map[string]interface{}{
			"servers_configured": len(registry.Servers),
			"config_paths": []string{
				"~/.ollama/mcp-servers.json",
				"/etc/ollama/mcp-servers.json",
				"./mcp-servers.json",
				"./examples/mcp-servers.json",
			},
		},
		"environment": map[string]interface{}{
			"OLLAMA_MCP_DISABLE":      os.Getenv("OLLAMA_MCP_DISABLE"),
			"OLLAMA_PYTHON_COMMAND":   os.Getenv("OLLAMA_PYTHON_COMMAND"),
			"OLLAMA_NPX_COMMAND":      os.Getenv("OLLAMA_NPX_COMMAND"),
			"OLLAMA_NODE_COMMAND":     os.Getenv("OLLAMA_NODE_COMMAND"),
			"OLLAMA_MCP_SERVERS":      os.Getenv("OLLAMA_MCP_SERVERS") != "",
		},
	}
}

// RunStartupValidation is called during server initialization
func RunStartupValidation() {
	validator := NewMCPValidator()
	validator.ValidateOnStartup()
}