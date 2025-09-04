package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEnterKey handles Enter key press
func (m *MainModel) handleEnterKey() (*MainModel, tea.Cmd) {
	// AI info request is now handled in handleKeyboardInput at the beginning

	// Handle AI selection modal if active
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		// Confirm selection (either custom input or selected option)
		selection := m.aiSelectionModal.GetSelection()
		selectionType := m.aiSelectionModal.SelectionType
		m.aiSelectionModal.Active = false
		return m, func() tea.Msg {
			return AISelectionResultMsg{
				Selection:     selection,
				SelectionType: selectionType,
				Cancelled:     false,
			}
		}
	}

	// Cancel exit confirmation if active
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	command := strings.TrimSpace(m.input.Value())

	// Note: We removed the follow-up mode check here since we're now handling
	// follow-up questions within the modal itself

	// Handle AI command
	if strings.HasPrefix(strings.ToUpper(command), ".AI") {
		return m.handleAICommand(command)
	}

	// If completions are showing, accept the selected one
	if m.showCompletions && len(m.completions) > 0 && m.completionIndex >= 0 {
		return m.handleCompletionSelection()
	}

	// Check if modal is showing FIRST before processing command
	if m.modal.Type != ModalNone {
		return m.handleModalConfirmation(command)
	}

	// Process the command from input (unless we're executing an AI command)
	{
		inputText := m.input.Value()
		command = strings.TrimSpace(inputText)
		
		// Check if this is just a comment line
		if strings.HasPrefix(command, "--") || strings.HasPrefix(command, "//") {
			return m.handleCommentLine(command)
		}
		
		// Check for block comment handling
		if strings.HasPrefix(command, "/*") {
			return m.handleBlockComment(command)
		}
		
		// If we're in multi-line mode and this ends a block comment
		if m.multiLineMode && len(m.multiLineBuffer) > 0 {
			updatedModel, cmd := m.handleMultiLineBlockComment(command)
			if updatedModel != nil {
				return updatedModel, cmd
			}
		}
		
		if command == "" && !m.multiLineMode {
			return m, nil
		}
	}

	// Check if this is a CQL statement (not a meta command)
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	isCQLStatement := !strings.HasPrefix(upperCommand, "DESCRIBE") &&
		!strings.HasPrefix(upperCommand, "DESC ") &&
		!strings.HasPrefix(upperCommand, "CONSISTENCY") &&
		!strings.HasPrefix(upperCommand, "OUTPUT") &&
		!strings.HasPrefix(upperCommand, "PAGING") &&
		!strings.HasPrefix(upperCommand, "TRACING") &&
		!strings.HasPrefix(upperCommand, "SOURCE") &&
		!strings.HasPrefix(upperCommand, "CAPTURE") &&
		!strings.HasPrefix(upperCommand, "EXPAND") &&
		!strings.HasPrefix(upperCommand, "SHOW") &&
		!strings.HasPrefix(upperCommand, "HELP") &&
		!strings.HasPrefix(upperCommand, "CLEAR") &&
		!strings.HasPrefix(upperCommand, "CLS") &&
		!strings.HasPrefix(upperCommand, "EXIT") &&
		!strings.HasPrefix(upperCommand, "QUIT")

	// For CQL statements, check for semicolon (skip for AI-generated commands)
	if isCQLStatement {
		if !strings.HasSuffix(strings.TrimSpace(command), ";") {
			// Enter multi-line mode
			if !m.multiLineMode {
				m.multiLineMode = true
				m.multiLineBuffer = []string{command}
				m.input.Placeholder = "... (multi-line mode, end with ;)"
			} else {
				// Add to buffer
				m.multiLineBuffer = append(m.multiLineBuffer, command)
			}
			m.input.Reset()

			// Update the prompt to show we're in multi-line mode
			m.input.SetValue("")
			return m, nil
		} else if m.multiLineMode {
			// We have a semicolon and we're in multi-line mode
			m.multiLineBuffer = append(m.multiLineBuffer, command)
			command = strings.Join(m.multiLineBuffer, " ")
			m.multiLineMode = false
			m.multiLineBuffer = nil
			m.input.Placeholder = "Enter CQL command..."
		}
	}

	// Check for dangerous commands (skip for AI commands - already checked)
	if m.sessionManager != nil && m.sessionManager.RequireConfirmation() && router.IsDangerousCommand(command) {
		// Show confirmation modal for dangerous commands
		m.modal = NewConfirmationModal(command)

		// Add command to history
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
		m.updateHistoryWrapping()
		m.historyViewport.GotoBottom()

		m.input.Reset()
		return m, nil
	}

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

	// Check for special commands
	model, cmd, handled := m.handleSpecialCommands(command)
	if handled {
		return model, cmd
	}

	start := time.Now()
	result := router.ProcessCommand(command, m.session)
	m.lastQueryTime = time.Since(start)

	// Add command to history viewport
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
	
	// Capture trace data if tracing is enabled and this was a query that returns results
	m.captureTraceData(command)

	// Handle different result types
	logger.DebugfToFile("HandleEnterKey", "Result type (2nd location): %T", result)
	return m.processCommandResult(command, result, start)
}