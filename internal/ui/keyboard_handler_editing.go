package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleCtrlK handles Ctrl+K - cut from cursor to end of line
func (m *MainModel) handleCtrlK() (*MainModel, tea.Cmd) {
	currentValue := m.input.Value()
	cursorPos := m.input.Position()
	if cursorPos < len(currentValue) {
		// Store the cut text in clipboard buffer
		m.clipboardBuffer = currentValue[cursorPos:]
		// Remove the text from cursor to end
		m.input.SetValue(currentValue[:cursorPos])
	}
	return m, nil
}

// handleCtrlU handles Ctrl+U - cut from beginning of line to cursor
func (m *MainModel) handleCtrlU() (*MainModel, tea.Cmd) {
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
}

// handleCtrlW handles Ctrl+W - delete word backward
func (m *MainModel) handleCtrlW() (*MainModel, tea.Cmd) {
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
}

// handleCtrlA handles Ctrl+A - move cursor to beginning of line
func (m *MainModel) handleCtrlA() (*MainModel, tea.Cmd) {
	m.input.CursorStart()
	return m, nil
}

// handleCtrlE handles Ctrl+E - move cursor to end of line
func (m *MainModel) handleCtrlE() (*MainModel, tea.Cmd) {
	m.input.CursorEnd()
	return m, nil
}

// handleCtrlLeft handles Ctrl+Left - jump backward by word
func (m *MainModel) handleCtrlLeft() (*MainModel, tea.Cmd) {
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
}

// handleCtrlRight handles Ctrl+Right - jump forward by word
func (m *MainModel) handleCtrlRight() (*MainModel, tea.Cmd) {
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
}

// handleCtrlY handles Ctrl+Y - paste (yank) from clipboard buffer
func (m *MainModel) handleCtrlY() (*MainModel, tea.Cmd) {
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
}