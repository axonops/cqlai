package ui

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/mcp"
)

// A change the MCP client proposes is put in front of the user and never run
// by cqlai: it waits in the prompt, and running it is always confirmed.

func proposalModel(t *testing.T) *MainModel {
	t.Helper()
	m := helpModel()
	m.input.Focus()
	m.showProposal(mcpProposalMsg{
		Statement:    "INSERT INTO shop.orders (id) VALUES (1);",
		Implications: []string{"INSERT overwrites a row with the same primary key"},
	})
	return m
}

// TestAProposalWaitsInThePrompt, with what it will do in the Console.
func TestAProposalWaitsInThePrompt(t *testing.T) {
	m := proposalModel(t)
	assert.Equal(t, "INSERT INTO shop.orders (id) VALUES (1);", m.input.Value())
	assert.Contains(t, m.fullHistoryContent, "cqlai has not run it")
	assert.Contains(t, m.fullHistoryContent, "INSERT overwrites a row")

	// Something being typed is not overwritten.
	m = helpModel()
	m.input.SetValue("SELECT * FROM shop.orders")
	m.showProposal(mcpProposalMsg{Statement: "DROP TABLE shop.orders;"})
	assert.Equal(t, "SELECT * FROM shop.orders", m.input.Value())
	assert.Contains(t, m.fullHistoryContent, "The prompt has text in it")
}

// TestAProposalIsAlwaysConfirmed, even an INSERT, and even with confirmation
// turned off in the settings: nobody typed it.
func TestAProposalIsAlwaysConfirmed(t *testing.T) {
	m := proposalModel(t)
	require.Nil(t, m.sessionManager, "no settings asking for confirmation")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, ModalConfirmDangerous, m.modal.Type)
	assert.Equal(t, "Run the statement from the MCP client?", m.modal.Title)
	assert.Equal(t, 0, m.modal.Selected, "Cancel is what Enter presses next")
}

// TestAnEditedProposalIsTheUsers: changed, it is what the user typed, and is
// confirmed only as a typed statement would be.
func TestAnEditedProposalIsTheUsers(t *testing.T) {
	m := proposalModel(t)
	m.input.SetValue("INSERT INTO shop.orders (id) VALUES (2);")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotEqual(t, "Run the statement from the MCP client?", m.modal.Title)
}

// TestUntickingSelectStopsQueriesAtOnce: saved in PREFERENCES while the
// shell is serving MCP, it applies to the next call.
func TestUntickingSelectStopsQueriesAtOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"mcp": {"auditLog": "-"}}`), 0o600))

	host := mcp.StartHost(mcp.Options{ConfigFile: path, Port: freePrefPort(t), RequestTimeout: 1}, "test")
	defer host.Close()
	require.True(t, host.Serving(), host.Status())
	host.Use(config.Config{Host: "127.0.0.1", Port: 1}, "")

	m := prefModel(t, &config.Config{SourcePath: path, MCP: &config.MCPConfig{AuditLog: "-"}})
	m.mcpHost = host
	m = toggleCommand(t, m, "SELECT")
	m.savePreferences()
	assert.Contains(t, m.fullHistoryContent, "uses its settings from now on")

	o := ai.CallMCPTool(context.Background(), host.Env(), "query", map[string]any{"cql": "SELECT * FROM shop.orders"})
	assert.True(t, o.Refused, o.Text)
	assert.Contains(t, o.Text, "SELECT is not permitted on this server")
}

func freePrefPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

// TestARefusedStatementIsHandedOver: in the prompt, with why the server did
// not run it, and confirmed before it runs.
func TestARefusedStatementIsHandedOver(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m.showProposal(mcpProposalMsg{Statement: "SELECT * FROM shop.orders;", Refused: "SELECT is not permitted on this server"})

	assert.Equal(t, "SELECT * FROM shop.orders;", m.input.Value())
	assert.Contains(t, m.fullHistoryContent, "was not permitted to run it (SELECT is not permitted on this server)")
	assert.NotContains(t, m.fullHistoryContent, "What it will do")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, "Run the statement from the MCP client?", m.modal.Title)
}
