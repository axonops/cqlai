package parquet_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionedParquetWriter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parquet_partitioned_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("Basic Partitioning", func(t *testing.T) {
		options := parquet.PartitionedWriterOptions{
			WriterOptions:    parquet.DefaultWriterOptions(),
			PartitionColumns: []string{"year", "month"},
			MaxOpenFiles:     5,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
		}

		columnNames := []string{"id", "name", "year", "month", "value"}
		columnTypes := []string{"int", "text", "int", "int", "double"}

		writer, err := parquet.NewPartitionedParquetWriter(tmpDir, columnNames, columnTypes, options)
		require.NoError(t, err)
		defer writer.Close()

		// Write rows to different partitions
		rows := []map[string]interface{}{
			{"id": 1, "name": "Alice", "year": 2024, "month": 1, "value": 100.0},
			{"id": 2, "name": "Bob", "year": 2024, "month": 1, "value": 200.0},
			{"id": 3, "name": "Charlie", "year": 2024, "month": 2, "value": 300.0},
			{"id": 4, "name": "David", "year": 2023, "month": 12, "value": 400.0},
			{"id": 5, "name": "Eve", "year": 2023, "month": 12, "value": 500.0},
		}

		err = writer.WriteRows(rows)
		assert.NoError(t, err)

		err = writer.Flush()
		assert.NoError(t, err)

		// Check partition structure
		partitionInfo := writer.GetPartitionInfo()
		assert.Len(t, partitionInfo, 3) // 2024/01, 2024/02, 2023/12

		// Verify directory structure
		assert.DirExists(t, filepath.Join(tmpDir, "year=2024", "month=1"))
		assert.DirExists(t, filepath.Join(tmpDir, "year=2024", "month=2"))
		assert.DirExists(t, filepath.Join(tmpDir, "year=2023", "month=12"))

		// Verify file existence
		assert.FileExists(t, filepath.Join(tmpDir, "year=2024", "month=1", "part-00000.parquet"))
		assert.FileExists(t, filepath.Join(tmpDir, "year=2024", "month=2", "part-00000.parquet"))
		assert.FileExists(t, filepath.Join(tmpDir, "year=2023", "month=12", "part-00000.parquet"))
	})

	t.Run("LRU File Handle Management", func(t *testing.T) {
		options := parquet.PartitionedWriterOptions{
			WriterOptions:    parquet.DefaultWriterOptions(),
			PartitionColumns: []string{"partition"},
			MaxOpenFiles:     2, // Only allow 2 open files at a time
			MaxFileSize:      10 * 1024 * 1024,
		}

		columnNames := []string{"id", "partition", "value"}
		columnTypes := []string{"int", "text", "double"}

		writer, err := parquet.NewPartitionedParquetWriter(tmpDir, columnNames, columnTypes, options)
		require.NoError(t, err)
		defer writer.Close()

		// Write to more partitions than max open files
		for i := 0; i < 5; i++ {
			rows := []map[string]interface{}{
				{"id": i, "partition": fmt.Sprintf("part%d", i), "value": float64(i) * 100},
			}
			err = writer.WriteRows(rows)
			assert.NoError(t, err)
		}

		// Write again to first partition (should reopen)
		rows := []map[string]interface{}{
			{"id": 99, "partition": "part0", "value": 999.0},
		}
		err = writer.WriteRows(rows)
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		// Verify all partitions were written
		for i := 0; i < 5; i++ {
			partPath := filepath.Join(tmpDir, fmt.Sprintf("partition=part%d", i))
			assert.DirExists(t, partPath)
			assert.FileExists(t, filepath.Join(partPath, "part-00000.parquet"))
		}
	})

	t.Run("Null Values in Partitions", func(t *testing.T) {
		options := parquet.PartitionedWriterOptions{
			WriterOptions:    parquet.DefaultWriterOptions(),
			PartitionColumns: []string{"category"},
			MaxOpenFiles:     5,
			MaxFileSize:      10 * 1024 * 1024,
		}

		columnNames := []string{"id", "category", "value"}
		columnTypes := []string{"int", "text", "double"}

		writer, err := parquet.NewPartitionedParquetWriter(tmpDir, columnNames, columnTypes, options)
		require.NoError(t, err)
		defer writer.Close()

		// Write rows with null partition values
		rows := []map[string]interface{}{
			{"id": 1, "category": "electronics", "value": 100.0},
			{"id": 2, "category": nil, "value": 200.0}, // NULL category
			{"id": 3, "category": "books", "value": 300.0},
			{"id": 4, "category": nil, "value": 400.0}, // NULL category
		}

		err = writer.WriteRows(rows)
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		// Verify NULL partition directory
		assert.DirExists(t, filepath.Join(tmpDir, "category=__NULL__"))
		assert.DirExists(t, filepath.Join(tmpDir, "category=electronics"))
		assert.DirExists(t, filepath.Join(tmpDir, "category=books"))
	})

	t.Run("Special Characters in Partition Values", func(t *testing.T) {
		options := parquet.PartitionedWriterOptions{
			WriterOptions:    parquet.DefaultWriterOptions(),
			PartitionColumns: []string{"path"},
			MaxOpenFiles:     5,
			MaxFileSize:      10 * 1024 * 1024,
		}

		columnNames := []string{"id", "path", "value"}
		columnTypes := []string{"int", "text", "double"}

		writer, err := parquet.NewPartitionedParquetWriter(tmpDir, columnNames, columnTypes, options)
		require.NoError(t, err)
		defer writer.Close()

		// Write rows with special characters in partition values
		rows := []map[string]interface{}{
			{"id": 1, "path": "dir/subdir", "value": 100.0},  // Contains slash
			{"id": 2, "path": "key=value", "value": 200.0},    // Contains equals
			{"id": 3, "path": "normal_value", "value": 300.0}, // Normal value
		}

		err = writer.WriteRows(rows)
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		// Verify escaped partition directories
		assert.DirExists(t, filepath.Join(tmpDir, "path=dir__SLASH__subdir"))
		assert.DirExists(t, filepath.Join(tmpDir, "path=key__EQ__value"))
		assert.DirExists(t, filepath.Join(tmpDir, "path=normal_value"))
	})
}

