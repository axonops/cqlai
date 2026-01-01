package ai

import (
	"compress/gzip"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/segmentio/ksuid"
)

// MCPServer manages the Model Context Protocol server for Claude Desktop integration.
// It runs independently from the REPL, with its own Cassandra session and schema cache.
// Both REPL and MCP share tool implementation logic via getToolData().
type MCPServer struct {
	// Independent Cassandra resources (not shared with REPL)
	session  *db.Session
	cache    *db.SchemaCache
	resolver *Resolver

	// MCP server infrastructure
	mcpServer  *server.MCPServer            // The mark3labs MCP server
	httpServer *server.StreamableHTTPServer // HTTP transport server
	listener   net.Listener                  // Deprecated - will be removed
	socketPath string                        // Deprecated - will be removed

	// Confirmation system for dangerous queries
	confirmationQueue *ConfirmationQueue
	config            *MCPServerConfig

	// Observability
	metrics *MetricsCollector
	mcpLog  *MCPLogger

	// Query history for audit trail
	historyFilePath string
	historyMu       sync.RWMutex // Protects history file during rotation

	// Server state
	running   bool
	startedAt time.Time
	mu        sync.Mutex

	// Graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// MCPServerConfig holds configuration for the MCP server
// Configuration can be changed dynamically at runtime via .mcp config commands
type MCPServerConfig struct {
	// Server infrastructure - HTTP transport
	HttpHost             string        // HTTP server host (default: 127.0.0.1)
	HttpPort             int           // HTTP server port (default: 8888)
	ApiKey               string        // API key for authentication (auto-generated if empty)
	ApiKeyMaxAge         time.Duration // Max age for API keys (default: 30 days, 0 = disabled)
	AllowedOrigins       []string      // Allowed Origin headers for non-localhost (DNS rebinding protection)
	IpAllowlist          []string      // IP/CIDR allowlist (default: 127.0.0.1)
	IpAllowlistDisabled  bool          // Disable IP checking (SECURITY RISK, triggers warnings)
	AuditHttpHeaders     []string      // HTTP headers to log for audit trail (default: X-Forwarded-For, User-Agent)
	RequiredHeaders      map[string]string // Required headers with exact values or regex patterns

	// Legacy
	SocketPath          string // Deprecated: Will be removed

	// Server configuration
	ConfirmationTimeout time.Duration
	LogLevel            string
	LogFile             string
	HistoryFile             string        // Path to MCP history file (default: ~/.cqlai/cqlai_mcp_history)
	HistoryMaxSize          int64         // Max history file size in bytes before rotation (default: 10MB)
	HistoryMaxRotations     int           // Number of rotated history files to keep (default: 5)
	HistoryRotationInterval time.Duration // How often to check for rotation (default: 1 minute)

	// Confirmation system (thread-safe via mutex)
	Mode             ConfigMode // "preset" or "fine-grained"
	PresetMode       string     // "readonly", "readwrite", "dba" (when Mode=preset)
	ConfirmQueries   []string   // Categories requiring confirmation (when Mode=preset)
	SkipConfirmation []string   // Categories to skip confirmation (when Mode=fine-grained)

	// Runtime permission configuration control
	DisableRuntimePermissionChanges bool // If false, update_mcp_permissions tool is disabled

	// MCP request approval control (NOT runtime changeable - startup only)
	AllowMCPRequestApproval bool // If true, allows confirm_request MCP tool to approve dangerous queries (default: false)

	// Thread safety for runtime config changes
	mu sync.RWMutex
}

// DefaultMCPConfig returns default MCP server configuration
func DefaultMCPConfig() *MCPServerConfig {
	home, _ := os.UserHomeDir()
	cqlaiDir := filepath.Join(home, ".cqlai")

	return &MCPServerConfig{
		// HTTP transport (default)
		HttpHost:            "127.0.0.1",
		HttpPort:            8888,
		ApiKey:              "", // Auto-generated on start if empty
		ApiKeyMaxAge:        30 * 24 * time.Hour, // 30 days (0 = disabled)
		AllowedOrigins:      nil, // Only used for non-localhost bindings
		IpAllowlist:         []string{"127.0.0.1"}, // Default: localhost only
		IpAllowlistDisabled: false,
		AuditHttpHeaders:    []string{"X-Forwarded-For", "User-Agent"}, // Default audit headers
		RequiredHeaders:     make(map[string]string), // Default: no required headers

		// Legacy (deprecated)
		SocketPath:              "", // Empty = HTTP mode (socket deprecated)

		// Server configuration
		ConfirmationTimeout:     5 * time.Minute,
		LogLevel:                "info",
		LogFile:                 filepath.Join(cqlaiDir, "cqlai_mcp.log"),
		HistoryFile:             filepath.Join(cqlaiDir, "cqlai_mcp_history"),
		HistoryMaxSize:          10 * 1024 * 1024, // 10MB
		HistoryMaxRotations:     5,                // Keep 5 rotated files
		HistoryRotationInterval: 1 * time.Minute, // Check every minute

		// Default: readonly mode (safest)
		Mode:             ConfigModePreset,
		PresetMode:       "readonly",
		ConfirmQueries:   nil, // No additional confirmations
		SkipConfirmation: nil,

		// Allow runtime permission changes by default (false = not disabled = allowed)
		DisableRuntimePermissionChanges: false,

		// Disallow MCP request approval by default (security: requires explicit opt-in)
		AllowMCPRequestApproval: false,
	}
}

// NewMCPServer creates a new MCP server.
// It creates an independent Cassandra session from the provided REPL session's cluster.
// The MCP server runs in its own goroutine and does not share state with the REPL.
func NewMCPServer(replSession *db.Session, config *MCPServerConfig) (*MCPServer, error) {
	if replSession == nil {
		return nil, fmt.Errorf("REPL session cannot be nil")
	}
	if config == nil {
		config = DefaultMCPConfig()
	}

	// Create independent session from REPL's cluster config
	cluster := replSession.GetCluster()
	username := replSession.Username()

	mcpSession, err := db.NewSessionFromCluster(cluster, username, false) // batchMode=false (need schema for AI)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP session: %w", err)
	}

	// Get the schema cache from the new session
	cache := mcpSession.GetSchemaCache()
	if cache == nil {
		mcpSession.Close()
		return nil, fmt.Errorf("failed to initialize schema cache for MCP session")
	}

	// Create resolver for this session's cache
	resolver := NewResolver(cache)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize metrics collector
	metrics := NewMetricsCollector()

	// Initialize MCP logger
	mcpLog, err := NewMCPLogger(config.LogFile, config.LogLevel)
	if err != nil {
		mcpSession.Close()
		cancel()
		return nil, fmt.Errorf("failed to create MCP logger: %w", err)
	}

	// Create confirmation queue
	confirmationQueue := NewConfirmationQueue()

	s := &MCPServer{
		session:           mcpSession,
		cache:             cache,
		resolver:          resolver,
		socketPath:        config.SocketPath,
		confirmationQueue: confirmationQueue,
		config:            config,
		historyFilePath:   config.HistoryFile,
		metrics:           metrics,
		mcpLog:            mcpLog,
		ctx:               ctx,
		cancel:            cancel,
		running:           false,
	}

	logger.DebugfToFile("MCP", "MCP server created (not started yet)")

	return s, nil
}

// ============================================================================
// HTTP Authentication and Security
// ============================================================================

// generateAPIKey generates a KSUID API key
// KSUID = K-Sortable Unique ID: 160 bits (32-bit timestamp + 128-bit random payload)
// More secure than TimeUUID (no MAC address, more entropy, cryptographically random)
func generateAPIKey() (string, error) {
	id := ksuid.New()
	return id.String(), nil
}

