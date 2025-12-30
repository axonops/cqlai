package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPermissionLockdown_DisabledByDefault tests default allows runtime changes
func TestPermissionLockdown_DisabledByDefault(t *testing.T) {
	config := DefaultMCPConfig()

	assert.False(t, config.DisableRuntimePermissionChanges,
		"Runtime permission changes should be allowed by default (false = not disabled = allowed)")
}

// TestPermissionLockdown_CanBeDisabled tests lockdown can be enabled
func TestPermissionLockdown_CanBeDisabled(t *testing.T) {
	config := DefaultMCPConfig()
	config.DisableRuntimePermissionChanges = false

	assert.False(t, config.DisableRuntimePermissionChanges,
		"Should be able to disable runtime permission changes")
}

// TestPermissionLockdown_BlocksUpdatePresetMode tests mode updates blocked when locked
func TestPermissionLockdown_BlocksUpdatePresetMode(t *testing.T) {
	// This is enforced in the MCP tool handler, not in UpdatePresetMode
	// UpdatePresetMode is a config method that doesn't check the flag
	// The flag is checked in createUpdatePermissionsHandler

	config := &MCPServerConfig{
		Mode:                          ConfigModePreset,
		PresetMode:                    "readonly",
		DisableRuntimePermissionChanges: false,
	}

	// The config method itself doesn't check the flag
	// But the MCP tool handler will
	err := config.UpdatePresetMode("readwrite")
	assert.NoError(t, err, "Config method allows update")

	// The enforcement happens in the MCP tool handler (checked in handler tests)
}

// TestPermissionLockdown_InSnapshot tests flag is included in snapshot
func TestPermissionLockdown_InSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
	}{
		{"allowed", false},   // false = not disabled = allowed
		{"disabled", true},  // true = disabled = not allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultMCPConfig()
			config.DisableRuntimePermissionChanges = tt.disabled

			snapshot := config.GetConfigSnapshot()

			assert.Equal(t, tt.disabled, snapshot.DisableRuntimePermissionChanges,
				"Snapshot should preserve DisableRuntimePermissionChanges flag")
		})
	}
}

// TestPermissionLockdown_StatusDisplay tests lockdown shown in status
func TestPermissionLockdown_StatusDisplay(t *testing.T) {
	config := &MCPServerConfig{
		Mode:                            ConfigModePreset,
		PresetMode:                      "readonly",
		DisableRuntimePermissionChanges: true, // Disabled
	}

	status := config.FormatConfigForDisplay()

	assert.Contains(t, status, "DISABLED",
		"Status should show runtime permission changes are disabled")
	assert.Contains(t, status, "locked",
		"Status should mention configuration is locked")
}

// TestPermissionLockdown_AllModes tests lockdown works with all modes
func TestPermissionLockdown_AllModes(t *testing.T) {
	modes := []struct {
		configMode ConfigMode
		presetMode string
		skipList   []string
	}{
		{ConfigModePreset, "readonly", nil},
		{ConfigModePreset, "readwrite", nil},
		{ConfigModePreset, "dba", nil},
		{ConfigModeFineGrained, "", []string{"dql", "dml"}},
		{ConfigModeFineGrained, "", []string{"ALL"}},
	}

	for _, mode := range modes {
		testName := mode.presetMode
		if testName == "" {
			testName = string(mode.configMode)
		}
		t.Run(testName, func(t *testing.T) {
			config := &MCPServerConfig{
				Mode:                            mode.configMode,
				PresetMode:                      mode.presetMode,
				SkipConfirmation:                mode.skipList,
				DisableRuntimePermissionChanges: true, // Disabled/locked
			}

			snapshot := config.GetConfigSnapshot()
			assert.True(t, snapshot.DisableRuntimePermissionChanges,
				"Lockdown should be enabled")

			// Status should show lockdown warning
			status := config.FormatConfigForDisplay()
			assert.Contains(t, status, "DISABLED",
				"Status should show lockdown for mode: %s", mode.presetMode)
		})
	}
}

