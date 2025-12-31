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
		err := appendQueryToHistory(historyPath, "SELECT * FROM users", 10*1024*1024, 5)
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
			err := appendQueryToHistory(historyPath, query, 10*1024*1024, 5)
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

		err := appendQueryToHistory(historyPath, "", 10*1024*1024, 5)
		require.NoError(t, err)

		err = appendQueryToHistory(historyPath, "   ", 10*1024*1024, 5)
		require.NoError(t, err)

		afterContent, _ := os.ReadFile(historyPath)
		afterLines := len(strings.Split(strings.TrimSpace(string(afterContent)), "\n"))

		// Should be same number of lines (empty queries not added)
		assert.Equal(t, beforeLines, afterLines)
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		newPath := filepath.Join(tmpDir, "subdir", "another", "history")
		err := appendQueryToHistory(newPath, "SELECT 1", 10*1024*1024, 5)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})

	t.Run("handles special characters", func(t *testing.T) {
		specialQuery := "INSERT INTO users (name) VALUES ('O''Brien')"
		err := appendQueryToHistory(historyPath, specialQuery, 10*1024*1024, 5)
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
			err := appendQueryToHistory(historyPath, fmt.Sprintf("QUERY_%d", idx), 10*1024*1024, 5)
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

// TestHistoryRotation tests file rotation with gzip compression
func TestHistoryRotation(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "rotation_test_history")

	// Use small size for testing (1KB)
	smallSize := int64(1024)
	maxRotations := 3

	t.Run("rotates when size exceeded", func(t *testing.T) {
		// Write enough data to exceed size limit
		largeQuery := strings.Repeat("SELECT * FROM users WHERE id = 'x'; ", 50) // ~2KB
		err := appendQueryToHistory(historyPath, largeQuery, smallSize, maxRotations)
		require.NoError(t, err)

		// Should have created .1.gz
		gzPath := historyPath + ".1.gz"
		_, err = os.Stat(gzPath)
		assert.NoError(t, err, "Should have created .1.gz file")

		// Original should be removed/started fresh
		info, err := os.Stat(historyPath)
		require.NoError(t, err)
		assert.Less(t, info.Size(), smallSize, "New file should be smaller than limit")
	})

	t.Run("keeps max rotations", func(t *testing.T) {
		// Trigger multiple rotations
		for i := 0; i < 5; i++ {
			largeQuery := strings.Repeat(fmt.Sprintf("QUERY %d: SELECT * FROM users; ", i), 100)
			err := appendQueryToHistory(historyPath, largeQuery, smallSize, maxRotations)
			require.NoError(t, err)
		}

		// Should have .1.gz, .2.gz, .3.gz (maxRotations=3)
		for i := 1; i <= maxRotations; i++ {
			gzPath := fmt.Sprintf("%s.%d.gz", historyPath, i)
			_, err := os.Stat(gzPath)
			assert.NoError(t, err, "Should have .%d.gz", i)
		}

		// Should NOT have .4.gz (exceeds maxRotations)
		gzPath4 := fmt.Sprintf("%s.%d.gz", historyPath, 4)
		_, err := os.Stat(gzPath4)
		assert.True(t, os.IsNotExist(err), "Should not have .4.gz (exceeds maxRotations)")
	})

	t.Run("zero rotations deletes file", func(t *testing.T) {
		zeroRotPath := filepath.Join(tmpDir, "zero_rotation_history")
		
		// Write large query
		largeQuery := strings.Repeat("SELECT * FROM big_table; ", 100)
		err := appendQueryToHistory(zeroRotPath, largeQuery, smallSize, 0)
		require.NoError(t, err)

		// Should have removed file (maxRotations=0)
		// New file should exist but be fresh
		info, err := os.Stat(zeroRotPath)
		require.NoError(t, err)
		assert.Less(t, info.Size(), int64(1000), "Should be a fresh small file")

		// Should not have any .gz files
		matches, _ := filepath.Glob(zeroRotPath + "*.gz")
		assert.Equal(t, 0, len(matches), "Should not have any .gz files with maxRotations=0")
	})
}

// TestCompressFile tests gzip compression
func TestCompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	destPath := filepath.Join(tmpDir, "compressed.gz")

	// Create source file with compressible content
	content := strings.Repeat("This is test data that compresses well. ", 1000)
	err := os.WriteFile(srcPath, []byte(content), 0600)
	require.NoError(t, err)

	// Compress it
	err = compressFile(srcPath, destPath)
	require.NoError(t, err)

	// Verify compressed file exists
	compInfo, err := os.Stat(destPath)
	require.NoError(t, err)

	srcInfo, _ := os.Stat(srcPath)
	assert.Less(t, compInfo.Size(), srcInfo.Size(), "Compressed file should be smaller than source")
}
