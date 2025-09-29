package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyboardInput handles keyboard input events
func (m *MainModel) handleKeyboardInput(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	// Check for save modal first (highest priority)
	if m.saveModalActive {
		return m.handleSaveModalKeyboard(msg)
	}

	// Check for AI CQL modal (high priority)
	if m.aiCQLModal != nil && m.aiCQLModal.Active {
		switch msg.String() {
		case "left", "h":
			m.aiCQLModal.PrevChoice()
			return m, nil
		case "right", "l":
			m.aiCQLModal.NextChoice()
			return m, nil
		case "enter":
			switch m.aiCQLModal.Selected {
			case 0:
				// Execute the CQL
				cql := m.aiCQLModal.CQL
				m.aiCQLModal.Active = false
				m.aiCQLModal = nil

				// Exit AI conversation view
				m.aiConversationActive = false
				m.viewMode = "history"

				// Execute the CQL command
				m.input.SetValue(cql)
				// Trigger the enter key handler to execute
				return m.handleEnterKey()
			case 1:
				// Edit - put the CQL in the input field and close modal
				cql := m.aiCQLModal.CQL
				m.aiCQLModal.Active = false
				m.aiCQLModal = nil

				// Exit AI conversation view and go to query view
				m.aiConversationActive = false
				m.viewMode = "history"

				// Put the CQL in the input field for editing
				m.input.SetValue(cql)
				m.input.CursorEnd()
				return m, nil
			case 2:
				// Cancel - close modal and stay in AI conversation
				m.aiCQLModal.Active = false
				m.aiCQLModal = nil
				return m, nil
			}
		case "esc", "ctrl+c":
			// Cancel - close modal and stay in AI conversation
			m.aiCQLModal.Active = false
			m.aiCQLModal = nil
			return m, nil
		}
		return m, nil
	}

	// Check for AI selection modal (second priority)
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		// Handle selection modal keyboard input
		if m.aiSelectionModal.InputMode {
			// In custom input mode
			switch msg.Type {
			case tea.KeyEscape:
				// Exit input mode
				m.aiSelectionModal.InputMode = false
				m.aiSelectionModal.CustomInput = ""
				return m, nil
			case tea.KeyEnter:
				// Submit custom input
				if m.aiSelectionModal.CustomInput != "" {
					return m, func() tea.Msg {
						return AISelectionResultMsg{
							Selection:     m.aiSelectionModal.CustomInput,
							SelectionType: m.aiSelectionModal.SelectionType,
							Cancelled:     false,
						}
					}
				}
				return m, nil
			case tea.KeyBackspace:
				if len(m.aiSelectionModal.CustomInput) > 0 {
					m.aiSelectionModal.CustomInput = m.aiSelectionModal.CustomInput[:len(m.aiSelectionModal.CustomInput)-1]
				}
				return m, nil
			default:
				// Add character to custom input
				if msg.Type == tea.KeyRunes {
					m.aiSelectionModal.CustomInput += string(msg.Runes)
				}
				return m, nil
			}
		} else {
			// In selection mode
			switch msg.String() {
			case "up", "k":
				m.aiSelectionModal.PrevOption()
				return m, nil
			case "down", "j":
				m.aiSelectionModal.NextOption()
				return m, nil
			case "i":
				// Enter custom input mode
				m.aiSelectionModal.InputMode = true
				m.aiSelectionModal.CustomInput = ""
				return m, nil
			case "enter":
				// Select current option
				return m, func() tea.Msg {
					return AISelectionResultMsg{
						Selection:     m.aiSelectionModal.GetSelection(),
						SelectionType: m.aiSelectionModal.SelectionType,
						Cancelled:     false,
					}
				}
			case "esc", "ctrl+c":
				// Cancel selection
				return m, func() tea.Msg {
					return AISelectionResultMsg{
						Cancelled: true,
					}
				}
			}
		}
		return m, nil
	}

	// Handle AI conversation view input
	if m.viewMode == "ai" && m.aiConversationActive {
		// Handle history search mode in AI view
		if m.historySearchMode {
			switch msg.Type {
			case tea.KeyCtrlR, tea.KeyCtrlC, tea.KeyEscape:
				// Exit history search mode
				m.historySearchMode = false
				m.historySearchQuery = ""
				m.historySearchResults = []string{}
				m.historySearchIndex = 0
				m.historySearchScrollOffset = 0
				return m, nil
			case tea.KeyEnter:
				// Select the current history entry
				if len(m.historySearchResults) > 0 && m.historySearchIndex < len(m.historySearchResults) {
					// Set the AI input value to the selected history entry
					m.aiConversationInput.SetValue(m.historySearchResults[m.historySearchIndex])
					// Exit history search mode
					m.historySearchMode = false
					m.historySearchQuery = ""
					m.historySearchResults = []string{}
					m.historySearchIndex = 0
					m.historySearchScrollOffset = 0
				}
				return m, nil
			case tea.KeyUp:
				// Navigate search results
				if m.historySearchIndex > 0 {
					m.historySearchIndex--
					// Adjust scroll offset if needed
					if m.historySearchIndex < m.historySearchScrollOffset {
						m.historySearchScrollOffset = m.historySearchIndex
					}
				}
				return m, nil
			case tea.KeyDown:
				// Navigate search results
				if m.historySearchIndex < len(m.historySearchResults)-1 {
					m.historySearchIndex++
					// Adjust scroll offset if needed
					if m.historySearchIndex >= m.historySearchScrollOffset+10 {
						m.historySearchScrollOffset = m.historySearchIndex - 9
					}
				}
				return m, nil
			default:
				// Handle typing for search query
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

				// Update search results from AI history
				if m.aiHistoryManager != nil {
					results := m.aiHistoryManager.SearchHistory(m.historySearchQuery)
					m.historySearchResults = results
					if len(results) > 0 {
						// Start at the bottom (newest matching command)
						m.historySearchIndex = len(results) - 1
						// Set scroll offset to show the bottom
						if len(results) > 10 {
							m.historySearchScrollOffset = len(results) - 10
						} else {
							m.historySearchScrollOffset = 0
						}
					}
				}
				return m, nil
			}
		}

		// Check for Ctrl+R to toggle history search mode
		if msg.Type == tea.KeyCtrlR {
			if m.historySearchMode {
				// Exit history search mode
				m.historySearchMode = false
				m.historySearchQuery = ""
				m.historySearchResults = []string{}
				m.historySearchIndex = 0
				m.historySearchScrollOffset = 0
			} else {
				// Enter history search mode
				m.historySearchMode = true
				m.historySearchQuery = ""
				// Initialize with all AI history items (empty search shows all)
				if m.aiHistoryManager != nil {
					m.historySearchResults = m.aiHistoryManager.SearchHistory("")
				} else {
					m.historySearchResults = m.commandHistory
				}
				// Start at the bottom (newest command)
				if len(m.historySearchResults) > 0 {
					m.historySearchIndex = len(m.historySearchResults) - 1
					// Set scroll offset to show the bottom
					if len(m.historySearchResults) > 10 {
						m.historySearchScrollOffset = len(m.historySearchResults) - 10
					} else {
						m.historySearchScrollOffset = 0
					}
				} else {
					m.historySearchIndex = 0
					m.historySearchScrollOffset = 0
				}
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			// Exit AI mode and return to history
			m.aiConversationActive = false
			m.viewMode = "history"
			m.aiConversationInput.SetValue("")
			m.aiProcessing = false
			return m, nil
		case tea.KeyEnter:
			// Submit message to AI
			message := strings.TrimSpace(m.aiConversationInput.Value())
			if message != "" && !m.aiProcessing {
				// Add to AI command history (separate from CQL history)
				m.aiCommandHistory = append(m.aiCommandHistory, message)
				m.aiHistoryIndex = -1

				// Save to AI history manager for Ctrl+R search in AI view
				if m.aiHistoryManager != nil {
					if err := m.aiHistoryManager.SaveCommand(message); err != nil {
						// Log error but don't fail
						logger.DebugfToFile("AI", "Could not save command to AI history: %v", err)
					}
				}

				// Add user message to raw messages
				m.aiConversationMessages = append(m.aiConversationMessages, AIMessage{
					Role:    "user",
					Content: message,
				})
				// Rebuild the conversation with proper wrapping
				m.rebuildAIConversation()

				// Clear input and set processing state
				userMessage := message
				m.aiConversationInput.SetValue("")
				m.aiProcessing = true

				// Send message to AI
				return m, func() tea.Msg {
					// If this is a follow-up in an existing conversation
					if m.aiConversationID != "" {
						return continueAIConversation(m.aiConfig, m.aiConversationID, userMessage)()
					}
					// Start new conversation
					return generateAICQL(m.session, m.aiConfig, userMessage)()
				}
			}
			return m, nil
		case tea.KeyUp:
			// Check for Alt modifier first for scrolling
			if msg.Alt {
				// Scroll conversation up by one line
				m.aiConversationViewport.YOffset = max(0, m.aiConversationViewport.YOffset-1)
				return m, nil
			}
			// Navigate AI command history (not CQL history)
			if len(m.aiCommandHistory) > 0 {
				if m.aiHistoryIndex == -1 {
					// Save current input before navigating history
					m.currentInput = m.aiConversationInput.Value()
					m.aiHistoryIndex = len(m.aiCommandHistory) - 1
				} else if m.aiHistoryIndex > 0 {
					m.aiHistoryIndex--
				}
				m.aiConversationInput.SetValue(m.aiCommandHistory[m.aiHistoryIndex])
			}
			return m, nil
		case tea.KeyDown:
			// Check for Alt modifier first for scrolling
			if msg.Alt {
				// Scroll conversation down by one line
				maxOffset := max(0, m.aiConversationViewport.TotalLineCount()-m.aiConversationViewport.Height)
				m.aiConversationViewport.YOffset = min(maxOffset, m.aiConversationViewport.YOffset+1)
				return m, nil
			}
			// Navigate AI command history (not CQL history)
			if m.aiHistoryIndex != -1 {
				if m.aiHistoryIndex < len(m.aiCommandHistory)-1 {
					m.aiHistoryIndex++
					m.aiConversationInput.SetValue(m.aiCommandHistory[m.aiHistoryIndex])
				} else {
					// Return to current input
					m.aiHistoryIndex = -1
					m.aiConversationInput.SetValue(m.currentInput)
				}
			}
			return m, nil
		case tea.KeyPgUp:
			// Scroll conversation up by multiple lines
			m.aiConversationViewport.ScrollUp(3)
			return m, nil
		case tea.KeyPgDown:
			// Scroll conversation down by multiple lines
			m.aiConversationViewport.ScrollDown(3)
			return m, nil
		case tea.KeyF2, tea.KeyF3, tea.KeyF4, tea.KeyF5:
			// Let function keys pass through to be handled by the main key handler
			// Fall through to main handler
		default:
			// Update the text input
			var cmd tea.Cmd
			m.aiConversationInput, cmd = m.aiConversationInput.Update(msg)
			return m, cmd
		}
	}

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
			m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
			m.updateHistoryWrapping()
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

			// Search with empty query shows all history
			if m.historyManager != nil {
				m.historySearchResults = m.historyManager.SearchHistory("")
			} else {
				// Fallback to in-memory history
				m.historySearchResults = make([]string, len(m.commandHistory))
				copy(m.historySearchResults, m.commandHistory)
			}

			// Start at the bottom (newest command)
			if len(m.historySearchResults) > 0 {
				m.historySearchIndex = len(m.historySearchResults) - 1
				// Set scroll offset to show the bottom
				if len(m.historySearchResults) > 10 {
					m.historySearchScrollOffset = len(m.historySearchResults) - 10
				} else {
					m.historySearchScrollOffset = 0
				}
			} else {
				m.historySearchIndex = 0
				m.historySearchScrollOffset = 0
			}
		} else {
			// Exit history search mode
			m.historySearchMode = false
			m.historySearchQuery = ""
			m.historySearchResults = []string{}
			m.historySearchIndex = 0
		}
		return m, nil

	case tea.KeyCtrlK:
		// Cut from cursor to end of line (kill line)
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		if cursorPos < len(currentValue) {
			// Store the cut text in clipboard buffer
			m.clipboardBuffer = currentValue[cursorPos:]
			// Remove the text from cursor to end
			m.input.SetValue(currentValue[:cursorPos])
		}
		return m, nil

	case tea.KeyCtrlU:
		// Cut from beginning of line to cursor (unix-line-discard)
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		if cursorPos > 0 {
			// Store the cut text in clipboard buffer
			m.clipboardBuffer = currentValue[:cursorPos]
			// Remove the text from beginning to cursor
			m.input.SetValue(currentValue[cursorPos:])
			m.input.SetCursor(0)
		}
		return m, nil

	case tea.KeyCtrlW:
		// Cut word backward (delete word before cursor)
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		if cursorPos > 0 {
			// Find the start of the word to cut
			start := cursorPos - 1
			
			// Skip trailing spaces
			for start >= 0 && currentValue[start] == ' ' {
				start--
			}
			
			// Find the beginning of the word
			for start >= 0 && currentValue[start] != ' ' {
				start--
			}
			start++ // Move to the first character of the word
			
			// Store the cut text in clipboard buffer
			m.clipboardBuffer = currentValue[start:cursorPos]
			
			// Remove the word from the input
			newValue := currentValue[:start] + currentValue[cursorPos:]
			m.input.SetValue(newValue)
			m.input.SetCursor(start)
		}
		return m, nil

	case tea.KeyCtrlP:
		// Move to previous line in history (same as Up arrow)
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

	case tea.KeyCtrlA:
		// Move cursor to beginning of line
		m.input.CursorStart()
		return m, nil

	case tea.KeyCtrlE:
		// Move cursor to end of line
		m.input.CursorEnd()
		return m, nil

	case tea.KeyCtrlLeft:
		// Jump backward by word or by 20 characters if no word boundary
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		if cursorPos > 0 {
			// Try to find previous word boundary
			newPos := cursorPos - 1
			// Skip spaces
			for newPos > 0 && currentValue[newPos] == ' ' {
				newPos--
			}
			// Skip word characters
			for newPos > 0 && currentValue[newPos-1] != ' ' {
				newPos--
			}
			// If we didn't move much, jump by 20 characters
			if cursorPos - newPos < 5 {
				newPos = cursorPos - 20
				if newPos < 0 {
					newPos = 0
				}
			}
			m.input.SetCursor(newPos)
		}
		return m, nil

	case tea.KeyCtrlRight:
		// Jump forward by word or by 20 characters if no word boundary
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		valueLen := len(currentValue)
		if cursorPos < valueLen {
			// Try to find next word boundary
			newPos := cursorPos + 1
			// Skip current word
			for newPos < valueLen && currentValue[newPos-1] != ' ' {
				newPos++
			}
			// Skip spaces
			for newPos < valueLen && currentValue[newPos] == ' ' {
				newPos++
			}
			// If we didn't move much, jump by 20 characters
			if newPos - cursorPos < 5 {
				newPos = cursorPos + 20
				if newPos > valueLen {
					newPos = valueLen
				}
			}
			m.input.SetCursor(newPos)
		}
		return m, nil

	case tea.KeyCtrlY:
		// Paste (yank) from clipboard buffer
		if m.clipboardBuffer != "" {
			currentValue := m.input.Value()
			cursorPos := m.input.Position()
			// Insert clipboard content at cursor position
			newValue := currentValue[:cursorPos] + m.clipboardBuffer + currentValue[cursorPos:]
			m.input.SetValue(newValue)
			// Move cursor to end of pasted text
			m.input.SetCursor(cursorPos + len(m.clipboardBuffer))
		}
		return m, nil

	case tea.KeyEsc:
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
		// If we're in the middle of paging through results, clear the paging state
		if m.viewMode == "table" && m.slidingWindow != nil && m.slidingWindow.hasMoreData {
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

	case tea.KeyTab:
		// If modal is showing, navigate choices
		if m.modal.Type != ModalNone {
			m.modal.NextChoice()
			return m, nil
		}
		return m.handleTabKey()

	case tea.KeyF2:
		// F2 to switch to query/history view
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

	case tea.KeyF3:
		// F3 to switch to table view
		if m.viewMode != "table" {
			m.viewMode = "table"
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

	case tea.KeyF4:
		// F4 to switch to trace view
		if m.viewMode != "trace" {
			m.viewMode = "trace"
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

	case tea.KeyF5:
		// F5 to switch to AI view
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

	case tea.KeyF6:
		// F6 to toggle showing data types in table headers
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
		logger.DebugfToFile("AI", "KeyUp pressed. aiSelectionModal.Active=%v, historySearchMode=%v",
			m.aiSelectionModal != nil && m.aiSelectionModal.Active, m.historySearchMode)
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
			// Use the index directly - the modal display handles the reversal
			if m.historyModalIndex >= 0 && m.historyModalIndex < len(m.commandHistory) {
				m.input.SetValue(m.commandHistory[m.historyModalIndex])
			}
			// Close the modal
			m.showHistoryModal = false
			m.historyModalIndex = 0
			m.historyModalScrollOffset = 0
			return m, nil
		}
		return m.handleEnterKey()

	default:
		// Handle Alt+N (move to next line in history, same as Down arrow)
		if msg.String() == "alt+n" {
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
		}

		// Handle Alt+D (delete word forward)
		if msg.String() == "alt+d" {
			currentValue := m.input.Value()
			cursorPos := m.input.Position()
			if cursorPos < len(currentValue) {
				// Find the end of the word to cut
				end := cursorPos
				
				// Skip leading spaces
				for end < len(currentValue) && currentValue[end] == ' ' {
					end++
				}
				
				// Find the end of the word
				for end < len(currentValue) && currentValue[end] != ' ' {
					end++
				}
				
				// Store the cut text in clipboard buffer
				m.clipboardBuffer = currentValue[cursorPos:end]
				
				// Remove the word from the input
				newValue := currentValue[:cursorPos] + currentValue[end:]
				m.input.SetValue(newValue)
				// Cursor stays at the same position
			}
			return m, nil
		}

		// AI info request view input is handled at the beginning of handleKeyboardInput
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

			// Start at the bottom (newest matching command) if results changed
			if len(m.historySearchResults) > 0 {
				m.historySearchIndex = len(m.historySearchResults) - 1
				// Set scroll offset to show the bottom
				if len(m.historySearchResults) > 10 {
					m.historySearchScrollOffset = len(m.historySearchResults) - 10
				} else {
					m.historySearchScrollOffset = 0
				}
			} else {
				m.historySearchIndex = 0
				m.historySearchScrollOffset = 0
			}
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
