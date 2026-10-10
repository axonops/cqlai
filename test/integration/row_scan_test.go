//go:build integration
// +build integration

package integration_test

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/db"
)

// TestScanRowTellsNullFromZero, keeps a tuple whole, and the display shows
// both, with a timestamp's milliseconds.
func TestScanRowTellsNullFromZero(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()

	const ks = "test_row_scan"
	require.NoError(t, sess.Query(`DROP KEYSPACE IF EXISTS `+ks).Exec())
	require.NoError(t, sess.Query(`CREATE KEYSPACE `+ks+` WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec())
	defer func() { _ = sess.Query(`DROP KEYSPACE IF EXISTS ` + ks).Exec() }()

	for _, stmt := range []string{
		`CREATE TABLE ` + ks + `.t (id int PRIMARY KEY, n int, s text, b boolean, u uuid, ts timestamp,
			big varint, dec decimal, l list<int>, m map<text, int>, tup tuple<int, text>)`,
		`INSERT INTO ` + ks + `.t (id) VALUES (1)`,
		`INSERT INTO ` + ks + `.t (id, n, s, b, u, ts, big, dec, l, m, tup) VALUES (2, 0, '', false,
			123e4567-e89b-12d3-a456-426614174000, '2024-01-02 03:04:05.678+0000', 12345678901234567890, -0.05,
			[1, 2], {'a': 1}, (5, 'five'))`,
	} {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}

	rows := map[int]map[string]interface{}{}
	var cols []gocql.ColumnInfo
	iter := sess.Query(`SELECT * FROM ` + ks + `.t`).Iter()
	cols = iter.Columns()
	for {
		row := map[string]interface{}{}
		if !db.ScanRow(iter, row) {
			break
		}
		rows[row["id"].(int)] = row
	}
	require.NoError(t, iter.Close())
	require.Len(t, rows, 2)

	// Every column but the key is NULL in row 1, and nil here - not 0, "",
	// false or the zero UUID.
	for name, v := range rows[1] {
		if name != "id" {
			assert.Nil(t, v, "%s was NULL", name)
		}
	}

	// Row 2 holds the zero values themselves, which are not NULL.
	r := rows[2]
	assert.Equal(t, 0, r["n"])
	assert.Equal(t, "", r["s"])
	assert.Equal(t, false, r["b"])
	assert.Equal(t, []interface{}{5, "five"}, r["tup"], "the tuple, whole, under its own name")

	// And the display.
	h := db.NewCQLTypeHandler()
	shown := map[string]string{}
	for _, c := range cols {
		shown[c.Name] = h.FormatValue(r[c.Name], c.TypeInfo)
	}
	assert.Equal(t, "2024-01-02 03:04:05.678+0000", shown["ts"], "with its milliseconds")
	assert.Equal(t, "(5, five)", shown["tup"], "as cqlai shows text in any collection")
	assert.Equal(t, "12345678901234567890", shown["big"])
	assert.Equal(t, "-0.05", shown["dec"])
	for _, c := range cols {
		if c.Name != "id" {
			assert.Equal(t, "null", h.FormatValue(rows[1][c.Name], c.TypeInfo), c.Name)
		}
	}
}
