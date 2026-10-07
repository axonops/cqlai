package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEachCommandIsClassified: the command, its kind, and the names it
// touches.
func TestEachCommandIsClassified(t *testing.T) {
	for _, tc := range []struct {
		cql     string
		command string
		kind    Kind
		names   []Name
	}{
		{"SELECT * FROM shop.orders", "SELECT", KindRead, []Name{{"shop", "orders", true}}},
		{"select id from Shop.Orders where id = 1", "SELECT", KindRead, []Name{{"shop", "orders", true}}},
		{`SELECT * FROM "Shop"."Orders"`, "SELECT", KindRead, []Name{{"Shop", "Orders", true}}},
		{"INSERT INTO shop.orders (id) VALUES (1)", "INSERT", KindWrite, []Name{{"shop", "orders", true}}},
		{"UPDATE shop.orders SET total = 1 WHERE id = 1", "UPDATE", KindWrite, []Name{{"shop", "orders", true}}},
		{"DELETE total FROM shop.orders WHERE id = 1", "DELETE", KindWrite, []Name{{"shop", "orders", true}}},
		{"TRUNCATE shop.orders", "TRUNCATE", KindSchema, []Name{{"shop", "orders", true}}},
		{"TRUNCATE TABLE shop.orders", "TRUNCATE", KindSchema, []Name{{"shop", "orders", true}}},
		{"CREATE TABLE IF NOT EXISTS shop.t (id int PRIMARY KEY)", "CREATE", KindSchema, []Name{{"shop", "t", true}}},
		{"CREATE KEYSPACE shop WITH replication = {'class': 'SimpleStrategy'}", "CREATE", KindSchema, []Name{{"shop", "", true}}},
		{"CREATE INDEX ON shop.orders (total)", "CREATE", KindSchema, []Name{{"shop", "orders", true}}},
		{"CREATE CUSTOM INDEX idx ON shop.orders (total) USING 'StorageAttachedIndex'", "CREATE", KindSchema, []Name{{"shop", "orders", true}}},
		{"CREATE OR REPLACE FUNCTION shop.f (a int) RETURNS NULL ON NULL INPUT RETURNS int LANGUAGE java AS $$ return a; $$", "CREATE", KindSchema, []Name{{"shop", "f", true}}},
		{"CREATE MATERIALIZED VIEW shop.v AS SELECT * FROM shop.orders WHERE id IS NOT NULL PRIMARY KEY (id)", "CREATE", KindSchema, []Name{{"shop", "v", true}, {"shop", "orders", true}}},
		{"ALTER TABLE shop.orders ADD note text", "ALTER", KindSchema, []Name{{"shop", "orders", true}}},
		{"ALTER KEYSPACE shop WITH durable_writes = true", "ALTER", KindSchema, []Name{{"shop", "", true}}},
		{"DROP TABLE IF EXISTS shop.orders", "DROP", KindSchema, []Name{{"shop", "orders", true}}},
		{"DROP INDEX shop.orders_total", "DROP", KindSchema, []Name{{"shop", "", true}}},
		{"DROP KEYSPACE shop", "DROP", KindSchema, []Name{{"shop", "", true}}},
		{"DESCRIBE TABLE shop.orders", "DESCRIBE", KindRead, nil},
		{"LIST ROLES", "LIST", KindRead, nil},
	} {
		s, err := Classify(tc.cql)
		require.NoError(t, err, tc.cql)
		assert.Equal(t, tc.command, s.Command, tc.cql)
		assert.Equal(t, tc.kind, s.Kind, tc.cql)
		assert.Equal(t, tc.names, s.Names, tc.cql)
	}
}

// TestNeverPermittedStatementsAreRefused, whatever they look like.
func TestNeverPermittedStatementsAreRefused(t *testing.T) {
	for _, cql := range []string{
		"GRANT SELECT ON ALL KEYSPACES TO bob",
		"REVOKE SELECT ON ALL KEYSPACES FROM bob",
		"CREATE ROLE bob WITH PASSWORD = 'x'",
		"ALTER ROLE cassandra WITH PASSWORD = 'x'",
		"DROP ROLE bob",
		"CREATE USER bob WITH PASSWORD 'x'",
		"ALTER USER cassandra WITH PASSWORD 'x'",
		"DROP USER bob",
		"USE shop",
		"COPY shop.orders TO 'out.csv'",
		"SOURCE 'file.cql'",
		"CONSISTENCY ALL",
		"",
		"   ;  ",
	} {
		_, err := Classify(cql)
		assert.Error(t, err, "%q should be refused", cql)
	}
}

