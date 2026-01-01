package ai

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerConfirmationTools registers MCP tools for confirmation lifecycle management
func (s *MCPServer) registerConfirmationTools() error {
	tools := []struct {
		tool    mcp.Tool
		handler server.ToolHandlerFunc
	}{
		{s.createGetMCPStatusTool(), s.createGetMCPStatusHandler()},
		{s.createGetPendingConfirmationsTool(), s.createGetPendingConfirmationsHandler()},
		{s.createGetApprovedConfirmationsTool(), s.createGetApprovedConfirmationsHandler()},
		{s.createGetDeniedConfirmationsTool(), s.createGetDeniedConfirmationsHandler()},
		{s.createGetCancelledConfirmationsTool(), s.createGetCancelledConfirmationsHandler()},
		{s.createGetConfirmationStateTool(), s.createGetConfirmationStateHandler()},
		{s.createConfirmRequestTool(), s.createConfirmRequestHandler()},
		{s.createDenyRequestTool(), s.createDenyRequestHandler()},
		{s.createCancelConfirmationTool(), s.createCancelConfirmationHandler()},
		{s.createGetTraceDataTool(), s.createGetTraceDataHandler()},
	}

	for _, t := range tools {
		s.mcpServer.AddTool(t.tool, t.handler)
	}

	logger.DebugfToFile("MCP", "Registered %d confirmation lifecycle tools", len(tools))
	return nil
}

// get_mcp_status tool
func (s *MCPServer) createGetMCPStatusTool() mcp.Tool {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_mcp_status",
		"Get current MCP server status including permission configuration, connection details, and metrics. Use this to understand current security settings.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetMCPStatusHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		config := s.config.GetConfigSnapshot()
		connInfo := s.GetConnectionInfo()
		metrics := s.GetMetrics()

		// Extract API key timestamp info
		apiKeyMasked := MaskAPIKey(config.ApiKey)
		var apiKeyTimestamp string
		var apiKeyAge string
		var apiKeyExpired bool
		if id, err := ParseKSUID(config.ApiKey); err == nil {
			keyTime := id.Time()
			keyAgeD := time.Since(keyTime)
			apiKeyTimestamp = keyTime.Format(time.RFC3339)
			apiKeyAge = keyAgeD.Round(time.Hour).String()
			if config.ApiKeyMaxAge > 0 {
				apiKeyExpired = keyAgeD > config.ApiKeyMaxAge
			}
		}

		status := map[string]any{
			"state":  "RUNNING",
			"config": map[string]any{
				"http_endpoint":                          fmt.Sprintf("http://%s:%d/mcp", config.HttpHost, config.HttpPort),
				"http_host":                              config.HttpHost,
				"http_port":                              config.HttpPort,
				"api_key_masked":                         apiKeyMasked,
				"api_key_timestamp":                      apiKeyTimestamp,
				"api_key_age":                            apiKeyAge,
				"api_key_max_age_days":                   int(config.ApiKeyMaxAge.Hours() / 24),
				"api_key_expired":                        apiKeyExpired,
				"allowed_origins":                        config.AllowedOrigins,
				"mode":                                   string(config.Mode),
				"preset_mode":                            config.PresetMode,
				"confirm_queries":                        config.ConfirmQueries,
				"skip_confirmation":                      config.SkipConfirmation,
				"disable_runtime_permission_changes":     config.DisableRuntimePermissionChanges,
				"allow_mcp_request_approval":             config.AllowMCPRequestApproval,
				"history_file":                           config.HistoryFile,
				"history_max_size_mb":                    config.HistoryMaxSize / (1024 * 1024),
				"history_max_rotations":                  config.HistoryMaxRotations,
				"history_rotation_interval_seconds":      int(config.HistoryRotationInterval.Seconds()),
			},
			"connection": map[string]any{
				"contact_point": connInfo.ContactPoint,
				"username":      connInfo.Username,
				"cluster_name":  connInfo.ClusterName,
			},
			"metrics": map[string]any{
				"total_requests":      metrics.TotalRequests,
				"successful_requests": metrics.SuccessfulRequests,
				"failed_requests":     metrics.FailedRequests,
				"success_rate":        metrics.SuccessRate,
				"tool_calls":          metrics.ToolCalls,
			},
		}

		jsonData, _ := json.MarshalIndent(status, "", "  ")
		s.metrics.RecordToolCall("get_mcp_status", true, time.Since(startTime))
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// get_pending_confirmations tool
func (s *MCPServer) createGetPendingConfirmationsTool() mcp.Tool {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_pending_confirmations",
		"Get list of confirmation requests waiting for user approval. Returns request ID, query, severity, and timestamp.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetPendingConfirmationsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		pending := s.GetPendingConfirmations()
		result := formatConfirmationList(pending)
		s.metrics.RecordToolCall("get_pending_confirmations", true, time.Since(startTime))
		return mcp.NewToolResultText(result), nil
	}
}

