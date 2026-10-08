package ui

import (
	tea "charm.land/bubbletea/v2"
)

// The prompt's line-editing keys: Ctrl+K, Ctrl+U, Ctrl+W, Alt+D, Ctrl+Y and
// the word jumps.
//
// The cursor counts characters, and the text is UTF-8, where a character can
// be one byte or four. Cutting the text at the cursor as if it counted bytes
// cut in the wrong place for anything outside ASCII: Ctrl+U on a line of
// Japanese took away a third of what it should, and could leave half a
// character behind. So every one of these works on characters.

// promptChars is the prompt's text as characters, and the cursor's place in
// them.
func (m *MainModel) promptChars() ([]rune, int) {
	chars := []rune(m.input.Value())
	return chars, min(max(m.input.Position(), 0), len(chars))
}

// setPrompt puts text in the prompt with the cursor at a character.
func (m *MainModel) setPrompt(chars []rune, cursor int) {
	m.input.SetValue(string(chars))
	m.input.SetCursor(cursor)
}

// handleCtrlK handles Ctrl+K - cut from cursor to end of line
func (m *MainModel) handleCtrlK() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor < len(chars) {
		m.clipboardBuffer = string(chars[cursor:])
		m.setPrompt(chars[:cursor], cursor)
	}
	return m, nil
}

// handleCtrlU handles Ctrl+U - cut from beginning of line to cursor
func (m *MainModel) handleCtrlU() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor > 0 {
		m.clipboardBuffer = string(chars[:cursor])
		m.setPrompt(chars[cursor:], 0)
	}
	return m, nil
}

// handleCtrlW handles Ctrl+W - delete word backward
func (m *MainModel) handleCtrlW() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor > 0 {
		start := cursor
		for start > 0 && chars[start-1] == ' ' { // the spaces before the cursor
			start--
		}
		for start > 0 && chars[start-1] != ' ' { // and the word before them
			start--
		}
		m.clipboardBuffer = string(chars[start:cursor])
		m.setPrompt(append(append([]rune{}, chars[:start]...), chars[cursor:]...), start)
	}
	return m, nil
}

// handleAltD handles Alt+D - delete word forward, leaving the cursor where it
// is.
func (m *MainModel) handleAltD() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor < len(chars) {
		end := cursor
		for end < len(chars) && chars[end] == ' ' { // the spaces after the cursor
			end++
		}
		for end < len(chars) && chars[end] != ' ' { // and the word after them
			end++
		}
		m.clipboardBuffer = string(chars[cursor:end])
		m.setPrompt(append(append([]rune{}, chars[:cursor]...), chars[end:]...), cursor)
	}
	return m, nil
}

// handleCtrlA handles Ctrl+A - move to beginning of line
func (m *MainModel) handleCtrlA() (*MainModel, tea.Cmd) {
	m.input.CursorStart()
	return m, nil
}

// handleCtrlE handles Ctrl+E - move to end of line
func (m *MainModel) handleCtrlE() (*MainModel, tea.Cmd) {
	m.input.CursorEnd()
	return m, nil
}

// handleCtrlLeft handles Ctrl+Left - jump back by word, or by 20 characters
// when the word is short.
func (m *MainModel) handleCtrlLeft() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor > 0 {
		pos := cursor
		for pos > 0 && chars[pos-1] == ' ' {
			pos--
		}
		for pos > 0 && chars[pos-1] != ' ' {
			pos--
		}
		if cursor-pos < 5 {
			pos = max(cursor-20, 0)
		}
		m.input.SetCursor(pos)
	}
	return m, nil
}

// handleCtrlRight handles Ctrl+Right - jump forward by word, or by 20
// characters when the word is short.
func (m *MainModel) handleCtrlRight() (*MainModel, tea.Cmd) {
	chars, cursor := m.promptChars()
	if cursor < len(chars) {
		pos := cursor
		for pos < len(chars) && chars[pos] != ' ' {
			pos++
		}
		for pos < len(chars) && chars[pos] == ' ' {
			pos++
		}
		if pos-cursor < 5 {
			pos = min(cursor+20, len(chars))
		}
		m.input.SetCursor(pos)
	}
	return m, nil
}

// handleCtrlY handles Ctrl+Y - paste (yank) from clipboard buffer
func (m *MainModel) handleCtrlY() (*MainModel, tea.Cmd) {
	if m.clipboardBuffer != "" {
		chars, cursor := m.promptChars()
		pasted := []rune(m.clipboardBuffer)
		value := append(append(append([]rune{}, chars[:cursor]...), pasted...), chars[cursor:]...)
		m.setPrompt(value, cursor+len(pasted))
	}
	return m, nil
}
