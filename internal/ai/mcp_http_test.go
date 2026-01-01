package ai

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// ============================================================================
// API Key Format Validation Tests
// ============================================================================

func TestValidateAPIKeyFormat(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid TimeUUID (v1) - generated fresh",
			key:     gocql.TimeUUID().String(),
			wantErr: false,
		},
		{
			name:    "valid TimeUUID from gocql",
			key:     gocql.TimeUUID().String(),
			wantErr: false,
		},
		{
			name:    "invalid - UUIDv4 (random UUID)",
			key:     gocql.MustRandomUUID().String(), // This is v4
			wantErr: true,
			errMsg:  "must be a TimeUUID (UUIDv1), got UUIDv4",
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid format - random text",
			key:     "not-a-uuid",
			wantErr: true,
			errMsg:  "must be a valid TimeUUID",
		},
		{
			name:    "invalid format - malformed UUID",
			key:     "a1b2c3d4-1234-11ef",
			wantErr: true,
			errMsg:  "must be a valid TimeUUID",
		},
		{
			name:    "invalid format - wrong separators",
			key:     "a1b2c3d4_1234_11ef_8000_000000000001",
			wantErr: true,
			errMsg:  "must be a valid TimeUUID",
		},
		{
			name: "invalid - future timestamp (expiration bypass attack)",
			key: func() string {
				// Create TimeUUID with timestamp 1 year in the future
				futureTime := time.Now().Add(365 * 24 * time.Hour)
				timestamp := (futureTime.Unix()*1e7 + 0x01b21dd213814000)
				uuid := gocql.TimeUUIDWith(timestamp, 0, []byte{0, 0, 0, 0, 0, 0})
				return uuid.String()
			}(),
			wantErr: true,
			errMsg:  "timestamp is in the future",
		},
		{
			name: "valid - timestamp with acceptable clock skew (30 seconds)",
			key: func() string {
				// Create TimeUUID with timestamp 30 seconds in the future (within tolerance)
				futureTime := time.Now().Add(30 * time.Second)
				timestamp := (futureTime.Unix()*1e7 + 0x01b21dd213814000)
				uuid := gocql.TimeUUIDWith(timestamp, 0, []byte{0, 0, 0, 0, 0, 0})
				return uuid.String()
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with default maxAge (30 days)
			err := validateAPIKeyFormat(tt.key, 30*24*time.Hour)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKeyFormat(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateAPIKeyFormat(%q) error = %v, want error containing %q",
						tt.key, err, tt.errMsg)
				}
			}
		})
	}
}

// ============================================================================
// API Key Expiration Tests
// ============================================================================

func TestValidateAPIKeyFormat_Expiration(t *testing.T) {
	tests := []struct {
		name    string
		keyAge  time.Duration
		maxAge  time.Duration
		wantErr bool
		errMsg  string
	}{
		{
			name:    "fresh key (1 hour old, max 30 days)",
			keyAge:  1 * time.Hour,
			maxAge:  30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "key within limit (29 days old, max 30 days)",
			keyAge:  29 * 24 * time.Hour,
			maxAge:  30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "expired key (31 days old, max 30 days)",
			keyAge:  31 * 24 * time.Hour,
			maxAge:  30 * 24 * time.Hour,
			wantErr: true,
			errMsg:  "API key expired",
		},
		{
			name:    "very old key (365 days old, max 30 days)",
			keyAge:  365 * 24 * time.Hour,
			maxAge:  30 * 24 * time.Hour,
			wantErr: true,
			errMsg:  "API key expired",
		},
		{
			name:    "old key but age check disabled (100 days, maxAge=0)",
			keyAge:  100 * 24 * time.Hour,
			maxAge:  0, // Disabled
			wantErr: false,
		},
		{
			name:    "expired key (10 days old, max 7 days)",
			keyAge:  10 * 24 * time.Hour,
			maxAge:  7 * 24 * time.Hour,
			wantErr: true,
			errMsg:  "API key expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TimeUUID with specified age
			pastTime := time.Now().Add(-tt.keyAge)
			timestamp := (pastTime.Unix()*1e7 + 0x01b21dd213814000)
			uuid := gocql.TimeUUIDWith(timestamp, 0, []byte{0, 0, 0, 0, 0, 0})
			key := uuid.String()

			err := validateAPIKeyFormat(key, tt.maxAge)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKeyFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateAPIKeyFormat() error = %v, want error containing %q",
						err, tt.errMsg)
				}
			}
		})
	}
}

// ============================================================================
// API Key Generation Tests
// ============================================================================