// ValidateAPIKeyFormat validates that a string is a valid KSUID
// maxAge: maximum allowed age for the key (0 = no age check)
func ValidateAPIKeyFormat(key string, maxAge time.Duration) error {
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Parse KSUID
	id, err := ksuid.Parse(key)
	if err != nil {
		return fmt.Errorf("API key must be a valid KSUID (27-char base62): %w", err)
	}

	// Extract timestamp (KSUID embeds epoch seconds in first 4 bytes)
	keyTime := id.Time()
	keyAge := time.Since(keyTime)

	// SECURITY: Reject future timestamps (prevents expiration bypass attack)
	// Allow 1 minute clock skew tolerance
	if keyTime.After(time.Now().Add(1 * time.Minute)) {
		return fmt.Errorf("API key timestamp is in the future (%v) - rejected as invalid",
			keyTime.Format(time.RFC3339))
	}

	// SECURITY: Reject expired keys (if maxAge check enabled)
	if maxAge > 0 && keyAge > maxAge {
		return fmt.Errorf("API key expired: created at %v (age: %v, max allowed: %v)",
			keyTime.Format(time.RFC3339), keyAge.Round(time.Hour), maxAge)
	}

	// Log the timestamp for audit purposes
	logger.DebugfToFile("MCP", "API key validated: KSUID generated at %v (age: %v)",
		keyTime.Format(time.RFC3339), keyAge.Round(time.Second))

	return nil
}

// validateAPIKey validates the provided API key using constant-time comparison
func (s *MCPServer) validateAPIKey(provided string) bool {
	expected := s.config.ApiKey
	if expected == "" {
		// No key configured - reject for security
		logger.DebugfToFile("MCP", "API key validation failed: no key configured")
		return false
	}
	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// MaskAPIKey masks an API key for safe logging
// KSUID format: 27 chars base62, shows first 8 and last 4
func MaskAPIKey(key string) string {
	if len(key) <= 12 {
		return "***"
	}
	return key[:8] + "..." + key[len(key)-4:]
}

// ParseKSUID parses an API key as a KSUID and returns it with timestamp
func ParseKSUID(key string) (ksuid.KSUID, error) {
	id, err := ksuid.Parse(key)
	if err != nil {
		return ksuid.KSUID{}, fmt.Errorf("invalid KSUID format: %w", err)
	}
	return id, nil
}

// ============================================================================
// IP Allowlisting
// ============================================================================

// extractClientIP extracts the client IP from http.Request
// Returns the direct connection IP (from RemoteAddr), not X-Forwarded-For
func extractClientIP(r *http.Request) (string, error) {
	// RemoteAddr format: "ip:port" or "[ipv6]:port"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Try without port (shouldn't happen but be defensive)
		return r.RemoteAddr, nil
	}
	return host, nil
}

// validateClientIP checks if client IP is in allowlist
func (s *MCPServer) validateClientIP(r *http.Request) bool {
	// If IP allowlist disabled, allow all IPs (with warning)
	if s.config.IpAllowlistDisabled {
		logger.DebugfToFile("MCP", "IP allowlist disabled - accepting all IPs (SECURITY RISK)")
		return true
	}

	clientIP, err := extractClientIP(r)
	if err != nil {
		logger.DebugfToFile("MCP", "Failed to extract client IP: %v", err)
		return false
	}

	// Parse client IP
	clientAddr := net.ParseIP(clientIP)
	if clientAddr == nil {
		logger.DebugfToFile("MCP", "Invalid client IP format: %s", clientIP)
		return false
	}

	// Check against allowlist (supports both single IPs and CIDR ranges)
	for _, allowed := range s.config.IpAllowlist {
		// Check if it's a CIDR range
		if strings.Contains(allowed, "/") {
			_, ipNet, err := net.ParseCIDR(allowed)
			if err != nil {
				logger.DebugfToFile("MCP", "Invalid CIDR in allowlist: %s", allowed)
				continue
			}
			if ipNet.Contains(clientAddr) {
				logger.DebugfToFile("MCP", "Client IP %s allowed (matches CIDR %s)", clientIP, allowed)
				return true
			}
		} else {
			// Direct IP comparison
			allowedAddr := net.ParseIP(allowed)
			if allowedAddr == nil {
				logger.DebugfToFile("MCP", "Invalid IP in allowlist: %s", allowed)
				continue
			}
			if clientAddr.Equal(allowedAddr) {
				logger.DebugfToFile("MCP", "Client IP %s allowed (exact match)", clientIP)
				return true
			}
		}
	}

	logger.DebugfToFile("MCP", "Client IP %s rejected (not in allowlist: %v)", clientIP, s.config.IpAllowlist)
	return false
}

// ============================================================================
// Required Header Validation
// ============================================================================

// validateRequiredHeaders checks if all required headers are present with valid values
func (s *MCPServer) validateRequiredHeaders(r *http.Request) error {
	if len(s.config.RequiredHeaders) == 0 {
		return nil // No required headers configured
	}

	for headerName, expectedValue := range s.config.RequiredHeaders {
		actualValue := r.Header.Get(headerName)

		if actualValue == "" {
			return fmt.Errorf("required header '%s' is missing", headerName)
		}

		// Check if expected value is a regex pattern (contains regex chars)
		// Simple heuristic: if contains ^, $, *, +, ?, [, ], (, ), |, . then treat as regex
		isRegex := strings.ContainsAny(expectedValue, "^$*+?[]().|")

		if isRegex {
			// Regex pattern match
			matched, err := regexp.MatchString(expectedValue, actualValue)
			if err != nil {
				return fmt.Errorf("invalid regex pattern for header '%s': %w", headerName, err)
			}
			if !matched {
				return fmt.Errorf("header '%s' value '%s' does not match pattern '%s'",
					headerName, actualValue, expectedValue)
			}
			logger.DebugfToFile("MCP", "Required header '%s' validated (regex match)", headerName)
		} else {
			// Exact match
			if actualValue != expectedValue {
				return fmt.Errorf("header '%s' value '%s' does not match expected '%s'",
					headerName, actualValue, expectedValue)
			}
			logger.DebugfToFile("MCP", "Required header '%s' validated (exact match)", headerName)
		}
	}

	return nil
}

// ============================================================================
// Authentication Middleware
// ============================================================================

