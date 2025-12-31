package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppendQueryToHistory tests query history append functionality
func TestAppendQueryToHistory(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "test_history")

	t.Run("append single query", func(t *testing.T) {
		err := appendQueryToHistory(historyPath, "SELECT * FROM users")
		require.NoError(t, err)

		// Verify file was created
		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "SELECT * FROM users")
		assert.Contains(t, contentStr, "[") // Timestamp prefix
	})

	t.Run("append multiple queries", func(t *testing.T) {
		queries := []string{
			"INSERT INTO users VALUES (...)",
			"UPDATE users SET name = 'test'",
			"DELETE FROM users WHERE id = 1",
		}

		for _, query := range queries {
			err := appendQueryToHistory(historyPath, query)
			require.NoError(t, err)
		}

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		contentStr := string(content)
		for _, query := range queries {
			assert.Contains(t, contentStr, query)
		}

		// Should have at least 4 lines (1 from first test + 3 from this test)
		lines := strings.Split(strings.TrimSpace(contentStr), "\n")
		assert.GreaterOrEqual(t, len(lines), 4)
	})

	t.Run("ignore empty queries", func(t *testing.T) {
		beforeContent, _ := os.ReadFile(historyPath)
		beforeLines := len(strings.Split(strings.TrimSpace(string(beforeContent)), "\n"))

		err := appendQueryToHistory(historyPath, "")
		require.NoError(t, err)

		err = appendQueryToHistory(historyPath, "   ")
		require.NoError(t, err)

		afterContent, _ := os.ReadFile(historyPath)
		afterLines := len(strings.Split(strings.TrimSpace(string(afterContent)), "\n"))

		// Should be same number of lines (empty queries not added)
		assert.Equal(t, beforeLines, afterLines)
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		newPath := filepath.Join(tmpDir, "subdir", "another", "history")
		err := appendQueryToHistory(newPath, "SELECT 1")
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})

	t.Run("handles special characters", func(t *testing.T) {
		specialQuery := "INSERT INTO users (name) VALUES ('O''Brien')"
		err := appendQueryToHistory(historyPath, specialQuery)
		require.NoError(t, err)

		content, err := os.ReadFile(historyPath)
		require.NoError(t, err)

		assert.Contains(t, string(content), specialQuery)
	})
}

// TestAppendQueryToHistory_ThreadSafety tests concurrent writes
func TestAppendQueryToHistory_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "concurrent_history")

	// Write 100 queries concurrently
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			err := appendQueryToHistory(historyPath, fmt.Sprintf("QUERY_%d", idx))
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify all 100 queries were written
	content, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Equal(t, 100, len(lines), "Should have exactly 100 lines (one per query)")
}
