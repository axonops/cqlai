package ai

import (
	"fmt"
	"strings"
	"time"
)

// ConfigMode represents the configuration mode type
type ConfigMode string

const (
	ConfigModePreset      ConfigMode = "preset"       // Using preset modes (readonly, readwrite, dba)
	ConfigModeFineGrained ConfigMode = "fine-grained" // Using skip-confirmation list
)

// UpdatePresetMode changes the preset mode (readonly/readwrite/dba)
func (c *MCPServerConfig) UpdatePresetMode(mode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate preset mode
	validModes := map[string]bool{
		"readonly":   true,
		"readwrite":  true,
		"dba":        true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid preset mode: %s (valid: readonly, readwrite, dba)", mode)
	}

	c.Mode = ConfigModePreset
	c.PresetMode = mode
	c.SkipConfirmation = nil // Clear fine-grained settings

	return nil
}

// UpdateConfirmQueries changes which categories require confirmation (preset mode only)
func (c *MCPServerConfig) UpdateConfirmQueries(categories []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Mode != ConfigModePreset {
		return fmt.Errorf("confirm-queries only allowed with preset modes (readonly/readwrite/dba), currently in %s mode", c.Mode)
	}

	// Handle special values
	if len(categories) == 1 {
		switch strings.ToLower(categories[0]) {
		case "none", "disable", "":
			c.ConfirmQueries = nil
			return nil
		case "all":
			c.ConfirmQueries = []string{"ALL"}
			return nil
		}
	}

	// Validate categories
	validCategories := map[string]bool{
		"dql": true, "session": true, "dml": true,
		"ddl": true, "dcl": true, "file": true,
	}

	for _, cat := range categories {
		if !validCategories[strings.ToLower(cat)] {
			return fmt.Errorf("invalid category: %s (valid: dql, session, dml, ddl, dcl, file)", cat)
		}
	}

	c.ConfirmQueries = categories
	return nil
}

// UpdateSkipConfirmation changes the skip-confirmation list (switches to fine-grained mode)
func (c *MCPServerConfig) UpdateSkipConfirmation(categories []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Switch to fine-grained mode
	c.Mode = ConfigModeFineGrained
	c.PresetMode = ""
	c.ConfirmQueries = nil

	// Handle special value "ALL" - skip everything
	if len(categories) == 1 && strings.ToUpper(categories[0]) == "ALL" {
		c.SkipConfirmation = []string{"ALL"}
		return nil
	}

	// Handle special value "none" - skip nothing (confirm everything)
	if len(categories) == 0 || (len(categories) == 1 && (strings.ToLower(categories[0]) == "none" || categories[0] == "")) {
		c.SkipConfirmation = []string{"session"} // Only SESSION, always allowed
		return nil
	}

	// Validate categories
	validCategories := map[string]bool{
		"dql": true, "session": true, "dml": true,
		"ddl": true, "dcl": true, "file": true,
		"none": true, // Special value
	}

	for _, cat := range categories {
		catLower := strings.ToLower(cat)
		if !validCategories[catLower] {
			return fmt.Errorf("invalid category: %s (valid: dql, session, dml, ddl, dcl, file, ALL, none)", cat)
		}
	}

	c.SkipConfirmation = categories

	// SESSION always auto-added (never requires confirmation)
	if !containsCategory(categories, "session") && !containsCategory(categories, "ALL") {
		c.SkipConfirmation = append(c.SkipConfirmation, "session")
	}

	return nil
}

// GetConfigSnapshot returns a thread-safe snapshot of current configuration
func (c *MCPServerConfig) GetConfigSnapshot() MCPConfigSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return MCPConfigSnapshot{
		Mode:                            c.Mode,
		PresetMode:                      c.PresetMode,
		ConfirmQueries:                  append([]string(nil), c.ConfirmQueries...),   // Copy slice
		SkipConfirmation:                append([]string(nil), c.SkipConfirmation...), // Copy slice
		SocketPath:                      c.SocketPath,                                  // Deprecated
		HttpHost:                        c.HttpHost,
		HttpPort:                        c.HttpPort,
		ApiKey:                          c.ApiKey,
		ApiKeyMaxAge:                    c.ApiKeyMaxAge,
		AllowedOrigins:                  append([]string(nil), c.AllowedOrigins...), // Copy slice
		LogLevel:                        c.LogLevel,
		DisableRuntimePermissionChanges: c.DisableRuntimePermissionChanges,
		AllowMCPRequestApproval:         c.AllowMCPRequestApproval,
		HistoryFile:                     c.HistoryFile,
		HistoryMaxSize:                  c.HistoryMaxSize,
		HistoryMaxRotations:             c.HistoryMaxRotations,
		HistoryRotationInterval:         c.HistoryRotationInterval,
	}
}

// MCPConfigSnapshot is a point-in-time snapshot of configuration
type MCPConfigSnapshot struct {
	Mode                            ConfigMode
	PresetMode                      string
	ConfirmQueries                  []string
	SkipConfirmation                []string
	SocketPath                      string // Deprecated - will be removed
	HttpHost                        string
	HttpPort                        int
	ApiKey                          string
	ApiKeyMaxAge                    time.Duration
	AllowedOrigins                  []string
	LogLevel                        string
	DisableRuntimePermissionChanges bool
	AllowMCPRequestApproval         bool
	HistoryFile                     string
	HistoryMaxSize                  int64
	HistoryMaxRotations             int
	HistoryRotationInterval         time.Duration
}

// GetModeDescription returns a human-readable description of the current mode
func (s *MCPConfigSnapshot) GetModeDescription() string {
	if s.Mode == ConfigModePreset {
		switch s.PresetMode {
		case "readonly":
			return "Read-only mode: Only queries and session settings allowed"
		case "readwrite":
			return "Read-write mode: Queries, data modifications, and file operations allowed"
		case "dba":
			return "DBA mode: All operations allowed"
		default:
			return fmt.Sprintf("Unknown preset mode: %s", s.PresetMode)
		}
	} else if s.Mode == ConfigModeFineGrained {
		if containsCategory(s.SkipConfirmation, "ALL") {
			return "Fine-grained mode: Skip confirmation on ALL operations"
		}
		return fmt.Sprintf("Fine-grained mode: Skip confirmation on %s", strings.Join(s.SkipConfirmation, ", "))
	}
	return "Unknown mode"
}

// containsCategory checks if a category is in the list (case-insensitive)
func containsCategory(list []string, category string) bool {
	catLower := strings.ToLower(category)
	for _, item := range list {
		if strings.ToLower(item) == catLower {
			return true
		}
	}
	return false
}

// ParseCategoryList parses a comma-separated category list
func ParseCategoryList(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(strings.ToLower(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
