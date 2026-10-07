package ai

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/policy"
)

func mcpPolicy(t *testing.T, mcp *config.MCPConfig) policy.Policy {
	t.Helper()
	p, err := policy.Load(&config.Config{MCP: mcp}, "", "", time.Second, policy.Flags{})
	require.NoError(t, err)
	return p
}

// TestTheChatViewIsOfferedWhatItAlwaysWas: adding MCP tools to the registry
// must not change what the CHAT view's planner sees.
func TestTheChatViewIsOfferedWhatItAlwaysWas(t *testing.T) {
	var names []string
	for _, def := range GetCommonToolDefinitions() {
		names = append(names, def.Name)
	}
	assert.Equal(t, []string{
		"fuzzy_search", "get_schema", "list_keyspaces", "list_tables", "user_selection",
		"not_enough_info", "not_relevant", "submit_query_plan", "info",
	}, names)

	// And the CHAT path refuses an MCP-only tool rather than running it
	// without a policy.
	result := ExecuteToolCall("query", map[string]any{"cql": "SELECT * FROM a.b"})
	require.Error(t, result.Error)
}

// TestEveryToolOfferedCanBeRun: a definition marked for MCP has a parameter
// parser and is a valid name, so it cannot be offered and then fail as
// unknown.
func TestEveryToolOfferedCanBeRun(t *testing.T) {
	for _, def := range allToolDefinitions() {
		if !def.MCP {
			continue
		}
		name := ParseToolName(def.Name)
		require.True(t, name.IsValid(), def.Name)
		_, err := ParseToolParams(name, []byte(`{}`))
		assert.NoError(t, err, "%s has no parameter parser", def.Name)
		for _, required := range def.Required {
			assert.Contains(t, def.Parameters, required, "%s requires %s and does not describe it", def.Name, required)
		}
	}
}

// TestToolsFollowThePolicy.
func TestToolsFollowThePolicy(t *testing.T) {
	names := func(p policy.Policy) []string {
		var out []string
		for _, def := range MCPToolDefinitions(p) {
			out = append(out, def.Name)
		}
		return out
	}

	assert.ElementsMatch(t, []string{
		"fuzzy_search", "get_schema", "list_keyspaces", "list_tables",
		"connection_info", "describe", "query", "trace_query",
		"node_status", "table_size", "list_roles", "propose_change",
	}, names(mcpPolicy(t, nil)))

	selectOnly := []string{"SELECT"}
	assert.ElementsMatch(t, []string{"connection_info", "query", "trace_query", "node_status", "table_size", "propose_change"},
		names(mcpPolicy(t, &config.MCPConfig{Permit: &selectOnly})))

	listOnly := []string{"LIST"}
	assert.ElementsMatch(t, []string{"connection_info", "list_roles", "propose_change"},
		names(mcpPolicy(t, &config.MCPConfig{Permit: &listOnly})))

	// propose_change runs nothing, so it is always there.
	var zero policy.Policy
	assert.Equal(t, []string{"connection_info", "propose_change"}, names(zero), "the zero policy offers nothing that touches the cluster")
}

// TestAHiddenTableIsLeftOutOfADescription, and nothing else is.
func TestAHiddenTableIsLeftOutOfADescription(t *testing.T) {
	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"shop.secrets"}})
	described := strings.Join([]string{
		"CREATE KEYSPACE shop WITH replication = {'class': 'SimpleStrategy'} AND durable_writes = true",
		"CREATE TABLE shop.orders (\n    id int PRIMARY KEY\n) WITH comment = ''",
		"CREATE TABLE shop.secrets (\n    id int PRIMARY KEY\n) WITH comment = ''",
		"CREATE INDEX secrets_note ON shop.secrets (note)",
		"CREATE TABLE shop.secrets_archive (\n    id int PRIMARY KEY\n)",
		"",
	}, ";\n")

	kept := withoutHidden(p, described)
	assert.Contains(t, kept, "CREATE KEYSPACE shop")
	assert.Contains(t, kept, "CREATE TABLE shop.orders")
	assert.Contains(t, kept, "CREATE TABLE shop.secrets_archive", "a table whose name starts the same is another table")
	assert.NotContains(t, kept, "CREATE TABLE shop.secrets (")
	assert.NotContains(t, kept, "secrets_note", "nor an index on it")
}

// TestAPageTokenBelongsToItsStatement.
func TestAPageTokenBelongsToItsStatement(t *testing.T) {
	token := encodePageToken([]byte{1, 2, 3}, "SELECT * FROM shop.orders")
	state, err := decodePageToken(token, "  SELECT * FROM shop.orders ")
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3}, state)

	_, err = decodePageToken(token, "SELECT * FROM shop.customers")
	assert.Error(t, err)
	_, err = decodePageToken("not-a-token", "SELECT * FROM shop.orders")
	assert.Error(t, err)

	assert.Empty(t, encodePageToken(nil, "SELECT 1"), "no token when there is nothing after the page")
}

