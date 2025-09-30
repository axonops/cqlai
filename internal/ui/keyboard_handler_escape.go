package ui

import (
	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEscapeKey handles the ESC key press with its many contexts
func (m *MainModel) handleEscapeKey() (*MainModel, tea.Cmd) {
	// If AI conversation is active and processing, cancel it
	if m.aiConversationActive && m.aiProcessing {
		m.aiProcessing = false
		m.aiConversationHistory += m.styles.MutedText.Render("\n(Cancelled)") + "\n"
		m.aiConversationViewport.SetContent(m.aiConversationHistory)
		m.aiConversationViewport.GotoBottom()
		return m, nil
	}

	// If AI selection modal is showing, handle it
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		if m.aiSelectionModal.InputMode {
			// Exit input mode
			m.aiSelectionModal.InputMode = false
			m.aiSelectionModal.CustomInput = ""
		} else {
			// Cancel the selection
			m.aiSelectionModal.Active = false
			return m, func() tea.Msg {
				return AISelectionResultMsg{Cancelled: true}
			}
		}
		return m, nil
	}

	// If in multi-line mode, exit it
	if m.multiLineMode {
		m.multiLineMode = false
		m.multiLineBuffer = nil
		m.input.Placeholder = "Enter CQL command..."
		m.input.Reset()

		// Add cancellation message to history
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Multi-line mode cancelled.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}

	// If in history search mode, exit it
	if m.historySearchMode {
		m.historySearchMode = false
		m.historySearchQuery = ""
		m.historySearchResults = []string{}
		m.historySearchIndex = 0
		return m, nil
	}

	// If history modal is showing, close it
	if m.showHistoryModal {
		m.showHistoryModal = false
		m.historyModalIndex = 0
		m.historyModalScrollOffset = 0
		return m, nil
	}

	// Toggle navigation mode in table/trace views - HIGH PRIORITY
	// Check this early because it's a common operation
	logger.DebugfToFile("ESC", "ESC key: viewMode=%s, hasTable=%v, hasTrace=%v, navigationMode=%v",
		m.viewMode, m.hasTable, m.hasTrace, m.navigationMode)

	if (m.viewMode == "table" && m.hasTable) || (m.viewMode == "trace" && m.hasTrace) {
		// If we're in navigation mode AND have more data to page through, offer to cancel paging
		if m.navigationMode && m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			logger.DebugfToFile("ESC", "In nav mode with paging - cancelling paging")
			// Clear the "more data" state and reset to normal nav mode
			m.slidingWindow.hasMoreData = false
			m.slidingWindow.streamingResult = nil
			// Stay in navigation mode but update the placeholder
			m.input.Placeholder = "[NAV MODE] ↑↓←→=scroll | j/k=line | d/u=½page | g/G=top/bottom | </>=10cols | ESC=exit"
			// Add a message to history to indicate paging was cancelled
			m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Paging cancelled. Showing partial results.")
			m.updateHistoryWrapping()
			m.historyViewport.GotoBottom()
			return m, nil
		}

		// Normal navigation mode toggle
		m.navigationMode = !m.navigationMode
		logger.DebugfToFile("ESC", "Toggling navigation mode to: %v", m.navigationMode)
		if m.navigationMode {
			m.input.Blur()
			m.input.Placeholder = "[NAV MODE] ↑↓←→=scroll | j/k=line | d/u=½page | g/G=top/bottom | </>=10cols | ESC=exit"
		} else {
			m.input.Focus()
			m.input.Placeholder = "Enter CQL command (ESC for navigation mode)..."
		}
		return m, nil
	}

	// If modal is showing, close it
	if m.modal.Type != ModalNone {
		m.modal = Modal{Type: ModalNone}
		m.input.Placeholder = "Enter CQL command..."
		m.input.Reset()

		// Add cancellation message to history
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}

	// If showing completions, hide them
	if m.showCompletions {
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		return m, nil
	}

	// MOVED: Navigation mode toggle should be checked BEFORE paging state
	// because ESC to toggle navigation is more important than canceling paging

	// If we're in the middle of paging through results AND not in navigation mode, offer to clear paging
	// (Don't do this if we're just trying to toggle navigation mode)
	if m.viewMode == "table" && m.slidingWindow != nil && m.slidingWindow.hasMoreData && !m.navigationMode {
		// Only clear paging if we're not toggling navigation
		logger.DebugfToFile("ESC", "Clearing paging state instead of toggling nav")
		// Clear the "more data" state and reset input placeholder
		m.slidingWindow.hasMoreData = false
		m.slidingWindow.streamingResult = nil
		m.input.Placeholder = "Enter CQL command..."
		// Add a message to history to indicate paging was cancelled
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Paging cancelled. Showing partial results.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}

	// Cancel any confirmation first
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	// Ask for confirmation
	m.confirmExit = true
	m.input.SetValue("")
	m.input.Placeholder = "Really exit? (Ctrl+C/Ctrl+D again to confirm, any other key to cancel)"
	return m, nil
}