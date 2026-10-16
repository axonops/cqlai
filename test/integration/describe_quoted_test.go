//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDescribeTakesQuotedNames: the quotes were trimmed off the ends of what
// was typed, so DESCRIBE TABLE "MyKeyspace"."TableA" looked for MyKeyspace"."TableA
// and k."_staging" for "_staging, and neither was found.
func TestDescribeTakesQuotedNames(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()
	mgr := session.NewManager(&config.Config{})
	defer func() { _ = sess.Query(`DROP KEYSPACE IF EXISTS "DescKs"`).Exec() }()

	for _, stmt := range []string{
		`DROP KEYSPACE IF EXISTS "DescKs"`,
		`CREATE KEYSPACE "DescKs" WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		`CREATE TABLE "DescKs"."TableA" (id int PRIMARY KEY, "order" int)`,
		`CREATE TABLE "DescKs"."_staging" (id int PRIMARY KEY)`,
		`CREATE TYPE "DescKs"."Addr" (street text)`,
		`CREATE INDEX "OrderIdx" ON "DescKs"."TableA" ("order")`,
	} {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}

	for command, want := range map[string]string{
		`DESCRIBE TABLE "DescKs"."TableA"`:   `"TableA"`,
		`DESCRIBE TABLE "DescKs"."_staging"`: `"_staging"`,
		`DESCRIBE "DescKs"."TableA"`:         `"TableA"`,
		`DESCRIBE KEYSPACE "DescKs"`:         `"DescKs"`,
		`DESCRIBE TYPE "DescKs"."Addr"`:      `"Addr"`,
		`DESCRIBE INDEX "DescKs"."OrderIdx"`: `"OrderIdx"`,
	} {
		result := router.ProcessCommand(command, sess, mgr)
		_, failed := result.(error)
		require.False(t, failed, "%s: %v", command, result)
		assert.Contains(t, described(result), want, command)
	}

	// The keyspace in use, named with capitals.
	router.ProcessCommand(`USE "DescKs"`, sess, mgr)
	assert.Contains(t, described(router.ProcessCommand(`DESCRIBE TABLE "TableA"`, sess, mgr)), `"TableA"`)
	assert.Contains(t, described(router.ProcessCommand(`DESCRIBE TYPE "Addr"`, sess, mgr)), `"Addr"`)

	// A name that is not there is said to be not there, rather than a panic.
	_, failed := router.ProcessCommand(`DESCRIBE "DescKs".nosuch`, sess, mgr).(error)
	assert.True(t, failed)
}

// described is what a DESCRIBE answered: its text, or the create statements of
// the rows a server-side DESCRIBE streams.
func described(result interface{}) string {
	streamed, ok := result.(db.StreamingQueryResult)
	if !ok {
		return fmt.Sprint(result)
	}
	defer streamed.Iterator.Close()
	out := ""
	for row := map[string]interface{}{}; db.ScanRow(streamed.Iterator, row); row = map[string]interface{}{} {
		out += fmt.Sprint(row["create_statement"]) + "\n"
	}
	return out
}
