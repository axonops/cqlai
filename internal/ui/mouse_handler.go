package ui

import (
	"encoding/json"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// handleMouseInput handles mouse events
func (m *MainModel) handleMouseInput(msg tea.MouseMsg) (*MainModel, tea.Cmd) {
	// Debug log ALL mouse events
	logger.DebugfToFile("Mouse", "MouseEvent: Action=%v, Button=%v, X=%d, Y=%d, Shift=%v, Alt=%v, Ctrl=%v",
		msg.Action, msg.Button, msg.X, msg.Y, msg.Shift, msg.Alt, msg.Ctrl)

	// Only handle wheel events - pass through everything else
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// Check for horizontal scrolling modes
		if (msg.Shift || msg.Alt || msg.Ctrl) && m.viewMode == "table" && m.hasTable {
			// Any modifier + WheelUp = Scroll left
			logger.DebugfToFile("Mouse", "Modified WheelUp detected (Shift=%v, Alt=%v, Ctrl=%v) - scrolling left",
				msg.Shift, msg.Alt, msg.Ctrl)
			return m.handleMouseWheelLeft()
		}
		// Regular scroll up
		return m.handleMouseWheelUp()
	case tea.MouseButtonWheelDown:
		// Check for horizontal scrolling modes
		if (msg.Shift || msg.Alt || msg.Ctrl) && m.viewMode == "table" && m.hasTable {
			// Any modifier + WheelDown = Scroll right
			logger.DebugfToFile("Mouse", "Modified WheelDown detected (Shift=%v, Alt=%v, Ctrl=%v) - scrolling right",
				msg.Shift, msg.Alt, msg.Ctrl)
			return m.handleMouseWheelRight()
		}
		// Regular scroll down
		return m.handleMouseWheelDown()
	case tea.MouseButtonWheelLeft:
		// Native horizontal scroll left (for mice/trackpads that support it)
		return m.handleMouseWheelLeft()
	case tea.MouseButtonWheelRight:
		// Native horizontal scroll right (for mice/trackpads that support it)
		return m.handleMouseWheelRight()
	default:
		// Ignore all other button events (left, middle, right clicks)
		// This allows terminal to handle text selection
		return m, nil
	}
}

// handleMouseWheelUp handles mouse wheel up scrolling
func (m *MainModel) handleMouseWheelUp() (*MainModel, tea.Cmd) {
	scrollAmount := 3 // Lines to scroll per wheel event

	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		// Scroll AI conversation up
		m.aiConversationViewport.YOffset = max(0, m.aiConversationViewport.YOffset-scrollAmount)
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace up
		m.traceViewport.YOffset = max(0, m.traceViewport.YOffset-scrollAmount)
	case m.viewMode == "table" && m.hasTable:
		// Scroll table up
		m.tableViewport.YOffset = max(0, m.tableViewport.YOffset-scrollAmount)
	default:
		// Scroll history up
		m.historyViewport.YOffset = max(0, m.historyViewport.YOffset-scrollAmount)
	}
	return m, nil
}

// handleMouseWheelDown handles mouse wheel down scrolling
func (m *MainModel) handleMouseWheelDown() (*MainModel, tea.Cmd) {
	scrollAmount := 3 // Lines to scroll per wheel event

	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		// Scroll AI conversation down
		maxOffset := max(0, m.aiConversationViewport.TotalLineCount()-m.aiConversationViewport.Height)
		m.aiConversationViewport.YOffset = min(maxOffset, m.aiConversationViewport.YOffset+scrollAmount)
	case m.viewMode == "trace" && m.hasTrace:
		// Scroll trace down
		maxOffset := max(0, m.traceViewport.TotalLineCount()-m.traceViewport.Height)
		m.traceViewport.YOffset = min(maxOffset, m.traceViewport.YOffset+scrollAmount)
	case m.viewMode == "table" && m.hasTable:
		// Log initial state
		logger.DebugfToFile("Mouse", "MouseWheelDown in table: YOffset=%d, TotalLines=%d, Height=%d",
			m.tableViewport.YOffset, m.tableViewport.TotalLineCount(), m.tableViewport.Height)

		// First, check if we need to load more data BEFORE calculating limits (like PageDown does)
		if m.slidingWindow != nil {
			logger.DebugfToFile("Mouse", "SlidingWindow state: hasMoreData=%v, currentRows=%d, totalSeen=%d",
				m.slidingWindow.hasMoreData, len(m.slidingWindow.Rows), m.slidingWindow.TotalRowsSeen)

			if m.slidingWindow.hasMoreData {
				totalLines := m.tableViewport.TotalLineCount()
				viewportHeight := m.tableViewport.Height
				currentOffset := m.tableViewport.YOffset

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
		viewportHeight := m.tableViewport.Height
		maxOffset := max(0, totalLines-viewportHeight)

		// Calculate and apply new offset
		newOffset := min(maxOffset, m.tableViewport.YOffset+scrollAmount)
		m.tableViewport.YOffset = newOffset
	default:
		// Scroll history down
		maxOffset := max(0, m.historyViewport.TotalLineCount()-m.historyViewport.Height)
		m.historyViewport.YOffset = min(maxOffset, m.historyViewport.YOffset+scrollAmount)
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
			oldOffset, m.horizontalOffset, m.tableWidth, m.tableViewport.Width)

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
		if m.tableWidth > m.tableViewport.Width {
			maxOffset := m.tableWidth - m.tableViewport.Width + 10 // Add some buffer
			m.horizontalOffset = min(maxOffset, m.horizontalOffset+10)
		} else {
			// Table fits in viewport, but allow some scrolling anyway
			m.horizontalOffset = m.horizontalOffset + 10
		}

		logger.DebugfToFile("Mouse", "Scrolling right: oldOffset=%d, newOffset=%d, tableWidth=%d, viewportWidth=%d",
			oldOffset, m.horizontalOffset, m.tableWidth, m.tableViewport.Width)

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

		// Format based on current output format
		var contentStr string
		if m.sessionManager != nil {
			switch m.sessionManager.GetOutputFormat() {
			case config.OutputFormatASCII:
				contentStr = FormatASCIITable(allData)
			case config.OutputFormatExpand:
				contentStr = FormatExpandTable(allData, m.styles)
			case config.OutputFormatJSON:
				contentStr = m.formatTableAsJSON()
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

