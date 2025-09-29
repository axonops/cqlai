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