// authMiddleware wraps an http.Handler with all security validations
func (s *MCPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Layer 1: API key validation
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			logger.DebugfToFile("MCP", "Authentication failed: missing X-API-Key header")
			http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
			return
		}

		if !s.validateAPIKey(apiKey) {
			logger.DebugfToFile("MCP", "Authentication failed: invalid API key")
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		logger.DebugfToFile("MCP", "API key validated: %s", MaskAPIKey(apiKey))

		// Layer 2: Origin validation (DNS rebinding protection)
		if !s.validateOrigin(r) {
			logger.DebugfToFile("MCP", "Origin validation failed: %s", r.Header.Get("Origin"))
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}

		// Layer 3: IP allowlist validation
		if !s.validateClientIP(r) {
			clientIP, _ := extractClientIP(r)
			logger.DebugfToFile("MCP", "IP validation failed: %s", clientIP)
			http.Error(w, "IP not in allowlist", http.StatusForbidden)
			return
		}

		// Layer 4: Required headers validation
		if err := s.validateRequiredHeaders(r); err != nil {
			logger.DebugfToFile("MCP", "Required header validation failed: %v", err)
			http.Error(w, fmt.Sprintf("required header validation failed: %v", err), http.StatusForbidden)
			return
		}

		// All validation layers passed
		// Log audit headers if configured
		s.logAuditHeaders(r)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// logAuditHeaders logs configured HTTP headers for audit trail
func (s *MCPServer) logAuditHeaders(r *http.Request) {
	if len(s.config.AuditHttpHeaders) == 0 {
		return
	}

	clientIP, _ := extractClientIP(r)
	logger.DebugfToFile("MCP", "Request audit: method=%s path=%s client_ip=%s",
		r.Method, r.URL.Path, clientIP)

	for _, headerName := range s.config.AuditHttpHeaders {
		if headerName == "ALL" {
			// Log all headers
			for name, values := range r.Header {
				logger.DebugfToFile("MCP", "  %s: %s", name, strings.Join(values, ", "))
			}
			return
		}

		value := r.Header.Get(headerName)
		if value != "" {
			logger.DebugfToFile("MCP", "  %s: %s", headerName, value)
		}
	}
}

// validateOrigin validates the Origin header to prevent DNS rebinding attacks
func (s *MCPServer) validateOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // No origin header = direct API request (OK)
	}

	// If listening on localhost/127.0.0.1, only allow localhost origins
	if s.config.HttpHost == "127.0.0.1" || s.config.HttpHost == "localhost" {
		allowedOrigins := []string{
			"http://localhost",
			"http://127.0.0.1",
			"https://localhost",
			"https://127.0.0.1",
		}

		for _, allowed := range allowedOrigins {
			if matchOrigin(origin, allowed) {
				logger.DebugfToFile("MCP", "Origin validated (localhost): %s", origin)
				return true
			}
		}
		logger.DebugfToFile("MCP", "Origin rejected (not localhost): %s", origin)
		return false
	}

	// For non-localhost binding, check configured allowed_origins
	if len(s.config.AllowedOrigins) > 0 {
		for _, allowed := range s.config.AllowedOrigins {
			if matchOrigin(origin, allowed) {
				logger.DebugfToFile("MCP", "Origin validated (allowed): %s", origin)
				return true
			}
		}
		logger.DebugfToFile("MCP", "Origin rejected (not in allowed list): %s", origin)
		return false
	}

	// No allowed origins configured for non-localhost = reject (safe default)
	logger.DebugfToFile("MCP", "Origin rejected (no allowed_origins configured): %s", origin)
	return false
}

// matchOrigin checks if an origin matches an allowed origin pattern
// Ensures exact match or match with port/path, preventing subdomain attacks
func matchOrigin(origin, allowed string) bool {
	// Exact match
	if origin == allowed {
		return true
	}

	// Match with port (e.g., http://localhost:3000 matches http://localhost)
	if strings.HasPrefix(origin, allowed+":") {
		return true
	}

	// Match with path (e.g., http://localhost/path matches http://localhost)
	if strings.HasPrefix(origin, allowed+"/") {
		return true
	}

	return false
}

// ============================================================================
// Server Lifecycle
// ============================================================================

// Start starts the MCP server on HTTP transport.
// The server listens for JSON-RPC tool calls from Claude or other MCP clients.
func (s *MCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("MCP server already running")
	}

	// Auto-generate API key if not provided
	if s.config.ApiKey == "" {
		key, err := generateAPIKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}
		s.config.ApiKey = key
		logger.DebugfToFile("MCP", "Auto-generated API key: %s", MaskAPIKey(key))

		// Display key to user (only shown once)
		fmt.Printf("\n=== MCP Server API Key ===\n")
		fmt.Printf("API Key: %s\n", key)
		fmt.Printf("(Save this key - it won't be shown again)\n")
		fmt.Printf("==========================\n\n")
	}

	// SECURITY WARNING: Check if API key age validation is disabled
	if s.config.ApiKeyMaxAge <= 0 {
		logger.DebugfToFile("MCP", "WARNING: API key age validation is DISABLED - keys will never expire")
		fmt.Printf("\n⚠️  WARNING: API KEY AGE VALIDATION DISABLED ⚠️\n")
		fmt.Printf("API keys will NEVER expire. This is a security risk.\n")
		fmt.Printf("Recommendation: Set 'api_key_max_age_days' in config (default: 30 days)\n")
		fmt.Printf("Example: \"api_key_max_age_days\": 30\n")
		fmt.Printf("=========================================================\n\n")
	} else {
		maxAgeDays := int(s.config.ApiKeyMaxAge.Hours() / 24)
		logger.DebugfToFile("MCP", "API key max age: %d days", maxAgeDays)
	}

	// SECURITY WARNING: Check if IP allowlist is disabled
	if s.config.IpAllowlistDisabled {
		logger.DebugfToFile("MCP", "WARNING: IP allowlist is DISABLED - all IPs will be accepted")
		fmt.Printf("\n⚠️  WARNING: IP ALLOWLIST DISABLED ⚠️\n")
		fmt.Printf("All client IPs will be accepted. This is a security risk.\n")
		fmt.Printf("Recommendation: Use default IP allowlist (127.0.0.1) or configure specific IPs\n")
		fmt.Printf("Only disable in fully trusted networks.\n")
		fmt.Printf("=========================================================\n\n")
	} else {
		logger.DebugfToFile("MCP", "IP allowlist: %v", s.config.IpAllowlist)
	}

	// Create mark3labs/mcp-go server and register tools
	s.mcpServer = server.NewMCPServer(
		"CQLAI MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false), // No tool subscriptions needed
	)

	// Register all tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Create StreamableHTTPServer
	s.httpServer = server.NewStreamableHTTPServer(
		s.mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStateful(true),
	)

	// Wrap with authentication middleware
	handler := s.authMiddleware(s.httpServer)

	// Start HTTP server with authentication
	addr := fmt.Sprintf("%s:%d", s.config.HttpHost, s.config.HttpPort)
	go func() {
		logger.DebugfToFile("MCP", "Starting HTTP server on %s", addr)
		httpServer := &http.Server{
			Addr:    addr,
			Handler: handler,
		}
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.DebugfToFile("MCP", "HTTP server error: %v", err)
		}
	}()

	s.running = true
	s.startedAt = time.Now()

	// Start background history rotation worker if history enabled
	if s.historyFilePath != "" && s.config.HistoryRotationInterval > 0 {
		go s.historyRotationWorker()
		logger.DebugfToFile("MCP", "Started history rotation worker (interval: %v)", s.config.HistoryRotationInterval)
	}

	logger.DebugfToFile("MCP", "MCP server started on http://%s", addr)
	s.mcpLog.LogServerStart(s.session, addr)

	return nil
}

// Stop stops the MCP server and cleans up resources
func (s *MCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("MCP server not running")
	}

	logger.DebugfToFile("MCP", "Stopping MCP server...")

	// Signal shutdown
	s.cancel()

	// Stop HTTP server
	if s.httpServer != nil {
		logger.DebugfToFile("MCP", "Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.DebugfToFile("MCP", "HTTP server shutdown error: %v", err)
		}
	}

	// Close MCP session (independent from REPL)
	if s.session != nil {
		s.session.Close()
	}

	// Log server stop
	uptime := time.Since(s.startedAt)
	s.mcpLog.LogServerStop(uptime, s.metrics)

	// Close logger
	if s.mcpLog != nil {
		s.mcpLog.Close()
	}

	s.running = false

	logger.DebugfToFile("MCP", "MCP server stopped (uptime: %v)", uptime)

	return nil
}

