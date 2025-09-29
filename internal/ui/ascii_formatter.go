package ui

import (
	"bytes"
	"fmt"
	"strings"
)

// FormatASCIITableHeader formats just the header part of an ASCII table
func FormatASCIITableHeader(headers [][]string) string {
	if len(headers) == 0 || len(headers[0]) == 0 {
		return ""
	}
	
	header := headers[0]
	
	// Calculate column widths (using rune count for proper Unicode handling)
	columnWidths := make([]int, len(header))
	for i, h := range header {
		columnWidths[i] = len([]rune(h)) // Count runes, not bytes
	}
	
	var buf bytes.Buffer
	
	// Top border
	buf.WriteString("+")
	for _, width := range columnWidths {
		buf.WriteString(strings.Repeat("-", width+2))
		buf.WriteString("+")
	}
	buf.WriteString("\n")
	
	// Header row
	buf.WriteString("|")
	for i, h := range header {
		buf.WriteString(fmt.Sprintf(" %-*s |", columnWidths[i], h))
	}
	buf.WriteString("\n")
	
	// Header separator
	buf.WriteString("+")
	for _, width := range columnWidths {
		buf.WriteString(strings.Repeat("-", width+2))
		buf.WriteString("+")
	}
	buf.WriteString("\n")
	
	return buf.String()
}

// FormatASCIITable formats query results as an ASCII table for display in the terminal
func FormatASCIITable(data [][]string) string {
	if len(data) == 0 {
		return "No results"
	}

	// Calculate column widths based on actual content (including multi-line)
	columnWidths := CalculateColumnWidths(data)

	var buf bytes.Buffer

	// Helper function to draw separator line
	drawSeparator := func(leftChar, midChar, rightChar string) {
		buf.WriteString(leftChar)
		for i, width := range columnWidths {
			for j := 0; j < width+2; j++ {
				buf.WriteString("-")
			}
			if i < len(columnWidths)-1 {
				buf.WriteString(midChar)
			}
		}
		buf.WriteString(rightChar)
		buf.WriteString("\n")
	}

	// Draw top border
	drawSeparator("+", "+", "+")

	// Draw header
	buf.WriteString("|")
	for i, header := range data[0] {
		buf.WriteString(" ")
		headerRunes := []rune(header)
		buf.WriteString(header)
		// Add padding
		for j := len(headerRunes); j < columnWidths[i]; j++ {
			buf.WriteString(" ")
		}
		buf.WriteString(" |")
	}
	buf.WriteString("\n")

	// Draw separator after header
	drawSeparator("+", "+", "+")

	// Draw data rows with multi-line support
	for _, row := range data[1:] {
		// First, render the first line of each cell
		buf.WriteString("|")
		for i, cell := range row {
			buf.WriteString(" ")
			lines := strings.Split(cell, "\n")
			firstLine := lines[0]
			buf.WriteString(firstLine)
			// Add padding (using rune count)
			cellWidth := len([]rune(firstLine))
			for j := cellWidth; j < columnWidths[i]; j++ {
				buf.WriteString(" ")
			}
			buf.WriteString(" |")
		}
		buf.WriteString("\n")

		// Handle additional lines in multi-line cells
		hasMoreLines := true
		lineIndex := 1
		for hasMoreLines {
			hasMoreLines = false
			extraLine := "|"
			for i, cell := range row {
				lines := strings.Split(cell, "\n")
				if lineIndex < len(lines) {
					hasMoreLines = true
					extraLine += " "
					extraLine += lines[lineIndex]
					cellWidth := len([]rune(lines[lineIndex]))
					for j := cellWidth; j < columnWidths[i]; j++ {
						extraLine += " "
					}
					extraLine += " |"
				} else {
					// Empty cell for this line
					extraLine += " "
					for j := 0; j < columnWidths[i]; j++ {
						extraLine += " "
					}
					extraLine += " |"
				}
			}
			if hasMoreLines {
				buf.WriteString(extraLine)
				buf.WriteString("\n")
			}
			lineIndex++
		}
	}

	// Draw bottom border
	drawSeparator("+", "+", "+")

	// Add row count
	rowCount := len(data) - 1
	if rowCount == 1 {
		buf.WriteString(fmt.Sprintf("\n(%d row)\n", rowCount))
	} else {
		buf.WriteString(fmt.Sprintf("\n(%d rows)\n", rowCount))
	}

	return buf.String()
}