// get_approved_confirmations tool
func (s *MCPServer) createGetApprovedConfirmationsTool() mcp.Tool {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_approved_confirmations",
		"Get list of confirmed/approved requests. Shows request history.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetApprovedConfirmationsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		approved := s.GetApprovedConfirmations()
		result := formatConfirmationList(approved)
		s.metrics.RecordToolCall("get_approved_confirmations", true, time.Since(startTime))
		return mcp.NewToolResultText(result), nil
	}
}

// get_denied_confirmations tool
func (s *MCPServer) createGetDeniedConfirmationsTool() mcp.Tool {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_denied_confirmations",
		"Get list of denied requests. Shows request history.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetDeniedConfirmationsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		denied := s.GetDeniedConfirmations()
		result := formatConfirmationList(denied)
		s.metrics.RecordToolCall("get_denied_confirmations", true, time.Since(startTime))
		return mcp.NewToolResultText(result), nil
	}
}

// get_cancelled_confirmations tool
func (s *MCPServer) createGetCancelledConfirmationsTool() mcp.Tool {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_cancelled_confirmations",
		"Get list of cancelled requests. Shows request history.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetCancelledConfirmationsHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		cancelled := s.GetCancelledConfirmations()
		result := formatConfirmationList(cancelled)
		s.metrics.RecordToolCall("get_cancelled_confirmations", true, time.Since(startTime))
		return mcp.NewToolResultText(result), nil
	}
}

// get_confirmation_state tool
func (s *MCPServer) createGetConfirmationStateTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"request_id": map[string]any{
				"type":        "string",
				"description": "The confirmation request ID (e.g., req_001)",
			},
		},
		"required": []string{"request_id"},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_confirmation_state",
		"Get detailed state of a specific confirmation request by ID. Returns status, query, timestamps, etc.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetConfirmationStateHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		argsMap := request.GetArguments()

		requestID, ok := argsMap["request_id"].(string)
		if !ok {
			s.metrics.RecordToolCall("get_confirmation_state", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid request_id parameter"), nil
		}

		req, err := s.GetConfirmationRequest(requestID)
		if err != nil {
			s.metrics.RecordToolCall("get_confirmation_state", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Request not found: %v", err)), nil
		}

		result := formatSingleConfirmation(req)
		s.metrics.RecordToolCall("get_confirmation_state", true, time.Since(startTime))
		return mcp.NewToolResultText(result), nil
	}
}

// confirm_request tool - Approves dangerous queries (requires user_confirmed AND allow_mcp_request_approval)
func (s *MCPServer) createConfirmRequestTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"request_id": map[string]any{
				"type":        "string",
				"description": "The confirmation request ID to approve (e.g., req_001)",
			},
			"user_confirmed": map[string]any{
				"type":        "boolean",
				"description": "REQUIRED: Must be true to indicate you have asked the user and they confirmed. Claude cannot approve dangerous operations without explicit user consent.",
			},
		},
		"required": []string{"request_id", "user_confirmed"},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"confirm_request",
		"Approve a dangerous query that requires confirmation. CRITICAL: You MUST set user_confirmed=true AND you MUST have explicitly asked the user for approval in your prompt. Never approve dangerous operations without user consent. Note: This tool is only available if allow_mcp_request_approval is enabled in config.",
		schemaJSON,
	)
}

