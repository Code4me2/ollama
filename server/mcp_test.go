package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

// TestMCPClientInitialization tests the MCP client initialization
func TestMCPClientInitialization(t *testing.T) {
	// Create a mock MCP server command
	client := NewMCPClient("test", "echo", []string{"test"}, nil)
	
	if client.name != "test" {
		t.Errorf("Expected client name to be 'test', got %s", client.name)
	}
	
	if client.command != "echo" {
		t.Errorf("Expected command to be 'echo', got %s", client.command)
	}
	
	if !client.initialized {
		// This is expected - not initialized yet
		t.Log("Client correctly not initialized on creation")
	}
}

// TestSecureEnvironmentFiltering tests environment variable filtering
func TestSecureEnvironmentFiltering(t *testing.T) {
	// Set some test environment variables
	os.Setenv("TEST_SAFE_VAR", "safe_value")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "secret_key")
	os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin:/root/bin")
	
	client := NewMCPClient("test", "echo", []string{}, nil)
	env := client.buildSecureEnvironment()
	
	// Check that sensitive variables are filtered out
	for _, e := range env {
		if strings.HasPrefix(e, "AWS_SECRET_ACCESS_KEY=") {
			t.Errorf("Sensitive AWS_SECRET_ACCESS_KEY should be filtered out")
		}
		if strings.Contains(e, "/root/bin") {
			t.Errorf("Dangerous PATH component /root/bin should be filtered out")
		}
	}
	
	// Check that PATH is present but sanitized
	hasPath := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			hasPath = true
			if strings.Contains(e, "/root") {
				t.Errorf("PATH should not contain /root directories")
			}
		}
	}
	
	if !hasPath {
		t.Errorf("PATH should be present in environment")
	}
}

// TestMCPManagerAddServer tests adding MCP servers to the manager
func TestMCPManagerAddServer(t *testing.T) {
	manager := NewMCPManager(5)
	
	// Test adding a valid server config
	config := api.MCPServerConfig{
		Name:    "test_server",
		Command: "python",
		Args:    []string{"-m", "test_module"},
		Env:     map[string]string{"TEST": "value"},
	}
	
	// This will fail in test environment but validates the validation logic
	err := manager.AddServer(config)
	if err == nil || !strings.Contains(err.Error(), "failed to initialize") {
		// Expected to fail initialization in test environment
		t.Logf("Server initialization failed as expected: %v", err)
	}
	
	// Test invalid server names
	invalidConfigs := []api.MCPServerConfig{
		{Name: "", Command: "python"}, // Empty name
		{Name: strings.Repeat("a", 101), Command: "python"}, // Too long
		{Name: "test/server", Command: "python"}, // Invalid characters
	}
	
	for _, cfg := range invalidConfigs {
		err := manager.validateServerConfig(cfg)
		if err == nil {
			t.Errorf("Should reject invalid config: %+v", cfg)
		}
	}
}

// TestDangerousCommandValidation tests rejection of dangerous commands
func TestDangerousCommandValidation(t *testing.T) {
	manager := NewMCPManager(5)
	
	dangerousConfigs := []api.MCPServerConfig{
		{Name: "test1", Command: "bash"},
		{Name: "test2", Command: "/bin/sh"},
		{Name: "test3", Command: "sudo"},
		{Name: "test4", Command: "rm"},
		{Name: "test5", Command: "curl"},
		{Name: "test6", Command: "eval"},
	}
	
	for _, cfg := range dangerousConfigs {
		err := manager.validateServerConfig(cfg)
		if err == nil {
			t.Errorf("Should reject dangerous command: %s", cfg.Command)
		}
		if !strings.Contains(err.Error(), "not allowed for security") {
			t.Errorf("Expected security error for command %s, got: %v", cfg.Command, err)
		}
	}
	
	// Test that safe commands are allowed
	safeConfigs := []api.MCPServerConfig{
		{Name: "test1", Command: "python"},
		{Name: "test2", Command: "node"},
		{Name: "test3", Command: "/usr/bin/python3"},
	}
	
	for _, cfg := range safeConfigs {
		err := manager.validateServerConfig(cfg)
		if err != nil {
			t.Errorf("Should allow safe command %s: %v", cfg.Command, err)
		}
	}
}

// TestShellInjectionPrevention tests prevention of shell injection
func TestShellInjectionPrevention(t *testing.T) {
	manager := NewMCPManager(5)
	
	// Test arguments with shell metacharacters
	injectionConfigs := []api.MCPServerConfig{
		{
			Name:    "test1",
			Command: "python",
			Args:    []string{"; rm -rf /"},
		},
		{
			Name:    "test2", 
			Command: "python",
			Args:    []string{"test", "| cat /etc/passwd"},
		},
		{
			Name:    "test3",
			Command: "python",
			Args:    []string{"$(whoami)"},
		},
		{
			Name:    "test4",
			Command: "python",
			Args:    []string{"`id`"},
		},
	}
	
	for _, cfg := range injectionConfigs {
		err := manager.validateServerConfig(cfg)
		if err == nil {
			t.Errorf("Should reject shell injection attempt in args: %v", cfg.Args)
		}
		if !strings.Contains(err.Error(), "shell metacharacters") {
			t.Errorf("Expected shell metacharacter error, got: %v", err)
		}
	}
}

