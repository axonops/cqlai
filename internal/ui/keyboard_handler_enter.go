package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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

		// Initialize AI conversation view if needed
		if m.aiConversationInput.Value() == "" {
			input := textinput.New()
			input.Placeholder = ""
			input.Prompt = ""  // Remove the ">" prompt
			input.Focus()
			input.CharLimit = 500
			input.Width = m.historyViewport.Width - 10
			m.aiConversationInput = input
			
			// Initialize conversation viewport
			m.aiConversationViewport = viewport.New(m.historyViewport.Width, m.historyViewport.Height)
			m.aiConversationHistory = m.styles.AccentText.Render("🤖 AI Conversation") + "\n" + 
				m.styles.MutedText.Render("────────────────────") + "\n"
		}
		
		// Add user's initial request to conversation
		m.aiConversationHistory += "\n" + m.styles.AccentText.Render("You: ") + userRequest + "\n"
		m.aiConversationViewport.SetContent(m.aiConversationHistory)
		m.aiConversationViewport.GotoBottom()
		
		// Switch to AI conversation view
		m.aiConversationActive = true
		m.viewMode = "ai"
		m.aiProcessing = true
		m.aiConversationInput.SetValue("")
		m.aiConversationInput.Focus()
		m.input.Reset()

		// Start AI generation in background
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
					logger.DebugfToFile("HandleEnterKey", "Got output format from session manager: %v", outputFormat)
				}
				logger.DebugfToFile("HandleEnterKey", "Using output format: %v", outputFormat)
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
					expandStr := FormatExpandTable(allData, m.styles)
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
					logger.DebugToFile("HandleEnterKey", "Formatting output as JSON")
					// JSON format - use table viewport for pagination support
					m.tableHeaders = v.Headers
					m.columnTypes = v.ColumnTypes
					m.hasTable = true
					m.viewMode = "table"

					// Format initial data as JSON
					allData := append([][]string{v.Headers}, m.slidingWindow.Rows...)
					m.lastTableData = allData // Store for pagination
					m.horizontalOffset = 0    // Reset horizontal scroll

					// Convert table data to JSON format
					jsonStr := ""
					// Check if we have a single [json] column from SELECT JSON
					if len(v.Headers) == 1 && v.Headers[0] == "[json]" {
						// This is already JSON from SELECT JSON - just extract it
						for _, row := range m.slidingWindow.Rows {
							if len(row) > 0 {
								jsonStr += row[0] + "\n"
							}
						}
					} else {
						// Convert regular table data to JSON
						for _, row := range m.slidingWindow.Rows {
							jsonMap := make(map[string]interface{})
							for i, header := range v.Headers {
								if i < len(row) {
									jsonMap[header] = row[i]
								}
							}
							jsonBytes, err := json.Marshal(jsonMap)
							if err == nil {
								jsonStr += string(jsonBytes) + "\n"
							} else {
								logger.DebugfToFile("HandleEnterKey", "Error marshaling row to JSON: %v", err)
							}
						}
					}
					logger.DebugfToFile("HandleEnterKey", "Generated JSON with %d characters", len(jsonStr))
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
						// ASCII format - use table viewport for scrolling support
						// Store table data and headers
						m.lastTableData = v.Data
						m.tableHeaders = v.Data[0]    // Store the header row
						m.columnTypes = v.ColumnTypes // Store column types
						m.horizontalOffset = 0
						m.hasTable = true
						m.viewMode = "table"

						// Format as ASCII table
						asciiOutput := FormatASCIITable(v.Data)
						m.tableViewport.SetContent(asciiOutput)
						m.tableViewport.GotoTop() // Start at top of table
					case config.OutputFormatExpand:
						// EXPAND format - use table viewport for scrolling support
						// Store table data and headers
						m.lastTableData = v.Data
						m.tableHeaders = v.Data[0]    // Store the header row
						m.columnTypes = v.ColumnTypes // Store column types
						m.horizontalOffset = 0
						m.hasTable = true
						m.viewMode = "table"

						// Format as expanded vertical table
						expandOutput := FormatExpandTable(v.Data, m.styles)
						m.tableViewport.SetContent(expandOutput)
						m.tableViewport.GotoTop() // Start at top of table
					case config.OutputFormatJSON:
						// JSON format - use table viewport for scrolling support
						// Store table data and headers
						m.lastTableData = v.Data
						m.tableHeaders = v.Data[0]    // Store the header row
						m.columnTypes = v.ColumnTypes // Store column types
						m.horizontalOffset = 0
						m.hasTable = true
						m.viewMode = "table"

						// Check if this is already JSON from SELECT JSON
						jsonOutput := ""
						if len(v.Data) > 1 {
							headers := v.Data[0]
							// Check if we have a single [json] column from SELECT JSON
							if len(headers) == 1 && headers[0] == "[json]" {
								// This is already JSON from SELECT JSON - just extract it
								for _, row := range v.Data[1:] {
									if len(row) > 0 {
										jsonOutput += row[0] + "\n"
									}
								}
							} else {
								// Convert regular table data to JSON
								for _, row := range v.Data[1:] {
									jsonMap := make(map[string]interface{})
									for i, header := range headers {
										if i < len(row) {
											jsonMap[header] = row[i]
										}
									}
									jsonBytes, err := json.Marshal(jsonMap)
									if err == nil {
										jsonOutput += string(jsonBytes) + "\n"
									}
								}
							}
						}
						// Display JSON output in table viewport
						m.tableViewport.SetContent(jsonOutput)
						m.tableViewport.GotoTop() // Start at top of table
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
					keyspace = strings.TrimSpace(keyspace)
					if m.sessionManager != nil {
						m.sessionManager.SetKeyspace(keyspace)
						// Update the status bar
						m.statusBar.Keyspace = keyspace
					}
					// Update the database session's keyspace
					if m.session != nil {
						if err := m.session.SetKeyspace(keyspace); err != nil {
							// Log error but don't fail - the keyspace change was already successful on the server
							logger.DebugfToFile("keyboard_handler_enter", "Failed to update session keyspace: %v", err)
						}
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
	{
		inputText := m.input.Value()
		command = strings.TrimSpace(inputText)
		
		// Check if this is just a comment line
		if strings.HasPrefix(command, "--") || strings.HasPrefix(command, "//") {
			// This is a line comment, process it as complete
			// The router will strip it and return empty result
			m.input.Reset()
			// Add to history to show it was entered
			m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
			// Process the comment (router will strip it)
			_ = router.ProcessCommand(command, m.session)
			return m, nil
		}
		
		// Check for block comment handling
		if strings.HasPrefix(command, "/*") {
			// Starting a block comment
			if strings.Contains(command, "*/") {
				// Single-line block comment - complete
				m.input.Reset()
				m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
				m.historyViewport.SetContent(m.fullHistoryContent)
				m.historyViewport.GotoBottom()
				_ = router.ProcessCommand(command, m.session)
				return m, nil
			} else if !m.multiLineMode {
				// Multi-line block comment - enter special mode
				m.multiLineMode = true
				m.multiLineBuffer = []string{command}
				m.input.Placeholder = "... (in block comment, end with */)"
				m.input.Reset()
				return m, nil
			}
		}
		
		// If we're in multi-line mode and this ends a block comment
		if m.multiLineMode && len(m.multiLineBuffer) > 0 {
			if strings.HasPrefix(m.multiLineBuffer[0], "/*") && strings.Contains(command, "*/") {
				// End of multi-line block comment
				m.multiLineBuffer = append(m.multiLineBuffer, command)
				fullComment := strings.Join(m.multiLineBuffer, "\n")
				m.multiLineMode = false
				m.multiLineBuffer = nil
				m.input.Placeholder = "Enter CQL command..."
				m.input.Reset()
				
				// Add to history
				for _, line := range strings.Split(fullComment, "\n") {
					m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+line)
				}
				m.historyViewport.SetContent(m.fullHistoryContent)
				m.historyViewport.GotoBottom()
				
				// Process (will be stripped as comment)
				_ = router.ProcessCommand(fullComment, m.session)
				return m, nil
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
	
	// Capture trace data if tracing is enabled and this was a query that returns results
	upperCmd := strings.ToUpper(strings.TrimSpace(command))
	if m.session != nil && m.session.Tracing() && 
	   (strings.HasPrefix(upperCmd, "SELECT") || 
	    strings.HasPrefix(upperCmd, "LIST") || 
	    strings.HasPrefix(upperCmd, "DESCRIBE") ||
	    strings.HasPrefix(upperCmd, "DESC")) {
		// Give Cassandra a moment to write trace data
		time.Sleep(50 * time.Millisecond)
		
		// Retrieve trace data
		traceData, traceHeaders, traceInfo, err := m.session.GetTraceData()
		if err == nil && len(traceData) > 0 {
			// Add summary info as a header to the trace content
			summaryLine := ""
			if traceInfo != nil {
				summaryLine = fmt.Sprintf("Trace Session - Coordinator: %s | Total Duration: %d μs\n",
					traceInfo.Coordinator, traceInfo.Duration)
			}
			
			// Combine headers and data into a single table structure
			fullTraceData := make([][]string, 0, len(traceData)+1)
			fullTraceData = append(fullTraceData, traceHeaders)
			fullTraceData = append(fullTraceData, traceData...)
			
			// Store trace data for refreshing
			m.traceData = fullTraceData
			m.traceHeaders = traceHeaders
			m.traceInfo = traceInfo
			m.hasTrace = true
			m.traceHorizontalOffset = 0 // Reset horizontal scroll
			
			// Use the existing formatTableForViewport method temporarily storing the offset
			originalOffset := m.horizontalOffset
			originalData := m.lastTableData
			originalWidth := m.tableWidth
			originalHeaders := m.tableHeaders
			originalColWidths := m.columnWidths
			
			// Set trace data temporarily
			m.horizontalOffset = m.traceHorizontalOffset
			m.lastTableData = fullTraceData
			
			// Format using existing table renderer
			traceTable := m.formatTableForViewport(fullTraceData)
			
			// Store trace-specific values
			m.traceTableWidth = m.tableWidth
			m.traceColumnWidths = m.columnWidths
			
			// Restore original table values
			m.horizontalOffset = originalOffset
			m.lastTableData = originalData
			m.tableWidth = originalWidth
			m.tableHeaders = originalHeaders
			m.columnWidths = originalColWidths
			
			// Prepend summary line to the table
			finalContent := summaryLine + traceTable
			m.traceViewport.SetContent(finalContent)
			m.traceViewport.GotoTop()
		}
	}

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
			expandOutput := FormatExpandTable(allData, m.styles)

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
				expandOutput := FormatExpandTable(v.Data, m.styles)
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
		
		// Check if this is a USE command result and update the keyspace
		if strings.HasPrefix(v, "Now using keyspace ") {
			// Extract the keyspace name
			keyspaceName := strings.TrimPrefix(v, "Now using keyspace ")
			keyspaceName = strings.TrimSpace(keyspaceName)
			
			// Update the session manager and database session
			if m.sessionManager != nil {
				m.sessionManager.SetKeyspace(keyspaceName)
				// Update the status bar
				m.statusBar.Keyspace = keyspaceName
			}
			// Update the database session's keyspace
			if m.session != nil {
				if err := m.session.SetKeyspace(keyspaceName); err != nil {
					// Log error but don't fail - the keyspace change was already successful on the server
					logger.DebugfToFile("keyboard_handler_enter", "Failed to update session keyspace: %v", err)
				}
			}
		}
		
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
