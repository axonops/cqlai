package ai

import (
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
	}, names(mcpPolicy(t, nil)))

	selectOnly := []string{"SELECT"}
	assert.ElementsMatch(t, []string{"connection_info", "query", "trace_query"},
		names(mcpPolicy(t, &config.MCPConfig{Permit: &selectOnly})))

	var zero policy.Policy
	assert.Equal(t, []string{"connection_info"}, names(zero), "the zero policy offers nothing that touches the cluster")
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