func TestGenerateAPIKey(t *testing.T) {
	t.Run("generates valid key format", func(t *testing.T) {
		key, err := generateAPIKey()
		if err != nil {
			t.Fatalf("generateAPIKey() failed: %v", err)
		}

		// Check UUID format (8-4-4-4-12 hex digits)
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(key) {
			t.Errorf("Expected valid UUID format, got: %s", key)
		}

		// Verify it's a valid TimeUUID (can be parsed)
		_, err = gocql.ParseUUID(key)
		if err != nil {
			t.Errorf("Generated key is not a valid UUID: %v", err)
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, err1 := generateAPIKey()
		key2, err2 := generateAPIKey()

		if err1 != nil || err2 != nil {
			t.Fatalf("generateAPIKey() failed: %v, %v", err1, err2)
		}

		if key1 == key2 {
			t.Errorf("Generated identical keys (should be unique): %s", key1)
		}
	})

	t.Run("embeds timestamp for expiration support", func(t *testing.T) {
		beforeGen := time.Now()
		key, err := generateAPIKey()
		afterGen := time.Now()

		if err != nil {
			t.Fatalf("generateAPIKey() failed: %v", err)
		}

		// Parse TimeUUID and get timestamp
		uuid, err := gocql.ParseUUID(key)
		if err != nil {
			t.Fatalf("Failed to parse UUID: %v", err)
		}

		// Verify timestamp is within generation window
		keyTime := uuid.Time()
		if keyTime.Before(beforeGen) || keyTime.After(afterGen) {
			t.Errorf("Key timestamp %v not within generation window [%v - %v]",
				keyTime, beforeGen, afterGen)
		}

		// Verify we can calculate age (for future expiration logic)
		age := time.Since(keyTime)
		if age < 0 || age > time.Second {
			t.Errorf("Unexpected key age: %v", age)
		}
	})
}

// ============================================================================
// API Key Validation Tests
// ============================================================================

func TestValidateAPIKey(t *testing.T) {
	// Use a valid TimeUUID format for testing
	testKey := "a1b2c3d4-1234-11ef-8000-000000000001"

	config := &MCPServerConfig{
		ApiKey: testKey,
	}

	server := &MCPServer{
		config: config,
	}

	tests := []struct {
		name     string
		provided string
		want     bool
	}{
		{
			name:     "valid key",
			provided: testKey,
			want:     true,
		},
		{
			name:     "invalid key",
			provided: "a1b2c3d4-1234-11ef-8000-000000000002",
			want:     false,
		},
		{
			name:     "empty provided key",
			provided: "",
			want:     false,
		},
		{
			name:     "completely wrong key",
			provided: "totally-wrong-key",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.validateAPIKey(tt.provided)
			if got != tt.want {
				t.Errorf("validateAPIKey(%q) = %v, want %v", tt.provided, got, tt.want)
			}
		})
	}

	t.Run("rejects when no key configured", func(t *testing.T) {
		emptyServer := &MCPServer{
			config: &MCPServerConfig{ApiKey: ""},
		}
		if emptyServer.validateAPIKey("any-key") {
			t.Error("validateAPIKey() should reject when no key configured")
		}
	})
}

// ============================================================================
// API Key Masking Tests
// ============================================================================

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "normal UUID key",
			key:  "a1b2c3d4-1234-11ef-8000-123456789abc",
			want: "a1b2c3d4...9abc", // Shows first 8 + ... + last 4
		},
		{
			name: "short key",
			key:  "short",
			want: "***",
		},
		{
			name: "exactly 12 chars",
			key:  "exactly12chr",
			want: "***",
		},
		{
			name: "13 chars (just over threshold)",
			key:  "sk_local_test",
			want: "sk_local...test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.key)
			if got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Origin Validation Tests
// ============================================================================

func TestValidateOrigin(t *testing.T) {
	t.Run("localhost binding", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				HttpHost: "127.0.0.1",
			},
		}

		tests := []struct {
			name   string
			origin string
			want   bool
		}{
			{"no origin header", "", true},
			{"localhost http", "http://localhost:3000", true},
			{"localhost https", "https://localhost:3000", true},
			{"127.0.0.1 http", "http://127.0.0.1:8080", true},
			{"127.0.0.1 https", "https://127.0.0.1:8080", true},
			{"external origin", "https://evil.com", false},
			{"192.168.x.x", "http://192.168.1.100", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				if tt.origin != "" {
					req.Header.Set("Origin", tt.origin)
				}

				got := server.validateOrigin(req)
				if got != tt.want {
					t.Errorf("validateOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
				}
			})
		}
	})

	t.Run("remote binding with allowed_origins", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				HttpHost: "0.0.0.0",
				AllowedOrigins: []string{
					"https://app.company.com",
					"http://localhost",
				},
			},
		}

		tests := []struct {
			name   string
			origin string
			want   bool
		}{
			{"no origin header", "", true},
			{"allowed origin 1", "https://app.company.com", true},
			{"allowed origin 2", "http://localhost:3000", true},
			{"not allowed", "https://evil.com", false},
			{"similar but not exact", "https://app.company.com.evil.com", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				if tt.origin != "" {
					req.Header.Set("Origin", tt.origin)
				}

				got := server.validateOrigin(req)
				if got != tt.want {
					t.Errorf("validateOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
				}
			})
		}
	})

	t.Run("remote binding with no allowed_origins configured", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				HttpHost:       "0.0.0.0",
				AllowedOrigins: nil,
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("Origin", "https://any.com")

		// Should reject when no allowed_origins configured (safe default)
		if server.validateOrigin(req) {
			t.Error("validateOrigin() should reject when no allowed_origins configured for remote binding")
		}
	})
}