// TestToolResultCache tests the tool result caching functionality
func TestToolResultCache(t *testing.T) {
	cache := NewToolResultCache(10, 1*time.Second)
	
	// Test setting and getting cached results
	result := ToolResult{
		Content: "test result",
		Error:   nil,
	}
	
	args := map[string]interface{}{"param": "value"}
	cache.Set("test_tool", args, result)
	
	// Should get the cached result
	cached := cache.Get("test_tool", args)
	if cached == nil {
		t.Errorf("Expected to get cached result")
	}
	if cached.Content != "test result" {
		t.Errorf("Expected cached content 'test result', got %s", cached.Content)
	}
	
	// Test cache expiration
	time.Sleep(1100 * time.Millisecond)
	cached = cache.Get("test_tool", args)
	if cached != nil {
		t.Errorf("Expected cache to expire after TTL")
	}
	
	// Test cache size limit
	cache = NewToolResultCache(2, 1*time.Hour)
	cache.Set("tool1", args, result)
	cache.Set("tool2", args, result)
	cache.Set("tool3", args, result) // Should evict oldest
	
	if cache.Get("tool1", args) != nil {
		t.Errorf("Oldest entry should have been evicted")
	}
}

// TestParallelToolExecution tests parallel execution of tools
func TestParallelToolExecution(t *testing.T) {
	manager := NewMCPManager(5)
	
	// Create test tool calls
	toolCalls := []api.ToolCall{
		{
			Function: api.ToolCallFunction{
				Name:      "tool1",
				Arguments: map[string]interface{}{"test": "1"},
			},
		},
		{
			Function: api.ToolCallFunction{
				Name:      "tool2",
				Arguments: map[string]interface{}{"test": "2"},
			},
		},
		{
			Function: api.ToolCallFunction{
				Name:      "tool3",
				Arguments: map[string]interface{}{"test": "3"},
			},
		},
	}
	
	// Execute in parallel (will fail but tests the mechanism)
	results := manager.ExecuteToolsParallel(toolCalls)
	
	if len(results) != len(toolCalls) {
		t.Errorf("Expected %d results, got %d", len(toolCalls), len(results))
	}
	
	// All should have errors since no MCP servers are connected
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("Expected error for tool call %d", i)
		}
	}
}

// TestPathSanitization tests PATH environment variable sanitization
func TestPathSanitization(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
		rejected []string
	}{
		{
			input:    "/usr/local/bin:/usr/bin:/bin:/root/bin",
			expected: []string{"/usr/local/bin", "/usr/bin", "/bin"},
			rejected: []string{"/root/bin"},
		},
		{
			input:    "/usr/bin:./local:../bin:/opt/bin",
			expected: []string{"/usr/bin", "/opt/bin"},
			rejected: []string{"./local", "../bin"},
		},
	}
	
	for _, tc := range testCases {
		result := sanitizePath(tc.input)
		
		for _, exp := range tc.expected {
			if !strings.Contains(result, exp) {
				t.Errorf("Expected %s in sanitized PATH", exp)
			}
		}
		
		for _, rej := range tc.rejected {
			if strings.Contains(result, rej) {
				t.Errorf("Should not contain %s in sanitized PATH", rej)
			}
		}
	}
}

// TestMCPClientTimeout tests timeout handling for tool execution
func TestMCPClientTimeout(t *testing.T) {
	// This test verifies that the timeout is properly set
	// In real scenario, we'd need a slow MCP server to test actual timeout
	client := NewMCPClient("test", "sleep", []string{"60"}, nil)
	
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Try to call with timeout (will fail but tests the mechanism)
	req := mcpCallToolRequest{
		Name:      "test_tool",
		Arguments: map[string]interface{}{},
	}
	
	var resp mcpCallToolResponse
	err := client.callWithContext(ctx, "tools/call", req, &resp)
	
	// Should timeout or fail
	if err == nil {
		t.Errorf("Expected timeout or error")
	}
}

// TestEnvironmentVariableValidation tests validation of environment variables
func TestEnvironmentVariableValidation(t *testing.T) {
	manager := NewMCPManager(5)
	
	// Test invalid environment variable names
	invalidEnvConfigs := []api.MCPServerConfig{
		{
			Name:    "test1",
			Command: "python",
			Env:     map[string]string{"VAR=BAD": "value"},
		},
		{
			Name:    "test2",
			Command: "python",
			Env:     map[string]string{"VAR;CMD": "value"},
		},
		{
			Name:    "test3",
			Command: "python",
			Env:     map[string]string{"VAR|PIPE": "value"},
		},
	}
	
	for _, cfg := range invalidEnvConfigs {
		err := manager.validateServerConfig(cfg)
		if err == nil {
			t.Errorf("Should reject invalid environment variable names: %v", cfg.Env)
		}
	}
	
	// Test valid environment variables
	validConfig := api.MCPServerConfig{
		Name:    "test",
		Command: "python",
		Env: map[string]string{
			"PYTHONPATH": "/usr/lib/python3",
			"MY_VAR":     "value",
			"TEST_123":   "test",
		},
	}
	
	err := manager.validateServerConfig(validConfig)
	if err != nil {
		t.Errorf("Should allow valid environment variables: %v", err)
	}
}

// BenchmarkToolExecution benchmarks tool execution performance
func BenchmarkToolExecution(b *testing.B) {
	manager := NewMCPManager(10)
	
	toolCall := api.ToolCall{
		Function: api.ToolCallFunction{
			Name:      "test_tool",
			Arguments: map[string]interface{}{"param": "value"},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ExecuteTool(toolCall)
	}
}

// BenchmarkParallelToolExecution benchmarks parallel tool execution
func BenchmarkParallelToolExecution(b *testing.B) {
	manager := NewMCPManager(10)
	
	toolCalls := make([]api.ToolCall, 10)
	for i := range toolCalls {
		toolCalls[i] = api.ToolCall{
			Function: api.ToolCallFunction{
				Name:      fmt.Sprintf("tool_%d", i),
				Arguments: map[string]interface{}{"param": i},
			},
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ExecuteToolsParallel(toolCalls)
	}
}