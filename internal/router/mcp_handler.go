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

	// Parse options (simple parsing for now, can enhance later)
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
		case "--readonly":
			config.ConfirmationMode = "readonly"
		case "--read-write":
			config.ConfirmationMode = "read_write"
		case "--confirm-on-dangerous":
			config.ConfirmationMode = "dangerous_only"
		case "--confirm-all":
			config.ConfirmationMode = "all"
		case "--no-confirmation":
			config.ConfirmationMode = "none"
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
	sb.WriteString("MCP server started successfully\n\n")
	sb.WriteString(fmt.Sprintf("Socket: %s\n", config.SocketPath))
	sb.WriteString("Cassandra connection: ACTIVE (using independent session)\n")
	sb.WriteString(fmt.Sprintf("Confirmation mode: %s\n", config.ConfirmationMode))
	sb.WriteString(fmt.Sprintf("Log level: %s\n", config.LogLevel))
	sb.WriteString(fmt.Sprintf("Log file: %s\n", config.LogFile))
	sb.WriteString("Available tools: 9 (FUZZY_SEARCH, GET_SCHEMA, LIST_KEYSPACES, LIST_TABLES, etc.)\n\n")
	sb.WriteString("Claude Code can now connect via:\n")
	sb.WriteString(fmt.Sprintf("  claude mcp add --transport stdio cqlai --scope project -- nc -U %s\n\n", config.SocketPath))
	sb.WriteString("Or add to .mcp.json:\n")
	sb.WriteString(fmt.Sprintf(`  {
    "mcpServers": {
      "cqlai": {
        "type": "stdio",
        "command": "nc",
        "args": ["-U", "%s"]
      }
    }
  }
`, config.SocketPath))

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

	// Confirmation mode with description
	sb.WriteString("Confirmation Mode:\n")
	sb.WriteString(fmt.Sprintf("  Mode: %s\n", config.ConfirmationMode))
	sb.WriteString(fmt.Sprintf("  Description: %s\n\n", ai.GetConfirmationModeDescription(config.ConfirmationMode)))

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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ Confirmed request %s\n\n", requestID))
	sb.WriteString(fmt.Sprintf("Query: %s\n", req.Query))
	sb.WriteString(fmt.Sprintf("Operation: %s (%s)\n", req.Classification.Operation, req.Classification.Severity))
	sb.WriteString("\nThe operation will now proceed.")

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

.mcp start [options]           Start MCP server
  Options:
    --socket-path <path>       Socket path (default: /tmp/cqlai-mcp.sock)
    --log-level <level>        Log level: debug, info, warning, error (default: info)
    --log-file <path>          Log file path (default: /tmp/cqlai-mcp.log)

  Confirmation Modes:
    --readonly                 Only SELECT/DESCRIBE allowed (safest)
    --read-write               SELECT/INSERT/UPDATE/DELETE allowed; DROP/TRUNCATE require confirmation
    --confirm-on-dangerous     All queries allowed; dangerous ones require confirmation (default)
    --confirm-all              All queries require confirmation (most restrictive)
    --no-confirmation          No confirmations required (NOT RECOMMENDED - dangerous!)

.mcp stop                      Stop MCP server
.mcp status                    Show server status and metrics
.mcp metrics                   Show detailed metrics
.mcp pending                   Show pending confirmation requests
.mcp confirm <req_id>          Confirm a dangerous query request
.mcp deny <req_id> [reason]    Deny a dangerous query request
.mcp log [options]             Show MCP logs (not yet implemented)

Examples:
  .mcp start
  .mcp start --confirm-on-dangerous --log-level debug
  .mcp status
  .mcp pending
  .mcp confirm req_001
  .mcp deny req_002 "Query too broad"
  .mcp stop

After starting, configure Claude Code:
  claude mcp add --transport stdio cqlai --scope project -- nc -U /tmp/cqlai-mcp.sock

Or add to .mcp.json:
  {
    "mcpServers": {
      "cqlai": {
        "type": "stdio",
        "command": "nc",
        "args": ["-U", "/tmp/cqlai-mcp.sock"]
      }
    }
  }
`
}
