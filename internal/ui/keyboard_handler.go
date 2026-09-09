package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// handleKeyboardInput handles keyboard input events
func (m *MainModel) handleKeyboardInput(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	// Any keypress drops the mouse selection: whatever happens next is likely
	// to move or replace the text under it, and a highlight left behind on
	// different text is worse than no highlight. Escape does nothing else, so
	// it is the way to take one down deliberately.
	if m.selection.active {
		m.clearSelection()
		if msg.String() == "esc" {
			return m, nil
		}
	}

	// The settings chooser takes keys while it is open: it was opened by a
	// click a moment ago, so it is what the next keypress is aimed at.
	if m.chooser.active {
		return m.handleSettingChooserKey(msg)
	}

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

	switch msg.String() {
	case "ctrl+c":
		return m.handleCtrlC()

	case "ctrl+d":
		return m.handleCtrlD()

	case "ctrl+r":
		return m.handleCtrlR()

	case "ctrl+k":
		return m.handleCtrlK()

	case "ctrl+u":
		return m.handleCtrlU()

	case "ctrl+w":
		return m.handleCtrlW()

	case "ctrl+p":
		return m.handleCtrlP()

	case "ctrl+a":
		return m.handleCtrlA()

	case "ctrl+e":
		return m.handleCtrlE()

	case "ctrl+left":
		return m.handleCtrlLeft()

	case "ctrl+right":
		return m.handleCtrlRight()

	case "ctrl+y":
		return m.handleCtrlY()

	case "esc":
		return m.handleEscapeKey()

	case "tab":
		// If modal is showing, navigate choices
		if m.modal.Type != ModalNone {
			m.modal.NextChoice()
			return m, nil
		}
		return m.handleTabKey()

	case "f2":
		return m.handleF2()

	case "f3":
		return m.handleF3()

	case "f4":
		return m.handleF4()

	case "f5":
		return m.handleF5()

	case "f6":
		return m.handleF6()

	case "space":
		// A space is part of what you are searching for. Without this it went
		// to the pager instead and never reached the query, so searching for
		// anything with a space in it silently matched nothing.
		if m.historySearchMode {
			return m.typeIntoHistorySearch(msg)
		}
		return m.handleSpaceKey(msg)

	case "pgup":
		if m.historySearchMode {
			return m.moveHistorySelection(-historyModalRows)
		}
		return m.handlePageUp(msg)

	case "pgdown":
		if m.historySearchMode {
			return m.moveHistorySelection(historyModalRows)
		}
		return m.handlePageDown(msg)

	// The modified forms route to the same handlers, which then read msg.Mod
	// themselves - Alt+arrow scrolls where a bare arrow moves the cursor or
	// walks history. v1 switched on msg.Type, which ignored modifiers; v2's
	// String() spells them out, so "alt+left" no longer matches "left" and the
	// Alt bindings would silently stop working.
	case "up", "alt+up", "shift+up", "ctrl+up":
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchUp()
		}
		return m.handleUpArrow(msg)

	case "down", "alt+down", "shift+down", "ctrl+down":
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchDown()
		}
		return m.handleDownArrow(msg)

	case "left", "alt+left", "shift+left":
		return m.handleLeftArrow(msg)

	case "right", "alt+right", "shift+right":
		return m.handleRightArrow(msg)

	case "enter":
		// If in history search mode, select the current entry
		if m.historySearchMode {
			return m.handleHistorySearchSelect()
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
			return m.typeIntoHistorySearch(msg)
		}

		// Pass the key to the input field for regular typing
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// If completions are showing, update them based on new input
		if m.showCompletions {
			newInput := m.input.Value()
			// If in multi-line mode, combine buffer with current input
			fullInput := newInput
			if m.multiLineMode && len(m.multiLineBuffer) > 0 {
				fullInput = strings.Join(m.multiLineBuffer, " ") + " " + newInput
			}
			m.completions = m.completionEngine.Complete(fullInput)

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
func (m *MainModel) handleUpArrow(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
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

	// If Alt is held, scroll viewport up by one line
	if msg.Mod.Contains(tea.ModAlt) {
		return m.handleAltScrollUp()
	}

	// While a result set is on screen, scroll it. Alternate scroll mode delivers
	// wheel events as plain Up presses, so this is also what the wheel does here.
	// Command history stays on Up in the normal view.
	if m.viewportOwnsArrows() {
		return m.handleAltScrollUp()
	}

	// Handle command history navigation up
	return m.handleCommandHistoryUp()
}

// handleDownArrow handles Down arrow key press
func (m *MainModel) handleDownArrow(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
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

	// If Alt is held, scroll viewport down by one line
	if msg.Mod.Contains(tea.ModAlt) {
		return m.handleAltScrollDown()
	}

	// See handleUpArrow: the wheel arrives here too while a result set is up.
	if m.viewportOwnsArrows() {
		return m.handleAltScrollDown()
	}

	// Handle command history navigation down
	return m.handleCommandHistoryDown()
}

// handleLeftArrow handles Left arrow key press
