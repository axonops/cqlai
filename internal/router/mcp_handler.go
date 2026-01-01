package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
)

// MCPHandler manages MCP server lifecycle
type MCPHandler struct {
	replSession *db.Session
	mcpServer   *ai.MCPServer
}

// mcpHandlerInstance is the singleton instance
var mcpHandlerInstance *MCPHandler

// InitMCPHandler initializes the MCP handler
func InitMCPHandler(session *db.Session) error {
	mcpHandlerInstance = &MCPHandler{
		replSession: session,
		mcpServer:   nil, // Created when .mcp start is called
	}
	return nil
}

// GetMCPHandler returns the singleton MCP handler
func GetMCPHandler() *MCPHandler {
	return mcpHandlerInstance
}

// GetPendingConfirmationCount returns the number of pending confirmations
func (h *MCPHandler) GetPendingConfirmationCount() int {
	if h == nil || h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return 0
	}
	pending := h.mcpServer.GetPendingConfirmations()
	return len(pending)
}

// HandleMCPCommand processes MCP commands
func (h *MCPHandler) HandleMCPCommand(command string) string {
	if h == nil {
		return "MCP system not initialized. Please restart the application."
	}

	// Remove .mcp prefix
	args := strings.TrimSpace(strings.TrimPrefix(command, ".mcp"))
	args = strings.TrimSpace(strings.TrimPrefix(args, ".MCP"))

	if args == "" {
		return h.showUsage()
	}

	// Parse subcommand
	parts := strings.Fields(args)
	subcommand := strings.ToLower(parts[0])

	switch subcommand {
	case "start":
		return h.handleStart(parts[1:])
	case "stop":
		return h.handleStop()
	case "status":
		return h.handleStatus()
	case "metrics":
		return h.handleMetrics()
	case "log":
		return h.handleLog(parts[1:])
	case "pending":
		return h.handlePending()
	case "confirm":
		if len(parts) < 2 {
			return "Usage: .mcp confirm <request_id>"
		}
		return h.handleConfirm(parts[1])
	case "deny":
		if len(parts) < 2 {
			return "Usage: .mcp deny <request_id> [reason]"
		}
		reason := ""
		if len(parts) > 2 {
			reason = strings.Join(parts[2:], " ")
		}
		return h.handleDeny(parts[1], reason)
	case "permissions-config":
		return h.handlePermissionsConfig(parts[1:])
	case "generate-api-key":
		return h.handleGenerateAPIKey()
	default:
		return fmt.Sprintf("Unknown MCP command: %s\n%s", subcommand, h.showUsage())
	}
}