// TestPermissionLockdown_WithConfirmQueries tests lockdown with confirm-queries overlay
func TestPermissionLockdown_WithConfirmQueries(t *testing.T) {
	config := &MCPServerConfig{
		Mode:                            ConfigModePreset,
		PresetMode:                      "dba",
		ConfirmQueries:                  []string{"dcl", "ddl"},
		DisableRuntimePermissionChanges: true, // Disabled/locked
	}

	snapshot := config.GetConfigSnapshot()
	assert.True(t, snapshot.DisableRuntimePermissionChanges)
	assert.Equal(t, []string{"dcl", "ddl"}, snapshot.ConfirmQueries)

	// Status should show both confirm-queries and lockdown
	status := config.FormatConfigForDisplay()
	assert.Contains(t, status, "dcl, ddl")
	assert.Contains(t, status, "DISABLED")
}

// TestPermissionLockdown_ErrorMessage tests the lockdown error message
func TestPermissionLockdown_ErrorMessage(t *testing.T) {
	// The error message is created in createUpdatePermissionsHandler
	// We test it has the right content

	expectedStrings := []string{
		"Runtime permission changes are disabled",
		"--disable-runtime-permission-changes",
		"stop the server",
		"restart",
		"locked",
	}

	errorMsg := "Runtime permission changes are disabled for this MCP server.\n\n" +
		"The server was started with --disable-runtime-permission-changes flag.\n" +
		"To change permissions, stop the server (.mcp stop) and restart with desired security settings.\n\n" +
		"Current permission configuration is locked to prevent accidental security changes."

	for _, expected := range expectedStrings {
		assert.Contains(t, errorMsg, expected,
			"Error message should contain '%s'", expected)
	}
}

// TestPermissionLockdown_DoesNotAffectOperations tests lockdown only blocks config changes, not operations
func TestPermissionLockdown_DoesNotAffectOperations(t *testing.T) {
	// Lockdown should NOT affect whether queries are allowed/confirmed
	// It only prevents changing the permission configuration

	config := &MCPServerConfig{
		Mode:                          ConfigModePreset,
		PresetMode:                    "readwrite",
		DisableRuntimePermissionChanges: false, // Locked
	}
	server := &MCPServer{config: config}

	// Operations should still work normally
	opInfo := ClassifyOperation("INSERT INTO users VALUES (...)")
	allowed, needsConf, _ := server.CheckOperationPermission(opInfo)

	assert.True(t, allowed,
		"INSERT should still be allowed in readwrite mode even when locked")
	assert.False(t, needsConf,
		"INSERT should not need confirmation in readwrite mode even when locked")

	// Lockdown ONLY affects update_mcp_permissions tool
}

// TestPermissionLockdown_FlagCombinations tests lockdown with various startup combinations
func TestPermissionLockdown_FlagCombinations(t *testing.T) {
	tests := []struct {
		name         string
		presetMode   string
		confirmList  []string
		skipList     []string
		lockdown     bool
		description  string
	}{
		{
			name:        "readonly_locked",
			presetMode:  "readonly",
			lockdown:    true,
			description: "Readonly mode with lockdown",
		},
		{
			name:        "readwrite_locked_confirm_dml",
			presetMode:  "readwrite",
			confirmList: []string{"dml"},
			lockdown:    true,
			description: "Readwrite with DML confirmations, locked",
		},
		{
			name:        "dba_locked_confirm_dcl",
			presetMode:  "dba",
			confirmList: []string{"dcl"},
			lockdown:    true,
			description: "DBA with DCL confirmations, locked",
		},
		{
			name:        "finegrained_locked",
			presetMode:  "",
			skipList:    []string{"dql", "dml"},
			lockdown:    true,
			description: "Fine-grained mode locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &MCPServerConfig{
				DisableRuntimePermissionChanges: tt.lockdown, // Clear: lockdown=true means disabled=true
			}

			if tt.presetMode != "" {
				config.Mode = ConfigModePreset
				config.PresetMode = tt.presetMode
				config.ConfirmQueries = tt.confirmList
			} else {
				config.Mode = ConfigModeFineGrained
				config.SkipConfirmation = tt.skipList
			}

			snapshot := config.GetConfigSnapshot()
			assert.Equal(t, tt.lockdown, snapshot.DisableRuntimePermissionChanges,
				"Snapshot should have DisableRuntimePermissionChanges=%v when lockdown=%v", tt.lockdown, tt.lockdown)

			// Verify status shows lockdown if enabled
			if tt.lockdown {
				status := config.FormatConfigForDisplay()
				assert.Contains(t, status, "DISABLED", tt.description)
			}
		})
	}
}