// FormatASCIITableRowsOnly formats only the data rows (no headers, no borders)
// Used for streaming subsequent batches
func FormatASCIITableRowsOnly(data [][]string) string {
	return FormatASCIITableRowsOnlyWithWidths(data, nil)
}

// FormatASCIITableRowsOnlyWithWidths formats only the data rows with specified column widths
func FormatASCIITableRowsOnlyWithWidths(data [][]string, columnWidths []int) string {
	if len(data) <= 1 {
		return "" // No data rows to output
	}

	// If no column widths provided, calculate them
	if columnWidths == nil {
		columnWidths = CalculateColumnWidths(data)
	}

	var buf bytes.Buffer

	// Draw data rows (skip header at index 0)
	for _, row := range data[1:] {
		buf.WriteString("|")
		for i, cell := range row {
			buf.WriteString(" ")

			// Handle multi-line cells
			lines := strings.Split(cell, "\n")
			if len(lines) > 1 {
				// For multi-line cells, format each line separately
				maxLineWidth := 0
				for _, line := range lines {
					lineWidth := len([]rune(line))
					if lineWidth > maxLineWidth {
						maxLineWidth = lineWidth
					}
				}

				// Use the first line for this row, pad to column width
				firstLine := lines[0]
				buf.WriteString(firstLine)
				cellWidth := len([]rune(firstLine))
				for j := cellWidth; j < columnWidths[i]; j++ {
					buf.WriteString(" ")
				}
			} else {
				// Single line cell
				cellRunes := []rune(cell)
				buf.WriteString(cell)
				// Add padding (using rune count)
				cellWidth := len(cellRunes)
				for j := cellWidth; j < columnWidths[i]; j++ {
					buf.WriteString(" ")
				}
			}
			buf.WriteString(" |")
		}
		buf.WriteString("\n")

		// Handle additional lines in multi-line cells
		hasMoreLines := true
		lineIndex := 1
		for hasMoreLines {
			hasMoreLines = false
			extraLine := "|"
			for i, cell := range row {
				lines := strings.Split(cell, "\n")
				if lineIndex < len(lines) {
					hasMoreLines = true
					extraLine += " "
					extraLine += lines[lineIndex]
					cellWidth := len([]rune(lines[lineIndex]))
					for j := cellWidth; j < columnWidths[i]; j++ {
						extraLine += " "
					}
					extraLine += " |"
				} else {
					// Empty cell for this line
					extraLine += " "
					for j := 0; j < columnWidths[i]; j++ {
						extraLine += " "
					}
					extraLine += " |"
				}
			}
			if hasMoreLines {
				buf.WriteString(extraLine)
				buf.WriteString("\n")
			}
			lineIndex++
		}
	}

	return buf.String()
}

// CalculateColumnWidths calculates the maximum width for each column in the data
func CalculateColumnWidths(data [][]string) []int {
	if len(data) == 0 {
		return []int{}
	}

	columnWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			// For multi-line cells, check each line's width
			lines := strings.Split(cell, "\n")
			for _, line := range lines {
				cellWidth := len([]rune(line))
				if cellWidth > columnWidths[i] {
					columnWidths[i] = cellWidth
				}
			}
		}
	}
	return columnWidths
}

// FormatASCIITableBottom draws just the bottom border based on data
func FormatASCIITableBottom(data [][]string) string {
	return FormatASCIITableBottomWithWidths(data, nil)
}

// FormatASCIITableBottomWithWidths draws just the bottom border with specified column widths
func FormatASCIITableBottomWithWidths(data [][]string, columnWidths []int) string {
	if len(data) == 0 {
		return ""
	}

	// If no column widths provided, calculate them
	if columnWidths == nil {
		columnWidths = CalculateColumnWidths(data)
	}

	var buf bytes.Buffer

	// Draw bottom border
	buf.WriteString("+")
	for i, width := range columnWidths {
		for j := 0; j < width+2; j++ {
			buf.WriteString("-")
		}
		if i < len(columnWidths)-1 {
			buf.WriteString("+")
		}
	}
	buf.WriteString("+\n")

	return buf.String()
}