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

// MCPTestContext holds test context
type MCPTestContext struct {
	Session    *db.Session
	MCPHandler *router.MCPHandler
	SocketPath string
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
