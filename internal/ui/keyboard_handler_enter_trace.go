package ui

import (
	"fmt"
	"strings"
	"time"
)

// captureTraceData captures trace data if tracing is enabled
func (m *MainModel) captureTraceData(command string) {
	upperCmd := strings.ToUpper(strings.TrimSpace(command))
	if m.session != nil && m.session.Tracing() && 
	   (strings.HasPrefix(upperCmd, "SELECT") || 
	    strings.HasPrefix(upperCmd, "LIST") || 
	    strings.HasPrefix(upperCmd, "DESCRIBE") ||
	    strings.HasPrefix(upperCmd, "DESC")) {
		// Give Cassandra a moment to write trace data
		time.Sleep(50 * time.Millisecond)
		
		// Retrieve trace data
		traceData, traceHeaders, traceInfo, err := m.session.GetTraceData()
		if err == nil && len(traceData) > 0 {
			// Add summary info as a header to the trace content
			summaryLine := ""
			if traceInfo != nil {
				summaryLine = fmt.Sprintf("Trace Session - Coordinator: %s | Total Duration: %d μs\n",
					traceInfo.Coordinator, traceInfo.Duration)
			}
			
			// Combine headers and data into a single table structure
			fullTraceData := make([][]string, 0, len(traceData)+1)
			fullTraceData = append(fullTraceData, traceHeaders)
			fullTraceData = append(fullTraceData, traceData...)
			
			// Store trace data for refreshing
			m.traceData = fullTraceData
			m.traceHeaders = traceHeaders
			m.traceInfo = traceInfo
			m.hasTrace = true
			m.traceHorizontalOffset = 0 // Reset horizontal scroll
			
			// Use the existing formatTableForViewport method temporarily storing the offset
			originalOffset := m.horizontalOffset
			originalData := m.lastTableData
			originalWidth := m.tableWidth
			originalHeaders := m.tableHeaders
			originalColWidths := m.columnWidths
			
			// Set trace data temporarily
			m.horizontalOffset = m.traceHorizontalOffset
			m.lastTableData = fullTraceData
			
			// Format using existing table renderer
			traceTable := m.formatTableForViewport(fullTraceData)
			
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
			m.traceViewport.GotoTop()
		}
	}
}