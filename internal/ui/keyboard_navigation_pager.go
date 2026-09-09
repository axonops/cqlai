package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
)

// handleSingleLineDown scrolls down by one line (j key)
func (m *MainModel) handleSingleLineDown() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceViewport.YOffset() < m.traceViewport.TotalLineCount()-m.traceViewport.Height() {
			m.traceViewport.SetYOffset(m.traceViewport.YOffset() + 1)
		}
	case m.viewMode == "table" && m.hasTable:
		maxOffset := m.tableViewport.TotalLineCount() - m.tableViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.tableViewport.YOffset() < maxOffset {
			newOffset := m.tableViewport.YOffset() + 1

			// Respect row boundaries for multi-line cells
			if len(m.tableRowBoundaries) > 0 {
				newOffset = snapLineDown(m.tableViewport.YOffset(), m.tableViewport.Height(), m.tableRowBoundaries)
			}

			if newOffset > maxOffset {
				newOffset = maxOffset
			}
			m.tableViewport.SetYOffset(newOffset)

			// Check if we need to load more data (same as Alt+Down)
			if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
				remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset() - m.tableViewport.Height()
				if remainingRows < 10 {
					m.loadMoreTableDataHelper()
				}
			}
		}
	default:
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.historyViewport.YOffset() < maxOffset {
			m.historyViewport.SetYOffset(m.historyViewport.YOffset() + 1)
		}
	}
	return m, nil
}

// handleSingleLineUp scrolls up by one line (k key)
func (m *MainModel) handleSingleLineUp() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceViewport.YOffset() > 0 {
			m.traceViewport.SetYOffset(m.traceViewport.YOffset() - 1)
		}
	case m.viewMode == "table" && m.hasTable:
		if m.tableViewport.YOffset() > 0 {
			newOffset := m.tableViewport.YOffset() - 1

			// Respect row boundaries for multi-line cells
			if len(m.tableRowBoundaries) > 0 {
				// Find current row
				currentRowIdx := -1
				for i, boundary := range m.tableRowBoundaries {
					if boundary >= m.tableViewport.YOffset() {
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

			m.tableViewport.SetYOffset(newOffset)
		}
	default:
		if m.historyViewport.YOffset() > 0 {
			m.historyViewport.SetYOffset(m.historyViewport.YOffset() - 1)
		}
	}
	return m, nil
}

// snapDownToBoundary aligns a downward scroll to the start of a record.
//
// It prefers the last boundary at or before the target, so a page lands on a
// record start rather than mid-record. Where a record is taller than the scroll
// amount - expand format, or a table with tall multi-line cells - there is no
// such boundary past the current position, and returning the current offset
// would wedge scrolling entirely. Step to the next record in that case.
func snapDownToBoundary(current, target, viewportHeight int, boundaries []int) int {
	if target <= current {
		return current
	}

	// Prefer the last record start at or before the target, so a page lands at
	// the top of a record rather than part way through one.
	snapped := current
	for _, boundary := range boundaries {
		if boundary <= target && boundary > snapped {
			snapped = boundary
		}
	}
	if snapped > current {
		return snapped
	}

	// Nothing between here and the target. Whether to jump to the next record
	// depends on whether this one fits on screen.
	if next, ok := nextBoundary(current, boundaries); ok && next-current <= viewportHeight {
		// The whole record is visible from here, so skipping to the next one
		// leaves nothing unread.
		return next
	}

	// Either the record is taller than the screen, or there is no next record.
	// Scroll within it: jumping to the next record start here would step over
	// everything that did not fit, which is unreachable any other way.
	return target
}

// nextBoundary returns the first record starting after a line.
func nextBoundary(after int, boundaries []int) (int, bool) {
	for _, boundary := range boundaries {
		if boundary > after {
			return boundary, true
		}
	}
	return 0, false
}

// snapLineDown advances a single-line scroll, honouring record starts.
//
// Aligning to the next record start is only right when the current record fits
// on screen. A record taller than the viewport has to be scrolled through a
// line at a time, or the part that did not fit can never be seen.
func snapLineDown(current, viewportHeight int, boundaries []int) int {
	if next, ok := nextBoundary(current, boundaries); ok && next-current <= viewportHeight {
		return next
	}
	return current + 1
}

// snapUpToBoundary aligns an upward scroll to the start of a record.
//
// The mirror of snapDownToBoundary: align to a record start when that record
// fits on screen, and scroll plainly when it does not, so a record taller than
// the viewport can be read on the way back up as well as down.
func snapUpToBoundary(current, target, viewportHeight int, boundaries []int) int {
	if target >= current {
		return current
	}

	best := -1
	for _, boundary := range boundaries {
		if boundary <= target {
			best = boundary
		} else {
			break
		}
	}

	if best >= 0 && current-best <= viewportHeight {
		return best
	}
	return target
}

// handleHalfPageDown scrolls down by half a page (d key)
func (m *MainModel) handleHalfPageDown() (*MainModel, tea.Cmd) {
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = m.traceViewport.Height() / 2
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
		scrollAmount = m.tableViewport.Height() / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		maxOffset := m.tableViewport.TotalLineCount() - m.tableViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		newOffset := m.tableViewport.YOffset() + scrollAmount
		if newOffset > maxOffset {
			newOffset = maxOffset
		}

		// Snap to row boundary if we have multi-line cells
		if len(m.tableRowBoundaries) > 0 {
			newOffset = snapDownToBoundary(m.tableViewport.YOffset(), newOffset, m.tableViewport.Height(), m.tableRowBoundaries)
			if newOffset > maxOffset {
				newOffset = maxOffset
			}
		}

		m.tableViewport.SetYOffset(newOffset)

		// Check if we need to load more data
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset() - m.tableViewport.Height()
			if remainingRows < 10 {
				m.loadMoreTableDataHelper()
			}
		}

	default:
		scrollAmount = m.historyViewport.Height() / 2
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

// handleHalfPageUp scrolls up by half a page (u key)
func (m *MainModel) handleHalfPageUp() (*MainModel, tea.Cmd) {
	scrollAmount := 0
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		scrollAmount = m.traceViewport.Height() / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.traceViewport.YOffset() - scrollAmount
		if newOffset < 0 {
			newOffset = 0
		}
		m.traceViewport.SetYOffset(newOffset)

	case m.viewMode == "table" && m.hasTable:
		scrollAmount = m.tableViewport.Height() / 2
		if scrollAmount < 1 {
			scrollAmount = 1
		}
		newOffset := m.tableViewport.YOffset() - scrollAmount
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

		m.tableViewport.SetYOffset(newOffset)

	default:
		scrollAmount = m.historyViewport.Height() / 2
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

// handleGoToTop jumps to the top (g key)
func (m *MainModel) handleGoToTop() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		m.traceViewport.SetYOffset(0)
	case m.viewMode == "table" && m.hasTable:
		m.tableViewport.SetYOffset(0)
	default:
		m.historyViewport.SetYOffset(0)
	}
	return m, nil
}

