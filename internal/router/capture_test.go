package router

import (
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

func TestCaptureParquet(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a mock session and session manager
	mockSession := &db.Session{}
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)

	// Create handler
	handler := NewMetaCommandHandler(mockSession, sessionMgr)

	t.Run("capture parquet file", func(t *testing.T) {
		// Start capture to parquet file
		outputPath := filepath.Join(tempDir, "test.parquet")
		command := "CAPTURE PARQUET '" + outputPath + "'"
		result := handler.handleCapture(command)

		// Check success message
		assert.Contains(t, result, "Now capturing")
		assert.Contains(t, result, "parquet")
		assert.Equal(t, "parquet", handler.captureFormat)
		assert.NotNil(t, handler.captureOutput)

		// Write some test data with column types
		// Simulate headers with (PK) and (C) suffixes as they come from the executor
		headers := []string{"id (PK)", "name", "age", "active (C)"}
		columnTypes := []string{"int", "text", "int", "boolean"}
		rows := [][]string{
			{"1", "Alice", "30", "true"},
			{"2", "Bob", "25", "false"},
			{"3", "Charlie", "35", "true"},
		}

		err := handler.WriteCaptureResultWithTypes("SELECT * FROM users", headers, columnTypes, rows, nil)
		assert.NoError(t, err)
		assert.NotNil(t, handler.parquetWriter)

		// Stop capture
		result = handler.handleCapture("CAPTURE OFF")
		assert.Contains(t, result, "Stopped capturing")
		assert.Nil(t, handler.captureOutput)
		assert.Nil(t, handler.parquetWriter)

		// Verify file was created
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})

	t.Run("capture with auto extension", func(t *testing.T) {
		// Start capture without .parquet extension
		outputPath := filepath.Join(tempDir, "test2")
		command := "CAPTURE PARQUET '" + outputPath + "'"
		result := handler.handleCapture(command)

		// Should add .parquet extension
		expectedPath := outputPath + ".parquet"
		assert.Contains(t, result, expectedPath)
		assert.Equal(t, expectedPath, handler.captureFile)

		// Stop capture
		handler.handleCapture("CAPTURE OFF")
	})

	t.Run("append rows to parquet", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "test3.parquet")
		command := "CAPTURE PARQUET '" + outputPath + "'"
		_ = handler.handleCapture(command)

		// Write initial data
		headers := []string{"id", "value"}
		columnTypes := []string{"int", "double"}
		rows1 := [][]string{
			{"1", "1.5"},
			{"2", "2.5"},
		}

		err := handler.WriteCaptureResultWithTypes("SELECT * FROM data", headers, columnTypes, rows1, nil)
		require.NoError(t, err)

		// Append more rows
		rows2 := [][]string{
			{"3", "3.5"},
			{"4", "4.5"},
		}
		err = handler.AppendCaptureRows(rows2)
		assert.NoError(t, err)

		// Stop capture
		handler.handleCapture("CAPTURE OFF")

		// Verify file exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)
	})

	t.Run("capture with raw data", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "test4.parquet")
		command := "CAPTURE PARQUET '" + outputPath + "'"
		_ = handler.handleCapture(command)

		// Write data with raw values (preserves types)
		headers := []string{"id", "name", "score"}
		columnTypes := []string{"int", "text", "float"}
		rows := [][]string{
			{"1", "Test", "95.5"},
		}
		rawData := []map[string]interface{}{
			{"id": int32(1), "name": "Test", "score": float32(95.5)},
		}

		err := handler.WriteCaptureResultWithTypes("SELECT * FROM scores", headers, columnTypes, rows, rawData)
		assert.NoError(t, err)

		// Stop capture
		handler.handleCapture("CAPTURE OFF")
	})

	t.Run("capture status", func(t *testing.T) {
		// Check status when not capturing
		result := handler.handleCapture("CAPTURE")
		assert.Equal(t, "Not currently capturing output", result)

		// Start capture
		outputPath := filepath.Join(tempDir, "test5.parquet")
		handler.handleCapture("CAPTURE PARQUET '" + outputPath + "'")

		// Check status when capturing
		result = handler.handleCapture("CAPTURE")
		assert.Contains(t, result, "Currently capturing to")
		assert.Contains(t, result, "parquet")

		// Stop capture
		handler.handleCapture("CAPTURE OFF")
	})

	t.Run("invalid capture format fallback", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "test6.txt")
		// If user doesn't specify format, it defaults to text
		command := "CAPTURE '" + outputPath + "'"
		result := handler.handleCapture(command)
		assert.Contains(t, result, "text")
		assert.Equal(t, "text", handler.captureFormat)

		handler.handleCapture("CAPTURE OFF")
	})
}

func TestCaptureFormats(t *testing.T) {
	// Test that all capture formats are recognized
	tempDir := t.TempDir()
	mockSession := &db.Session{}
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)
	handler := NewMetaCommandHandler(mockSession, sessionMgr)

	formats := []struct {
		command      string
		format       string
		extension    string
	}{
		{"CAPTURE JSON 'test'", "json", ".json"},
		{"CAPTURE CSV 'test'", "csv", ".csv"},
		{"CAPTURE PARQUET 'test'", "parquet", ".parquet"},
		{"CAPTURE 'test'", "text", ""},
	}

	for _, f := range formats {
		t.Run(f.format, func(t *testing.T) {
			// Use full path to avoid file conflicts
			outputPath := filepath.Join(tempDir, "test_" + f.format)
			command := strings.Replace(f.command, "'test'", "'"+outputPath+"'", 1)

			result := handler.handleCapture(command)
			assert.Contains(t, result, f.format)
			assert.Equal(t, f.format, handler.captureFormat)

			// Check extension was added if needed
			if f.extension != "" {
				assert.True(t, strings.HasSuffix(handler.captureFile, f.extension))
			}

			// Clean up
			handler.handleCapture("CAPTURE OFF")
		})
	}
}