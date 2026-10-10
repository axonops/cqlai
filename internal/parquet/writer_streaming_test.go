package parquet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRowsGoToTheFileAsTheyCome: each chunk is written once it fills. They
// were kept until Close and written as one table, so an export held every
// row in memory until the end, and the file was empty until then.
func TestRowsGoToTheFileAsTheyCome(t *testing.T) {
	file := filepath.Join(t.TempDir(), "big.parquet")
	opts := DefaultWriterOptions()
	opts.ChunkSize = 1000
	w, err := NewParquetCaptureWriter(file, []string{"id", "name"}, []string{"int", "text"}, opts)
	require.NoError(t, err)

	for i := 0; i < 20000; i++ {
		require.NoError(t, w.WriteRow(map[string]interface{}{"id": i, "name": "a row of some length"}))
	}
	info, err := os.Stat(file)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(10000), "rows are on disk before Close")

	require.NoError(t, w.Close())
	r, err := NewParquetReader(file)
	require.NoError(t, err)
	defer r.Close()
	rows, err := r.ReadAll()
	require.NoError(t, err)
	assert.Len(t, rows, 20000)
}

// TestAFileWithNoRowsStillReads: it has the schema and no rows.
func TestAFileWithNoRowsStillReads(t *testing.T) {
	file := filepath.Join(t.TempDir(), "empty.parquet")
	w, err := NewParquetCaptureWriter(file, []string{"id"}, []string{"int"}, DefaultWriterOptions())
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := NewParquetReader(file)
	require.NoError(t, err)
	defer r.Close()
	rows, err := r.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, rows)
}
