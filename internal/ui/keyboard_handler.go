package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/logger"
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
		return m.handleAICQLModal(msg)
	}

	// Check for AI selection modal (second priority)
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		return m.handleAISelectionModal(msg)
	}

	// Handle AI conversation view input
	if m.viewMode == "ai" && m.aiConversationActive {
		// Try to handle in AI conversation handler first
		result, cmd := m.handleAIConversationInput(msg)
		if result != nil {
			// Key was handled by AI conversation handler
			return result, cmd
		}
		// Key wasn't handled, let it fall through to main switch
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		return m.handleCtrlC()

	case tea.KeyCtrlD:
		return m.handleCtrlD()

	case tea.KeyCtrlR:
		return m.handleCtrlR()

	case tea.KeyCtrlK:
		return m.handleCtrlK()

	case tea.KeyCtrlU:
		return m.handleCtrlU()

	case tea.KeyCtrlW:
		return m.handleCtrlW()

	case tea.KeyCtrlP:
		return m.handleCtrlP()

	case tea.KeyCtrlA:
		return m.handleCtrlA()

	case tea.KeyCtrlE:
		return m.handleCtrlE()

	case tea.KeyCtrlLeft:
		return m.handleCtrlLeft()

	case tea.KeyCtrlRight:
		return m.handleCtrlRight()

	case tea.KeyCtrlY:
		return m.handleCtrlY()

	case tea.KeyEsc:
		return m.handleEscapeKey()

	case tea.KeyTab:
		// If modal is showing, navigate choices
		if m.modal.Type != ModalNone {
			m.modal.NextChoice()
			return m, nil
		}
		return m.handleTabKey()

	case tea.KeyF2:
		return m.handleF2()

	case tea.KeyF3:
		return m.handleF3()

	case tea.KeyF4:
		return m.handleF4()

	case tea.KeyF5:
		return m.handleF5()

	case tea.KeyF6:
		return m.handleF6()

	case tea.KeySpace:
		return m.handleSpaceKey(msg)

	case tea.KeyPgUp:
		return m.handlePageUp(msg)

	case tea.KeyPgDown:
		return m.handlePageDown(msg)

	case tea.KeyUp:
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchUp()
		}
		return m.handleUpArrow(msg)

	case tea.KeyDown:
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchDown()
		}
		return m.handleDownArrow(msg)

	case tea.KeyLeft:
		return m.handleLeftArrow(msg)

	case tea.KeyRight:
		return m.handleRightArrow(msg)

	case tea.KeyEnter:
		// If in history search mode, select the current entry
		if m.historySearchMode {
			return m.handleHistorySearchSelect()
		}
		// If history modal is showing, select the current entry
		if m.showHistoryModal {
			return m.handleHistoryModalSelect()
		}
		return m.handleEnterKey()

	default:
		// Handle Alt+N (move to next line in history, same as Down arrow)
		if msg.String() == "alt+n" {
			// If in history search mode, navigate search results
			if m.historySearchMode {
				return m.handleHistorySearchDown()
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

		// Handle navigation mode keys (when in table/trace view with navigation mode active)
		if m.navigationMode && (m.viewMode == "table" || m.viewMode == "trace") {
			switch msg.String() {
			case "j":
				// Single line down
				return m.handleSingleLineDown()
			case "k":
				// Single line up
				return m.handleSingleLineUp()
			case "d":
				// Half page down
				return m.handleHalfPageDown()
			case "u":
				// Half page up
				return m.handleHalfPageUp()
			case "g":
				// Go to top
				return m.handleGoToTop()
			case "G":
				// Go to bottom
				return m.handleGoToBottom()
			case "<":
				// Scroll left by 10 columns
				return m.handlePageLeftScroll()
			case ">":
				// Scroll right by 10 columns
				return m.handlePageRightScroll()
			case "h":
				// Scroll left by one column
				return m.handleHorizontalScrollLeft()
			case "l":
				// Scroll right by one column
				return m.handleHorizontalScrollRight()
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
func (m *MainModel) handleUpArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleUpArrow called.")

	// If completions are showing, navigate up
	if m.showCompletions && len(m.completions) > 0 {
		m.completionIndex--
		if m.completionIndex < 0 {
			m.completionIndex = len(m.completions) - 1
			// Jump to the end of the list
			if len(m.completions) > 10 {
				m.completionScrollOffset = len(m.completions) - 10
			}
		}

		// Adjust scroll offset if selection moves out of view
		if m.completionIndex < m.completionScrollOffset {
			m.completionScrollOffset = m.completionIndex
		}
		return m, nil
	}


	// If history modal is showing, navigate up (go to older command)
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		// Navigate to older commands (decrease index in original array)
		if m.historyModalIndex > 0 {
			m.historyModalIndex--
			// Adjust scroll offset if selection moves out of view
			if m.historyModalIndex < m.historyModalScrollOffset {
				m.historyModalScrollOffset = m.historyModalIndex
			}
		}
		return m, nil
	}

	// If Alt is held, scroll viewport up by one line
	if msg.Alt {
		return m.handleAltScrollUp()
	}

	// Handle command history navigation up
	return m.handleCommandHistoryUp()
}

// handleDownArrow handles Down arrow key press
func (m *MainModel) handleDownArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleDownArrow called.")

	// If completions are showing, navigate down
	if m.showCompletions && len(m.completions) > 0 {
		m.completionIndex = (m.completionIndex + 1) % len(m.completions)

		// Reset scroll to top when wrapping around
		if m.completionIndex == 0 {
			m.completionScrollOffset = 0
		}

		// Adjust scroll offset if selection moves out of view
		if m.completionIndex >= m.completionScrollOffset+10 {
			m.completionScrollOffset = m.completionIndex - 9
		}
		return m, nil
	}


	// If history modal is showing, navigate down (go to newer command)
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		// Navigate to newer commands (increase index in original array)
		if m.historyModalIndex < len(m.commandHistory)-1 {
			m.historyModalIndex++
			// Adjust scroll offset if selection moves out of view
			if m.historyModalIndex >= m.historyModalScrollOffset + 10 {
				m.historyModalScrollOffset = m.historyModalIndex - 9
			}
		}
		return m, nil
	}

	// If Alt is held, scroll viewport down by one line
	if msg.Alt {
		return m.handleAltScrollDown()
	}

	// Handle command history navigation down
	return m.handleCommandHistoryDown()
}


// handleLeftArrow handles Left arrow key press
