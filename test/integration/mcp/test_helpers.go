package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestCluster returns test cluster config
func getTestCluster(t *testing.T) *gocql.ClusterConfig {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 5 * time.Second
	cluster.Consistency = gocql.LocalOne

	return cluster
}

// createTestREPLSession creates a test session
func createTestREPLSession(t *testing.T) (*db.Session, error) {
	cluster := getTestCluster(t)

	session, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	if err != nil {
		return nil, err
	}

	// Verify session works
	var version string
	iter := session.Query("SELECT release_version FROM system.local").Iter()
	if !iter.Scan(&version) {
		session.Close()
		return nil, iter.Close()
	}
	if err := iter.Close(); err != nil {
		session.Close()
		return nil, err
	}

	return session, nil
}

// MCPTestContext holds test context for socket-based tests (DEPRECATED)
type MCPTestContext struct {
	Session    *db.Session
	MCPHandler *router.MCPHandler
	SocketPath string
}

// HTTPTestContext holds test context for HTTP-based tests
type HTTPTestContext struct {
	Session      *db.Session
	MCPHandler   *router.MCPHandler
	HTTPHost     string
	HTTPPort     int
	APIKey       string
	BaseURL      string
	MCPSessionID string // MCP protocol session ID
}

// startMCPFromConfig starts MCP using JSON config file
// Uses the ACTUAL .mcp start --config-file implementation
func startMCPFromConfig(t *testing.T, configPath string) *MCPTestContext {
	// Create session
	replSession, err := createTestREPLSession(t)
	require.NoError(t, err)

	// Init MCP handler
	err = router.InitMCPHandler(replSession)
	require.NoError(t, err)

	mcpHandler := router.GetMCPHandler()
	require.NotNil(t, mcpHandler)

	// Extract socket path from config first
	data, _ := os.ReadFile(configPath)
	var jsonConfig map[string]any
	json.Unmarshal(data, &jsonConfig)

	socketPath := "/tmp/cqlai-mcp-test.sock"
	if sp, ok := jsonConfig["socket_path"].(string); ok {
		socketPath = sp
	}

	// Clean up any existing socket from previous failed tests
	os.Remove(socketPath) // Ignore errors - socket may not exist

	// Use the REAL .mcp start command with --config-file
	// This tests the actual code path users will use!
	cmd := fmt.Sprintf(".mcp start --config-file %s", configPath)

	// Start server
	result := mcpHandler.HandleMCPCommand(cmd)
	require.Contains(t, result, "started successfully", "MCP start failed: %s", result)

	time.Sleep(500 * time.Millisecond)

	return &MCPTestContext{
		Session:    replSession,
		MCPHandler: mcpHandler,
		SocketPath: socketPath,
	}
}

// buildStartCommandFromConfig builds .mcp start command from JSON config
func buildStartCommandFromConfig(config map[string]any, socketPath string) string {
	cmd := fmt.Sprintf(".mcp start --socket-path %s", socketPath)

	// Add mode
	if mode, ok := config["mode"].(string); ok && mode != "" {
		switch mode {
		case "readonly":
			cmd += " --readonly_mode"
		case "readwrite":
			cmd += " --readwrite_mode"
		case "dba":
			cmd += " --dba_mode"
		}

		// Add confirm-queries if present
		if confirmQueries, ok := config["confirm_queries"].([]any); ok && len(confirmQueries) > 0 {
			cats := make([]string, len(confirmQueries))
			for i, c := range confirmQueries {
				cats[i] = c.(string)
			}
			cmd += " --confirm-queries " + strings.Join(cats, ",")
		}
	}

	// Add skip-confirmation if present
	if skipConf, ok := config["skip_confirmation"].([]any); ok && len(skipConf) > 0 {
		cats := make([]string, len(skipConf))
		for i, c := range skipConf {
			cats[i] = c.(string)
		}
		cmd += " --skip-confirmation " + strings.Join(cats, ",")
	}

	// Add lockdown if present
	if lockdown, ok := config["disable_runtime_permission_changes"].(bool); ok && lockdown {
		cmd += " --disable-runtime-permission-changes"
	}

	// Add log level if present
	if logLevel, ok := config["log_level"].(string); ok && logLevel != "" {
		cmd += " --log-level " + logLevel
	}

	return cmd
}

