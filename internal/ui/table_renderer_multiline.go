package ui

import (
	"strings"
)

// splitCellIntoLines splits a cell's content into multiple lines if it contains newlines
// or if it exceeds the maximum width
func splitCellIntoLines(cell string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 80 // Default reasonable width
	}

	// First split by newlines
	lines := strings.Split(cell, "\n")
	
	var result []string
	for _, line := range lines {
		// Further split long lines that exceed maxWidth
		if len([]rune(line)) <= maxWidth {
			result = append(result, line)
		} else {
			// Word wrap long lines
			wrapped := wrapLine(line, maxWidth)
			result = append(result, wrapped...)
		}
	}
	
	if len(result) == 0 {
		result = []string{""} // Ensure at least one line
	}
	
	return result
}

// wrapLine wraps a single line of text to fit within maxWidth
func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{line}
	}
	
	var result []string
	runes := []rune(line)
	
	for len(runes) > 0 {
		if len(runes) <= maxWidth {
			result = append(result, string(runes))
			break
		}
		
		// Find a good break point (prefer spaces)
		breakPoint := maxWidth
		for i := maxWidth - 1; i > maxWidth*2/3; i-- {
			if runes[i] == ' ' {
				breakPoint = i + 1 // Include the space at the end of the line
				break
			}
		}
		
		// Take up to the break point
		result = append(result, strings.TrimRight(string(runes[:breakPoint]), " "))
		runes = runes[breakPoint:]
		
		// Skip leading spaces on the next line
		for len(runes) > 0 && runes[0] == ' ' {
			runes = runes[1:]
		}
	}
	
	return result
}

// buildFullTableMultiline builds the complete table with borders and formatting, handling multi-line cells
func (m *MainModel) buildFullTableMultiline(data [][]string, colWidths []int) []string {
	if len(data) == 0 {
		return nil
	}

	var lines []string
	m.tableRowBoundaries = []int{} // Reset row boundaries

	// Determine max width for columns with very long content
	maxColWidth := 80 // Maximum width for any column to prevent excessive wrapping
	adjustedWidths := make([]int, len(colWidths))
	for i, width := range colWidths {
		if width > maxColWidth {
			adjustedWidths[i] = maxColWidth
		} else {
			adjustedWidths[i] = width
		}
	}

	// Top border
	topBorder := "┌"
	for i, width := range adjustedWidths {
		topBorder += strings.Repeat("─", width+2)
		if i < len(adjustedWidths)-1 {
			topBorder += "┬"
		}
	}
	topBorder += "┐"
	lines = append(lines, topBorder)

	// Process each row
	for rowIdx, row := range data {
		// Skip header row for boundary tracking (we only care about data rows)
		if rowIdx > 0 {
			// Record the starting line of this data row
			m.tableRowBoundaries = append(m.tableRowBoundaries, len(lines))
		}

		// Split each cell into lines
		cellLines := make([][]string, len(row))
		maxLines := 1

		for colIdx, cell := range row {
			if colIdx >= len(adjustedWidths) {
				continue
			}

			// Split the cell content into lines
			cellLines[colIdx] = splitCellIntoLines(stripAnsi(cell), adjustedWidths[colIdx])
			if len(cellLines[colIdx]) > maxLines {
				maxLines = len(cellLines[colIdx])
			}
		}

		// Render each line of the row
		for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
			line := "│"
			for colIdx := range row {
				if colIdx >= len(adjustedWidths) {
					continue
				}

				cellContent := ""
				if lineIdx < len(cellLines[colIdx]) {
					cellContent = cellLines[colIdx][lineIdx]
				}

				// Apply styling for headers
				if rowIdx == 0 {
					cellContent = m.styles.AccentText.Bold(true).Render(cellContent) + "\x1b[0m"
				}

				// Calculate padding
				plainContent := stripAnsi(cellContent)
				padding := adjustedWidths[colIdx] - len([]rune(plainContent))
				if padding < 0 {
					padding = 0
				}

				line += " " + cellContent + strings.Repeat(" ", padding) + " │"
			}
			lines = append(lines, line)
		}

		// Add separator after header row
		if rowIdx == 0 && len(data) > 1 {
			separator := "├"
			for i, width := range adjustedWidths {
				separator += strings.Repeat("─", width+2)
				if i < len(adjustedWidths)-1 {
					separator += "┼"
				}
			}
			separator += "┤"
			lines = append(lines, separator)
		}
	}

	// Bottom border
	bottomBorder := "└"
	for i, width := range adjustedWidths {
		bottomBorder += strings.Repeat("─", width+2)
		if i < len(adjustedWidths)-1 {
			bottomBorder += "┴"
		}
	}
	bottomBorder += "┘"
	lines = append(lines, bottomBorder)

	// Add the bottom border line as a boundary so it's included when scrolling to bottom
	m.tableRowBoundaries = append(m.tableRowBoundaries, len(lines)-1)

	return lines
}