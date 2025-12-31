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

		status := map[string]any{
			"state":  "RUNNING",
			"config": map[string]any{
				"mode":                            string(config.Mode),
				"preset_mode":                     config.PresetMode,
				"confirm_queries":                 config.ConfirmQueries,
				"skip_confirmation":               config.SkipConfirmation,
				"disable_runtime_permission_changes": config.DisableRuntimePermissionChanges,
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

		err := s.CancelRequest(requestID, "claude", reason)
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
