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

// TestMapsKeyedByBlobsAndCollectionsCanBeRead: gocql cannot make a Go map
// with a slice for its key, and panicked reading one - in a column, or in a
// UDT - which took the shell down with it.
func TestMapsKeyedByBlobsAndCollectionsCanBeRead(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()

	for _, stmt := range []string{
		`DROP TABLE IF EXISTS odd_keys`,
		`CREATE TYPE IF NOT EXISTS odd_udt (dec decimal, dur duration, bm map<blob, int>, im map<inet, int>)`,
		`CREATE TABLE odd_keys (id int PRIMARY KEY, bm map<blob, int>, lm map<frozen<list<int>>, int>, o frozen<odd_udt>)`,
		`INSERT INTO odd_keys (id, bm, lm, o) VALUES (1, {0xcafe: 1}, {[1, 2]: 3},
			{dec: -0.05, dur: 1mo2d1h30m, bm: {0xbeef: 4}, im: {'10.0.0.1': 5}})`,
		`INSERT INTO odd_keys (id) VALUES (2)`,
	} {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}
	defer func() { _ = sess.Query(`DROP TABLE IF EXISTS odd_keys`).Exec() }()

	var rows map[interface{}]map[string]interface{}
	require.NotPanics(t, func() { rows = readAll(t, sess, `SELECT * FROM odd_keys`) })
	require.Len(t, rows, 2)

	h := db.NewCQLTypeHandler()
	for id, row := range rows {
		if h.FormatValue(id, nil) == "2" {
			assert.Nil(t, row["bm"])
			assert.Nil(t, row["o"])
			continue
		}
		assert.Equal(t, "{0xcafe: 1}", h.FormatValue(row["bm"], nil))
		assert.Equal(t, "{[1, 2]: 3}", h.FormatValue(row["lm"], nil))
		o, ok := row["o"].(map[string]interface{})
		require.True(t, ok, "%T", row["o"])
		assert.Equal(t, gocql.Duration{Months: 1, Days: 2, Nanoseconds: 5_400_000_000_000}, o["dur"])
		assert.Equal(t, "{0xbeef: 4}", h.FormatValue(o["bm"], nil))
		assert.Equal(t, "{'10.0.0.1': 5}", h.FormatValue(o["im"], nil))
	}
}