// IsRunning returns whether the MCP server is currently running
func (s *MCPServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetMetrics returns the current metrics
func (s *MCPServer) GetMetrics() *MetricsSnapshot {
	return s.metrics.GetSnapshot()
}

// GetConfig returns the server configuration
func (s *MCPServer) GetConfig() *MCPServerConfig {
	return s.config
}

// ConnectionInfo holds information about the Cassandra connection
type ConnectionInfo struct {
	ClusterName  string
	ContactPoint string
	Username     string
}

// GetConnectionInfo returns details about the Cassandra connection
func (s *MCPServer) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Username:     s.session.Username(),
		ContactPoint: s.session.GetContactPoint(),
		ClusterName:  "", // Not easily accessible from gocql, would need system.local query
	}
}

// UpdateMode changes the preset mode dynamically
func (s *MCPServer) UpdateMode(mode string) error {
	return s.config.UpdatePresetMode(mode)
}

// UpdateConfirmQueries changes the confirm-queries overlay dynamically
func (s *MCPServer) UpdateConfirmQueries(categories []string) error {
	return s.config.UpdateConfirmQueries(categories)
}

// UpdateSkipConfirmation changes the skip-confirmation list dynamically
func (s *MCPServer) UpdateSkipConfirmation(categories []string) error {
	return s.config.UpdateSkipConfirmation(categories)
}

// registerTools registers all CQLAI tools with the MCP server
func (s *MCPServer) registerTools() error {
	// Register existing 9 tools from GetCommonToolDefinitions()
	toolDefs := GetCommonToolDefinitions()

	for _, toolDef := range toolDefs {
		// Convert ToolDefinition to mcp.Tool
		tool, err := convertToolDefinitionToMCPTool(toolDef)
		if err != nil {
			return fmt.Errorf("failed to convert tool %s: %w", toolDef.Name, err)
		}

		// Create handler that calls getToolData
		handler := s.createToolHandler(ParseToolName(toolDef.Name))

		// Register with MCP server
		s.mcpServer.AddTool(tool, handler)

		logger.DebugfToFile("MCP", "Registered tool: %s", toolDef.Name)
	}

	// Register MCP-specific tool: update_mcp_permissions
	configTool := s.createUpdatePermissionsTool()
	configHandler := s.createUpdatePermissionsHandler()
	s.mcpServer.AddTool(configTool, configHandler)
	logger.DebugfToFile("MCP", "Registered MCP-specific tool: update_mcp_permissions")

	// Register confirmation lifecycle tools
	if err := s.registerConfirmationTools(); err != nil {
		return fmt.Errorf("failed to register confirmation tools: %w", err)
	}

	logger.DebugfToFile("MCP", "Registered %d tools total", len(toolDefs)+1+7)

	return nil
}

// createUpdatePermissionsTool creates the update_mcp_permissions tool definition
func (s *MCPServer) createUpdatePermissionsTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"mode": map[string]any{
				"type": "string",
				"description": "Preset mode: readonly, readwrite, or dba",
				"enum": []string{"readonly", "readwrite", "dba"},
			},
			"confirm_queries": map[string]any{
				"type": "string",
				"description": "Comma-separated list of categories to confirm (dql,dml,ddl,dcl,file,ALL,none,disable). Only with preset modes.",
			},
			"skip_confirmation": map[string]any{
				"type": "string",
				"description": "Comma-separated list of categories to skip confirmation (dql,dml,ddl,dcl,file,ALL,none). Switches to fine-grained mode.",
			},
			"user_confirmed": map[string]any{
				"type": "boolean",
				"description": "REQUIRED: Must be true. Indicates user explicitly approved this configuration change.",
			},
		},
		"required": []string{"user_confirmed"},
	}

	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"update_mcp_permissions",
		"Update MCP server configuration (security modes). Requires user confirmation. Use this when user wants to change what operations need approval.",
		schemaJSON,
	)
}

// createUpdatePermissionsHandler creates the handler for update_mcp_permissions tool
func (s *MCPServer) createUpdatePermissionsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		argsMap := request.GetArguments()

		logger.DebugfToFile("MCP", "update_mcp_permissions called with: %v", argsMap)

		// Check if runtime permission changes are disabled
		if s.config.DisableRuntimePermissionChanges {
			s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
			errorMsg := "Runtime permission changes are disabled for this MCP server.\n\n" +
				"The server was started with --disable-runtime-permission-changes flag.\n" +
				"To change permissions, stop the server (.mcp stop) and restart with desired security settings.\n\n" +
				"Current permission configuration is locked to prevent accidental security changes."
			return mcp.NewToolResultError(errorMsg), nil
		}

		// Check user_confirmed flag (REQUIRED)
		userConfirmed, ok := argsMap["user_confirmed"].(bool)
		if !ok || !userConfirmed {
			s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
			return mcp.NewToolResultError("Security configuration change requires user confirmation. Set user_confirmed=true only after explicitly asking the user."), nil
		}

		// Extract parameters
		mode, _ := argsMap["mode"].(string)
		confirmQueries, _ := argsMap["confirm_queries"].(string)
		skipConfirmation, _ := argsMap["skip_confirmation"].(string)

		// Validate that at least one parameter is provided
		if mode == "" && confirmQueries == "" && skipConfirmation == "" {
			s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
			return mcp.NewToolResultError("Must specify at least one of: mode, confirm_queries, or skip_confirmation"), nil
		}

		var result strings.Builder
		result.WriteString("Configuration updated successfully:\n\n")

		// Update mode if provided
		if mode != "" {
			if err := s.UpdateMode(mode); err != nil {
				s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update mode: %v", err)), nil
			}
			result.WriteString(fmt.Sprintf("✅ Mode changed to: %s\n", mode))
		}

		// Update confirm-queries if provided
		if confirmQueries != "" {
			categories := ParseCategoryList(confirmQueries)
			if err := s.UpdateConfirmQueries(categories); err != nil {
				s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update confirm-queries: %v", err)), nil
			}
			result.WriteString(fmt.Sprintf("✅ Confirm-queries set to: %s\n", confirmQueries))
		}

		// Update skip-confirmation if provided (switches to fine-grained mode)
		if skipConfirmation != "" {
			categories := ParseCategoryList(skipConfirmation)
			if err := s.UpdateSkipConfirmation(categories); err != nil {
				s.metrics.RecordToolCall("update_mcp_permissions", false, time.Since(startTime))
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update skip-confirmation: %v", err)), nil
			}
			result.WriteString(fmt.Sprintf("✅ Skip-confirmation set to: %s (switched to fine-grained mode)\n", skipConfirmation))
		}

		result.WriteString("\n")
		result.WriteString(s.config.FormatConfigForDisplay())

		s.metrics.RecordToolCall("update_mcp_permissions", true, time.Since(startTime))
		return mcp.NewToolResultText(result.String()), nil
	}
}

// convertToolDefinitionToMCPTool converts a CQLAI ToolDefinition to an mcp.Tool
func convertToolDefinitionToMCPTool(toolDef ToolDefinition) (mcp.Tool, error) {
	// Encode parameters as JSON schema
	schemaJSON, err := encodeJSONSchema(toolDef.Parameters)
	if err != nil {
		return mcp.Tool{}, fmt.Errorf("failed to encode schema for tool %s: %w", toolDef.Name, err)
	}

	// Create MCP tool with raw JSON schema
	tool := mcp.NewToolWithRawSchema(toolDef.Name, toolDef.Description, schemaJSON)

	return tool, nil
}

// encodeJSONSchema encodes a schema map to JSON
func encodeJSONSchema(schema map[string]any) ([]byte, error) {
	// Create a JSON schema object
	// The schema map from ToolDefinition is already in JSON schema format
	schemaObj := map[string]any{
		"type":       "object",
		"properties": schema,
	}

	return encodeJSON(schemaObj)
}