// ensureTestDataExists verifies test_mcp keyspace and users table exist
// Call this at start of test functions that need the test data
func ensureTestDataExists(t *testing.T, session *db.Session) {
	// Create test_mcp keyspace if it doesn't exist
	err := session.Query("CREATE KEYSPACE IF NOT EXISTS test_mcp WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
	if err != nil {
		t.Logf("Warning: CREATE KEYSPACE test_mcp failed: %v (may already exist)", err)
	}

	// Create table in test_mcp keyspace
	err = session.Query("CREATE TABLE IF NOT EXISTS test_mcp.users (id uuid PRIMARY KEY, email text, name text, created_at timestamp, role text, is_active boolean)").Exec()
	if err != nil {
		t.Logf("Warning: CREATE TABLE IF NOT EXISTS failed: %v (table may already exist)", err)
	}

	// Ensure at least one row exists (ignore errors if it already exists)
	session.Query("INSERT INTO test_mcp.users (id, email, name, role, is_active) VALUES (11111111-1111-1111-1111-111111111111, 'alice@example.com', 'Alice Admin', 'admin', true)").Exec()
}

// stopMCP stops MCP and cleans up
func stopMCP(ctx *MCPTestContext) {
	if ctx == nil {
		return
	}
	if ctx.MCPHandler != nil {
		ctx.MCPHandler.HandleMCPCommand(".mcp stop")
	}
	if ctx.Session != nil {
		ctx.Session.Close()
	}
	// Clean up socket file
	if ctx.SocketPath != "" {
		time.Sleep(100 * time.Millisecond) // Brief pause for socket to close
		os.Remove(ctx.SocketPath)         // Ignore errors
	}
}

// callTool calls MCP tool via socket
func callTool(t *testing.T, socketPath string, toolName string, args map[string]any) map[string]any {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Logf("Failed to connect: %v", err)
		return nil
	}
	defer conn.Close()

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestJSON, _ := json.Marshal(request)
	conn.Write(requestJSON)
	conn.Write([]byte("\n"))

	reader := bufio.NewReader(conn)
	responseLine, err := reader.ReadBytes('\n')
	if err != nil {
		t.Logf("Failed to read: %v", err)
		return nil
	}

	var response map[string]any
	json.Unmarshal(responseLine, &response)
	return response
}

// Helper assertions
func assertIsError(t *testing.T, resp map[string]any, msg string) {
	if resp == nil {
		t.Fatalf("nil response: %s", msg)
	}
	result := resp["result"].(map[string]any)
	isError, _ := result["isError"].(bool)
	assert.True(t, isError, msg)
}

func assertNotError(t *testing.T, resp map[string]any, msg string) {
	if resp == nil {
		t.Fatalf("nil response: %s", msg)
	}
	result := resp["result"].(map[string]any)
	isError, _ := result["isError"].(bool)

	// Debug: Print actual error if present
	if isError {
		content, ok := result["content"].([]any)
		if ok && len(content) > 0 {
			if textObj, ok := content[0].(map[string]any); ok {
				if text, ok := textObj["text"].(string); ok {
					t.Logf("ERROR DETAILS: %s", text)
				}
			}
		}
	}

	assert.False(t, isError, msg)
}

func assertContains(t *testing.T, resp map[string]any, substring string) {
	if resp == nil {
		return
	}
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	if len(content) > 0 {
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, substring)
	}
}

func extractText(t *testing.T, resp map[string]any) string {
	if resp == nil {
		return ""
	}
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	if len(content) > 0 {
		return content[0].(map[string]any)["text"].(string)
	}
	return ""
}

func extractRequestID(text string) string {
	for i := 0; i < len(text)-6; i++ {
		if text[i:i+4] == "req_" {
			end := i + 7
			if end <= len(text) {
				return text[i:end]
			}
		}
	}
	return ""
}

