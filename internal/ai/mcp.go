package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	mcpServer  *server.MCPServer // The mark3labs MCP server
	listener   net.Listener
	socketPath string

	// Confirmation system for dangerous queries
	confirmationQueue *ConfirmationQueue
	config            *MCPServerConfig

	// Observability
	metrics *MetricsCollector
	mcpLog  *MCPLogger

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
	// Server infrastructure
	SocketPath          string
	ConfirmationTimeout time.Duration
	LogLevel            string
	LogFile             string

	// Confirmation system (thread-safe via mutex)
	Mode             ConfigMode // "preset" or "fine-grained"
	PresetMode       string     // "readonly", "readwrite", "dba" (when Mode=preset)
	ConfirmQueries   []string   // Categories requiring confirmation (when Mode=preset)
	SkipConfirmation []string   // Categories to skip confirmation (when Mode=fine-grained)

	// Runtime permission configuration control
	DisableRuntimePermissionChanges bool // If false, update_mcp_permissions tool is disabled

	// Thread safety for runtime config changes
	mu sync.RWMutex
}

// DefaultMCPConfig returns default MCP server configuration
func DefaultMCPConfig() *MCPServerConfig {
	return &MCPServerConfig{
		SocketPath:          "/tmp/cqlai-mcp.sock",
		ConfirmationTimeout: 5 * time.Minute,
		LogLevel:            "info",
		LogFile:             "/tmp/cqlai-mcp.log",

		// Default: readonly mode (safest)
		Mode:             ConfigModePreset,
		PresetMode:       "readonly",
		ConfirmQueries:   nil, // No additional confirmations
		SkipConfirmation: nil,

		// Allow runtime permission changes by default (false = not disabled = allowed)
		DisableRuntimePermissionChanges: false,
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
		metrics:           metrics,
		mcpLog:            mcpLog,
		ctx:               ctx,
		cancel:            cancel,
		running:           false,
	}

	logger.DebugfToFile("MCP", "MCP server created (not started yet)")

	return s, nil
}

// Start starts the MCP server on a Unix domain socket.
// The server listens for JSON-RPC tool calls from Claude Desktop.
func (s *MCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("MCP server already running")
	}

	// Remove existing socket if it exists
	if err := removeSocketIfExists(s.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix domain socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}
	s.listener = listener

	// Create mark3labs/mcp-go server and register tools
	s.mcpServer = server.NewMCPServer(
		"CQLAI MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false), // No tool subscriptions needed
	)

	// Register all 9 existing tools
	if err := s.registerTools(); err != nil {
		listener.Close()
		return fmt.Errorf("failed to register tools: %w", err)
	}

	s.running = true
	s.startedAt = time.Now()

	logger.DebugfToFile("MCP", "MCP server started on socket: %s", s.socketPath)
	s.mcpLog.LogServerStart(s.session, s.socketPath)

	// Start accepting connections in a goroutine
	go s.acceptConnections()

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

	// Close listener
	if s.listener != nil {
		s.listener.Close()
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

	// Remove socket file
	removeSocketIfExists(s.socketPath)

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

// acceptConnections accepts connections from Claude Desktop
func (s *MCPServer) acceptConnections() {
	for {
		select {
		case <-s.ctx.Done():
			logger.DebugfToFile("MCP", "Accept loop shutting down")
			return
		default:
			// Accept connection with timeout
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					// Server shutting down, expected error
					return
				default:
					logger.DebugfToFile("MCP", "Error accepting connection: %v", err)
					s.mcpLog.LogError("ACCEPT_ERROR", err)
					continue
				}
			}

			logger.DebugfToFile("MCP", "New MCP connection accepted")
			s.mcpLog.LogClaudeConnected()

			// Handle connection in separate goroutine
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single MCP connection from Claude Desktop.
// The connection is a Unix socket that nc bridges to Claude's stdin/stdout.
func (s *MCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	logger.DebugfToFile("MCP", "Handling MCP connection from Claude Desktop")

	// Create StdioServer to handle JSON-RPC protocol over the socket
	stdioServer := server.NewStdioServer(s.mcpServer)

	// Use the socket connection as stdin/stdout for MCP protocol
	// The connection implements io.Reader and io.Writer
	err := stdioServer.Listen(s.ctx, conn, conn)
	if err != nil {
		logger.DebugfToFile("MCP", "Connection error: %v", err)
		s.mcpLog.LogError("CONNECTION_ERROR", err)
	}

	logger.DebugfToFile("MCP", "MCP connection closed")
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

		s.metrics.RecordToolCall(string(ToolSubmitQueryPlan), false, time.Since(startTime))
		// Return error with request ID and hints on how to confirm or update permissions
		errorMsg := CreateConfirmationRequiredError(opInfo, s.config.GetConfigSnapshot(), req.ID)
		return mcp.NewToolResultError(errorMsg), nil
	}

	// Query approved or no confirmation needed - EXECUTE IT
	logger.DebugfToFile("MCP", "Executing approved query: %s", query)

	// Handle shell commands specially - don't send to Cassandra
	var execResult db.QueryExecutionResult
	if tempClassify.Category == "SESSION" || tempClassify.Category == "FILE" {
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
		// PAGING - display setting, not applicable to MCP
		return db.QueryExecutionResult{
			Result:   "Paging setting acknowledged",
			Duration: time.Since(start),
		}

	case "COPY":
		// COPY TO/FROM - file operations
		// This is complex - for now return not implemented
		return db.QueryExecutionResult{
			Result:   fmt.Errorf("COPY operations not yet implemented in MCP"),
			Duration: time.Since(start),
		}

	case "SOURCE":
		// SOURCE - execute CQL from file
		return db.QueryExecutionResult{
			Result:   fmt.Errorf("SOURCE operations not yet implemented in MCP"),
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

// removeSocketIfExists removes a Unix socket file if it exists
func removeSocketIfExists(path string) error {
	// Check if socket file exists by trying to dial it
	if conn, err := net.Dial("unix", path); err == nil {
		// Socket is active and accepting connections
		conn.Close()
		return fmt.Errorf("socket %s is already in use by another process", path)
	}

	// Socket file may exist but not accepting connections
	// Try to remove it (if it doesn't exist, that's ok)
	if err := removeFile(path); err != nil {
		// Only return error if file exists but can't be removed
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	return nil
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
	return s.confirmationQueue.ConfirmRequest(requestID, confirmedBy)
}

// DenyRequest denies a pending dangerous query request
func (s *MCPServer) DenyRequest(requestID, deniedBy, reason string) error {
	if s.confirmationQueue == nil {
		return fmt.Errorf("confirmation queue not initialized")
	}
	return s.confirmationQueue.DenyRequest(requestID, deniedBy, reason)
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
	return s.confirmationQueue.CancelRequest(requestID, cancelledBy, reason)
}

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
