package ui

import (
	"encoding/json"
	"strings"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyboardInput handles keyboard input events
func (m *MainModel) handleKeyboardInput(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	// Check for save modal first (highest priority)
	if m.saveModalActive {
		return m.handleSaveModalKeyboard(msg)
	}

	// Check for AI CQL modal (high priority)
	if m.aiCQLModal != nil && m.aiCQLModal.Active {
		return m.handleAICQLModal(msg)
	}

	// Check for AI selection modal (second priority)
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		return m.handleAISelectionModal(msg)
	}

	// Handle AI conversation view input
	if m.viewMode == "ai" && m.aiConversationActive {
		// Try to handle in AI conversation handler first
		result, cmd := m.handleAIConversationInput(msg)
		if result != nil {
			// Key was handled by AI conversation handler
			return result, cmd
		}
		// Key wasn't handled, let it fall through to main switch
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		return m.handleCtrlC()

	case tea.KeyCtrlD:
		return m.handleCtrlD()

	case tea.KeyCtrlR:
		return m.handleCtrlR()

	case tea.KeyCtrlK:
		return m.handleCtrlK()

	case tea.KeyCtrlU:
		return m.handleCtrlU()

	case tea.KeyCtrlW:
		return m.handleCtrlW()

	case tea.KeyCtrlP:
		return m.handleCtrlP()

	case tea.KeyCtrlA:
		return m.handleCtrlA()

	case tea.KeyCtrlE:
		return m.handleCtrlE()

	case tea.KeyCtrlLeft:
		return m.handleCtrlLeft()

	case tea.KeyCtrlRight:
		return m.handleCtrlRight()

	case tea.KeyCtrlY:
		return m.handleCtrlY()

	case tea.KeyEsc:
		return m.handleEscapeKey()

	case tea.KeyTab:
		// If modal is showing, navigate choices
		if m.modal.Type != ModalNone {
			m.modal.NextChoice()
			return m, nil
		}
		return m.handleTabKey()

	case tea.KeyF2:
		return m.handleF2()

	case tea.KeyF3:
		return m.handleF3()

	case tea.KeyF4:
		return m.handleF4()

	case tea.KeyF5:
		return m.handleF5()

	case tea.KeyF6:
		return m.handleF6()

	case tea.KeySpace:
		return m.handleSpaceKey(msg)

	case tea.KeyPgUp:
		return m.handlePageUp(msg)

	case tea.KeyPgDown:
		return m.handlePageDown(msg)

	case tea.KeyUp:
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchUp()
		}
		return m.handleUpArrow(msg)

	case tea.KeyDown:
		// If in history search mode, navigate search results
		if m.historySearchMode {
			return m.handleHistorySearchDown()
		}
		return m.handleDownArrow(msg)

	case tea.KeyLeft:
		return m.handleLeftArrow(msg)

	case tea.KeyRight:
		return m.handleRightArrow(msg)

	case tea.KeyEnter:
		// If in history search mode, select the current entry
		if m.historySearchMode {
			return m.handleHistorySearchSelect()
		}
		// If history modal is showing, select the current entry
		if m.showHistoryModal {
			return m.handleHistoryModalSelect()
		}
		return m.handleEnterKey()

	default:
		// Handle Alt+N (move to next line in history, same as Down arrow)
		if msg.String() == "alt+n" {
			// If in history search mode, navigate search results
			if m.historySearchMode {
				return m.handleHistorySearchDown()
			}
			return m.handleDownArrow(msg)
		}

		// Handle Alt+D (delete word forward)
		if msg.String() == "alt+d" {
			currentValue := m.input.Value()
			cursorPos := m.input.Position()
			if cursorPos < len(currentValue) {
				// Find the end of the word to cut
				end := cursorPos
				
				// Skip leading spaces
				for end < len(currentValue) && currentValue[end] == ' ' {
					end++
				}
				
				// Find the end of the word
				for end < len(currentValue) && currentValue[end] != ' ' {
					end++
				}
				
				// Store the cut text in clipboard buffer
				m.clipboardBuffer = currentValue[cursorPos:end]
				
				// Remove the word from the input
				newValue := currentValue[:cursorPos] + currentValue[end:]
				m.input.SetValue(newValue)
				// Cursor stays at the same position
			}
			return m, nil
		}

		// Handle navigation mode keys (when in table/trace view with navigation mode active)
		if m.navigationMode && (m.viewMode == "table" || m.viewMode == "trace") {
			switch msg.String() {
			case "j":
				// Single line down
				return m.handleSingleLineDown()
			case "k":
				// Single line up
				return m.handleSingleLineUp()
			case "d":
				// Half page down
				return m.handleHalfPageDown()
			case "u":
				// Half page up
				return m.handleHalfPageUp()
			case "g":
				// Go to top
				return m.handleGoToTop()
			case "G":
				// Go to bottom
				return m.handleGoToBottom()
			case "<":
				// Scroll left by 10 columns
				return m.handlePageLeftScroll()
			case ">":
				// Scroll right by 10 columns
				return m.handlePageRightScroll()
			case "h":
				// Scroll left by one column
				return m.handleHorizontalScrollLeft()
			case "l":
				// Scroll right by one column
				return m.handleHorizontalScrollRight()
			}
		}

		// Cancel exit confirmation on any other key
		if m.confirmExit {
			m.confirmExit = false
			m.input.Placeholder = "Enter CQL command..."
		}

		// If in history search mode, handle typing for search query
		if m.historySearchMode {
			// Update search query based on key press
			switch msg.String() {
			case "backspace", "delete":
				if len(m.historySearchQuery) > 0 {
					m.historySearchQuery = m.historySearchQuery[:len(m.historySearchQuery)-1]
				}
			default:
				// Add character to search query if it's a printable character
				if len(msg.Runes) > 0 && len(m.historySearchQuery) < 100 {
					m.historySearchQuery += string(msg.Runes)
				}
			}

			// Update search results
			if m.historyManager != nil {
				m.historySearchResults = m.historyManager.SearchHistory(m.historySearchQuery)
			} else {
				// Fallback to in-memory history search
				m.historySearchResults = []string{}
				queryLower := strings.ToLower(m.historySearchQuery)
				for i := len(m.commandHistory) - 1; i >= 0; i-- {
					if strings.Contains(strings.ToLower(m.commandHistory[i]), queryLower) {
						m.historySearchResults = append(m.historySearchResults, m.commandHistory[i])
					}
				}
			}

			// Start at the bottom (newest matching command) if results changed
			if len(m.historySearchResults) > 0 {
				m.historySearchIndex = len(m.historySearchResults) - 1
				// Set scroll offset to show the bottom
				if len(m.historySearchResults) > 10 {
					m.historySearchScrollOffset = len(m.historySearchResults) - 10
				} else {
					m.historySearchScrollOffset = 0
				}
			} else {
				m.historySearchIndex = 0
				m.historySearchScrollOffset = 0
			}
			return m, nil
		}

		// Pass the key to the input field for regular typing
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// If completions are showing, update them based on new input
		if m.showCompletions {
			newInput := m.input.Value()
			m.completions = m.completionEngine.Complete(newInput)

			// If no completions match, hide the modal
			if len(m.completions) == 0 {
				m.showCompletions = false
				m.completionIndex = -1
				m.completionScrollOffset = 0
			} else {
				// Reset selection and scroll to first item when list changes
				m.completionIndex = 0
				m.completionScrollOffset = 0
			}
		}

		return m, cmd
	}
}
func (m *MainModel) handleUpArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleUpArrow called.")

	// If completions are showing, navigate up
	if m.showCompletions && len(m.completions) > 0 {
		m.completionIndex--
		if m.completionIndex < 0 {
			m.completionIndex = len(m.completions) - 1
			// Jump to the end of the list
			if len(m.completions) > 10 {
				m.completionScrollOffset = len(m.completions) - 10
			}
		}

		// Adjust scroll offset if selection moves out of view
		if m.completionIndex < m.completionScrollOffset {
			m.completionScrollOffset = m.completionIndex
		}
		return m, nil
	}


	// If history modal is showing, navigate up (go to older command)
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		// Navigate to older commands (decrease index in original array)
		if m.historyModalIndex > 0 {
			m.historyModalIndex--
			// Adjust scroll offset if selection moves out of view
			if m.historyModalIndex < m.historyModalScrollOffset {
				m.historyModalScrollOffset = m.historyModalIndex
			}
		}
		return m, nil
	}

	// If Alt is held, scroll viewport up by one line
	if msg.Alt {
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			// Scroll trace up by one line
			if m.traceViewport.YOffset > 0 {
				m.traceViewport.YOffset--
			}
		case m.viewMode == "table" && m.hasTable:
			// Scroll table up to previous row boundary
			if m.tableViewport.YOffset > 0 {
				newOffset := m.tableViewport.YOffset - 1

				// Find the previous row boundary
				if len(m.tableRowBoundaries) > 0 {
					for i := len(m.tableRowBoundaries) - 1; i >= 0; i-- {
						if m.tableRowBoundaries[i] < m.tableViewport.YOffset {
							newOffset = m.tableRowBoundaries[i]
							break
						}
					}
					// If we didn't find a boundary, go to top
					if newOffset == m.tableViewport.YOffset - 1 {
						newOffset = 0
					}
				}

				m.tableViewport.YOffset = newOffset
			}
		default:
			// Scroll history up by one line
			if m.historyViewport.YOffset > 0 {
				m.historyViewport.YOffset--
			}
		}
		return m, nil
	}

	// Handle command history navigation up
	return m.handleCommandHistoryUp()
}

