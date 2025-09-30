package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// handleF2 handles F2 key - switch to query/history view
func (m *MainModel) handleF2() (*MainModel, tea.Cmd) {
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

// handleF3 handles F3 key - switch to table view
func (m *MainModel) handleF3() (*MainModel, tea.Cmd) {
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

// handleF4 handles F4 key - switch to trace view
func (m *MainModel) handleF4() (*MainModel, tea.Cmd) {
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

// handleF5 handles F5 key - switch to AI view
func (m *MainModel) handleF5() (*MainModel, tea.Cmd) {
	if m.viewMode != "ai" {
		m.viewMode = "ai"
		m.aiConversationActive = true

		// Clear any existing conversation ID when entering AI view via F5
		// This ensures we start fresh
		m.aiConversationID = ""

		// Initialize AI conversation input if not initialized
		// Check if Width is 0 as a proxy for uninitialized state
		if m.aiConversationInput.Width == 0 {
			input := textinput.New()
			input.Placeholder = ""
			input.Prompt = "> "
			input.CharLimit = 4096 // Increased to support long queries
			input.Width = m.historyViewport.Width - 2 // Reduced margin for better scrolling
			input.Focus()
			m.aiConversationInput = input

			// Initialize conversation viewport if needed
			if m.aiConversationViewport.Width == 0 {
				m.aiConversationViewport = viewport.New(m.historyViewport.Width, m.historyViewport.Height)
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

// handleF6 handles F6 key - toggle showing data types in table headers
func (m *MainModel) handleF6() (*MainModel, tea.Cmd) {
	if m.hasTable && m.viewMode == "table" {
		m.showDataTypes = !m.showDataTypes
		// Refresh the table view with new headers
		if len(m.lastTableData) > 0 {
			// Update ALL headers in the stored data (not just visible ones)
			if len(m.columnTypes) > 0 {
				// Process all columns in the header row
				for i := 0; i < len(m.lastTableData[0]) && i < len(m.columnTypes); i++ {
					// Parse the original header to extract base name and key indicators
					original := m.tableHeaders[i]

					// Remove any existing type info [...]
					if idx := strings.Index(original, " ["); idx != -1 {
						if endIdx := strings.Index(original[idx:], "]"); endIdx != -1 {
							original = original[:idx] + original[idx+endIdx+1:]
						}
					}

					// Extract base name and key indicator
					baseName := original
					keyIndicator := ""
					if strings.HasSuffix(original, " (PK)") {
						baseName = strings.TrimSuffix(original, " (PK)")
						keyIndicator = " (PK)"
					} else if strings.HasSuffix(original, " (C)") {
						baseName = strings.TrimSuffix(original, " (C)")
						keyIndicator = " (C)"
					}

					// Build the new header
					newHeader := baseName
					if m.showDataTypes && m.columnTypes[i] != "" {
						newHeader += " [" + m.columnTypes[i] + "]"
					}
					newHeader += keyIndicator

					// Update the actual stored data
					m.lastTableData[0][i] = newHeader
				}
			}

			// Clear cache and initial widths to force full rebuild with new headers
			m.cachedTableLines = nil
			m.initialColumnWidths = nil // Allow column widths to be recalculated
			// Refresh the table display with the updated data
			tableStr := m.formatTableForViewport(m.lastTableData)
			m.tableViewport.SetContent(tableStr)
		}
	}
	return m, nil
}