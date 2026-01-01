package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/ksuid"
)

// ============================================================================
// IP Extraction Tests
// ============================================================================

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
		wantErr    bool
	}{
		{
			name:       "IPv4 with port",
			remoteAddr: "192.168.1.100:54321",
			want:       "192.168.1.100",
			wantErr:    false,
		},
		{
			name:       "IPv6 with port",
			remoteAddr: "[::1]:54321",
			want:       "::1",
			wantErr:    false,
		},
		{
			name:       "localhost IPv4",
			remoteAddr: "127.0.0.1:12345",
			want:       "127.0.0.1",
			wantErr:    false,
		},
		{
			name:       "localhost IPv6",
			remoteAddr: "[::1]:12345",
			want:       "::1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/mcp", nil)
			req.RemoteAddr = tt.remoteAddr

			got, err := extractClientIP(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractClientIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// IP Allowlist Validation Tests
// ============================================================================

func TestValidateClientIP(t *testing.T) {
	t.Run("default allowlist (127.0.0.1)", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"127.0.0.1"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"localhost allowed", "127.0.0.1:54321", true},
			{"external IP rejected", "192.168.1.100:54321", false},
			{"public IP rejected", "203.0.113.10:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("single IP allowlist", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"203.0.113.10"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"allowed IP", "203.0.113.10:54321", true},
			{"different IP", "203.0.113.11:54321", false},
			{"localhost", "127.0.0.1:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("multiple IP allowlist", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"203.0.113.10", "203.0.113.11", "192.168.1.100"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"first allowed IP", "203.0.113.10:54321", true},
			{"second allowed IP", "203.0.113.11:54321", true},
			{"third allowed IP", "192.168.1.100:54321", true},
			{"not in list", "203.0.113.50:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("CIDR subnet allowlist", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"10.0.1.0/24"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"first IP in subnet", "10.0.1.1:54321", true},
			{"middle IP in subnet", "10.0.1.100:54321", true},
			{"last IP in subnet", "10.0.1.254:54321", true},
			{"IP outside subnet", "10.0.2.1:54321", false},
			{"completely different IP", "192.168.1.100:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("mixed IP and CIDR allowlist", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"127.0.0.1", "10.0.1.0/24", "203.0.113.10"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"localhost", "127.0.0.1:54321", true},
			{"in subnet", "10.0.1.50:54321", true},
			{"specific IP", "203.0.113.10:54321", true},
			{"not in list", "192.168.1.100:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("IP allowlist disabled - accept all", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"127.0.0.1"}, // Configured but disabled
				IpAllowlistDisabled: true,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"localhost", "127.0.0.1:54321", true},
			{"external IP", "192.168.1.100:54321", true},
			{"public IP", "203.0.113.50:54321", true},
			{"any IP accepted", "1.2.3.4:54321", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v (allowlist disabled)", tt.clientIP, got, tt.want)
				}
			})
		}
	})

	t.Run("IPv6 support", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				IpAllowlist:         []string{"::1", "fe80::/10"},
				IpAllowlistDisabled: false,
			},
		}

		tests := []struct {
			name     string
			clientIP string
			want     bool
		}{
			{"IPv6 localhost", "[::1]:54321", true},
			{"IPv6 in subnet", "[fe80::1]:54321", true},
			{"IPv6 not in subnet", "[2001:db8::1]:54321", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/mcp", nil)
				req.RemoteAddr = tt.clientIP

				got := server.validateClientIP(req)
				if got != tt.want {
					t.Errorf("validateClientIP(%s) = %v, want %v", tt.clientIP, got, tt.want)
				}
			})
		}
	})
}

// ============================================================================
// Required Header Validation Tests
// ============================================================================

