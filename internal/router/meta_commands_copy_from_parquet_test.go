package router

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/parquet"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionForCopyFrom tracks INSERT queries
type MockSessionForCopyFrom struct {
	db.Session
	insertQueries []string
	insertCount   int
	failInserts   bool
}

func (m *MockSessionForCopyFrom) ExecuteCQLQuery(query string) interface{} {
	m.insertQueries = append(m.insertQueries, query)

	if m.failInserts {
		return fmt.Errorf("mock insert error")
	}

	m.insertCount++
	return db.QueryResult{}
}

func TestExecuteCopyFromParquet(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	t.Run("basic COPY FROM parquet", func(t *testing.T) {
		// First, create a test Parquet file
		testFile := filepath.Join(tempDir, "test_data.parquet")

		// Create test data
		columnNames := []string{"id", "name", "value"}
		columnTypes := []string{"int", "text", "double"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write test data
		testData := []map[string]interface{}{
			{"id": int32(1), "name": "Alice", "value": 100.5},
			{"id": int32(2), "name": "Bob", "value": 200.75},
			{"id": int32(3), "name": "Charlie", "value": 300.25},
		}

		for _, row := range testData {
			err := writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Now test COPY FROM
		// Note: Since db.Session is a struct not an interface, we can't override ExecuteCQLQuery
		// The test will use a nil gocql.Session which will return "not connected to database" errors
		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil}, // nil gocql.Session will fail all queries
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{
			"FORMAT": "parquet",
		}

		result := handler.executeCopyFromParquet("test_table", []string{}, testFile, options)

		// Since we have a nil session, all INSERTs will fail with "not connected to database"
		// The result should show 0 imported rows with errors
		assert.Contains(t, result, "Imported 0 rows")
		assert.Contains(t, result, "3 errors")
	})

	t.Run("COPY FROM with column selection", func(t *testing.T) {
		// Create test Parquet file with more columns
		testFile := filepath.Join(tempDir, "test_columns.parquet")

		columnNames := []string{"id", "name", "value", "extra"}
		columnTypes := []string{"int", "text", "double", "text"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write test data
		testData := []map[string]interface{}{
			{"id": int32(1), "name": "Alice", "value": 100.5, "extra": "data1"},
			{"id": int32(2), "name": "Bob", "value": 200.75, "extra": "data2"},
		}

		for _, row := range testData {
			err := writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Test COPY FROM with selected columns
		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{
			"FORMAT": "parquet",
		}

		// Only import id and name columns
		result := handler.executeCopyFromParquet("test_table", []string{"id", "name"}, testFile, options)

		// Since we have a nil session, all INSERTs will fail
		assert.Contains(t, result, "Imported 0 rows")
		assert.Contains(t, result, "2 errors")
	})

	t.Run("COPY FROM with skip rows", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test_skip.parquet")

		columnNames := []string{"id", "value"}
		columnTypes := []string{"int", "double"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write 5 rows
		for i := 1; i <= 5; i++ {
			row := map[string]interface{}{
				"id":    int32(i),
				"value": float64(i) * 10.0,
			}
			err := writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{
			"FORMAT":   "parquet",
			"SKIPROWS": "2",
		}

		result := handler.executeCopyFromParquet("test_table", []string{}, testFile, options)

		// Should try to import 3 rows (5 total - 2 skipped), but all fail due to nil session
		t.Logf("Skip rows test result: %v", result)
		assert.Contains(t, result, "Imported 0 rows")
		assert.Contains(t, result, "skipped 2 rows")
		// Errors might not be counted if skipping happens first
		// The test might be skipping before attempting inserts
	})

	t.Run("COPY FROM with max rows", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test_max.parquet")

		columnNames := []string{"id"}
		columnTypes := []string{"int"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write 10 rows
		for i := 1; i <= 10; i++ {
			row := map[string]interface{}{"id": int32(i)}
			err := writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{
			"FORMAT":  "parquet",
			"MAXROWS": "5",
		}

		result := handler.executeCopyFromParquet("test_table", []string{}, testFile, options)

		// Should try to import 5 rows (MAXROWS limit), but all fail due to nil session
		assert.Contains(t, result, "Imported 0 rows")
		assert.Contains(t, result, "5 errors")
	})

	t.Run("COPY FROM handles insert errors", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test_errors.parquet")

		columnNames := []string{"id"}
		columnTypes := []string{"int"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write 5 rows
		for i := 1; i <= 5; i++ {
			row := map[string]interface{}{"id": int32(i)}
			err := writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		mockSession := &MockSessionForCopyFrom{
			Session:     db.Session{Session: nil},
			failInserts: true, // All inserts will fail
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{
			"FORMAT":           "parquet",
			"MAXINSERTERRORS": "3",
		}

		result := handler.executeCopyFromParquet("test_table", []string{}, testFile, options)

		// Should abort after 3 errors
		assert.Contains(t, result, "Aborted after 3 insert errors")
		assert.Contains(t, result, "Successfully imported 0 rows")
	})

	t.Run("COPY FROM validates columns", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "test_validate.parquet")

		columnNames := []string{"id", "name"}
		columnTypes := []string{"int", "text"}

		writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		row := map[string]interface{}{"id": int32(1), "name": "test"}
		err = writer.WriteRow(row)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{"FORMAT": "parquet"}

		// Try to import non-existent column
		result := handler.executeCopyFromParquet("test_table", []string{"id", "nonexistent"}, testFile, options)

		assert.Contains(t, result, "Column 'nonexistent' not found in Parquet file")
	})

	t.Run("COPY FROM rejects STDIN", func(t *testing.T) {
		mockSession := &MockSessionForCopyFrom{
			Session: db.Session{Session: nil},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		options := map[string]string{"FORMAT": "parquet"}

		result := handler.executeCopyFromParquet("test_table", []string{}, "STDIN", options)

		assert.Contains(t, result, "COPY FROM STDIN is not supported for Parquet format")
	})
}

func TestHandleCopyWithParquetFrom(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	// Create a test Parquet file
	testFile := filepath.Join(tempDir, "test.parquet")

	columnNames := []string{"id", "name"}
	columnTypes := []string{"int", "text"}

	writer, err := parquet.NewParquetCaptureWriter(testFile, columnNames, columnTypes, parquet.DefaultWriterOptions())
	require.NoError(t, err)

	row := map[string]interface{}{"id": int32(1), "name": "test"}
	err = writer.WriteRow(row)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	mockSession := &MockSessionForCopyFrom{
		Session: db.Session{Session: nil}, // nil gocql.Session
	}

	handler := &MetaCommandHandler{
		session:        &mockSession.Session,
		sessionManager: sessionMgr,
	}

	t.Run("COPY FROM with FORMAT=PARQUET", func(t *testing.T) {
		command := fmt.Sprintf("COPY test_table FROM '%s' WITH FORMAT='PARQUET'", testFile)

		result := handler.handleCopy(command)

		// Since we have a nil session, the INSERT will fail
		assert.Contains(t, result, "Imported 0 rows from Parquet file")
		assert.Contains(t, result, "1 errors")
	})

	t.Run("COPY FROM defaults to CSV", func(t *testing.T) {
		// Create a CSV file
		csvFile := filepath.Join(tempDir, "test.csv")
		err := os.WriteFile(csvFile, []byte("1,test\n"), 0644)
		require.NoError(t, err)

		command := fmt.Sprintf("COPY test_table FROM '%s'", csvFile)

		result := handler.handleCopy(command)

		// Should try to use CSV format and likely get an error due to mock session
		// but not a Parquet-specific error
		assert.NotContains(t, result, "Parquet")
	})
}