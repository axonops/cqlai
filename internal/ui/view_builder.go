package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/config"
)

// newView wraps rendered content in the terminal state cqlai wants.
//
// v2 makes this declarative: rather than startup options and imperative
// commands, the alternate screen and mouse mode are fields recomputed on every
// render, so they cannot drift out of step with what is on screen.
//
// Taking the mouse means the terminal stops selecting text for us, so cqlai
// draws the selection itself instead - see selection.go. That is what lets the
// tabs and the bottom line be clickable and text still be selectable, which no
// choice of DECSET mode gives you on its own. MOUSE OFF hands the mouse back
// for anyone who would rather have the terminal do it.
//
// Either way the wheel still scrolls: with reporting on it arrives as a wheel
// event, and with it off, through alternate scroll mode, which v2 does not
// manage and mouse_mode.go sets up.
func (m *MainModel) newView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true

	// CellMotion is DECSET 1002: presses, releases, the wheel, and motion while
	// a button is held, which is what following a drag needs. AllMotion adds
	// motion with no button down, and nothing here tracks the pointer at rest.
	if m.mouseEnabled {
		v.MouseMode = tea.MouseModeCellMotion
	} else {
		v.MouseMode = tea.MouseModeNone
	}

	return v
}

// View renders the main model.
func (m *MainModel) View() tea.View {
	if !m.ready {
		return m.newView("")
	}

	m.topBar.LastCommand = m.lastCommand
	if m.session != nil {
		autoFetch := m.session.AutoFetch()
		m.statusBar.AutoFetch = autoFetch
		m.topBar.AutoFetch = autoFetch // drives the "+" on the row count
	}
	if m.slidingWindow != nil {
		m.topBar.HasMoreData = m.slidingWindow.hasMoreData
	}
	if m.session != nil {
		currentKeyspace := ""
		if m.sessionManager != nil {
			currentKeyspace = m.sessionManager.CurrentKeyspace()
		}
		m.statusBar.Keyspace = currentKeyspace
		m.statusBar.Username = m.session.Username()
		m.statusBar.Tracing = m.session.Tracing()
		m.statusBar.HasTraceData = m.hasTrace
		m.statusBar.Consistency = m.session.Consistency()
		m.statusBar.PagingSize = m.session.PageSize()
		m.statusBar.Version = m.session.CassandraVersion()
		// Get the current output format
		if m.sessionManager != nil {
			switch m.sessionManager.GetOutputFormat() {
			case config.OutputFormatTable:
				m.statusBar.OutputFormat = "TABLE"
			case config.OutputFormatASCII:
				m.statusBar.OutputFormat = "ASCII"
			case config.OutputFormatExpand:
				m.statusBar.OutputFormat = "EXPAND"
			case config.OutputFormatJSON:
				m.statusBar.OutputFormat = "JSON"
			default:
				m.statusBar.OutputFormat = "TABLE"
			}
		} else {
			m.statusBar.OutputFormat = "TABLE"
		}
	}

	// Get the active viewport for scroll info
	activeViewport := m.historyViewport
	switch m.viewMode {
	case "table":
		if m.hasTable {
			activeViewport = m.tableViewport
		}
	case "trace":
		if m.hasTrace {
			activeViewport = m.traceViewport
		}
	}

	// Add mode indicator and available view keys
	var scrollInfo string
	switch m.viewMode {
	case "ai":
		// No need to show navigation keys in AI view since they're less relevant
		scrollInfo = ""
	case "trace":
		modeIndicator := m.styles.AccentText.Render("[TRACE VIEW]")
		scrollInfo = " " + modeIndicator
		scrollInfo += " " + m.styles.MutedText.Render("[F2: History | F5: AI]")
		if m.hasTable {
			scrollInfo += " " + m.styles.MutedText.Render("[F3: Table]")
		}
	case "table":
		modeIndicator := m.styles.AccentText.Render("[TABLE VIEW]")
		scrollInfo = " " + modeIndicator
		scrollInfo += " " + m.styles.MutedText.Render("[F2: History | F5: AI | F6: Toggle Types]")
		if m.hasTrace {
			scrollInfo += " " + m.styles.MutedText.Render("[F4: Trace]")
		}
	default: // history view
		modeIndicator := m.styles.MutedText.Render("[CQL VIEW]")
		scrollInfo = " " + modeIndicator
		scrollInfo += " " + m.styles.MutedText.Render("[F5: AI]")
		if m.hasTable {
			scrollInfo += " " + m.styles.MutedText.Render("[F3: Table]")
		}
		if m.hasTrace {
			scrollInfo += " " + m.styles.MutedText.Render("[F4: Trace]")
		}
	}

	// Add scroll indicator if content is scrollable
	if activeViewport.TotalLineCount() > activeViewport.Height() {
		scrollPercent := activeViewport.ScrollPercent()
		if scrollPercent == 0 { //nolint:gocritic // more readable as if
			scrollInfo += " [TOP]"
		} else if scrollPercent >= 0.99 {
			scrollInfo += " [BOTTOM]"
		} else {
			scrollInfo += fmt.Sprintf(" [%d%%]", int(scrollPercent*100))
		}
	}

	// Add horizontal scroll indicator if table/trace is wider than viewport
	switch {
	case m.viewMode == "trace" && m.hasTrace && m.traceTableWidth > m.traceViewport.Width():
		if m.traceHorizontalOffset == 0 { //nolint:gocritic // more readable as if
			scrollInfo += " | H:LEFT"
		} else if m.traceHorizontalOffset >= m.traceTableWidth-m.traceViewport.Width() {
			scrollInfo += " | H:RIGHT"
		} else {
			// Calculate horizontal scroll percentage
			maxOffset := m.traceTableWidth - m.traceViewport.Width()
			hScrollPercent := float64(m.traceHorizontalOffset) / float64(maxOffset)
			scrollInfo += fmt.Sprintf(" | H:%d%%", int(hScrollPercent*100))
		}
	case m.hasTable && m.viewMode == "table" && m.tableWidth > m.tableViewport.Width():
		if m.horizontalOffset == 0 { //nolint:gocritic // more readable as if
			scrollInfo += " | H:LEFT"
		} else if m.horizontalOffset >= m.tableWidth-m.tableViewport.Width() {
			scrollInfo += " | H:RIGHT"
		} else {
			// Calculate horizontal scroll percentage
			maxOffset := m.tableWidth - m.tableViewport.Width()
			hScrollPercent := float64(m.horizontalOffset) / float64(maxOffset)
			scrollInfo += fmt.Sprintf(" | H:%d%%", int(hScrollPercent*100))
		}
	}

	// Add sliding window indicator if data has been dropped
	if m.slidingWindow != nil && m.slidingWindow.DataDroppedAtStart {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		scrollInfo += " " + warningStyle.Render(fmt.Sprintf("[Rows %d-%d, earlier rows dropped]",
			m.slidingWindow.FirstRowIndex+1,
			m.slidingWindow.FirstRowIndex+int64(len(m.slidingWindow.Rows))))
	}

	// Add output format indicator
	if m.sessionManager != nil {
		outputFormat := m.sessionManager.GetOutputFormat()
		switch outputFormat {
		case config.OutputFormatExpand:
			expandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF")).Bold(true)
			scrollInfo += " " + expandStyle.Render("[EXPAND ON]")
		case config.OutputFormatASCII:
			asciiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF"))
			scrollInfo += " " + asciiStyle.Render("[ASCII]")
		case config.OutputFormatJSON:
			jsonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF"))
			scrollInfo += " " + jsonStyle.Render("[JSON]")
		}
	}

	// Say which way the mouse is set. Without this, someone who has turned it
	// off, or run into a terminal that will not report it, just finds that
	// clicking the tabs does nothing, with nothing on screen explaining why.
	if m.mouseEnabled {
		mouseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF"))
		scrollInfo += " " + mouseStyle.Render("[MOUSE]")
	} else {
		mouseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5F5F5F"))
		scrollInfo += " " + mouseStyle.Render("[MOUSE OFF]")
	}

	// Build the input section
	var inputSection string
	if m.viewMode == "ai" && m.aiConversationActive {
		// Use AI conversation input
		if m.aiProcessing {
			inputSection = m.styles.MutedText.Render("Processing... ")
		} else {
			inputSection = m.aiConversationInput.View()
		}
	} else {
		// Update placeholder if there's more data to fetch
		if m.viewMode == "table" && m.slidingWindow != nil && m.slidingWindow.hasMoreData {
			if m.session != nil && !m.session.AutoFetch() {
				// Temporarily change placeholder to show more data hint
				originalPlaceholder := m.input.Placeholder
				m.input.Placeholder = "▼ MORE DATA AVAILABLE - Press Space or PgDn to load next page ▼"
				inputSection = m.input.View()
				m.input.Placeholder = originalPlaceholder // Restore original
			} else {
				inputSection = m.input.View()
			}
		} else {
			inputSection = m.input.View()
		}

		// If in multi-line mode, show the buffered lines above the input
		if m.multiLineMode && len(m.multiLineBuffer) > 0 {
			bufferedLines := m.styles.MutedText.Render("... " + strings.Join(m.multiLineBuffer, "\n... "))
			inputSection = bufferedLines + "\n" + inputSection
		}
	}

	// Function key hints are now shown in the status bar

	// Use the appropriate viewport based on mode
	var viewportContent string
	var viewportWidth int

	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		// For AI view, just get the viewport content
		viewportWidth = m.aiConversationViewport.Width()
		viewportContent = m.aiConversationViewport.View()
	case m.viewMode == "trace":
		viewportWidth = m.historyViewport.Width()
		if m.hasTrace {
			viewportWidth = m.traceViewport.Width()
			viewportContent = m.traceViewport.View()
		} else {
			// Create a temporary viewport for empty trace message
			emptyMsg := m.styles.MutedText.Render("\n  No trace data available. Enable tracing with 'TRACING ON' to capture query traces.\n")
			tempViewport := m.historyViewport
			tempViewport.SetContent(emptyMsg)
			viewportContent = tempViewport.View()
		}
	case m.viewMode == "table":
		viewportWidth = m.historyViewport.Width()
		if m.hasTable {
			viewportWidth = m.tableViewport.Width()
			viewportContent = m.tableViewport.View()
		} else {
			// Create a temporary viewport for empty table message
			emptyMsg := m.styles.MutedText.Render("\n  No table data available. Execute a SELECT query to view results in table format.\n")
			tempViewport := m.historyViewport
			tempViewport.SetContent(emptyMsg)
			viewportContent = tempViewport.View()
		}
	default:
		viewportContent = m.historyViewport.View()
		viewportWidth = m.historyViewport.Width()
	}

	// Build the main view with proper sticky header overlay
	var finalView string

	// Query facts - AutoFetch, the last command, timing, row count. These sit at
	// the bottom now, just above the connection bar, so the top of the screen is
	// only the tabs and the result starts a line higher.
	infoBar := m.topBar.View(viewportWidth, m.styles, m.viewMode)
	if infoBar == "" {
		// Never empty, or the row collapses and the layout shifts
		infoBar = " "
	}

	// Build the viewport section with sticky header for tables.
	//
	// stickyHeaderRows decides whether the header is drawn, and the selection
	// reads the same function to know which rows are showing the top of the
	// table rather than the lines the viewport is scrolled to.
	viewportSection := viewportContent
	if headerLineCount := m.stickyHeaderRows(); headerLineCount > 0 {
		header := m.buildTableStickyHeader()
		lines := strings.Split(viewportContent, "\n")
		if len(lines) > headerLineCount {
			viewportSection = header + "\n" + strings.Join(lines[headerLineCount:], "\n")
		} else {
			viewportSection = header + "\n" + viewportContent
		}
	}

	// Paint the mouse selection on last, so it covers whatever styling the
	// content already had.
	viewportSection = m.highlightSelection(viewportSection)

	// Build the final view. The tabs get their own line above the status bar so
	// the modes and their keys are always on screen, rather than something you
	// have to have read the README to know about.
	finalView = lipgloss.JoinVertical(lipgloss.Left,
		m.ViewTabBar(viewportWidth),
		viewportSection,
		inputSection,
		infoBar,
		m.statusBar.View(viewportWidth, m.styles, m.viewMode)+scrollInfo,
	)

	// Calculate the actual screen dimensions
	screenHeight := strings.Count(finalView, "\n") + 1
	screenWidth := viewportWidth

	// Create a layer manager for overlays
	layerManager := NewLayerManager(screenWidth, screenHeight)

	// If completions are showing, add as a layer
	if m.showCompletions && len(m.completions) > 0 {
		completionModal := NewCompletionModal(m.completions, m.completionIndex)
		completionModal.scrollOffset = m.completionScrollOffset
		content := completionModal.RenderContent(m.styles)

		// Position at bottom left, just above the prompt
		modalHeight := strings.Count(content, "\n") + 1
		layer := Layer{
			Content: content,
			X:       0,
			Y:       screenHeight - modalHeight - 2,
			Width:   lipgloss.Width(content),
			Height:  modalHeight,
			ZIndex:  100,
		}
		layerManager.AddLayer(layer)
	}

	// If history modal is showing, add as a layer
	if m.showHistoryModal && len(m.commandHistory) > 0 {
		historyModal := NewHistoryModal(m.commandHistory, m.historyModalIndex, viewportWidth)
		historyModal.scrollOffset = m.historyModalScrollOffset
		content := historyModal.RenderContent(m.styles)

		// Position at bottom left, just above the prompt
		modalHeight := strings.Count(content, "\n") + 1
		layer := Layer{
			Content: content,
			X:       0,
			Y:       screenHeight - modalHeight - 2,
			Width:   lipgloss.Width(content),
			Height:  modalHeight,
			ZIndex:  100,
		}
		layerManager.AddLayer(layer)
	}

	// If history search modal is showing, add as a layer
	if m.historySearchMode {
		searchModal := NewHistorySearchModal(m.historySearchQuery, m.historySearchResults, m.historySearchIndex, viewportWidth)
		searchModal.scrollOffset = m.historySearchScrollOffset
		content := searchModal.RenderContent(m.styles)

		// Calculate the actual width from the rendered content
		modalLines := strings.Split(content, "\n")
		modalWidth := 0
		for _, line := range modalLines {
			lineWidth := lipgloss.Width(line)
			if lineWidth > modalWidth {
				modalWidth = lineWidth
			}
		}

		// Position at bottom left, just above the prompt
		modalHeight := len(modalLines)
		layer := Layer{
			Content: content,
			X:       0,
			Y:       screenHeight - modalHeight - 2,
			Width:   modalWidth,
			Height:  modalHeight,
			ZIndex:  100,
		}
		layerManager.AddLayer(layer)
	}

	// If AI CQL modal is showing, add as a layer
	if m.aiCQLModal != nil && m.aiCQLModal.Active {
		content := m.aiCQLModal.Render(screenWidth, screenHeight, m.styles)
		// The modal renders as a full overlay, so we return it directly
		return m.newView(content)
	}

	// If AI selection modal is showing, add as a layer
	if m.aiSelectionModal != nil && m.aiSelectionModal.Active {
		content := m.aiSelectionModal.Render(screenWidth, screenHeight, m.styles)
		// The modal renders as a full overlay, so we return it directly
		return m.newView(content)
	}

	// If save modal is showing, add it as a layer
	if m.saveModalActive {
		modalContent := m.renderSaveModal(screenWidth, screenHeight)
		modalLayer := RenderModal(modalContent, screenWidth, screenHeight)
		layerManager.AddLayer(modalLayer)
	}

	// Apply all layers to the final view
	// The settings chooser sits above everything: it is the thing just clicked.
	if layer, ok := m.viewSettingChooser(screenWidth, screenHeight); ok {
		layerManager.AddLayer(layer)
	}

	finalView = layerManager.Render(finalView)

	// If modal is showing, render it as an overlay
	if m.modal.Type != ModalNone {
		// Use the actual window dimensions
		screenWidth := m.windowWidth
		screenHeight := m.windowHeight
		if screenWidth == 0 {
			// Fallback if window dimensions not yet set
			screenWidth = viewportWidth
		}
		if screenHeight == 0 {
			// Fallback if window dimensions not yet set
			screenHeight = m.historyViewport.Height() + 3
		}

		// Render the modal overlay with the current view as background
		return m.newView(m.modal.Render(screenWidth, screenHeight, m.styles, finalView))
	}

	return m.newView(finalView)
}

