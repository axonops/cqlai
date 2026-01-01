package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// ============================================================================
// DefaultMCPConfig Tests
// ============================================================================

func TestDefaultMCPConfig(t *testing.T) {
	config := DefaultMCPConfig()

	// HTTP transport defaults
	if config.HttpHost != "127.0.0.1" {
		t.Errorf("Expected HttpHost '127.0.0.1', got %q", config.HttpHost)
	}
	if config.HttpPort != 8888 {
		t.Errorf("Expected HttpPort 8888, got %d", config.HttpPort)
	}
	if config.ApiKey != "" {
		t.Errorf("Expected ApiKey empty (auto-generated), got %q", config.ApiKey)
	}
	if config.AllowedOrigins != nil {
		t.Errorf("Expected AllowedOrigins nil, got %v", config.AllowedOrigins)
	}

	// Other defaults
	if config.Mode != ConfigModePreset {
		t.Errorf("Expected Mode %v, got %v", ConfigModePreset, config.Mode)
	}
	if config.PresetMode != "readonly" {
		t.Errorf("Expected PresetMode 'readonly', got %q", config.PresetMode)
	}
	if config.ConfirmationTimeout != 5*time.Minute {
		t.Errorf("Expected ConfirmationTimeout 5m, got %v", config.ConfirmationTimeout)
	}
}

// ============================================================================
// LoadMCPConfigFromFile Tests
// ============================================================================

func TestLoadMCPConfigFromFile(t *testing.T) {
	// Create temp directory for test configs
	tmpDir := t.TempDir()

	t.Run("valid HTTP config with all fields", func(t *testing.T) {
		validKey := gocql.TimeUUID().String()
		configJSON := `{
			"http_host": "0.0.0.0",
			"http_port": 9999,
			"api_key": "` + validKey + `",
			"allowed_origins": ["https://app.example.com", "http://localhost"],
			"mode": "readonly",
			"log_level": "DEBUG"
		}`

		configPath := filepath.Join(tmpDir, "valid_http.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.HttpHost != "0.0.0.0" {
			t.Errorf("Expected HttpHost '0.0.0.0', got %q", config.HttpHost)
		}
		if config.HttpPort != 9999 {
			t.Errorf("Expected HttpPort 9999, got %d", config.HttpPort)
		}
		if config.ApiKey != validKey {
			t.Errorf("Expected ApiKey %q, got %q", validKey, config.ApiKey)
		}
		if len(config.AllowedOrigins) != 2 {
			t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
		}
		if config.AllowedOrigins[0] != "https://app.example.com" {
			t.Errorf("Expected first origin 'https://app.example.com', got %q", config.AllowedOrigins[0])
		}
	})

	t.Run("HTTP config with environment variable expansion", func(t *testing.T) {
		validKey := gocql.TimeUUID().String()
		os.Setenv("TEST_MCP_API_KEY", validKey)
		defer os.Unsetenv("TEST_MCP_API_KEY")

		configJSON := `{
			"http_host": "127.0.0.1",
			"http_port": 8888,
			"api_key": "${TEST_MCP_API_KEY}"
		}`

		configPath := filepath.Join(tmpDir, "env_var.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.ApiKey != validKey {
			t.Errorf("Expected ApiKey to be expanded to %q, got %q", validKey, config.ApiKey)
		}
	})

	t.Run("HTTP config with default value expansion", func(t *testing.T) {
		defaultKey := gocql.TimeUUID().String()
		configJSON := `{
			"api_key": "${NONEXISTENT_KEY:-` + defaultKey + `}"
		}`

		configPath := filepath.Join(tmpDir, "default_val.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.ApiKey != defaultKey {
			t.Errorf("Expected ApiKey to use default %q, got %q", defaultKey, config.ApiKey)
		}
	})

	t.Run("invalid API key - not a TimeUUID", func(t *testing.T) {
		configJSON := `{
			"api_key": "not-a-valid-timeuuid"
		}`

		configPath := filepath.Join(tmpDir, "invalid_key.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := LoadMCPConfigFromFile(configPath)
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should reject invalid API key")
		}
		if err != nil && err.Error() == "" {
			t.Error("Error should have a message")
		}
	})

	t.Run("invalid API key - UUIDv4 instead of TimeUUID", func(t *testing.T) {
		randomUUID := gocql.MustRandomUUID().String() // This is v4
		configJSON := `{
			"api_key": "` + randomUUID + `"
		}`

		configPath := filepath.Join(tmpDir, "uuidv4_key.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := LoadMCPConfigFromFile(configPath)
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should reject UUIDv4 (must be TimeUUID/UUIDv1)")
		}
	})

	t.Run("partial HTTP config uses defaults", func(t *testing.T) {
		configJSON := `{
			"http_host": "192.168.1.100"
		}`

		configPath := filepath.Join(tmpDir, "partial.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.HttpHost != "192.168.1.100" {
			t.Errorf("Expected HttpHost '192.168.1.100', got %q", config.HttpHost)
		}
		// Should use default port
		if config.HttpPort != 8888 {
			t.Errorf("Expected default HttpPort 8888, got %d", config.HttpPort)
		}
		// Should have empty API key (will be auto-generated)
		if config.ApiKey != "" {
			t.Errorf("Expected ApiKey empty (auto-generated), got %q", config.ApiKey)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadMCPConfigFromFile("/nonexistent/path/config.json")
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should fail for nonexistent file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		configJSON := `{ invalid json }`

		configPath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := LoadMCPConfigFromFile(configPath)
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should fail for invalid JSON")
		}
	})
}

