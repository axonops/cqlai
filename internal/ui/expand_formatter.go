package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FormatExpandTable formats query results in expanded vertical format.
func FormatExpandTable(data [][]string, styles *Styles) string {
	out, _ := FormatExpandTableWithBoundaries(data, styles)
	return out
}

// FormatExpandTableWithBoundaries formats query results in expanded vertical
// format and reports the line each record starts on.
//
// Paging snaps to those line numbers. In this layout a record is as tall as the
// table is wide - often ten or more lines - so boundaries taken from the boxed
// table renderer, where a record is a single line, cap scrolling near the row
// count and leave the rest of the result unreachable.
func FormatExpandTableWithBoundaries(data [][]string, styles *Styles) (string, []int) {
	if len(data) == 0 {
		return "No results", nil
	}

	if len(data) == 1 {
		// Only headers, no data rows
		return "No results", nil
	}

	headers := data[0]
	rows := data[1:]

	// Calculate the maximum column name width for alignment (including indicators)
	maxColWidth := 0
	for _, header := range headers {
		// Include (PK) and (C) indicators in width calculation
		if len(header) > maxColWidth {
			maxColWidth = len(header)
		}
	}

	var buf bytes.Buffer

	// Line the next write lands on, and where each record begins.
	line := 0
	boundaries := make([]int, 0, len(rows)+1)

	// Define styles for different parts
	rowHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	columnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87D7FF")) // Light blue for column names

	pipeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	// Process each row
	for rowIdx, row := range rows {
		// Blank line between records
		buf.WriteString("\n")
		line++

		// The record starts here, so this is what paging should land on
		boundaries = append(boundaries, line)

		// Row header with color
		buf.WriteString(rowHeaderStyle.Render(fmt.Sprintf("@ Row %d", rowIdx+1)))
		buf.WriteString("\n")
		line++

		// Separator line with muted color
		separator := strings.Repeat("-", maxColWidth+2) + "+" + strings.Repeat("-", 40)
		buf.WriteString(separatorStyle.Render(separator))
		buf.WriteString("\n")
		line++

		// Column name and value pairs
		for colIdx, value := range row {
			if colIdx < len(headers) {
				// Get the header (already includes (PK)/(C) if present)
				header := headers[colIdx]

				// Choose appropriate style for column based on type
				var colStyle lipgloss.Style
				switch {
				case strings.Contains(header, "(PK)"):
					colStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FFD787")). // Gold for primary key
						Bold(true)
				case strings.Contains(header, "(C)"):
					colStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#87FFD7")). // Aqua for clustering key
						Bold(true)
				default:
					colStyle = columnStyle
				}

				// Format: column_name | value
				buf.WriteString(" ")
				buf.WriteString(colStyle.Render(header))

				// Add padding to align the separator
				padding := maxColWidth - len(header)
				buf.WriteString(strings.Repeat(" ", padding))

				buf.WriteString(pipeStyle.Render(" | "))
				buf.WriteString(value)
				buf.WriteString("\n")
				line++
			}
		}
	}

	// Add row count with muted style
	rowCountStyle := lipgloss.NewStyle().
		Foreground(styles.Muted)

	buf.WriteString("\n")
	line++
	if len(rows) == 1 {
		buf.WriteString(rowCountStyle.Render(fmt.Sprintf("(%d row)", len(rows))))
	} else {
		buf.WriteString(rowCountStyle.Render(fmt.Sprintf("(%d rows)", len(rows))))
	}
	buf.WriteString("\n")

	// A final boundary on the last line, so paging can reach the end of the
	// last record rather than stopping at the line it starts on.
	boundaries = append(boundaries, line)

	return buf.String(), boundaries
}