// getWelcomeMessage returns the welcome message for the application
func (m *MainModel) getWelcomeMessage() string {
	var welcome strings.Builder

	// Welcome banner
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("╔═══════════════════════════════════════════════════════╗"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("║                  Welcome to CQLAI                     ║"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("╚═══════════════════════════════════════════════════════╝"))
	welcome.WriteString("\n\n")

	// Connection status
	if m.session != nil && m.session.Session != nil {
		welcome.WriteString(m.styles.SuccessText.Render("✓ Connected to Cassandra"))
		welcome.WriteString("\n")
		currentKeyspace := ""
		if m.sessionManager != nil {
			currentKeyspace = m.sessionManager.CurrentKeyspace()
		}
		if currentKeyspace != "" {
			welcome.WriteString(m.styles.MutedText.Render(fmt.Sprintf("  Keyspace: %s", currentKeyspace)))
			welcome.WriteString("\n")
		}
	} else {
		welcome.WriteString(m.styles.ErrorText.Render("✗ Not connected to Cassandra"))
		welcome.WriteString("\n")
	}
	welcome.WriteString("\n")

	// Quick help
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("Quick Commands:"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • Type CQL queries and press Enter"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • DESCRIBE KEYSPACES - List all keyspaces"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • DESCRIBE TABLES - List tables in current keyspace"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • USE <keyspace> - Switch to a keyspace"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • CLEAR - Clear the screen"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • EXIT or Ctrl+C - Quit the application"))
	welcome.WriteString("\n\n")

	// Keyboard shortcuts
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("Keyboard Shortcuts:"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • Tab - Auto-complete commands"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • ↑/↓ - Navigate command history"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • Ctrl+R - Search command history"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • Alt+↑ / Alt+↓ - Scroll table line by line"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • PgUp/PgDn - Scroll through results page by page"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • Alt+← / Alt+→ - Horizontal scroll for wide tables"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • F2 - Switch to history/query view"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • F3 - Switch to table view"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • F4 - Switch to trace view"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • F5 - Switch to AI assistant mode"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.MutedText.Render("  • F6 - Toggle column data types (in table view)"))
	welcome.WriteString("\n\n")

	return welcome.String()
}
