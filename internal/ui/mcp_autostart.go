package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// MCPStartConfig holds configuration for auto-starting MCP server
type MCPStartConfig struct {
	SocketPath                       string   `json:"socket_path"`
	LogLevel                         string   `json:"log_level"`
	LogFile                          string   `json:"log_file"`
	Mode                             string   `json:"mode"` // "readonly", "readwrite", "dba"
	ConfirmQueries                   []string `json:"confirm_queries"`
	SkipConfirmation                 []string `json:"skip_confirmation"`
	DisableRuntimePermissionChanges  bool     `json:"disable_runtime_permission_changes"`
	ConfirmationTimeoutSeconds       int      `json:"confirmation_timeout_seconds"`
}

// DefaultMCPStartConfig returns default MCP auto-start configuration
func DefaultMCPStartConfig() *MCPStartConfig {
	return &MCPStartConfig{
		SocketPath:                      "/tmp/cqlai-mcp.sock",
		LogLevel:                        "info",
		LogFile:                         "/tmp/cqlai-mcp.log",
		Mode:                            "readonly",
		ConfirmQueries:                  nil,
		SkipConfirmation:                nil,
		DisableRuntimePermissionChanges: false,
		ConfirmationTimeoutSeconds:      300,
	}
}

// LoadMCPConfig loads MCP configuration from JSON file
func LoadMCPConfig(filePath string) (*MCPStartConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MCPStartConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Validate configuration
	if config.SocketPath == "" {
		config.SocketPath = "/tmp/cqlai-mcp.sock"
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.LogFile == "" {
		config.LogFile = "/tmp/cqlai-mcp.log"
	}
	if config.Mode == "" {
		config.Mode = "readonly"
	}
	if config.ConfirmationTimeoutSeconds == 0 {
		config.ConfirmationTimeoutSeconds = 300
	}

	return &config, nil
}

// buildMCPStartCommand builds the .mcp start command string from config
func buildMCPStartCommand(config *MCPStartConfig) string {
	var parts []string
	parts = append(parts, ".mcp start")

	// Add socket path if not default
	if config.SocketPath != "/tmp/cqlai-mcp.sock" {
		parts = append(parts, fmt.Sprintf("--socket-path %s", config.SocketPath))
	}

	// Add log level if not default
	if config.LogLevel != "info" {
		parts = append(parts, fmt.Sprintf("--log-level %s", config.LogLevel))
	}

	// Add log file if not default
	if config.LogFile != "/tmp/cqlai-mcp.log" {
		parts = append(parts, fmt.Sprintf("--log-file %s", config.LogFile))
	}

	// Add mode based on configuration
	if config.Mode != "" && len(config.SkipConfirmation) == 0 {
		// Preset mode
		switch config.Mode {
		case "readonly":
			parts = append(parts, "--readonly_mode")
		case "readwrite":
			parts = append(parts, "--readwrite_mode")
		case "dba":
			parts = append(parts, "--dba_mode")
		}

		// Add confirm-queries overlay if specified
		if len(config.ConfirmQueries) > 0 {
			parts = append(parts, fmt.Sprintf("--confirm-queries %s", strings.Join(config.ConfirmQueries, ",")))
		}
	}

	// Add skip-confirmation if specified (fine-grained mode)
	if len(config.SkipConfirmation) > 0 {
		parts = append(parts, fmt.Sprintf("--skip-confirmation %s", strings.Join(config.SkipConfirmation, ",")))
	}

	// Add runtime permission lockdown if enabled
	if config.DisableRuntimePermissionChanges {
		parts = append(parts, "--disable-runtime-permission-changes")
	}

	return strings.Join(parts, " ")
}
