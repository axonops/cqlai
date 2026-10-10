//go:build integration
// +build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPReadsATableWithAUDT: the server's session had no UDT registry, so a
// table with a UDT column could not be read through query at all.
func TestMCPReadsATableWithAUDT(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	for _, stmt := range []string{
		`CREATE TYPE IF NOT EXISTS test_mcp.address (street text, city text)`,
		`CREATE TABLE IF NOT EXISTS test_mcp.people (id int PRIMARY KEY, home frozen<address>)`,
		`INSERT INTO test_mcp.people (id, home) VALUES (1, {street: '1 Main St', city: 'Leeds'})`,
	} {
		require.NoError(t, env.Session.Query(stmt).Exec(), stmt)
	}
	t.Cleanup(func() { _ = env.Session.Query(`DROP TABLE IF EXISTS test_mcp.people`).Exec() })

	o := callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.people WHERE id = 1"})
	t.Logf("%s", o.Text)
	r := decode(t, o)
	require.Len(t, r.Rows, 1)
	home, ok := r.Rows[0]["home"].(map[string]any)
	require.True(t, ok, "%T %v", r.Rows[0]["home"], r.Rows[0]["home"])
	assert.Equal(t, "Leeds", home["city"])
}