// TestRowsAreShapedForTheModel: redacted, converted to JSON values, and cut.
func TestRowsAreShapedForTheModel(t *testing.T) {
	p := mcpPolicy(t, &config.MCPConfig{
		Redact:        []string{"shop.customers.email"},
		MaxValueBytes: 10,
	})
	page := db.QueryPageResult{Rows: []map[string]interface{}{
		{"id": 1, "email": "ann@example.com", "bio": "ééééééééééééé", "tags": []string{"aaaaa", "bbbbb"}},
		{"id": 2, "email": nil, "bio": "short", "tags": nil},
	}}

	rows, cut := shapeRows(p, "shop", "customers", page)
	require.Len(t, rows, 2)
	assert.Equal(t, policy.RedactedValue, rows[0]["email"])
	assert.Equal(t, policy.RedactedValue, rows[1]["email"], "a NULL is hidden too: whether there is a value is a value")
	assert.Equal(t, "short", rows[1]["bio"])
	assert.Equal(t, []string{"bio", "tags"}, cut)

	bio := rows[0]["bio"].(string)
	assert.True(t, strings.HasSuffix(bio, "…"))
	kept := strings.TrimSuffix(bio, "…")
	full := page.Rows[0]["bio"].(string)
	assert.LessOrEqual(t, len(kept), 10)
	assert.True(t, strings.HasPrefix(full, kept), "cut on a character boundary")
}

// TestTheClusterToolsCheckTheirArguments before anything runs.
func TestTheClusterToolsCheckTheirArguments(t *testing.T) {
	assert.NoError(t, NodeStatusParams{Section: "thread_pools"}.Validate())
	assert.Error(t, NodeStatusParams{Section: "passwords"}.Validate())
	assert.Error(t, TableSizeParams{Keyspace: "shop", Table: "orders; DROP"}.Validate())
	assert.NoError(t, TableSizeParams{Keyspace: "shop", Table: "orders"}.Validate())
}

// TestAHiddenTableIsNotMeasured, and nothing runs without a cluster.
func TestAHiddenTableIsNotMeasured(t *testing.T) {
	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"shop.secrets"}})
	o := CallMCPTool(t.Context(), ToolEnv{Policy: p}, "table_size", map[string]any{"keyspace": "shop", "table": "orders"})
	assert.True(t, o.IsError)
	assert.Contains(t, o.Text, "not connected")

	o = tableSize(t.Context(), ToolEnv{Policy: p}, TableSizeParams{Keyspace: "shop", Table: "secrets"})
	assert.True(t, o.Refused)
}

// TestThePromptsFollowThePolicy, and use the shell's own instructions.
func TestThePromptsFollowThePolicy(t *testing.T) {
	names := func(p policy.Policy) []string {
		var out []string
		for _, d := range MCPPrompts(p) {
			out = append(out, d.Name)
		}
		return out
	}
	assert.Equal(t, []string{"review_table", "diagnose_query"}, names(mcpPolicy(t, nil)))
	selectOnly := []string{"SELECT"}
	assert.Equal(t, []string{"diagnose_query"}, names(mcpPolicy(t, &config.MCPConfig{Permit: &selectOnly})))

	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"shop.secrets"}})
	text, err := GetMCPPrompt(ToolEnv{Policy: p}, "diagnose_query", map[string]string{"cql": "SELECT * FROM shop.orders"})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(text, TraceInstructions))
	assert.Contains(t, text, "SELECT * FROM shop.orders")

	text, err = GetMCPPrompt(ToolEnv{Policy: p}, "review_table", map[string]string{"table": "shop.orders"})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(text, SchemaInstructions))
	assert.Contains(t, text, "describe tool", "without a cluster the model is asked to fetch the definition")

	_, err = GetMCPPrompt(ToolEnv{Policy: p}, "review_table", map[string]string{"table": "shop.secrets"})
	assert.Error(t, err, "not for a hidden table")
	_, err = GetMCPPrompt(ToolEnv{Policy: p}, "review_table", map[string]string{"table": "orders"})
	assert.Error(t, err, "keyspace.table")
	_, err = GetMCPPrompt(ToolEnv{Policy: p}, "review_table", nil)
	assert.Error(t, err)
}

// TestResourcesFollowThePolicy: none without DESCRIBE or a cluster, and a
// hidden table's cannot be read.
func TestResourcesFollowThePolicy(t *testing.T) {
	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"shop.secrets"}})
	assert.Empty(t, SchemaResources(ToolEnv{Policy: p}))
	assert.Equal(t, "cql://schema/shop/orders", SchemaResourceURI("shop", "orders"))

	o := ReadSchemaResource(ToolEnv{Policy: p}, "cql://schema/shop/orders")
	assert.True(t, o.IsError)
	assert.Contains(t, o.Text, "not connected")

	o = ReadSchemaResource(ToolEnv{Policy: p}, "file:///etc/passwd")
	assert.True(t, o.IsError)

	selectOnly := []string{"SELECT"}
	o = ReadSchemaResource(ToolEnv{Policy: mcpPolicy(t, &config.MCPConfig{Permit: &selectOnly})}, "cql://schema/shop")
	assert.True(t, o.Refused, "DESCRIBE is not permitted")
}

