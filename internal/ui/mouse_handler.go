package ui

import (
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
		// The right button pastes. Taking the mouse takes the terminal's own
		// context menu away, and with it the paste everyone reaches for; most
		// terminals still hand a shifted click back to themselves for anyone
		// who wants the menu.
		if mouse.Button == tea.MouseRight {
			return m.requestPaste()
		}
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
		// The candidate list has the wheel while it is showing.
		if m.capture.active && len(m.capture.matches) > 0 {
			switch mouse.Button {
			case tea.MouseWheelUp:
				m.showMatch(m.capture.match - wheelLines)
			case tea.MouseWheelDown:
				m.showMatch(m.capture.match + wheelLines)
			}
			return m, nil
		}

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

	// A question takes every press while it is up. A click on an answer gives
	// it; one anywhere else is ignored rather than taken as an answer, since a
	// dialog asking whether to quit should not be dismissed by a stray click.
	if m.modal.Type != ModalNone {
		if choice, hit := m.modal.buttonAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			return m.answerModal(choice)
		}
		return m, nil
	}

	// A press inside the preferences window works it; anywhere else closes it,
	// rather than acting on something underneath it.
	if m.preferences.active {
		if m.inPreferences(m.windowWidth, m.windowHeight, mouse.X, mouse.Y) {
			return m.handlePreferencesClick(mouse.X, mouse.Y)
		}
		m.closePreferences()
		return m, nil
	}

	// A press inside a form moves to the field pressed; anywhere else closes it,
	// rather than acting on something underneath it.
	if m.form.active {
		if button, hit := m.formButtonAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			if button == "cancel" {
				m.closeFileForm()
				return m, nil
			}
			return m.submitFileForm()
		}
		if i, hit := m.formMatchAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			m.form.match = i

			// A column list is ticked rather than picked from, and stays open:
			// choosing columns is a several-at-a-time job.
			if m.form.awaiting != "" {
				return m.useOptionValue()
			}
			if field := m.form.current(); field != nil && field.kind == fieldColumns {
				return m.toggleColumn()
			}

			updated, cmd := m.useMatchInForm()
			updated.clearDependentFields()
			return updated, cmd
		}
		if i, hit := m.formFieldAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			// Clicking a field that completes shows its list straight away, and
			// shows all of it: clicking "Keyspace" is asking which keyspaces
			// there are, not asking to have the one already in it completed.
			m.form.focusField(i)
			if field := m.form.current(); field != nil {
				switch field.kind {
				case fieldKeyspace, fieldTable, fieldPath, fieldColumns:
					return m.listField()
				case fieldOption:
					m.showOptionValues()
					return m, nil
				case fieldYesNo:
					return m.toggleField()
				}
			}
			return m, nil
		}
		if m.inFileForm(m.windowWidth, m.windowHeight, mouse.X, mouse.Y) {
			return m, nil
		}
		m.closeFileForm()
		return m, nil
	}

	// A press inside the FILE menu picks an entry; anywhere else closes it,
	// rather than acting on something underneath it. A press on FILE itself
	// falls through to the tab line, which toggles the menu shut.
	if m.fileMenu.active {
		if i, hit := m.fileMenuItemAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			// A press on the separator, or on an entry that cannot be picked,
			// leaves the menu as it is rather than moving the highlight
			// somewhere nothing can be chosen from.
			if !m.fileItemAvailable(fileMenuItems()[i]) {
				return m, nil
			}
			m.fileMenu.selected = i
			return m.chooseFileMenuItem()
		}
		m.closeFileMenu()
		if mode, _ := m.modeAt(m.windowWidth, mouse.X); mouse.Y == 0 && mode == fileMode {
			return m, nil
		}
	}

	// A press inside the capture window picks a format; anywhere else closes
	// it, rather than acting on something underneath it.
	if m.capture.active {
		if i, hit := m.formatAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			m.capture.format = i
			return m.chooseCaptureFormat()
		}
		if i, hit := m.matchAt(m.windowWidth, m.windowHeight, mouse.X, mouse.Y); hit {
			m.capture.match = i
			return m.useMatch()
		}
		if m.inCapturePanel(m.windowWidth, m.windowHeight, mouse.X, mouse.Y) {
			// The path step: a press in the box is aimed at the text field, so
			// leave what has been typed alone.
			return m, nil
		}

		// Clicking the control that opened it closes it, rather than closing
		// and reopening in the same press - which looks like nothing happened.
		reopens := m.pressOpensCapture(mouse)
		m.closeCapturePanel()
		if reopens {
			return m, nil
		}
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
	if m.historySearchMode {
		if updated, cmd, hit := m.clickHistoryCommand(mouse.X, mouse.Y); hit {
			return updated, cmd
		}

		// Inside the box but not on a command - the title, the search line, the
		// scrollbar - does nothing rather than closing it, so an aimed press
		// that misses by a row is not the same as walking away.
		if m.inHistoryOverlay(mouse.X, mouse.Y) {
			return m, nil
		}

		// Anywhere else closes it, the same as every other window, and the
		// press carries on to whatever it landed on: the tabs and the settings
		// are still live underneath.
		//
		// Except the control that opened it. Left to carry on, the press closed
		// the list and the History field reopened it in the same press, so it
		// looked like nothing happened - the bug the settings lists, Capture and
		// SAVE RESULTS each had in turn.
		reopens := m.pressOpensHistory(mouse)
		m.closeHistorySearch()
		if reopens {
			return m, nil
		}
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

	// A press on the schema tree picks what it lands on. The definition beside
	// it is text like any other view, and a press there starts a selection, so
	// what is on screen can still be copied out.
	if m.viewMode == "schema" {
		if _, hit := m.schemaRowAt(mouse.X, mouse.Y); hit {
			return m.clickSchema(mouse.X, mouse.Y)
		}
	}

	return m.beginSelection(mouse.X, mouse.Y)
}

