package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// handleAIConversationInput handles keyboard input for the AI conversation view
func (m *MainModel) handleAIConversationInput(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
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

	// Check for specific keys that should bypass AI input
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
		// Don't handle function keys here - let them fall through to main handler
		// by not returning anything in this case
	default:
		// Pass other keys to the AI input field
		var cmd tea.Cmd
		m.aiConversationInput, cmd = m.aiConversationInput.Update(msg)
		return m, cmd
	}

	// If we get here, the key wasn't handled, so return nil to let it fall through
	return nil, nil
}

