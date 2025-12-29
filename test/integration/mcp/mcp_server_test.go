// +build integration

package mcp_test

import (
	"testing"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServer_NewMCPServer tests creating MCP server from REPL session
func TestMCPServer_NewMCPServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create REPL session
	replSession, err := createTestREPLSession(t)
	if err != nil {
		t.Fatalf("Failed to create REPL session: %v", err)
	}
	defer replSession.Close()

	// Create MCP server
	config := ai.DefaultMCPConfig()
	config.SocketPath = "/tmp/cqlai-test-mcp.sock"

	mcpServer, err := ai.NewMCPServer(replSession, config)
	require.NoError(t, err, "Failed to create MCP server")
	require.NotNil(t, mcpServer, "MCP server should not be nil")

	// Verify server is not running yet
	assert.False(t, mcpServer.IsRunning())

	// Clean up (server has its own session)
	// The server's Close() method (via Stop) should clean up its session
}

// TestMCPServer_StartStop tests server lifecycle
func TestMCPServer_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create REPL session
	replSession, err := createTestREPLSession(t)
	if err != nil {
		t.Fatalf("Failed to create REPL session: %v", err)
	}
	defer replSession.Close()

	// Create MCP server
	config := ai.DefaultMCPConfig()
	config.SocketPath = "/tmp/cqlai-test-mcp-lifecycle.sock"

	mcpServer, err := ai.NewMCPServer(replSession, config)
	require.NoError(t, err)

	// Start server
	err = mcpServer.Start()
	require.NoError(t, err, "Failed to start MCP server")

	assert.True(t, mcpServer.IsRunning(), "Server should be running after Start()")

	// Try to start again - should error
	err = mcpServer.Start()
	assert.Error(t, err, "Starting already-running server should error")
	assert.Contains(t, err.Error(), "already running")

	// Stop server
	err = mcpServer.Stop()
	require.NoError(t, err, "Failed to stop MCP server")

	assert.False(t, mcpServer.IsRunning(), "Server should not be running after Stop()")

	// Try to stop again - should error
	err = mcpServer.Stop()
	assert.Error(t, err, "Stopping stopped server should error")
	assert.Contains(t, err.Error(), "not running")
}

// TestMCPServer_IndependentSession tests that MCP session is independent from REPL
func TestMCPServer_IndependentSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create REPL session
	replSession, err := createTestREPLSession(t)
	if err != nil {
		t.Fatalf("Failed to create REPL session: %v", err)
	}
	defer replSession.Close()

	// Create MCP server
	config := ai.DefaultMCPConfig()
	config.SocketPath = "/tmp/cqlai-test-mcp-independent.sock"

	mcpServer, err := ai.NewMCPServer(replSession, config)
	require.NoError(t, err)
	defer mcpServer.Stop()

	// Verify MCP has its own session by modifying REPL session
	replSession.SetConsistency("QUORUM")
	replSession.SetPageSize(500)

	// MCP server should have started with different defaults
	// We can't directly access MCP's session, but we can verify it was created
	// by checking the server started successfully
	err = mcpServer.Start()
	require.NoError(t, err)
	defer mcpServer.Stop()

	// Both sessions should be able to query independently
	// We'll verify this through metrics - if MCP crashed it wouldn't work
	assert.True(t, mcpServer.IsRunning())

	// Verify metrics exist
	metrics := mcpServer.GetMetrics()
	assert.NotNil(t, metrics)
}

// TestMCPServer_MetricsCollection tests metrics tracking
func TestMCPServer_MetricsCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create REPL session
	replSession, err := createTestREPLSession(t)
	if err != nil {
		t.Fatalf("Failed to create REPL session: %v", err)
	}
	defer replSession.Close()

	// Create MCP server
	config := ai.DefaultMCPConfig()
	config.SocketPath = "/tmp/cqlai-test-mcp-metrics.sock"

	mcpServer, err := ai.NewMCPServer(replSession, config)
	require.NoError(t, err)

	// Get initial metrics
	metrics := mcpServer.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
	assert.Equal(t, 0.0, metrics.SuccessRate)

	// Start server (this will register tools and do some internal operations)
	err = mcpServer.Start()
	require.NoError(t, err)
	defer mcpServer.Stop()

	// Metrics should still show zero requests (no actual tool calls yet)
	metrics = mcpServer.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
}

// Helper function to create a test REPL session
func createTestREPLSession(t *testing.T) (*db.Session, error) {
	cluster := getTestCluster(t)

	// Create session
	session, err := db.NewSessionFromCluster(cluster, "cassandra", false)
	if err != nil {
		return nil, err
	}

	// Verify session can query
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

	t.Logf("Connected to Cassandra %s", version)

	return session, nil
}
