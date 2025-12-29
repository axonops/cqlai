package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
	confirmQueue  chan *ConfirmationRequest
	responseQueue chan *ConfirmationResponse

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

// MCPServerConfig holds configuration for starting the MCP server
type MCPServerConfig struct {
	SocketPath         string
	ConfirmationMode   string // "dangerous_only", "all", "none", "interactive"
	ConfirmationTimeout time.Duration
	LogLevel           string // "debug", "info", "warning", "error"
	LogFile            string
}

// DefaultMCPConfig returns default MCP server configuration
func DefaultMCPConfig() *MCPServerConfig {
	return &MCPServerConfig{
		SocketPath:          "/tmp/cqlai-mcp.sock",
		ConfirmationMode:    "dangerous_only",
		ConfirmationTimeout: 5 * time.Minute,
		LogLevel:            "info",
		LogFile:             "/tmp/cqlai-mcp.log",
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

	// Create confirmation channels
	confirmQueue := make(chan *ConfirmationRequest, 10)
	responseQueue := make(chan *ConfirmationResponse, 10)

	s := &MCPServer{
		session:       mcpSession,
		cache:         cache,
		resolver:      resolver,
		socketPath:    config.SocketPath,
		confirmQueue:  confirmQueue,
		responseQueue: responseQueue,
		metrics:       metrics,
		mcpLog:        mcpLog,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
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

	// Close confirmation channels
	close(s.confirmQueue)
	close(s.responseQueue)

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

	logger.DebugfToFile("MCP", "Registered %d tools", len(toolDefs))

	return nil
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
