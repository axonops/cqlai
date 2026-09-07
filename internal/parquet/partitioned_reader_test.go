package parquet

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePartitionedDataset builds a Hive-style dataset:
//
//	<base>/year=2024/month=01/data.parquet
//	<base>/year=2024/month=02/data.parquet
//	<base>/year=2025/month=01/data.parquet
//
// Each file holds rowsPerFile rows with a globally increasing id, so the
// concatenation of every file is 0..(3*rowsPerFile-1) in order.
func writePartitionedDataset(t *testing.T, rowsPerFile, chunkSize int) (base string, total int) {
	t.Helper()

	base = t.TempDir()
	parts := []struct{ year, month string }{
		{"2024", "01"},
		{"2024", "02"},
		{"2025", "01"},
	}

	id := 0
	for _, p := range parts {
		dir := filepath.Join(base, "year="+p.year, "month="+p.month)
		require.NoError(t, os.MkdirAll(dir, 0o750))

		opts := DefaultWriterOptions()
		opts.ChunkSize = int64(chunkSize)

		w, err := NewParquetCaptureWriter(filepath.Join(dir, "data.parquet"),
			[]string{"id", "name"}, []string{"int", "text"}, opts)
		require.NoError(t, err)

		for range rowsPerFile {
			require.NoError(t, w.WriteRow(map[string]interface{}{
				"id":   id,
				"name": fmt.Sprintf("row-%d", id),
			}))
			id++
		}
		require.NoError(t, w.Close())
	}

	return base, id
}

func drainPartitioned(t *testing.T, pr *PartitionedParquetReader, batchSize int) []map[string]interface{} {
	t.Helper()

	var all []map[string]interface{}
	for range 1000 {
		batch, err := pr.ReadBatch(batchSize)
		if errors.Is(err, io.EOF) {
			all = append(all, batch...)
			return all
		}
		require.NoError(t, err)
		require.NotEmpty(t, batch, "ReadBatch returned no rows and no io.EOF")
		all = append(all, batch...)
	}

	t.Fatal("ReadBatch never returned io.EOF")
	return nil
}

// TestPartitionedReadBatchAcrossFiles is the reachable path: COPY FROM a
// partitioned dataset streams through it.
func TestPartitionedReadBatchAcrossFiles(t *testing.T) {
	const rowsPerFile = 250

	base, total := writePartitionedDataset(t, rowsPerFile, 100)

	for _, batchSize := range []int{1, 30, 250, 400, 10000} {
		t.Run(fmt.Sprintf("batch=%d", batchSize), func(t *testing.T) {
			pr, err := NewPartitionedParquetReader(base)
			require.NoError(t, err)
			defer pr.Close()

			rows := drainPartitioned(t, pr, batchSize)

			require.Equal(t, total, len(rows), "rows lost or duplicated across partition files")
			for i, row := range rows {
				assert.Equal(t, int32(i), row["id"], "row %d out of order or wrong", i)
			}
		})
	}
}

// TestPartitionedReadBatchAddsPartitionColumns checks the partition values from
// the directory names land on the rows, with the numeric coercion the reader does.
func TestPartitionedReadBatchAddsPartitionColumns(t *testing.T) {
	base, total := writePartitionedDataset(t, 10, 100)

	pr, err := NewPartitionedParquetReader(base)
	require.NoError(t, err)
	defer pr.Close()

	assert.Equal(t, []string{"month", "year"}, pr.GetPartitionColumns())

	rows := drainPartitioned(t, pr, 7)
	require.Equal(t, total, len(rows))

	// Files sort by path, so the first 10 rows are year=2024/month=01.
	assert.Equal(t, int64(2024), rows[0]["year"], "partition value should be coerced to a number")
	assert.Equal(t, int64(1), rows[0]["month"], "month=01 should parse as 1")

	// Last file is year=2025/month=01.
	assert.Equal(t, int64(2025), rows[total-1]["year"])
}

// TestPartitionedGetRowCount pins down what GetRowCount actually reports.
func TestPartitionedGetRowCount(t *testing.T) {
	const rowsPerFile = 100

	base, total := writePartitionedDataset(t, rowsPerFile, 1000)

	pr, err := NewPartitionedParquetReader(base)
	require.NoError(t, err)
	defer pr.Close()

	// Files are opened lazily, so before reading only the first one is counted.
	assert.Equal(t, int64(rowsPerFile), pr.GetRowCount(),
		"straight after construction only the first file has been opened")

	rows := drainPartitioned(t, pr, 50)
	require.Equal(t, total, len(rows))

	assert.Equal(t, int64(total), pr.GetRowCount(),
		"once every file has been read the tally should equal the dataset")
}

// TestPartitionedReadAllRowCount exercises ReadAll, which has no production
// caller today but is exported.
func TestPartitionedReadAllRowCount(t *testing.T) {
	base, total := writePartitionedDataset(t, 100, 1000)

	pr, err := NewPartitionedParquetReader(base)
	require.NoError(t, err)
	defer pr.Close()

	rows, err := pr.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, total, len(rows))

	// ReadAll restarts the walk from the first file. It must reset the tally
	// rather than adding the first file's rows on top of the constructor's.
	assert.Equal(t, int64(total), pr.GetRowCount(),
		"ReadAll double-counted the file opened by the constructor")
}
