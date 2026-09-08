package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// handleSpecialCommands handles special commands like EXIT, QUIT, CLEAR
func (m *MainModel) handleSpecialCommands(command string) (*MainModel, tea.Cmd, bool) {
	upperCommand := strings.ToUpper(command)

	if upperCommand == "EXIT" || upperCommand == "QUIT" {
		// Give the wheel back to the terminal on exit.
		DisableAlternateScroll()
		return m, tea.Quit, true
	}

	// MOUSE is handled here rather than in the router because it is UI state:
	// the mode is a field on the View returned each render, not something the
	// session knows about.
	if upperCommand == "MOUSE" || strings.HasPrefix(upperCommand, "MOUSE ") {
		return m.handleMouseCommand(upperCommand)
	}

	if upperCommand == "CLEAR" || upperCommand == "CLS" {
		m.fullHistoryContent = ""
		m.updateHistoryWrapping()
		m.input.Reset()
		m.lastCommand = ""
		m.rowCount = 0
		m.horizontalOffset = 0
		m.lastTableData = nil
		m.tableWidth = 0
		m.tableHeaders = nil
		m.columnWidths = nil
		m.hasTable = false
		m.cachedTableLines = nil // Clear table cache
		m.viewMode = "history"
		return m, nil, true
	}

	return m, nil, false
}

// handleMouseCommand turns mouse reporting on or off.
//
// The two are exclusive at the terminal level, so this is a choice between
// clicking the mode tabs and the terminal's own text selection. Saying which is
// active matters: someone who has never heard of this command would otherwise
// just find that selecting text stopped working, with no clue why.
func (m *MainModel) handleMouseCommand(upperCommand string) (*MainModel, tea.Cmd, bool) {
	arg := strings.TrimSpace(strings.TrimPrefix(upperCommand, "MOUSE"))

	switch arg {
	case "ON":
		m.mouseEnabled = true
	case "OFF":
		m.mouseEnabled = false
	case "":
		// Report rather than change, the way SHOW does.
		m.appendMouseStatus()
		m.input.Reset()
		return m, nil, true
	default:
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(
			"Usage: MOUSE [ON|OFF]") + "\n"
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.input.Reset()
		return m, nil, true
	}

	m.appendMouseStatus()
	m.input.Reset()
	return m, nil, true
}

// appendMouseStatus writes the current mouse mode and what it costs.
func (m *MainModel) appendMouseStatus() {
	if m.mouseEnabled {
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("Mouse: ON") +
			m.styles.MutedText.Render(" - tabs are clickable; hold Shift to select text") + "\n"
	} else {
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("Mouse: OFF") +
			m.styles.MutedText.Render(" - select and paste as usual; tabs are not clickable") + "\n"
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
}