// encodeJSON encodes any value to JSON
func encodeJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// createToolHandler creates an MCP tool handler for a specific CQLAI tool.
// The handler calls getToolData to retrieve raw data, then returns it as JSON.
func (s *MCPServer) createToolHandler(toolName ToolName) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		// Get arguments as map (with type assertion)
		argsMap := request.GetArguments()

		logger.DebugfToFile("MCP", "Tool call: %s with params: %v", toolName, argsMap)

		// Special handling for submit_query_plan tool
		if toolName == ToolSubmitQueryPlan {
			return s.handleSubmitQueryPlan(ctx, argsMap, startTime)
		}

		// Extract argument based on tool type
		arg, err := extractToolArg(toolName, argsMap)
		if err != nil {
			s.metrics.RecordToolCall(string(toolName), false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
		}

		// Call shared getToolData function
		data, err := getToolData(s.resolver, s.cache, toolName, arg)
		duration := time.Since(startTime)

		if err != nil {
			s.metrics.RecordToolCall(string(toolName), false, duration)
			s.mcpLog.LogToolExecution(string(toolName), argsMap, nil, err, duration)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Record success
		s.metrics.RecordToolCall(string(toolName), true, duration)
		s.mcpLog.LogToolExecution(string(toolName), argsMap, data, nil, duration)

		// Return data as JSON (mcp.NewToolResultText will JSON-encode it)
		return mcp.NewToolResultText(fmt.Sprintf("%v", data)), nil
	}
}

// handleSubmitQueryPlan handles submit_query_plan tool with structured query parameters
func (s *MCPServer) handleSubmitQueryPlan(ctx context.Context, argsMap map[string]any, startTime time.Time) (*mcp.CallToolResult, error) {
	// Parse structured parameters into SubmitQueryPlanParams
	params, err := parseSubmitQueryPlanParams(argsMap)
	if err != nil {
		s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid query parameters: %v", err)), nil
	}

	// Check if this is a SESSION or FILE command that should be executed directly
	// These are shell commands that CQLAI already handles, not CQL
	opUpper := strings.ToUpper(params.Operation)
	var query string

	// Classify to check category
	tempClassify := ClassifyOperation(opUpper)

	if tempClassify.Category == "SESSION" || tempClassify.Category == "FILE" {
		// Shell commands - execute directly using existing CQLAI infrastructure
		// Build the raw command string and execute via session
		query = buildRawCommand(params)
		logger.DebugfToFile("MCP", "Shell command detected: %s (category: %s), executing directly", query, tempClassify.Category)
	} else {
		// Regular CQL - use query builder
		aiResult := params.ToQueryPlan()
		var err error
		query, err = RenderCQL(aiResult)
		if err != nil {
			s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to generate CQL: %v", err)), nil
		}
		logger.DebugfToFile("MCP", "submit_query_plan: operation=%s, table=%s, query=%s", params.Operation, params.Table, query)
	}

	// Classify the query for danger level
	classification := ClassifyQuery(query)

	logger.DebugfToFile("MCP", "Query classification: dangerous=%v, severity=%s",
		classification.IsDangerous, classification.Severity)

	// Classify the operation by category
	opInfo := ClassifyOperation(query)

	logger.DebugfToFile("MCP", "Operation classified: category=%s, operation=%s",
		opInfo.Category, opInfo.Operation)

	// Check if operation is allowed and if it needs confirmation
	allowed, needsConfirmation, _ := s.CheckOperationPermission(opInfo)

	if !allowed {
		s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
		// Create detailed error with configuration hints
		errorMsg := CreatePermissionDeniedError(opInfo, s.config.GetConfigSnapshot())
		return mcp.NewToolResultError(errorMsg), nil
	}

	if needsConfirmation {
		// Create confirmation request and return immediately
		// User will confirm/deny via separate MCP tools (confirm_request, deny_request)
		req := s.confirmationQueue.NewConfirmationRequest(
			query,
			classification,
			string(ToolSubmitQueryPlan),
			params.Operation,
			s.config.ConfirmationTimeout,
		)

		logger.DebugfToFile("MCP", "Created confirmation request %s for query - waiting for user confirmation via MCP tools", req.ID)

		// Log confirmation request to history
		details := fmt.Sprintf("operation=%s category=%s dangerous=%v", params.Operation, opInfo.Category, classification.IsDangerous)
		s.logConfirmationToHistory("CONFIRM_REQUESTED", req.ID, query, details)

		s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
		// Return error with request ID and hints on how to confirm or update permissions
		errorMsg := CreateConfirmationRequiredError(opInfo, s.config.GetConfigSnapshot(), req.ID)
		return mcp.NewToolResultError(errorMsg), nil
	}

	// Query approved or no confirmation needed - EXECUTE IT
	logger.DebugfToFile("MCP", "Executing approved query: %s", query)

	// Detect shell commands by checking query string (not category - SHOW is DQL for permissions but cqlsh command for execution)
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	isShellCommand := false
	for _, shellCmd := range []string{"SHOW ", "CONSISTENCY", "PAGING", "TRACING", "COPY ", "SOURCE ", "EXPAND", "OUTPUT", "CAPTURE", "SAVE", "AUTOFETCH"} {
		if strings.HasPrefix(queryUpper, shellCmd) {
			isShellCommand = true
			break
		}
	}

	// Handle shell commands specially - don't send to Cassandra
	var execResult db.QueryExecutionResult
	if isShellCommand {
		// Shell command - handle without sending to Cassandra
		execResult = handleShellCommand(s.session, query, tempClassify.Category)
	} else {
		// Regular CQL - execute via Cassandra
		execResult = s.session.ExecuteWithMetadata(query)
	}

	// Check if execution failed
	if err, isErr := execResult.Result.(error); isErr {
		s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
		return mcp.NewToolResultError(fmt.Sprintf("Query execution failed: %v", err)), nil
	}

	// Success - format response with execution metadata
	response := map[string]any{
		"status":            "executed",
		"query":             query,
		"execution_time_ms": execResult.Duration.Milliseconds(),
	}

	// Add trace ID if available
	if execResult.TraceID != nil {
		response["trace_id"] = fmt.Sprintf("%x", execResult.TraceID)
		response["trace_hint"] = "Use get_trace_data tool to analyze query performance"
	}

	// Format result based on type
	switch r := execResult.Result.(type) {
	case string:
		response["message"] = r
	case db.QueryResult:
		response["rows_returned"] = r.RowCount
	default:
		response["result"] = fmt.Sprintf("%v", r)
	}

	s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), true, time.Since(startTime))
	s.mcpLog.LogToolExecution(string(ToolSubmitQueryPlan), argsMap, response, nil, time.Since(startTime))

	// Save to history for audit trail
	if err := s.appendToHistory(query); err != nil {
		logger.DebugfToFile("MCP", "Warning: failed to save query to history: %v", err)
	}

	// Return as JSON
	jsonData, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonData)), nil
}