// TestAProposedChangeIsNotRun: it comes back with where it would run and what
// it would do, and the shell is handed it unrun.
func TestAProposedChangeIsNotRun(t *testing.T) {
	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"shop.secrets"}})
	var handed string
	env := ToolEnv{Policy: p, Propose: func(h Handover) { handed = h.Statement }}

	o := CallMCPTool(t.Context(), env, "propose_change", map[string]any{
		"cql": "  DROP TABLE shop.orders  ", "reason": "it is no longer used",
	})
	require.False(t, o.IsError, o.Text)
	assert.True(t, o.Proposed)
	assert.Equal(t, "DROP TABLE shop.orders;", handed)

	var out struct {
		Statement    string   `json:"statement"`
		Changes      string   `json:"changes"`
		Implications []string `json:"implications"`
		NotRun       string   `json:"not_run"`
		Reason       string   `json:"reason"`
		InCqlai      string   `json:"in_cqlai"`
	}
	require.NoError(t, json.Unmarshal([]byte(o.Text), &out))
	assert.Equal(t, "DROP TABLE shop.orders;", out.Statement)
	assert.Equal(t, "schema", out.Changes)
	assert.Contains(t, out.NotRun, "has not run this")
	assert.Equal(t, "it is no longer used", out.Reason)
	assert.NotEmpty(t, out.InCqlai)
	require.NotEmpty(t, out.Implications)
	assert.Contains(t, out.Implications[0], "deletes the table and all its data")

	// A proposal still cannot reach what is hidden, and is not a way to run
	// a read.
	o = CallMCPTool(t.Context(), env, "propose_change", map[string]any{"cql": "TRUNCATE shop.secrets"})
	assert.True(t, o.Refused)
	o = CallMCPTool(t.Context(), env, "propose_change", map[string]any{"cql": "SELECT * FROM shop.orders"})
	assert.True(t, o.Refused)
}

// TestARefusalHandsBackTheStatement: refused by the settings, the caller
// gets the statement for the user to run, and the shell is handed it.
func TestARefusalHandsBackTheStatement(t *testing.T) {
	selectOnly := []string{"SELECT"}
	p := mcpPolicy(t, &config.MCPConfig{
		Permit:    &selectOnly,
		Keyspaces: []string{"shop"},
		Redact:    []string{"shop.customers.email"},
	})
	var handed Handover
	env := ToolEnv{Policy: p, Propose: func(h Handover) { handed = h }}

	for cql, why := range map[string]string{
		"SELECT * FROM billing.invoices":                        "keyspace billing is not visible",
		"SELECT * FROM shop.orders WHERE x > 1 ALLOW FILTERING": "ALLOW FILTERING is not permitted",
		"SELECT JSON * FROM shop.customers":                     "hidden columns",
		"DESCRIBE TABLE shop.orders":                            "DESCRIBE is not permitted",
	} {
		handed = Handover{}
		o := CallMCPTool(t.Context(), env, "query", map[string]any{"cql": cql})
		require.True(t, o.Refused, cql)
		assert.Contains(t, o.Text, why, cql)
		assert.Contains(t, o.Text, `"statement": "`+cql+`;"`, cql)
		assert.Equal(t, cql+";", handed.Statement, cql)
		assert.Contains(t, handed.Refused, why, cql)
	}

	// Tools the settings do not offer hand back what they would have run.
	for _, c := range []struct {
		tool string
		args map[string]any
		want string
	}{
		{"describe", map[string]any{"kind": "table", "keyspace": "shop", "name": "orders"}, "DESCRIBE TABLE shop.orders;"},
		{"list_roles", map[string]any{"role": "o'brien"}, "LIST ALL PERMISSIONS OF 'o''brien';"},
		{"get_schema", map[string]any{"keyspace": "shop", "table": "orders"}, "DESCRIBE TABLE shop.orders;"},
	} {
		o := CallMCPTool(t.Context(), env, c.tool, c.args)
		require.True(t, o.Refused, c.tool)
		assert.Equal(t, c.want, o.Statement, c.tool)
		assert.Contains(t, o.Text, "is not permitted on this server", c.tool)
	}

	// Not what is never offered, nor what is not one statement.
	for _, cql := range []string{
		"GRANT ALL ON ALL KEYSPACES TO bob",
		"SELECT * FROM shop.a; DROP TABLE shop.a",
	} {
		handed = Handover{}
		o := CallMCPTool(t.Context(), env, "query", map[string]any{"cql": cql})
		assert.True(t, o.Refused, cql)
		assert.NotContains(t, o.Text, `"statement"`, cql)
		assert.Empty(t, handed.Statement, cql)
	}
}

// TestResultsAreNotEscapedForHTML: a statement or a value with <, > or & in
// it reads, and copies, as it was.
func TestResultsAreNotEscapedForHTML(t *testing.T) {
	o := jsonOutcome(map[string]any{"statement": "SELECT * FROM a.b WHERE x > 1 AND y < 2", "v": "a & b"})
	assert.Contains(t, o.Text, "x > 1 AND y < 2")
	assert.Contains(t, o.Text, "a & b")
}
