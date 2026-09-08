package parquet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReaderCloseReportsSuccess guards against Close reporting a failure on a
// perfectly good reader.
//
// The parquet reader owns the file handle and closes it. Closing the handle
// again afterwards returned "file already closed" from every successful Close,
// which every caller had to ignore for the code to work at all.
func TestReaderCloseReportsSuccess(t *testing.T) {
	path := writeNumberedRows(t, 10, 100)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)

	assert.NoError(t, reader.Close(), "closing a healthy reader must not report an error")
}

// TestReaderCloseIsIdempotent covers the common pattern of a deferred Close
// alongside an explicit one.
func TestReaderCloseIsIdempotent(t *testing.T) {
	path := writeNumberedRows(t, 10, 100)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)

	require.NoError(t, reader.Close())
	assert.NoError(t, reader.Close(), "a second Close must be a no-op")
	assert.NoError(t, reader.Close())
}

// TestReaderCloseAfterReadingReportsSuccess checks the same once the reader has
// actually done some work, so the batch and row-group state is populated.
func TestReaderCloseAfterReadingReportsSuccess(t *testing.T) {
	path := writeNumberedRows(t, 350, 100)

	reader, err := NewParquetReader(path)
	require.NoError(t, err)

	rows := drainBatches(t, reader, 60)
	require.Equal(t, 350, len(rows))

	assert.NoError(t, reader.Close())
}
