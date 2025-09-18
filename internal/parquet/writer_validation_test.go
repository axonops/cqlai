package parquet

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAllCassandraTypes tests writing and reading all Cassandra data types
func TestAllCassandraTypes(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "all_types.parquet")

	// Define all Cassandra types
	columnNames := []string{
		"id", "text_col", "int_col", "bigint_col", "float_col", "double_col",
		"boolean_col", "uuid_col", "timeuuid_col", "timestamp_col", "date_col",
		"time_col", "inet_col", "blob_col", "list_col", "set_col", "map_col",
	}

	columnTypes := []string{
		"uuid", "text", "int", "bigint", "float", "double",
		"boolean", "uuid", "timeuuid", "timestamp", "date",
		"time", "inet", "blob", "list<text>", "set<int>", "map<text, int>",
	}

	// Create writer
	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)

	// Write test data
	testData := []map[string]interface{}{
		{
			"id":           uuid.New().String(),
			"text_col":     "Test text",
			"int_col":      int32(42),
			"bigint_col":   int64(9223372036854775807),
			"float_col":    float32(3.14),
			"double_col":   float64(2.71828),
			"boolean_col":  true,
			"uuid_col":     uuid.New().String(),
			"timeuuid_col": uuid.New().String(),
			"timestamp_col": time.Now(),
			"date_col":     time.Now().Truncate(24 * time.Hour),
			"time_col":     int64(time.Hour + 30*time.Minute + 45*time.Second),
			"inet_col":     "192.168.1.1",
			"blob_col":     []byte("binary data"),
			"list_col":     []string{"item1", "item2", "item3"},
			"set_col":      []int32{1, 2, 3},
			"map_col":      map[string]int32{"key1": 1, "key2": 2},
		},
		{
			"id":           uuid.New().String(),
			"text_col":     "Another text",
			"int_col":      int32(-100),
			"bigint_col":   int64(-9223372036854775808),
			"float_col":    float32(-1.5),
			"double_col":   float64(-999.999),
			"boolean_col":  false,
			"uuid_col":     uuid.New().String(),
			"timeuuid_col": uuid.New().String(),
			"timestamp_col": time.Now().Add(-24 * time.Hour),
			"date_col":     time.Now().Add(-7 * 24 * time.Hour).Truncate(24 * time.Hour),
			"time_col":     int64(23*time.Hour + 59*time.Minute + 59*time.Second),
			"inet_col":     "::1",
			"blob_col":     []byte{0xFF, 0x00, 0xAA, 0x55},
			"list_col":     []string{},
			"set_col":      []int32{99},
			"map_col":      map[string]int32{},
		},
	}

	for _, row := range testData {
		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Validate the written file
	validateParquetFile(t, outputPath, testData, columnNames)
}

