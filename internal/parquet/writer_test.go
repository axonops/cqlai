package parquet

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParquetCaptureWriter(t *testing.T) {
	t.Run("create writer with stdout", func(t *testing.T) {
		columnNames := []string{"id", "name", "age"}
		columnTypes := []string{"int", "text", "int"}

		writer, err := NewParquetCaptureWriter("-", columnNames, columnTypes, DefaultWriterOptions())
		require.NoError(t, err)
		require.NotNil(t, writer)

		assert.Equal(t, os.Stdout, writer.writer)
		assert.NotNil(t, writer.schema)
		assert.NotNil(t, writer.builder)
		assert.Equal(t, int64(10000), writer.chunkSize)
	})

	t.Run("create writer with file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "test.parquet")

		columnNames := []string{"id", "name"}
		columnTypes := []string{"uuid", "text"}

		writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
		require.NoError(t, err)
		require.NotNil(t, writer)

		defer writer.Close()

		assert.NotNil(t, writer.schema)
		assert.Equal(t, 2, writer.schema.NumFields())
	})

	t.Run("schema mismatch", func(t *testing.T) {
		columnNames := []string{"id", "name", "age"}
		columnTypes := []string{"int", "text"} // Mismatch: fewer types than names

		writer, err := NewParquetCaptureWriter("-", columnNames, columnTypes, DefaultWriterOptions())
		assert.Error(t, err)
		assert.Nil(t, writer)
	})
}

func TestWriteRow(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_write.parquet")

	columnNames := []string{"id", "name", "age", "active"}
	columnTypes := []string{"int", "text", "int", "boolean"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)
	defer writer.Close()

	t.Run("write single row", func(t *testing.T) {
		row := map[string]interface{}{
			"id":     int32(1),
			"name":   "Alice",
			"age":    int32(30),
			"active": true,
		}

		err := writer.WriteRow(row)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), writer.rowCount)
		assert.Equal(t, int64(1), writer.totalRows)
	})

	t.Run("write multiple rows", func(t *testing.T) {
		rows := []map[string]interface{}{
			{"id": int32(2), "name": "Bob", "age": int32(25), "active": false},
			{"id": int32(3), "name": "Charlie", "age": int32(35), "active": true},
		}

		err := writer.WriteRows(rows)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), writer.totalRows)
	})

	t.Run("write with nil values", func(t *testing.T) {
		row := map[string]interface{}{
			"id":     int32(4),
			"name":   nil, // NULL value
			"age":    int32(40),
			"active": nil, // NULL value
		}

		err := writer.WriteRow(row)
		assert.NoError(t, err)
		assert.Equal(t, int64(4), writer.totalRows)
	})
}

func TestChunking(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_chunks.parquet")

	columnNames := []string{"id", "value"}
	columnTypes := []string{"int", "double"}

	options := WriterOptions{
		ChunkSize:   5, // Small chunk size for testing
		Compression: compress.Codecs.Snappy,
	}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, options)
	require.NoError(t, err)

	// Write 12 rows (should create 3 chunks)
	for i := 1; i <= 12; i++ {
		row := map[string]interface{}{
			"id":    int32(i),
			"value": float64(i) * 1.5,
		}
		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	// Check total rows
	assert.Equal(t, int64(12), writer.totalRows)

	// Close and verify file was written
	err = writer.Close()
	require.NoError(t, err)

	// Read back the file to verify
	// Create a parquet file reader
	pqReader, err := file.OpenParquetFile(outputPath, false)
	require.NoError(t, err)
	defer pqReader.Close()

	reader, err := pqarrow.NewFileReader(pqReader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	table, err := reader.ReadTable(context.Background())
	require.NoError(t, err)
	defer table.Release()

	assert.Equal(t, int64(12), table.NumRows())
	assert.Equal(t, 2, int(table.NumCols()))
}

func TestWriteRawRows(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_raw.parquet")

	columnNames := []string{"id", "name", "score"}
	columnTypes := []string{"int", "text", "float"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)
	defer writer.Close()

	headers := []string{"id", "name", "score"}
	rows := [][]interface{}{
		{int32(1), "Alice", float32(95.5)},
		{int32(2), "Bob", float32(87.3)},
		{int32(3), "Charlie", float32(92.8)},
	}

	err = writer.WriteRawRows(headers, rows)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), writer.totalRows)
}

func TestWriteStringRows(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_string.parquet")

	columnNames := []string{"id", "name", "value"}
	columnTypes := []string{"text", "text", "text"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)
	defer writer.Close()

	headers := []string{"id", "name", "value"}
	rows := [][]string{
		{"1", "Alice", "100"},
		{"2", "Bob", "200"},
		{"3", "Charlie", "300"},
	}

	err = writer.WriteStringRows(headers, rows)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), writer.totalRows)
}

