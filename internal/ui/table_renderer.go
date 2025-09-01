package ui

import (
	"fmt"
	"regexp"
	"strings"
)

// formatTableForViewport formats a 2D array of strings as a table for the viewport
func (m *MainModel) formatTableForViewport(data [][]string) string {
	if len(data) == 0 {
		return ""
	}

	// Calculate column widths (using rune count for proper Unicode handling)
	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			plainCell := stripAnsi(cell)
			cellWidth := len([]rune(plainCell)) // Count runes, not bytes
			if cellWidth > colWidths[i] {
				colWidths[i] = cellWidth
			}
		}
	}

	// Store column widths for sticky header alignment
	m.columnWidths = colWidths

	// Build the full table
	fullLines := m.buildFullTable(data, colWidths)

	// Calculate total table width
	if len(fullLines) > 0 {
		m.tableWidth = len(stripAnsi(fullLines[0]))
	}

	// Apply horizontal scrolling only if offset is > 0
	// If table is wider but offset is 0, show the leftmost part completely
	if m.horizontalOffset > 0 {
		scrolledLines := applyHorizontalScrollWithANSI(fullLines, m.horizontalOffset, m.tableViewport.Width)
		return strings.Join(scrolledLines, "\n")
	}

	// Join all lines
	return strings.Join(fullLines, "\n")
}

// buildFullTable builds the complete table with borders and formatting
func (m *MainModel) buildFullTable(data [][]string, colWidths []int) []string {
	var lines []string

	// Top border
	topBorder := "┌"
	for i, width := range colWidths {
		topBorder += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			topBorder += "┬"
		}
	}
	topBorder += "┐"
	lines = append(lines, topBorder)

	// Header row (if exists)
	if len(data) > 0 {
		headerRow := "│"
		for i, cell := range data[0] {
			// Style header cells and ensure reset at end
			styledCell := m.styles.AccentText.Bold(true).Render(cell) + "\x1b[0m"
			plainCell := stripAnsi(cell)
			padding := colWidths[i] - len([]rune(plainCell))
			headerRow += " " + styledCell + strings.Repeat(" ", padding) + " │"
		}
		lines = append(lines, headerRow)

		// Header separator
		separator := "├"
		for i, width := range colWidths {
			separator += strings.Repeat("─", width+2)
			if i < len(colWidths)-1 {
				separator += "┼"
			}
		}
		separator += "┤"
		lines = append(lines, separator)
	}

	// Data rows
	for i, row := range data {
		if i == 0 {
			continue // Skip header row, already processed
		}
		dataRow := "│"
		for j, cell := range row {
			plainCell := stripAnsi(cell)
			padding := colWidths[j] - len([]rune(plainCell))
			dataRow += " " + cell + strings.Repeat(" ", padding) + " │"
		}
		lines = append(lines, dataRow)
	}

	// Bottom border
	bottomBorder := "└"
	for i, width := range colWidths {
		bottomBorder += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			bottomBorder += "┴"
		}
	}
	bottomBorder += "┘"
	lines = append(lines, bottomBorder)

	return lines
}

// refreshTableView refreshes the table view with the current data and scroll position
func (m *MainModel) refreshTableView() {
	if m.lastTableData == nil {
		return
	}

	// Since we're using separate viewports, update the table viewport directly
	if m.viewMode == "table" && m.hasTable {
		// Rebuild the table with new scroll position
		tableStr := m.formatTableForViewport(m.lastTableData)

		// Store the current scroll position
		currentYOffset := m.tableViewport.YOffset

		// Update the table viewport content
		m.tableViewport.SetContent(tableStr)

		// Restore scroll position
		m.tableViewport.YOffset = currentYOffset
	}
}

// buildTableStickyHeader builds a complete sticky header with border and column names
func (m *MainModel) buildTableStickyHeader() string {
	if len(m.tableHeaders) == 0 || m.lastTableData == nil || m.columnWidths == nil {
		return ""
	}

	var lines []string
	colWidths := m.columnWidths

	// Build top border
	topBorder := "┌"
	for i, width := range colWidths {
		topBorder += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			topBorder += "┬"
		}
	}
	topBorder += "┐"
	lines = append(lines, topBorder)

	// Build header row with styling
	headerRow := "│"
	for i, header := range m.tableHeaders {
		if i < len(colWidths) {
			padding := colWidths[i] - len([]rune(header))
			if padding < 0 {
				padding = 0
			}
			// Style the header
			styledHeader := m.styles.AccentText.Bold(true).Render(header) + "\x1b[0m"
			headerRow += " " + styledHeader + strings.Repeat(" ", padding) + " │"
		}
	}
	lines = append(lines, headerRow)

	// Build separator
	separator := "├"
	for i, width := range colWidths {
		separator += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			separator += "┼"
		}
	}
	separator += "┤"
	lines = append(lines, separator)

	// Apply horizontal scrolling if needed
	if m.horizontalOffset > 0 || m.tableWidth > m.tableViewport.Width {
		scrolledLines := applyHorizontalScrollWithANSI(lines, m.horizontalOffset, m.tableViewport.Width)
		return strings.Join(scrolledLines, "\n")
	}

	return strings.Join(lines, "\n")
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// refreshTraceView refreshes the trace viewport using the existing table renderer
func (m *MainModel) refreshTraceView() {
	if !m.hasTrace || m.traceData == nil {
		return
	}

	// Create summary line
	summaryLine := ""
	if m.traceInfo != nil {
		// Highlight "Trace Session" with accent color
		highlightedTitle := m.styles.AccentText.Bold(true).Render("Trace Session")
		summaryLine = fmt.Sprintf("%s - Coordinator: %s | Total Duration: %d μs\n",
			highlightedTitle, m.traceInfo.Coordinator, m.traceInfo.Duration)
	}

	// Temporarily swap in trace data and settings
	originalOffset := m.horizontalOffset
	originalData := m.lastTableData
	originalWidth := m.tableWidth
	originalHeaders := m.tableHeaders
	originalColWidths := m.columnWidths

	// Set trace data temporarily
	m.horizontalOffset = m.traceHorizontalOffset
	m.lastTableData = m.traceData

	// Format using existing table renderer
	traceTable := m.formatTableForViewport(m.traceData)

	// Store trace-specific values
	m.traceTableWidth = m.tableWidth
	m.traceColumnWidths = m.columnWidths

	// Restore original table values
	m.horizontalOffset = originalOffset
	m.lastTableData = originalData
	m.tableWidth = originalWidth
	m.tableHeaders = originalHeaders
	m.columnWidths = originalColWidths

	// Prepend summary line to the table
	finalContent := summaryLine + traceTable
	m.traceViewport.SetContent(finalContent)
}
