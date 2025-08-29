package ui

import (
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePageUp handles PageUp key press
func (m MainModel) handlePageUp(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	// Cancel exit confirmation if active
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	// If AI modal is showing, scroll its viewport
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		cmd := m.aiModal.Update(msg)
		return m, cmd
	}

	// Scroll the appropriate viewport
	var cmd tea.Cmd
	if m.viewMode == "table" && m.hasTable {
		m.tableViewport, cmd = m.tableViewport.Update(msg)
	} else {
		m.historyViewport, cmd = m.historyViewport.Update(msg)
	}
	return m, cmd
}

// handlePageDown handles PageDown key press
func (m MainModel) handlePageDown(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	// If AI modal is showing, scroll its viewport
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		cmd := m.aiModal.Update(msg)
		return m, cmd
	}

	// Scroll the appropriate viewport
	var cmd tea.Cmd
	if m.viewMode == "table" && m.hasTable {
		m.tableViewport, cmd = m.tableViewport.Update(msg)

		// Check if we need to load more data
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			// If we're within 10 rows of the bottom, load more
			remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset - m.tableViewport.Height
			if remainingRows < 10 {
				// Load the next page
				newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
				if newRows > 0 {
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
							contentStr = FormatExpandTable(allData)
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
	} else {
		m.historyViewport, cmd = m.historyViewport.Update(msg)
	}
	return m, cmd
}

// handleUpArrow handles Up arrow key press
func (m MainModel) handleUpArrow(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleUpArrow called. showAIModal=%v, aiModal.State=%v",
		m.showAIModal, m.aiModal.State)

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

	// If AI modal is showing and in preview state, handle scrolling
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		logger.DebugfToFile("AI", "AI modal scrolling up.")
		// Use the Update method to handle scrolling
		cmd := m.aiModal.Update(msg)
		logger.DebugfToFile("AI", "AI modal scrolled up.")
		return m, cmd
	}

	// If history modal is showing, navigate up
	if m.showHistoryModal && len(m.commandHistory) > 0 {
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
		if m.viewMode == "table" && m.hasTable {
			// Scroll up by one line
			if m.tableViewport.YOffset > 0 {
				m.tableViewport.YOffset--
			}
		} else {
			// Scroll history up by one line
			if m.historyViewport.YOffset > 0 {
				m.historyViewport.YOffset--
			}
		}
		return m, nil
	}

	// Show history modal if there's history to show
	if len(m.commandHistory) > 0 && !m.showHistoryModal {
		m.showHistoryModal = true
		m.historyModalIndex = 0 // Start at most recent
		m.historyModalScrollOffset = 0
	}
	return m, nil
}

// handleDownArrow handles Down arrow key press
func (m MainModel) handleDownArrow(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	logger.DebugfToFile("AI", "handleDownArrow called. showAIModal=%v, aiModal.State=%v",
		m.showAIModal, m.aiModal.State)

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

	// If AI modal is showing and in preview state, handle scrolling
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		logger.DebugfToFile("AI", "AI modal scrolling down.")
		// Use the Update method to handle scrolling
		cmd := m.aiModal.Update(msg)
		logger.DebugfToFile("AI", "AI modal scrolled down.")
		return m, cmd
	}

	// If history modal is showing, navigate down
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		// Remember we're showing newest first, so down means newer (higher index)
		if m.historyModalIndex < len(m.commandHistory)-1 {
			m.historyModalIndex++
			// Adjust scroll offset if selection moves out of view
			if m.historyModalIndex >= m.historyModalScrollOffset+10 {
				m.historyModalScrollOffset = m.historyModalIndex - 9
			}
		}
		return m, nil
	}

	// If Alt is held, scroll viewport down by one line
	if msg.Alt {
		if m.viewMode == "table" && m.hasTable {
			// Scroll down by one line
			maxOffset := m.tableViewport.TotalLineCount() - m.tableViewport.Height
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.tableViewport.YOffset < maxOffset {
				m.tableViewport.YOffset++

				// Check if we need to load more data
				if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
					// If we're within 10 rows of the bottom, load more
					remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset - m.tableViewport.Height
					if remainingRows < 10 {
						// Load the next page
						newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
						if newRows > 0 {
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
									contentStr = FormatExpandTable(allData)
								case config.OutputFormatJSON:
									// Format JSON output - each row is a JSON string
									jsonStr := ""
									for _, row := range m.slidingWindow.Rows {
										if len(row) > 0 {
											jsonStr += row[0] + "\n"
										}
									}
									contentStr = jsonStr
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
		} else {
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

	// If history modal is not showing, down arrow does nothing special
	// (Could show history modal here too if desired)
	return m, nil
}


// handleLeftArrow handles Left arrow key press
func (m MainModel) handleLeftArrow(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	// If AI modal is showing, navigate left
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		m.aiModal.PrevChoice()
		return m, nil
	}

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.PrevChoice()
		return m, nil
	}
	// If Alt is held, scroll table left
	if msg.Alt {
		if m.horizontalOffset > 0 {
			m.horizontalOffset -= 10
			if m.horizontalOffset < 0 {
				m.horizontalOffset = 0
			}
			// Refresh the table view if we have table data
			if m.lastTableData != nil {
				m.refreshTableView()
			}
		}
		return m, nil
	}
	// If Alt is not held, pass the key to the input for cursor movement
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handleRightArrow handles Right arrow key press
func (m MainModel) handleRightArrow(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	// If AI modal is showing, navigate right
	if m.showAIModal && m.aiModal.State == AIModalStatePreview {
		m.aiModal.NextChoice()
		return m, nil
	}

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.NextChoice()
		return m, nil
	}
	// If Alt is held, scroll table right
	if msg.Alt {
		// Only scroll right if table is wider than viewport
		if m.tableWidth > m.tableViewport.Width {
			maxOffset := m.tableWidth - m.tableViewport.Width + 10 // Add some buffer
			if m.horizontalOffset < maxOffset {
				m.horizontalOffset += 10
				if m.horizontalOffset > maxOffset {
					m.horizontalOffset = maxOffset
				}
				// Refresh the table view if we have table data
				if m.lastTableData != nil {
					m.refreshTableView()
				}
			}
		}
		return m, nil
	}
	// If Alt is not held, pass the key to the input for cursor movement
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