func (s *MCPServer) createConfirmRequestHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		// Security check: Only allow if explicitly enabled in config
		if !s.config.AllowMCPRequestApproval {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError("confirm_request tool is disabled. Set allow_mcp_request_approval=true in MCP config to enable (security setting)."), nil
		}

		argsMap := request.GetArguments()

		requestID, ok := argsMap["request_id"].(string)
		if !ok {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid request_id parameter"), nil
		}

		userConfirmed, ok := argsMap["user_confirmed"].(bool)
		if !ok {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid user_confirmed parameter - must be boolean"), nil
		}

		if !userConfirmed {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Cannot confirm request: user_confirmed must be true. You must ask the user for explicit approval before confirming dangerous operations."), nil
		}

		err := s.ConfirmRequest(requestID, "mcp")
		if err != nil {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to confirm request: %v", err)), nil
		}

		// Execute the confirmed query
		if err := s.ExecuteConfirmedQuery(requestID); err != nil {
			s.metrics.RecordToolCall("confirm_request", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Request confirmed but execution failed: %v", err)), nil
		}

		// Get updated request to show execution details
		req, _ := s.GetConfirmationRequest(requestID)

		result := map[string]any{
			"status":      "confirmed_and_executed",
			"request_id":  requestID,
			"query":       req.Query,
			"executed":    req.Executed,
			"executed_at": req.ExecutedAt.Format("2006-01-02 15:04:05"),
		}

		if req.ExecutionTime > 0 {
			result["execution_time_ms"] = req.ExecutionTime.Milliseconds()
		}
		if req.TraceID != nil {
			result["trace_id"] = fmt.Sprintf("%x", req.TraceID)
		}
		if req.RowsAffected > 0 {
			result["rows_affected"] = req.RowsAffected
		}

		s.metrics.RecordToolCall("confirm_request", true, time.Since(startTime))

		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// deny_request tool - Denies dangerous queries (always allowed - no security risk in denying)
func (s *MCPServer) createDenyRequestTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"request_id": map[string]any{
				"type":        "string",
				"description": "The confirmation request ID to deny (e.g., req_001)",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Reason for denying the request (e.g., 'User declined', 'Too risky')",
			},
			"user_confirmed": map[string]any{
				"type":        "boolean",
				"description": "REQUIRED: Must be true to indicate you have asked the user and they declined. Claude cannot deny without user interaction.",
			},
		},
		"required": []string{"request_id", "user_confirmed"},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"deny_request",
		"Deny a dangerous query that requires confirmation. CRITICAL: You MUST set user_confirmed=true AND you MUST have explicitly asked the user, who declined. Provide a clear reason for the denial.",
		schemaJSON,
	)
}

func (s *MCPServer) createDenyRequestHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		argsMap := request.GetArguments()

		requestID, ok := argsMap["request_id"].(string)
		if !ok {
			s.metrics.RecordToolCall("deny_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid request_id parameter"), nil
		}

		userConfirmed, ok := argsMap["user_confirmed"].(bool)
		if !ok {
			s.metrics.RecordToolCall("deny_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid user_confirmed parameter - must be boolean"), nil
		}

		if !userConfirmed {
			s.metrics.RecordToolCall("deny_request", false, time.Since(startTime))
			return mcp.NewToolResultError("Cannot deny request: user_confirmed must be true. You must ask the user before denying on their behalf."), nil
		}

		reason, _ := argsMap["reason"].(string)
		if reason == "" {
			reason = "User declined"
		}

		err := s.DenyRequest(requestID, "mcp", reason)
		if err != nil {
			s.metrics.RecordToolCall("deny_request", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to deny request: %v", err)), nil
		}

		s.metrics.RecordToolCall("deny_request", true, time.Since(startTime))
		return mcp.NewToolResultText(fmt.Sprintf("Request %s denied. Reason: %s", requestID, reason)), nil
	}
}

// cancel_confirmation tool
func (s *MCPServer) createCancelConfirmationTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"request_id": map[string]any{
				"type":        "string",
				"description": "The confirmation request ID to cancel (e.g., req_001)",
			},
			"reason": map[string]any{
				"type":        "string",
				"description": "Optional reason for cancellation",
			},
		},
		"required": []string{"request_id"},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"cancel_confirmation",
		"Cancel a confirmation request. Can be called in any state (PENDING, CONFIRMED, DENIED, TIMEOUT).",
		schemaJSON,
	)
}

func (s *MCPServer) createCancelConfirmationHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		argsMap := request.GetArguments()

		requestID, ok := argsMap["request_id"].(string)
		if !ok {
			s.metrics.RecordToolCall("cancel_confirmation", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid request_id parameter"), nil
		}

		reason, _ := argsMap["reason"].(string)
		if reason == "" {
			reason = "Cancelled by Claude"
		}

		err := s.CancelRequest(requestID, "mcp", reason)
		if err != nil {
			s.metrics.RecordToolCall("cancel_confirmation", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to cancel request: %v", err)), nil
		}

		s.metrics.RecordToolCall("cancel_confirmation", true, time.Since(startTime))
		return mcp.NewToolResultText(fmt.Sprintf("Request %s cancelled successfully. Reason: %s", requestID, reason)), nil
	}
}

