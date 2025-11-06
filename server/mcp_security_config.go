package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// MCPSecurityConfig defines security policies for MCP servers
type MCPSecurityConfig struct {
	// Commands that are never allowed as MCP servers
	BlockedCommands []string `json:"blocked_commands"`
	
	// Shell metacharacters that are not allowed in arguments
	BlockedMetacharacters []string `json:"blocked_metacharacters"`
	
	// Environment variables that should be filtered
	FilteredEnvironmentVars []string `json:"filtered_environment_vars"`
	
	// Patterns for sensitive environment variables (supports wildcards)
	FilteredEnvironmentPatterns []string `json:"filtered_environment_patterns"`
	
	// Additional allowed commands (overrides defaults)
	AllowedCommands []string `json:"allowed_commands,omitempty"`
	
	// Strict mode - only allow explicitly listed commands
	StrictMode bool `json:"strict_mode,omitempty"`
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *MCPSecurityConfig {
	return &MCPSecurityConfig{
		BlockedCommands: []string{
			// Shells
			"sh", "bash", "zsh", "fish", "csh", "ksh", "dash", "tcsh",
			"cmd", "cmd.exe", "powershell", "powershell.exe", "pwsh", "pwsh.exe",
			
			// System commands
			"sudo", "su", "doas", "runas", "pkexec",
			"rm", "del", "rmdir", "format", "dd", "shred",
			"kill", "killall", "pkill", "shutdown", "reboot",
			"systemctl", "service", "init",
			
			// Network tools
			"curl", "wget", "nc", "netcat", "telnet", "ssh", "scp", "sftp",
			"nmap", "ping", "traceroute", "dig", "nslookup",
			
			// Script interpreters
			"eval", "exec", "source", ".",
			"perl", "ruby", "php", "lua", "tcl",
			
			// File manipulation
			"chmod", "chown", "chgrp", "mount", "umount",
			"ln", "mkfifo", "mknod",
			
			// Package managers (prevent system modification)
			"apt", "apt-get", "yum", "dnf", "pacman", "zypper",
			"brew", "port", "snap", "flatpak",
		},
		
		BlockedMetacharacters: []string{
			";", "|", "&", "$(", "`", ">", "<", ">>", "<<",
			"||", "&&", "\n", "\r", "$", "!", "*", "?",
		},
		
		FilteredEnvironmentVars: []string{
			// AWS
			"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
			
			// Cloud providers
			"GOOGLE_APPLICATION_CREDENTIALS", "AZURE_CLIENT_SECRET",
			
			// API Keys
			"GITHUB_TOKEN", "GITLAB_TOKEN", "OPENAI_API_KEY", "ANTHROPIC_API_KEY",
			
			// Database
			"DATABASE_URL", "DB_PASSWORD", "MYSQL_ROOT_PASSWORD", "POSTGRES_PASSWORD",
			
			// Authentication
			"JWT_SECRET", "SESSION_SECRET", "AUTH_TOKEN", "API_KEY", "API_SECRET",
			
			// SSH
			"SSH_AUTH_SOCK", "SSH_AGENT_PID",
		},
		
		FilteredEnvironmentPatterns: []string{
			"*_TOKEN", "*_SECRET", "*_PASSWORD", "*_KEY", "*_CREDENTIALS",
			"*_AUTH", "*_APIKEY", "*_PASS", "*PWD*",
		},
	}
}

// LoadSecurityConfig loads security configuration from file or environment
func LoadSecurityConfig() (*MCPSecurityConfig, error) {
	config := DefaultSecurityConfig()
	
	// Check for custom security config file
	configPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".ollama", "mcp-security.json"),
		"/etc/ollama/mcp-security.json",
		"./mcp-security.json",
	}
	
	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			var customConfig MCPSecurityConfig
			if err := json.Unmarshal(data, &customConfig); err == nil {
				// Merge with defaults (custom config adds to defaults, doesn't replace)
				config.mergeWith(&customConfig)
				break
			}
		}
	}
	
	// Check for environment variable override
	if securityJSON := os.Getenv("OLLAMA_MCP_SECURITY"); securityJSON != "" {
		var envConfig MCPSecurityConfig
		if err := json.Unmarshal([]byte(securityJSON), &envConfig); err == nil {
			config.mergeWith(&envConfig)
		}
	}
	
	return config, nil
}

