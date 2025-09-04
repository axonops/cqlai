package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSpecialCommands handles special commands like EXIT, QUIT, CLEAR
func (m *MainModel) handleSpecialCommands(command string) (*MainModel, tea.Cmd, bool) {
	upperCommand := strings.ToUpper(command)
	
	if upperCommand == "EXIT" || upperCommand == "QUIT" {
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
		m.viewMode = "history"
		return m, nil, true
	}

	return m, nil, false
}