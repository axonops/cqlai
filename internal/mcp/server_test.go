package mcp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/policy"
)

// These tests talk to the server with the official Go SDK's client, which is
// linked into the test binary only. If the server drifts from the protocol,
// the SDK's client is what notices.

// noCluster is a source with a policy and no cluster.
type noCluster struct{ pol policy.Policy }

func (n noCluster) Policy() policy.Policy { return n.pol }
func (n noCluster) Env() ai.ToolEnv {
	return ai.ToolEnv{Policy: n.pol, NotConnected: "not connected to nowhere:9042"}
}

// serve starts a server on in-process pipes and connects the SDK's client to
// it. No cluster: what is tested is the protocol and what the policy offers.
func serve(t *testing.T, file *config.Config, flags policy.Flags) (*sdk.ClientSession, string) {
	t.Helper()

	pol, err := policy.Load(file, "", "", 5*time.Second, flags)
	require.NoError(t, err)

	auditPath := filepath.Join(t.TempDir(), "audit.log")
	audit, err := policy.OpenAudit(auditPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = audit.Close() })

	toServer, fromClient := io.Pipe()
	toClient, fromServer := io.Pipe()

	server := NewServer(noCluster{pol}, policy.NewLimiter(1000), audit, "test")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.Serve(ctx, toServer, fromServer)
		_ = fromServer.Close()
	}()

	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: toClient, Writer: fromClient}, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = session.Close()
		_ = fromClient.Close()
		cancel()
		<-done
	})
	return session, auditPath
}

func toolNames(t *testing.T, session *sdk.ClientSession) []string {
	t.Helper()
	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	return names
}

func call(t *testing.T, session *sdk.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(*sdk.TextContent)
	require.True(t, ok, "a text result")
	return text.Text, result.IsError
}

// TestTheSDKClientConnects and lists the tools, each with a schema it accepts.
func TestTheSDKClientConnects(t *testing.T) {
	session, _ := serve(t, &config.Config{}, policy.Flags{})

	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, tools.Tools)
	for _, tool := range tools.Tools {
		assert.NotEmpty(t, tool.Description, tool.Name)
		require.NotNil(t, tool.InputSchema, tool.Name)
		require.NotNil(t, tool.Annotations, tool.Name)
		assert.True(t, tool.Annotations.ReadOnlyHint, "%s only reads", tool.Name)
	}
	require.NoError(t, session.Ping(context.Background(), nil))
}

// TestTheToolsFollowThePermittedCommands: a tool the policy does not allow is
// not offered at all.
func TestTheToolsFollowThePermittedCommands(t *testing.T) {
	session, _ := serve(t, &config.Config{}, policy.Flags{})
	assert.ElementsMatch(t, []string{
		"fuzzy_search", "get_schema", "list_keyspaces", "list_tables",
		"connection_info", "describe", "query", "trace_query",
	}, toolNames(t, session))

	noSelect := []string{"DESCRIBE"}
	session, _ = serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &noSelect}}, policy.Flags{})
	names := toolNames(t, session)
	assert.NotContains(t, names, "query")
	assert.NotContains(t, names, "trace_query")
	assert.Contains(t, names, "describe")

	nothing := []string{}
	session, _ = serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &nothing}}, policy.Flags{})
	assert.Equal(t, []string{"connection_info"}, toolNames(t, session),
		"with nothing permitted, only what the server may do can be asked")

	// And none of the CHAT view's planner tools, ever.
	for _, name := range []string{"submit_query_plan", "user_selection", "not_enough_info", "not_relevant", "info"} {
		assert.NotContains(t, names, name)
	}
}

// TestAToolNotOfferedCannotBeCalled, and says the same as one that does not
// exist.
func TestAToolNotOfferedCannotBeCalled(t *testing.T) {
	nothing := []string{}
	session, _ := serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &nothing}}, policy.Flags{})

	text, isError := call(t, session, "query", map[string]any{"cql": "SELECT * FROM a.b"})
	assert.True(t, isError)
	assert.Equal(t, "refused: there is no tool called query", text)

	text, isError = call(t, session, "submit_query_plan", map[string]any{"operation": "DROP"})
	assert.True(t, isError)
	assert.Equal(t, "refused: there is no tool called submit_query_plan", text)
}

// TestTheGateAnswersThroughTheProtocol.
func TestTheGateAnswersThroughTheProtocol(t *testing.T) {
	session, auditPath := serve(t, &config.Config{MCP: &config.MCPConfig{
		Keyspaces: []string{"shop"},
	}}, policy.Flags{})

	for cql, why := range map[string]string{
		"DROP TABLE shop.orders":                  "DROP is not permitted",
		"SELECT * FROM billing.invoices":          "keyspace billing is not visible",
		"SELECT * FROM orders":                    "name the keyspace",
		"SELECT * FROM system_auth.roles":         "not visible",
		"SELECT * FROM shop.a; DROP TABLE shop.a": "one statement at a time",
	} {
		text, isError := call(t, session, "query", map[string]any{"cql": cql})
		assert.True(t, isError, cql)
		assert.Contains(t, text, why, cql)
	}

	// Arguments are checked before anything runs.
	text, isError := call(t, session, "describe", map[string]any{"kind": "table", "keyspace": "shop", "name": "x; DROP TABLE y"})
	assert.True(t, isError)
	assert.Contains(t, text, "letters, digits and underscores")

	// Every call is in the audit log, refusals too, with no values.
	time.Sleep(50 * time.Millisecond)
	data, err := os.ReadFile(auditPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 6)
	for _, line := range lines {
		var e policy.Entry
		require.NoError(t, json.Unmarshal([]byte(line), &e))
		assert.NotEmpty(t, e.Decision)
	}
	assert.Contains(t, string(data), `"decision":"refused"`)
}

// TestConnectionInfoSaysWhatIsAllowedAndNotWhatIsHidden.
func TestConnectionInfoSaysWhatIsAllowedAndNotWhatIsHidden(t *testing.T) {
	session, _ := serve(t, &config.Config{MCP: &config.MCPConfig{
		Deny: []string{"shop.secrets"},
	}}, policy.Flags{ReadOnly: true})

	text, isError := call(t, session, "connection_info", nil)
	require.False(t, isError, text)
	assert.NotContains(t, text, "secrets", "a hidden table is not named")

	var info struct {
		Policy struct {
			Permitted []string `json:"permitted_commands"`
			Hidden    bool     `json:"some_tables_hidden"`
		} `json:"policy"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &info))
	assert.Equal(t, []string{"SELECT", "DESCRIBE", "LIST"}, info.Policy.Permitted)
	assert.True(t, info.Policy.Hidden)
}

// TestTheServerRunsWithoutACluster: it starts, lists its tools, says it is
// not connected, and still refuses what it would refuse.
func TestTheServerRunsWithoutACluster(t *testing.T) {
	session, _ := serve(t, &config.Config{}, policy.Flags{})

	assert.Contains(t, toolNames(t, session), "query")

	text, isError := call(t, session, "connection_info", nil)
	require.False(t, isError, text)
	assert.Contains(t, text, `"connected": false`)
	assert.Contains(t, text, "not connected to nowhere:9042")

	text, isError = call(t, session, "query", map[string]any{"cql": "SELECT * FROM shop.orders"})
	assert.True(t, isError)
	assert.Contains(t, text, "not connected to nowhere:9042; the server tries again on the next call")

	text, _ = call(t, session, "query", map[string]any{"cql": "DROP TABLE shop.orders"})
	assert.Contains(t, text, "DROP is not permitted", "the gate answers first")
}
