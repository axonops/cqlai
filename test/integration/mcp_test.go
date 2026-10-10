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
	pol = mcpserver.SessionPolicy(pol, sess)

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

	// With every keyspace allowed but some denied, it still reads them: the
	// statement is its own and the secrets are hidden.
	all := mcpEnv(t, &config.MCPConfig{Deny: []string{"test_mcp.secrets"}})
	r = decode(t, callTool(t, all, "node_status", map[string]any{"section": "thread_pools"}))
	assert.NotEmpty(t, r.Rows)
}

// TestMCPDenyIsNotUndoneThroughTheSystemKeyspaces: with every keyspace allowed
// but some denied, a SELECT on the system keyspaces would list what deny
// hides, so it is refused.
func TestMCPDenyIsNotUndoneThroughTheSystemKeyspaces(t *testing.T) {
	permit := []string{"SELECT", "DESCRIBE"}
	env := mcpEnv(t, &config.MCPConfig{Permit: &permit, Deny: []string{"test_mcp.secrets"}})

	for _, cql := range []string{
		"SELECT keyspace_name, table_name FROM system_schema.tables",
		"SELECT * FROM system_traces.sessions",
		"SELECT * FROM system.size_estimates WHERE keyspace_name = 'test_mcp' AND table_name = 'secrets'",
	} {
		o := callTool(t, env, "query", map[string]any{"cql": cql})
		assert.True(t, o.Refused, "%s: %s", cql, o.Text)
		assert.NotContains(t, o.Text, "secrets\"", cql)
	}

	// The tables deny does not hide are still there.
	o := callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.orders LIMIT 1"})
	assert.False(t, o.IsError, o.Text)
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

	// A permission on a hidden table would name it.
	for _, stmt := range []string{
		`CREATE ROLE IF NOT EXISTS test_mcp_reader`,
		`GRANT SELECT ON test_mcp.orders TO test_mcp_reader`,
		`GRANT SELECT ON test_mcp.secrets TO test_mcp_reader`,
	} {
		require.NoError(t, env.Session.Query(stmt).Exec(), stmt)
	}
	t.Cleanup(func() { _ = env.Session.Query(`DROP ROLE IF EXISTS test_mcp_reader`).Exec() })
	o = callTool(t, env, "list_roles", map[string]any{"role": "test_mcp_reader"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, "test_mcp.orders")
	assert.NotContains(t, o.Text, "secrets")
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

// TestMCPSystemSchemaRowsLeaveOutWhatIsHidden: with system_schema listed, its
// rows about a hidden keyspace or table are left out, and the rest are there.
func TestMCPSystemSchemaRowsLeaveOutWhatIsHidden(t *testing.T) {
	permit := []string{"SELECT"}
	env := mcpEnv(t, &config.MCPConfig{
		Permit:    &permit,
		Keyspaces: []string{"test_mcp", "system_schema", "system"},
		Deny:      []string{"test_mcp.secrets"},
		AutoFetch: true,
	})

	r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM system_schema.tables"}))
	seen := map[string]bool{}
	for _, row := range r.Rows {
		ks, table := row["keyspace_name"].(string), row["table_name"].(string)
		seen[ks+"."+table] = true
		assert.NotEqual(t, "test_mcp.secrets", ks+"."+table, "a denied table is not described")
		assert.Contains(t, []string{"test_mcp", "system_schema", "system", "system_auth"}, ks,
			"the keyspaces listed, and system_auth's schema, which is the same everywhere")
	}
	assert.True(t, seen["test_mcp.orders"], "what is visible is still there")
	assert.True(t, seen["system_schema.columns"])

	r = decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM system_schema.columns WHERE keyspace_name = 'test_mcp' AND table_name = 'secrets'"}))
	assert.Empty(t, r.Rows, "asking for a denied table by name finds nothing")

	// The columns that say what a row is about have to come back.
	for _, cql := range []string{
		"SELECT table_name, column_name FROM system_schema.columns",
		"SELECT table_name AS keyspace_name, column_name FROM system_schema.columns",
		"SELECT COUNT(*) FROM system_schema.columns",
	} {
		o := callTool(t, env, "query", map[string]any{"cql": cql})
		assert.True(t, o.Refused, "%s: %s", cql, o.Text)
	}

	// A system table whose rows are not about other keyspaces is read as it is.
	r = decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM system.local"}))
	assert.Len(t, r.Rows, 1)
}

// TestMCPPageSizeIsTheClients: a page size asked for is used, whatever the
// setting, which is only the page size used when none is asked for.
func TestMCPPageSizeIsTheClients(t *testing.T) {
	env := mcpEnv(t, &config.MCPConfig{MaxRows: 1, Keyspaces: []string{"test_mcp"}})

	r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'"}))
	assert.Len(t, r.Rows, 1, "the setting's page size")
	assert.NotEmpty(t, r.Next)

	r = decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'", "page_size": 1000}))
	assert.Len(t, r.Rows, 3, "the client's page size, above the setting")
	assert.Empty(t, r.Next)
}

// TestMCPAutoFetchReturnsEveryRow, paging through them, with no next page.
func TestMCPAutoFetchReturnsEveryRow(t *testing.T) {
	env := mcpEnv(t, &config.MCPConfig{MaxRows: 1, AutoFetch: true, Keyspaces: []string{"test_mcp"}})

	r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'"}))
	assert.Len(t, r.Rows, 3, "every row, a page of one at a time")
	assert.Empty(t, r.Next)

	o := callTool(t, env, "trace_query", map[string]any{"cql": "SELECT * FROM test_mcp.orders WHERE customer = 'ann' AND day = 'd1'"})
	require.False(t, o.IsError, o.Text)
	assert.Contains(t, o.Text, `"trace"`)
}

// TestMCPOpenForADemo: every keyspace and the system ones, scans and auto
// fetch on, nothing denied. The whole of system_schema.columns comes back in
// one call, traced, with any selection.
func TestMCPOpenForADemo(t *testing.T) {
	env := mcpEnv(t, &config.MCPConfig{SystemKeyspaces: true, AllowScans: true, AutoFetch: true})

	o := callTool(t, env, "trace_query", map[string]any{"cql": "SELECT * FROM system_schema.columns"})
	require.False(t, o.IsError, o.Text)
	var r queryResult
	require.NoError(t, json.Unmarshal([]byte(o.Text), &r))
	assert.Greater(t, len(r.Rows), 100, "more than one page, in one call")
	assert.Empty(t, r.Next)

	r = decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT table_name, column_name FROM system_schema.columns"}))
	assert.Greater(t, len(r.Rows), 100)

	r = decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT COUNT(*) FROM system_schema.columns"}))
	assert.Len(t, r.Rows, 1)

	o = callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM system_auth.roles"})
	assert.True(t, o.Refused, "system_auth's data, never")
}

// TestMCPAViewIsHiddenAsItsBaseTableIs: a materialized view holds its base
// table's rows, so the deny entry and the redaction for the table hold for
// it. Views are off unless cassandra.yaml turns them on; without them this is
// skipped.
func TestMCPAViewIsHiddenAsItsBaseTableIs(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	for _, stmt := range []string{
		`CREATE MATERIALIZED VIEW IF NOT EXISTS test_mcp.secrets_by_note AS SELECT * FROM test_mcp.secrets
			WHERE note IS NOT NULL AND id IS NOT NULL PRIMARY KEY (note, id)`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS test_mcp.customers_by_name AS SELECT * FROM test_mcp.customers
			WHERE name IS NOT NULL AND id IS NOT NULL PRIMARY KEY (name, id)`,
	} {
		if err := env.Session.Query(stmt).Exec(); err != nil {
			t.Skipf("views are not on here: %v", err)
		}
	}
	t.Cleanup(func() {
		_ = env.Session.Query(`DROP MATERIALIZED VIEW IF EXISTS test_mcp.secrets_by_note`).Exec()
		_ = env.Session.Query(`DROP MATERIALIZED VIEW IF EXISTS test_mcp.customers_by_name`).Exec()
	})
	require.NoError(t, env.Session.AwaitSchemaAgreement(context.Background()))

	o := callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.secrets_by_note WHERE note = 'top secret'"})
	assert.True(t, o.Refused, o.Text)

	// The view's rows can take a moment to be built.
	var rows []map[string]any
	require.Eventually(t, func() bool {
		r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT * FROM test_mcp.customers_by_name WHERE name = 'Ann'"}))
		rows = r.Rows
		return len(rows) == 1
	}, 20*time.Second, 500*time.Millisecond)
	assert.Equal(t, policy.RedactedValue, rows[0]["email"])
	assert.Equal(t, "Ann", rows[0]["name"])
}

// TestMCPQueryHidesSecretSettings: query hides a setting that holds a secret
// as node_status does, whether or not Cassandra hides it, and refuses the
// ways round that.
func TestMCPQueryHidesSecretSettings(t *testing.T) {
	env := mcpEnv(t, defaultMCP())
	const name = "client_encryption_options.keystore_password"

	for _, cql := range []string{
		"SELECT name, value FROM system_views.settings WHERE name = '" + name + "'",
		"SELECT value FROM system_views.settings WHERE name = '" + name + "'",
	} {
		r := decode(t, callTool(t, env, "query", map[string]any{"cql": cql}))
		require.Len(t, r.Rows, 1, cql)
		assert.Equal(t, policy.RedactedValue, r.Rows[0]["value"], cql)
	}

	r := decode(t, callTool(t, env, "query", map[string]any{"cql": "SELECT name, value FROM system_views.settings WHERE name = 'cluster_name'"}))
	require.Len(t, r.Rows, 1)
	assert.NotEqual(t, policy.RedactedValue, r.Rows[0]["value"], "a setting that is not a secret is shown")

	for _, cql := range []string{
		"SELECT JSON * FROM system_views.settings WHERE name = '" + name + "'",
		"SELECT name FROM system_views.settings WHERE value = 'x' ALLOW FILTERING",
	} {
		assert.True(t, callTool(t, env, "query", map[string]any{"cql": cql}).Refused, cql)
	}
}
