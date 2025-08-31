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
	
	// Calculate column widths
	columnWidths := make([]int, len(header))
	for i, h := range header {
		columnWidths[i] = len(h)
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
	
	// Calculate column widths
	columnWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
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
	buf.WriteString("|")
	for i, header := range data[0] {
		buf.WriteString(" ")
		buf.WriteString(header)
		// Add padding
		for j := len(header); j < columnWidths[i]; j++ {
			buf.WriteString(" ")
		}
		buf.WriteString(" |")
	}
	buf.WriteString("\n")
	
	// Draw separator after header
	drawSeparator("+", "+", "+")
	
	// Draw data rows
	for _, row := range data[1:] {
		buf.WriteString("|")
		for i, cell := range row {
			buf.WriteString(" ")
			buf.WriteString(cell)
			// Add padding
			for j := len(cell); j < columnWidths[i]; j++ {
				buf.WriteString(" ")
			}
			buf.WriteString(" |")
		}
		buf.WriteString("\n")
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