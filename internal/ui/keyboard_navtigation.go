package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
)

// handlePageUp handles PageUp key press
func (m *MainModel) handlePageUp(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
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
		scrollAmount = int(float64(m.traceViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.traceViewport.YOffset() - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.traceViewport.SetYOffset(newOffset)
	case m.viewMode == "table" && m.hasTable:
		scrollAmount = int(float64(m.tableViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.tableViewport.YOffset() - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}

		// Align to a record start, unless the record is taller than the screen,
		// in which case scroll within it so every line can be read.
		if len(m.tableRowBoundaries) > 0 && newOffset > 0 {
			newOffset = snapUpToBoundary(m.tableViewport.YOffset(), newOffset, m.tableViewport.Height(), m.tableRowBoundaries)
		}

		m.tableViewport.SetYOffset(newOffset)
	default:
		scrollAmount = int(float64(m.historyViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.historyViewport.YOffset() - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.historyViewport.SetYOffset(newOffset)
	}
	return m, nil
}

// handlePageDown handles PageDown key press
func (m *MainModel) handlePageDown(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
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
		scrollAmount = int(float64(m.traceViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.traceViewport.YOffset() + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}
		m.traceViewport.SetYOffset(newOffset)
	case m.viewMode == "table" && m.hasTable:
		scrollAmount = int(float64(m.tableViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}

		// First, check if we need to load more data BEFORE calculating limits
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			totalLines := m.tableViewport.TotalLineCount()
			viewportHeight := m.tableViewport.Height()
			currentOffset := m.tableViewport.YOffset()

			// Check if scrolling would take us near the bottom
			potentialOffset := currentOffset + scrollAmount
			remainingRows := totalLines - potentialOffset - viewportHeight

			// Load more data if we're getting close to the bottom
			if remainingRows < 20 {
				newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
				logger.DebugfToFile("PageDown", "Pre-loading more rows: requested=%d, got=%d, total=%d",
					m.session.PageSize(), newRows, len(m.slidingWindow.Rows))

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

					m.refreshTableContent(allData)

					// Update row count
					m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
					m.rowCount = int(m.slidingWindow.TotalRowsSeen)
				}
			}
		}

		// NOW calculate the limits with potentially updated data
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height()
		maxOffset := totalLines - viewportHeight
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.tableViewport.YOffset() + scrollAmount
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
				// Align to a record start, but only when the record fits on
				// screen. A record taller than the viewport has to be scrolled
				// through, and snapping would step over the part that did not
				// fit. Sharing the helper keeps this in step with the d key,
				// which had the same fault.
				newOffset = snapDownToBoundary(m.tableViewport.YOffset(), newOffset, viewportHeight, m.tableRowBoundaries)
			}
		}

		m.tableViewport.SetYOffset(newOffset)

		// Data loading is now done BEFORE calculating limits, so no need to load here
	default:
		scrollAmount = int(float64(m.historyViewport.Height()) * 0.8)
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.historyViewport.YOffset() + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}
		m.historyViewport.SetYOffset(newOffset)
	}
	return m, nil
}

// handleLeftArrow handles Left arrow key press
func (m *MainModel) handleLeftArrow(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.PrevChoice()
		return m, nil
	}
	// If Alt is held, scroll table/trace left
	if msg.Mod.Contains(tea.ModAlt) {
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
func (m *MainModel) handleRightArrow(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {

	// If modal is showing, navigate choices
	if m.modal.Type != ModalNone {
		m.modal.NextChoice()
		return m, nil
	}
	// If Alt is held, scroll table/trace right
	if msg.Mod.Contains(tea.ModAlt) {
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			// Scroll trace table right
			if m.traceTableWidth > m.traceViewport.Width() {
				maxOffset := m.traceTableWidth - m.traceViewport.Width() + 10 // Add some buffer
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
			if m.tableWidth > m.tableViewport.Width() {
				maxOffset := m.tableWidth - m.tableViewport.Width() + 10 // Add some buffer
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
