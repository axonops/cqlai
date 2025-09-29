package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
)

// formatTableForViewport formats a 2D array of strings as a table for the viewport
func (m *MainModel) formatTableForViewport(data [][]string) string {
	if len(data) == 0 {
		return ""
	}

	// Check if we need to rebuild the table (data changed)
	needsRebuild := m.cachedTableLines == nil || !m.isSameTableData(data)

	if needsRebuild {
		// Use initial column widths if available, otherwise calculate them
		var colWidths []int
		if m.initialColumnWidths != nil && len(m.initialColumnWidths) == len(data[0]) {
			// Use preserved initial widths for consistency
			colWidths = make([]int, len(m.initialColumnWidths))
			copy(colWidths, m.initialColumnWidths)
		} else {
			// Calculate column widths (using rune count for proper Unicode handling)
			colWidths = make([]int, len(data[0]))
			for _, row := range data {
				for i, cell := range row {
					plainCell := stripAnsi(cell)
					cellWidth := len([]rune(plainCell)) // Count runes, not bytes
					if cellWidth > colWidths[i] {
						colWidths[i] = cellWidth
					}
				}
			}

			// Apply maximum column width cap for multi-line content
			// This ensures consistent column widths across all pages
			maxColWidth := 80
			for i := range colWidths {
				if colWidths[i] > maxColWidth {
					colWidths[i] = maxColWidth
				}
			}

			// Store as initial widths for consistency across pagination
			m.initialColumnWidths = make([]int, len(colWidths))
			copy(m.initialColumnWidths, colWidths)
		}

		// Store column widths for sticky header alignment
		m.columnWidths = colWidths

		// Check if any cell has multi-line content
		hasMultiLine := false
		for _, row := range data {
			for _, cell := range row {
				if strings.Contains(cell, "\n") || len([]rune(stripAnsi(cell))) > 80 {
					hasMultiLine = true
					break
				}
			}
			if hasMultiLine {
				break
			}
		}

		var fullLines []string
		if hasMultiLine {
			// Use multi-line capable table builder
			fullLines = m.buildFullTableMultiline(data, colWidths)
		} else {
			// Use regular table builder for better performance on simple tables
			fullLines = m.buildFullTable(data, colWidths)
		}

		// Cache the rendered lines
		m.cachedTableLines = fullLines

		// Debug: Log the last few lines to verify bottom border is included
		if len(fullLines) >= 3 {
			lastIdx := len(fullLines) - 1
			logger.DebugfToFile("Table", "Last 3 cached lines: [%d]=%q, [%d]=%q, [%d]=%q",
				lastIdx-2, stripAnsi(fullLines[lastIdx-2]),
				lastIdx-1, stripAnsi(fullLines[lastIdx-1]),
				lastIdx, stripAnsi(fullLines[lastIdx]))
		}

		// Calculate total table width
		if len(fullLines) > 0 {
			m.tableWidth = len(stripAnsi(fullLines[0]))
		}
	}

	// Apply horizontal scrolling to cached lines
	if m.horizontalOffset > 0 && m.cachedTableLines != nil {
		scrolledLines := applyHorizontalScrollWithANSI(m.cachedTableLines, m.horizontalOffset, m.tableViewport.Width)
		return strings.Join(scrolledLines, "\n")
	}

	// Return cached lines
	if m.cachedTableLines != nil {
		return strings.Join(m.cachedTableLines, "\n")
	}

	return ""
}

// isSameTableData checks if the table data has changed
func (m *MainModel) isSameTableData(data [][]string) bool {
	if m.lastTableData == nil || data == nil {
		return m.lastTableData == nil && data == nil
	}
	if len(m.lastTableData) != len(data) {
		return false
	}
	// For now, just check if it's the same length and first/last elements match
	// This is a heuristic that works because we typically reuse the same slice
	if len(m.lastTableData) > 0 && len(data) > 0 {
		// Check if first and last row are the same (pointer comparison)
		return &m.lastTableData[0] == &data[0]
	}
	return true
}


// buildFullTable builds the complete table with borders and formatting
func (m *MainModel) buildFullTable(data [][]string, colWidths []int) []string {
	var lines []string
	m.tableRowBoundaries = []int{} // Reset row boundaries

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
		// Record the starting line of this data row
		m.tableRowBoundaries = append(m.tableRowBoundaries, len(lines))

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

	// Add the bottom border line as a boundary so it's included when scrolling to bottom
	m.tableRowBoundaries = append(m.tableRowBoundaries, len(lines)-1)

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

		// Debug: Log viewport state
		logger.DebugfToFile("Table", "refreshTableView: TotalLineCount=%d, Height=%d, YOffset=%d, cachedLines=%d",
			m.tableViewport.TotalLineCount(), m.tableViewport.Height, currentYOffset, len(m.cachedTableLines))

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
