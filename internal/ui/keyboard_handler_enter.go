package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEnterKey handles Enter key press
func (m MainModel) handleEnterKey() (MainModel, tea.Cmd) {
	// Cancel exit confirmation if active
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	command := strings.TrimSpace(m.input.Value())

	// Handle AI command
	if strings.HasPrefix(strings.ToUpper(command), ".AI") {
		// Extract the natural language request
		userRequest := strings.TrimSpace(command[3:])
		if userRequest == "" {
			// Show error for empty request
			historyContent := m.historyViewport.View() + "\n" + m.styles.ErrorText.Render("Error: Please provide a request after .ai")
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()
			m.input.Reset()
			return m, nil
		}
		
		// Create and show AI modal
		m.aiModal = NewAIModal(userRequest)
		m.showAIModal = true
		m.input.Reset()
		
		// Start AI generation in background (will be handled in Update)
		return m, generateAICQL(m.session, userRequest)
	}

	// If completions are showing, accept the selected one
	if m.showCompletions && len(m.completions) > 0 && m.completionIndex >= 0 {
		// Get the selected completion (just the next word)
		selectedCompletion := m.completions[m.completionIndex]

		// Apply the completion by appending to current input
		currentInput := m.input.Value()
		newValue := ""

		// Special case: if input ends with a dot (keyspace.), just append the table name
		if strings.HasSuffix(currentInput, ".") {
			newValue = currentInput + selectedCompletion
		} else if strings.HasSuffix(currentInput, " ") {
			// Just append the completion
			newValue = currentInput + selectedCompletion
		} else {
			// Check if we have a partial word to replace
			lastSpace := strings.LastIndex(currentInput, " ")
			if lastSpace >= 0 {
				// Check if the last word is a complete token that shouldn't be replaced
				lastWord := currentInput[lastSpace+1:]

				// Check for keyspace.table pattern
				if strings.Contains(lastWord, ".") {
					// For keyspace.table patterns, always replace the part after the dot
					// The completion engine returns just the table name
					dotIndex := strings.LastIndex(currentInput, ".")
					newValue = currentInput[:dotIndex+1] + selectedCompletion
				} else if lastWord == "*" || strings.HasSuffix(lastWord, ")") {
					// Don't replace, just append
					newValue = currentInput + " " + selectedCompletion
				} else {
					// Replace the partial word
					newValue = currentInput[:lastSpace+1] + selectedCompletion
				}
			} else {
				// Replace the entire input (single partial word)
				newValue = selectedCompletion
			}
		}

		m.input.SetValue(newValue)
		m.input.SetCursor(len(newValue))

		// Hide completions
		m.showCompletions = false
		m.completions = []string{}
		m.completionIndex = -1
		m.completionScrollOffset = 0
		return m, nil
	}

	// Variable to track if we should execute a command from AI modal
	executeAICommand := false
	
	// Check if AI modal is showing
	if m.showAIModal {
		if m.aiModal.State == AIModalStatePreview {
			switch m.aiModal.Selected {
			case 0: // Cancel
				m.showAIModal = false
				m.aiModal = AIModal{}
				m.input.Placeholder = "Enter CQL command..."
				return m, nil
			case 1: // Execute
				// Get the generated CQL
				command = m.aiModal.CQL
				m.showAIModal = false
				m.aiModal = AIModal{}
				
				// Check if it's a dangerous command
				if m.aiModal.Plan != nil && !m.aiModal.Plan.ReadOnly && m.session.RequireConfirmation() {
					// Show confirmation modal for dangerous AI-generated commands
					m.modal = NewConfirmationModal(command)
					return m, nil
				}
				
				// Mark that we should execute this command
				executeAICommand = true
			case 2: // Edit
				// Put the CQL in the input for editing
				m.input.SetValue(m.aiModal.CQL)
				m.input.SetCursor(len(m.aiModal.CQL))
				m.showAIModal = false
				m.aiModal = AIModal{}
				return m, nil
			}
		} else if m.aiModal.State == AIModalStateError {
			// Close error modal
			m.showAIModal = false
			m.aiModal = AIModal{}
			return m, nil
		}
		// If we're not executing an AI command, just return
		if !executeAICommand {
			return m, nil
		}
	}
	
	// Check if modal is showing FIRST before processing command
	if m.modal.Type != ModalNone {
		if m.modal.Selected == 1 { // "Execute" button
			// Execute the dangerous command
			command = m.modal.Command
			m.modal = Modal{Type: ModalNone}
			m.input.Placeholder = "Enter CQL command..."
			// Continue with normal command execution - process it below

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

			// Add command to history viewport
			historyContent := m.historyViewport.View() + "\n" + m.styles.AccentText.Render("> "+command)
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()

			// Handle different result types
			switch v := result.(type) {
			case db.QueryResult:
				// Query result with metadata - use table viewport
				if len(v.Data) > 0 {
					// Update top bar with query metadata
					m.topBar.QueryTime = v.Duration
					m.topBar.RowCount = v.RowCount
					m.topBar.HasQueryData = true

					m.rowCount = v.RowCount
					// Store table data and headers
					m.lastTableData = v.Data
					m.tableHeaders = v.Data[0]    // Store the header row
					m.columnTypes = v.ColumnTypes // Store column types
					m.horizontalOffset = 0
					m.hasTable = true
					m.viewMode = "table"

					// Format and display in table viewport
					tableStr := m.formatTableForViewport(v.Data)
					m.tableViewport.SetContent(tableStr)
					m.tableViewport.GotoTop() // Start at top of table

					// Write to capture file if capturing
					metaHandler := router.GetMetaHandler()
					if metaHandler != nil && metaHandler.IsCapturing() && len(v.Data) > 1 {
						// Extract headers and rows from data
						headers := v.Data[0]
						rows := v.Data[1:]
						metaHandler.WriteCaptureResult(command, headers, rows)
					}
				}
			case [][]string:
				// Table data without metadata (for backward compatibility)
				if len(v) > 0 {
					m.rowCount = len(v) - 1 // Exclude header
					// Store table data and headers
					m.lastTableData = v
					m.tableHeaders = v[0] // Store the header row
					m.horizontalOffset = 0
					m.hasTable = true
					m.viewMode = "table"

					// Format and display in table viewport
					tableStr := m.formatTableForViewport(v)
					m.tableViewport.SetContent(tableStr)
					m.tableViewport.GotoTop() // Start at top of table

					// Write to capture file if capturing
					metaHandler := router.GetMetaHandler()
					if metaHandler != nil && metaHandler.IsCapturing() && len(v) > 1 {
						// Extract headers and rows from data
						headers := v[0]
						rows := v[1:]
						metaHandler.WriteCaptureResult(command, headers, rows)
					}
				}
			case string:
				// Text result - add to history
				m.tableHeaders = nil
				m.columnWidths = nil
				m.hasTable = false
				m.viewMode = "history"
				// Clear query metadata from top bar
				m.topBar.HasQueryData = false
				// Wrap long lines to prevent truncation
				wrappedResult := wrapLongLines(v, m.historyViewport.Width)
				newContent := m.historyViewport.View() + "\n" + wrappedResult
				m.historyViewport.SetContent(newContent)
				m.historyViewport.GotoBottom()
			case error:
				// Error result - add to history
				m.tableHeaders = nil
				m.columnWidths = nil
				// Clear query metadata from top bar
				m.topBar.HasQueryData = false
				m.hasTable = false
				m.viewMode = "history"
				errorMsg := m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", v))
				newContent := m.historyViewport.View() + "\n" + errorMsg
				m.historyViewport.SetContent(newContent)
				m.historyViewport.GotoBottom()
			}

			m.input.Reset()
			return m, nil
		} else { // "Cancel" button
			// Cancel the command
			m.modal = Modal{Type: ModalNone}
			m.input.Placeholder = "Enter CQL command..."
			m.input.Reset()

			// Add cancellation message to history
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("Command cancelled.")
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()
			return m, nil
		}
	}

	// Process the command from input (unless we're executing an AI command)
	if !executeAICommand {
		inputText := m.input.Value()
		command = strings.TrimSpace(inputText)
		if command == "" && !m.multiLineMode {
			return m, nil
		}
	}

	// Check if this is a CQL statement (not a meta command)
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	isCQLStatement := !strings.HasPrefix(upperCommand, "DESCRIBE") &&
		!strings.HasPrefix(upperCommand, "DESC ") &&
		!strings.HasPrefix(upperCommand, "CONSISTENCY") &&
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
	if isCQLStatement && !executeAICommand {
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
	if !executeAICommand && m.session.RequireConfirmation() && router.IsDangerousCommand(command) {
		// Show confirmation modal for dangerous commands
		m.modal = NewConfirmationModal(command)

		// Add command to history
		historyContent := m.historyViewport.View() + "\n" + m.styles.AccentText.Render("> "+command)
		m.historyViewport.SetContent(historyContent)
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
	upperCommand = strings.ToUpper(command)
	if upperCommand == "EXIT" || upperCommand == "QUIT" {
		return m, tea.Quit
	}

	if upperCommand == "CLEAR" || upperCommand == "CLS" {
		m.historyViewport.SetContent("")
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
		return m, nil
	}

	start := time.Now()
	result := router.ProcessCommand(command, m.session)
	m.lastQueryTime = time.Since(start)

	// Add command to history viewport
	historyContent := m.historyViewport.View() + "\n" + m.styles.AccentText.Render("> "+command)
	m.historyViewport.SetContent(historyContent)
	m.historyViewport.GotoBottom()

	// Handle different result types
	switch v := result.(type) {
	case db.QueryResult:
		// Query result with metadata - use table viewport
		if len(v.Data) > 0 {
			// Update top bar with query metadata
			m.topBar.QueryTime = v.Duration
			m.topBar.RowCount = v.RowCount
			m.topBar.HasQueryData = true

			m.rowCount = v.RowCount
			// Store table data and headers
			m.lastTableData = v.Data
			m.tableHeaders = v.Data[0]    // Store the header row
			m.columnTypes = v.ColumnTypes // Store column types
			m.horizontalOffset = 0
			m.hasTable = true
			m.viewMode = "table"

			// Format and display in table viewport
			tableStr := m.formatTableForViewport(v.Data)
			m.tableViewport.SetContent(tableStr)
			m.tableViewport.GotoTop() // Start at top of table

			// Write to capture file if capturing
			metaHandler := router.GetMetaHandler()
			if metaHandler != nil && metaHandler.IsCapturing() && len(v.Data) > 1 {
				// Extract headers and rows from data
				headers := v.Data[0]
				rows := v.Data[1:]
				metaHandler.WriteCaptureResult(command, headers, rows)
			}
		}
	case [][]string:
		// Table data without metadata (for backward compatibility)
		if len(v) > 0 {
			m.rowCount = len(v) - 1 // Exclude header
			// Store table data and headers
			m.lastTableData = v
			m.tableHeaders = v[0] // Store the header row
			m.horizontalOffset = 0
			m.hasTable = true
			m.viewMode = "table"

			// Format and display in table viewport
			tableStr := m.formatTableForViewport(v)
			m.tableViewport.SetContent(tableStr)
			m.tableViewport.GotoTop() // Start at top of table

			// Write to capture file if capturing
			metaHandler := router.GetMetaHandler()
			if metaHandler != nil && metaHandler.IsCapturing() && len(v) > 1 {
				// Extract headers and rows from data
				headers := v[0]
				rows := v[1:]
				metaHandler.WriteCaptureResult(command, headers, rows)
			}
		}
	case string:
		// Text result - add to history
		m.tableHeaders = nil
		m.columnWidths = nil
		m.hasTable = false
		m.viewMode = "history"
		// Clear query metadata from top bar
		m.topBar.HasQueryData = false
		// Wrap long lines to prevent truncation
		wrappedResult := wrapLongLines(v, m.historyViewport.Width)
		newContent := m.historyViewport.View() + "\n" + wrappedResult
		m.historyViewport.SetContent(newContent)
		m.historyViewport.GotoBottom()
	case error:
		// Error result - add to history
		m.tableHeaders = nil
		m.columnWidths = nil
		m.hasTable = false
		m.viewMode = "history"
		// Clear query metadata from top bar
		m.topBar.HasQueryData = false
		errorMsg := m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", v))
		newContent := m.historyViewport.View() + "\n" + errorMsg
		m.historyViewport.SetContent(newContent)
		m.historyViewport.GotoBottom()
	}

	m.input.Reset()
	return m, nil
}
