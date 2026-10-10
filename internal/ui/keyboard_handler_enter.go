package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// handleEnterKey handles Enter key press
func (m *MainModel) handleEnterKey() (*MainModel, tea.Cmd) {
	// AI info request is now handled in handleKeyboardInput at the beginning

	// If we have more data to load and input is empty, load next page (like PgDn)
	if m.slidingWindow != nil && m.slidingWindow.hasMoreData && m.input.Value() == "" {
		// Treat Enter as PageDown when there's more data available
		return m.handlePageDown(tea.KeyPressMsg{Code: tea.KeyPgDown})
	}

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

	// The prompt is emptied after each line of a statement, so a line the
	// same as the one before it was typed again. It used to be taken for a
	// second press of Enter and dropped, writing different data from what
	// was typed.
	command := strings.TrimSpace(m.input.Value())

	// Note: We removed the follow-up mode check here since we're now handling
	// follow-up questions within the modal itself

	// Handle AI command
	if strings.HasPrefix(strings.ToUpper(command), ".AI") {
		return m.handleAICommand(command)
	}

	// If completions are showing, accept the selected one - unless it is a
	// hint, which is a note about what to type and has nothing to accept. The
	// statement runs instead, so Enter does not have to be pressed twice.
	if m.showCompletions && len(m.completions) > 0 && m.completionIndex >= 0 {
		if !completion.IsHint(m.completions[m.completionIndex]) {
			return m.handleCompletionSelection()
		}
		m.clearCompletions()
	}

	// Process the command from input (unless we're executing an AI command)
	{
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
	// Empty commands in multi-line mode should be treated as CQL to continue multi-line input
	isCQLStatement := (command == "" && m.multiLineMode) ||
		(!strings.HasPrefix(upperCommand, "DESCRIBE") &&
			!strings.HasPrefix(upperCommand, "DESC ") &&
			!strings.HasPrefix(upperCommand, "CONSISTENCY") &&
			!strings.HasPrefix(upperCommand, "OUTPUT") &&
			!strings.HasPrefix(upperCommand, "PAGING") &&
			!strings.HasPrefix(upperCommand, "AUTOFETCH") &&
			!strings.HasPrefix(upperCommand, "TRACING") &&
			!strings.HasPrefix(upperCommand, "SOURCE") &&
			!strings.HasPrefix(upperCommand, "AUTOSAVE") &&
			!strings.HasPrefix(upperCommand, "CAPTURE") &&
			!strings.HasPrefix(upperCommand, "EXPAND") &&
			!strings.HasPrefix(upperCommand, "SHOW") &&
			!strings.HasPrefix(upperCommand, "HELP") &&
			!strings.HasPrefix(upperCommand, "SAVE") &&
			!strings.HasPrefix(upperCommand, "CLEAR") &&
			!strings.HasPrefix(upperCommand, "CLS") &&
			!strings.HasPrefix(upperCommand, "EXIT") &&
			!strings.HasPrefix(upperCommand, "QUIT"))

	// echoed says the statement is already in the console: one typed over
	// several lines is written there a line at a time as it is typed, and
	// writing the whole thing again when it runs would be two of it.
	echoed := false

	// For CQL statements, check the statement is whole. Not at the first line
	// that ends with a semicolon: a BATCH goes on to its APPLY BATCH, and a
	// semicolon in a string or a function's $$ body ends nothing.
	if isCQLStatement {
		pending := command
		if m.multiLineMode {
			pending = strings.Join(append(append([]string{}, m.multiLineBuffer...), command), "\n")
		}
		switch {
		case !router.StatementComplete(pending):
			// Enter multi-line mode
			if !m.multiLineMode {
				m.multiLineMode = true
				m.multiLineBuffer = []string{command}
				m.continueStatement()
			} else {
				// Add to buffer (including empty lines for proper formatting)
				m.multiLineBuffer = append(m.multiLineBuffer, command)
			}
			m.echoInput(command)

			// Create a new empty textinput to ensure it's properly reset
			newInput := textinput.New()
			newInput.Placeholder = m.input.Placeholder
			newInput.CharLimit = m.input.CharLimit
			newInput.SetWidth(m.input.Width())
			newInput.Prompt = m.input.Prompt
			// v2 keeps the prompt and placeholder styling in one Styles value
			// rather than separate fields, so carry the whole thing over.
			newInput.SetStyles(m.input.Styles())
			newInput.Focus()
			m.input = newInput

			return m, nil
		case m.multiLineMode:
			// The lines as they were typed, so a comment ends at the end of
			// its own line rather than taking the lines after it.
			m.multiLineBuffer = append(m.multiLineBuffer, command)
			m.echoInput(command)
			command = strings.Join(m.multiLineBuffer, "\n")
			echoed = true
			m.endStatement()
		}
	}

	// A change the MCP client proposed is confirmed whatever the settings
	// say: nobody typed it.
	proposal := m.isProposal(command)
	m.proposed = ""

	// Check for dangerous commands (skip for AI commands - already checked)
	if proposal || (m.sessionManager != nil && m.sessionManager.RequireConfirmation() && router.IsDangerousCommand(command)) {
		// Show confirmation modal for dangerous commands
		m.modal = NewConfirmationModal(command)
		if proposal {
			m.modal = NewProposalModal(command)
		}

		// Add command to history
		if !echoed {
			m.fullHistoryContent += "\n" + m.styles.AccentText.Render(promptMark+command)
			m.updateHistoryWrapping()
			m.historyViewport.GotoBottom()
		}

		m.input.Reset()
		return m, nil
	}

	// Add to history, as one line: the file holds one statement a line, and
	// the comments are left out, since on one line a comment would take what
	// followed it.
	remembered := historyLine(command)
	m.commandHistory = append(m.commandHistory, remembered)
	m.historyIndex = -1
	m.lastCommand = remembered

	// Save to persistent history
	if m.historyManager != nil {
		if err := m.historyManager.SaveCommand(remembered); err != nil {
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
	result := m.processCommand(command)
	m.lastQueryTime = time.Since(start)

	// Add command to history viewport
	if !echoed {
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render(promptMark+command)
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	// Capture trace data if tracing is enabled and this was a query that returns results
	m.captureTraceData(command)

	// Handle different result types
	logger.DebugfToFile("HandleEnterKey", "Result type (2nd location): %T", result)
	return m.processCommandResult(command, result, start)
}

// historyLine is a statement as history keeps it: without its comments, on
// one line.
func historyLine(command string) string {
	if !strings.ContainsAny(command, "\n\r") {
		return command
	}
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(router.StripComments(command))
}