// assertRequiresConfirmation verifies a query requires confirmation by submitting it
// in a goroutine, checking pending list, and cancelling to cleanup
func assertRequiresConfirmation(t *testing.T, ctx *HTTPTestContext, toolName string, args map[string]any) {
	done := make(chan bool)

	// Submit query in goroutine (will block if confirmation needed)
	go func() {
		callToolHTTP(t, ctx, toolName, args)
		done <- true
	}()

	// Give it time to create confirmation request
	time.Sleep(500 * time.Millisecond)

	// Check pending confirmations
	pendingResp := callToolHTTP(t, ctx, "get_pending_confirmations", map[string]any{})
	assertNotError(t, pendingResp, "get_pending should work")
	text := extractText(t, pendingResp)
	assert.Contains(t, text, "req_", "Should have pending confirmation request")

	// Cancel to unblock goroutine
	requestID := extractRequestID(text)
	if requestID != "" {
		callToolHTTP(t, ctx, "cancel_confirmation", map[string]any{"request_id": requestID})
	}

	// Wait for goroutine
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Log("Warning: Goroutine may still be running")
	}
}
func buildOperationParams(operation, keyspace, table string) map[string]any {
	// Parse compound operations like "LIST ROLES" -> operation="LIST", options={"object_type":"ROLES"}
	parts := strings.Fields(operation)
	mainOp := parts[0]

	params := map[string]any{
		"operation": mainOp,
		"keyspace":  keyspace,
		"table":     table,
	}

	// Handle compound operations (LIST ROLES, SHOW VERSION, etc.)
	if len(parts) > 1 {
		subType := strings.Join(parts[1:], " ")

		switch strings.ToUpper(mainOp) {
		case "LIST":
			params["options"] = map[string]any{
				"object_type": subType,
			}
			return params
		case "SHOW":
			params["options"] = map[string]any{
				"show_type": subType,
			}
			return params
		case "DESCRIBE", "DESC":
			// Keep operation as DESCRIBE and use existing renderDescribe logic
			params["operation"] = "DESCRIBE"
			// Set table to the describe target for specific describes
			if strings.Contains(subType, "TABLE") && !strings.Contains(subType, "TABLES") {
				params["table"] = table
			} else if strings.Contains(subType, "TYPE") && !strings.Contains(subType, "TYPES") {
				params["table"] = "TYPE"
			} else {
				params["table"] = subType
			}
			return params
		case "CREATE", "DROP", "ALTER":
			// Handle CREATE/DROP/ALTER with specific types (INDEX, TYPE, FUNCTION, etc.)
			params["options"] = map[string]any{
				"object_type": subType,
			}
			// Fall through to add type-specific parameters below
		case "GRANT":
			// GRANT ROLE vs GRANT permission
			if strings.Contains(subType, "ROLE") {
				params["options"] = map[string]any{
					"grant_type": "ROLE",
					"role":       "developer",
					"to_role":    "app_admin",
				}
				return params
			}
			// Fall through for regular permission grants
		case "REVOKE":
			// REVOKE ROLE vs REVOKE permission
			if strings.Contains(subType, "ROLE") {
				params["options"] = map[string]any{
					"grant_type": "ROLE",
					"role":       "developer",
					"from_role":  "app_admin",
				}
				return params
			}
			// Fall through for regular permission revokes
		}
	}

	// After compound operation parsing, add type-specific parameters for CREATE/DROP operations
	if opts, ok := params["options"].(map[string]any); ok {
		if objType, ok := opts["object_type"].(string); ok {
			objTypeUpper := strings.ToUpper(objType)
			switch objTypeUpper {
			case "INDEX":
				opts["index_name"] = "test_idx"
				opts["column"] = "email"
				// Don't need schema for INDEX
				delete(params, "schema")
			case "CUSTOM INDEX":
				opts["index_name"] = "test_idx"
				opts["column"] = "email"
				opts["using"] = "org.apache.cassandra.index.sasi.SASIIndex"
				// For CUSTOM INDEX, we need to normalize to INDEX for the renderer
				opts["object_type"] = "INDEX"
				opts["custom"] = true
				delete(params, "schema")
			case "TYPE":
				opts["type_name"] = "test_type"
				// Keep schema for TYPE (UDT fields)
				params["schema"] = map[string]any{
					"street": "text",
					"city":   "text",
				}
			case "FUNCTION", "OR REPLACE FUNCTION":
				opts["function_name"] = "test_func"
				opts["returns"] = "int"
				opts["language"] = "java"
				opts["body"] = "return 0;"
				delete(params, "schema")
			case "AGGREGATE", "OR REPLACE AGGREGATE":
				opts["aggregate_name"] = "test_agg"
				opts["sfunc"] = "state_func"
				opts["stype"] = "int"
				opts["input_type"] = "int"
				delete(params, "schema")
			case "TRIGGER":
				opts["trigger_name"] = "test_trigger"
				opts["trigger_class"] = "org.apache.cassandra.triggers.TestTrigger"
				delete(params, "schema")
			case "MATERIALIZED VIEW":
				opts["view_name"] = "test_mv"
				opts["base_table"] = table
				opts["primary_key"] = "(id, name)"
				params["columns"] = []string{"id", "name"}
				params["where"] = []any{
					map[string]any{
						"column":   "id",
						"operator": "IS NOT NULL",
						"value":    nil,
					},
					map[string]any{
						"column":   "name",
						"operator": "IS NOT NULL",
						"value":    nil,
					},
				}
				delete(params, "schema")
			case "ROLE":
				opts["role_name"] = "test_role"
				opts["password"] = "test_pass"
				opts["login"] = true
				opts["superuser"] = false
				delete(params, "schema")
			case "USER":
				opts["user_name"] = "test_user"
				opts["password"] = "test_pass"
				opts["superuser"] = false
				delete(params, "schema")
			}
		}
	}

	// Add required parameters based on operation
	switch mainOp {
	case "INSERT":
		params["values"] = map[string]any{
			"id":    "00000000-0000-0000-0000-000000000060",
			"name":  "Test User",
			"email": "test@example.com",
		}
	case "UPDATE":
		params["values"] = map[string]any{"name": "Updated Name"}
		params["where"] = []any{
			map[string]any{
				"column":   "id",
				"operator": "=",
				"value":    "00000000-0000-0000-0000-000000000060",
			},
		}
	case "DELETE":
		params["where"] = []any{
			map[string]any{
				"column":   "id",
				"operator": "=",
				"value":    "00000000-0000-0000-0000-000000000060",
			},
		}
	case "CREATE TABLE", "CREATE COLUMNFAMILY", "CREATE", "CREATE KEYSPACE", "CREATE INDEX", "CREATE CUSTOM INDEX",
		"CREATE MATERIALIZED VIEW", "CREATE TYPE", "CREATE FUNCTION", "CREATE OR REPLACE FUNCTION",
		"CREATE AGGREGATE", "CREATE OR REPLACE AGGREGATE", "CREATE TRIGGER", "CREATE ROLE", "CREATE USER":
		// Add IF NOT EXISTS to prevent conflicts when tests run multiple times
		if params["options"] == nil {
			params["options"] = map[string]any{}
		}
		if opts, ok := params["options"].(map[string]any); ok {
			opts["if_not_exists"] = true
		}
		// Default schema for TABLE/KEYSPACE only (others handled above)
		if _, hasSchema := params["schema"]; !hasSchema {
			if opts, ok := params["options"].(map[string]any); ok {
				if objType, ok := opts["object_type"].(string); ok {
					objTypeUpper := strings.ToUpper(objType)
					// Only add schema for TABLE, COLUMNFAMILY, KEYSPACE, or no object_type
					if objTypeUpper == "TABLE" || objTypeUpper == "COLUMNFAMILY" || objTypeUpper == "KEYSPACE" || objTypeUpper == "" {
						params["schema"] = map[string]any{
							"id":   "uuid PRIMARY KEY",
							"name": "text",
						}
					}
				}
			} else {
				// No options, default to table schema
				params["schema"] = map[string]any{
					"id":   "uuid PRIMARY KEY",
					"name": "text",
				}
			}
		}
	case "ALTER TABLE", "ALTER COLUMNFAMILY", "ALTER", "ALTER KEYSPACE", "ALTER MATERIALIZED VIEW",
		"ALTER TYPE", "ALTER ROLE", "ALTER USER":
		params["options"] = map[string]any{
			"object_type": "TABLE",
			"action":      "ADD",
			"column_name": "age",
			"column_type": "int",
		}
	case "GRANT", "GRANT ROLE":
		params["options"] = map[string]any{
			"permission": "SELECT",
			"role":       "app_readonly",
		}
	case "REVOKE":
		params["options"] = map[string]any{
			"permission": "SELECT",
			"role":       "app_readonly",
		}
	case "REVOKE ROLE":
		params["options"] = map[string]any{
			"grant_type": "ROLE",
			"role":       "developer",
			"from_role":  "app_admin",
		}
	case "ADD IDENTITY":
		params["operation"] = "ALTER"
		params["options"] = map[string]any{
			"object_type": "ROLE",
			"action":      "ADD_IDENTITY",
			"identity":    "user@REALM",
			"role_name":   "app_role",
		}
	case "DROP IDENTITY":
		params["operation"] = "ALTER"
		params["options"] = map[string]any{
			"object_type": "ROLE",
			"action":      "DROP_IDENTITY",
			"identity":    "user@REALM",
			"role_name":   "app_role",
		}
	}

	return params
}