// ============================================================================
// MCPConfigSnapshot Tests
// ============================================================================

func TestMCPConfigSnapshot(t *testing.T) {
	validKey := gocql.TimeUUID().String()
	config := &MCPServerConfig{
		HttpHost:                        "0.0.0.0",
		HttpPort:                        9999,
		ApiKey:                          validKey,
		AllowedOrigins:                  []string{"https://example.com"},
		Mode:                            ConfigModePreset,
		PresetMode:                      "readonly",
		ConfirmQueries:                  []string{"DDL"},
		SkipConfirmation:                []string{"SELECT"},
		LogLevel:                        "DEBUG",
		DisableRuntimePermissionChanges: true,
		AllowMCPRequestApproval:         true,
		HistoryFile:                     "/tmp/history",
		HistoryMaxSize:                  1024 * 1024,
		HistoryMaxRotations:             3,
		HistoryRotationInterval:         2 * time.Minute,
	}

	snapshot := config.GetConfigSnapshot()

	// Verify HTTP fields are copied
	if snapshot.HttpHost != "0.0.0.0" {
		t.Errorf("Expected HttpHost '0.0.0.0', got %q", snapshot.HttpHost)
	}
	if snapshot.HttpPort != 9999 {
		t.Errorf("Expected HttpPort 9999, got %d", snapshot.HttpPort)
	}
	if snapshot.ApiKey != validKey {
		t.Errorf("Expected ApiKey %q, got %q", validKey, snapshot.ApiKey)
	}
	if len(snapshot.AllowedOrigins) != 1 {
		t.Errorf("Expected 1 allowed origin, got %d", len(snapshot.AllowedOrigins))
	}
	if snapshot.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("Expected origin 'https://example.com', got %q", snapshot.AllowedOrigins[0])
	}

	// Verify other fields
	if snapshot.Mode != ConfigModePreset {
		t.Errorf("Expected Mode %v, got %v", ConfigModePreset, snapshot.Mode)
	}
	if snapshot.PresetMode != "readonly" {
		t.Errorf("Expected PresetMode 'readonly', got %q", snapshot.PresetMode)
	}

	// Verify slices are copied (not shared references)
	if &config.AllowedOrigins == &snapshot.AllowedOrigins {
		t.Error("AllowedOrigins should be a copy, not a shared reference")
	}
	if &config.ConfirmQueries == &snapshot.ConfirmQueries {
		t.Error("ConfirmQueries should be a copy, not a shared reference")
	}

	// Modify original and verify snapshot is unchanged
	config.HttpHost = "127.0.0.1"
	config.AllowedOrigins[0] = "https://modified.com"

	if snapshot.HttpHost == "127.0.0.1" {
		t.Error("Snapshot should not be affected by changes to original config")
	}
	if snapshot.AllowedOrigins[0] == "https://modified.com" {
		t.Error("Snapshot AllowedOrigins should not be affected by changes to original")
	}
}

// ============================================================================
// Config Integration Tests
// ============================================================================

func TestConfigRoundTrip(t *testing.T) {
	// Test that we can load a config, modify it, and snapshot it
	tmpDir := t.TempDir()
	validKey := gocql.TimeUUID().String()

	configJSON := `{
		"http_host": "0.0.0.0",
		"http_port": 7777,
		"api_key": "` + validKey + `",
		"allowed_origins": ["https://test.com"],
		"mode": "dba",
		"confirmation_timeout_seconds": 600
	}`

	configPath := filepath.Join(tmpDir, "roundtrip.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	config, err := LoadMCPConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
	}

	// Get snapshot
	snapshot := config.GetConfigSnapshot()

	// Verify all HTTP fields match
	if snapshot.HttpHost != "0.0.0.0" {
		t.Errorf("Snapshot HttpHost mismatch")
	}
	if snapshot.HttpPort != 7777 {
		t.Errorf("Snapshot HttpPort mismatch")
	}
	if snapshot.ApiKey != validKey {
		t.Errorf("Snapshot ApiKey mismatch")
	}
	if len(snapshot.AllowedOrigins) != 1 || snapshot.AllowedOrigins[0] != "https://test.com" {
		t.Errorf("Snapshot AllowedOrigins mismatch")
	}
}
