package ui

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
)

// handleMouseInput handles mouse events.
//
// Nothing reaches this today: cqlai no longer turns on mouse reporting, because
// doing so takes the buttons away from the terminal and breaks its text
// selection and right-click paste. The wheel arrives as Up/Down key presses via
// alternate scroll mode instead - see mouse_mode.go.
//
// It is kept for the day something clickable needs mouse reporting switched on
// while it is on screen, which is the point at which wheel events start
// arriving here again.
func (m *MainModel) handleMouseInput(msg tea.MouseMsg) (*MainModel, tea.Cmd) {
	mouse := msg.Mouse()

	logger.DebugfToFile("Mouse", "MouseEvent: %T Button=%v X=%d Y=%d Mod=%v",
		msg, mouse.Button, mouse.X, mouse.Y, mouse.Mod)

	// Only wheel events matter here. v2 splits clicks, releases, motion and
	// wheel into separate types, so anything else is simply not ours. Note that
	// ignoring a click here is not what lets the terminal select text - by the
	// time an event arrives the terminal has already given the button up.
	// Leaving MouseMode off is what does that.
	if _, isWheel := msg.(tea.MouseWheelMsg); !isWheel {
		return m, nil
	}

	// Any modifier plus a vertical wheel means horizontal scrolling.
	modified := mouse.Mod.Contains(tea.ModShift) ||
		mouse.Mod.Contains(tea.ModAlt) ||
		mouse.Mod.Contains(tea.ModCtrl)

	switch mouse.Button {
	case tea.MouseWheelUp:
		// Check for horizontal scrolling modes
		if modified && m.viewMode == "table" && m.hasTable {
			// Any modifier + WheelUp = Scroll left
			logger.DebugfToFile("Mouse", "Modified WheelUp detected (Mod=%v) - scrolling left",
				mouse.Mod)
			return m.handleMouseWheelLeft()
		}
		// Regular scroll up
		return m.handleMouseWheelUp()
	case tea.MouseWheelDown:
		// Check for horizontal scrolling modes
		if modified && m.viewMode == "table" && m.hasTable {
			// Any modifier + WheelDown = Scroll right
			logger.DebugfToFile("Mouse", "Modified WheelDown detected (Mod=%v) - scrolling right",
				mouse.Mod)
			return m.handleMouseWheelRight()
		}
		// Regular scroll down
		return m.handleMouseWheelDown()
	case tea.MouseWheelLeft:
		// Native horizontal scroll left (for mice/trackpads that support it)
		return m.handleMouseWheelLeft()
	case tea.MouseWheelRight:
		// Native horizontal scroll right (for mice/trackpads that support it)
		return m.handleMouseWheelRight()
	default:
		// Ignore left, middle and right clicks. Note that ignoring them here is
		// not what lets the terminal select text - by the time an event reaches
		// this function the terminal has already given the button up. Leaving
		// mouse reporting off is what does that.
		return m, nil
	}
}

// handleMouseWheelUp handles mouse wheel up scrolling
func (m *MainModel) handleMouseWheelUp() (*MainModel, tea.Cmd) {
	scrollAmount := 3 // Lines to scroll per wheel event

	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		// Scroll AI conversation up
		m.aiConversationViewport.SetYOffset(max(0, m.aiConversationViewport.YOffset()-scrollAmount))
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace up
		m.traceViewport.SetYOffset(max(0, m.traceViewport.YOffset()-scrollAmount))
	case m.viewMode == "table" && m.hasTable:
		// Scroll table up
		m.tableViewport.SetYOffset(max(0, m.tableViewport.YOffset()-scrollAmount))
	default:
		// Scroll history up
		m.historyViewport.SetYOffset(max(0, m.historyViewport.YOffset()-scrollAmount))
	}
	return m, nil
}

