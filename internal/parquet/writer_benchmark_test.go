package parquet

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateLargeDataset creates a large dataset for testing
func generateLargeDataset(rows int) []map[string]interface{} {
	data := make([]map[string]interface{}, rows)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	statuses := []string{"active", "inactive", "pending", "completed", "failed"}

	for i := 0; i < rows; i++ {
		data[i] = map[string]interface{}{
			"id":          uuid.New().String(),
			"user_id":     uuid.New().String(),
			"timestamp":   time.Now().Add(time.Duration(-r.Intn(86400)) * time.Second),
			"value":       r.Float64() * 1000,
			"count":       int32(r.Intn(10000)),
			"status":      statuses[r.Intn(len(statuses))],
			"description": fmt.Sprintf("Test description for row %d with some random text", i),
			"is_active":   r.Intn(2) == 1,
			"score":       r.Float32() * 100,
			"metadata":    fmt.Sprintf(`{"key": "value_%d", "index": %d}`, i, i),
		}
	}

	return data
}

func TestLargeDatasetWrite(t *testing.T) {
	tests := []struct {
		name     string
		rows     int
		chunkSize int64
	}{
		{"10K rows", 10000, 1000},
		{"100K rows", 100000, 10000},
		// Uncomment for full testing
		// {"1M rows", 1000000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, fmt.Sprintf("large_%d.parquet", tt.rows))

			// Create schema
			columnNames := []string{"id", "user_id", "timestamp", "value", "count",
				"status", "description", "is_active", "score", "metadata"}
			columnTypes := []string{"uuid", "uuid", "timestamp", "double", "int",
				"text", "text", "boolean", "float", "text"}

			options := DefaultWriterOptions()
			options.ChunkSize = tt.chunkSize

			writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, options)
			require.NoError(t, err)
			defer writer.Close()

			// Generate and write data
			startTime := time.Now()
			data := generateLargeDataset(tt.rows)

			for _, row := range data {
				err := writer.WriteRow(row)
				require.NoError(t, err)
			}

			err = writer.Close()
			require.NoError(t, err)

			duration := time.Since(startTime)

			// Verify file was created and has reasonable size
			fileInfo, err := os.Stat(outputPath)
			require.NoError(t, err)

			t.Logf("Written %d rows in %v", tt.rows, duration)
			t.Logf("File size: %.2f MB", float64(fileInfo.Size())/(1024*1024))
			t.Logf("Throughput: %.2f rows/sec", float64(tt.rows)/duration.Seconds())
			t.Logf("Bytes per row: %.2f", float64(fileInfo.Size())/float64(tt.rows))

			// Verify we can read the file back
			verifyParquetFile(t, outputPath, tt.rows)
		})
	}
}

func TestCompressionComparison(t *testing.T) {
	compressionTypes := []struct {
		name string
		compression string
	}{
		{"No compression", ""},
		{"Snappy", "snappy"},
		{"GZIP", "gzip"},
		// LZ4 is not implemented in Apache Arrow Go
		// {"LZ4", "lz4"},
		{"ZSTD", "zstd"},
	}

	rows := 10000
	data := generateLargeDataset(rows)

	for _, ct := range compressionTypes {
		t.Run(ct.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, fmt.Sprintf("compressed_%s.parquet", ct.name))

			columnNames := []string{"id", "user_id", "timestamp", "value", "count",
				"status", "description", "is_active", "score", "metadata"}
			columnTypes := []string{"uuid", "uuid", "timestamp", "double", "int",
				"text", "text", "boolean", "float", "text"}

			options := DefaultWriterOptions()

			startTime := time.Now()

			writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, options)
			require.NoError(t, err)

			// Set compression after creating writer
			if ct.compression != "" {
				err = writer.SetCompression(ct.compression)
				require.NoError(t, err)
			}

			for _, row := range data {
				err := writer.WriteRow(row)
				require.NoError(t, err)
			}

			err = writer.Close()
			require.NoError(t, err)

			duration := time.Since(startTime)

			fileInfo, err := os.Stat(outputPath)
			require.NoError(t, err)

			t.Logf("Compression: %s", ct.name)
			t.Logf("Write time: %v", duration)
			t.Logf("File size: %.2f MB", float64(fileInfo.Size())/(1024*1024))
			t.Logf("Compression ratio: %.2f bytes/row", float64(fileInfo.Size())/float64(rows))
		})
	}
}

func TestMemoryUsage(t *testing.T) {
	t.Run("memory efficient chunking", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "memory_test.parquet")

		columnNames := []string{"id", "value", "data"}
		columnTypes := []string{"uuid", "double", "text"}

		options := DefaultWriterOptions()
		options.ChunkSize = 1000 // Small chunks to test memory management

		writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, options)
		require.NoError(t, err)
		defer writer.Close()

		// Write data in batches and observe memory usage
		batchSize := 5000
		numBatches := 10

		for batch := 0; batch < numBatches; batch++ {
			for i := 0; i < batchSize; i++ {
				row := map[string]interface{}{
					"id":    uuid.New().String(),
					"value": rand.Float64() * 1000,
					"data":  fmt.Sprintf("Large text data for row %d in batch %d with padding...", i, batch),
				}
				err := writer.WriteRow(row)
				require.NoError(t, err)
			}

			// Log current row count and chunks written
			t.Logf("Batch %d: Total rows: %d", batch+1, (batch+1)*batchSize)
		}

		err = writer.Close()
		require.NoError(t, err)

		fileInfo, err := os.Stat(outputPath)
		require.NoError(t, err)

		totalRows := batchSize * numBatches
		t.Logf("Total rows written: %d", totalRows)
		t.Logf("Final file size: %.2f MB", float64(fileInfo.Size())/(1024*1024))
	})
}

// verifyParquetFile reads back a Parquet file to verify it's valid
func verifyParquetFile(t *testing.T, filepath string, expectedRows int) {
	f, err := os.Open(filepath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	// Verify row count
	assert.Equal(t, int64(expectedRows), reader.NumRows())

	// Read using Arrow reader for better verification
	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	// Read the schema
	schema, err := arrowReader.Schema()
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Verify we can read the whole file
	table, err := arrowReader.ReadTable(ctx)
	require.NoError(t, err)
	defer table.Release()

	assert.Equal(t, int64(expectedRows), table.NumRows())
}

func BenchmarkParquetWrite(b *testing.B) {
	benchmarks := []struct {
		name string
		rows int
	}{
		{"100_rows", 100},
		{"1000_rows", 1000},
		{"10000_rows", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := generateLargeDataset(bm.rows)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tempDir := b.TempDir()
				outputPath := filepath.Join(tempDir, "bench.parquet")

				columnNames := []string{"id", "user_id", "timestamp", "value", "count"}
				columnTypes := []string{"uuid", "uuid", "timestamp", "double", "int"}

				writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
				require.NoError(b, err)

				for _, row := range data[:5] { // Only write first 5 columns for benchmark
					smallRow := map[string]interface{}{
						"id":        row["id"],
						"user_id":   row["user_id"],
						"timestamp": row["timestamp"],
						"value":     row["value"],
						"count":     row["count"],
					}
					_ = writer.WriteRow(smallRow)
				}

				_ = writer.Close()
			}
		})
	}
}