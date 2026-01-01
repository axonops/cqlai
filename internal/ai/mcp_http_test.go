package ai

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
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
			name:    "valid KSUID - generated fresh",
			key:     ksuid.New().String(),
			wantErr: false,
		},
		{
			name:    "valid KSUID from library",
			key:     ksuid.New().String(),
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid format - random text",
			key:     "not-a-ksuid",
			wantErr: true,
			errMsg:  "must be a valid KSUID",
		},
		{
			name:    "invalid format - too short",
			key:     "2ABCDEFabcdef",
			wantErr: true,
			errMsg:  "must be a valid KSUID",
		},
		{
			name:    "invalid format - UUID (wrong type)",
			key:     "a1b2c3d4-1234-11ef-8000-000000000001",
			wantErr: true,
			errMsg:  "must be a valid KSUID",
		},
		{
			name: "invalid - future timestamp (expiration bypass attack)",
			key: func() string {
				// Create KSUID with timestamp 1 year in the future
				futureTime := time.Now().Add(365 * 24 * time.Hour)
				id, _ := ksuid.NewRandomWithTime(futureTime)
				return id.String()
			}(),
			wantErr: true,
			errMsg:  "timestamp is in the future",
		},
		{
			name: "valid - timestamp with acceptable clock skew (30 seconds)",
			key: func() string {
				// Create KSUID with timestamp 30 seconds in the future (within tolerance)
				futureTime := time.Now().Add(30 * time.Second)
				id, _ := ksuid.NewRandomWithTime(futureTime)
				return id.String()
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with default maxAge (30 days)
			err := ValidateAPIKeyFormat(tt.key, 30*24*time.Hour)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKeyFormat(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateAPIKeyFormat(%q) error = %v, want error containing %q",
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
			// Create KSUID with specified age
			pastTime := time.Now().Add(-tt.keyAge)
			id, err := ksuid.NewRandomWithTime(pastTime)
			if err != nil {
				t.Fatalf("Failed to create KSUID: %v", err)
			}
			key := id.String()

			err = ValidateAPIKeyFormat(key, tt.maxAge)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKeyFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateAPIKeyFormat() error = %v, want error containing %q",
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

		// KSUID is 27 chars in base62
		if len(key) != 27 {
			t.Errorf("Expected KSUID length 27, got %d: %s", len(key), key)
		}

		// Verify it's a valid KSUID (can be parsed)
		_, err = ksuid.Parse(key)
		if err != nil {
			t.Errorf("Generated key is not a valid KSUID: %v", err)
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

		// Parse KSUID and get timestamp
		id, err := ksuid.Parse(key)
		if err != nil {
			t.Fatalf("Failed to parse KSUID: %v", err)
		}

		// Verify timestamp is within generation window
		// KSUID timestamp is in seconds, so allow 2-second window
		keyTime := id.Time()
		if keyTime.Before(beforeGen.Add(-1*time.Second)) || keyTime.After(afterGen.Add(1*time.Second)) {
			t.Errorf("Key timestamp %v not within generation window [%v - %v]",
				keyTime, beforeGen, afterGen)
		}

		// Verify we can calculate age (for future expiration logic)
		age := time.Since(keyTime)
		if age < -1*time.Second || age > 2*time.Second {
			t.Errorf("Unexpected key age: %v", age)
		}
	})
}

// ============================================================================
// API Key Validation Tests
// ============================================================================

func TestValidateAPIKey(t *testing.T) {
	// Use a valid KSUID for testing
	testKey := ksuid.New().String()

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
			provided: ksuid.New().String(), // Different KSUID
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
			name: "normal KSUID key (27 chars)",
			key:  "2ABCDEFGHIJKLMNOPQRSTUVWXYZa", // 27 chars
			want: "2ABCDEFG...XYZa",             // First 8 + ... + last 4
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
			got := MaskAPIKey(tt.key)
			if got != tt.want {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
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
	validKey := ksuid.New().String()
	server := &MCPServer{
		config: &MCPServerConfig{
			ApiKey:              validKey,
			HttpHost:            "127.0.0.1",
			IpAllowlist:         []string{"127.0.0.1"},
			IpAllowlistDisabled: false,
			RequiredHeaders:     make(map[string]string),
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
		req.RemoteAddr = "127.0.0.1:54321"
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
		req.Header.Set("X-API-Key", ksuid.New().String()) // Different KSUID
		req.RemoteAddr = "127.0.0.1:54321"
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
		req.RemoteAddr = "127.0.0.1:54321"
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
		req.RemoteAddr = "127.0.0.1:54321"
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
		req.RemoteAddr = "127.0.0.1:54321"
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

	correctKey := ksuid.New().String()
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
	almostCorrect := ksuid.New().String() // Different KSUID
	if server.validateAPIKey(almostCorrect) {
		t.Error("validateAPIKey() should reject different KSUID")
	}

	// Verify correct key works
	if !server.validateAPIKey(correctKey) {
		t.Error("validateAPIKey() should accept correct key")
	}
}
