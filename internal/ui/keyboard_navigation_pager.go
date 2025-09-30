package ui

import (
	"encoding/json"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleSingleLineDown scrolls down by one line (j key)
func (m *MainModel) handleSingleLineDown() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceViewport.YOffset < m.traceViewport.TotalLineCount()-m.traceViewport.Height {
			m.traceViewport.YOffset++
		}
	case m.viewMode == "table" && m.hasTable:
		maxOffset := m.tableViewport.TotalLineCount() - m.tableViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.tableViewport.YOffset < maxOffset {
			newOffset := m.tableViewport.YOffset + 1

			// Respect row boundaries for multi-line cells
			if len(m.tableRowBoundaries) > 0 {
				// Find next row boundary
				for _, boundary := range m.tableRowBoundaries {
					if boundary > m.tableViewport.YOffset {
						newOffset = boundary
						break
					}
				}
			}

			if newOffset > maxOffset {
				newOffset = maxOffset
			}
			m.tableViewport.YOffset = newOffset

			// Check if we need to load more data (same as Alt+Down)
			if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
				remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset - m.tableViewport.Height
				if remainingRows < 10 {
					m.loadMoreTableDataHelper()
				}
			}
		}
	default:
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

// handleSingleLineUp scrolls up by one line (k key)
func (m *MainModel) handleSingleLineUp() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceViewport.YOffset > 0 {
			m.traceViewport.YOffset--
		}
	case m.viewMode == "table" && m.hasTable:
		if m.tableViewport.YOffset > 0 {
			newOffset := m.tableViewport.YOffset - 1

			// Respect row boundaries for multi-line cells
			if len(m.tableRowBoundaries) > 0 {
				// Find current row
				currentRowIdx := -1
				for i, boundary := range m.tableRowBoundaries {
					if boundary >= m.tableViewport.YOffset {
						currentRowIdx = i
						break
					}
				}

				// Move to previous row boundary
				if currentRowIdx > 0 {
					newOffset = m.tableRowBoundaries[currentRowIdx-1]
				} else if currentRowIdx == 0 {
					newOffset = 0
				}
			}

			m.tableViewport.YOffset = newOffset
		}
	default:
		if m.historyViewport.YOffset > 0 {
			m.historyViewport.YOffset--
		}
	}
	return m, nil
}

// handleHalfPageDown scrolls down by half a page (d key)
func (m *MainModel) handleHalfPageDown() (*MainModel, tea.Cmd) {
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = m.traceViewport.Height / 2
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
		scrollAmount = m.tableViewport.Height / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.tableViewport.TotalLineCount() - m.tableViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.tableViewport.YOffset + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}

		// Snap to row boundary if we have multi-line cells
		if len(m.tableRowBoundaries) > 0 {
			bestOffset := m.tableViewport.YOffset
			for _, boundary := range m.tableRowBoundaries {
				if boundary <= newOffset && boundary > bestOffset {
					bestOffset = boundary
				}
			}
			newOffset = bestOffset
		}

		m.tableViewport.YOffset = newOffset

		// Check if we need to load more data
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset - m.tableViewport.Height
			if remainingRows < 10 {
				m.loadMoreTableDataHelper()
			}
		}

	default:
		scrollAmount = m.historyViewport.Height / 2
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

// handleHalfPageUp scrolls up by half a page (u key)
func (m *MainModel) handleHalfPageUp() (*MainModel, tea.Cmd) {
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = m.traceViewport.Height / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.traceViewport.YOffset - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.traceViewport.YOffset = newOffset

	case m.viewMode == "table" && m.hasTable:
		scrollAmount = m.tableViewport.Height / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.tableViewport.YOffset - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}

		// Snap to row boundary if we have multi-line cells
		if len(m.tableRowBoundaries) > 0 && newOffset > 0 {
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
		scrollAmount = m.historyViewport.Height / 2
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

// handleGoToTop jumps to the top (g key)
func (m *MainModel) handleGoToTop() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		m.traceViewport.YOffset = 0
	case m.viewMode == "table" && m.hasTable:
		m.tableViewport.YOffset = 0
	default:
		m.historyViewport.YOffset = 0
	}
	return m, nil
}

// handleGoToBottom jumps to the bottom (G key)
func (m *MainModel) handleGoToBottom() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.traceViewport.YOffset = maxOffset

	case m.viewMode == "table" && m.hasTable:
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height
		maxOffset := totalLines - viewportHeight
		if maxOffset < 0 {
			maxOffset = 0
		}

		// Check if there's no more data
		noMoreData := m.slidingWindow == nil || !m.slidingWindow.hasMoreData

		// If we have row boundaries and no more data, ensure bottom border is visible
		if len(m.tableRowBoundaries) > 0 && noMoreData {
			lastBoundary := m.tableRowBoundaries[len(m.tableRowBoundaries)-1]
			desiredOffset := lastBoundary - viewportHeight + 1
			if desiredOffset < 0 {
				desiredOffset = 0
			}
			if desiredOffset <= maxOffset {
				maxOffset = desiredOffset
			}
		}

		m.tableViewport.YOffset = maxOffset

		// Load all remaining data if we have more
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			// Could implement loading all data here if desired
			logger.DebugfToFile("Nav", "Jump to bottom requested but more data available")
		}

	default:
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.historyViewport.YOffset = maxOffset
	}
	return m, nil
}

