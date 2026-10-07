package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func implications(t *testing.T, cql string, f Facts) string {
	t.Helper()
	s, err := Classify(cql)
	require.NoError(t, err, cql)
	return strings.Join(Implications(s, f), "\n")
}

var ordersKey = Facts{PartitionKey: []string{"customer", "day"}, ClusteringKey: []string{"id"}}

// TestEachChangeSaysWhatItWillDo, in cqlai's words.
func TestEachChangeSaysWhatItWillDo(t *testing.T) {
	for cql, want := range map[string][]string{
		"DROP TABLE shop.orders":                {"deletes the table and all its data", "auto_snapshot", "schema change"},
		"DROP KEYSPACE shop":                    {"every table, type, function and view in it"},
		"TRUNCATE shop.orders":                  {"every row of the table on every node", "Every node has to be up"},
		"ALTER TABLE shop.orders DROP note":     {"unreadable at once", "cannot be added back with a different type", "nodetool snapshot"},
		"ALTER TABLE shop.orders ADD note text": {"changes only the schema"},
		"ALTER TABLE shop.orders WITH compaction = {'class': 'LeveledCompactionStrategy'}":                   {"rewrites the table's data"},
		"ALTER TABLE shop.orders WITH gc_grace_seconds = 3600":                                               {"deleted data can come back"},
		"ALTER KEYSPACE shop WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3}":              {"moves no data", "nodetool repair -full", "nodetool cleanup"},
		"CREATE KEYSPACE shop WITH replication = {'class': 'SimpleStrategy'} AND durable_writes = false":     {"SimpleStrategy is for one data centre", "skips the commit log"},
		"CREATE TABLE shop.t (id int PRIMARY KEY)":                                                           {"primary key cannot be changed later"},
		"CREATE INDEX ON shop.orders (total)":                                                                {"reads all of the table's data", "asks every node"},
		"CREATE MATERIALIZED VIEW shop.v AS SELECT * FROM shop.orders WHERE id IS NOT NULL PRIMARY KEY (id)": {"experimental"},
		"INSERT INTO shop.orders (customer, day, id) VALUES ('a', 'b', 1)":                                   {"overwrites a row", "IF NOT EXISTS"},
		"INSERT INTO shop.orders (customer, day, id) VALUES ('a', 'b', 1) IF NOT EXISTS":                     {"lightweight transaction"},
		"UPDATE shop.orders USING TTL 60 SET total = 1 WHERE customer = 'a' AND day = 'b' AND id = 1":        {"creates the row", "expires after its TTL"},
		"UPDATE shop.orders SET total = 1 WHERE customer = 'a' AND day = 'b' AND id = 1 IF total = 0":        {"Paxos"},
		"BEGIN UNLOGGED BATCH INSERT INTO shop.a (id) VALUES (1); APPLY BATCH":                               {"not atomic"},
		"BEGIN BATCH DELETE FROM shop.a WHERE id = 1; APPLY BATCH":                                           {"all of its statements or none", "tombstone"},
	} {
		got := implications(t, cql, ordersKey)
		for _, w := range want {
			assert.Contains(t, got, w, cql)
		}
	}

	// A plain INSERT is not told it is a schema change, and the IF NOT
	// EXISTS advice is not given to one that has it.
	got := implications(t, "INSERT INTO shop.orders (customer, day, id) VALUES ('a', 'b', 1) IF NOT EXISTS", ordersKey)
	assert.NotContains(t, got, "add IF NOT EXISTS")
	assert.NotContains(t, got, "schema change")
}

// TestADeleteSaysHowMuchItRemoves, from which key columns it fixes.
func TestADeleteSaysHowMuchItRemoves(t *testing.T) {
	for cql, want := range map[string]string{
		"DELETE FROM shop.orders WHERE customer = 'a' AND day = 'b' AND id = 1":       "one row",
		"DELETE FROM shop.orders WHERE customer = 'a' AND day = 'b'":                  "the whole partition",
		"DELETE FROM shop.orders WHERE customer = 'a' AND day = 'b' AND id > 1":       "the whole partition", // a range on the first clustering column is not fixed
		"DELETE FROM shop.orders WHERE customer = 'a'":                                "Cassandra refuses",
		"DELETE total FROM shop.orders WHERE customer = 'a' AND day = 'b' AND id = 1": "Only the named columns",
	} {
		assert.Contains(t, implications(t, cql, ordersKey), want, cql)
	}

	twoClustering := Facts{PartitionKey: []string{"customer"}, ClusteringKey: []string{"day", "id"}}
	assert.Contains(t, implications(t, "DELETE FROM shop.orders WHERE customer = 'a' AND day = 'b'", twoClustering), "a range of rows")

	assert.Contains(t, implications(t, "DELETE FROM shop.orders WHERE customer = 'a'", Facts{}), "could not be found",
		"without the table's key it says what it cannot know")
}
