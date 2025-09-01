package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEnterKey handles Enter key press
func (m *MainModel) handleEnterKey() (*MainModel, tea.Cmd) {
	// Handle AI info request modal if active
	if m.aiInfoReplyModal != nil && m.aiInfoReplyModal.Active {
		response := m.aiInfoReplyModal.GetResponse()
		if response != "" {
			m.aiInfoReplyModal.Active = false
			return m, func() tea.Msg {
				return AIInfoResponseMsg{
					Response:  response,
					Cancelled: false,
				}
			}
		}
		// Don't close if empty response
		return m, nil
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

	command := strings.TrimSpace(m.input.Value())

	// Note: We removed the follow-up mode check here since we're now handling
	// follow-up questions within the modal itself

	// Handle AI command
	if strings.HasPrefix(strings.ToUpper(command), ".AI") {
		// Log the AI command
		logger.DebugfToFile("AI", "User AI command: %s", command)

		// Add AI command to command history
		m.commandHistory = append(m.commandHistory, command)
		m.historyIndex = -1
		m.lastCommand = command

		// Save to persistent history
		if m.historyManager != nil {
			if err := m.historyManager.SaveCommand(command); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save AI command to history: %v\n", err)
			}
		}

		// Add command to history
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
		m.historyViewport.SetContent(m.fullHistoryContent)
		m.historyViewport.GotoBottom()

		// Extract the natural language request
		userRequest := strings.TrimSpace(command[3:])
		if userRequest == "" {
			// Show error for empty request
			m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: Please provide a request after .ai")
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
			m.input.Reset()
			return m, nil
		}

		// Log the extracted request
		logger.DebugfToFile("AI", "Extracted AI request: %s", userRequest)

		// Clear any existing conversation ID for new .ai command
		// The new conversation ID will be set when we receive the response
		m.aiConversationID = ""

		// Create and show AI modal
		m.aiModal = NewAIModal(userRequest)
		m.showAIModal = true
		m.input.Reset()

		// Start AI generation in background (will be handled in Update)
		return m, generateAICQL(m.session, m.aiConfig, userRequest)
	}

	// If completions are showing, accept the selected one
	if m.showCompletions && len(m.completions) > 0 && m.completionIndex >= 0 {
		// Get the selected completion (just the next word)
		selectedCompletion := m.completions[m.completionIndex]

		// Apply the completion by appending to current input
		currentInput := m.input.Value()
		newValue := ""

		// Special case: if input ends with a dot (keyspace.), just append the table name
		if strings.HasSuffix(currentInput, ".") { //nolint:gocritic // more readable as if
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
				if strings.Contains(lastWord, ".") { //nolint:gocritic // more readable as if
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
		switch m.aiModal.State {
		case AIModalStatePreview:
			// Check if this is an INFO operation
			if m.aiModal.Plan != nil && m.aiModal.Plan.Operation == "INFO" {
				// If there's follow-up input, submit it
				if m.aiModal.FollowUpInput != "" {
					followUpQuestion := m.aiModal.FollowUpInput
					logger.DebugfToFile("AI", "User submitting follow-up question: %s", followUpQuestion)
					logger.DebugfToFile("AI", "Current conversation ID: %s", m.aiConversationID)

					// Clear the input and set state to generating
					m.aiModal.FollowUpInput = ""
					m.aiModal.CursorPosition = 0
					m.aiModal.State = AIModalStateGenerating
					m.aiModal.UserRequest = followUpQuestion

					// Continue the conversation
					return m, continueAIConversation(m.aiConfig, m.aiConversationID, followUpQuestion)
				}
				// If no input, do nothing (stay in info view)
				return m, nil
			} else {
				// Regular CQL operation
				switch m.aiModal.Selected {
				case 0: // Cancel
					m.showAIModal = false
					m.aiModal = AIModal{}
					m.input.Placeholder = "Enter CQL command..."
					return m, nil
				case 1: // Execute
					// Get the generated CQL
					command = m.aiModal.CQL
					logger.DebugfToFile("AI", "User executing AI-generated CQL: %s", command)

					// Store the plan before clearing the modal
					aiPlan := m.aiModal.Plan

					m.showAIModal = false
					m.aiModal = AIModal{}

					// Check if it's a dangerous command
					if aiPlan != nil && !aiPlan.ReadOnly && m.sessionManager != nil && m.sessionManager.RequireConfirmation() {
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
			}
		case AIModalStateError:
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

			// Add command to full history and viewport
			m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()

			// Handle different result types
			logger.DebugfToFile("HandleEnterKey", "Result type: %T", result)
			switch v := result.(type) {
			case db.StreamingQueryResult:
				// Handle streaming query result
				logger.DebugfToFile("HandleEnterKey", "Got StreamingQueryResult with %d headers", len(v.Headers))
				logger.DebugfToFile("HandleEnterKey", "Headers: %v", v.Headers)
				logger.DebugfToFile("HandleEnterKey", "ColumnNames: %v", v.ColumnNames)

				// Initialize sliding window (10MB memory limit, 10000 rows max)
				m.slidingWindow = NewSlidingWindowTable(10000, 10)
				m.slidingWindow.Headers = v.Headers
				m.slidingWindow.ColumnNames = v.ColumnNames
				m.slidingWindow.ColumnTypes = v.ColumnTypes

				// Load initial batch of rows
				initialRows := 0
				maxInitialRows := 100 // Show first 100 rows immediately

				for initialRows < maxInitialRows {
					rowMap := make(map[string]interface{})
					if !v.Iterator.MapScan(rowMap) {
						// Check for iterator error
						if err := v.Iterator.Close(); err != nil {
							logger.DebugfToFile("HandleEnterKey", "Iterator error: %v", err)
							// Show error to user
							m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", err))
							m.historyViewport.SetContent(m.fullHistoryContent)
							m.historyViewport.GotoBottom()
							m.viewMode = "history"
							m.hasTable = false
							m.input.Reset()
							return m, nil
						}
						logger.DebugfToFile("HandleEnterKey", "MapScan returned false after %d rows", initialRows)
						break
					}
					logger.DebugfToFile("HandleEnterKey", "Row %d map keys: %v", initialRows, rowMap)

					// Convert row to string array using original column names
					row := make([]string, len(v.ColumnNames))
					for i, colName := range v.ColumnNames {
						if val, ok := rowMap[colName]; ok {
							if val == nil {
								row[i] = "null"
							} else {
								// Handle different types appropriately
								switch typed := val.(type) {
								case gocql.UUID:
									row[i] = typed.String()
								case []byte:
									row[i] = fmt.Sprintf("0x%x", typed)
								case time.Time:
									row[i] = typed.Format(time.RFC3339)
								default:
									row[i] = fmt.Sprintf("%v", val)
								}
							}
						} else {
							row[i] = "null"
						}
					}

					m.slidingWindow.AddRow(row)
					initialRows++
				}

				logger.DebugfToFile("HandleEnterKey", "Loaded %d initial rows", initialRows)
				logger.DebugfToFile("HandleEnterKey", "Sliding window has %d rows", len(m.slidingWindow.Rows))

				// Check if we got any data
				if initialRows == 0 {
					// No data returned
					_ = v.Iterator.Close()
					m.fullHistoryContent += "\n" + "No results"
					m.historyViewport.SetContent(m.fullHistoryContent)
					m.historyViewport.GotoBottom()
					m.viewMode = "history"
					m.hasTable = false
					m.input.Reset()
					return m, nil
				}

				// Check if there's more data by trying to peek at the next row
				// Store the iterator for later use
				m.slidingWindow.iterator = v.Iterator
				m.slidingWindow.hasMoreData = true // Assume more data until proven otherwise

				// Write initial rows to capture file if capturing
				metaHandler := router.GetMetaHandler()
				if metaHandler != nil && metaHandler.IsCapturing() && len(m.slidingWindow.Rows) > 0 {
					_ = metaHandler.WriteCaptureResult(command, v.Headers, m.slidingWindow.Rows)
					m.slidingWindow.MarkRowsAsCaptured(len(m.slidingWindow.Rows))
				}

				// Update UI
				m.topBar.HasQueryData = true
				m.topBar.QueryTime = time.Since(v.StartTime)
				m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
				m.rowCount = int(m.slidingWindow.TotalRowsSeen)

				logger.DebugfToFile("HandleEnterKey", "TopBar.RowCount set to %d", m.topBar.RowCount)

				// Prepare display based on format
				outputFormat := config.OutputFormatTable
				if m.sessionManager != nil {
					outputFormat = m.sessionManager.GetOutputFormat()
				}
				switch outputFormat {
				case config.OutputFormatExpand:
					// EXPAND format - use table viewport for pagination support
					m.tableHeaders = v.Headers
					m.columnTypes = v.ColumnTypes
					m.hasTable = true
					m.viewMode = "table"

					// Format initial data as expanded vertical format
					allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
					m.lastTableData = allData // Store for pagination
					m.horizontalOffset = 0    // Reset horizontal scroll

					// Format as expanded vertical table
					expandStr := FormatExpandTable(allData)
					m.tableViewport.SetContent(expandStr)
					m.tableViewport.GotoTop()
				case config.OutputFormatASCII:
					// ASCII format - use table viewport for pagination support
					m.tableHeaders = v.Headers
					m.columnTypes = v.ColumnTypes
					m.hasTable = true
					m.viewMode = "table"

					// Format initial data as ASCII table
					allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
					m.lastTableData = allData // Store for pagination
					m.horizontalOffset = 0    // Reset horizontal scroll

					// Format as ASCII table
					asciiStr := FormatASCIITable(allData)
					m.tableViewport.SetContent(asciiStr)
					m.tableViewport.GotoTop()
				case config.OutputFormatJSON:
					// JSON format - use table viewport for pagination support
					m.tableHeaders = v.Headers
					m.columnTypes = v.ColumnTypes
					m.hasTable = true
					m.viewMode = "table"

					// Format initial data as JSON
					allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
					m.lastTableData = allData // Store for pagination
					m.horizontalOffset = 0    // Reset horizontal scroll

					// Format JSON output - each row is a JSON string
					jsonStr := ""
					for _, row := range m.slidingWindow.Rows {
						if len(row) > 0 {
							jsonStr += row[0] + "\n"
						}
					}
					m.tableViewport.SetContent(jsonStr)
					m.tableViewport.GotoTop()
				default:
					// TABLE format - use table viewport
					m.tableHeaders = v.Headers
					m.columnTypes = v.ColumnTypes
					m.hasTable = true
					m.viewMode = "table"

					// Format initial data for display
					allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
					m.lastTableData = allData // Store for horizontal scrolling
					m.horizontalOffset = 0    // Reset horizontal scroll
					logger.DebugfToFile("HandleEnterKey", "Formatting table with %d rows (including header)", len(allData))
					tableStr := m.formatTableForViewport(allData)
					logger.DebugfToFile("HandleEnterKey", "Table string length: %d", len(tableStr))
					m.tableViewport.SetContent(tableStr)
					m.tableViewport.GotoTop()
					logger.DebugfToFile("HandleEnterKey", "Table viewport content set, viewMode: %s", m.viewMode)
				}

			case db.QueryResult:
				// Query result with metadata
				if len(v.Data) > 0 {
					// Update top bar with query metadata
					m.topBar.QueryTime = v.Duration
					m.topBar.RowCount = v.RowCount
					m.topBar.HasQueryData = true

					m.rowCount = v.RowCount

					// Get output format from session manager
					outputFormat := config.OutputFormatTable
					if m.sessionManager != nil {
						outputFormat = m.sessionManager.GetOutputFormat()
					}
					logger.DebugfToFile("keyboard_handler_enter", "QueryResult Format: %v", outputFormat)

					// Check output format
					switch outputFormat {
					case config.OutputFormatASCII:
						// Format as ASCII table in the UI layer
						asciiOutput := FormatASCIITable(v.Data)
						// Display ASCII formatted output in history viewport
						m.fullHistoryContent += "\n" + asciiOutput
						m.historyViewport.SetContent(m.fullHistoryContent)
						m.historyViewport.GotoBottom()
						m.viewMode = "history"
						m.hasTable = false
					case config.OutputFormatExpand:
						// Format as expanded vertical table in the UI layer
						expandOutput := FormatExpandTable(v.Data)
						// Display expanded output in history viewport
						m.fullHistoryContent += "\n" + expandOutput
						m.historyViewport.SetContent(m.fullHistoryContent)
						m.historyViewport.GotoBottom()
						m.viewMode = "history"
						m.hasTable = false
					case config.OutputFormatJSON:
						// JSON format - display raw JSON rows
						// With SELECT JSON, each row is a JSON string
						jsonOutput := ""
						if len(v.Data) > 1 {
							// Skip header row for JSON output
							for _, row := range v.Data[1:] {
								if len(row) > 0 {
									jsonOutput += row[0] + "\n"
								}
							}
						}
						// Display JSON output in history viewport
						m.fullHistoryContent += "\n" + jsonOutput
						m.historyViewport.SetContent(m.fullHistoryContent)
						m.historyViewport.GotoBottom()
						m.viewMode = "history"
						m.hasTable = false
					default:
						// Use table viewport for TABLE format
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
					}

					// Write to capture file if capturing
					metaHandler := router.GetMetaHandler()
					if metaHandler != nil && metaHandler.IsCapturing() && len(v.Data) > 1 {
						// Extract headers and rows from data
						headers := v.Data[0]
						rows := v.Data[1:]
						_ = metaHandler.WriteCaptureResult(command, headers, rows)
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
						_ = metaHandler.WriteCaptureResult(command, headers, rows)
					}
				}
			case string:
				// Check if this is a USE command result
				if strings.HasPrefix(v, "Now using keyspace ") {
					// Extract keyspace name and update session manager
					keyspace := strings.TrimPrefix(v, "Now using keyspace ")
					if m.sessionManager != nil {
						m.sessionManager.SetKeyspace(keyspace)
					}
				}
				// Text result - add to history
				m.tableHeaders = nil
				m.columnWidths = nil
				m.hasTable = false
				m.viewMode = "history"
				// Clear query metadata from top bar
				m.topBar.HasQueryData = false
				// Wrap long lines to prevent truncation
				wrappedResult := wrapLongLines(v, m.historyViewport.Width)
				
				// Save the current line count before adding new content
				oldLineCount := m.historyViewport.TotalLineCount()
				
				m.fullHistoryContent += "\n" + wrappedResult
				m.historyViewport.SetContent(m.fullHistoryContent)
				
				// Write to capture file if capturing
				metaHandler := router.GetMetaHandler()
				if metaHandler != nil && metaHandler.IsCapturing() {
					_ = metaHandler.WriteCaptureText(command, v)
				}
				
				// Check if this is a DESCRIBE command
				upperCmd := strings.ToUpper(strings.TrimSpace(command))
				isDescribe := strings.HasPrefix(upperCmd, "DESCRIBE") || strings.HasPrefix(upperCmd, "DESC ")
				
				if isDescribe {
					// For DESCRIBE commands, position to show the start of the output
					// Calculate where the new output starts
					if oldLineCount > 0 {
						m.historyViewport.YOffset = oldLineCount + 1 // +1 for the newline we added
					} else {
						m.historyViewport.YOffset = 0
					}
				} else {
					// For other commands, scroll to bottom as usual
					m.historyViewport.GotoBottom()
				}
			case error:
				// Error result - add to history
				m.tableHeaders = nil
				m.columnWidths = nil
				// Clear query metadata from top bar
				m.topBar.HasQueryData = false
				m.hasTable = false
				m.viewMode = "history"
				errorMsg := m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", v))
				m.fullHistoryContent += "\n" + errorMsg
				m.historyViewport.SetContent(m.fullHistoryContent)
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
			m.fullHistoryContent += "\n" + m.styles.MutedText.Render("Command cancelled.")
			m.historyViewport.SetContent(m.fullHistoryContent)
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
	if !executeAICommand && m.sessionManager != nil && m.sessionManager.RequireConfirmation() && router.IsDangerousCommand(command) {
		// Show confirmation modal for dangerous commands
		m.modal = NewConfirmationModal(command)

		// Add command to history
		m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
		m.historyViewport.SetContent(m.fullHistoryContent)
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
		m.fullHistoryContent = ""
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
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	m.historyViewport.SetContent(m.fullHistoryContent)
	m.historyViewport.GotoBottom()

	// Handle different result types
	logger.DebugfToFile("HandleEnterKey", "Result type (2nd location): %T", result)
	switch v := result.(type) {
	case db.StreamingQueryResult:
		// Handle streaming query result
		logger.DebugfToFile("HandleEnterKey", "Got StreamingQueryResult with %d headers", len(v.Headers))
		logger.DebugfToFile("HandleEnterKey", "Headers: %v", v.Headers)
		logger.DebugfToFile("HandleEnterKey", "ColumnNames: %v", v.ColumnNames)

		// Initialize sliding window (10MB memory limit, 10000 rows max)
		m.slidingWindow = NewSlidingWindowTable(10000, 10)
		m.slidingWindow.Headers = v.Headers
		m.slidingWindow.ColumnNames = v.ColumnNames
		m.slidingWindow.ColumnTypes = v.ColumnTypes

		// Load initial batch of rows
		initialRows := 0
		maxInitialRows := 100 // Show first 100 rows immediately

		for initialRows < maxInitialRows {
			rowMap := make(map[string]interface{})
			if !v.Iterator.MapScan(rowMap) {
				// Check for iterator error
				if err := v.Iterator.Close(); err != nil {
					logger.DebugfToFile("HandleEnterKey", "Iterator error: %v", err)
					// Show error to user
					m.fullHistoryContent += "\n" + m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", err))
					m.historyViewport.SetContent(m.fullHistoryContent)
					m.historyViewport.GotoBottom()
					m.viewMode = "history"
					m.hasTable = false
					m.input.Reset()
					return m, nil
				}
				logger.DebugfToFile("HandleEnterKey", "MapScan returned false after %d rows", initialRows)
				break
			}
			logger.DebugfToFile("HandleEnterKey", "Row %d map keys: %v", initialRows, rowMap)

			// Convert row to string array using original column names
			row := make([]string, len(v.ColumnNames))
			for i, colName := range v.ColumnNames {
				if val, ok := rowMap[colName]; ok {
					if val == nil {
						row[i] = "null"
					} else {
						// Handle different types appropriately
						switch typed := val.(type) {
						case gocql.UUID:
							row[i] = typed.String()
						case []byte:
							row[i] = fmt.Sprintf("0x%x", typed)
						case time.Time:
							row[i] = typed.Format(time.RFC3339)
						default:
							row[i] = fmt.Sprintf("%v", val)
						}
					}
				} else {
					row[i] = "null"
				}
			}

			m.slidingWindow.AddRow(row)
			initialRows++
		}

		logger.DebugfToFile("HandleEnterKey", "Loaded %d initial rows", initialRows)
		logger.DebugfToFile("HandleEnterKey", "Sliding window has %d rows", len(m.slidingWindow.Rows))

		// Check if we got any data
		if initialRows == 0 {
			// No data returned
			_ = v.Iterator.Close()
			m.fullHistoryContent += "\n" + "No results"
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
			m.viewMode = "history"
			m.hasTable = false
			m.input.Reset()
			return m, nil
		}

		// Check if there's more data by trying to peek at the next row
		// Store the iterator for later use
		m.slidingWindow.iterator = v.Iterator
		m.slidingWindow.hasMoreData = true // Assume more data until proven otherwise

		// Update UI
		m.topBar.HasQueryData = true
		m.topBar.QueryTime = time.Since(v.StartTime)
		m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)

		logger.DebugfToFile("HandleEnterKey", "TopBar.RowCount set to %d", m.topBar.RowCount)

		// Get output format from session manager
		outputFormat := config.OutputFormatTable
		if m.sessionManager != nil {
			outputFormat = m.sessionManager.GetOutputFormat()
		}

		// Prepare display based on format
		switch outputFormat {
		case config.OutputFormatExpand:
			// Format as expanded vertical table
			allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
			expandOutput := FormatExpandTable(allData)

			// Add note if there's more data
			if m.slidingWindow.hasMoreData {
				expandOutput += fmt.Sprintf("\n(Showing first %d rows, more available - use OUTPUT TABLE for pagination)\n", len(m.slidingWindow.Rows))
			}

			// Display expanded output in history viewport
			m.fullHistoryContent += "\n" + expandOutput
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
			m.viewMode = "history"
			m.hasTable = false

			// Close iterator since we won't paginate in expand mode
			if v.Iterator != nil {
				_ = v.Iterator.Close()
			}
			m.slidingWindow.iterator = nil
			m.slidingWindow.hasMoreData = false
		case config.OutputFormatASCII:
			// Format as ASCII table
			allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
			asciiOutput := FormatASCIITable(allData)

			// Add note if there's more data
			if m.slidingWindow.hasMoreData {
				asciiOutput += fmt.Sprintf("\n(Showing first %d rows, more available - use OUTPUT TABLE for pagination)\n", len(m.slidingWindow.Rows))
			}

			// Display ASCII formatted output in history viewport
			m.fullHistoryContent += "\n" + asciiOutput
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
			m.viewMode = "history"
			m.hasTable = false

			// Close iterator since we won't paginate in ASCII mode
			if v.Iterator != nil {
				_ = v.Iterator.Close()
			}
			m.slidingWindow.iterator = nil
			m.slidingWindow.hasMoreData = false
		default:
			// TABLE format - use table viewport
			m.tableHeaders = v.Headers
			m.columnTypes = v.ColumnTypes
			m.hasTable = true
			m.viewMode = "table"

			// Format initial data for display
			allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
			m.lastTableData = allData // Store for horizontal scrolling
			m.horizontalOffset = 0    // Reset horizontal scroll
			logger.DebugfToFile("HandleEnterKey", "Formatting table with %d rows (including header)", len(allData))
			tableStr := m.formatTableForViewport(allData)
			logger.DebugfToFile("HandleEnterKey", "Table string length: %d", len(tableStr))
			m.tableViewport.SetContent(tableStr)
			m.tableViewport.GotoTop()
			logger.DebugfToFile("HandleEnterKey", "Table viewport content set, viewMode: %s", m.viewMode)
		}

		// Write to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() && len(m.slidingWindow.Rows) > 0 {
			_ = metaHandler.WriteCaptureResult(command, v.Headers, m.slidingWindow.Rows)
		}

	case db.QueryResult:
		// Query result with metadata
		if len(v.Data) > 0 {
			// Update top bar with query metadata
			m.topBar.QueryTime = v.Duration
			m.topBar.RowCount = v.RowCount
			m.topBar.HasQueryData = true
			m.rowCount = v.RowCount

			// Get output format from session manager
			outputFormat := config.OutputFormatTable
			if m.sessionManager != nil {
				outputFormat = m.sessionManager.GetOutputFormat()
			}

			// Check output format
			switch outputFormat {
			case config.OutputFormatASCII:
				// Format as ASCII table in the UI layer
				asciiOutput := FormatASCIITable(v.Data)
				// Display ASCII formatted output in history viewport
				m.fullHistoryContent += "\n" + asciiOutput
				m.historyViewport.SetContent(m.fullHistoryContent)
				m.historyViewport.GotoBottom()
				m.viewMode = "history"
				m.hasTable = false
			case config.OutputFormatExpand:
				// Format as expanded vertical table in the UI layer
				expandOutput := FormatExpandTable(v.Data)
				// Display expanded output in history viewport
				m.fullHistoryContent += "\n" + expandOutput
				m.historyViewport.SetContent(m.fullHistoryContent)
				m.historyViewport.GotoBottom()
				m.viewMode = "history"
				m.hasTable = false
			default:
				// Use table viewport for TABLE format
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
			}

			// Write to capture file if capturing
			metaHandler := router.GetMetaHandler()
			if metaHandler != nil && metaHandler.IsCapturing() && len(v.Data) > 1 {
				// Extract headers and rows from data
				headers := v.Data[0]
				rows := v.Data[1:]
				_ = metaHandler.WriteCaptureResult(command, headers, rows)
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
				_ = metaHandler.WriteCaptureResult(command, headers, rows)
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
		
		// Save the current line count before adding new content
		oldLineCount := m.historyViewport.TotalLineCount()
		
		m.fullHistoryContent += "\n" + wrappedResult
		m.historyViewport.SetContent(m.fullHistoryContent)
		
		// Write to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() {
			_ = metaHandler.WriteCaptureText(command, v)
		}
		
		// Check if this is a DESCRIBE command
		upperCmd := strings.ToUpper(strings.TrimSpace(command))
		isDescribe := strings.HasPrefix(upperCmd, "DESCRIBE") || strings.HasPrefix(upperCmd, "DESC ")
		
		if isDescribe {
			// For DESCRIBE commands, position to show the start of the output
			// Calculate where the new output starts
			if oldLineCount > 0 {
				m.historyViewport.YOffset = oldLineCount + 1 // +1 for the newline we added
			} else {
				m.historyViewport.YOffset = 0
			}
		} else {
			// For other commands, scroll to bottom as usual
			m.historyViewport.GotoBottom()
		}
	case error:
		// Error result - add to history
		m.tableHeaders = nil
		m.columnWidths = nil
		m.hasTable = false
		m.viewMode = "history"
		// Clear query metadata from top bar
		m.topBar.HasQueryData = false
		errorMsg := m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", v))
		m.fullHistoryContent += "\n" + errorMsg
		m.historyViewport.SetContent(m.fullHistoryContent)
		m.historyViewport.GotoBottom()
	}

	m.input.Reset()
	return m, nil
}