// pressOpensHistory reports whether a press is on the History field, which is
// what opens the command list.
func (m *MainModel) pressOpensHistory(mouse tea.Mouse) bool {
	if mouse.Y != m.windowHeight-2 {
		return false
	}
	field, ok := m.topBar.fieldAt(m.windowWidth, mouse.X)
	return ok && field == infoHistory
}

// pressOpensCapture reports whether a press is on a control that opens the
// window: the FILE menu on the tab line, or the Capture field on the bottom
// line, which still opens it directly.
//
// A press on one of those closes what it opened rather than closing and
// reopening in the same press, which is what made clicking look like nothing
// had happened.
func (m *MainModel) pressOpensCapture(mouse tea.Mouse) bool {
	switch {
	case mouse.Y == 0:
		mode, _ := m.modeAt(m.windowWidth, mouse.X)
		return mode == fileMode

	case m.capture.kind == capturing && mouse.Y == m.windowHeight-1:
		setting, _, ok := m.statusBar.settingAt(m.windowWidth, mouse.X)
		return ok && setting == settingCapture
	}
	return false
}

// handleMouseWheel scrolls the view under the pointer.
func (m *MainModel) handleMouseWheel(mouse tea.Mouse) (*MainModel, tea.Cmd) {
	// The schema view is two panes, and the wheel moves whichever one it is
	// over rather than whichever was touched last.
	if m.viewMode == "schema" && !m.preferences.active && !m.form.active && !m.help.active {
		switch mouse.Button {
		case tea.MouseWheelUp:
			return m.scrollSchema(mouse.X, -wheelLines)
		case tea.MouseWheelDown:
			return m.scrollSchema(mouse.X, wheelLines)
		}
		return m, nil
	}

	// The preferences window is longer than the screen, so the wheel over it
	// moves whichever of its two lists is under the pointer.
	if m.preferences.active {
		delta := 0
		switch mouse.Button {
		case tea.MouseWheelUp:
			delta = -wheelLines
		case tea.MouseWheelDown:
			delta = wheelLines
		}
		if delta == 0 {
			return m, nil
		}
		if len(m.preferences.matches) > 0 {
			return m.scrollPrefMatches(delta)
		}
		return m.scrollPreferences(delta)
	}

	// A candidate list takes the wheel while it is showing: it is over the view
	// and it is the thing just asked for.
	if m.form.active && len(m.form.matches) > 0 {
		switch mouse.Button {
		case tea.MouseWheelUp:
			return m.scrollMatches(-wheelLines)
		case tea.MouseWheelDown:
			return m.scrollMatches(wheelLines)
		}
		return m, nil
	}

	// Any modifier plus a vertical wheel means horizontal scrolling.
	modified := mouse.Mod.Contains(tea.ModShift) ||
		mouse.Mod.Contains(tea.ModAlt) ||
		mouse.Mod.Contains(tea.ModCtrl)

	switch mouse.Button {
	case tea.MouseWheelUp:
		if modified && m.scrollsSideways() {
			return m.handleMouseWheelLeft()
		}
		return m.handleMouseWheelUp()
	case tea.MouseWheelDown:
		if modified && m.scrollsSideways() {
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

// scrollsSideways reports whether the view in front of you has anything to
// scroll sideways.
//
// The Trace view was left out of this, so a modifier and the wheel scrolled it
// up and down instead - on a trace far wider than the screen, which is the one
// view where reaching sideways matters most.
func (m *MainModel) scrollsSideways() bool {
	return (m.viewMode == "table" && m.hasTable) || (m.viewMode == "trace" && m.hasTrace)
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

		// Clear cache to force rebuild (important!)
		m.cachedTableLines = nil
		m.renderResults(m.resultRows())

		// Update row count
		m.rowCount = int(m.slidingWindow.TotalRowsSeen)
	}
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
	if mode == fileMode {
		span, ok := m.tabSpanFor(m.windowWidth, fileMode)
		if !ok {
			return m, nil
		}
		return m.openFileMenu(span.start)
	}

	switch mode {
	case "history":
		return m.showConsole()
	case "table":
		return m.showResults()
	case "trace":
		return m.showTrace()
	case "schema":
		return m.showSchema()
	case "ai":
		return m.showChat()
	}
	return m, nil
}

// clickStatusSetting opens the list of values for a setting on the status line,
// or closes the one already open.
func (m *MainModel) clickStatusSetting(col int) (*MainModel, tea.Cmd) {
	wasOpen, openSetting := m.chooser.active, m.chooser.setting
	m.closeSettingChooser()

	setting, anchorX, ok := m.statusBar.settingAt(m.windowWidth, col)
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

	// Capture is not one either: it takes a format and then a path, which is a
	// step more than a list of values can do.
	if setting == settingCapture {
		return m.openCapturePanel(anchorX)
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
	field, ok := m.topBar.fieldAt(m.windowWidth, col)
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