// handleStart starts the MCP server
func (h *MCPHandler) handleStart(args []string) string {
	// Check if already running
	if h.mcpServer != nil && h.mcpServer.IsRunning() {
		return "MCP server is already running. Use '.mcp status' to check status or '.mcp stop' to stop it first."
	}

	// Create default config
	config := ai.DefaultMCPConfig()

	// Check for --config-file first (loads JSON config)
	var configFile string
	for i := 0; i < len(args); i++ {
		if args[i] == "--config-file" && i+1 < len(args) {
			configFile = args[i+1]
			break
		}
	}

	// Load JSON config if specified
	if configFile != "" {
		loadedConfig, err := ai.LoadMCPConfigFromFile(configFile)
		if err != nil {
			return fmt.Sprintf("Failed to load config file: %v", err)
		}
		config = loadedConfig
	}

	// Parse options (these override JSON config)
	hasPresetMode := false
	hasSkipConf := false
	var confirmQueriesArg []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--socket-path":
			if i+1 < len(args) {
				config.SocketPath = args[i+1]
				i++
			}

		// HTTP transport configuration
		case "--http-host":
			if i+1 < len(args) {
				config.HttpHost = args[i+1]
				i++
			}
		case "--http-port":
			if i+1 < len(args) {
				var port int
				if _, err := fmt.Sscanf(args[i+1], "%d", &port); err != nil {
					return fmt.Sprintf("Error: --http-port must be a number, got %q", args[i+1])
				}
				if port <= 0 || port > 65535 {
					return fmt.Sprintf("Error: --http-port must be between 1-65535, got %d", port)
				}
				config.HttpPort = port
				i++
			}
		case "--api-key":
			if i+1 < len(args) {
				apiKey := args[i+1]
				if err := ai.ValidateAPIKeyFormat(apiKey, config.ApiKeyMaxAge); err != nil {
					return fmt.Sprintf("Error: Invalid --api-key: %v", err)
				}
				config.ApiKey = apiKey
				i++
			}
		case "--api-key-max-age-days":
			if i+1 < len(args) {
				var days float64
				if _, err := fmt.Sscanf(args[i+1], "%f", &days); err != nil {
					return fmt.Sprintf("Error: --api-key-max-age-days must be a number, got %q", args[i+1])
				}
				if days <= 0 {
					config.ApiKeyMaxAge = 0 // Disabled
				} else {
					config.ApiKeyMaxAge = time.Duration(days*24) * time.Hour
				}
				i++
			}
		case "--disable-api-key-age-check":
			config.ApiKeyMaxAge = 0
		case "--allowed-origins":
			if i+1 < len(args) {
				// Parse comma-separated list of origins
				origins := strings.Split(args[i+1], ",")
				config.AllowedOrigins = make([]string, 0, len(origins))
				for _, origin := range origins {
					trimmed := strings.TrimSpace(origin)
					if trimmed != "" {
						config.AllowedOrigins = append(config.AllowedOrigins, trimmed)
					}
				}
				i++
			}
		case "--ip-allowlist":
			if i+1 < len(args) {
				// Parse comma-separated list of IPs/CIDRs
				ips := strings.Split(args[i+1], ",")
				config.IpAllowlist = make([]string, 0, len(ips))
				for _, ip := range ips {
					trimmed := strings.TrimSpace(ip)
					if trimmed != "" {
						config.IpAllowlist = append(config.IpAllowlist, trimmed)
					}
				}
				i++
			}
		case "--ip-allowlist-disabled":
			config.IpAllowlistDisabled = true
		case "--audit-http-headers":
			if i+1 < len(args) {
				// Parse comma-separated list of headers
				headers := strings.Split(args[i+1], ",")
				config.AuditHttpHeaders = make([]string, 0, len(headers))
				for _, header := range headers {
					trimmed := strings.TrimSpace(header)
					if trimmed != "" {
						config.AuditHttpHeaders = append(config.AuditHttpHeaders, trimmed)
					}
				}
				i++
			}
		case "--require-headers":
			if i+1 < len(args) {
				// Parse comma-separated list of header:value pairs
				pairs := strings.Split(args[i+1], ",")
				config.RequiredHeaders = make(map[string]string)
				for _, pair := range pairs {
					parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
					if len(parts) == 2 {
						headerName := strings.TrimSpace(parts[0])
						headerValue := strings.TrimSpace(parts[1])
						config.RequiredHeaders[headerName] = headerValue
					}
				}
				i++
			}

		case "--log-level":
			if i+1 < len(args) {
				config.LogLevel = args[i+1]
				i++
			}
		case "--log-file":
			if i+1 < len(args) {
				config.LogFile = args[i+1]
				i++
			}

		// Preset modes
		case "--readonly_mode":
			if hasSkipConf {
				return "Error: Cannot use --readonly_mode with --skip-confirmation"
			}
			config.Mode = ai.ConfigModePreset
			config.PresetMode = "readonly"
			hasPresetMode = true

		case "--readwrite_mode":
			if hasSkipConf {
				return "Error: Cannot use --readwrite_mode with --skip-confirmation"
			}
			config.Mode = ai.ConfigModePreset
			config.PresetMode = "readwrite"
			hasPresetMode = true

		case "--dba_mode":
			if hasSkipConf {
				return "Error: Cannot use --dba_mode with --skip-confirmation"
			}
			config.Mode = ai.ConfigModePreset
			config.PresetMode = "dba"
			hasPresetMode = true

		// Confirmation overlay (only with preset modes)
		case "--confirm-queries":
			if i+1 < len(args) {
				confirmQueriesArg = ai.ParseCategoryList(args[i+1])
				i++
			}

		// Fine-grained mode
		case "--skip-confirmation":
			if hasPresetMode {
				return "Error: Cannot use --skip-confirmation with preset modes (--readonly_mode, --readwrite_mode, --dba_mode)"
			}
			if i+1 < len(args) {
				categories := ai.ParseCategoryList(args[i+1])
				if err := config.UpdateSkipConfirmation(categories); err != nil {
					return fmt.Sprintf("Error in --skip-confirmation: %v", err)
				}
				hasSkipConf = true
				i++
			}

		// Runtime permission config control
		case "--disable-runtime-permission-changes":
			config.DisableRuntimePermissionChanges = true

		case "--allow-runtime-permission-changes":
			config.DisableRuntimePermissionChanges = false

		// MCP request approval (security: explicit opt-in)
		case "--allow-mcp-request-approval":
			config.AllowMCPRequestApproval = true
		}
	}

	// Validate confirm-queries only with preset modes
	if len(confirmQueriesArg) > 0 && !hasPresetMode {
		return "Error: --confirm-queries only allowed with preset modes (--readonly_mode, --readwrite_mode, --dba_mode)"
	}

	// Apply confirm-queries overlay if specified
	if len(confirmQueriesArg) > 0 {
		if err := config.UpdateConfirmQueries(confirmQueriesArg); err != nil {
			return fmt.Sprintf("Error in --confirm-queries: %v", err)
		}
	}

	// Create MCP server
	mcpServer, err := ai.NewMCPServer(h.replSession, config)
	if err != nil {
		return fmt.Sprintf("Failed to create MCP server: %v", err)
	}

	// Start the server
	if err := mcpServer.Start(); err != nil {
		return fmt.Sprintf("Failed to start MCP server: %v", err)
	}

	h.mcpServer = mcpServer

	// Build success message
	var sb strings.Builder
	sb.WriteString("✅ MCP server started successfully\n\n")

	// Show configuration
	sb.WriteString(config.FormatConfigForDisplay())
	sb.WriteString("\n")

	// Server details
	sb.WriteString("Server Details:\n")
	sb.WriteString(fmt.Sprintf("  HTTP endpoint: http://%s:%d/mcp\n", config.HttpHost, config.HttpPort))

	// API key with timestamp
	apiKeyInfo := formatAPIKeyInfo(config.ApiKey, config.ApiKeyMaxAge)
	sb.WriteString(fmt.Sprintf("  API key: %s\n", apiKeyInfo))

	sb.WriteString(fmt.Sprintf("  Log level: %s\n", config.LogLevel))
	sb.WriteString(fmt.Sprintf("  Log file: %s\n", config.LogFile))
	sb.WriteString("  Available tools: 9 (FUZZY_SEARCH, GET_SCHEMA, LIST_KEYSPACES, LIST_TABLES, etc.)\n\n")

	sb.WriteString("Claude Code can now connect via .mcp.json:\n")
	sb.WriteString("  {\n")
	sb.WriteString("    \"mcpServers\": {\n")
	sb.WriteString("      \"cqlai\": {\n")
	sb.WriteString(fmt.Sprintf("        \"url\": \"http://%s:%d/mcp\",\n", config.HttpHost, config.HttpPort))
	sb.WriteString("        \"headers\": {\n")
	sb.WriteString(fmt.Sprintf("          \"X-API-Key\": \"%s\"\n", config.ApiKey))
	sb.WriteString("        }\n")
	sb.WriteString("      }\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n\n")

	sb.WriteString("Use '.mcp status' to view detailed status\n")
	sb.WriteString("Use '.mcp permissions-config mode <readonly|readwrite|dba>' to change mode dynamically\n")

	return sb.String()
}