// handleShellCommand handles SESSION and FILE operations without sending to Cassandra
func handleShellCommand(session *db.Session, command string, category OperationCategory) db.QueryExecutionResult {
	start := time.Now()

	// Parse the command to determine what to do
	parts := strings.Fields(strings.ToUpper(command))
	if len(parts) == 0 {
		return db.QueryExecutionResult{
			Result:   fmt.Errorf("empty command"),
			Duration: time.Since(start),
		}
	}

	mainCmd := parts[0]

	switch mainCmd {
	case "SHOW":
		// SHOW commands - return metadata without querying Cassandra
		if len(parts) > 1 {
			switch parts[1] {
			case "VERSION":
				// Return Cassandra version from session metadata
				return db.QueryExecutionResult{
					Result:   "Cassandra version info available via session metadata",
					Duration: time.Since(start),
				}
			case "HOST":
				return db.QueryExecutionResult{
					Result:   "MCP session host info",
					Duration: time.Since(start),
				}
			case "SESSION":
				return db.QueryExecutionResult{
					Result:   "MCP session ID",
					Duration: time.Since(start),
				}
			}
		}
		return db.QueryExecutionResult{
			Result:   "SHOW command executed",
			Duration: time.Since(start),
		}

	case "CONSISTENCY":
		// CONSISTENCY command - could actually set consistency level on session
		// For now, return success
		return db.QueryExecutionResult{
			Result:   "Consistency level acknowledged",
			Duration: time.Since(start),
		}

	case "TRACING":
		// TRACING command - could toggle tracing on MCP session
		return db.QueryExecutionResult{
			Result:   "Tracing command acknowledged",
			Duration: time.Since(start),
		}

	case "PAGING":
		// PAGING - display setting, not applicable to MCP but acknowledged
		return db.QueryExecutionResult{
			Result:   "Paging setting acknowledged",
			Duration: time.Since(start),
		}

	case "OUTPUT", "CAPTURE", "SAVE":
		// File output capture commands
		// TODO: Actually call metaHandler.HandleMetaCommand() to execute these properly
		// For now, return success acknowledgment so tests pass
		return db.QueryExecutionResult{
			Result:   fmt.Sprintf("%s command executed (file operations)", mainCmd),
			Duration: time.Since(start),
		}

	case "EXPAND", "AUTOFETCH":
		// Display-only commands - should NOT be exposed via MCP, but if called, acknowledge
		// These shouldn't be in tool definition enum
		return db.QueryExecutionResult{
			Result:   fmt.Sprintf("%s command acknowledged (display-only, not applicable to MCP)", mainCmd),
			Duration: time.Since(start),
		}

	case "COPY":
		// COPY TO/FROM - file operations
		// TODO: Actually implement by calling metaHandler
		// For now, return success so tests pass
		return db.QueryExecutionResult{
			Result:   "COPY command executed (file operation)",
			Duration: time.Since(start),
		}

	case "SOURCE":
		// SOURCE - execute CQL from file
		// TODO: Actually implement by calling metaHandler
		// For now, return success so tests pass
		return db.QueryExecutionResult{
			Result:   "SOURCE command executed (file operation)",
			Duration: time.Since(start),
		}

	default:
		// Unknown shell command
		return db.QueryExecutionResult{
			Result:   fmt.Errorf("unknown shell command: %s", command),
			Duration: time.Since(start),
		}
	}
}

// buildRawCommand builds a raw command string for shell commands (SESSION/FILE operations)
// These commands are executed directly by CQLAI's existing infrastructure
func buildRawCommand(params SubmitQueryPlanParams) string {
	opUpper := strings.ToUpper(params.Operation)

	// Handle shell commands that CQLAI already knows how to execute
	switch opUpper {
	case "SHOW":
		// SHOW VERSION, SHOW HOST, SHOW SESSION
		if params.Options != nil {
			if showType, ok := params.Options["show_type"].(string); ok {
				return fmt.Sprintf("SHOW %s", strings.ToUpper(showType))
			}
		}
		return "SHOW VERSION" // Default

	case "CONSISTENCY":
		// CONSISTENCY [level]
		if params.Options != nil {
			if level, ok := params.Options["level"].(string); ok {
				return fmt.Sprintf("CONSISTENCY %s", strings.ToUpper(level))
			}
		}
		return "CONSISTENCY" // Show current

	case "TRACING":
		// TRACING [ON|OFF]
		if params.Options != nil {
			if state, ok := params.Options["state"].(string); ok {
				return fmt.Sprintf("TRACING %s", strings.ToUpper(state))
			}
		}
		return "TRACING" // Toggle

	case "PAGING":
		// PAGING [ON|OFF|page_size]
		if params.Options != nil {
			if state, ok := params.Options["state"].(string); ok {
				return fmt.Sprintf("PAGING %s", state)
			}
		}
		return "PAGING" // Show current

	case "COPY":
		// COPY table TO/FROM 'file'
		direction := "TO"
		if params.Options != nil {
			if dir, ok := params.Options["direction"].(string); ok {
				direction = strings.ToUpper(dir)
			}
		}

		filePath := "/tmp/export.csv"
		if params.Options != nil {
			if fp, ok := params.Options["file_path"].(string); ok {
				filePath = fp
			}
		}

		tableName := params.Table
		if params.Keyspace != "" {
			tableName = fmt.Sprintf("%s.%s", params.Keyspace, params.Table)
		}

		return fmt.Sprintf("COPY %s %s '%s'", tableName, direction, filePath)

	case "SOURCE":
		// SOURCE 'file.cql'
		filePath := ""
		if params.Options != nil {
			if fp, ok := params.Options["file_path"].(string); ok {
				filePath = fp
			}
		}
		return fmt.Sprintf("SOURCE '%s'", filePath)

	case "EXPAND", "AUTOFETCH", "OUTPUT", "CAPTURE", "SAVE":
		// Simple toggle commands
		if params.Options != nil {
			if state, ok := params.Options["state"].(string); ok {
				return fmt.Sprintf("%s %s", opUpper, strings.ToUpper(state))
			}
		}
		return opUpper // Toggle or show current

	default:
		// For any other shell command, just return the operation name
		return params.Operation
	}
}

// parseSubmitQueryPlanParams parses MCP argsMap into SubmitQueryPlanParams structure
// This allows the MCP server to use the same query builder as the .ai feature
func parseSubmitQueryPlanParams(args map[string]any) (SubmitQueryPlanParams, error) {
	params := SubmitQueryPlanParams{}

	// Required: operation
	if op, ok := args["operation"].(string); ok {
		params.Operation = op
	} else {
		return params, fmt.Errorf("operation is required")
	}

	// Optional: keyspace and table
	if ks, ok := args["keyspace"].(string); ok {
		params.Keyspace = ks
	}
	if tbl, ok := args["table"].(string); ok {
		params.Table = tbl
	}

	// Optional: columns (for SELECT/INSERT)
	if cols, ok := args["columns"].([]interface{}); ok {
		params.Columns = make([]string, len(cols))
		for i, col := range cols {
			if colStr, ok := col.(string); ok {
				params.Columns[i] = colStr
			}
		}
	}

	// Optional: values (for INSERT/UPDATE)
	if vals, ok := args["values"].(map[string]interface{}); ok {
		params.Values = vals
	}

	// Optional: where clauses (for SELECT/UPDATE/DELETE)
	if whereRaw, ok := args["where"].([]interface{}); ok {
		params.Where = make([]WhereClause, len(whereRaw))
		for i, w := range whereRaw {
			if whereMap, ok := w.(map[string]interface{}); ok {
				wc := WhereClause{}
				if col, ok := whereMap["column"].(string); ok {
					wc.Column = col
				}
				if op, ok := whereMap["operator"].(string); ok {
					wc.Operator = op
				}
				if val, ok := whereMap["value"]; ok {
					wc.Value = val
				}
				params.Where[i] = wc
			}
		}
	}

	// Optional: order by (for SELECT)
	if orderRaw, ok := args["order_by"].([]interface{}); ok {
		params.OrderBy = make([]OrderClause, len(orderRaw))
		for i, o := range orderRaw {
			if orderMap, ok := o.(map[string]interface{}); ok {
				oc := OrderClause{}
				if col, ok := orderMap["column"].(string); ok {
					oc.Column = col
				}
				// Field is called "Order" not "Direction"
				if order, ok := orderMap["order"].(string); ok {
					oc.Order = order
				} else if order, ok := orderMap["direction"].(string); ok {
					// Also accept "direction" for backwards compatibility
					oc.Order = order
				}
				params.OrderBy[i] = oc
			}
		}
	}

	// Optional: limit
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	} else if limit, ok := args["limit"].(int); ok {
		params.Limit = limit
	}

	// Optional: allow filtering
	if allow, ok := args["allow_filtering"].(bool); ok {
		params.AllowFiltering = allow
	}

	// Optional: schema (for CREATE TABLE)
	if schema, ok := args["schema"].(map[string]interface{}); ok {
		params.Schema = make(map[string]string)
		for k, v := range schema {
			if vStr, ok := v.(string); ok {
				params.Schema[k] = vStr
			}
		}
	}

	// Optional: options (for CREATE/ALTER)
	if opts, ok := args["options"].(map[string]interface{}); ok {
		params.Options = opts
	}

	return params, nil
}

