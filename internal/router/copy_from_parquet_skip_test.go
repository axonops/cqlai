package router

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/parquet"
)

// TestSkipRowsLeavesTheRest: SKIPROWS read a whole batch to skip a few rows,
// and the rest of that batch was lost to the import as well.
func TestSkipRowsLeavesTheRest(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ten.parquet")
	w, err := parquet.NewParquetCaptureWriter(file, []string{"id"}, []string{"int"}, parquet.DefaultWriterOptions())
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		require.NoError(t, w.WriteRow(map[string]interface{}{"id": i}))
	}
	require.NoError(t, w.Close())

	for _, batchSize := range []int{1, 2, 1000} {
		reader, err := parquet.NewParquetReader(file)
		require.NoError(t, err)

		opts := &copyOptions{skipRows: 3, batchSize: batchSize}
		stats := &copyStats{}
		require.NoError(t, (&MetaCommandHandler{}).skipRows(reader, opts, stats))
		assert.Equal(t, 3, stats.skippedRows, "batch %d", batchSize)

		var ids []interface{}
		for {
			batch, err := reader.ReadBatch(1000)
			for _, row := range batch {
				ids = append(ids, row["id"])
			}
			if err != nil || len(batch) == 0 {
				break
			}
		}
		require.Len(t, ids, 7, "batch %d", batchSize)
		assert.EqualValues(t, 3, ids[0], "batch %d", batchSize)
		_ = reader.Close()
	}
}