// handleHorizontalScrollLeft scrolls left (< or , key)
func (m *MainModel) handleHorizontalScrollLeft() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceHorizontalOffset > 0 {
			m.traceHorizontalOffset -= 10
			if m.traceHorizontalOffset < 0 {
				m.traceHorizontalOffset = 0
			}
			m.refreshTraceView()
		}
	case m.viewMode == "table" && m.hasTable:
		if m.horizontalOffset > 0 {
			m.horizontalOffset -= 10
			if m.horizontalOffset < 0 {
				m.horizontalOffset = 0
			}
			if m.lastTableData != nil {
				m.refreshTableView()
			}
		}
	}
	return m, nil
}

// handlePageLeftScroll scrolls left by page width (Alt+PageUp)
func (m *MainModel) handlePageLeftScroll() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceHorizontalOffset > 0 {
			scrollAmount := m.traceViewport.Width / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			m.traceHorizontalOffset -= scrollAmount
			if m.traceHorizontalOffset < 0 {
				m.traceHorizontalOffset = 0
			}
			m.refreshTraceView()
		}
	case m.viewMode == "table" && m.hasTable:
		if m.horizontalOffset > 0 {
			scrollAmount := m.tableViewport.Width / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			m.horizontalOffset -= scrollAmount
			if m.horizontalOffset < 0 {
				m.horizontalOffset = 0
			}
			if m.lastTableData != nil {
				m.refreshTableView()
			}
		}
	}
	return m, nil
}

// handlePageRightScroll scrolls right by page width (Alt+PageDown)
func (m *MainModel) handlePageRightScroll() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceTableWidth > m.traceViewport.Width {
			scrollAmount := m.traceViewport.Width / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			maxOffset := m.traceTableWidth - m.traceViewport.Width + 10
			m.traceHorizontalOffset += scrollAmount
			if m.traceHorizontalOffset > maxOffset {
				m.traceHorizontalOffset = maxOffset
			}
			m.refreshTraceView()
		}
	case m.viewMode == "table" && m.hasTable:
		if m.tableWidth > m.tableViewport.Width {
			scrollAmount := m.tableViewport.Width / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			maxOffset := m.tableWidth - m.tableViewport.Width + 10
			m.horizontalOffset += scrollAmount
			if m.horizontalOffset > maxOffset {
				m.horizontalOffset = maxOffset
			}
			if m.lastTableData != nil {
				m.refreshTableView()
			}
		}
	}
	return m, nil
}

// handleHorizontalScrollRight scrolls right (> or . key)
func (m *MainModel) handleHorizontalScrollRight() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceTableWidth > m.traceViewport.Width {
			maxOffset := m.traceTableWidth - m.traceViewport.Width + 10
			if m.traceHorizontalOffset < maxOffset {
				m.traceHorizontalOffset += 10
				if m.traceHorizontalOffset > maxOffset {
					m.traceHorizontalOffset = maxOffset
				}
				m.refreshTraceView()
			}
		}
	case m.viewMode == "table" && m.hasTable:
		if m.tableWidth > m.tableViewport.Width {
			maxOffset := m.tableWidth - m.tableViewport.Width + 10
			if m.horizontalOffset < maxOffset {
				m.horizontalOffset += 10
				if m.horizontalOffset > maxOffset {
					m.horizontalOffset = maxOffset
				}
				if m.lastTableData != nil {
					m.refreshTableView()
				}
			}
		}
	}
	return m, nil
}

// loadMoreTableDataHelper loads more rows when scrolling near the bottom
func (m *MainModel) loadMoreTableDataHelper() {
	if m.slidingWindow == nil || !m.slidingWindow.hasMoreData || m.session == nil {
		return
	}

	// Load the next page
	newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
	if newRows > 0 {
		// Write uncaptured rows to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() {
			uncapturedRows := m.slidingWindow.GetUncapturedRows()
			if len(uncapturedRows) > 0 {
				_ = metaHandler.AppendCaptureRows(uncapturedRows)
				m.slidingWindow.MarkRowsAsCaptured(len(uncapturedRows))
			}
		}

		// Update the table data and refresh the view
		allData := append([][]string{m.slidingWindow.Headers}, m.slidingWindow.Rows...)
		// Clear cache to force rebuild
		m.cachedTableLines = nil
		// NOTE: Don't update m.lastTableData - formatTableForViewport will handle it

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

// handleAltScrollUp handles Alt+Up key for scrolling viewports up
func (m *MainModel) handleAltScrollUp() (*MainModel, tea.Cmd) {
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
				// Note: If we didn't find a boundary, keep the newOffset as YOffset - 1
				// Don't jump to top (0) as that's too aggressive
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

// handleAltScrollDown handles Alt+Down key for scrolling viewports down
func (m *MainModel) handleAltScrollDown() (*MainModel, tea.Cmd) {
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
					// Load more data using the helper function
					m.loadMoreTableDataHelper()
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
