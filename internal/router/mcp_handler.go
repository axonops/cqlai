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

	// Parse options
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
	sb.WriteString(fmt.Sprintf("  Socket: %s\n", config.SocketPath))
	sb.WriteString(fmt.Sprintf("  Log level: %s\n", config.LogLevel))
	sb.WriteString(fmt.Sprintf("  Log file: %s\n", config.LogFile))
	sb.WriteString("  Available tools: 9 (FUZZY_SEARCH, GET_SCHEMA, LIST_KEYSPACES, LIST_TABLES, etc.)\n\n")

	sb.WriteString("Claude Code can now connect via:\n")
	sb.WriteString(fmt.Sprintf("  claude mcp add --transport stdio cqlai --scope project -- nc -U %s\n\n", config.SocketPath))

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
	sb.WriteString(fmt.Sprintf("  Socket: %s\n", config.SocketPath))
	sb.WriteString(fmt.Sprintf("  Log level: %s\n", config.LogLevel))
	sb.WriteString(fmt.Sprintf("  Log file: %s\n\n", config.LogFile))

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

	// Confirm the request (username would come from session, using "user" for now)
	err = h.mcpServer.ConfirmRequest(requestID, "user")
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

	// Deny the request (username would come from session, using "user" for now)
	err = h.mcpServer.DenyRequest(requestID, "user", reason)
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
.mcp log [options]             Show MCP logs (not yet implemented)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
START OPTIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Server Configuration:
  --socket-path <path>         Socket path (default: /tmp/cqlai-mcp.sock)
  --log-level <level>          Log level: debug, info, warning, error (default: info)
  --log-file <path>            Log file path (default: /tmp/cqlai-mcp.log)

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
