package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/mcp"
	"github.com/axonops/cqlai/internal/router"
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
		b.WriteString("\n\nAdd this to your MCP client's configuration.")
		if m.mcpHost.TokenRequired() {
			b.WriteString(" The token is kept in ")
			b.WriteString(mcp.TokenFile())
			b.WriteString(", readable only by you; anyone with it can use what this server allows.")
		} else if !m.mcpHost.ClientCertRequired() {
			b.WriteString(" No token is asked for: anything on this machine that can reach the address can use what this server allows. Require token, in PREFERENCES, asks for one.")
		}
		if m.mcpHost.ClientCertRequired() {
			b.WriteString(" The client also has to show a certificate the TLS client CA signed.")
		}
		b.WriteString("\n\n")
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

// mcpProposalMsg is a change a model proposed through the MCP server.
type mcpProposalMsg mcp.Proposal

// waitForProposal waits for the next proposed change.
func waitForProposal(h *mcp.Host) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-h.Proposals()
		if !ok {
			return nil
		}
		return mcpProposalMsg(p)
	}
}

// showProposal puts a proposed change in front of the user, unrun: in the
// Console with what it will do, and in the prompt when nothing is being typed
// there. Running it is the user's to do, by pressing Enter - and a dangerous
// statement is still confirmed as any typed one is.
func (m *MainModel) showProposal(p mcpProposalMsg) (*MainModel, tea.Cmd) {
	var b strings.Builder
	if p.Refused != "" {
		b.WriteString("The MCP client asked for this, and the MCP server was not permitted to run it (" +
			p.Refused + "). It is yours to run, if you choose.\n\n")
		b.WriteString(p.Statement)
	} else {
		b.WriteString("The MCP client proposes this change. cqlai has not run it.\n\n")
		b.WriteString(p.Statement)
		b.WriteString("\n\nWhat it will do:")
		for _, line := range p.Implications {
			b.WriteString("\n  - " + line)
		}
	}

	// The prompt is one line: it turns a newline into a space, and on one
	// line a comment would take everything after it. So it gets the
	// statement without its comments, on one line - and that is what is
	// remembered as the proposal, so Enter on it is recognised and confirmed.
	prompt := proposalLine(p.Statement)
	switch {
	case strings.TrimSpace(m.input.Value()) != "":
		b.WriteString("\n\nThe prompt has text in it, so the statement is only here: copy it to run it.")
	case m.input.CharLimit > 0 && len([]rune(prompt)) > m.input.CharLimit:
		// The prompt would cut it short, and a statement cut short is not one
		// to leave a press of Enter away.
		b.WriteString("\n\nIt is too long for the prompt, so it is only here: copy it to run it.")
	default:
		m.input.SetValue(prompt)
		m.input.CursorEnd()
		m.proposed = prompt
		b.WriteString("\n\nIt is in the prompt. Read it, and press Enter to run it, or clear it.")
	}
	updated, _ := m.report(b.String())
	return updated, waitForProposal(m.mcpHost)
}

// isProposal reports whether a command is the change the MCP client put in
// the prompt, as it was put there. Spacing is not compared: the prompt can
// change it, and a proposal that slipped past this would run unconfirmed.
func (m *MainModel) isProposal(command string) bool {
	if m.proposed == "" {
		return false
	}
	same := func(s string) string { return strings.TrimRight(strings.Join(strings.Fields(s), " "), "; ") }
	return same(command) == same(m.proposed)
}

// proposalLine is a statement as it goes into the prompt: without its
// comments, on one line. Only what the prompt changes anyway - a newline or a
// tab to a space - so the spacing inside a string is left as it was proposed.
func proposalLine(statement string) string {
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(router.StripComments(statement))
}