// handleMouseWheelDown handles mouse wheel down scrolling
func (m *MainModel) handleMouseWheelDown() (*MainModel, tea.Cmd) {
	scrollAmount := 3 // Lines to scroll per wheel event

	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		// Scroll AI conversation down
		maxOffset := max(0, m.aiConversationViewport.TotalLineCount()-m.aiConversationViewport.Height())
		m.aiConversationViewport.SetYOffset(min(maxOffset, m.aiConversationViewport.YOffset()+scrollAmount))
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace down
		maxOffset := max(0, m.traceViewport.TotalLineCount()-m.traceViewport.Height())
		m.traceViewport.SetYOffset(min(maxOffset, m.traceViewport.YOffset()+scrollAmount))
	case m.viewMode == "table" && m.hasTable:
		// Log initial state
		logger.DebugfToFile("Mouse", "MouseWheelDown in table: YOffset=%d, TotalLines=%d, Height=%d",
			m.tableViewport.YOffset(), m.tableViewport.TotalLineCount(), m.tableViewport.Height())

		// First, check if we need to load more data BEFORE calculating limits (like PageDown does)
		if m.slidingWindow != nil {
			logger.DebugfToFile("Mouse", "SlidingWindow state: hasMoreData=%v, currentRows=%d, totalSeen=%d",
				m.slidingWindow.hasMoreData, len(m.slidingWindow.Rows), m.slidingWindow.TotalRowsSeen)

			if m.slidingWindow.hasMoreData {
				totalLines := m.tableViewport.TotalLineCount()
				viewportHeight := m.tableViewport.Height()
				currentOffset := m.tableViewport.YOffset()

				// Check if scrolling would take us near the bottom
				potentialOffset := currentOffset + scrollAmount
				// Calculate how many lines remain below the current view
				remainingRows := totalLines - (currentOffset + viewportHeight)

				logger.DebugfToFile("Mouse", "Scroll check: currentOffset=%d, potentialOffset=%d, remainingRows=%d, threshold=20",
					currentOffset, potentialOffset, remainingRows)

				// Load more data if we're getting close to the bottom
				if remainingRows < 20 {
					logger.DebugfToFile("Mouse", "Triggering data load: remainingRows=%d < 20",
						remainingRows)
					m.loadMoreTableData()
				}
			}
		} else {
			logger.DebugfToFile("Mouse", "No sliding window available")
		}

		// NOW calculate the limits with potentially updated data
		totalLines := m.tableViewport.TotalLineCount()
		viewportHeight := m.tableViewport.Height()
		maxOffset := max(0, totalLines-viewportHeight)

		// Calculate and apply new offset
		newOffset := min(maxOffset, m.tableViewport.YOffset()+scrollAmount)
		m.tableViewport.SetYOffset(newOffset)
	default:
		// Scroll history down
		maxOffset := max(0, m.historyViewport.TotalLineCount()-m.historyViewport.Height())
		m.historyViewport.SetYOffset(min(maxOffset, m.historyViewport.YOffset()+scrollAmount))
	}
	return m, nil
}

// handleMouseWheelLeft handles horizontal scrolling left
func (m *MainModel) handleMouseWheelLeft() (*MainModel, tea.Cmd) {
	logger.DebugfToFile("Mouse", "handleMouseWheelLeft: viewMode=%s, hasTable=%v, hasData=%v",
		m.viewMode, m.hasTable, m.lastTableData != nil)

	if m.viewMode == "table" && m.hasTable && m.lastTableData != nil {
		// Scroll left by 10 columns
		oldOffset := m.horizontalOffset
		m.horizontalOffset = max(0, m.horizontalOffset-10)

		logger.DebugfToFile("Mouse", "Scrolling left: oldOffset=%d, newOffset=%d, tableWidth=%d, viewportWidth=%d",
			oldOffset, m.horizontalOffset, m.tableWidth, m.tableViewport.Width())

		// Only re-render if offset actually changed
		if oldOffset != m.horizontalOffset {
			// Refresh the table view (same as arrow keys do)
			m.refreshTableView()
			logger.DebugfToFile("Mouse", "Table refreshed with left scroll")
		}
	}
	return m, nil
}