// handleStop stops the MCP server
func (h *MCPHandler) handleStop() string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running. Use '.mcp start' to start it."
	}

	// Get metrics before stopping
	metrics := h.mcpServer.GetMetrics()

	// Stop the server
	if err := h.mcpServer.Stop(); err != nil {
		return fmt.Sprintf("Failed to stop MCP server: %v", err)
	}

	// Build stop message with summary
	var sb strings.Builder
	sb.WriteString("MCP server stopped\n\n")
	sb.WriteString("Session Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total requests: %d\n", metrics.TotalRequests))
	sb.WriteString(fmt.Sprintf("  Successful: %d (%.1f%%)\n", metrics.SuccessfulRequests, metrics.SuccessRate))
	sb.WriteString(fmt.Sprintf("  Failed: %d\n", metrics.FailedRequests))

	if len(metrics.ToolCalls) > 0 {
		sb.WriteString("\n  Tool Breakdown:\n")
		for tool, count := range metrics.ToolCalls {
			sb.WriteString(fmt.Sprintf("    - %s: %d calls\n", tool, count))
		}
	}

	h.mcpServer = nil

	return sb.String()
}

// handleStatus shows MCP server status
func (h *MCPHandler) handleStatus() string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running.\nUse '.mcp start' to start it."
	}

	metrics := h.mcpServer.GetMetrics()
	config := h.mcpServer.GetConfig()
	connInfo := h.mcpServer.GetConnectionInfo()

	var sb strings.Builder
	sb.WriteString("MCP Server Status:\n\n")
	sb.WriteString("  State: RUNNING\n\n")

	// Confirmation configuration (detailed format #3)
	sb.WriteString("Confirmation Configuration:\n")
	sb.WriteString(config.FormatConfigForDisplay())
	sb.WriteString("\n")

	// Cassandra connection details
	sb.WriteString("Cassandra Connection:\n")
	sb.WriteString("  Status: Connected (independent session)\n")
	if connInfo.ClusterName != "" {
		sb.WriteString(fmt.Sprintf("  Cluster: %s\n", connInfo.ClusterName))
	}
	sb.WriteString(fmt.Sprintf("  Contact point: %s\n", connInfo.ContactPoint))
	sb.WriteString(fmt.Sprintf("  Username: %s\n\n", connInfo.Username))

	// Server configuration
	sb.WriteString("Server Configuration:\n")
	sb.WriteString(fmt.Sprintf("  HTTP endpoint: http://%s:%d/mcp\n", config.HttpHost, config.HttpPort))

	// API key with timestamp
	apiKeyInfo := formatAPIKeyInfo(config.ApiKey, config.ApiKeyMaxAge)
	sb.WriteString(fmt.Sprintf("  API key: %s\n", apiKeyInfo))

	// Allowed origins
	if len(config.AllowedOrigins) > 0 {
		sb.WriteString(fmt.Sprintf("  Allowed origins: %v\n", config.AllowedOrigins))
	} else if config.HttpHost == "127.0.0.1" || config.HttpHost == "localhost" {
		sb.WriteString("  Allowed origins: localhost only (secure default)\n")
	} else {
		sb.WriteString("  Allowed origins: NONE (will reject browser requests)\n")
	}

	sb.WriteString(fmt.Sprintf("  Log level: %s\n", config.LogLevel))
	sb.WriteString(fmt.Sprintf("  Log file: %s\n", config.LogFile))
	sb.WriteString(fmt.Sprintf("  History file: %s\n", config.HistoryFile))
	sb.WriteString(fmt.Sprintf("  MCP request approval: %v (security: must opt-in to enable)\n\n", config.AllowMCPRequestApproval))

	// Metrics
	sb.WriteString("Metrics:\n")
	sb.WriteString(fmt.Sprintf("  Total requests: %d\n", metrics.TotalRequests))
	sb.WriteString(fmt.Sprintf("  Success rate: %.1f%%\n", metrics.SuccessRate))

	if len(metrics.ToolCalls) > 0 {
		sb.WriteString("\n  Tool Usage:\n")
		for tool, count := range metrics.ToolCalls {
			sb.WriteString(fmt.Sprintf("    - %s: %d\n", tool, count))
		}
	}

	return sb.String()
}