func TestValidateRequiredHeaders(t *testing.T) {
	t.Run("no required headers - always pass", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: make(map[string]string),
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		err := server.validateRequiredHeaders(req)
		if err != nil {
			t.Errorf("validateRequiredHeaders() should pass when no headers required, got error: %v", err)
		}
	})

	t.Run("exact match validation - success", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Proxy-Verified", "true")

		err := server.validateRequiredHeaders(req)
		if err != nil {
			t.Errorf("validateRequiredHeaders() should pass with correct header, got: %v", err)
		}
	})

	t.Run("exact match validation - failure (wrong value)", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Proxy-Verified", "false")

		err := server.validateRequiredHeaders(req)
		if err == nil {
			t.Error("validateRequiredHeaders() should fail with wrong header value")
		}
	})

	t.Run("exact match validation - failure (missing header)", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		// Don't set the header

		err := server.validateRequiredHeaders(req)
		if err == nil {
			t.Error("validateRequiredHeaders() should fail with missing required header")
		}
	})

	t.Run("regex pattern validation - success", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Request-ID": "^req_[0-9a-f]{16}$",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Request-ID", "req_a1b2c3d4e5f60708") // 16 hex chars

		err := server.validateRequiredHeaders(req)
		if err != nil {
			t.Errorf("validateRequiredHeaders() should pass with matching pattern, got: %v", err)
		}
	})

	t.Run("regex pattern validation - failure (doesn't match)", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Request-ID": "^req_[0-9a-f]{16}$",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Request-ID", "invalid-format")

		err := server.validateRequiredHeaders(req)
		if err == nil {
			t.Error("validateRequiredHeaders() should fail with non-matching pattern")
		}
	})

	t.Run("multiple required headers - all valid", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
					"X-Request-ID":     "^req_.*",
					"X-Version":        "v1",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Proxy-Verified", "true")
		req.Header.Set("X-Request-ID", "req_12345")
		req.Header.Set("X-Version", "v1")

		err := server.validateRequiredHeaders(req)
		if err != nil {
			t.Errorf("validateRequiredHeaders() should pass with all headers valid, got: %v", err)
		}
	})

	t.Run("multiple required headers - one missing", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
					"X-Request-ID":     "^req_.*",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Proxy-Verified", "true")
		// X-Request-ID missing

		err := server.validateRequiredHeaders(req)
		if err == nil {
			t.Error("validateRequiredHeaders() should fail with missing required header")
		}
	})

	t.Run("multiple required headers - one invalid", func(t *testing.T) {
		server := &MCPServer{
			config: &MCPServerConfig{
				RequiredHeaders: map[string]string{
					"X-Proxy-Verified": "true",
					"X-Version":        "v1",
				},
			},
		}

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Proxy-Verified", "true")
		req.Header.Set("X-Version", "v2") // Wrong value

		err := server.validateRequiredHeaders(req)
		if err == nil {
			t.Error("validateRequiredHeaders() should fail with invalid header value")
		}
	})
}

// ============================================================================
// Integration Tests - Full Auth Middleware with IP Allowlist
// ============================================================================

func TestAuthMiddleware_WithIPAllowlist(t *testing.T) {
	validKey := ksuid.New().String()
	server := &MCPServer{
		config: &MCPServerConfig{
			ApiKey:              validKey,
			HttpHost:            "127.0.0.1",
			IpAllowlist:         []string{"127.0.0.1", "10.0.1.0/24"},
			IpAllowlistDisabled: false,
			RequiredHeaders:     make(map[string]string),
		},
	}

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("all layers pass - localhost", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.RemoteAddr = "127.0.0.1:54321"
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Next handler should be called when all layers pass")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("all layers pass - subnet", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.RemoteAddr = "10.0.1.50:54321"
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Next handler should be called for IP in subnet")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("IP not in allowlist", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.RemoteAddr = "192.168.1.100:54321" // Not in allowlist
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should NOT be called when IP not in allowlist")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware_WithRequiredHeaders(t *testing.T) {
	validKey := ksuid.New().String()
	server := &MCPServer{
		config: &MCPServerConfig{
			ApiKey:              validKey,
			HttpHost:            "127.0.0.1",
			IpAllowlist:         []string{"127.0.0.1"},
			IpAllowlistDisabled: false,
			RequiredHeaders: map[string]string{
				"X-Proxy-Verified": "true",
				"X-Request-ID":     "^req_.*",
			},
		},
	}

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("all required headers present and valid", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.Header.Set("X-Proxy-Verified", "true")
		req.Header.Set("X-Request-ID", "req_12345")
		req.RemoteAddr = "127.0.0.1:54321"
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Error("Next handler should be called when all headers valid")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("required header missing", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.Header.Set("X-Proxy-Verified", "true")
		// X-Request-ID missing
		req.RemoteAddr = "127.0.0.1:54321"
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should NOT be called when required header missing")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})

	t.Run("required header value invalid", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-API-Key", validKey)
		req.Header.Set("X-Proxy-Verified", "false") // Should be "true"
		req.Header.Set("X-Request-ID", "req_12345")
		req.RemoteAddr = "127.0.0.1:54321"
		rec := httptest.NewRecorder()

		handler := server.authMiddleware(nextHandler)
		handler.ServeHTTP(rec, req)

		if nextCalled {
			t.Error("Next handler should NOT be called when required header value invalid")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}
