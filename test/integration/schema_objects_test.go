//go:build integration
// +build integration

package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/db"
)

// TestKeyspaceObjectsListsAndDescribesEachKind: the SCHEMA view's groups,
// against a cluster. Views and functions are off unless cassandra.yaml turns
// them on, so a kind the cluster will not create is skipped, not failed.
func TestKeyspaceObjectsListsAndDescribesEachKind(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()

	const ks = "test_schema_objects"
	require.NoError(t, sess.Query(`DROP KEYSPACE IF EXISTS `+ks).Exec())
	require.NoError(t, sess.Query(`CREATE KEYSPACE `+ks+` WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec())
	// Deferred after the session's own cleanup, so it runs first, while the
	// session is still open.
	defer func() { _ = sess.Query(`DROP KEYSPACE IF EXISTS ` + ks).Exec() }()

	must := []string{
		`CREATE TYPE ` + ks + `.address (street text, city text)`,
		`CREATE TABLE ` + ks + `.users (id int PRIMARY KEY, name text, email text, home frozen<address>)`,
		`CREATE INDEX users_name_idx ON ` + ks + `.users (name)`,
	}
	for _, stmt := range must {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}
	optional := map[string]string{
		db.KindViews: `CREATE MATERIALIZED VIEW ` + ks + `.users_by_email AS SELECT * FROM ` + ks + `.users
			WHERE email IS NOT NULL AND id IS NOT NULL PRIMARY KEY (email, id)`,
		db.KindFunctions: `CREATE FUNCTION ` + ks + `.plus (state int, v int) CALLED ON NULL INPUT RETURNS int
			LANGUAGE java AS 'return (state == null ? 0 : state) + (v == null ? 0 : v);'`,
	}
	created := map[string]bool{db.KindTypes: true, db.KindIndexes: true}
	for _, kind := range []string{db.KindViews, db.KindFunctions} {
		if err := sess.Query(optional[kind]).Exec(); err != nil {
			t.Logf("%s not created here: %v", kind, err)
			continue
		}
		created[kind] = true
	}
	if created[db.KindFunctions] {
		if err := sess.Query(`CREATE AGGREGATE ` + ks + `.total (int) SFUNC plus STYPE int INITCOND 0`).Exec(); err == nil {
			created[db.KindAggregates] = true
		}
	}
	require.NoError(t, sess.AwaitSchemaAgreement(context.Background()))

	// The MCP policy hides and redacts a view as its base table.
	if created[db.KindViews] {
		assert.Equal(t, "users", sess.ViewBase(ks, "users_by_email"))
		assert.Equal(t, "", sess.ViewBase(ks, "users"), "a table is not a view")
	}

	objects := sess.KeyspaceObjects(ks)
	want := map[string]string{
		db.KindTypes:      "address",
		db.KindIndexes:    "users_name_idx",
		db.KindViews:      "users_by_email",
		db.KindFunctions:  "plus",
		db.KindAggregates: "total",
	}
	starts := map[string]string{
		db.KindTypes:      "CREATE TYPE",
		db.KindIndexes:    "CREATE INDEX",
		db.KindViews:      "CREATE MATERIALIZED VIEW",
		db.KindFunctions:  "CREATE FUNCTION",
		db.KindAggregates: "CREATE AGGREGATE",
	}
	for kind, name := range want {
		if !created[kind] {
			continue
		}
		assert.Equal(t, []string{name}, objects[kind], kind)

		text, err := sess.DescribeSchemaObject(kind, ks, name)
		require.NoError(t, err, kind)
		assert.True(t, strings.HasPrefix(strings.TrimSpace(text), starts[kind]), "%s: %s", kind, text)
		assert.Contains(t, text, name, kind)
	}
	assert.Empty(t, objects[db.KindTriggers], "no triggers here")

	// A keyspace with none of them has no groups but its tables.
	for _, kind := range db.SchemaObjectKinds {
		assert.Empty(t, sess.KeyspaceObjects("system_auth")[kind], kind)
	}
}
