package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyboardInput handles keyboard input events
func (m *MainModel) handleKeyboardInput(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		// If in history search mode, exit it
		if m.historySearchMode {
			m.historySearchMode = false
			m.historySearchQuery = ""
			m.historySearchResults = []string{}
			m.historySearchIndex = 0
			m.historySearchScrollOffset = 0
			return m, nil
		}
		// If modal is showing, close it
		if m.modal.Type != ModalNone {
			m.modal = Modal{Type: ModalNone}
			m.input.Placeholder = "Enter CQL command..."
			m.input.Reset()

			// Add cancellation message to history
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("Command cancelled.")
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()
			return m, nil
		}
		// If there's text in the input, clear it. Otherwise ask for confirmation.
		if m.input.Value() != "" {
			m.input.Reset()
			// Also clear any completions and confirmation
			m.showCompletions = false
			m.completions = []string{}
			m.completionIndex = -1
			m.historyIndex = -1
			m.confirmExit = false
			return m, nil
		}
		// If already confirming, exit. Otherwise show confirmation.
		if m.confirmExit {
			return m, tea.Quit
		}
		m.confirmExit = true
		m.input.SetValue("")
		m.input.Placeholder = "Really exit? (Ctrl+C/Ctrl+D again to confirm, any other key to cancel)"
		return m, nil

	case tea.KeyCtrlD:
		// If confirming exit, quit. Otherwise show confirmation.
		if m.confirmExit {
			return m, tea.Quit
		}
		m.confirmExit = true
		m.input.SetValue("")
		m.input.Placeholder = "Really exit? (Ctrl+C/Ctrl+D again to confirm, any other key to cancel)"
		return m, nil

	case tea.KeyCtrlR:
		// Toggle history search mode
		if !m.historySearchMode {
			// Enter history search mode
			m.historySearchMode = true
			m.historySearchQuery = ""
			m.historySearchIndex = 0

			// Search with empty query shows all history
			if m.historyManager != nil {
				m.historySearchResults = m.historyManager.SearchHistory("")
			} else {
				// Fallback to in-memory history
				m.historySearchResults = make([]string, len(m.commandHistory))
				copy(m.historySearchResults, m.commandHistory)
			}

			// Reset scroll offset
			m.historySearchScrollOffset = 0
		} else {
			// Exit history search mode
			m.historySearchMode = false
			m.historySearchQuery = ""
			m.historySearchResults = []string{}
			m.historySearchIndex = 0
		}
		return m, nil

	case tea.KeyEsc:
		// If AI info request modal is showing, handle it
		if m.aiInfoReplyModal != nil && m.aiInfoReplyModal.Active {
			// Cancel the info request
			m.aiInfoReplyModal.Active = false
			return m, func() tea.Msg {
				return AIInfoResponseMsg{Cancelled: true}
			}
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
		// If AI modal is showing, close it
		if m.showAIModal {
			m.showAIModal = false
			m.aiModal = AIModal{}
			m.aiConversationID = ""
			m.input.Placeholder = "Enter CQL command..."
			return m, nil
		}
		// If in multi-line mode, exit it
		if m.multiLineMode {
			m.multiLineMode = false
			m.multiLineBuffer = nil
			m.input.Placeholder = "Enter CQL command..."
			m.input.Reset()

			// Add cancellation message to history
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("Multi-line mode cancelled.")
			m.historyViewport.SetContent(historyContent)
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
		// If modal is showing, close it
		if m.modal.Type != ModalNone {
			m.modal = Modal{Type: ModalNone}
			m.input.Placeholder = "Enter CQL command..."
			m.input.Reset()

			// Add cancellation message to history
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("Command cancelled.")
			m.historyViewport.SetContent(historyContent)
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

	case tea.KeyTab:
		// If modal is showing, navigate choices
		if m.modal.Type != ModalNone {
			m.modal.NextChoice()
			return m, nil
		}
		return m.handleTabKey()

	case tea.KeyF1:
		// F1 to switch between history and table view
		if m.hasTable {
			if m.viewMode == "history" {
				m.viewMode = "table"
			} else {
				m.viewMode = "history"
			}
		}
		return m, nil

	case tea.KeyF2:
		// F2 to toggle showing data types in table headers
		if m.hasTable && m.viewMode == "table" {
			m.showDataTypes = !m.showDataTypes
			// Refresh the table view with new headers
			if m.lastTableData != nil && len(m.lastTableData) > 0 {
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

						// Refresh the table display with the updated data
					tableStr := m.formatTableForViewport(m.lastTableData)
					m.tableViewport.SetContent(tableStr)
			}
		}
		return m, nil

	case tea.KeySpace:
		return m.handleSpaceKey(msg)

	case tea.KeyPgUp:
		return m.handlePageUp(msg)

	case tea.KeyPgDown:
		return m.handlePageDown(msg)

	case tea.KeyUp:
		logger.DebugfToFile("AI", "KeyUp pressed. showAIModal=%v, aiSelectionModal.Active=%v, historySearchMode=%v", 
			m.showAIModal, m.aiSelectionModal != nil && m.aiSelectionModal.Active, m.historySearchMode)
		// If AI selection modal is showing, navigate options
		if m.aiSelectionModal != nil && m.aiSelectionModal.Active && !m.aiSelectionModal.InputMode {
			m.aiSelectionModal.PrevOption()
			return m, nil
		}
		// If in history search mode, navigate search results
		if m.historySearchMode {
			if m.historySearchIndex > 0 {
				m.historySearchIndex--
				// Adjust scroll offset if needed
				if m.historySearchIndex < m.historySearchScrollOffset {
					m.historySearchScrollOffset = m.historySearchIndex
				}
			}
			return m, nil
		}
		return m.handleUpArrow(msg)

	case tea.KeyDown:
		// If AI selection modal is showing, navigate options
		if m.aiSelectionModal != nil && m.aiSelectionModal.Active && !m.aiSelectionModal.InputMode {
			m.aiSelectionModal.NextOption()
			return m, nil
		}
		// If in history search mode, navigate search results
		if m.historySearchMode {
			if m.historySearchIndex < len(m.historySearchResults)-1 {
				m.historySearchIndex++
				// Adjust scroll offset if needed
				if m.historySearchIndex >= m.historySearchScrollOffset+10 {
					m.historySearchScrollOffset = m.historySearchIndex - 9
				}
			}
			return m, nil
		}
		return m.handleDownArrow(msg)

	case tea.KeyLeft:
		return m.handleLeftArrow(msg)

	case tea.KeyRight:
		return m.handleRightArrow(msg)

	case tea.KeyEnter:
		// If in history search mode, select the current entry
		if m.historySearchMode {
			if len(m.historySearchResults) > 0 && m.historySearchIndex < len(m.historySearchResults) {
				// Set the input value to the selected history entry
				m.input.SetValue(m.historySearchResults[m.historySearchIndex])
				// Exit history search mode
				m.historySearchMode = false
				m.historySearchQuery = ""
				m.historySearchResults = []string{}
				m.historySearchIndex = 0
			}
			return m, nil
		}
		// If history modal is showing, select the current entry
		if m.showHistoryModal && len(m.commandHistory) > 0 {
			// Get the selected command (remember we're showing newest first)
			selectedIndex := len(m.commandHistory) - 1 - m.historyModalIndex
			if selectedIndex >= 0 && selectedIndex < len(m.commandHistory) {
				m.input.SetValue(m.commandHistory[selectedIndex])
			}
			// Close the modal
			m.showHistoryModal = false
			m.historyModalIndex = 0
			m.historyModalScrollOffset = 0
			return m, nil
		}
		return m.handleEnterKey()

	default:
		// Handle AI info request modal text input
		if m.aiInfoReplyModal != nil && m.aiInfoReplyModal.Active {
			var cmd tea.Cmd
			m.aiInfoReplyModal, cmd = m.aiInfoReplyModal.Update(msg)
			return m, cmd
		}
		// Handle AI selection modal 'i' key for custom input
		if m.aiSelectionModal != nil && m.aiSelectionModal.Active && !m.aiSelectionModal.InputMode {
			if msg.String() == "i" || msg.String() == "I" {
				m.aiSelectionModal.ToggleInputMode()
				return m, nil
			}
		}
		// Handle AI selection modal text input when in input mode
		if m.aiSelectionModal != nil && m.aiSelectionModal.Active && m.aiSelectionModal.InputMode {
			// Handle character input
			if msg.Type == tea.KeyRunes {
				m.aiSelectionModal.CustomInput += string(msg.Runes)
				return m, nil
			}
			// Handle backspace
			if msg.Type == tea.KeyBackspace && len(m.aiSelectionModal.CustomInput) > 0 {
				m.aiSelectionModal.CustomInput = m.aiSelectionModal.CustomInput[:len(m.aiSelectionModal.CustomInput)-1]
				return m, nil
			}
		}
		// Handle AI modal info input (when INFO operation is showing)
		if m.showAIModal && m.aiModal.State == AIModalStatePreview && 
			m.aiModal.Plan != nil && m.aiModal.Plan.Operation == "INFO" {
			// Handle character input
			if msg.Type == tea.KeyRunes {
				runes := string(msg.Runes)
				if len(m.aiModal.FollowUpInput) < 500 {
					m.aiModal.FollowUpInput = m.aiModal.FollowUpInput[:m.aiModal.CursorPosition] +
						runes +
						m.aiModal.FollowUpInput[m.aiModal.CursorPosition:]
					m.aiModal.CursorPosition += len(runes)
				}
				return m, nil
			}
			// Handle backspace
			if msg.Type == tea.KeyBackspace && m.aiModal.CursorPosition > 0 {
				m.aiModal.FollowUpInput = m.aiModal.FollowUpInput[:m.aiModal.CursorPosition-1] +
					m.aiModal.FollowUpInput[m.aiModal.CursorPosition:]
				m.aiModal.CursorPosition--
				return m, nil
			}
		}
		
		// Handle AI modal keyboard input
		if m.showAIModal && m.aiModal.State == AIModalStatePreview {





			// Regular modal - handle 'P' key for toggling plan/CQL view
			if msg.String() == "p" || msg.String() == "P" {
				m.aiModal.ToggleView()
				return m, nil
			}
		}

		// Cancel exit confirmation on any other key
		if m.confirmExit {
			m.confirmExit = false
			m.input.Placeholder = "Enter CQL command..."
		}

		// If in history search mode, handle typing for search query
		if m.historySearchMode {
			// Update search query based on key press
			switch msg.String() {
			case "backspace", "delete":
				if len(m.historySearchQuery) > 0 {
					m.historySearchQuery = m.historySearchQuery[:len(m.historySearchQuery)-1]
				}
			default:
				// Add character to search query if it's a printable character
				if len(msg.Runes) > 0 && len(m.historySearchQuery) < 100 {
					m.historySearchQuery += string(msg.Runes)
				}
			}

			// Update search results
			if m.historyManager != nil {
				m.historySearchResults = m.historyManager.SearchHistory(m.historySearchQuery)
			} else {
				// Fallback to in-memory history search
				m.historySearchResults = []string{}
				queryLower := strings.ToLower(m.historySearchQuery)
				for i := len(m.commandHistory) - 1; i >= 0; i-- {
					if strings.Contains(strings.ToLower(m.commandHistory[i]), queryLower) {
						m.historySearchResults = append(m.historySearchResults, m.commandHistory[i])
					}
				}
			}

			// Reset index and scroll offset if results changed
			m.historySearchIndex = 0
			m.historySearchScrollOffset = 0
			return m, nil
		}

		// Pass the key to the input field for regular typing
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// If completions are showing, update them based on new input
		if m.showCompletions {
			newInput := m.input.Value()
			m.completions = m.completionEngine.Complete(newInput)

			// If no completions match, hide the modal
			if len(m.completions) == 0 {
				m.showCompletions = false
				m.completionIndex = -1
				m.completionScrollOffset = 0
			} else {
				// Reset selection and scroll to first item when list changes
				m.completionIndex = 0
				m.completionScrollOffset = 0
			}
		}

		return m, cmd
	}
}

// wrapLongLines wraps lines that exceed the specified width
func wrapLongLines(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		// Don't wrap lines that are part of CREATE TABLE definition
		if strings.Contains(line, "CREATE TABLE") || len(line) <= maxWidth {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Check if this is a CQL property line (starts with AND or has = sign)
		if strings.TrimSpace(line) != "" && (strings.Contains(line, " = ") || strings.HasPrefix(strings.TrimSpace(line), "AND ")) {
			// For CQL WITH clause properties, wrap intelligently
			currentLine := line
			indentLevel := len(line) - len(strings.TrimLeft(line, " "))
			indent := strings.Repeat(" ", indentLevel+4) // Extra indent for continuation

			for len(currentLine) > maxWidth {
				// Find a good break point
				breakPoint := -1

				// First try to break after a comma within a reasonable distance
				for i := maxWidth - 1; i > maxWidth-20 && i >= 0; i-- {
					if i < len(currentLine) && currentLine[i] == ',' {
						breakPoint = i + 1
						break
					}
				}

				// If no comma found, try to break at a space
				if breakPoint == -1 {
					for i := maxWidth - 1; i > maxWidth/2 && i >= 0; i-- {
						if i < len(currentLine) && currentLine[i] == ' ' {
							breakPoint = i + 1
							break
						}
					}
				}

				// If still no good break point, just break at maxWidth
				if breakPoint == -1 || breakPoint >= len(currentLine) {
					breakPoint = maxWidth
				}

				// Add the wrapped part
				wrappedLines = append(wrappedLines, currentLine[:breakPoint])

				// Continue with the rest, adding indentation
				if breakPoint < len(currentLine) {
					currentLine = indent + strings.TrimSpace(currentLine[breakPoint:])
				} else {
					currentLine = ""
					break
				}
			}

			// Add any remaining part
			if len(currentLine) > 0 {
				wrappedLines = append(wrappedLines, currentLine)
			}
		} else {
			// For other lines, just add them as-is
			wrappedLines = append(wrappedLines, line)
		}
	}

	return strings.Join(wrappedLines, "\n")
}