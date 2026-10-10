//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/db"
)

// TestCSVRoundTripKeepsEveryType: COPY TO CSV and back through COPY FROM
// gives the same values, of every type. COPY FROM used to guess a type from
// the text: text that looked like a number went in as one and failed, and a
// timestamp, a blob or a duration could not be read at all.
func TestCSVRoundTripKeepsEveryType(t *testing.T) {
	sess, handler, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, sess.Query(`DROP TABLE IF EXISTS csv_types`).Exec())
	require.NoError(t, sess.Query(`CREATE TYPE IF NOT EXISTS csv_addr (street text, zip text)`).Exec())
	require.NoError(t, sess.Query(`CREATE TABLE csv_types (
		id int PRIMARY KEY,
		zip text, flag boolean, big bigint, small smallint, tiny tinyint,
		ratio double, f float, dec decimal, vi varint,
		ts timestamp, d date, tm time, dur duration,
		u uuid, tu timeuuid, ip inet, b blob,
		tags list<text>, nums set<int>, attrs map<text, int>,
		pair tuple<int, text>, vec vector<float, 2>, home frozen<csv_addr>, "MixedCase" text
	)`).Exec())
	defer func() { _ = sess.Query(`DROP TABLE IF EXISTS csv_types`).Exec() }()

	require.NoError(t, sess.Query(`INSERT INTO csv_types (id, zip, flag, big, small, tiny, ratio, f, dec, vi,
		ts, d, tm, dur, u, tu, ip, b, tags, nums, attrs) VALUES (1, '02134', true, 9007199254740993, -12, 7, 0.1, 1.5,
		-12.345, 123456789012345678901234567890, '2024-01-02 03:04:05.678+0000', '2024-02-29', '13:14:15.123456789',
		1h30m, 123e4567-e89b-12d3-a456-426614174000, 50554d6e-29bb-11e5-b345-feff819cdc9f, '10.0.0.1', 0xcafebabe,
		['a', 'b, c', 'it''s'], {3, 1, 2}, {'x': 1, 'y': 2})`).Exec())
	require.NoError(t, sess.Query(`UPDATE csv_types SET pair = (5, 'five'), vec = [1.5, 2.5],
		home = {street: '1 "Main" St', zip: '02134'}, "MixedCase" = 'kept' WHERE id = 1`).Exec())
	// A row of NULLs, and one whose text is empty rather than NULL.
	require.NoError(t, sess.Query(`INSERT INTO csv_types (id) VALUES (2)`).Exec())
	require.NoError(t, sess.Query(`INSERT INTO csv_types (id, zip) VALUES (3, '')`).Exec())

	before := readAll(t, sess, `SELECT * FROM csv_types`)
	require.Len(t, before, 3)

	for _, header := range []string{"true", "false"} {
		file := filepath.Join(t.TempDir(), "types.csv")
		out := handler.HandleMetaCommand(fmt.Sprintf("COPY csv_types TO '%s' WITH HEADER = %s", file, header))
		require.Equal(t, "Exported 3 rows to "+file, fmt.Sprint(out))
		data, _ := os.ReadFile(file)
		t.Logf("exported, header %s:\n%s", header, data)

		require.NoError(t, sess.Query(`TRUNCATE csv_types`).Exec())
		in := handler.HandleMetaCommand(fmt.Sprintf("COPY csv_types FROM '%s' WITH HEADER = %s", file, header))
		require.Equal(t, "Imported 3 rows from "+file, fmt.Sprint(in))

		after := readAll(t, sess, `SELECT * FROM csv_types`)
		require.Len(t, after, len(before))
		for id, row := range before {
			for column, want := range row {
				assert.Equal(t, fmt.Sprint(want), fmt.Sprint(after[id][column]), "header %s, row %v, %s", header, id, column)
			}
		}
	}
	for id, row := range readAll(t, sess, `SELECT * FROM csv_types`) {
		if fmt.Sprint(id) == "3" {
			assert.Equal(t, "", row["zip"], "empty text is not NULL")
		}
	}

	// Columns a file does not have are left as they are.
	some := filepath.Join(t.TempDir(), "some.csv")
	require.NoError(t, os.WriteFile(some, []byte("1,99999\n"), 0o600))
	in := handler.HandleMetaCommand(fmt.Sprintf("COPY csv_types (id, zip) FROM '%s'", some))
	require.Equal(t, "Imported 1 rows from "+some, fmt.Sprint(in))
	for id, row := range readAll(t, sess, `SELECT * FROM csv_types`) {
		if fmt.Sprint(id) == "1" {
			assert.Equal(t, "99999", row["zip"])
			assert.Equal(t, "kept", row["MixedCase"], "a column not in the file is not set to NULL")
		}
	}
}

// readAll is a table's rows by id, each value as ScanRow reads it.
func readAll(t *testing.T, sess *db.Session, cql string) map[interface{}]map[string]interface{} {
	t.Helper()
	iter := sess.Query(cql).Iter()
	rows := map[interface{}]map[string]interface{}{}
	for {
		row := map[string]interface{}{}
		if !db.ScanRow(iter, row) {
			break
		}
		rows[row["id"]] = row
	}
	require.NoError(t, iter.Close())
	return rows
}