func TestPartitionedParquetReader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parquet_partitioned_read_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a partitioned dataset first
	createPartitionedDataset := func() {
		options := parquet.PartitionedWriterOptions{
			WriterOptions:    parquet.DefaultWriterOptions(),
			PartitionColumns: []string{"year", "month"},
			MaxOpenFiles:     5,
			MaxFileSize:      10 * 1024 * 1024,
		}

		columnNames := []string{"id", "name", "year", "month", "value"}
		columnTypes := []string{"int", "text", "int", "int", "double"}

		writer, err := parquet.NewPartitionedParquetWriter(tmpDir, columnNames, columnTypes, options)
		require.NoError(t, err)

		rows := []map[string]interface{}{
			{"id": 1, "name": "Alice", "year": 2024, "month": 1, "value": 100.0},
			{"id": 2, "name": "Bob", "year": 2024, "month": 1, "value": 200.0},
			{"id": 3, "name": "Charlie", "year": 2024, "month": 2, "value": 300.0},
			{"id": 4, "name": "David", "year": 2023, "month": 12, "value": 400.0},
			{"id": 5, "name": "Eve", "year": 2023, "month": 12, "value": 500.0},
		}

		err = writer.WriteRows(rows)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)
	}

	createPartitionedDataset()

	t.Run("Read All Partitions", func(t *testing.T) {
		reader, err := parquet.NewPartitionedParquetReader(tmpDir)
		require.NoError(t, err)
		defer reader.Close()

		// Get schema including partition columns
		columns, _ := reader.GetSchema()
		assert.Contains(t, columns, "year")
		assert.Contains(t, columns, "month")
		assert.Contains(t, columns, "id")
		assert.Contains(t, columns, "name")
		assert.Contains(t, columns, "value")

		// Read all rows
		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 5)

		// Check partition values are included in rows
		for _, row := range rows {
			assert.Contains(t, row, "year")
			assert.Contains(t, row, "month")
			assert.Contains(t, row, "id")
			assert.Contains(t, row, "name")
			assert.Contains(t, row, "value")
		}
	})

	t.Run("Read Batch Across Partitions", func(t *testing.T) {
		reader, err := parquet.NewPartitionedParquetReader(tmpDir)
		require.NoError(t, err)
		defer reader.Close()

		totalRows := 0
		for {
			batch, err := reader.ReadBatch(2)
			if err != nil || len(batch) == 0 {
				break
			}

			totalRows += len(batch)

			// Verify batch has partition values
			for _, row := range batch {
				assert.Contains(t, row, "year")
				assert.Contains(t, row, "month")
			}
		}

		assert.Equal(t, 5, totalRows)
	})

	t.Run("Partition Discovery", func(t *testing.T) {
		reader, err := parquet.NewPartitionedParquetReader(tmpDir)
		require.NoError(t, err)
		defer reader.Close()

		partitionFiles := reader.GetPartitionFiles()
		assert.Len(t, partitionFiles, 3) // 3 partition files

		partitionColumns := reader.GetPartitionColumns()
		assert.Equal(t, []string{"month", "year"}, partitionColumns) // Sorted order

		// Verify partition values were parsed correctly
		foundPartitions := make(map[string]bool)
		for _, pf := range partitionFiles {
			key := fmt.Sprintf("%v_%v", pf.PartitionValues["year"], pf.PartitionValues["month"])
			foundPartitions[key] = true
		}

		assert.True(t, foundPartitions["2024_1"])
		assert.True(t, foundPartitions["2024_2"])
		assert.True(t, foundPartitions["2023_12"])
	})

	t.Run("Empty Directory", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "empty_parquet_test")
		require.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		_, err = parquet.NewPartitionedParquetReader(emptyDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no Parquet files found")
	})

	t.Run("Single File Error", func(t *testing.T) {
		// Try to read a single file as partitioned dataset
		singleFile := filepath.Join(tmpDir, "year=2024", "month=1", "part-00000.parquet")
		_, err := parquet.NewPartitionedParquetReader(singleFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}