// TestOneStatementAtATime: a second statement after the first is the oldest
// way past a check on the first.
func TestOneStatementAtATime(t *testing.T) {
	for _, cql := range []string{
		"SELECT * FROM shop.orders; DROP TABLE shop.orders",
		"SELECT * FROM shop.orders;DROP TABLE shop.orders;",
		"SELECT * FROM shop.orders -- comment\n; TRUNCATE shop.orders",
		"BEGIN BATCH INSERT INTO shop.t (id) VALUES (1); APPLY BATCH; DROP TABLE shop.t",
	} {
		_, err := Classify(cql)
		assert.Error(t, err, "%q should be refused", cql)
	}

	// A trailing semicolon, or several, is still one statement.
	s, err := Classify("SELECT * FROM shop.orders;;")
	require.NoError(t, err)
	assert.Equal(t, "SELECT", s.Command)
}

// TestTextInsideStringsAndCommentsIsNotTheStatement: a semicolon or a keyword
// in a string or a comment does not split or change the statement.
func TestTextInsideStringsAndCommentsIsNotTheStatement(t *testing.T) {
	s, err := Classify("INSERT INTO shop.notes (id, body) VALUES (1, 'x; DROP TABLE shop.orders')")
	require.NoError(t, err)
	assert.Equal(t, "INSERT", s.Command)

	s, err = Classify("/* DROP TABLE shop.orders; */ SELECT * FROM shop.orders")
	require.NoError(t, err)
	assert.Equal(t, "SELECT", s.Command)

	s, err = Classify("-- DROP\n// TRUNCATE\nSELECT * FROM shop.orders")
	require.NoError(t, err)
	assert.Equal(t, "SELECT", s.Command)

	s, err = Classify("INSERT INTO shop.notes (id, body) VALUES (1, 'it''s; fine')")
	require.NoError(t, err)
	assert.Equal(t, "INSERT", s.Command)

	// And a comment cannot hide the start of the real statement.
	s, err = Classify("SELECT/**/ * FROM shop.orders")
	require.NoError(t, err)
	assert.Equal(t, "SELECT", s.Command)
}

// TestWhatCannotBeReadIsRefused rather than guessed at.
func TestWhatCannotBeReadIsRefused(t *testing.T) {
	for _, cql := range []string{
		"SELECT * FROM shop.orders WHERE body = 'not closed",
		`SELECT * FROM "shop.orders`,
		"/* not closed SELECT * FROM shop.orders",
		"SELECT * FROM ѕhop.orders",              // a Cyrillic s
		"ЅELECT * FROM shop.orders",              // a Cyrillic S
		"SELECT * FROM shop.orders WHERE id = 1", // a no-break space
		`SELECT * FROM ""."orders"`,
		"SELECT * FROM",
		"SELECT * FROM shop.",
		"SELECT count(*)",
	} {
		_, err := Classify(cql)
		assert.Error(t, err, "%q should be refused", cql)
	}

	// Outside ASCII is fine inside a string or a quoted name.
	_, err := Classify("INSERT INTO shop.notes (id, body) VALUES (1, 'café')")
	assert.NoError(t, err)
	_, err = Classify(`SELECT * FROM shop."Café"`)
	assert.NoError(t, err)
}

// TestAnUnqualifiedTableIsReportedAsOne, so the gate can refuse it: the
// session's keyspace would otherwise decide which table it is.
func TestAnUnqualifiedTableIsReportedAsOne(t *testing.T) {
	s, err := Classify("SELECT * FROM orders")
	require.NoError(t, err)
	assert.Equal(t, []Name{{Table: "orders"}}, s.Names)
	assert.False(t, s.Names[0].Qualified)
}

// TestABatchIsEveryStatementInIt.
func TestABatchIsEveryStatementInIt(t *testing.T) {
	s, err := Classify(`BEGIN UNLOGGED BATCH USING TIMESTAMP 1
		INSERT INTO shop.a (id) VALUES (1);
		UPDATE shop.b SET x = 1 WHERE id = 1;
		DELETE FROM shop.c WHERE id = 1;
		APPLY BATCH;`)
	require.NoError(t, err)
	assert.Equal(t, "BATCH", s.Command)
	assert.Equal(t, KindWrite, s.Kind)
	require.Len(t, s.Inner, 3)
	assert.Equal(t, []string{"INSERT", "UPDATE", "DELETE"},
		[]string{s.Inner[0].Command, s.Inner[1].Command, s.Inner[2].Command})
	assert.Equal(t, []Name{{"shop", "a", true}, {"shop", "b", true}, {"shop", "c", true}}, s.Names)

	// Only writes can be in one.
	for _, cql := range []string{
		"BEGIN BATCH TRUNCATE shop.a; APPLY BATCH",
		"BEGIN BATCH DROP TABLE shop.a; APPLY BATCH",
		"BEGIN BATCH GRANT ALL ON ALL KEYSPACES TO bob; APPLY BATCH",
		"BEGIN BATCH APPLY BATCH",
		"BEGIN BATCH INSERT INTO shop.a (id) VALUES (1);",
	} {
		_, err := Classify(cql)
		assert.Error(t, err, "%q should be refused", cql)
	}
}

