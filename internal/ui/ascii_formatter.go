package ui

import (
	"bytes"
	"fmt"
	"github.com/axonops/cqlai/internal/db"
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
		fmt.Fprintf(&buf, " %-*s |", columnWidths[i], h)
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

// FormatASCIITable formats query results as an ASCII table for display in the
// terminal, without a detail row under the column names.
//
// Batch mode uses this: its output is piped somewhere and read by something
// else as often as by a person, and a second header row is one more line to
// have to skip.
func FormatASCIITable(data [][]string) string {
	return FormatASCIITableWithTypes(data, nil)
}

// FormatASCIITableWithTypes draws the table with the key marker and the type on
// the row under the column names, the way the boxed table does.
//
// columnTypes may be nil or the wrong length, which is what a grid the shell
// made up rather than a result over a table looks like - DESCRIBE and the
// listings. Then there is no detail row, the same as there are no key markers.
func FormatASCIITableWithTypes(data [][]string, columnTypes []string) string {
	if len(data) == 0 {
		return "No results"
	}

	details := asciiHeaderDetails(data[0], columnTypes)

	// Calculate column widths based on actual content (including multi-line)
	columnWidths := CalculateColumnWidths(data)
	// With a detail row the name row is shorter than the header string it came
	// from - the marker has moved down - so the widths are worked out again
	// from what will actually be drawn.
	for i, detail := range details {
		columnWidths[i] = max(len([]rune(db.StripKeyMarker(data[0][i]))), len([]rune(detail)))
		for _, row := range data[1:] {
			if i < len(row) {
				for _, line := range strings.Split(row[i], "\n") {
					columnWidths[i] = max(columnWidths[i], len([]rune(line)))
				}
			}
		}
	}

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
	drawHeaderRow := func(cells []string) {
		buf.WriteString("|")
		for i, cell := range cells {
			buf.WriteString(" ")
			buf.WriteString(cell)
			for j := len([]rune(cell)); j < columnWidths[i]; j++ {
				buf.WriteString(" ")
			}
			buf.WriteString(" |")
		}
		buf.WriteString("\n")
	}

	if details == nil {
		drawHeaderRow(data[0])
	} else {
		drawHeaderRow(db.StripKeyMarkers(data[0]))
		drawHeaderRow(details)
	}

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
		fmt.Fprintf(&buf, "\n(%d row)\n", rowCount)
	} else {
		fmt.Fprintf(&buf, "\n(%d rows)\n", rowCount)
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

// asciiHeaderDetails is the detail row for an ASCII table, or nil when there is
// not one to draw.
func asciiHeaderDetails(headers, columnTypes []string) []string {
	if len(columnTypes) != len(headers) {
		return nil
	}

	details := make([]string, len(headers))
	for i, header := range headers {
		details[i] = headerDetail(header, columnTypes[i])
	}
	return details
}
