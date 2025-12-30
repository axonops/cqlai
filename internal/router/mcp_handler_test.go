package router

import (
	"strings"
	"testing"
	"time"
)

// TestMCPHandler_Lifecycle tests basic MCP handler lifecycle
func TestMCPHandler_Lifecycle(t *testing.T) {
	// This test just verifies the handler can be created
	// We can't easily test Start/Stop without a real Cassandra cluster

	handler := &MCPHandler{
		replSession: nil,
		mcpServer:   nil,
	}

	if handler == nil {
		t.Fatal("Failed to create MCPHandler")
	}
}

// TestMCPHandler_ShowUsage tests the usage command
func TestMCPHandler_ShowUsage(t *testing.T) {
	handler := &MCPHandler{}

	usage := handler.showUsage()

	// Verify all commands are documented
	expectedCommands := []string{
		".mcp start",
		".mcp stop",
		".mcp status",
		".mcp metrics",
		".mcp pending",
		".mcp confirm",
		".mcp deny",
		".mcp log",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(usage, cmd) {
			t.Errorf("Usage text missing command: %s", cmd)
		}
	}
}

// TestMCPHandler_HandleMCPCommand_NoServer tests commands when server not running
func TestMCPHandler_HandleMCPCommand_NoServer(t *testing.T) {
	handler := &MCPHandler{
		replSession: nil,
		mcpServer:   nil,
	}

	tests := []struct {
		name     string
		command  string
		contains string
	}{
		{
			name:     "status when not running",
			command:  ".mcp status",
			contains: "not running",
		},
		{
			name:     "stop when not running",
			command:  ".mcp stop",
			contains: "not running",
		},
		{
			name:     "metrics when not running",
			command:  ".mcp metrics",
			contains: "not running",
		},
		{
			name:     "pending when not running",
			command:  ".mcp pending",
			contains: "not running",
		},
		{
			name:     "confirm when not running",
			command:  ".mcp confirm req_001",
			contains: "not running",
		},
		{
			name:     "deny when not running",
			command:  ".mcp deny req_001",
			contains: "not running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleMCPCommand(tt.command)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

// TestMCPHandler_HandleMCPCommand_InvalidCommands tests invalid command handling
func TestMCPHandler_HandleMCPCommand_InvalidCommands(t *testing.T) {
	handler := &MCPHandler{}

	tests := []struct {
		name     string
		command  string
		contains string
	}{
		{
			name:     "unknown command",
			command:  ".mcp foobar",
			contains: "Unknown MCP command",
		},
		{
			name:     "empty command",
			command:  ".mcp",
			contains: "MCP (Model Context Protocol) Server Commands",
		},
		{
			name:     "confirm without request_id",
			command:  ".mcp confirm",
			contains: "Usage: .mcp confirm <request_id>",
		},
		{
			name:     "deny without request_id",
			command:  ".mcp deny",
			contains: "Usage: .mcp deny <request_id>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleMCPCommand(tt.command)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

// TestMCPHandler_GetPendingConfirmationCount tests the confirmation count method
func TestMCPHandler_GetPendingConfirmationCount(t *testing.T) {
	tests := []struct {
		name     string
		handler  *MCPHandler
		expected int
	}{
		{
			name:     "nil handler",
			handler:  nil,
			expected: 0,
		},
		{
			name: "no server",
			handler: &MCPHandler{
				replSession: nil,
				mcpServer:   nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.handler.GetPendingConfirmationCount()
			if count != tt.expected {
				t.Errorf("Expected count %d, got %d", tt.expected, count)
			}
		})
	}
}

// TestFormatDuration tests the duration formatting function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		contains string
	}{
		{
			name:     "future time",
			input:    time.Now().Add(5 * time.Minute),
			contains: "in",
		},
		{
			name:     "past time",
			input:    time.Now().Add(-30 * time.Second),
			contains: "ago",
		},
		{
			name:     "duration",
			input:    5 * time.Minute,
			contains: "5m",
		},
		{
			name:     "unknown type",
			input:    "not a time",
			contains: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

// TestMCPHandler_ConfirmDenyWithNoServer tests confirm/deny error handling
func TestMCPHandler_ConfirmDenyWithNoServer(t *testing.T) {
	// Verify the error messages when server is nil
	handler := &MCPHandler{
		replSession: nil,
		mcpServer:   nil,
	}

	// Test confirm with no server
	result := handler.handleConfirm("req_001")
	if !strings.Contains(result, "not running") {
		t.Errorf("Expected 'not running' error, got: %s", result)
	}

	// Test deny with no server
	result = handler.handleDeny("req_001", "test reason")
	if !strings.Contains(result, "not running") {
		t.Errorf("Expected 'not running' error, got: %s", result)
	}

	// Test pending with no server
	result = handler.handlePending()
	if !strings.Contains(result, "not running") {
		t.Errorf("Expected 'not running' error, got: %s", result)
	}
}

// TestMCPHandler_CommandParsing tests that commands are parsed correctly
func TestMCPHandler_CommandParsing(t *testing.T) {
	handler := &MCPHandler{}

	tests := []struct {
		name        string
		command     string
		shouldError bool
		errorText   string
	}{
		{
			name:        "confirm with id",
			command:     ".mcp confirm req_001",
			shouldError: true,
			errorText:   "not running",
		},
		{
			name:        "confirm with MCP prefix",
			command:     ".MCP confirm req_001",
			shouldError: true,
			errorText:   "not running",
		},
		{
			name:        "deny with id and reason",
			command:     ".mcp deny req_002 Too dangerous",
			shouldError: true,
			errorText:   "not running",
		},
		{
			name:        "deny with id only",
			command:     ".mcp deny req_002",
			shouldError: true,
			errorText:   "not running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleMCPCommand(tt.command)
			if tt.shouldError && !strings.Contains(result, tt.errorText) {
				t.Errorf("Expected error containing %q, got: %s", tt.errorText, result)
			}
		})
	}
}