// formatConfirmationList formats a list of confirmation requests as JSON
func formatConfirmationList(requests []*ConfirmationRequest) string {
	if len(requests) == 0 {
		return "[]"
	}

	data := make([]map[string]any, len(requests))
	for i, req := range requests {
		data[i] = map[string]any{
			"request_id":   req.ID,
			"status":       req.Status,
			"query":        req.Query,
			"operation":    req.Classification.Operation,
			"severity":     req.Classification.Severity,
			"tool":         req.Tool,
			"timestamp":    req.Timestamp.Format(time.RFC3339),
			"timeout":      req.Timeout.Format(time.RFC3339),
			"confirmed_by": req.ConfirmedBy,
			"confirmed_at": formatTimePtr(req.ConfirmedAt),
		}
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

// formatSingleConfirmation formats a single confirmation request as JSON
func formatSingleConfirmation(req *ConfirmationRequest) string {
	data := map[string]any{
		"request_id":   req.ID,
		"status":       req.Status,
		"query":        req.Query,
		"classification": map[string]any{
			"is_dangerous": req.Classification.IsDangerous,
			"severity":     req.Classification.Severity,
			"operation":    req.Classification.Operation,
			"table":        req.Classification.Table,
			"description":  req.Classification.Description,
		},
		"tool":           req.Tool,
		"tool_operation": req.ToolOperation,
		"timestamp":      req.Timestamp.Format(time.RFC3339),
		"timeout":        req.Timeout.Format(time.RFC3339),
		"user_confirmed": req.UserConfirmed,
		"confirmed_by":   req.ConfirmedBy,
		"confirmed_at":   formatTimePtr(req.ConfirmedAt),
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}

// formatTimePtr formats a time.Time, handling zero values
func formatTimePtr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// get_trace_data tool - NEW (18th MCP tool)
func (s *MCPServer) createGetTraceDataTool() mcp.Tool {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"trace_id": map[string]any{
				"type":        "string",
				"description": "Cassandra trace ID (hex string) from query execution",
			},
		},
		"required": []string{"trace_id"},
	}
	schemaJSON, _ := json.Marshal(schema)
	return mcp.NewToolWithRawSchema(
		"get_trace_data",
		"Get detailed Cassandra trace data for a query by trace ID. Returns coordinator, duration, and trace events for performance analysis.",
		schemaJSON,
	)
}

func (s *MCPServer) createGetTraceDataHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()
		argsMap := request.GetArguments()

		traceIDHex, ok := argsMap["trace_id"].(string)
		if !ok || traceIDHex == "" {
			s.metrics.RecordToolCall("get_trace_data", false, time.Since(startTime))
			return mcp.NewToolResultError("Missing or invalid trace_id parameter"), nil
		}

		// Parse hex trace ID to bytes
		traceIDBytes, err := parseHexTraceID(traceIDHex)
		if err != nil {
			s.metrics.RecordToolCall("get_trace_data", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Invalid trace_id format: %v", err)), nil
		}

		// Query trace data from Cassandra
		traceData, err := getTraceDataByID(s.session, traceIDBytes)
		if err != nil {
			s.metrics.RecordToolCall("get_trace_data", false, time.Since(startTime))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve trace data: %v", err)), nil
		}

		s.metrics.RecordToolCall("get_trace_data", true, time.Since(startTime))

		result := traceData

		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// parseHexTraceID parses a hex string trace ID to bytes
func parseHexTraceID(hexStr string) ([]byte, error) {
	traceIDBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex trace ID: %v", err)
	}
	if len(traceIDBytes) != 16 {
		return nil, fmt.Errorf("trace ID must be 16 bytes (UUID), got %d bytes", len(traceIDBytes))
	}
	return traceIDBytes, nil
}

// getTraceDataByID retrieves trace data for a specific trace ID
func getTraceDataByID(session *db.Session, traceIDBytes []byte) (map[string]any, error) {
	// Convert bytes to UUID
	traceID, err := gocql.UUIDFromBytes(traceIDBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trace ID: %v", err)
	}

	// Query trace events
	query := `SELECT event_id, activity, source, source_elapsed, thread
	          FROM system_traces.events
	          WHERE session_id = ?
	          ORDER BY event_id`

	iter := session.Query(query, traceID).Consistency(gocql.LocalOne).Iter()
	defer iter.Close()

	var events []map[string]any
	var eventID gocql.UUID
	var activity, source, thread string
	var sourceElapsed int

	for iter.Scan(&eventID, &activity, &source, &sourceElapsed, &thread) {
		event := map[string]any{
			"event_id":       eventID.String(),
			"activity":       activity,
			"source":         source,
			"source_elapsed": sourceElapsed,
			"thread":         thread,
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to retrieve trace events: %v", err)
	}

	// Get session info
	var coordinator string
	var duration int
	sessionIter := session.Query(`SELECT coordinator, duration
	                                FROM system_traces.sessions
	                                WHERE session_id = ?`, traceID).Consistency(gocql.LocalOne).Iter()

	hasSession := sessionIter.Scan(&coordinator, &duration)
	_ = sessionIter.Close()

	result := map[string]any{
		"trace_id": traceID.String(),
		"events":   events,
	}

	if hasSession {
		result["coordinator"] = coordinator
		result["duration_us"] = duration
	}

	return result, nil
}
