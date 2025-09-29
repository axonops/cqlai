package ui

import (
	"encoding/json"
	
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePageUp handles PageUp key press
func (m *MainModel) handlePageUp(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	// Cancel exit confirmation if active
	if m.confirmExit {
		m.confirmExit = false
		m.input.Placeholder = "Enter CQL command..."
		return m, nil
	}

	// If input has focus and contains text, page left in the input
	if m.input.Focused() && len(m.input.Value()) > 0 {
		cursorPos := m.input.Position()
		// Page left by half the viewport width
		pageSize := m.windowWidth / 2
		if pageSize < 20 {
			pageSize = 20
		}
		newPos := cursorPos - pageSize
		if newPos < 0 {
			newPos = 0
		}
		m.input.SetCursor(newPos)
		return m, nil
	}

	// Scroll by 80% of viewport height to maintain context
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = int(float64(m.traceViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.traceViewport.YOffset - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.traceViewport.YOffset = newOffset
	case m.viewMode == "table" && m.hasTable:
		scrollAmount = int(float64(m.tableViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.tableViewport.YOffset - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}

		// Snap to row boundary to avoid cutting through multi-line cells
		if len(m.tableRowBoundaries) > 0 && newOffset > 0 {
			// Find the closest row boundary that's >= newOffset
			// When scrolling up, we want to align to the start of a row
			bestOffset := 0
			for _, boundary := range m.tableRowBoundaries {
				if boundary <= newOffset {
					bestOffset = boundary
				} else {
					break
				}
			}
			newOffset = bestOffset
		}

		m.tableViewport.YOffset = newOffset
	default:
		scrollAmount = int(float64(m.historyViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.historyViewport.YOffset - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.historyViewport.YOffset = newOffset
	}
	return m, nil
}

// handlePageDown handles PageDown key press
func (m *MainModel) handlePageDown(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	// If input has focus and contains text, page right in the input
	if m.input.Focused() && len(m.input.Value()) > 0 {
		currentValue := m.input.Value()
		cursorPos := m.input.Position()
		valueLen := len(currentValue)
		// Page right by half the viewport width
		pageSize := m.windowWidth / 2
		if pageSize < 20 {
			pageSize = 20
		}
		newPos := cursorPos + pageSize
		if newPos > valueLen {
			newPos = valueLen
		}
		m.input.SetCursor(newPos)
		return m, nil
	}

	// Scroll by 80% of viewport height to maintain context
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = int(float64(m.traceViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.traceViewport.YOffset + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}
		m.traceViewport.YOffset = newOffset
	case m.viewMode == "table" && m.hasTable:
		scrollAmount = int(float64(m.tableViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		// Calculate max offset - ensure we can see the last line
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height
		// The max offset should allow the last line to be visible at the bottom
		maxOffset := totalLines - viewportHeight
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.tableViewport.YOffset + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}

		// Check if we're at the bottom and there's no more data to load
		atBottom := newOffset >= maxOffset
		noMoreData := m.slidingWindow == nil || !m.slidingWindow.hasMoreData
		logger.DebugfToFile("Nav", "PageDown: totalLines=%d, viewportHeight=%d, maxOffset=%d, newOffset=%d, atBottom=%v",
			totalLines, viewportHeight, maxOffset, newOffset, atBottom)

		// Snap to row boundary to avoid cutting through multi-line cells
		if len(m.tableRowBoundaries) > 0 {
			// Special case: if we're at the bottom with no more data,
			// make sure the last boundary (bottom border) is visible
			if atBottom && noMoreData && len(m.tableRowBoundaries) > 0 {
				// The last boundary is the bottom border line
				lastBoundary := m.tableRowBoundaries[len(m.tableRowBoundaries)-1]
				// Position the viewport so the bottom border is at the bottom of the screen
				// This means the offset should be: lastBoundary - viewportHeight + 1
				// (+1 to include the border line itself)
				desiredOffset := lastBoundary - viewportHeight + 1
				if desiredOffset < 0 {
					desiredOffset = 0
				}
				if desiredOffset <= maxOffset {
					newOffset = desiredOffset
				} else {
					newOffset = maxOffset
				}
				logger.DebugfToFile("Nav", "PageDown at bottom: lastBoundary=%d, desiredOffset=%d, maxOffset=%d, finalOffset=%d",
					lastBoundary, desiredOffset, maxOffset, newOffset)
			} else {
				// Normal case: find the closest row boundary that's <= newOffset
				bestOffset := m.tableViewport.YOffset
				for _, boundary := range m.tableRowBoundaries {
					if boundary <= newOffset && boundary > bestOffset {
						bestOffset = boundary
					}
				}
				newOffset = bestOffset
			}
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
							// Use AppendCaptureRows for continuation data
							_ = metaHandler.AppendCaptureRows(uncapturedRows)
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
	default:
		scrollAmount = int(float64(m.historyViewport.Height) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.historyViewport.YOffset + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}
		m.historyViewport.YOffset = newOffset
	}
	return m, nil
}

// handleUpArrow handles Up arrow key press
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

	// Show history modal if there's history to show
	if len(m.commandHistory) > 0 && !m.showHistoryModal {
		m.showHistoryModal = true
		m.historyModalIndex = len(m.commandHistory) - 1 // Start at most recent (last in array)
		// Set scroll offset to show the bottom of the list (newest commands)
		// If we have more than maxShow (10) items, scroll to show the last 10
		if len(m.commandHistory) > 10 {
			m.historyModalScrollOffset = len(m.commandHistory) - 10
		} else {
			m.historyModalScrollOffset = 0
		}
		// Debug log
		logger.DebugfToFile("History", "Opening history modal: index=%d, scroll=%d, total=%d",
			m.historyModalIndex, m.historyModalScrollOffset, len(m.commandHistory))
	}
	return m, nil
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

	// If history modal is not showing, down arrow does nothing special
	// (Could show history modal here too if desired)
	return m, nil
}


// handleLeftArrow handles Left arrow key press
func (m *MainModel) handleLeftArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.PrevChoice()
		return m, nil
	}
	// If Alt is held, scroll table/trace left
	if msg.Alt {
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			// Scroll trace table left
			if m.traceHorizontalOffset > 0 {
				m.traceHorizontalOffset -= 10
				if m.traceHorizontalOffset < 0 {
					m.traceHorizontalOffset = 0
				}
				// Refresh the trace view using existing table renderer
				m.refreshTraceView()
			}
		case m.viewMode == "table" && m.hasTable:
			// Scroll data table left
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
		}
		return m, nil
	}
	// If Alt is not held, pass the key to the input for cursor movement
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handleRightArrow handles Right arrow key press
func (m *MainModel) handleRightArrow(msg tea.KeyMsg) (*MainModel, tea.Cmd) {

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.NextChoice()
		return m, nil
	}
	// If Alt is held, scroll table/trace right
	if msg.Alt {
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			// Scroll trace table right
			if m.traceTableWidth > m.traceViewport.Width {
				maxOffset := m.traceTableWidth - m.traceViewport.Width + 10 // Add some buffer
				if m.traceHorizontalOffset < maxOffset {
					m.traceHorizontalOffset += 10
					if m.traceHorizontalOffset > maxOffset {
						m.traceHorizontalOffset = maxOffset
					}
					// Refresh the trace view using existing table renderer
					m.refreshTraceView()
				}
			}
		case m.viewMode == "table" && m.hasTable:
			// Scroll data table right
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
		}
		return m, nil
	}
	// If Alt is not held, pass the key to the input for cursor movement
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
