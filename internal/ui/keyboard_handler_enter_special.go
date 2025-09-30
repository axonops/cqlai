package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSpecialCommands handles special commands like EXIT, QUIT, CLEAR
func (m *MainModel) handleSpecialCommands(command string) (*MainModel, tea.Cmd, bool) {
	upperCommand := strings.ToUpper(command)
	
	if upperCommand == "EXIT" || upperCommand == "QUIT" {
		// Disable mouse tracking on exit
		fmt.Print("\x1b[?1000l") // Disable basic mouse tracking
		fmt.Print("\x1b[?1006l") // Disable SGR mouse mode
		return m, tea.Quit, true
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