// handleGoToBottom jumps to the bottom (G key)
func (m *MainModel) handleGoToBottom() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.traceViewport.SetYOffset(maxOffset)

	case m.viewMode == "table" && m.hasTable:
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height()
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

		m.tableViewport.SetYOffset(maxOffset)

		// Load all remaining data if we have more
		if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			// Could implement loading all data here if desired
			logger.DebugfToFile("Nav", "Jump to bottom requested but more data available")
		}

	default:
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.historyViewport.SetYOffset(maxOffset)
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
	default:
		m.scrollHorizontally(-horizontalStep)
	}
	return m, nil
}

// handlePageLeftScroll scrolls left by page width (Alt+PageUp)
func (m *MainModel) handlePageLeftScroll() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceHorizontalOffset > 0 {
			scrollAmount := m.traceViewport.Width() / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			m.traceHorizontalOffset -= scrollAmount
			if m.traceHorizontalOffset < 0 {
				m.traceHorizontalOffset = 0
			}
			m.refreshTraceView()
		}
	default:
		m.scrollHorizontally(-m.halfScreen())
	}
	return m, nil
}

// handlePageRightScroll scrolls right by page width (Alt+PageDown)
func (m *MainModel) handlePageRightScroll() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceTableWidth > m.traceViewport.Width() {
			scrollAmount := m.traceViewport.Width() / 2
			if scrollAmount < 10 {
				scrollAmount = 10
			}
			maxOffset := m.traceTableWidth - m.traceViewport.Width() + 10
			m.traceHorizontalOffset += scrollAmount
			if m.traceHorizontalOffset > maxOffset {
				m.traceHorizontalOffset = maxOffset
			}
			m.refreshTraceView()
		}
	default:
		m.scrollHorizontally(m.halfScreen())
	}
	return m, nil
}

// handleHorizontalScrollRight scrolls right (> or . key)
func (m *MainModel) handleHorizontalScrollRight() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		if m.traceTableWidth > m.traceViewport.Width() {
			maxOffset := m.traceTableWidth - m.traceViewport.Width() + 10
			if m.traceHorizontalOffset < maxOffset {
				m.traceHorizontalOffset += 10
				if m.traceHorizontalOffset > maxOffset {
					m.traceHorizontalOffset = maxOffset
				}
				m.refreshTraceView()
			}
		}
	default:
		m.scrollHorizontally(horizontalStep)
	}
	return m, nil
}

// horizontalStep is how far one press of h, l or an arrow scrolls sideways.
const horizontalStep = 10

// halfScreen is how far < and > scroll: half the width on screen, so a page
// sideways leaves half of what you were reading in view.
func (m *MainModel) halfScreen() int {
	return max(m.tableViewport.Width()/2, horizontalStep)
}

