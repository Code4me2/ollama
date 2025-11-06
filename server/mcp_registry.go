package server

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// MCPRegistry manages MCP managers globally with session support
type MCPRegistry struct {
	mu       sync.RWMutex
	managers map[string]*SessionManager // session ID -> manager
	ttl      time.Duration              // session timeout
}

// SessionManager wraps an MCPManager with session metadata
type SessionManager struct {
	*MCPManager
	lastAccess time.Time
	sessionID  string
	configs    []api.MCPServerConfig
}

var (
	globalRegistry *MCPRegistry
	registryOnce   sync.Once
)

// GetMCPRegistry returns the singleton MCP registry
func GetMCPRegistry() *MCPRegistry {
	registryOnce.Do(func() {
		globalRegistry = &MCPRegistry{
			managers: make(map[string]*SessionManager),
			ttl:      30 * time.Minute, // Sessions expire after 30 min
		}
		// Start cleanup goroutine
		go globalRegistry.cleanupExpired()
	})
	return globalRegistry
}

// GetOrCreateManager gets existing or creates new MCP manager for session
func (r *MCPRegistry) GetOrCreateManager(sessionID string, configs []api.MCPServerConfig) (*MCPManager, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if manager exists and configs match
	if sm, exists := r.managers[sessionID]; exists {
		if configsMatch(sm.configs, configs) {
			sm.lastAccess = time.Now()
			slog.Debug("Reusing existing MCP manager", "session", sessionID, "clients", len(sm.clients))
			return sm.MCPManager, nil
		}
		// Configs changed, shutdown old manager
		slog.Info("MCP configs changed, recreating manager", "session", sessionID)
		sm.Shutdown()
		delete(r.managers, sessionID)
	}

	// Create new manager
	slog.Info("Creating new MCP manager", "session", sessionID, "configs", len(configs))
	manager := NewMCPManager(10)
	for _, config := range configs {
		if err := manager.AddServer(config); err != nil {
			slog.Warn("Failed to add MCP server", "name", config.Name, "error", err)
		}
	}

	r.managers[sessionID] = &SessionManager{
		MCPManager: manager,
		lastAccess: time.Now(),
		sessionID:  sessionID,
		configs:    configs,
	}

	return manager, nil
}

// GetManagerForToolsPath creates a manager for a tools directory path
func (r *MCPRegistry) GetManagerForToolsPath(model string, toolsPath string) (*MCPManager, error) {
	// Generate consistent session ID for model + tools path
	sessionID := generateToolsSessionID(model, toolsPath)
	
	// Create MCP config for filesystem server
	configs := []api.MCPServerConfig{
		{
			Name:    "filesystem",
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", toolsPath},
		},
	}
	
	return r.GetOrCreateManager(sessionID, configs)
}

// cleanupExpired removes expired sessions
func (r *MCPRegistry) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for sessionID, sm := range r.managers {
			if now.Sub(sm.lastAccess) > r.ttl {
				slog.Info("Cleaning up expired MCP session", "session", sessionID)
				sm.Shutdown()
				delete(r.managers, sessionID)
			}
		}
		r.mu.Unlock()
	}
}

// Shutdown closes all managers and stops the registry
func (r *MCPRegistry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Shutting down MCP registry", "sessions", len(r.managers))
	for sessionID, sm := range r.managers {
		slog.Debug("Shutting down session", "session", sessionID)
		sm.Shutdown()
	}
	r.managers = make(map[string]*SessionManager)
}

// configsMatch checks if two sets of MCP configs are equivalent
func configsMatch(a, b []api.MCPServerConfig) bool {
	if len(a) != len(b) {
		return false
	}
	// Simple comparison - could be enhanced
	return reflect.DeepEqual(a, b)
}

// generateToolsSessionID creates a consistent session ID for model + tools path
func generateToolsSessionID(model, toolsPath string) string {
	h := sha256.New()
	h.Write([]byte(model))
	h.Write([]byte(toolsPath))
	return "tools-" + hex.EncodeToString(h.Sum(nil))[:16]
}

// GenerateSessionID creates a session ID based on the request
func GenerateSessionID(req api.ChatRequest) string {
	// If explicit session ID provided
	if req.SessionID != "" {
		return req.SessionID
	}

	// For interactive mode with tools path
	if req.ToolsPath != "" {
		return generateToolsSessionID(req.Model, req.ToolsPath)
	}

	// Default: use request-specific ID (no persistence)
	// Use timestamp and random component for uniqueness
	h := sha256.New()
	h.Write([]byte(time.Now().Format(time.RFC3339Nano)))
	h.Write([]byte(req.Model))
	return "req-" + hex.EncodeToString(h.Sum(nil))[:16]
}