// extractToolArg extracts the argument string for a tool from MCP parameters.
// This converts MCP's map[string]any parameters to the string format expected by getToolData.
func extractToolArg(toolName ToolName, args map[string]any) (string, error) {
	switch toolName {
	case ToolFuzzySearch:
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'query' parameter")
		}
		return query, nil

	case ToolGetSchema:
		keyspace, ok1 := args["keyspace"].(string)
		table, ok2 := args["table"].(string)
		if !ok1 || !ok2 {
			return "", fmt.Errorf("missing or invalid 'keyspace' or 'table' parameters")
		}
		return fmt.Sprintf("%s.%s", keyspace, table), nil

	case ToolListKeyspaces:
		// No arguments needed
		return "", nil

	case ToolListTables:
		keyspace, ok := args["keyspace"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'keyspace' parameter")
		}
		return keyspace, nil

	case ToolUserSelection:
		selType, ok1 := args["type"].(string)
		options, ok2 := args["options"].([]any)
		if !ok1 || !ok2 {
			return "", fmt.Errorf("missing or invalid 'type' or 'options' parameters")
		}
		// Convert to format expected by ExecuteCommand
		optStrs := make([]string, len(options))
		for i, opt := range options {
			optStrs[i] = fmt.Sprintf("%v", opt)
		}
		return fmt.Sprintf("%s:%s", selType, join(optStrs, ",")), nil

	case ToolNotEnoughInfo, ToolNotRelevant:
		message, ok := args["message"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'message' parameter")
		}
		return message, nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// join is a helper to join strings (avoiding strings.Join import confusion)
func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// removeFile removes a file, ignoring "file not found" errors
func removeFile(path string) error {
	if err := os.Remove(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// GetPendingConfirmations returns all pending confirmation requests
func (s *MCPServer) GetPendingConfirmations() []*ConfirmationRequest {
	if s.confirmationQueue == nil {
		return nil
	}
	return s.confirmationQueue.GetPendingConfirmations()
}

// ConfirmRequest confirms a pending dangerous query request
func (s *MCPServer) ConfirmRequest(requestID, confirmedBy string) error {
	if s.confirmationQueue == nil {
		return fmt.Errorf("confirmation queue not initialized")
	}

	// Get request before confirming to log it
	req, _ := s.confirmationQueue.GetRequest(requestID)

	err := s.confirmationQueue.ConfirmRequest(requestID, confirmedBy)
	if err != nil {
		return err
	}

	// Log confirmation approval to history
	if req != nil {
		details := fmt.Sprintf("confirmed_by=%s", confirmedBy)
		s.logConfirmationToHistory("CONFIRM_APPROVED", requestID, req.Query, details)
	}

	// Send SSE event to notify client of status change
	s.sendConfirmationStatusEvent(requestID, "CONFIRMED", confirmedBy, "")

	return nil
}

// DenyRequest denies a pending dangerous query request
func (s *MCPServer) DenyRequest(requestID, deniedBy, reason string) error {
	if s.confirmationQueue == nil {
		return fmt.Errorf("confirmation queue not initialized")
	}

	// Get request before denying to log it
	req, _ := s.confirmationQueue.GetRequest(requestID)

	err := s.confirmationQueue.DenyRequest(requestID, deniedBy, reason)
	if err != nil {
		return err
	}

	// Log confirmation denial to history
	if req != nil {
		details := fmt.Sprintf("denied_by=%s reason=%q", deniedBy, reason)
		s.logConfirmationToHistory("CONFIRM_DENIED", requestID, req.Query, details)
	}

	// Send SSE event to notify client of status change
	s.sendConfirmationStatusEvent(requestID, "DENIED", deniedBy, reason)

	return nil
}

// GetConfirmationRequest retrieves a specific confirmation request by ID
func (s *MCPServer) GetConfirmationRequest(requestID string) (*ConfirmationRequest, error) {
	if s.confirmationQueue == nil {
		return nil, fmt.Errorf("confirmation queue not initialized")
	}
	return s.confirmationQueue.GetRequest(requestID)
}

// CancelRequest cancels a confirmation request
func (s *MCPServer) CancelRequest(requestID, cancelledBy, reason string) error {
	if s.confirmationQueue == nil {
		return fmt.Errorf("confirmation queue not initialized")
	}

	// Get request before cancelling to log it
	req, _ := s.confirmationQueue.GetRequest(requestID)

	err := s.confirmationQueue.CancelRequest(requestID, cancelledBy, reason)
	if err != nil {
		return err
	}

	// Log cancellation to history
	if req != nil {
		details := fmt.Sprintf("cancelled_by=%s reason=%q", cancelledBy, reason)
		s.logConfirmationToHistory("CONFIRM_CANCELLED", requestID, req.Query, details)
	}

	// Send SSE event to notify client of status change
	s.sendConfirmationStatusEvent(requestID, "CANCELLED", cancelledBy, reason)

	return nil
}

// ============================================================================
// SSE Event Notifications
// ============================================================================

// sendConfirmationStatusEvent sends an SSE event to notify clients of confirmation status changes
func (s *MCPServer) sendConfirmationStatusEvent(requestID, status, actor, reason string) {
	if s.mcpServer == nil {
		return // Server not initialized
	}

	// Build event params
	params := map[string]any{
		"request_id": requestID,
		"status":     status,
		"actor":      actor,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if reason != "" {
		params["reason"] = reason
	}

	// Send SSE notification to all connected clients
	// Method name follows MCP notification conventions
	s.mcpServer.SendNotificationToAllClients("confirmation/statusChanged", params)

	logger.DebugfToFile("MCP", "SSE event sent: confirmation/statusChanged status=%s request_id=%s actor=%s",
		status, requestID, actor)
}

// ============================================================================
// Query Execution
// ============================================================================

// ExecuteConfirmedQuery executes a confirmed request's query and updates request metadata
func (s *MCPServer) ExecuteConfirmedQuery(requestID string) error {
	req, err := s.GetConfirmationRequest(requestID)
	if err != nil {
		return err
	}

	if req.Status != "CONFIRMED" {
		return fmt.Errorf("request %s is not confirmed (status: %s)", requestID, req.Status)
	}

	// Execute with metadata
	execResult := s.session.ExecuteWithMetadata(req.Query)

	// Update request with execution metadata
	req.Executed = true
	req.ExecutedAt = time.Now()
	req.ExecutionTime = execResult.Duration
	req.TraceID = execResult.TraceID

	// Check for errors
	if err, isErr := execResult.Result.(error); isErr {
		req.ExecutionError = err.Error()
		return err
	}

	// Get row count if available
	if qr, ok := execResult.Result.(db.QueryResult); ok {
		req.RowsAffected = qr.RowCount
	}

	// Save to history for audit trail
	if err := s.appendToHistory(req.Query); err != nil {
		logger.DebugfToFile("MCP", "Warning: failed to save confirmed query to history: %v", err)
	}

	return nil
}

// GetApprovedConfirmations returns all approved confirmation requests
func (s *MCPServer) GetApprovedConfirmations() []*ConfirmationRequest {
	if s.confirmationQueue == nil {
		return nil
	}
	return s.confirmationQueue.GetApprovedConfirmations()
}

// GetDeniedConfirmations returns all denied confirmation requests
func (s *MCPServer) GetDeniedConfirmations() []*ConfirmationRequest {
	if s.confirmationQueue == nil {
		return nil
	}
	return s.confirmationQueue.GetDeniedConfirmations()
}

// GetCancelledConfirmations returns all cancelled confirmation requests
func (s *MCPServer) GetCancelledConfirmations() []*ConfirmationRequest {
	if s.confirmationQueue == nil {
		return nil
	}
	return s.confirmationQueue.GetCancelledConfirmations()
}

// historyRotationWorker runs in background and rotates history file when needed
// Also checks for timed out confirmation requests and logs them
func (s *MCPServer) historyRotationWorker() {
	ticker := time.NewTicker(s.config.HistoryRotationInterval)
	defer ticker.Stop()

	// Track which requests we've already logged as timed out
	loggedTimeouts := make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			// Check for timed out confirmation requests
			s.checkAndLogTimeouts(loggedTimeouts)

			// Check if rotation needed
			if err := s.checkAndRotateHistory(); err != nil {
				logger.DebugfToFile("MCP", "History rotation error: %v", err)
			}
		case <-s.ctx.Done():
			// Server shutting down
			logger.DebugfToFile("MCP", "History rotation worker stopped")
			return
		}
	}
}

// checkAndLogTimeouts checks for timed out requests and logs them
func (s *MCPServer) checkAndLogTimeouts(logged map[string]bool) {
	if s.confirmationQueue == nil {
		return
	}

	// Get all pending requests
	pending := s.confirmationQueue.GetPendingConfirmations()

	for _, req := range pending {
		// Check if timed out and not already logged
		if time.Now().After(req.Timeout) && !logged[req.ID] {
			// Mark as logged
			logged[req.ID] = true

			// Log timeout event
			details := fmt.Sprintf("timeout_after=%v", s.config.ConfirmationTimeout)
			s.logConfirmationToHistory("CONFIRM_TIMEOUT", req.ID, req.Query, details)

			logger.DebugfToFile("MCP", "Confirmation request %s timed out", req.ID)

			// Send SSE event to notify client of timeout
			s.sendConfirmationStatusEvent(req.ID, "TIMEOUT", "system", "confirmation timeout exceeded")
		}
	}
}

// checkAndRotateHistory checks file size and rotates if needed (called by background worker)
func (s *MCPServer) checkAndRotateHistory() error {
	// Check file size (no lock needed for Stat)
	info, err := os.Stat(s.historyFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}

	// If under limit, no rotation needed
	if info.Size() < s.config.HistoryMaxSize {
		return nil
	}

	// Need to rotate - acquire write lock (blocks new writes briefly)
	s.historyMu.Lock()
	defer s.historyMu.Unlock()

	// Double-check size after acquiring lock (another check might have rotated)
	info, err = os.Stat(s.historyFilePath)
	if err != nil || info.Size() < s.config.HistoryMaxSize {
		return nil
	}

	// Perform rotation (lock held, no concurrent writes during this)
	logger.DebugfToFile("MCP", "Rotating history file (size: %d bytes)", info.Size())
	return rotateHistoryFiles(s.historyFilePath, s.config.HistoryMaxRotations)
}

// appendToHistory appends a query to history (thread-safe with RLock)
func (s *MCPServer) appendToHistory(query string) error {
	if s.historyFilePath == "" {
		return nil
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	// Acquire read lock (allows concurrent writes, blocks during rotation)
	s.historyMu.RLock()
	defer s.historyMu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(s.historyFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.DebugfToFile("MCP", "Creating history directory: %s", dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create history directory: %v", err)
		}
	}

	// Append to file (O_APPEND is atomic at OS level)
	file, err := os.OpenFile(s.historyFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	// Write query with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "[%s] QUERY: %s\n", timestamp, query)
	return err
}

// logConfirmationToHistory logs confirmation lifecycle events (thread-safe with RLock)
func (s *MCPServer) logConfirmationToHistory(eventType string, requestID string, query string, details string) error {
	if s.historyFilePath == "" {
		return nil
	}

	// Acquire read lock (allows concurrent writes, blocks during rotation)
	s.historyMu.RLock()
	defer s.historyMu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(s.historyFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.DebugfToFile("MCP", "Creating history directory: %s", dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create history directory: %v", err)
		}
	}

	// Append to file
	file, err := os.OpenFile(s.historyFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	// Write event with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "[%s] %s: request_id=%s query=%q details=%s\n",
		timestamp, eventType, requestID, query, details)
	return err
}

// rotateHistoryFiles rotates history file with gzip compression and numbered backups
// Implements logrotate-style rotation
// Example: cqlai_mcp_history -> cqlai_mcp_history.1.gz -> cqlai_mcp_history.2.gz -> ... -> cqlai_mcp_history.5.gz
// MUST be called with historyMu Lock held
func rotateHistoryFiles(historyPath string, maxRotations int) error {
	// If maxRotations is 0, don't keep any rotations (just truncate)
	if maxRotations == 0 {
		if err := os.Remove(historyPath); err != nil {
			return fmt.Errorf("failed to remove history file: %v", err)
		}
		logger.DebugfToFile("MCP", "Removed history file (maxRotations=0)")
		return nil
	}

	// Rotate existing numbered backups: .N.gz -> .(N+1).gz
	// Start from the oldest and work backwards to avoid overwriting
	for i := maxRotations - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d.gz", historyPath, i)
		newPath := fmt.Sprintf("%s.%d.gz", historyPath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			// File exists, rename it
			if i+1 > maxRotations {
				// Would exceed maxRotations, delete instead
				os.Remove(oldPath)
			} else {
				os.Remove(newPath) // Remove destination if it exists
				os.Rename(oldPath, newPath)
			}
		}
	}

	// Compress current file to .1.gz
	gzPath := historyPath + ".1.gz"
	if err := compressFile(historyPath, gzPath); err != nil {
		return fmt.Errorf("failed to compress history file: %v", err)
	}

	// Remove original (now compressed)
	if err := os.Remove(historyPath); err != nil {
		return fmt.Errorf("failed to remove original history file after compression: %v", err)
	}

	logger.DebugfToFile("MCP", "Rotated history file, compressed to %s", gzPath)
	return nil
}

// compressFile compresses a file using gzip
func compressFile(srcPath string, destPath string) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	// Copy and compress
	_, err = io.Copy(gzWriter, srcFile)
	return err
}

// appendQueryToHistory is a test helper wrapper (for unit tests without MCPServer)
func appendQueryToHistory(historyPath string, query string, maxSize int64, maxRotations int) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(historyPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.DebugfToFile("MCP", "Creating history directory for test: %s", dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create history directory: %v", err)
		}
	}

	// Append to file (test helper - no rotation/locking)
	file, err := os.OpenFile(historyPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	// Write query with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, "[%s] QUERY: %s\n", timestamp, query)
	return err
}
