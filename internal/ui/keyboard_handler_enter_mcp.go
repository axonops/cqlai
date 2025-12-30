package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMCPCommand handles .mcp commands
func (m *MainModel) handleMCPCommand(command string) (*MainModel, tea.Cmd) {
	// Log the MCP command
	logger.DebugfToFile("MCP", "User MCP command: %s", command)

	// Add command to history
	m.commandHistory = append(m.commandHistory, command)
	m.historyIndex = -1
	m.lastCommand = command

	// Save to persistent history
	if m.historyManager != nil {
		if err := m.historyManager.SaveCommand(command); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save MCP command to history: %v\n", err)
		}
	}

	// Add command to history display
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	// Get MCP handler
	mcpHandler := router.GetMCPHandler()
	if mcpHandler == nil {
		result := "MCP handler not initialized"
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(result)
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		m.input.Reset()
		return m, nil
	}

	// Execute MCP command
	result := mcpHandler.HandleMCPCommand(command)

	// Display result
	if result != "" {
		// Format multiline results properly
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				m.fullHistoryContent += "\n" + line
			}
		}
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
	}

	// Reset input
	m.input.Reset()

	return m, nil
}
