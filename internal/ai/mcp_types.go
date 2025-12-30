package ai

import (
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// ConfirmationRequest represents a request for user confirmation of a dangerous query
type ConfirmationRequest struct {
	ID              string                // Unique ID: req_001, req_002, etc
	Timestamp       time.Time             // When requested
	Timeout         time.Time             // Expiration time
	Query           string                // CQL query string
	Classification  QueryClassification   // Danger level, operation type
	Tool            string                // Which tool requested
	ToolOperation   string                // Specific operation (if applicable)
	Status          string                // PENDING, CONFIRMED, DENIED, CANCELLED, TIMEOUT
	UserConfirmed   bool
	ConfirmedBy     string                // Username who confirmed
	ConfirmedAt     time.Time

	// Execution metadata (set after query executes)
	Executed        bool          // Whether query was actually executed
	ExecutionTime   time.Duration // How long execution took
	TraceID         []byte        // Cassandra trace ID (if tracing enabled)
	ExecutedAt      time.Time     // When execution happened
	ExecutionError  string        // Error message if execution failed
	RowsAffected    int           // Rows affected (for DML) or returned (for SELECT)
}

// ConfirmationResponse represents user's response to a confirmation request
type ConfirmationResponse struct {
	RequestID string // req_001, etc
	Approved  bool   // true = confirm, false = deny
	Reason    string // Why (if denied)
}

// QueryClassification contains information about a query's danger level
type QueryClassification struct {
	IsDangerous bool
	Severity    string // CRITICAL, HIGH, MEDIUM, LOW, SAFE
	Operation   string // DELETE, UPDATE, CREATE, etc
	Table       string
	Description string
}

// MetricsCollector collects metrics about MCP server operations
type MetricsCollector struct {
	totalRequests    int64
	successfulRequests int64
	failedRequests   int64
	toolCalls        map[string]int64
	mu               sync.Mutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		toolCalls: make(map[string]int64),
	}
}

// RecordToolCall records a tool invocation
func (m *MetricsCollector) RecordToolCall(toolName string, success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	if success {
		m.successfulRequests++
	} else {
		m.failedRequests++
	}
	m.toolCalls[toolName]++
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	TotalRequests    int64
	SuccessfulRequests int64
	FailedRequests   int64
	ToolCalls        map[string]int64
	SuccessRate      float64
}

// GetSnapshot returns a snapshot of current metrics
func (m *MetricsCollector) GetSnapshot() *MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	successRate := 0.0
	if m.totalRequests > 0 {
		successRate = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// Copy tool calls map
	toolCallsCopy := make(map[string]int64, len(m.toolCalls))
	for k, v := range m.toolCalls {
		toolCallsCopy[k] = v
	}

	return &MetricsSnapshot{
		TotalRequests:      m.totalRequests,
		SuccessfulRequests: m.successfulRequests,
		FailedRequests:     m.failedRequests,
		ToolCalls:          toolCallsCopy,
		SuccessRate:        successRate,
	}
}

// MCPLogger handles logging for MCP operations
type MCPLogger struct {
	// TODO: Implement full logging system
	logFile  string
	logLevel string
}

// NewMCPLogger creates a new MCP logger
func NewMCPLogger(logFile, logLevel string) (*MCPLogger, error) {
	// TODO: Open log file, validate log level
	return &MCPLogger{
		logFile:  logFile,
		logLevel: logLevel,
	}, nil
}

// LogServerStart logs MCP server startup
func (l *MCPLogger) LogServerStart(session *db.Session, socketPath string) {
	logger.DebugfToFile("MCPLogger", "Server started on %s", socketPath)
}

// LogServerStop logs MCP server shutdown
func (l *MCPLogger) LogServerStop(uptime time.Duration, metrics *MetricsCollector) {
	snapshot := metrics.GetSnapshot()
	logger.DebugfToFile("MCPLogger", "Server stopped (uptime: %v, requests: %d, success: %.1f%%)",
		uptime, snapshot.TotalRequests, snapshot.SuccessRate)
}

// LogClaudeConnected logs when Claude Desktop connects
func (l *MCPLogger) LogClaudeConnected() {
	logger.DebugfToFile("MCPLogger", "Claude Desktop connected")
}

// LogError logs an error
func (l *MCPLogger) LogError(eventType string, err error) {
	logger.DebugfToFile("MCPLogger", "ERROR [%s]: %v", eventType, err)
}

// LogToolExecution logs a tool execution
func (l *MCPLogger) LogToolExecution(toolName string, params map[string]any, result any, err error, duration time.Duration) {
	if err != nil {
		logger.DebugfToFile("MCPLogger", "Tool %s failed in %v: %v", toolName, duration, err)
	} else {
		logger.DebugfToFile("MCPLogger", "Tool %s succeeded in %v", toolName, duration)
	}
}

// Close closes the logger
func (l *MCPLogger) Close() error {
	// TODO: Close log file
	return nil
}
