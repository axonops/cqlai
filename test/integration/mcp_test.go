//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	mcpserver "github.com/axonops/cqlai/internal/mcp"
	"github.com/axonops/cqlai/internal/policy"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
)

// The MCP server's tools against a real cluster: what a model is given, and
// what it is refused.
//
// Each test calls the tools the way the server does, through CallMCPTool, so
// what is tested is the policy and the tools together; the protocol itself is
// tested against the official SDK's client in internal/mcp.

// mcpEnv is a session on a keyspace of three tables - one read freely, one
// with a redacted column, one hidden - under the policy given.
func mcpEnv(t *testing.T, mcp *config.MCPConfig) ai.ToolEnv {
	t.Helper()
	sess, _, cleanup := getTestSession(t)
	t.Cleanup(cleanup)

	for _, stmt := range []string{
		`CREATE KEYSPACE IF NOT EXISTS test_mcp WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		`CREATE TABLE IF NOT EXISTS test_mcp.orders (customer text, day text, id int, total int, PRIMARY KEY ((customer, day), id))`,
		`CREATE TABLE IF NOT EXISTS test_mcp.customers (id int PRIMARY KEY, name text, email text)`,
		`CREATE TABLE IF NOT EXISTS test_mcp.secrets (id int PRIMARY KEY, note text)`,
		`INSERT INTO test_mcp.orders (customer, day, id, total) VALUES ('ann', 'd1', 1, 10)`,
		`INSERT INTO test_mcp.orders (customer, day, id, total) VALUES ('ann', 'd1', 2, 20)`,
		`INSERT INTO test_mcp.orders (customer, day, id, total) VALUES ('ann', 'd1', 3, 30)`,
		`INSERT INTO test_mcp.customers (id, name, email) VALUES (1, 'Ann', 'ann@example.com')`,
		`INSERT INTO test_mcp.secrets (id, note) VALUES (1, 'top secret')`,
	} {
		require.NoError(t, sess.Query(stmt).Exec(), stmt)
	}
	require.NoError(t, sess.AwaitSchemaAgreement(context.Background()))

	pol, err := policy.Load(&config.Config{MCP: mcp}, "", "", 10*time.Second, policy.Flags{})
	require.NoError(t, err)
	pol = pol.
		WithPartitionKey(func(ks, table string) []string { key, _ := sess.TableKey(ks, table); return key }).
		WithColumns(func(ks, table string) []string { _, columns := sess.TableKey(ks, table); return columns })

	// A schema cache of its own on this session, as the server has.
	schema, err := ai.NewSchemaTools(sess)
	require.NoError(t, err)
	mgr := session.NewManager(&config.Config{})
	return ai.ToolEnv{
		Policy:   pol,
		Session:  sess,
		Schema:   schema,
		Describe: mcpserver.Describer(sess, mgr),
	}
}

func defaultMCP() *config.MCPConfig {
	return &config.MCPConfig{
		Keyspaces: []string{"test_mcp", "system_views"},
		Deny:      []string{"test_mcp.secrets"},
		Redact:    []string{"test_mcp.customers.email"},
		MaxRows:   2,
	}
}

func callTool(t *testing.T, env ai.ToolEnv, name string, args map[string]any) ai.ToolOutcome {
	t.Helper()
	return ai.CallMCPTool(context.Background(), env, name, args)
}

type queryResult struct {
	Columns []db.PageColumn  `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Next    string           `json:"next_page_token"`
	Trace   map[string]any   `json:"trace"`
}

func decode(t *testing.T, o ai.ToolOutcome) queryResult {
	t.Helper()
	require.False(t, o.IsError, o.Text)
	var r queryResult
	require.NoError(t, json.Unmarshal([]byte(o.Text), &r))
	return r
}

// TestMCPQueryPages: a page at a time, at most the server's limit, and the
// token carries on from where the page stopped.
func TestMCPQueryPages(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	cql := "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'"

	first := decode(t, callTool(t, env, "query", map[string]any{"cql": cql}))
	require.Len(t, first.Rows, 2, "max rows is 2")
	require.NotEmpty(t, first.Next)
	assert.Equal(t, "int", first.Columns[2].Type)

	second := decode(t, callTool(t, env, "query", map[string]any{"cql": cql, "page_token": first.Next}))
	require.Len(t, second.Rows, 1)
	assert.EqualValues(t, 3, second.Rows[0]["id"])
	assert.Empty(t, second.Next)
}

// TestMCPHidesWhatThePolicyHides.
func TestMCPHidesWhatThePolicyHides(t *testing.T) {
	env := mcpEnv(t, defaultMCP())

	// A redacted value does not come back, by name or another way.
	customers := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.customers"}))
	require.Len(t, customers.Rows, 1)
	assert.Equal(t, policy.RedactedValue, customers.Rows[0]["email"])
	assert.Equal(t, "Ann", customers.Rows[0]["name"])
	for _, cql := range []string{
		"SELECT JSON * FROM test_mcp.customers",
		"SELECT email AS e FROM test_mcp.customers",
		"SELECT id FROM test_mcp.customers WHERE email = 'ann@example.com' ALLOW FILTERING",
	} {
		o := callTool(t, env, "query", map[string]any{"cql": cql})
		assert.True(t, o.Refused, cql)
		// Nothing from the table comes back: only the statement, the model's
		// own, handed back for the user to run - a guess written into a WHERE
		// clause included.
		assert.NotContains(t, o.Text, `"rows"`, cql)
		assert.Equal(t, cql+";", o.Statement, cql)
	}

	// A hidden table is refused, and left out of everything that lists.
	o := callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.secrets"})
	assert.True(t, o.Refused)
	for _, call := range []struct {
		tool string
		args map[string]any
	}{
		{"list_tables", map[string]any{"keyspace": "test_mcp"}},
		{"describe", map[string]any{"kind": "keyspace", "keyspace": "test_mcp"}},
		{"fuzzy_search", map[string]any{"query": "secrets"}},
		{"connection_info", map[string]any{}},
	} {
		o := callTool(t, env, call.tool, call.args)
		require.False(t, o.IsError, "%s: %s", call.tool, o.Text)
		// fuzzy_search repeats the term it was given, which is the model's
		// own word; what it must not do is find the table.
		assert.NotContains(t, o.Text, "test_mcp.secrets", "%s should not name the hidden table", call.tool)
		assert.NotContains(t, o.Text, "secrets (", "%s should not list the hidden table", call.tool)
		assert.NotContains(t, o.Text, "CREATE TABLE test_mcp.secrets", "%s should not describe the hidden table", call.tool)
		assert.NotContains(t, o.Text, "top secret")
	}

	// A keyspace not in the list is refused, and not listed.
	o = callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM system_auth.roles"})
	assert.True(t, o.Refused)
	o = callTool(t, env, "list_keyspaces", map[string]any{})
	assert.NotContains(t, o.Text, "system_auth")
	assert.Contains(t, o.Text, "test_mcp")
}

// TestMCPRefusesWhatIsNotPermitted, and the table is still there after.
func TestMCPRefusesWhatIsNotPermitted(t *testing.T) {
	env := mcpEnv(t, defaultMCP())

	for _, cql := range []string{
		"DROP TABLE test_mcp.orders",
		"TRUNCATE test_mcp.orders",
		"INSERT INTO test_mcp.orders (customer, day, id, total) VALUES ('x', 'y', 1, 1)",
		"SELECT * FROM test_mcp.orders; DROP TABLE test_mcp.orders",
		"SELECT COUNT(*) FROM test_mcp.orders",
	} {
		o := callTool(t, env, "query", map[string]any{"cql": cql})
		assert.True(t, o.Refused, cql)
	}

	still := decode(t, callTool(t, env, "query", map[string]any{
		"cql": "SELECT COUNT(*) FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'",
	}))
	assert.EqualValues(t, 3, still.Rows[0]["count"], "nothing was changed, and a count in one partition is allowed")
}

// TestMCPTracesAQuery.
func TestMCPTracesAQuery(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	r := decode(t, callTool(t, env, "trace_query", map[string]any{
		"cql": "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'",
	}))
	require.NotNil(t, r.Trace)
	if note, ok := r.Trace["note"]; ok {
		t.Skipf("the trace was not written in time: %v", note)
	}
	assert.NotEmpty(t, r.Trace["coordinator"])
	assert.NotEmpty(t, r.Trace["events"])
}

// TestMCPReadsVirtualTables, which are listed with their keys.
func TestMCPReadsVirtualTables(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	if !env.Session.IsVersion4OrHigher() {
		t.Skip("virtual tables arrived in Cassandra 4.0")
	}

	o := callTool(t, env, "list_tables", map[string]any{"keyspace": "system_views"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "clients (partition key: address; clustering: port)")

	r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT address, port FROM system_views.clients"}))
	assert.NotEmpty(t, r.Rows, "this test's own connection is a client")

	o = callTool(t, env, "describe", map[string]any{"kind": "table", "keyspace": "system_views", "name": "clients"})
	require.False(t, o.IsError, o.Text)
	assert.True(t, strings.Contains(o.Text, "VIRTUAL TABLE system_views.clients"), o.Text)
}

// TestMCPNodeStatus reads the node's virtual tables, and hides a setting that
// holds a secret.
func TestMCPNodeStatus(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	if !env.Session.IsVersion4OrHigher() {
		o := callTool(t, env, "node_status", map[string]any{"section": "thread_pools"})
		assert.True(t, o.IsError, "before 4.0 it says there is nothing to read")
		return
	}

	r := decode(t, callTool(t, env, "node_status", map[string]any{"section": "thread_pools", "filter": "read"}))
	require.NotEmpty(t, r.Rows)
	for _, row := range r.Rows {
		assert.Contains(t, strings.ToLower(row["name"].(string)), "read")
	}

	r = decode(t, callTool(t, env, "node_status", map[string]any{"section": "settings", "filter": "password"}))
	for _, row := range r.Rows {
		assert.Equal(t, policy.RedactedValue, row["value"], "%v", row["name"])
	}

	// Without system_views visible, it is refused like any other read.
	hidden := mcpEnv(t, &config.MCPConfig{Keyspaces: []string{"test_mcp"}})
	o := callTool(t, hidden, "node_status", map[string]any{"section": "clients"})
	assert.True(t, o.Refused, o.Text)
}

// TestMCPTableSizeAndRoles.
func TestMCPTableSizeAndRoles(t *testing.T) {
	env := mcpEnv(t, defaultMCP())

	o := callTool(t, env, "table_size", map[string]any{"keyspace": "test_mcp", "table": "orders"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, `"table": "orders"`)

	o = callTool(t, env, "table_size", map[string]any{"keyspace": "test_mcp", "table": "secrets"})
	assert.True(t, o.Refused)

	o = callTool(t, env, "list_roles", map[string]any{})
	if o.IsError && (strings.Contains(o.Text, "anonymous") || strings.Contains(o.Text, "logged in")) {
		t.Skip("the cluster has no authentication, so it has no roles to list")
	}
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "cassandra")
	assert.NotContains(t, o.Text, "salted_hash")

	o = callTool(t, env, "list_roles", map[string]any{"role": "cassandra' OR '1'='1"})
	assert.NotContains(t, o.Text, "salted_hash", "a quote in the role name stays inside the name")
}

// TestMCPSchemaResources: a keyspace's definition, without its hidden table.
func TestMCPSchemaResources(t *testing.T) {
	env := mcpEnv(t, defaultMCP())

	assert.Contains(t, ai.SchemaResources(env), "test_mcp")
	assert.NotContains(t, ai.SchemaResources(env), "system_auth")

	o := ai.ReadSchemaResource(env, "cql://schema/test_mcp")
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "CREATE TABLE test_mcp.orders")
	assert.NotContains(t, o.Text, "test_mcp.secrets")

	o = ai.ReadSchemaResource(env, "cql://schema/test_mcp/secrets")
	assert.True(t, o.IsError)
}

// TestMCPProposesWithoutRunning: a proposed change says what it will do,
// from the table's real key, and nothing happens to the table.
func TestMCPProposesWithoutRunning(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	var handed string
	env.Propose = func(h ai.Handover) { handed = h.Statement }

	o := callTool(t, env, "propose_change", map[string]any{"cql": "DELETE FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "the whole partition", "the key came from the table")
	assert.Equal(t, "DELETE FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1';", handed)

	o = callTool(t, env, "propose_change", map[string]any{"cql": "DROP TABLE test_mcp.orders"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "has not run this")

	still := decode(t, callTool(t, env, "query", map[string]any{
		"cql": "SELECT COUNT(*) FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'",
	}))
	assert.EqualValues(t, 3, still.Rows[0]["count"], "nothing was deleted, and the table is still there")
}

// TestMCPDoesNotTakeOverTheShellsSettings: a describe from the MCP server
// does not build the router's meta-command handler. Built from here, it bound
// the shell's TRACING ON to the server's session, and the shell's tracing
// never came on.
func TestMCPDoesNotTakeOverTheShellsSettings(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	before := router.GetMetaHandler()

	o := callTool(t, env, "describe", map[string]any{"kind": "table", "keyspace": "test_mcp", "name": "orders"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "CREATE TABLE test_mcp.orders")

	assert.Same(t, before, router.GetMetaHandler(), "the shell's handler is left as it was")
}
