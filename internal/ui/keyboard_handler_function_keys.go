package ui

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// The views, one function each.
//
// Named for what they show rather than for the key that reaches them. They were
// handleF2 to handleF6, and moving a tab along the line meant renaming three
// functions and everything that called them - which is a rename with nothing
// behind it, since none of them has anything to do with a particular key.

// showConsole switches to the running transcript.
func (m *MainModel) showConsole() (*MainModel, tea.Cmd) {
	if m.viewMode != "history" {
		m.viewMode = "history"
		// If in AI conversation mode, also deactivate it
		if m.aiConversationActive {
			m.aiConversationActive = false
			m.aiConversationInput.SetValue("")
			m.aiProcessing = false
			m.input.Placeholder = "Enter CQL command..."
			m.input.SetValue("")
			m.input.Focus()
		}
	}
	return m, nil
}

// showResults switches to the last query's output.
func (m *MainModel) showResults() (*MainModel, tea.Cmd) {
	if m.viewMode != "table" {
		m.viewMode = "table"
		// If in AI conversation mode, also deactivate it
		if m.aiConversationActive {
			m.aiConversationActive = false
			m.aiConversationInput.SetValue("")
			m.aiProcessing = false
			m.input.SetValue("")
		}
		// Update placeholder to show ESC hint
		if m.hasTable {
			m.input.Placeholder = "Enter CQL command (ESC for navigation mode)..."
		} else {
			m.input.Placeholder = "Enter CQL command..."
		}
		m.input.Focus()
	}
	return m, nil
}

// showTrace switches to the query trace.
func (m *MainModel) showTrace() (*MainModel, tea.Cmd) {
	if m.viewMode != "trace" {
		m.viewMode = "trace"
		// If in AI conversation mode, also deactivate it
		if m.aiConversationActive {
			m.aiConversationActive = false
			m.aiConversationInput.SetValue("")
			m.aiProcessing = false
			m.input.SetValue("")
		}
		// Refresh the trace view if we have trace data
		if m.hasTrace {
			m.refreshTraceView()
			// Update placeholder to show ESC hint
			m.input.Placeholder = "Enter CQL command (ESC for navigation mode)..."
		} else {
			m.input.Placeholder = "Enter CQL command..."
		}
		m.input.Focus()
	}
	return m, nil
}

// showSchema switches to the schema browser, or asks the cluster again when it
// is already showing.
func (m *MainModel) showSchema() (*MainModel, tea.Cmd) {
	if !m.connected() {
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Not connected.") +
			m.styles.MutedText.Render(" There is no schema to browse.") + "\n"
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}
	return m.openSchema()
}

// showChat switches to the AI conversation.
func (m *MainModel) showChat() (*MainModel, tea.Cmd) {
	// Say why rather than doing nothing. A key that silently does nothing is
	// how someone concludes the build is broken.
	if !m.aiAvailable() {
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("AI is not configured.") +
			m.styles.MutedText.Render(" Set a provider and key in cqlai.json, or export ANTHROPIC_API_KEY.") + "\n"
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}

	if m.viewMode != "ai" {
		m.viewMode = "ai"
		m.aiConversationActive = true

		// Clear any existing conversation ID when entering the chat view
		// This ensures we start fresh
		m.aiConversationID = ""

		// Initialize AI conversation input if not initialized
		// Check if Width is 0 as a proxy for uninitialized state
		if m.aiConversationInput.Width() == 0 {
			input := textinput.New()
			input.Placeholder = ""
			input.Prompt = "> "
			input.CharLimit = 4096                        // Increased to support long queries
			input.SetWidth(m.historyViewport.Width() - 2) // Reduced margin for better scrolling
			input.Focus()
			m.aiConversationInput = input

			// Initialize conversation viewport if needed
			if m.aiConversationViewport.Width() == 0 {
				m.aiConversationViewport = viewport.New(viewport.WithWidth(m.historyViewport.Width()), viewport.WithHeight(m.historyViewport.Height()))
			}
		} else {
			// If already initialized, just clear and focus
			m.aiConversationInput.SetValue("")
			m.aiConversationInput.Focus()
		}

		// Always rebuild conversation to ensure proper wrapping with current viewport width
		// (header is added automatically by rebuildAIConversation if messages are empty)
		m.rebuildAIConversation()
	}
	return m, nil
}
