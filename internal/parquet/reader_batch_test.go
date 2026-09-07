package parquet

import (
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeNumberedRows writes rowCount rows of {id, name} into a Parquet file whose
// row groups hold chunkSize rows each, and returns the path.
func writeNumberedRows(t *testing.T, rowCount, chunkSize int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "batches.parquet")

	opts := DefaultWriterOptions()
	opts.ChunkSize = int64(chunkSize)

	writer, err := NewParquetCaptureWriter(path, []string{"id", "name"}, []string{"int", "text"}, opts)
	require.NoError(t, err)

	for i := range rowCount {
		require.NoError(t, writer.WriteRow(map[string]interface{}{
			"id":   i,
			"name": string(rune('a'+i%26)) + "-row",
		}))
	}
	require.NoError(t, writer.Close())

	return path
}

// drainBatches calls ReadBatch until io.EOF and returns everything it handed back.
func drainBatches(t *testing.T, r *ParquetReader, batchSize int) []map[string]any {
	t.Helper()

	var all []map[string]any
	for range 1000 { // guard against a reader that never reports EOF
		batch, err := r.ReadBatch(batchSize)
		if errors.Is(err, io.EOF) {
			return all
		}
		require.NoError(t, err)
		require.NotEmpty(t, batch, "ReadBatch returned no rows and no io.EOF, which would spin forever")
		all = append(all, batch...)
	}

	t.Fatal("ReadBatch never returned io.EOF")
	return nil
}

// TestReadBatchSpansRowGroups is the regression guard for row-group streaming.
//
// This path has broken before against an arrow-go upgrade, in a way that made
// every batch come back with zero rows while ReadAll kept working - so the rest
// of the suite stayed green. It is worth testing on its own for that reason.
func TestReadBatchSpansRowGroups(t *testing.T) {
	const (
		rowCount  = 350
		chunkSize = 100
	)

	path := writeNumberedRows(t, rowCount, chunkSize)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)
	defer reader.Close()

	require.Greater(t, reader.NumRowGroups(), 1,
		"test needs more than one row group to be meaningful")
	require.Equal(t, int64(rowCount), reader.GetRowCount())

	// A batch size that divides neither the row count nor the row group size,
	// so batches straddle row group boundaries.
	rows := drainBatches(t, reader, 60)

	require.Len(t, rows, rowCount, "streaming read lost or duplicated rows")

	for i, row := range rows {
		assert.Equal(t, int32(i), row["id"], "row %d out of order or wrong", i)
	}
}

// TestReadBatchSizes checks the batch size edge cases around row group
// boundaries, since that is where the streaming bookkeeping goes wrong.
func TestReadBatchSizes(t *testing.T) {
	const (
		rowCount  = 250
		chunkSize = 100
	)

	tests := []struct {
		name      string
		batchSize int
	}{
		{"one row at a time", 1},
		{"smaller than a row group", 30},
		{"exactly a row group", 100},
		{"larger than a row group", 175},
		{"larger than the whole file", 5000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeNumberedRows(t, rowCount, chunkSize)

			reader, err := NewParquetReader(path)
			require.NoError(t, err)
			defer reader.Close()

			rows := drainBatches(t, reader, tt.batchSize)

			require.Len(t, rows, rowCount)
			for i, row := range rows {
				assert.Equal(t, int32(i), row["id"], "row %d out of order or wrong", i)
			}
		})
	}
}

// TestReadBatchAfterExhaustion checks that a reader keeps reporting io.EOF once
// it is drained, rather than restarting or erroring.
func TestReadBatchAfterExhaustion(t *testing.T) {
	path := writeNumberedRows(t, 50, 100)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)
	defer reader.Close()

	rows := drainBatches(t, reader, 20)
	require.Len(t, rows, 50)

	for range 3 {
		_, err := reader.ReadBatch(20)
		assert.ErrorIs(t, err, io.EOF)
	}
}

// TestReadBatchAtProductionDefaults uses the sizes COPY actually runs with.
//
// The writer defaults to 10000 rows per row group and COPY FROM PARQUET reads
// in batches of 1000, so the batch size is well under the row group size - the
// case that used to drop every row after the first batch of each group.
func TestReadBatchAtProductionDefaults(t *testing.T) {
	const (
		writerChunkSize = 10000 // DefaultWriterOptions().ChunkSize
		copyBatchSize   = 1000  // copy_from_parquet.go
		rowCount        = 25000 // spans three row groups, last one partial
	)

	path := writeNumberedRows(t, rowCount, writerChunkSize)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)
	defer reader.Close()

	require.Equal(t, int64(rowCount), reader.GetRowCount())

	rows := drainBatches(t, reader, copyBatchSize)

	require.Len(t, rows, rowCount,
		"COPY FROM PARQUET would import %d of %d rows", len(rows), rowCount)

	for i, row := range rows {
		assert.Equal(t, int32(i), row["id"], "row %d out of order or wrong", i)
	}
}