// TestScansAreNoticed: ALLOW FILTERING and aggregates, and which columns the
// WHERE clause fixes.
func TestScansAreNoticed(t *testing.T) {
	s, err := Classify("SELECT * FROM shop.orders WHERE total > 10 allow Filtering")
	require.NoError(t, err)
	assert.True(t, s.AllowFiltering, "in any case")

	s, err = Classify("SELECT COUNT(*) FROM shop.orders")
	require.NoError(t, err)
	assert.True(t, s.Aggregate)
	assert.Empty(t, s.Restricted)

	s, err = Classify("SELECT max(total) FROM shop.orders WHERE customer = 'a' AND day IN ('x', 'y') AND total > 1")
	require.NoError(t, err)
	assert.True(t, s.Aggregate)
	assert.Equal(t, []string{"customer", "day"}, s.Restricted, "a range does not fix a column")

	s, err = Classify("SELECT * FROM shop.orders WHERE token(customer) > 5 LIMIT 10")
	require.NoError(t, err)
	assert.Empty(t, s.Restricted, "a token range is a range")

	s, err = Classify("SELECT * FROM shop.orders WHERE customer = 'a' LIMIT 10")
	require.NoError(t, err)
	assert.Equal(t, []string{"customer"}, s.Restricted)

	// A column called count is not the aggregate.
	s, err = Classify("SELECT count FROM shop.stats")
	require.NoError(t, err)
	assert.False(t, s.Aggregate)
}

// TestTheTableIsTheOneThePreferencesShow: every command the gate can return is
// one the window offers, and every command the window offers is one the gate
// returns.
func TestTheTableIsTheOneThePreferencesShow(t *testing.T) {
	seen := map[string]bool{}
	for _, cql := range []string{
		"SELECT * FROM a.b", "DESCRIBE a", "LIST ROLES",
		"INSERT INTO a.b (x) VALUES (1)", "UPDATE a.b SET x = 1", "DELETE FROM a.b",
		"BEGIN BATCH INSERT INTO a.b (x) VALUES (1); APPLY BATCH",
		"CREATE TABLE a.b (x int PRIMARY KEY)", "ALTER TABLE a.b ADD y int", "DROP TABLE a.b",
		"TRUNCATE a.b",
	} {
		s, err := Classify(cql)
		require.NoError(t, err, cql)
		command, ok := CommandNamed(s.Command)
		require.True(t, ok, "%s is not in the table", s.Command)
		assert.Equal(t, command.Kind, s.Kind, s.Command)
		seen[s.Command] = true
	}
	for _, c := range Commands {
		assert.True(t, seen[c.Name], "%s is in the table but nothing classifies as it", c.Name)
	}
	assert.Equal(t, []string{"SELECT", "DESCRIBE", "LIST"}, DefaultCommands())
}

// TestTheShellAsksAboutTheSameStatementsAsBefore, and a few more: a DELETE in
// a BATCH, and one behind a comment.
func TestTheShellAsksAboutTheSameStatementsAsBefore(t *testing.T) {
	for cql, want := range map[string]bool{
		"DROP TABLE shop.orders":                   true,
		"drop table orders":                        true,
		"ALTER TABLE shop.orders ADD x int":        true,
		"DELETE FROM shop.orders WHERE id = 1":     true,
		"TRUNCATE shop.orders":                     true,
		"REVOKE SELECT ON ALL KEYSPACES FROM bob":  true,
		"SELECT * FROM shop.orders":                false,
		"INSERT INTO shop.orders (id) VALUES (1)":  false,
		"CREATE TABLE shop.t (id int PRIMARY KEY)": false,
		"/* tidy up */ DROP TABLE shop.orders":     true,
		"BEGIN BATCH INSERT INTO shop.a (id) VALUES (1); DELETE FROM shop.b WHERE id = 1; APPLY BATCH": true,
		"BEGIN BATCH INSERT INTO shop.a (id) VALUES (1); APPLY BATCH":                                  false,
	} {
		assert.Equal(t, want, IsDangerousCommand(cql), cql)
	}
}
