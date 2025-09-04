package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleModalConfirmation handles Enter key when a modal is showing
func (m *MainModel) handleModalConfirmation(_ string) (*MainModel, tea.Cmd) {
	if m.modal.Selected == 1 { // "Execute" button
		// Execute the dangerous command
		command := m.modal.Command
		m.modal = Modal{Type: ModalNone}
		m.input.Placeholder = "Enter CQL command..."
		// Continue with normal command execution

		// Add to history
		m.commandHistory = append(m.commandHistory, command)
		m.historyIndex = -1
		m.lastCommand = command

		// Save to persistent history
		if m.historyManager != nil {
			if err := m.historyManager.SaveCommand(command); err != nil {
				// Log error but don't fail command execution
				fmt.Fprintf(os.Stderr, "Warning: could not save command to history: %v\n", err)
			}
		}

		// Process the command
		start := time.Now()
		result := router.ProcessCommand(command, m.session)
		m.lastQueryTime = time.Since(start)

		// Add command to full history and viewport
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()

		// Handle different result types
		logger.DebugfToFile("HandleEnterKey", "Result type: %T", result)
		updatedModel, cmd := m.processCommandResult(command, result, start)
		return updatedModel, cmd
		
	} else { // "Cancel" button
		// Cancel the command
		m.modal = Modal{Type: ModalNone}
		m.input.Placeholder = "Enter CQL command..."
		m.input.Reset()

		// Add cancellation message to history
		m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()
		return m, nil
	}
}