// validateParquetFile reads and validates a Parquet file
func validateParquetFile(t *testing.T, filepath string, expectedData []map[string]interface{}, columnNames []string) {
	f, err := os.Open(filepath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	// Check basic properties
	assert.Equal(t, int64(len(expectedData)), reader.NumRows())
	t.Logf("File has %d rows", reader.NumRows())
	t.Logf("File has %d row groups", reader.NumRowGroups())
	t.Logf("File has %d columns", reader.MetaData().Schema.NumColumns())

	// Verify schema
	schema := reader.MetaData().Schema
	for i := 0; i < schema.NumColumns(); i++ {
		col := schema.Column(i)
		t.Logf("Column %d: %s", i, col.Name())
	}

	// Read using Arrow reader
	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	arrowSchema, err := arrowReader.Schema()
	require.NoError(t, err)

	// Log Arrow schema
	t.Logf("Arrow schema: %v", arrowSchema)

	// Read all data
	table, err := arrowReader.ReadTable(ctx)
	require.NoError(t, err)
	defer table.Release()

	assert.Equal(t, int64(len(expectedData)), table.NumRows())
	assert.Equal(t, len(columnNames), int(table.NumCols()))

	// Verify we can iterate through the data
	for i := 0; i < int(table.NumCols()); i++ {
		col := table.Column(i)
		field := table.Schema().Field(i)
		t.Logf("Column %s: %d values", field.Name, col.Len())

		// Sample first value if available
		if col.Len() > 0 {
			t.Logf("  First value type: %s", col.DataType())
		}
	}
}

// TestParquetCompatibility tests that files can be read by Arrow's Parquet reader
func TestParquetCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "compatible.parquet")

	// Create a simple dataset
	columnNames := []string{"id", "name", "value", "active"}
	columnTypes := []string{"int", "text", "double", "boolean"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)

	// Write data
	for i := 0; i < 100; i++ {
		row := map[string]interface{}{
			"id":     int32(i),
			"name":   fmt.Sprintf("Name_%d", i),
			"value":  float64(i) * 1.5,
			"active": i%2 == 0,
		}
		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Now read it back using pure Arrow/Parquet libraries
	f, err := os.Open(outputPath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	// Create Arrow reader
	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	// Read as record batches
	rr, err := arrowReader.GetRecordReader(ctx, nil, nil)
	require.NoError(t, err)
	defer rr.Release()

	recordCount := 0
	for rr.Next() {
		rec := rr.Record()
		recordCount += int(rec.NumRows())

		// Verify we can access the data
		for i := 0; i < int(rec.NumCols()); i++ {
			col := rec.Column(i)
			assert.NotNil(t, col)
		}
	}

	assert.Equal(t, 100, recordCount)
}

// TestNullHandling tests proper NULL value handling
func TestNullHandling(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "nulls.parquet")

	columnNames := []string{"id", "nullable_text", "nullable_int", "nullable_bool"}
	columnTypes := []string{"int", "text", "int", "boolean"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)

	// Write data with nulls
	testData := []map[string]interface{}{
		{"id": int32(1), "nullable_text": "text", "nullable_int": int32(10), "nullable_bool": true},
		{"id": int32(2), "nullable_text": nil, "nullable_int": int32(20), "nullable_bool": false},
		{"id": int32(3), "nullable_text": "text3", "nullable_int": nil, "nullable_bool": true},
		{"id": int32(4), "nullable_text": "text4", "nullable_int": int32(40), "nullable_bool": nil},
		{"id": int32(5), "nullable_text": nil, "nullable_int": nil, "nullable_bool": nil},
	}

	for _, row := range testData {
		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Verify nulls are properly handled
	f, err := os.Open(outputPath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	table, err := arrowReader.ReadTable(ctx)
	require.NoError(t, err)
	defer table.Release()

	// Check null counts
	for i := 0; i < int(table.NumCols()); i++ {
		col := table.Column(i)
		field := table.Schema().Field(i)

		nullCount := 0
		if col.NullN() > 0 {
			nullCount = col.NullN()
		}

		t.Logf("Column %s: %d nulls out of %d values",
			field.Name, nullCount, col.Len())

		// Verify specific null positions for nullable columns
		if field.Name == "nullable_text" {
			assert.Equal(t, 2, nullCount) // rows 2 and 5 have null
		}
	}
}

// TestJSONCompatibility tests writing JSON data as text fields
func TestJSONCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "json_data.parquet")

	columnNames := []string{"id", "metadata", "config"}
	columnTypes := []string{"uuid", "text", "text"}

	writer, err := NewParquetCaptureWriter(outputPath, columnNames, columnTypes, DefaultWriterOptions())
	require.NoError(t, err)

	// Write JSON data
	for i := 0; i < 10; i++ {
		metadata := map[string]interface{}{
			"index": i,
			"timestamp": time.Now().Unix(),
			"values": []int{i, i*2, i*3},
		}

		config := map[string]interface{}{
			"enabled": i%2 == 0,
			"threshold": float64(i) * 0.5,
			"tags": []string{fmt.Sprintf("tag_%d", i)},
		}

		metadataJSON, _ := json.Marshal(metadata)
		configJSON, _ := json.Marshal(config)

		row := map[string]interface{}{
			"id":       uuid.New().String(),
			"metadata": string(metadataJSON),
			"config":   string(configJSON),
		}

		err := writer.WriteRow(row)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Verify the file
	fileInfo, err := os.Stat(outputPath)
	require.NoError(t, err)
	t.Logf("JSON data file size: %.2f KB", float64(fileInfo.Size())/1024)

	// Read and verify JSON can be parsed back
	f, err := os.Open(outputPath)
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	ctx := context.Background()
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	table, err := arrowReader.ReadTable(ctx)
	require.NoError(t, err)
	defer table.Release()

	// Get metadata column
	for i := 0; i < int(table.NumCols()); i++ {
		field := table.Schema().Field(i)
		if field.Name == "metadata" {
			col := table.Column(i)

			// Get the first chunk of the column
			if col.Len() > 0 {
				// Access the data through the column's chunks
				chunkedData := col.Data()
				for j := 0; j < chunkedData.Len(); j++ {
					chunk := chunkedData.Chunk(j)

					// Verify it's a string array
					strArray, ok := chunk.(*array.String)
					assert.True(t, ok, "metadata should be a string column")

					// Try to parse first value as JSON
					if strArray.Len() > 0 && !strArray.IsNull(0) {
						jsonStr := strArray.Value(0)

						var parsed map[string]interface{}
						err := json.Unmarshal([]byte(jsonStr), &parsed)
						assert.NoError(t, err, "Should be able to parse JSON back")
						t.Logf("Successfully parsed JSON: %v", parsed)
						break
					}
				}
			}
		}
	}
}