// handleDownArrow handles Down arrow key press
func (m *MainModel) handleDownArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleDownArrow called.")

	// If completions are showing, navigate down
	if m.showCompletions && len(m.completions) > 0 {
		m.completionIndex = (m.completionIndex + 1) % len(m.completions)

		// Reset scroll to top when wrapping around
		if m.completionIndex == 0 {
			m.completionScrollOffset = 0
		}

		// Adjust scroll offset if selection moves out of view
		if m.completionIndex >= m.completionScrollOffset+10 {
			m.completionScrollOffset = m.completionIndex - 9
		}
		return m, nil
	}


	// If history modal is showing, navigate down (go to newer command)
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		// Navigate to newer commands (increase index in original array)
		if m.historyModalIndex < len(m.commandHistory)-1 {
			m.historyModalIndex++
			// Adjust scroll offset if selection moves out of view
			if m.historyModalIndex >= m.historyModalScrollOffset + 10 {
				m.historyModalScrollOffset = m.historyModalIndex - 9
			}
		}
		return m, nil
	}

	// If Alt is held, scroll viewport down by one line
	if msg.Alt {
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			// Scroll trace down by one line
			maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.traceViewport.YOffset < maxOffset {
				m.traceViewport.YOffset++
			}
		case m.viewMode == "table" && m.hasTable:
			// Scroll table down to next row boundary
			totalLines := m.tableViewport.TotalLineCount()
			viewportHeight := m.tableViewport.Height
			maxOffset := totalLines - viewportHeight
			if maxOffset < 0 {
				maxOffset = 0
			}

			// Check if we're at the end with no more data
			noMoreData := m.slidingWindow == nil || !m.slidingWindow.hasMoreData

			if m.tableViewport.YOffset < maxOffset {
				newOffset := m.tableViewport.YOffset + 1

				// Find the next row boundary
				if len(m.tableRowBoundaries) > 0 {
					for _, boundary := range m.tableRowBoundaries {
						if boundary > m.tableViewport.YOffset {
							newOffset = boundary
							break
						}
					}
				}

				// Special handling for the last boundary (bottom border)
				if noMoreData && len(m.tableRowBoundaries) > 0 {
					lastBoundary := m.tableRowBoundaries[len(m.tableRowBoundaries)-1]
					if newOffset >= lastBoundary - viewportHeight {
						// Position to show the bottom border
						desiredOffset := lastBoundary - viewportHeight + 1
						if desiredOffset < 0 {
							desiredOffset = 0
						}
						if desiredOffset <= maxOffset {
							newOffset = desiredOffset
						} else {
							newOffset = maxOffset
						}
					}
				} else if newOffset > maxOffset {
					newOffset = maxOffset
				}

				m.tableViewport.YOffset = newOffset

				// Check if we need to load more data
				if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
					// If we're within 10 rows of the bottom, load more
					remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset - m.tableViewport.Height
					if remainingRows < 10 {
						// Load the next page
						newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
						if newRows > 0 {
							// Write uncaptured rows to capture file if capturing
							metaHandler := router.GetMetaHandler()
							if metaHandler != nil && metaHandler.IsCapturing() {
								uncapturedRows := m.slidingWindow.GetUncapturedRows()
								if len(uncapturedRows) > 0 {
									// Note: WriteCaptureResult expects the command as first param,
									// but for paged data we use empty string since it's continuation
									_ = metaHandler.WriteCaptureResult("", m.slidingWindow.Headers, uncapturedRows)
									m.slidingWindow.MarkRowsAsCaptured(len(uncapturedRows))
								}
							}

							// Update the table data and refresh the view
							allData := append([][]string{m.slidingWindow.Headers}, m.slidingWindow.Rows...)
							m.lastTableData = allData

							// Format based on current output format
							var contentStr string
							if m.sessionManager != nil {
								switch m.sessionManager.GetOutputFormat() {
								case config.OutputFormatASCII:
									contentStr = FormatASCIITable(allData)
								case config.OutputFormatExpand:
									contentStr = FormatExpandTable(allData, m.styles)
								case config.OutputFormatJSON:
									// Check if we have a single [json] column from SELECT JSON
									if len(m.slidingWindow.Headers) == 1 && m.slidingWindow.Headers[0] == "[json]" {
										// This is already JSON from SELECT JSON - just extract it
										jsonStr := ""
										for _, row := range m.slidingWindow.Rows {
											if len(row) > 0 {
												jsonStr += row[0] + "\n"
											}
										}
										contentStr = jsonStr
									} else {
										// Convert regular table data to JSON
										jsonStr := ""
										for _, row := range m.slidingWindow.Rows {
											jsonMap := make(map[string]interface{})
											for i, header := range m.slidingWindow.Headers {
												if i < len(row) {
													jsonMap[header] = row[i]
												}
											}
											jsonBytes, err := json.Marshal(jsonMap)
											if err == nil {
												jsonStr += string(jsonBytes) + "\n"
											}
										}
										contentStr = jsonStr
									}
								default:
									contentStr = m.formatTableForViewport(allData)
								}
							} else {
								contentStr = m.formatTableForViewport(allData)
							}
							m.tableViewport.SetContent(contentStr)

							// Update row count
							m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
							m.rowCount = int(m.slidingWindow.TotalRowsSeen)
						}
					}
				}
			}
		default:
			// Scroll history down by one line
			maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.historyViewport.YOffset < maxOffset {
				m.historyViewport.YOffset++
			}
		}
		return m, nil
	}

	// Handle command history navigation down
	return m.handleCommandHistoryDown()
}


// handleLeftArrow handles Left arrow key press
