package mcp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
func (n noCluster) SchemaVersion() string { return "" }
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
		"node_status", "table_size", "list_roles", "propose_change",
	}, toolNames(t, session))

	noSelect := []string{"DESCRIBE"}
	session, _ = serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &noSelect}}, policy.Flags{})
	names := toolNames(t, session)
	assert.NotContains(t, names, "query")
	assert.NotContains(t, names, "trace_query")
	assert.Contains(t, names, "describe")

	nothing := []string{}
	session, _ = serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &nothing}}, policy.Flags{})
	assert.Equal(t, []string{"connection_info", "propose_change"}, toolNames(t, session),
		"with nothing permitted, only what the server may do can be asked, and changes proposed")

	// And none of the CHAT view's planner tools, ever.
	for _, name := range []string{"submit_query_plan", "user_selection", "not_enough_info", "not_relevant", "info"} {
		assert.NotContains(t, names, name)
	}
}

// TestAToolNotOfferedHandsBackItsStatement: it does not run, and the caller
// gets the statement to give the user. A tool that does not exist, or one
// the server never offers, is still just refused.
func TestAToolNotOfferedHandsBackItsStatement(t *testing.T) {
	nothing := []string{}
	session, _ := serve(t, &config.Config{MCP: &config.MCPConfig{Permit: &nothing}}, policy.Flags{})

	text, isError := call(t, session, "query", map[string]any{"cql": "SELECT * FROM a.b"})
	assert.True(t, isError)
	assert.Contains(t, text, `"refused": "SELECT is not permitted on this server"`)
	assert.Contains(t, text, `"statement": "SELECT * FROM a.b;"`)

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
	}}, policy.Flags{})

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

// TestPromptsAndResourcesThroughTheSDK.
func TestPromptsAndResourcesThroughTheSDK(t *testing.T) {
	session, _ := serve(t, &config.Config{}, policy.Flags{})
	ctx := context.Background()

	prompts, err := session.ListPrompts(ctx, nil)
	require.NoError(t, err)
	require.Len(t, prompts.Prompts, 2)

	got, err := session.GetPrompt(ctx, &sdk.GetPromptParams{Name: "diagnose_query", Arguments: map[string]string{"cql": "SELECT * FROM shop.orders"}})
	require.NoError(t, err)
	require.Len(t, got.Messages, 1)
	assert.Contains(t, got.Messages[0].Content.(*sdk.TextContent).Text, "trace_query")

	resources, err := session.ListResources(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, resources.Resources, "no cluster, no keyspaces")

	templates, err := session.ListResourceTemplates(ctx, nil)
	require.NoError(t, err)
	require.Len(t, templates.ResourceTemplates, 1)
	assert.Equal(t, "cql://schema/{keyspace}/{table}", templates.ResourceTemplates[0].URITemplate)

	_, err = session.ReadResource(ctx, &sdk.ReadResourceParams{URI: "cql://schema/shop"})
	assert.Error(t, err, "no cluster to describe it from")
}

// changingSource is a source whose schema version changes when told to.
type changingSource struct {
	noCluster
	version func() string
}

func (c changingSource) SchemaVersion() string { return c.version() }

// TestASchemaChangeIsNotified over stdio, to the SDK's client.
func TestASchemaChangeIsNotified(t *testing.T) {
	pol, err := policy.Load(&config.Config{}, "", "", time.Second, policy.Flags{})
	require.NoError(t, err)
	version := "a"
	var mu sync.Mutex
	source := changingSource{noCluster{pol}, func() string { mu.Lock(); defer mu.Unlock(); return version }}

	audit, _ := policy.OpenAudit(policy.AuditOff)
	server := NewServer(source, policy.NewLimiter(100), audit, "test")
	toServer, fromClient := io.Pipe()
	toClient, fromServer := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx, toServer, fromServer) }()
	go server.WatchSchema(ctx, 10*time.Millisecond)

	changed := make(chan struct{}, 1)
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0"}, &sdk.ClientOptions{
		ResourceListChangedHandler: func(context.Context, *sdk.ResourceListChangedRequest) {
			select {
			case changed <- struct{}{}:
			default:
			}
		},
	})
	session, err := client.Connect(ctx, &sdk.IOTransport{Reader: toClient, Writer: fromClient}, nil)
	require.NoError(t, err)
	defer session.Close()

	time.Sleep(50 * time.Millisecond) // the watcher has seen "a"
	mu.Lock()
	version = "b"
	mu.Unlock()

	select {
	case <-changed:
	case <-time.After(2 * time.Second):
		t.Fatal("the client was not told the schema changed")
	}
}

// TestEveryReadIsWithinTheRate: a resource read and a prompt each reach the
// cluster, so each counts as a call; and the rate is the settings' own, not
// the one the limiter was made with.
func TestEveryReadIsWithinTheRate(t *testing.T) {
	session, _ := serve(t, &config.Config{MCP: &config.MCPConfig{MaxCallsPerMinute: 2}}, policy.Flags{})
	ctx := context.Background()

	_, err := session.ReadResource(ctx, &sdk.ReadResourceParams{URI: "cql://schema/shop"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "calls a minute", "the first is within the rate")

	_, err = session.GetPrompt(ctx, &sdk.GetPromptParams{Name: "diagnose_query", Arguments: map[string]string{"cql": "SELECT * FROM shop.orders"}})
	require.NoError(t, err, "the second is within the rate")

	_, err = session.GetPrompt(ctx, &sdk.GetPromptParams{Name: "diagnose_query", Arguments: map[string]string{"cql": "SELECT * FROM shop.orders"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "more than 2 calls a minute")

	_, err = session.ReadResource(ctx, &sdk.ReadResourceParams{URI: "cql://schema/shop"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "more than 2 calls a minute")

	text, isError := call(t, session, "list_keyspaces", map[string]any{})
	assert.True(t, isError)
	assert.Contains(t, text, "more than 2 calls a minute")
}
