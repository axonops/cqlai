//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParquetRoundTripKeepsDecimalsAndDurations: a decimal and a duration
// were written to Parquet as NULL. Each is now written as its CQL text, every
// digit and every month and day kept.
func TestParquetRoundTripKeepsDecimalsAndDurations(t *testing.T) {
	sess, handler, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, sess.Query(`DROP TABLE IF EXISTS pq_types`).Exec())
	require.NoError(t, sess.Query(`CREATE TABLE pq_types (id int PRIMARY KEY, dec decimal, dur duration, vi varint)`).Exec())
	defer func() { _ = sess.Query(`DROP TABLE IF EXISTS pq_types`).Exec() }()
	for _, stmt := range []string{
		`INSERT INTO pq_types (id, dec, dur, vi) VALUES (1, -12.345, 1h30m, 123456789012345678901234567890)`,
		`INSERT INTO pq_types (id, dec, dur) VALUES (2, 123456789012345678901234567890.123456789012345678901234567890, 14mo3d2ms)`,
		`INSERT INTO pq_types (id) VALUES (3)`,
	} {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}
	before := readAll(t, sess, `SELECT * FROM pq_types`)

	file := filepath.Join(t.TempDir(), "types.parquet")
	out := handler.HandleMetaCommand(fmt.Sprintf("COPY pq_types TO '%s'", file))
	require.Contains(t, fmt.Sprint(out), "3 rows", out)

	require.NoError(t, sess.Query(`TRUNCATE pq_types`).Exec())
	in := handler.HandleMetaCommand(fmt.Sprintf("COPY pq_types FROM '%s'", file))
	require.Contains(t, fmt.Sprint(in), "3 rows", in)

	after := readAll(t, sess, `SELECT * FROM pq_types`)
	require.Len(t, after, 3)
	for id, row := range before {
		for column, want := range row {
			assert.Equal(t, fmt.Sprint(want), fmt.Sprint(after[id][column]), "row %v, %s", id, column)
		}
	}
}

// TestPartitionedParquetKeepsEveryRow: more partition values than files are
// kept open, with the rows for each arriving in turn. A partition's file was
// opened again over itself, so its earlier rows were lost.
func TestPartitionedParquetKeepsEveryRow(t *testing.T) {
	sess, handler, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, sess.Query(`DROP TABLE IF EXISTS pq_parts`).Exec())
	require.NoError(t, sess.Query(`CREATE TABLE pq_parts (id int PRIMARY KEY, region text)`).Exec())
	defer func() { _ = sess.Query(`DROP TABLE IF EXISTS pq_parts`).Exec() }()
	const rows = 300
	for i := 0; i < rows; i++ {
		require.NoError(t, sess.Query(`INSERT INTO pq_parts (id, region) VALUES (?, ?)`, i, fmt.Sprintf("r%02d", i%25)).Exec())
	}

	dir := filepath.Join(t.TempDir(), "parts")
	out := handler.HandleMetaCommand(fmt.Sprintf("COPY pq_parts TO '%s' WITH FORMAT = 'parquet' AND PARTITION = 'region'", dir))
	assert.Equal(t, fmt.Sprintf("Exported %d rows to 25 partitions in %s", rows, dir), fmt.Sprint(out))

	require.NoError(t, sess.Query(`TRUNCATE pq_parts`).Exec())
	in := handler.HandleMetaCommand(fmt.Sprintf("COPY pq_parts FROM '%s' WITH FORMAT = 'parquet'", dir))
	t.Logf("%v", in)

	var n int
	require.NoError(t, sess.Query(`SELECT COUNT(*) FROM pq_parts`).Scan(&n))
	assert.Equal(t, rows, n)
}
