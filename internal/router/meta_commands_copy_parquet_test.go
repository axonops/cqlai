package router

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionForCopy embeds db.Session and overrides ExecuteCQLQuery
type MockSessionForCopy struct {
	db.Session
	queryResult interface{}
}

func (m *MockSessionForCopy) ExecuteCQLQuery(query string) interface{} {
	return m.queryResult
}

func TestExecuteCopyToParquet(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	t.Run("basic COPY TO parquet", func(t *testing.T) {
		// Create mock session with test data
		mockSession := &MockSessionForCopy{
			queryResult: db.QueryResult{
				Headers:     []string{"id (PK)", "name", "value"},
				ColumnTypes: []string{"int", "text", "double"},
				Data: [][]string{
					{"1", "Alice", "100.5"},
					{"2", "Bob", "200.75"},
					{"3", "Charlie", "300.25"},
				},
				RawData: []map[string]interface{}{
					{"id (PK)": int32(1), "name": "Alice", "value": 100.5},
					{"id (PK)": int32(2), "name": "Bob", "value": 200.75},
					{"id (PK)": int32(3), "name": "Charlie", "value": 300.25},
				},
			},
		}

		// Create a handler with mock session that implements the interface
		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		outputPath := filepath.Join(tempDir, "test_copy")
		options := map[string]string{
			"FORMAT": "parquet",
		}

		result := handler.executeCopyToParquet("users", []string{}, outputPath, options)

		// Check result type and message
		t.Logf("Result: %v (type: %T)", result, result)

		// Check success message
		resultStr := fmt.Sprintf("%v", result)
		assert.Contains(t, resultStr, "Exported 3 rows")
		assert.Contains(t, resultStr, "test_copy.parquet")
		assert.Contains(t, resultStr, "(Parquet format)")

		// Verify the file was created with .parquet extension
		expectedPath := outputPath + ".parquet"
		fileInfo, err := os.Stat(expectedPath)
		require.NoError(t, err)
		assert.True(t, fileInfo.Size() > 0)

		// Verify the Parquet file is valid and contains correct data
		verifyParquetFile(t, expectedPath, 3, []string{"id", "name", "value"})
	})

	t.Run("COPY TO with column selection", func(t *testing.T) {
		mockSession := &MockSessionForCopy{
			queryResult: db.QueryResult{
				Headers:     []string{"id (PK)", "name"},
				ColumnTypes: []string{"int", "text"},
				Data: [][]string{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		outputPath := filepath.Join(tempDir, "test_columns.parquet")
		options := map[string]string{
			"FORMAT": "parquet",
		}

		result := handler.executeCopyToParquet("users", []string{"id", "name"}, outputPath, options)

		assert.Contains(t, result, "Exported 2 rows")

		// Verify the file
		fileInfo, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.True(t, fileInfo.Size() > 0)
	})

	t.Run("COPY TO with compression", func(t *testing.T) {
		mockSession := &MockSessionForCopy{
			queryResult: db.QueryResult{
				Headers:     []string{"id", "data"},
				ColumnTypes: []string{"int", "text"},
				Data: [][]string{
					{"1", "Test data for compression"},
					{"2", "More test data"},
					{"3", "Even more data"},
				},
			},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		compressionTypes := []string{"snappy", "gzip", "zstd"}

		for _, compression := range compressionTypes {
			t.Run(compression, func(t *testing.T) {
				outputPath := filepath.Join(tempDir, "test_"+compression+".parquet")
				options := map[string]string{
					"FORMAT":      "parquet",
					"COMPRESSION": compression,
				}

				result := handler.executeCopyToParquet("data_table", []string{}, outputPath, options)

				assert.Contains(t, result, "Exported 3 rows")

				// Verify file exists and has data
				fileInfo, err := os.Stat(outputPath)
				require.NoError(t, err)
				assert.True(t, fileInfo.Size() > 0)

				// Verify it's a valid Parquet file
				verifyParquetFile(t, outputPath, 3, []string{"id", "data"})
			})
		}
	})

	t.Run("COPY TO handles no data gracefully", func(t *testing.T) {
		mockSession := &MockSessionForCopy{
			queryResult: db.QueryResult{
				Headers: []string{},
				Data:    [][]string{},
			},
		}

		handler := &MetaCommandHandler{
			session:        &mockSession.Session,
			sessionManager: sessionMgr,
		}

		outputPath := filepath.Join(tempDir, "test_empty.parquet")
		options := map[string]string{
			"FORMAT": "parquet",
		}

		result := handler.executeCopyToParquet("empty_table", []string{}, outputPath, options)

		assert.Equal(t, "No data to export", result)

		// File should not be created
		_, err := os.Stat(outputPath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestHandleCopyWithParquet(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	mockSession := &MockSessionForCopy{
		queryResult: db.QueryResult{
			Headers:     []string{"id", "name"},
			ColumnTypes: []string{"int", "text"},
			Data: [][]string{
				{"1", "Test"},
			},
		},
	}

	handler := &MetaCommandHandler{
		session:        &mockSession.Session,
		sessionManager: sessionMgr,
	}

	t.Run("COPY TO with FORMAT=PARQUET", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "format_test")
		command := "COPY users TO '" + outputPath + "' WITH FORMAT='PARQUET'"

		result := handler.handleCopy(command)

		assert.Contains(t, result, "Exported 1 rows")
		assert.Contains(t, result, "(Parquet format)")

		// Should add .parquet extension
		expectedPath := outputPath + ".parquet"
		_, err := os.Stat(expectedPath)
		require.NoError(t, err)
	})

	t.Run("COPY TO with FORMAT=CSV (default)", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "csv_test.csv")
		command := "COPY users TO '" + outputPath + "'"

		result := handler.handleCopy(command)

		// Should use CSV format by default
		assert.Contains(t, result, "Exported 1 rows")
		assert.NotContains(t, result, "Parquet")

		_, err := os.Stat(outputPath)
		require.NoError(t, err)
	})
}

// verifyParquetFile helper function to validate Parquet files
func verifyParquetFile(t *testing.T, filepath string, expectedRows int, expectedColumns []string) {
	f, err := os.Open(filepath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	// Verify row count
	assert.Equal(t, int64(expectedRows), reader.NumRows())

	// Read using Arrow reader to verify schema
	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	schema, err := arrowReader.Schema()
	require.NoError(t, err)

	// Verify column count and names
	assert.Equal(t, len(expectedColumns), schema.NumFields())

	for i, expectedName := range expectedColumns {
		field := schema.Field(i)
		assert.Equal(t, expectedName, field.Name, "Column %d name mismatch", i)
	}

	// Read the data to ensure it's valid
	table, err := arrowReader.ReadTable(ctx)
	require.NoError(t, err)
	defer table.Release()

	assert.Equal(t, int64(expectedRows), table.NumRows())
}