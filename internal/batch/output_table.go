package batch

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/ui"
)

// outputTable outputs data in ASCII table format
func (e *Executor) outputTable(data [][]string) error {
	if len(data) == 0 {
		return nil
	}

	output := ui.FormatASCIITable(data)
	fmt.Fprint(e.writer, output)
	return nil
}

// outputStreamingRows outputs rows during streaming
func (e *Executor) outputStreamingRows(rows [][]string, headers []string) error {
	if len(rows) == 0 {
		return nil
	}

	// For streaming, we just output the row content without headers or borders
	// since the header was already printed
	for _, row := range rows {
		// Format each row with proper spacing
		fmt.Fprintf(e.writer, "| ")
		for i, cell := range row {
			// Calculate column width based on header
			colWidth := len(headers[i])
			if colWidth < len(cell) {
				colWidth = len(cell)
			}
			fmt.Fprintf(e.writer, "%-*s | ", colWidth, cell)
		}
		fmt.Fprintln(e.writer)
	}

	return nil
}

// printTableBottom prints the bottom border of a table
func (e *Executor) printTableBottom(headers []string) {
	// Calculate the total width needed
	totalWidth := 1 // Start with initial '+'
	for _, header := range headers {
		totalWidth += len(header) + 3 // +3 for " | " or "-+-"
	}

	// Print bottom border
	fmt.Fprint(e.writer, "+")
	for i, header := range headers {
		for j := 0; j < len(header)+2; j++ {
			fmt.Fprint(e.writer, "-")
		}
		if i < len(headers)-1 {
			fmt.Fprint(e.writer, "+")
		}
	}
	fmt.Fprintln(e.writer, "+")
}

// printTraceData prints tracing information if available
func (e *Executor) printTraceData() {
	// Get trace data from the session
	traceData, headers, traceInfo, err := e.session.GetTraceData()
	if err != nil {
		fmt.Fprintf(e.writer, "Failed to get trace data: %v\n", err)
		return
	}

	if len(traceData) == 0 {
		return
	}

	// Print trace info
	fmt.Fprintln(e.writer, "\nTracing Information:")
	fmt.Fprintf(e.writer, "Duration: %v\n", traceInfo.Duration)
	fmt.Fprintf(e.writer, "Coordinator: %s\n", traceInfo.Coordinator)

	// Add headers to the beginning of data
	allData := [][]string{headers}
	allData = append(allData, traceData...)

	// Format and print trace data
	fmt.Fprintln(e.writer, "\nTrace Events:")
	output := ui.FormatASCIITable(allData)
	fmt.Fprint(e.writer, strings.TrimSpace(output))
	fmt.Fprintln(e.writer)
}