package ui

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
)

// handleMouseInput handles mouse events.
//
// cqlai owns the mouse while reporting is on, which is what makes the tabs and
// the settings on the bottom line clickable. It also means the terminal will
// not select text for us any more, so we do that ourselves - see selection.go.
// MOUSE OFF hands the mouse back and the terminal behaves as it always did.
func (m *MainModel) handleMouseInput(msg tea.MouseMsg) (*MainModel, tea.Cmd) {
	mouse := msg.Mouse()

	switch msg.(type) {
	case tea.MouseClickMsg:
		logger.DebugfToFile("Mouse", "Press Button=%v X=%d Y=%d Mod=%v",
			mouse.Button, mouse.X, mouse.Y, mouse.Mod)
		// Only the left button does anything. A right click is left alone so
		// the terminal's own context menu, and whatever it binds paste to,
		// still work.
		if mouse.Button != tea.MouseLeft {
			return m, nil
		}
		return m.handleMousePress(mouse)

	case tea.MouseMotionMsg:
		// Motion only arrives with a button held, and only matters mid-drag;
		// extendSelection returns straight away otherwise.
		return m.extendSelection(mouse.X, mouse.Y)

	case tea.MouseReleaseMsg:
		return m.endSelection()

	case tea.MouseWheelMsg:
		// The help window is over everything, so it has the wheel first.
		if m.help.active {
			switch mouse.Button {
			case tea.MouseWheelUp:
				return m.scrollHelp(-wheelLines)
			case tea.MouseWheelDown:
				return m.scrollHelp(wheelLines)
			}
			return m, nil
		}

		// The command list has the wheel while it is open, for the same reason
		// the settings lists do: otherwise it goes to the view behind and a
		// list longer than the box cannot be got through.
		if m.historySearchMode {
			switch mouse.Button {
			case tea.MouseWheelUp:
				return m.moveHistorySelection(-wheelLines)
			case tea.MouseWheelDown:
				return m.moveHistorySelection(wheelLines)
			}
			return m, nil
		}

		// While the list is open the wheel belongs to it rather than to the
		// view behind it, or the choices past the first screenful cannot be
		// reached at all.
		if m.chooser.active {
			switch mouse.Button {
			case tea.MouseWheelUp:
				return m.scrollChooser(-chooserScrollStep)
			case tea.MouseWheelDown:
				return m.scrollChooser(chooserScrollStep)
			}
			return m, nil
		}
		return m.handleMouseWheel(mouse)
	}

	return m, nil
}

// handleMousePress routes a left button press: the controls first, then the
// viewport, where a press starts a selection.
func (m *MainModel) handleMousePress(mouse tea.Mouse) (*MainModel, tea.Cmd) {
	// A press anywhere drops whatever was selected.
	m.clearSelection()

	// With the help open, a press dismisses it rather than acting on whatever
	// is underneath - which you cannot see to aim at.
	if m.help.active {
		if mode, _ := m.modeAt(m.windowWidth, mouse.X); mouse.Y == 0 && mode == helpMode {
			return m.toggleHelp()
		}
		m.help = helpWindow{}
		return m, nil
	}

	// A press inside the open list picks that value.
	if m.chooser.active {
		if choice, inside := m.choiceAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); inside {
			return m.applySettingChoice(choice)
		}
		// Anywhere else dismisses it, except on the tabs and the two bars at
		// the bottom, which still act on the click. The status line is left to
		// close the list itself, because it has to see which field the list
		// belongs to before it can tell a close from a reopen.
		if mouse.Y != 0 && mouse.Y < m.windowHeight-2 {
			m.closeSettingChooser()
			return m, nil
		}
	}

	// A press on a command in either history list puts it in the prompt. This
	// comes before the rows below, because the lists are drawn over them.
	if updated, cmd, hit := m.clickHistoryCommand(mouse.X, mouse.Y); hit {
		return updated, cmd
	}

	// Row 0 is the tabs, the last row is the connection bar and the one above
	// it carries the query facts; everything else is content, and a press there
	// starts a selection.
	switch mouse.Y {
	case 0:
		m.closeSettingChooser()
		return m.clickTab(mouse.X)
	case m.windowHeight - 2:
		return m.clickInfoField(mouse.X)
	case m.windowHeight - 1:
		return m.clickStatusSetting(mouse.X)
	}

	return m.beginSelection(mouse.X, mouse.Y)
}

