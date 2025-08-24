package ui

import (
	"bytes"
	"fmt"
)

// FormatExpandTable formats query results in expanded vertical format
func FormatExpandTable(data [][]string) string {
	if len(data) == 0 {
		return "No results"
	}
	
	if len(data) == 1 {
		// Only headers, no data rows
		return "No results"
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
	
	// Process each row
	for rowIdx, row := range rows {
		// Row header
		buf.WriteString(fmt.Sprintf("\n@ Row %d\n", rowIdx+1))
		
		// Separator line
		for i := 0; i < maxColWidth+1; i++ {
			buf.WriteString("-")
		}
		buf.WriteString("+")
		// Add dashes for the value column (use a reasonable width)
		for i := 0; i < 40; i++ {
			buf.WriteString("-")
		}
		buf.WriteString("\n")
		
		// Column name and value pairs
		for colIdx, value := range row {
			if colIdx < len(headers) {
				// Get the header (already includes (PK)/(C) if present)
				header := headers[colIdx]
				
				// Format: column_name | value
				buf.WriteString(" ")
				buf.WriteString(header)
				
				// Add padding to align the separator
				padding := maxColWidth - len(header)
				for i := 0; i < padding; i++ {
					buf.WriteString(" ")
				}
				
				buf.WriteString(" | ")
				buf.WriteString(value)
				buf.WriteString("\n")
			}
		}
	}
	
	// Add row count
	buf.WriteString("\n")
	if len(rows) == 1 {
		buf.WriteString(fmt.Sprintf("(%d row)\n", len(rows)))
	} else {
		buf.WriteString(fmt.Sprintf("(%d rows)\n", len(rows)))
	}
	
	return buf.String()
}