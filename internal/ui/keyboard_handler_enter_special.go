package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// handleSpecialCommands handles special commands like EXIT, QUIT, CLEAR
func (m *MainModel) handleSpecialCommands(command string) (*MainModel, tea.Cmd, bool) {
	upperCommand := strings.ToUpper(command)

	if upperCommand == "EXIT" || upperCommand == "QUIT" {
		// Give the wheel and the buttons back to the terminal on exit.
		DisableAlternateScroll()
		ResetMouseReporting()
		return m, tea.Quit, true
	}

	// MOUSE is handled here rather than in the router because it is UI state:
	// the mode is a field on the View returned each render, not something the
	// session knows about.
	if upperCommand == "MOUSE" || strings.HasPrefix(upperCommand, "MOUSE ") {
		return m.handleMouseCommand(upperCommand)
	}

	// CAPTURE on its own opens the window rather than reporting the status.
	// The bottom line already says whether capture is running, so the status
	// was telling you something you could see; the window is what you wanted.
	// CAPTURE with a file, or OFF, still goes to the command.
	if upperCommand == "CAPTURE" {
		m.input.Reset()
		// Centred, not above the field: you typed it at the prompt, so there
		// is nothing on the bottom line for it to be pointing at.
		updated, cmd := m.openCapturePanelCentred()
		return updated, cmd, true
	}

	if upperCommand == "CLEAR" || upperCommand == "CLS" {
		m.fullHistoryContent = ""
		m.updateHistoryWrapping()
		m.input.Reset()
		m.lastCommand = ""
		m.rowCount = 0
		m.resetHorizontalScroll()
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
// On, which is the default, cqlai has the mouse: the tabs and the settings on
// the bottom line are clickable and dragging selects text, because cqlai draws
// the selection itself. Off hands the mouse back to the terminal, for anyone
// who would rather have its selection, its right-click paste and its own idea
// of what a double click means.
//
// Nothing is written to the terminal here. MouseMode is a field on the View, so
// the next render carries the change; writing escape sequences from under the
// renderer fights it for the screen.
func (m *MainModel) handleMouseCommand(upperCommand string) (*MainModel, tea.Cmd, bool) {
	arg := strings.TrimSpace(strings.TrimPrefix(upperCommand, "MOUSE"))

	switch arg {
	case "ON":
		m.mouseEnabled = true
	case "OFF":
		m.mouseEnabled = false
		m.clearSelection()
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

// appendMouseStatus writes which mode the mouse is in and what each one does.
func (m *MainModel) appendMouseStatus() {
	if m.mouseEnabled {
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("Mouse: ON") +
			m.styles.MutedText.Render(" - click the tabs and the settings on the bottom line; drag to select, and the text goes to the clipboard") + "\n"
	} else {
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("Mouse: OFF") +
			m.styles.MutedText.Render(" - the terminal handles selection and paste; the tabs and settings are not clickable") + "\n"
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
}