// handleMetrics shows detailed metrics
func (h *MCPHandler) handleMetrics() string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	metrics := h.mcpServer.GetMetrics()

	var sb strings.Builder
	sb.WriteString("MCP Server Metrics:\n\n")
	sb.WriteString(fmt.Sprintf("Total Requests: %d\n", metrics.TotalRequests))
	sb.WriteString(fmt.Sprintf("  Successful: %d (%.1f%%)\n", metrics.SuccessfulRequests, metrics.SuccessRate))
	sb.WriteString(fmt.Sprintf("  Failed: %d\n\n", metrics.FailedRequests))

	if len(metrics.ToolCalls) > 0 {
		sb.WriteString("Tool Breakdown:\n")
		for tool, count := range metrics.ToolCalls {
			sb.WriteString(fmt.Sprintf("  %s: %d calls\n", tool, count))
		}
	} else {
		sb.WriteString("No tool calls recorded yet.\n")
	}

	return sb.String()
}

// handleLog shows MCP logs
func (h *MCPHandler) handleLog(args []string) string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	// TODO: Implement log querying
	return "Log querying not yet implemented.\nLog file: /tmp/cqlai-mcp.log"
}

// handlePending shows pending confirmation requests
func (h *MCPHandler) handlePending() string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	pending := h.mcpServer.GetPendingConfirmations()

	if len(pending) == 0 {
		return "No pending confirmation requests."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pending Confirmation Requests (%d):\n\n", len(pending)))

	for _, req := range pending {
		sb.WriteString(fmt.Sprintf("⚠️  Request ID: %s\n", req.ID))
		sb.WriteString(fmt.Sprintf("   Severity: %s\n", req.Classification.Severity))
		sb.WriteString(fmt.Sprintf("   Operation: %s\n", req.Classification.Operation))
		sb.WriteString(fmt.Sprintf("   Query: %s\n", req.Query))
		sb.WriteString(fmt.Sprintf("   Tool: %s\n", req.Tool))
		sb.WriteString(fmt.Sprintf("   Requested: %s ago\n", formatDuration(req.Timestamp)))
		sb.WriteString(fmt.Sprintf("   Expires: %s\n\n", formatDuration(req.Timeout)))
		sb.WriteString(fmt.Sprintf("   To approve: .mcp confirm %s\n", req.ID))
		sb.WriteString(fmt.Sprintf("   To deny:    .mcp deny %s [reason]\n\n", req.ID))
	}

	return sb.String()
}

