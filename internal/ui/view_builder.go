package ui

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// View renders the main model.
func (m MainModel) View() string {
	if !m.ready {
		return ""
	}

	m.topBar.LastCommand = m.lastCommand
	if m.session != nil {
		m.statusBar.Keyspace = m.session.CurrentKeyspace()
		m.statusBar.Tracing = m.session.Tracing()
		m.statusBar.Consistency = m.session.Consistency()
		m.statusBar.PagingSize = m.session.PageSize()
		m.statusBar.Version = m.session.CassandraVersion()
		// Get the current output format
		switch m.session.GetOutputFormat() {
		case db.OutputFormatTable:
			m.statusBar.OutputFormat = "TABLE"
		case db.OutputFormatASCII:
			m.statusBar.OutputFormat = "ASCII"
		case db.OutputFormatExpand:
			m.statusBar.OutputFormat = "EXPAND"
		case db.OutputFormatJSON:
			m.statusBar.OutputFormat = "JSON"
		default:
			m.statusBar.OutputFormat = "TABLE"
		}
	}

	// Get the active viewport for scroll info
	activeViewport := m.historyViewport
	if m.viewMode == "table" && m.hasTable {
		activeViewport = m.tableViewport
	}

	// Add mode indicator and scroll info
	var scrollInfo string
	if m.hasTable {
		// Show view mode indicator when table is available
		if m.viewMode == "table" {
			modeIndicator := m.styles.AccentText.Render("[TABLE VIEW]")
			scrollInfo = " " + modeIndicator
			// Add F2 hint for data types toggle
			if m.showDataTypes {
				scrollInfo += " " + m.styles.MutedText.Render("[F2: Hide Types]")
			} else {
				scrollInfo += " " + m.styles.MutedText.Render("[F2: Show Types]")
			}
		} else {
			modeIndicator := m.styles.MutedText.Render("[HISTORY VIEW]")
			scrollInfo = " " + modeIndicator
		}
	}

	// Add scroll indicator if content is scrollable
	if activeViewport.TotalLineCount() > activeViewport.Height {
		scrollPercent := activeViewport.ScrollPercent()
		if scrollPercent == 0 {
			scrollInfo += " [TOP]"
		} else if scrollPercent >= 0.99 {
			scrollInfo += " [BOTTOM]"
		} else {
			scrollInfo += fmt.Sprintf(" [%d%%]", int(scrollPercent*100))
		}
	}

	// Add horizontal scroll indicator if table is wider than viewport
	if m.hasTable && m.viewMode == "table" && m.tableWidth > m.tableViewport.Width {
		if m.horizontalOffset == 0 {
			scrollInfo += " | H:LEFT"
		} else if m.horizontalOffset >= m.tableWidth-m.tableViewport.Width {
			scrollInfo += " | H:RIGHT"
		} else {
			// Calculate horizontal scroll percentage
			maxOffset := m.tableWidth - m.tableViewport.Width
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
	if m.session != nil {
		outputFormat := m.session.GetOutputFormat()
		if outputFormat == db.OutputFormatExpand {
			expandStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF")).Bold(true)
			scrollInfo += " " + expandStyle.Render("[EXPAND ON]")
		} else if outputFormat == db.OutputFormatASCII {
			asciiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87D7FF"))
			scrollInfo += " " + asciiStyle.Render("[ASCII]")
		}
	}

	// Build the input section
	inputSection := m.input.View()

	// If in multi-line mode, show the buffered lines above the input
	if m.multiLineMode && len(m.multiLineBuffer) > 0 {
		bufferedLines := m.styles.MutedText.Render("... " + strings.Join(m.multiLineBuffer, "\n... "))
		inputSection = bufferedLines + "\n" + inputSection
	}

	// Add hints about function keys when table is available
	if m.hasTable && m.input.Value() == "" {
		hints := []string{}
		hints = append(hints, "F1: switch views")
		// Don't show F2 hint here since it's in the status bar for table view
		if len(hints) > 0 {
			hint := m.styles.MutedText.Render("  (" + strings.Join(hints, " | ") + ")")
			inputSection = inputSection + hint
		}
	}

	// Use the appropriate viewport based on mode
	var viewportContent string
	var viewportWidth int

	if m.viewMode == "table" && m.hasTable {
		viewportWidth = m.tableViewport.Width
		viewportContent = m.tableViewport.View()
	} else {
		viewportContent = m.historyViewport.View()
		viewportWidth = m.historyViewport.Width
	}

	// Build the main view with proper sticky header overlay
	var finalView string

	// Always show the top bar
	topBar := m.topBar.View(viewportWidth, m.styles)

	// Build the viewport section with sticky header for tables
	var viewportSection string
	if m.viewMode == "table" && m.hasTable && m.tableViewport.YOffset > 0 {
		// When table is scrolled, prepend the header
		header := m.buildTableStickyHeader()
		if header != "" {
			// Get viewport lines and remove the top lines that would duplicate the header
			lines := strings.Split(viewportContent, "\n")

			// Skip the lines that would be covered by the sticky header (3 lines: border, header, separator)
			headerLineCount := 3
			if len(lines) > headerLineCount {
				remainingLines := lines[headerLineCount:]
				viewportSection = header + "\n" + strings.Join(remainingLines, "\n")
			} else {
				viewportSection = header + "\n" + viewportContent
			}
		} else {
			viewportSection = viewportContent
		}
	} else {
		viewportSection = viewportContent
	}

	// Build the final view
	finalView = lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		viewportSection,
		inputSection,
		m.statusBar.View(viewportWidth, m.styles)+scrollInfo,
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

	// Apply all layers to the final view
	finalView = layerManager.Render(finalView)

	// If AI modal is showing, render it as an overlay
	if m.showAIModal {
		// Get the window dimensions from the viewport
		screenWidth := viewportWidth
		screenHeight := m.historyViewport.Height + 3 // Include top bar, input, and status bar

		// Render the AI modal overlay with the current view as background
		return m.aiModal.Render(screenWidth, screenHeight, m.styles)
	}
	
	// If modal is showing, render it as an overlay
	if m.modal.Type != ModalNone {
		// Get the window dimensions from the viewport
		screenWidth := viewportWidth
		screenHeight := m.historyViewport.Height + 3 // Include top bar, input, and status bar

		// Render the modal overlay with the current view as background
		return m.modal.Render(screenWidth, screenHeight, m.styles, finalView)
	}

	return finalView
}

// getWelcomeMessage returns the welcome message for the application
func (m MainModel) getWelcomeMessage() string {
	var welcome strings.Builder

	// Welcome banner
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("╔═══════════════════════════════════════════════════════╗"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("║            Welcome to CQLAI - CQL Shell               ║"))
	welcome.WriteString("\n")
	welcome.WriteString(m.styles.AccentText.Bold(true).Render("╚═══════════════════════════════════════════════════════╝"))
	welcome.WriteString("\n\n")

	// Connection status
	if m.session != nil && m.session.Session != nil {
		welcome.WriteString(m.styles.SuccessText.Render("✓ Connected to Cassandra"))
		welcome.WriteString("\n")
		if m.session.CurrentKeyspace() != "" {
			welcome.WriteString(m.styles.MutedText.Render(fmt.Sprintf("  Keyspace: %s", m.session.CurrentKeyspace())))
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
	welcome.WriteString(m.styles.MutedText.Render("  • F1 - Switch between history/table view"))
	welcome.WriteString("\n\n")

	return welcome.String()
}