// handleMouseWheelRight handles horizontal scrolling right
func (m *MainModel) handleMouseWheelRight() (*MainModel, tea.Cmd) {
	logger.DebugfToFile("Mouse", "handleMouseWheelRight: viewMode=%s, hasTable=%v, hasData=%v",
		m.viewMode, m.hasTable, m.lastTableData != nil)

	if m.viewMode == "table" && m.hasTable && m.lastTableData != nil {
		// Scroll right by 10 columns
		oldOffset := m.horizontalOffset

		// Calculate max offset based on actual table width
		if m.tableWidth > m.tableViewport.Width() {
			maxOffset := m.tableWidth - m.tableViewport.Width() + 10 // Add some buffer
			m.horizontalOffset = min(maxOffset, m.horizontalOffset+10)
		} else {
			// Table fits in viewport, but allow some scrolling anyway
			m.horizontalOffset += 10
		}

		logger.DebugfToFile("Mouse", "Scrolling right: oldOffset=%d, newOffset=%d, tableWidth=%d, viewportWidth=%d",
			oldOffset, m.horizontalOffset, m.tableWidth, m.tableViewport.Width())

		// Only re-render if offset actually changed
		if oldOffset != m.horizontalOffset {
			// Refresh the table view (same as arrow keys do)
			m.refreshTableView()
			logger.DebugfToFile("Mouse", "Table refreshed with right scroll")
		}
	}
	return m, nil
}

// Helper function to load more table data (extracted from keyboard handler logic)
func (m *MainModel) loadMoreTableData() {
	if m.slidingWindow == nil {
		logger.DebugfToFile("Mouse", "loadMoreTableData: No sliding window")
		return
	}
	if !m.slidingWindow.hasMoreData {
		logger.DebugfToFile("Mouse", "loadMoreTableData: No more data available")
		return
	}

	logger.DebugfToFile("Mouse", "loadMoreTableData: Loading more rows, pageSize=%d", m.session.PageSize())
	newRows := m.slidingWindow.LoadMoreRows(m.session.PageSize())
	logger.DebugfToFile("Mouse", "loadMoreTableData: Loaded %d new rows, total rows now=%d",
		newRows, len(m.slidingWindow.Rows))

	if newRows > 0 {
		// Write uncaptured rows to capture file if capturing
		metaHandler := router.GetMetaHandler()
		if metaHandler != nil && metaHandler.IsCapturing() {
			uncapturedRows := m.slidingWindow.GetUncapturedRows()
			if len(uncapturedRows) > 0 {
				_ = metaHandler.WriteCaptureResult("", m.slidingWindow.Headers, uncapturedRows)
				m.slidingWindow.MarkRowsAsCaptured(len(uncapturedRows))
			}
		}

		// Update the table data and refresh the view
		allData := append([][]string{m.slidingWindow.Headers}, m.slidingWindow.Rows...)

		// Clear cache to force rebuild (important!)
		m.cachedTableLines = nil
		m.lastTableData = allData

		m.refreshTableContent(allData)

		// Update row count
		m.topBar.RowCount = int(m.slidingWindow.TotalRowsSeen)
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)
	}
}

// Helper function to format table as JSON
func (m *MainModel) formatTableAsJSON() string {
	if m.slidingWindow == nil {
		return ""
	}

	// Check if we have a single [json] column from SELECT JSON
	if len(m.slidingWindow.Headers) == 1 && m.slidingWindow.Headers[0] == "[json]" {
		// This is already JSON from SELECT JSON - just extract it
		jsonStr := ""
		for _, row := range m.slidingWindow.Rows {
			if len(row) > 0 {
				jsonStr += row[0] + "\n"
			}
		}
		return jsonStr
	}

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
	return jsonStr
}