// handleConfirm confirms a pending dangerous query request
func (h *MCPHandler) handleConfirm(requestID string) string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	// Get request details first
	req, err := h.mcpServer.GetConfirmationRequest(requestID)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Confirm the request (use session username or "mcp_client" as identifier)
	confirmedBy := "mcp_client"
	if h.replSession != nil {
		if username := h.replSession.Username(); username != "" {
			confirmedBy = username
		}
	}
	err = h.mcpServer.ConfirmRequest(requestID, confirmedBy)
	if err != nil {
		return fmt.Sprintf("Failed to confirm request: %v", err)
	}

	// Execute the confirmed query
	err = h.mcpServer.ExecuteConfirmedQuery(requestID)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Confirmed request %s\n\n", requestID))
	sb.WriteString(fmt.Sprintf("Query: %s\n", req.Query))
	sb.WriteString(fmt.Sprintf("Operation: %s (%s)\n", req.Classification.Operation, req.Classification.Severity))

	if err != nil {
		sb.WriteString(fmt.Sprintf("\n❌ Execution failed: %v\n", err))
	} else {
		sb.WriteString("\n✅ Query executed successfully\n")

		// Get updated request with execution metadata
		updatedReq, _ := h.mcpServer.GetConfirmationRequest(requestID)
		if updatedReq != nil && updatedReq.Executed {
			sb.WriteString(fmt.Sprintf("Execution time: %v\n", updatedReq.ExecutionTime))
			if updatedReq.RowsAffected > 0 {
				sb.WriteString(fmt.Sprintf("Rows affected: %d\n", updatedReq.RowsAffected))
			}
			if updatedReq.TraceID != nil {
				sb.WriteString(fmt.Sprintf("Trace ID: %x\n", updatedReq.TraceID))
			}
		}
	}

	return sb.String()
}