// handleMouseWheel scrolls the view under the pointer.
func (m *MainModel) handleMouseWheel(mouse tea.Mouse) (*MainModel, tea.Cmd) {
	// Any modifier plus a vertical wheel means horizontal scrolling.
	modified := mouse.Mod.Contains(tea.ModShift) ||
		mouse.Mod.Contains(tea.ModAlt) ||
		mouse.Mod.Contains(tea.ModCtrl)

	switch mouse.Button {
	case tea.MouseWheelUp:
		if modified && m.viewMode == "table" && m.hasTable {
			return m.handleMouseWheelLeft()
		}
		return m.handleMouseWheelUp()
	case tea.MouseWheelDown:
		if modified && m.viewMode == "table" && m.hasTable {
			return m.handleMouseWheelRight()
		}
		return m.handleMouseWheelDown()
	case tea.MouseWheelLeft:
		// Native horizontal scroll, for mice and trackpads that report it.
		return m.handleMouseWheelLeft()
	case tea.MouseWheelRight:
		return m.handleMouseWheelRight()
	default:
		return m, nil
	}
}

// handleMouseWheelUp handles mouse wheel up scrolling
func (m *MainModel) handleMouseWheelUp() (*MainModel, tea.Cmd) {
	scrollAmount := wheelLines

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
	scrollAmount := wheelLines

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
	m.scrollHorizontally(-horizontalStep)
	return m, nil
}

// handleMouseWheelRight handles horizontal scrolling right
func (m *MainModel) handleMouseWheelRight() (*MainModel, tea.Cmd) {
	m.scrollHorizontally(horizontalStep)
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

// clickTab switches to whichever mode's tab covers a column.
//
// A tab with nothing to show is dimmed and reports itself unavailable, so
// clicking it does nothing rather than dropping you on an empty screen. The
// F-keys stay unconditional, matching what they did before there were tabs.
func (m *MainModel) clickTab(col int) (*MainModel, tea.Cmd) {
	mode, available := m.modeAt(m.windowWidth, col)
	if !available || mode == m.viewMode {
		return m, nil
	}

	logger.DebugfToFile("Mouse", "Tab clicked at column %d: switching to %s", col, mode)

	if mode == helpMode {
		return m.toggleHelp()
	}

	switch mode {
	case "history":
		return m.handleF2()
	case "table":
		return m.handleF3()
	case "trace":
		return m.handleF4()
	case "ai":
		return m.handleF5()
	}
	return m, nil
}

// clickStatusSetting opens the list of values for a setting on the status line,
// or closes the one already open.
func (m *MainModel) clickStatusSetting(col int) (*MainModel, tea.Cmd) {
	wasOpen, openSetting := m.chooser.active, m.chooser.setting
	m.closeSettingChooser()

	setting, anchorX, ok := m.statusBar.settingAt(col)
	if !ok {
		return m, nil
	}

	// Clicking the field whose list is already open closes it. Without this the
	// press closed the list and reopened it in the same breath, so the list
	// never appeared to go away and the only way to dismiss it was to click
	// somewhere else entirely.
	if wasOpen && openSetting == setting {
		return m, nil
	}

	// Connection is not a setting: it opens a panel of facts about how we are
	// connected, with nothing to pick.
	if setting == settingConnection {
		m.openConnectionPanel(anchorX)
		return m, nil
	}

	current := m.currentSettingValue(setting)

	// Every setting but KS has a fixed set of values. Keyspaces come from the
	// cluster, so they are fetched here rather than listed in settingChoices.
	var choices []string
	if setting == settingKeyspace {
		choices = m.keyspaceChoices()
	} else {
		choices, _ = settingChoices(setting, current)
	}
	if len(choices) == 0 {
		return m, nil
	}

	logger.DebugfToFile("Mouse", "Status setting %q clicked at column %d, current %q", setting, col, current)
	m.openSettingChooser(setting, anchorX, choices, current)
	return m, nil
}

// clickInfoField acts on a press on the query info bar.
func (m *MainModel) clickInfoField(col int) (*MainModel, tea.Cmd) {
	field, ok := m.topBar.fieldAt(col)
	if !ok || field != infoHistory {
		return m, nil
	}

	logger.DebugfToFile("Mouse", "History clicked at column %d", col)
	return m.openHistoryList()
}

// keyspaceChoices lists the keyspaces to offer for KS.
//
// Asked for on each click rather than cached, because keyspaces are created and
// dropped while cqlai is running and a list that quietly goes stale is worse
// than one small query against a system table. System keyspaces are included:
// USE system_schema is a reasonable thing to want.
func (m *MainModel) keyspaceChoices() []string {
	if m.session == nil {
		return nil
	}

	keyspaces, err := m.session.DescribeKeyspacesQuery()
	if err != nil {
		logger.DebugfToFile("Mouse", "Listing keyspaces for the KS chooser: %v", err)
		return nil
	}

	names := make([]string, 0, len(keyspaces))
	for _, ks := range keyspaces {
		names = append(names, ks.Name)
	}
	return names
}

// currentSettingValue reads a setting as it is displayed, so the chooser can
// mark it and open on it.
func (m *MainModel) currentSettingValue(setting string) string {
	for _, seg := range m.statusBar.segments() {
		if seg.setting == setting {
			return seg.value
		}
	}
	return ""
}
