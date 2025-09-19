package parquet_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleTypes(t *testing.T) {
	tmpFile := "/tmp/test_simple_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Write and Read Basic Types", func(t *testing.T) {
		// Define schema with all basic Cassandra types
		columns := []string{
			"id", "name", "age", "salary", "active",
			"created_date", "created_time", "score",
			"data", "ip_address", "user_id",
		}
		types := []string{
			"int", "text", "int", "double", "boolean",
			"date", "timestamp", "float",
			"blob", "inet", "uuid",
		}

		// Create writer
		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write test data
		testData := []map[string]interface{}{
			{
				"id":           1,
				"name":         "Alice",
				"age":          30,
				"salary":       75000.50,
				"active":       true,
				"created_date": time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				"created_time": time.Now(),
				"score":        float32(4.5),
				"data":         []byte("binary data"),
				"ip_address":   "192.168.1.1",
				"user_id":      "550e8400-e29b-41d4-a716-446655440000",
			},
			{
				"id":           2,
				"name":         "Bob",
				"age":          25,
				"salary":       65000.00,
				"active":       false,
				"created_date": time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
				"created_time": time.Now().Add(-24 * time.Hour),
				"score":        float32(3.8),
				"data":         []byte("more data"),
				"ip_address":   "10.0.0.1",
				"user_id":      "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		// Verify schema
		readColumns, _ := reader.GetSchema()
		assert.Equal(t, columns, readColumns)

		// Verify row count
		assert.Equal(t, int64(2), reader.GetRowCount())

		// Read all data
		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 2)

		// Verify first row
		assert.Equal(t, int32(1), rows[0]["id"])
		assert.Equal(t, "Alice", rows[0]["name"])
		assert.Equal(t, int32(30), rows[0]["age"])
		assert.Equal(t, true, rows[0]["active"])
	})
}

func TestNullHandling(t *testing.T) {
	tmpFile := "/tmp/test_null_handling.parquet"
	defer os.Remove(tmpFile)

	t.Run("Write and Read Null Values", func(t *testing.T) {
		columns := []string{"id", "optional_text", "optional_int"}
		types := []string{"int", "text", "int"}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Write rows with null values
		testData := []map[string]interface{}{
			{"id": 1, "optional_text": "present", "optional_int": 100},
			{"id": 2, "optional_text": nil, "optional_int": 200},
			{"id": 3, "optional_text": "another", "optional_int": nil},
			{"id": 4, "optional_text": nil, "optional_int": nil},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify nulls are preserved
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 4)

		// Verify null handling
		assert.Equal(t, "present", rows[0]["optional_text"])
		assert.Nil(t, rows[1]["optional_text"])
		assert.Equal(t, int32(200), rows[1]["optional_int"])
		assert.Nil(t, rows[2]["optional_int"])
		assert.Nil(t, rows[3]["optional_text"])
		assert.Nil(t, rows[3]["optional_int"])
	})
}

func TestLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tmpFile := "/tmp/test_large_dataset.parquet"
	defer os.Remove(tmpFile)

	t.Run("Write and Read 100K Rows", func(t *testing.T) {
		columns := []string{"id", "value", "timestamp"}
		types := []string{"bigint", "double", "timestamp"}

		options := parquet.DefaultWriterOptions()
		options.ChunkSize = 10000 // Use 10K chunk size for performance

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, options)
		require.NoError(t, err)

		// Write 100K rows
		rowCount := 100000
		startTime := time.Now()

		for i := 0; i < rowCount; i++ {
			row := map[string]interface{}{
				"id":        int64(i),
				"value":     float64(i) * 1.5,
				"timestamp": time.Now(),
			}
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		writeTime := time.Since(startTime)
		t.Logf("Wrote %d rows in %v (%.0f rows/sec)", rowCount, writeTime, float64(rowCount)/writeTime.Seconds())

		// Check file size
		fileInfo, err := os.Stat(tmpFile)
		require.NoError(t, err)
		t.Logf("File size: %.2f MB", float64(fileInfo.Size())/(1024*1024))

		// Read back sample
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, int64(rowCount), reader.GetRowCount())

		// Read first batch
		batch, err := reader.ReadBatch(100)
		require.NoError(t, err)
		assert.Len(t, batch, 100)

		// Verify first and last in batch
		assert.Equal(t, int64(0), batch[0]["id"])
		assert.Equal(t, int64(99), batch[99]["id"])
	})
}

func TestCompressionFormats(t *testing.T) {
	compressionTypes := []string{"SNAPPY", "GZIP", "ZSTD"}

	for _, compression := range compressionTypes {
		t.Run(fmt.Sprintf("Compression_%s", compression), func(t *testing.T) {
			tmpFile := fmt.Sprintf("/tmp/test_compression_%s.parquet", compression)
			defer os.Remove(tmpFile)

			columns := []string{"id", "data"}
			types := []string{"int", "text"}

			writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
			require.NoError(t, err)

			// Set compression
			err = writer.SetCompression(compression)
			require.NoError(t, err)

			// Write test data with repetitive patterns (good for compression)
			for i := 0; i < 1000; i++ {
				row := map[string]interface{}{
					"id":   i,
					"data": fmt.Sprintf("This is row %d with some repetitive text pattern", i%10),
				}
				err = writer.WriteRow(row)
				require.NoError(t, err)
			}

			err = writer.Close()
			require.NoError(t, err)

			// Check file size
			fileInfo, err := os.Stat(tmpFile)
			require.NoError(t, err)
			t.Logf("%s compressed file size: %.2f KB", compression, float64(fileInfo.Size())/1024)

			// Verify data integrity
			reader, err := parquet.NewParquetReader(tmpFile)
			require.NoError(t, err)
			defer reader.Close()

			rows, err := reader.ReadBatch(10)
			require.NoError(t, err)
			assert.Len(t, rows, 10)
			assert.Equal(t, int32(0), rows[0]["id"])
		})
	}
}