// handleDeny denies a pending dangerous query request
func (h *MCPHandler) handleDeny(requestID, reason string) string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	// Get request details first
	req, err := h.mcpServer.GetConfirmationRequest(requestID)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Default reason if not provided
	if reason == "" {
		reason = "User denied"
	}

	// Deny the request (use session username or "mcp_client" as identifier)
	deniedBy := "mcp_client"
	if h.replSession != nil {
		if username := h.replSession.Username(); username != "" {
			deniedBy = username
		}
	}
	err = h.mcpServer.DenyRequest(requestID, deniedBy, reason)
	if err != nil {
		return fmt.Sprintf("Failed to deny request: %v", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("❌ Denied request %s\n\n", requestID))
	sb.WriteString(fmt.Sprintf("Query: %s\n", req.Query))
	sb.WriteString(fmt.Sprintf("Operation: %s (%s)\n", req.Classification.Operation, req.Classification.Severity))
	sb.WriteString(fmt.Sprintf("Reason: %s\n", reason))
	sb.WriteString("\nThe operation has been rejected.")

	return sb.String()
}

// handlePermissionsConfig handles permission configuration changes at runtime
func (h *MCPHandler) handlePermissionsConfig(args []string) string {
	if h.mcpServer == nil || !h.mcpServer.IsRunning() {
		return "MCP server is not running."
	}

	// No args or "show" = display current config
	if len(args) == 0 || (len(args) == 1 && strings.ToLower(args[0]) == "show") {
		config := h.mcpServer.GetConfig()
		return "Current Permission Configuration:\n\n" + config.FormatConfigForDisplay()
	}

	// Parse permissions-config subcommand
	if len(args) < 2 {
		return "Usage: .mcp permissions-config <setting> <value>\nSettings: mode, confirm-queries, skip-confirmation\nExample: .mcp permissions-config mode readwrite"
	}

	setting := strings.ToLower(args[0])
	value := strings.Join(args[1:], " ")

	switch setting {
	case "mode":
		return h.handleConfigMode(args[1])

	case "confirm-queries":
		categories := ai.ParseCategoryList(value)
		return h.handleConfigConfirmQueries(categories)

	case "skip-confirmation":
		categories := ai.ParseCategoryList(value)
		return h.handleConfigSkipConfirmation(categories)

	default:
		return fmt.Sprintf("Unknown config setting: %s\nValid settings: mode, confirm-queries, skip-confirmation", setting)
	}
}

// handleGenerateAPIKey generates a new KSUID API key
func (h *MCPHandler) handleGenerateAPIKey() string {
	// Generate new KSUID
	key, err := ai.GenerateAPIKey()
	if err != nil {
		return fmt.Sprintf("Failed to generate API key: %v", err)
	}

	// Extract timestamp for display
	id, _ := ai.ParseKSUID(key)
	keyTime := id.Time()

	var sb strings.Builder
	sb.WriteString("✅ New API Key Generated\n\n")
	sb.WriteString(fmt.Sprintf("API Key: %s\n\n", key))
	sb.WriteString("Key Details:\n")
	sb.WriteString(fmt.Sprintf("  Format: KSUID (K-Sortable Unique ID)\n"))
	sb.WriteString(fmt.Sprintf("  Length: 27 characters (base62 encoding)\n"))
	sb.WriteString(fmt.Sprintf("  Generated: %s\n", keyTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("  Entropy: 128 bits of cryptographically secure random data\n\n"))

	sb.WriteString("⚠️  IMPORTANT:\n")
	sb.WriteString("1. Save this key securely (it won't be shown again)\n")
	sb.WriteString("2. Update your .mcp.json config file:\n")
	sb.WriteString(fmt.Sprintf("   \"api_key\": \"%s\"\n\n", key))
	sb.WriteString("3. Restart MCP server with new key:\n")
	sb.WriteString(fmt.Sprintf("   .mcp stop\n"))
	sb.WriteString(fmt.Sprintf("   .mcp start --api-key=%s\n\n", key))

	sb.WriteString("Or use environment variable:\n")
	sb.WriteString(fmt.Sprintf("   export MCP_API_KEY=\"%s\"\n", key))
	sb.WriteString("   Then in config: \"api_key\": \"${MCP_API_KEY}\"\n")

	return sb.String()
}

// handleConfigMode changes the preset mode
func (h *MCPHandler) handleConfigMode(mode string) string {
	mode = strings.ToLower(mode)

	// Validate mode
	if err := ai.ValidatePresetMode(mode); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Update mode
	if err := h.mcpServer.UpdateMode(mode); err != nil {
		return fmt.Sprintf("Failed to update mode: %v", err)
	}

	// Get new config for display
	config := h.mcpServer.GetConfig()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Mode changed to: %s\n\n", mode))
	sb.WriteString(config.FormatConfigForDisplay())

	return sb.String()
}

// handleConfigConfirmQueries changes the confirm-queries overlay
func (h *MCPHandler) handleConfigConfirmQueries(categories []string) string {
	// Update configuration
	if err := h.mcpServer.UpdateConfirmQueries(categories); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Get new config for display
	config := h.mcpServer.GetConfig()

	var sb strings.Builder
	if len(categories) == 0 || (len(categories) == 1 && (categories[0] == "none" || categories[0] == "disable")) {
		sb.WriteString("✅ Confirmations disabled\n\n")
	} else if len(categories) == 1 && categories[0] == "ALL" {
		sb.WriteString("✅ Confirmations required for ALL operations\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ Confirmations required for: %s\n\n", strings.Join(categories, ", ")))
	}

	sb.WriteString(config.FormatConfigForDisplay())

	return sb.String()
}

