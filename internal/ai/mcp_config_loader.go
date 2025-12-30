package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadMCPConfigFromFile loads MCP configuration from JSON file and returns MCPServerConfig
func LoadMCPConfigFromFile(filePath string) (*MCPServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var jsonConfig map[string]any
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create MCPServerConfig from JSON
	config := DefaultMCPConfig()

	// Socket path
	if sp, ok := jsonConfig["socket_path"].(string); ok && sp != "" {
		config.SocketPath = sp
	}

	// Log level
	if ll, ok := jsonConfig["log_level"].(string); ok && ll != "" {
		config.LogLevel = ll
	}

	// Log file
	if lf, ok := jsonConfig["log_file"].(string); ok && lf != "" {
		config.LogFile = lf
	}

	// Mode (preset or skip_confirmation for fine-grained)
	if mode, ok := jsonConfig["mode"].(string); ok && mode != "" {
		// Preset mode
		config.Mode = ConfigModePreset
		config.PresetMode = mode

		// Confirm queries overlay
		if cq, ok := jsonConfig["confirm_queries"].([]any); ok && len(cq) > 0 {
			cats := make([]string, len(cq))
			for i, c := range cq {
				cats[i] = c.(string)
			}
			config.ConfirmQueries = cats
		}
	}

	// Skip confirmation (fine-grained mode)
	if sc, ok := jsonConfig["skip_confirmation"].([]any); ok && len(sc) > 0 {
		cats := make([]string, len(sc))
		for i, c := range sc {
			cats[i] = c.(string)
		}
		config.Mode = ConfigModeFineGrained
		config.PresetMode = ""
		config.SkipConfirmation = cats
		// SESSION auto-added by UpdateSkipConfirmation
		if !containsCategory(cats, "session") && !containsCategory(cats, "ALL") {
			config.SkipConfirmation = append(config.SkipConfirmation, "session")
		}
	}

	// Lockdown
	if lockdown, ok := jsonConfig["disable_runtime_permission_changes"].(bool); ok {
		config.DisableRuntimePermissionChanges = lockdown
	}

	// Timeout
	if timeout, ok := jsonConfig["confirmation_timeout_seconds"].(float64); ok && timeout > 0 {
		config.ConfirmationTimeout = time.Duration(timeout) * time.Second
	}

	return config, nil
}
