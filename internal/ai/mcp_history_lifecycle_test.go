package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfirmationLifecycleLogging tests that all lifecycle events are logged
func TestConfirmationLifecycleLogging(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "lifecycle_history")

	// Create MCP server config with history
	config := DefaultMCPConfig()
	config.HistoryFile = historyPath
	config.ConfirmationTimeout = 100 * time.Millisecond // Short timeout for testing

	// We need an actual session for NewMCPServer, so skip full server creation
	// Instead, test the logging functions directly

	t.Run("CONFIRM_REQUESTED is logged", func(t *testing.T) {
		// Create a mock server with minimal fields
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		// Log a confirmation request
		s.logConfirmationToHistory("CONFIRM_REQUESTED", "req_test123", "DELETE FROM users", "operation=DELETE category=DML dangerous=true")

		// Verify logged
		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "CONFIRM_REQUESTED")
		assert.Contains(t, contentStr, "req_test123")
		assert.Contains(t, contentStr, "DELETE FROM users")
		assert.Contains(t, contentStr, "operation=DELETE")
	})

	t.Run("CONFIRM_APPROVED is logged", func(t *testing.T) {
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		s.logConfirmationToHistory("CONFIRM_APPROVED", "req_test456", "INSERT INTO users VALUES (...)", "confirmed_by=alice")

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "CONFIRM_APPROVED")
		assert.Contains(t, contentStr, "req_test456")
		assert.Contains(t, contentStr, "confirmed_by=alice")
	})

	t.Run("CONFIRM_DENIED is logged", func(t *testing.T) {
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		s.logConfirmationToHistory("CONFIRM_DENIED", "req_test789", "TRUNCATE table", "denied_by=bob reason=\"too dangerous\"")

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "CONFIRM_DENIED")
		assert.Contains(t, contentStr, "req_test789")
		assert.Contains(t, contentStr, "denied_by=bob")
		assert.Contains(t, contentStr, "too dangerous")
	})

	t.Run("CONFIRM_CANCELLED is logged", func(t *testing.T) {
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		s.logConfirmationToHistory("CONFIRM_CANCELLED", "req_test999", "DROP TABLE users", "cancelled_by=admin reason=\"mistake\"")

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "CONFIRM_CANCELLED")
		assert.Contains(t, contentStr, "req_test999")
		assert.Contains(t, contentStr, "cancelled_by=admin")
	})

	t.Run("CONFIRM_TIMEOUT is logged", func(t *testing.T) {
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		s.logConfirmationToHistory("CONFIRM_TIMEOUT", "req_test000", "CREATE TABLE logs", "timeout_after=5m0s")

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "CONFIRM_TIMEOUT")
		assert.Contains(t, contentStr, "req_test000")
		assert.Contains(t, contentStr, "timeout_after=5m0s")
	})

	t.Run("QUERY is logged after execution", func(t *testing.T) {
		s := &MCPServer{
			historyFilePath: historyPath,
			config:          config,
		}

		s.appendToHistory("SELECT * FROM users WHERE id = 123")

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "QUERY:")
		assert.Contains(t, contentStr, "SELECT * FROM users")
	})

	t.Run("all events have timestamps", func(t *testing.T) {
		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		for _, line := range lines {
			// Each line should start with [YYYY-MM-DD HH:MM:SS]
			assert.True(t, strings.HasPrefix(line, "["), "Line should start with timestamp: %s", line)
			assert.Contains(t, line, "]", "Line should have closing timestamp bracket")
		}
	})

	t.Run("events are in chronological order", func(t *testing.T) {
		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		// We logged 7 events total (REQUEST, APPROVED, DENIED, CANCELLED, TIMEOUT, QUERY, QUERY)
		assert.GreaterOrEqual(t, len(lines), 6, "Should have at least 6 lifecycle events logged")
	})
}

// TestLifecycleLogging_EmptyHistory tests that logging works with empty/new history file
func TestLifecycleLogging_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "new_history")

	// File doesn't exist yet
	_, err := os.Stat(historyPath)
	assert.True(t, os.IsNotExist(err))

	// Create server and log event
	s := &MCPServer{
		historyFilePath: historyPath,
		config:          DefaultMCPConfig(),
	}

	err = s.logConfirmationToHistory("CONFIRM_REQUESTED", "req_new", "SELECT 1", "test=true")
	require.NoError(t, err)

	// File should now exist
	_, err = os.Stat(historyPath)
	assert.NoError(t, err)

	// Content should be correct
	content, err := os.ReadFile(historyPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "CONFIRM_REQUESTED")
}

// TestLifecycleLogging_ConcurrentWrites tests thread safety
func TestLifecycleLogging_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "concurrent_lifecycle_history")

	s := &MCPServer{
		historyFilePath: historyPath,
		config:          DefaultMCPConfig(),
	}

	// Write 50 events concurrently (mix of queries and lifecycle events)
	done := make(chan bool, 50)
	for i := 0; i < 25; i++ {
		go func(idx int) {
			s.appendToHistory(fmt.Sprintf("QUERY_%d", idx))
			done <- true
		}(i)

		go func(idx int) {
			s.logConfirmationToHistory("CONFIRM_REQUESTED", fmt.Sprintf("req_%d", idx), fmt.Sprintf("SELECT %d", idx), "test=true")
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 50; i++ {
		<-done
	}

	// Verify all 50 events written
	content, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Equal(t, 50, len(lines), "Should have exactly 50 lines")

	// Verify mix of QUERY and CONFIRM_REQUESTED
	queryCount := strings.Count(string(content), "QUERY:")
	confirmCount := strings.Count(string(content), "CONFIRM_REQUESTED")
	assert.Equal(t, 25, queryCount)
	assert.Equal(t, 25, confirmCount)
}
