package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/mcp"
)

// The MCP server inside the app, as the app shows it.

// mcpStatus is the status bar's MCP field: the port when serving, or OFF.
func mcpStatus(h *mcp.Host) string {
	if h == nil {
		return ""
	}
	if !h.Serving() {
		return "OFF"
	}
	return fmt.Sprintf(":%d", h.Port())
}

// showMCPServer writes where MCP is served, and what a client needs to
// connect, to the Console - where it can be selected and copied.
func (m *MainModel) showMCPServer() (*MainModel, tea.Cmd) {
	if m.mcpHost == nil {
		return m.report("MCP is not being served. Start cqlai with `cqlai mcp` to serve it to an AI client from this window, " +
			"or with `cqlai mcp --headless` for a client that starts cqlai itself.")
	}

	var b strings.Builder
	b.WriteString(m.mcpHost.Status())
	if m.mcpHost.Serving() {
		b.WriteString("\n\nAdd this to your MCP client's configuration. The token is kept in ")
		b.WriteString(mcp.TokenFile())
		b.WriteString(", readable only by you; anyone with it can use what this server allows.\n\n")
		b.WriteString(m.mcpHost.ClientConfig())
	}
	b.WriteString("\n\nWhat the model may do is set in PREFERENCES and CONNECT, under MCP SERVER.")
	return m.report(b.String())
}

// CloseMCP stops serving MCP, if the app was. It is called on the way out.
func (m *MainModel) CloseMCP() {
	if m.mcpHost != nil {
		m.mcpHost.Close()
	}
}
