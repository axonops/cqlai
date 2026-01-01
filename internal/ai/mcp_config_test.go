package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
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
	expectedMaxAge := 30 * 24 * time.Hour
	if config.ApiKeyMaxAge != expectedMaxAge {
		t.Errorf("Expected ApiKeyMaxAge %v (30 days), got %v", expectedMaxAge, config.ApiKeyMaxAge)
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
		validKey := ksuid.New().String()
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
		validKey := ksuid.New().String()
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
		defaultKey := ksuid.New().String()
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

	t.Run("invalid API key - not a KSUID", func(t *testing.T) {
		configJSON := `{
			"api_key": "not-a-valid-ksuid"
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

	t.Run("invalid API key - UUID instead of KSUID", func(t *testing.T) {
		uuidKey := "a1b2c3d4-1234-11ef-8000-000000000001"
		configJSON := `{
			"api_key": "` + uuidKey + `"
		}`

		configPath := filepath.Join(tmpDir, "uuid_key.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := LoadMCPConfigFromFile(configPath)
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should reject UUID (must be KSUID)")
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

	t.Run("custom api_key_max_age_days", func(t *testing.T) {
		validKey := ksuid.New().String()
		configJSON := `{
			"api_key": "` + validKey + `",
			"api_key_max_age_days": 7
		}`

		configPath := filepath.Join(tmpDir, "custom_age.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		expectedMaxAge := 7 * 24 * time.Hour
		if config.ApiKeyMaxAge != expectedMaxAge {
			t.Errorf("Expected ApiKeyMaxAge %v (7 days), got %v", expectedMaxAge, config.ApiKeyMaxAge)
		}
	})

	t.Run("disable age check with 0", func(t *testing.T) {
		validKey := ksuid.New().String()
		configJSON := `{
			"api_key": "` + validKey + `",
			"api_key_max_age_days": 0
		}`

		configPath := filepath.Join(tmpDir, "age_disabled.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.ApiKeyMaxAge != 0 {
			t.Errorf("Expected ApiKeyMaxAge 0 (disabled), got %v", config.ApiKeyMaxAge)
		}
	})

	t.Run("disable age check with negative value", func(t *testing.T) {
		validKey := ksuid.New().String()
		configJSON := `{
			"api_key": "` + validKey + `",
			"api_key_max_age_days": -1
		}`

		configPath := filepath.Join(tmpDir, "age_negative.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		if config.ApiKeyMaxAge != 0 {
			t.Errorf("Expected ApiKeyMaxAge 0 (negative treated as disabled), got %v", config.ApiKeyMaxAge)
		}
	})

	t.Run("expired key rejected at config load", func(t *testing.T) {
		// Create KSUID that's 60 days old
		pastTime := time.Now().Add(-60 * 24 * time.Hour)
		id, _ := ksuid.NewRandomWithTime(pastTime)
		expiredKey := id.String()

		configJSON := `{
			"api_key": "` + expiredKey + `",
			"api_key_max_age_days": 30
		}`

		configPath := filepath.Join(tmpDir, "expired_key.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := LoadMCPConfigFromFile(configPath)
		if err == nil {
			t.Error("LoadMCPConfigFromFile() should reject expired API key")
		}
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Error should mention expiration, got: %v", err)
		}
	})

	t.Run("env var expansion for all string fields", func(t *testing.T) {
		// Set up environment variables
		os.Setenv("TEST_HOST", "192.168.1.100")
		os.Setenv("TEST_ORIGIN", "https://app.example.com")
		os.Setenv("TEST_IP", "10.0.1.0/24")
		os.Setenv("TEST_HEADER", "X-Custom-Header")
		os.Setenv("TEST_LOG_LEVEL", "DEBUG")
		defer func() {
			os.Unsetenv("TEST_HOST")
			os.Unsetenv("TEST_ORIGIN")
			os.Unsetenv("TEST_IP")
			os.Unsetenv("TEST_HEADER")
			os.Unsetenv("TEST_LOG_LEVEL")
		}()

		validKey := ksuid.New().String()
		configJSON := `{
			"http_host": "${TEST_HOST}",
			"allowed_origins": ["${TEST_ORIGIN}"],
			"ip_allowlist": ["${TEST_IP}"],
			"audit_http_headers": ["${TEST_HEADER}"],
			"log_level": "${TEST_LOG_LEVEL}",
			"api_key": "` + validKey + `"
		}`

		configPath := filepath.Join(tmpDir, "all_env_vars.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		// Verify all env vars were expanded
		if config.HttpHost != "192.168.1.100" {
			t.Errorf("HttpHost not expanded: got %q, want \"192.168.1.100\"", config.HttpHost)
		}
		if len(config.AllowedOrigins) != 1 || config.AllowedOrigins[0] != "https://app.example.com" {
			t.Errorf("AllowedOrigins not expanded: got %v", config.AllowedOrigins)
		}
		if len(config.IpAllowlist) != 1 || config.IpAllowlist[0] != "10.0.1.0/24" {
			t.Errorf("IpAllowlist not expanded: got %v", config.IpAllowlist)
		}
		if len(config.AuditHttpHeaders) != 1 || config.AuditHttpHeaders[0] != "X-Custom-Header" {
			t.Errorf("AuditHttpHeaders not expanded: got %v", config.AuditHttpHeaders)
		}
		if config.LogLevel != "DEBUG" {
			t.Errorf("LogLevel not expanded: got %q, want \"DEBUG\"", config.LogLevel)
		}
	})

	t.Run("required headers with env var expansion", func(t *testing.T) {
		os.Setenv("PROXY_HEADER", "X-Proxy-Verified")
		os.Setenv("PROXY_VALUE", "true")
		os.Setenv("REQ_ID_PATTERN", "^req_[0-9a-f]{16}$")
		defer func() {
			os.Unsetenv("PROXY_HEADER")
			os.Unsetenv("PROXY_VALUE")
			os.Unsetenv("REQ_ID_PATTERN")
		}()

		validKey := ksuid.New().String()
		configJSON := `{
			"api_key": "` + validKey + `",
			"required_headers": {
				"${PROXY_HEADER}": "${PROXY_VALUE}",
				"X-Request-ID": "${REQ_ID_PATTERN}"
			}
		}`

		configPath := filepath.Join(tmpDir, "required_headers_env.json")
		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := LoadMCPConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("LoadMCPConfigFromFile() failed: %v", err)
		}

		// Verify env vars were expanded in both keys and values
		if val, ok := config.RequiredHeaders["X-Proxy-Verified"]; !ok || val != "true" {
			t.Errorf("Required header key not expanded: got %v", config.RequiredHeaders)
		}
		if val, ok := config.RequiredHeaders["X-Request-ID"]; !ok || val != "^req_[0-9a-f]{16}$" {
			t.Errorf("Required header value not expanded: got %v", config.RequiredHeaders)
		}
	})
}

// ============================================================================
// MCPConfigSnapshot Tests
// ============================================================================

func TestMCPConfigSnapshot(t *testing.T) {
	validKey := ksuid.New().String()
	maxAge := 15 * 24 * time.Hour
	config := &MCPServerConfig{
		HttpHost:                        "0.0.0.0",
		HttpPort:                        9999,
		ApiKey:                          validKey,
		ApiKeyMaxAge:                    maxAge,
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
	if snapshot.ApiKeyMaxAge != maxAge {
		t.Errorf("Expected ApiKeyMaxAge %v, got %v", maxAge, snapshot.ApiKeyMaxAge)
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
	validKey := ksuid.New().String()

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