// handleConfigSkipConfirmation changes the skip-confirmation list (switches to fine-grained mode)
func (h *MCPHandler) handleConfigSkipConfirmation(categories []string) string {
	// Update configuration (switches to fine-grained mode)
	if err := h.mcpServer.UpdateSkipConfirmation(categories); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Get new config for display
	config := h.mcpServer.GetConfig()

	var sb strings.Builder
	if len(categories) == 1 && strings.ToUpper(categories[0]) == "ALL" {
		sb.WriteString("✅ Switched to fine-grained mode: Skip confirmation on ALL operations\n\n")
	} else if len(categories) == 0 || (len(categories) == 1 && categories[0] == "none") {
		sb.WriteString("✅ Switched to fine-grained mode: Confirm ALL operations\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ Switched to fine-grained mode: Skip confirmation on %s\n\n", strings.Join(categories, ", ")))
	}

	sb.WriteString(config.FormatConfigForDisplay())

	return sb.String()
}

// formatDuration formats a time or duration for display
// formatAPIKeyInfo formats API key information with masking and timestamp
func formatAPIKeyInfo(apiKey string, maxAge time.Duration) string {
	if apiKey == "" {
		return "not set (will be auto-generated)"
	}

	// Mask the key
	masked := ai.MaskAPIKey(apiKey)

	// Extract timestamp from KSUID
	id, err := ai.ParseKSUID(apiKey)
	if err != nil {
		return fmt.Sprintf("%s (invalid: %v)", masked, err)
	}

	keyTime := id.Time()
	keyAge := time.Since(keyTime)

	// Format the output
	var info strings.Builder
	info.WriteString(fmt.Sprintf("%s (generated: %s, age: %v",
		masked,
		keyTime.Format("2006-01-02 15:04:05"),
		keyAge.Round(time.Hour)))

	// Show expiration status
	if maxAge > 0 {
		remaining := maxAge - keyAge
		if remaining < 0 {
			info.WriteString(" ⚠️  EXPIRED")
		} else {
			daysRemaining := int(remaining.Hours() / 24)
			info.WriteString(fmt.Sprintf(", expires in %d days", daysRemaining))
		}
	} else {
		info.WriteString(", never expires ⚠️")
	}

	info.WriteString(")")
	return info.String()
}

func formatDuration(t interface{}) string {
	var d time.Duration

	switch v := t.(type) {
	case time.Time:
		d = time.Until(v)
		if d < 0 {
			d = time.Since(v)
			return fmt.Sprintf("%v ago", d.Round(time.Second))
		}
		return fmt.Sprintf("in %v", d.Round(time.Second))
	case time.Duration:
		return fmt.Sprintf("%v", v.Round(time.Second))
	default:
		return "unknown"
	}
}