func TestCompression(t *testing.T) {
	tests := []struct {
		compression string
		shouldError bool
	}{
		{"snappy", false},
		{"gzip", false},
		{"lz4", false},
		{"zstd", false},
		{"none", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.compression, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "test_compression.parquet")

			columnNames := []string{"id"}
			columnTypes := []string{"int"}

			writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
			require.NoError(t, err)
			defer writer.Close()

			err = writer.SetCompression(tt.compression)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCloseWithoutData(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_empty.parquet")

	columnNames := []string{"id", "name"}
	columnTypes := []string{"int", "text"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)

	// Close without writing any data
	err = writer.Close()
	assert.NoError(t, err)

	// File should exist but be empty or minimal
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.NotNil(t, info)
}

func TestWriterMethods(t *testing.T) {
	t.Run("GetRowCount", func(t *testing.T) {
		writer, err := NewParquetCaptureWriter("-", []string{"id"}, []string{"int"}, DefaultWriterOptions())
		require.NoError(t, err)
		defer writer.Close()

		assert.Equal(t, int64(0), writer.GetRowCount())

		err = writer.WriteRow(map[string]interface{}{"id": int32(1)})
		require.NoError(t, err)
		assert.Equal(t, int64(1), writer.GetRowCount())
	})

	t.Run("IsStreaming", func(t *testing.T) {
		writer, err := NewParquetCaptureWriter("-", []string{"id"}, []string{"int"}, DefaultWriterOptions())
		require.NoError(t, err)
		defer writer.Close()

		assert.True(t, writer.IsStreaming())
	})

	t.Run("WriteHeader", func(t *testing.T) {
		writer, err := NewParquetCaptureWriter("-", []string{"id"}, []string{"int"}, DefaultWriterOptions())
		require.NoError(t, err)
		defer writer.Close()

		err = writer.WriteHeader()
		assert.NoError(t, err) // Should be a no-op
	})

	t.Run("Flush", func(t *testing.T) {
		// Use a temp file instead of stdout to avoid issues
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "test_flush.parquet")

		writer, err := NewParquetCaptureWriter(outputPath, []string{"id"}, []string{"int"}, DefaultWriterOptions())
		require.NoError(t, err)
		defer writer.Close()

		err = writer.Flush()
		assert.NoError(t, err)
	})
}

func TestWriterAfterClose(t *testing.T) {
	writer, err := NewParquetCaptureWriter("-", []string{"id"}, []string{"int"}, DefaultWriterOptions())
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// Try to write after closing
	err = writer.WriteRow(map[string]interface{}{"id": int32(1)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writer is closed")
}

func TestWriterCollectionTypes(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_collections.parquet")

	columnNames := []string{"id", "tags", "scores"}
	columnTypes := []string{"int", "list<text>", "list<int>"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)
	defer writer.Close()

	row := map[string]interface{}{
		"id":     int32(1),
		"tags":   []string{"tag1", "tag2", "tag3"},
		"scores": []int32{100, 200, 300},
	}

	err = writer.WriteRow(row)
	assert.NoError(t, err)

	// Close and read back
	err = writer.Close()
	require.NoError(t, err)

	// Verify the file was created
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)
}

func TestBufferWriter(t *testing.T) {
	// Test writing to a buffer instead of file
	buf := &bytes.Buffer{}

	columnNames := []string{"id", "value"}
	columnTypes := []string{"int", "text"}

	// Create writer with buffer by modifying the writer field after creation
	writer, err := NewParquetCaptureWriter("-", columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)
	writer.writer = buf // Replace stdout with buffer

	// Write some data
	for i := 1; i <= 5; i++ {
		row := map[string]interface{}{
			"id":    int32(i),
			"value": "test",
		}
		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Buffer should contain data
	assert.Greater(t, buf.Len(), 0)
}

func BenchmarkWriteRow(b *testing.B) {
	tempDir := b.TempDir()
	outputPath := filepath.Join(tempDir, "bench.parquet")

	columnNames := []string{"id", "name", "value", "timestamp"}
	columnTypes := []string{"int", "text", "double", "timestamp"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(b, err)
	defer writer.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		row := map[string]interface{}{
			"id":        int32(i),
			"name":      "benchmark",
			"value":     float64(i) * 1.5,
			"timestamp": arrow.Timestamp(i * 1000),
		}
		_ = writer.WriteRow(row)
	}
}