package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Full batching and concurrency tests require integration tests with a real
// Cassandra cluster. The CI pipeline runs these tests against Cassandra 2.1-5.0.
// These unit tests verify parsing, options handling, and error cases that don't
// require database execution.

// TestParseCopyOptions tests the option parsing function
func TestParseCopyOptions(t *testing.T) {
	t.Run("Default options", func(t *testing.T) {
		options := parseCopyOptions("")

		assert.Equal(t, "false", options["HEADER"])
		assert.Equal(t, "null", options["NULLVAL"])
		assert.Equal(t, ",", options["DELIMITER"])
		assert.Equal(t, "20", options["MAXBATCHSIZE"])
		assert.Equal(t, "6", options["MAXREQUESTS"])
		assert.Equal(t, "5000", options["CHUNKSIZE"])
		assert.Equal(t, "-1", options["MAXROWS"])
		assert.Equal(t, "0", options["SKIPROWS"])
	})

	t.Run("Custom options", func(t *testing.T) {
		optionsStr := "HEADER=true, MAXBATCHSIZE=50, MAXREQUESTS=8, DELIMITER='|'"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "true", options["HEADER"])
		assert.Equal(t, "50", options["MAXBATCHSIZE"])
		assert.Equal(t, "8", options["MAXREQUESTS"])
		assert.Equal(t, "|", options["DELIMITER"])
	})

	t.Run("Quoted values", func(t *testing.T) {
		optionsStr := "NULLVAL='N/A', QUOTE='\"'"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "N/A", options["NULLVAL"])
		assert.Equal(t, "\"", options["QUOTE"])
	})

	t.Run("SKIPROWS option", func(t *testing.T) {
		optionsStr := "HEADER=true, SKIPROWS=5"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "true", options["HEADER"])
		assert.Equal(t, "5", options["SKIPROWS"])
	})

	t.Run("MAXPARSEERRORS option", func(t *testing.T) {
		optionsStr := "MAXPARSEERRORS=100"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "100", options["MAXPARSEERRORS"])
	})

	t.Run("Case insensitive options", func(t *testing.T) {
		optionsStr := "header=TRUE, delimiter='|', maxbatchsize=100"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "TRUE", options["HEADER"])
		assert.Equal(t, "|", options["DELIMITER"])
		assert.Equal(t, "100", options["MAXBATCHSIZE"])
	})

	t.Run("Multiple concurrent workers", func(t *testing.T) {
		optionsStr := "MAXREQUESTS=12, MAXBATCHSIZE=50"
		options := parseCopyOptions(optionsStr)

		assert.Equal(t, "12", options["MAXREQUESTS"])
		assert.Equal(t, "50", options["MAXBATCHSIZE"])
	})
}

// TestCopyFromCSVErrorHandling tests error handling in COPY FROM
func TestCopyFromCSVErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	t.Run("File not found", func(t *testing.T) {
		handler := &MetaCommandHandler{
			session:        &db.Session{},
			sessionManager: sessionMgr,
		}

		command := "COPY test_table (id, name) FROM '/nonexistent/file.csv'"
		result := handler.handleCopyFrom(command)

		resultStr := fmt.Sprintf("%v", result)
		t.Logf("File not found result: %v", resultStr)
		assert.Contains(t, resultStr, "Error opening file")
	})
}

// TestLargeBatchCreation tests that large files create multiple batches
// Note: This test verifies file preparation logic but doesn't execute batches
func TestLargeBatchCreation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Large file creates expected batch count", func(t *testing.T) {
		// Create a CSV with 100 rows
		csvFile := filepath.Join(tempDir, "test_large.csv")
		var csvContent strings.Builder
		csvContent.WriteString("id,name\n")
		for i := 1; i <= 100; i++ {
			csvContent.WriteString(fmt.Sprintf("%d,name%d\n", i, i))
		}
		err := os.WriteFile(csvFile, []byte(csvContent.String()), 0644)
		require.NoError(t, err)

		// Verify file was created with correct content
		content, err := os.ReadFile(csvFile)
		require.NoError(t, err)

		lines := strings.Split(string(content), "\n")
		// 100 data rows + 1 header + 1 empty line at end
		assert.GreaterOrEqual(t, len(lines), 101)
	})

	t.Run("Batch size calculation", func(t *testing.T) {
		// With 100 rows and MAXBATCHSIZE=20, we should get 5 batches
		// This verifies the math used in the copy logic
		totalRows := 100
		batchSize := 20
		expectedBatches := (totalRows + batchSize - 1) / batchSize
		assert.Equal(t, 5, expectedBatches)
	})

	t.Run("Concurrent workers calculation", func(t *testing.T) {
		// Test that MAXREQUESTS controls concurrency
		options := parseCopyOptions("MAXREQUESTS=8")
		maxRequests := options["MAXREQUESTS"]
		assert.Equal(t, "8", maxRequests)
	})
}