// scrollHorizontally moves the Results view sideways by n columns.
//
// Every route sideways comes through here - h and l, the arrows, < and >, and
// the wheel. They used to be six copies of the same clamp-then-rebuild, in
// three files, and #115 fixed two of them: the other four went on rebuilding
// ASCII output as a boxed table the moment you scrolled.
func (m *MainModel) scrollHorizontally(n int) {
	if m.viewMode != "table" || !m.hasTable {
		return
	}

	maxOffset := max(m.tableWidth-m.tableViewport.Width()+horizontalStep, 0)
	offset := min(max(m.horizontalOffset+n, 0), maxOffset)
	if offset == m.horizontalOffset {
		return
	}

	m.horizontalOffset = offset
	m.applyHorizontalOffset()
}

// resetHorizontalScroll puts a new result back at the left-hand edge.
//
// Both halves matter: the offset the table renderer reads, and the viewport's
// own, which is what slides text sideways. Resetting only the first left ASCII
// output still scrolled off the left when the next query landed.
func (m *MainModel) resetHorizontalScroll() {
	m.horizontalOffset = 0
	m.tableViewport.SetXOffset(0)
}

// applyHorizontalOffset moves the Results view sideways.
//
// A boxed table is rebuilt, because it drops whole columns rather than cutting
// one down the middle. Everything else is text - ASCII art, JSON lines - and is
// slid under the viewport instead. Rebuilding those as a table is what replaced
// what you were reading with a mangled one on the first sideways scroll.
func (m *MainModel) applyHorizontalOffset() {
	if m.resultFormat == config.OutputFormatTable {
		if m.lastTableData != nil {
			m.refreshTableView()
		}
		return
	}
	m.tableViewport.SetXOffset(m.horizontalOffset)
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

		m.refreshTableContent(allData)

		// Update row count
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)
	}
}

// handleAltScrollUp handles Alt+Up key for scrolling viewports up
func (m *MainModel) handleAltScrollUp() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace up by one line
		if m.traceViewport.YOffset() > 0 {
			m.traceViewport.SetYOffset(m.traceViewport.YOffset() - 1)
		}
	case m.viewMode == "table" && m.hasTable:
		// Scroll table up to previous row boundary
		if m.tableViewport.YOffset() > 0 {
			newOffset := m.tableViewport.YOffset() - 1

			// Find the previous row boundary
			if len(m.tableRowBoundaries) > 0 {
				for i := len(m.tableRowBoundaries) - 1; i >= 0; i-- {
					if m.tableRowBoundaries[i] < m.tableViewport.YOffset() {
						newOffset = m.tableRowBoundaries[i]
						break
					}
				}
				// Note: If we didn't find a boundary, keep the newOffset as YOffset - 1
				// Don't jump to top (0) as that's too aggressive
			}

			m.tableViewport.SetYOffset(newOffset)
		}
	default:
		// Scroll history up by one line
		if m.historyViewport.YOffset() > 0 {
			m.historyViewport.SetYOffset(m.historyViewport.YOffset() - 1)
		}
	}
	return m, nil
}

// handleAltScrollDown handles Alt+Down key for scrolling viewports down
func (m *MainModel) handleAltScrollDown() (*MainModel, tea.Cmd) {
	switch {
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace down by one line
		maxOffset := m.traceViewport.TotalLineCount() - m.traceViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.traceViewport.YOffset() < maxOffset {
			m.traceViewport.SetYOffset(m.traceViewport.YOffset() + 1)
		}
	case m.viewMode == "table" && m.hasTable:
		// Scroll table down to next row boundary
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height()
		maxOffset := totalLines - viewportHeight
		if maxOffset < 0 {
			maxOffset = 0
		}

		// Check if we're at the end with no more data
		noMoreData := m.slidingWindow == nil || !m.slidingWindow.hasMoreData

		if m.tableViewport.YOffset() < maxOffset {
			newOffset := m.tableViewport.YOffset() + 1

			// Find the next row boundary
			if len(m.tableRowBoundaries) > 0 {
				newOffset = snapLineDown(m.tableViewport.YOffset(), m.tableViewport.Height(), m.tableRowBoundaries)
			}

			// Special handling for the last boundary (bottom border)
			if noMoreData && len(m.tableRowBoundaries) > 0 {
				lastBoundary := m.tableRowBoundaries[len(m.tableRowBoundaries)-1]
				if newOffset >= lastBoundary-viewportHeight {
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

			m.tableViewport.SetYOffset(newOffset)

			// Check if we need to load more data
			if m.slidingWindow != nil && m.slidingWindow.hasMoreData {
				// If we're within 10 rows of the bottom, load more
				remainingRows := m.tableViewport.TotalLineCount() - m.tableViewport.YOffset() - m.tableViewport.Height()
				if remainingRows < 10 {
					// Load more data using the helper function
					m.loadMoreTableDataHelper()
				}
			}
		}
	default:
		// Scroll history down by one line
		maxOffset := m.historyViewport.TotalLineCount() - m.historyViewport.Height()
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.historyViewport.YOffset() < maxOffset {
			m.historyViewport.SetYOffset(m.historyViewport.YOffset() + 1)
		}
	}
	return m, nil
}