// mergeWith merges another security config into this one
func (c *MCPSecurityConfig) mergeWith(other *MCPSecurityConfig) {
	if other.StrictMode {
		c.StrictMode = true
	}
	
	// Add additional blocked commands
	if len(other.BlockedCommands) > 0 {
		c.BlockedCommands = append(c.BlockedCommands, other.BlockedCommands...)
	}
	
	// Add additional blocked metacharacters
	if len(other.BlockedMetacharacters) > 0 {
		c.BlockedMetacharacters = append(c.BlockedMetacharacters, other.BlockedMetacharacters...)
	}
	
	// Add additional filtered environment variables
	if len(other.FilteredEnvironmentVars) > 0 {
		c.FilteredEnvironmentVars = append(c.FilteredEnvironmentVars, other.FilteredEnvironmentVars...)
	}
	
	// Add additional filtered patterns
	if len(other.FilteredEnvironmentPatterns) > 0 {
		c.FilteredEnvironmentPatterns = append(c.FilteredEnvironmentPatterns, other.FilteredEnvironmentPatterns...)
	}
	
	// Add allowed commands (for strict mode)
	if len(other.AllowedCommands) > 0 {
		c.AllowedCommands = append(c.AllowedCommands, other.AllowedCommands...)
	}
}

// IsCommandAllowed checks if a command is allowed by security policy
func (c *MCPSecurityConfig) IsCommandAllowed(command string) bool {
	// Extract base command name
	baseName := filepath.Base(command)
	
	// In strict mode, only explicitly allowed commands are permitted
	if c.StrictMode {
		for _, allowed := range c.AllowedCommands {
			if baseName == allowed || command == allowed {
				return true
			}
		}
		return false
	}
	
	// Check against blocked commands
	for _, blocked := range c.BlockedCommands {
		if baseName == blocked || strings.HasSuffix(command, "/"+blocked) {
			return false
		}
	}
	
	// Check if explicitly allowed (overrides blocks)
	for _, allowed := range c.AllowedCommands {
		if baseName == allowed || command == allowed {
			return true
		}
	}
	
	// Default allow if not blocked
	return true
}

// HasShellMetacharacters checks if a string contains shell metacharacters
func (c *MCPSecurityConfig) HasShellMetacharacters(s string) bool {
	for _, meta := range c.BlockedMetacharacters {
		if strings.Contains(s, meta) {
			return true
		}
	}
	return false
}

// ShouldFilterEnvironmentVar checks if an environment variable should be filtered
func (c *MCPSecurityConfig) ShouldFilterEnvironmentVar(key string) bool {
	// Check exact matches
	for _, filtered := range c.FilteredEnvironmentVars {
		if key == filtered {
			return true
		}
	}
	
	// Check patterns
	upperKey := strings.ToUpper(key)
	for _, pattern := range c.FilteredEnvironmentPatterns {
		if matchesPattern(upperKey, strings.ToUpper(pattern)) {
			return true
		}
	}
	
	return false
}

// matchesPattern checks if a string matches a wildcard pattern
func matchesPattern(str, pattern string) bool {
	// Simple wildcard matching (only supports * at beginning and end)
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(str, pattern[1:len(pattern)-1])
	} else if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(str, pattern[1:])
	} else if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(str, pattern[:len(pattern)-1])
	}
	return str == pattern
}

// Global security config instance
var globalSecurityConfig *MCPSecurityConfig

// GetSecurityConfig returns the global security configuration
func GetSecurityConfig() *MCPSecurityConfig {
	if globalSecurityConfig == nil {
		config, err := LoadSecurityConfig()
		if err != nil {
			// Fall back to defaults on error
			config = DefaultSecurityConfig()
		}
		globalSecurityConfig = config
	}
	return globalSecurityConfig
}