// showUsage shows MCP command usage
func (h *MCPHandler) showUsage() string {
	return `MCP (Model Context Protocol) Server Commands:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SERVER CONTROL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.mcp start [options]           Start MCP server
.mcp stop                      Stop MCP server
.mcp status                    Show server status and configuration
.mcp metrics                   Show detailed request metrics
.mcp generate-api-key          Generate a new KSUID API key for rotation
.mcp log [options]             Show MCP logs (not yet implemented)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
START OPTIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Server Configuration:
  --config-file <path>         Load config from JSON file

HTTP Transport:
  --http-host <host>           HTTP server host (default: 127.0.0.1)
  --http-port <port>           HTTP server port (default: 8888)
  --api-key <ksuid>            KSUID API key (default: auto-generated)
  --api-key-max-age-days <n>   Max API key age in days (default: 30, 0=disabled)
  --disable-api-key-age-check  Disable age validation (SECURITY RISK)
  --allowed-origins <list>     Comma-separated allowed origins (for non-localhost)

Security (Defense-in-Depth):
  --ip-allowlist <list>        IP/CIDR allowlist (default: 127.0.0.1)
                               Examples: "203.0.113.10,10.0.1.0/24"
  --ip-allowlist-disabled      Disable IP checking (SECURITY RISK)
  --audit-http-headers <list>  Headers to log (default: X-Forwarded-For,User-Agent)
                               Use "ALL" to log all headers
  --require-headers <list>     Required header:value pairs
                               Examples: "X-Proxy-Verified:true,X-Request-ID:^req_.*"

Logging:
  --log-level <level>          Log level: debug, info, warning, error (default: info)
  --log-file <path>            Log file path (default: ~/.cqlai/cqlai_mcp.log)

Legacy (Deprecated):
  --socket-path <path>         Unix socket path (deprecated, use HTTP)

Security Modes (choose ONE - mutually exclusive):

  PRESET MODES (Recommended):
    --readonly_mode            Queries and session settings only (DEFAULT - safest)
    --readwrite_mode           + Data modifications (INSERT/UPDATE/DELETE)
    --dba_mode                 All operations allowed

  FINE-GRAINED MODE (Advanced):
    --skip-confirmation <list> Comma-separated categories to skip confirmation
                               Categories: dql, dml, ddl, dcl, file, ALL, none
                               SESSION always skipped automatically

  Confirmation Overlay (only with preset modes):
    --confirm-queries <list>   Require confirmation even for allowed operations
                               Values: dql, dml, ddl, dcl, file, ALL, none, disable

  Runtime Permission Control:
    --disable-runtime-permission-changes  Lock configuration (prevent update_mcp_permissions tool)
    --allow-runtime-permission-changes    Allow runtime changes (DEFAULT)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RUNTIME CONFIGURATION (Change settings without restart)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.mcp permissions-config                    Show current configuration
.mcp permissions-config mode <mode>        Change preset mode (readonly|readwrite|dba)
.mcp permissions-config confirm-queries <list>    Change confirmation overlay
.mcp permissions-config skip-confirmation <list>  Switch to fine-grained mode

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CONFIRMATION MANAGEMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.mcp pending                   Show pending confirmation requests
.mcp confirm <req_id>          Approve a pending request
.mcp deny <req_id> [reason]    Reject a pending request

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EXAMPLES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Start in readonly (default):
  .mcp start

Start in readwrite with confirmations on data mods:
  .mcp start --readwrite_mode --confirm-queries dml

Start in DBA mode with security confirmations:
  .mcp start --dba_mode --confirm-queries dcl

Fine-grained control (skip on queries and data only):
  .mcp start --skip-confirmation dql,dml

Change mode at runtime:
  .mcp permissions-config mode dba
  .mcp permissions-config confirm-queries disable
  .mcp permissions-config skip-confirmation ALL

Operation Categories:
  dql     - Data queries (SELECT, LIST, DESCRIBE) - 14 operations
  session - Session settings (CONSISTENCY, PAGING, etc) - 8 operations
  dml     - Data manipulation (INSERT, UPDATE, DELETE) - 8 operations
  ddl     - Schema definition (CREATE, ALTER, DROP) - 28 operations
  dcl     - Access control (roles, users, permissions) - 13 operations
  file    - File operations (COPY, SOURCE) - 3 operations
`
}