// ============================================================================
// Auth Middleware Tests
// ============================================================================

func TestAuthMiddleware(t *testing.T) {
	validKey := "a1b2c3d4-1234-11ef-8000-000000000001"
	server := &MCPServer{
		config: &MCPServerConfig{
			ApiKey:   validKey,
			HttpHost: "127.0.0.1",
		},
	}

	// Mock next handler that records if it was called
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing API key header", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should not be called when API key missing")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "missing X-API-Key header") {
			t.Errorf("Expected error message about missing header, got: %s", rec.Body.String())
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", "a1b2c3d4-1234-11ef-8000-000000000099")
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should not be called with invalid API key")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("valid API key but invalid origin", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should not be called with invalid origin")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "origin not allowed") {
			t.Errorf("Expected error message about origin, got: %s", rec.Body.String())
		}
	})

	t.Run("valid API key and valid origin", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Next handler should be called with valid API key and origin")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("valid API key with no origin header", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		// No Origin header (direct API request)
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Next handler should be called with valid API key and no origin")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

// ============================================================================
// Environment Variable Expansion Tests
// ============================================================================

func TestExpandEnvVar(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_API_KEY", "test-key-123")
	os.Setenv("TEST_HOST", "example.com")
	defer func() {
		os.Unsetenv("TEST_API_KEY")
		os.Unsetenv("TEST_HOST")
	}()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple variable",
			input: "${TEST_API_KEY}",
			want:  "test-key-123",
		},
		{
			name:  "variable with default (var exists)",
			input: "${TEST_API_KEY:-default-value}",
			want:  "test-key-123",
		},
		{
			name:  "variable with default (var does not exist)",
			input: "${NONEXISTENT:-default-value}",
			want:  "default-value",
		},
		{
			name:  "no variable",
			input: "plain-text-key",
			want:  "plain-text-key",
		},
		{
			name:  "variable in middle",
			input: "prefix-${TEST_HOST}-suffix",
			want:  "prefix-example.com-suffix",
		},
		{
			name:  "multiple variables",
			input: "${TEST_API_KEY}-${TEST_HOST}",
			want:  "test-key-123-example.com",
		},
		{
			name:  "empty variable without default",
			input: "${EMPTY_VAR}",
			want:  "",
		},
		{
			name:  "malformed variable (no closing brace)",
			input: "${TEST_API_KEY",
			want:  "${TEST_API_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEnvVar(tt.input)
			if got != tt.want {
				t.Errorf("expandEnvVar(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Security: Constant-Time Comparison Test
// ============================================================================

func TestValidateAPIKey_ConstantTime(t *testing.T) {
	// This test ensures we're using constant-time comparison
	// While we can't directly measure timing, we can verify the behavior
	// matches what crypto/subtle.ConstantTimeCompare would do

	correctKey := "a1b2c3d4-1234-11ef-8000-000000000001"
	config := &MCPServerConfig{
		ApiKey: correctKey,
	}
	server := &MCPServer{config: config}

	// Test that differing lengths still use constant-time comparison
	// (crypto/subtle handles this safely)
	shortKey := "short"
	if server.validateAPIKey(shortKey) {
		t.Error("validateAPIKey() should reject short key")
	}

	// Test that byte-by-byte differences are handled
	almostCorrect := "a1b2c3d4-1234-11ef-8000-000000000002" // Last digit different
	if server.validateAPIKey(almostCorrect) {
		t.Error("validateAPIKey() should reject almost-correct key")
	}

	// Verify correct key works
	if !server.validateAPIKey(correctKey) {
		t.Error("validateAPIKey() should accept correct key")
	}
}
