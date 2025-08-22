package router

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/axonops/cqlai/internal/db"
)

// MetaCommandHandler handles non-CQL meta commands
type MetaCommandHandler struct {
	session      *db.Session
	expandMode   bool
	captureFile  string
	captureOutput *os.File
	captureFormat string // "text", "json", or "csv"
	csvWriter    *csv.Writer
}

// NewMetaCommandHandler creates a new meta command handler
func NewMetaCommandHandler(session *db.Session) *MetaCommandHandler {
	return &MetaCommandHandler{
		session:       session,
		expandMode:    false,
		captureFormat: "text",
	}
}

// HandleMetaCommand processes meta commands that aren't CQL
func (h *MetaCommandHandler) HandleMetaCommand(command string) interface{} {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	parts := strings.Fields(upperCommand)
	
	if len(parts) == 0 {
		return ""
	}
	
	switch parts[0] {
	case "CONSISTENCY":
		return h.handleConsistency(command)
	case "SHOW":
		return h.handleShow(command)
	case "TRACING":
		return h.handleTracing(command)
	case "PAGING":
		return h.handlePaging(command)
	case "EXPAND":
		return h.handleExpand(command)
	case "SOURCE":
		return h.handleSource(command)
	case "CAPTURE":
		return h.handleCapture(command)
	case "HELP":
		return h.handleHelp()
	default:
		return fmt.Sprintf("Unknown meta command: %s", parts[0])
	}
}

// handleConsistency handles CONSISTENCY command
func (h *MetaCommandHandler) handleConsistency(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))
	
	if len(parts) == 1 {
		// Show current consistency
		return fmt.Sprintf("Current consistency level: %s", h.session.Consistency())
	}
	
	if len(parts) == 2 {
		// Set consistency level
		level := parts[1]
		if err := h.session.SetConsistency(level); err != nil {
			return err
		}
		return fmt.Sprintf("Consistency level set to %s", level)
	}
	
	return "Usage: CONSISTENCY [level]\nValid levels: ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE"
}

// handleShow handles SHOW commands
func (h *MetaCommandHandler) handleShow(command string) interface{} {
	upperCommand := strings.ToUpper(command)
	
	if strings.Contains(upperCommand, "VERSION") {
		// Show Cassandra version
		iter := h.session.Query("SELECT release_version FROM system.local").Iter()
		var version string
		if iter.Scan(&version) {
			iter.Close()
			return fmt.Sprintf("Cassandra version: %s", version)
		}
		iter.Close()
		return "Unable to get Cassandra version"
	}
	
	if strings.Contains(upperCommand, "HOST") {
		// Show current host connection
		iter := h.session.Query("SELECT rpc_address, data_center, rack FROM system.local").Iter()
		var host, datacenter, rack string
		if iter.Scan(&host, &datacenter, &rack) {
			iter.Close()
			result := fmt.Sprintf("Connected to: %s\n", host)
			result += fmt.Sprintf("Datacenter: %s\n", datacenter)
			result += fmt.Sprintf("Rack: %s", rack)
			return result
		}
		iter.Close()
		return "Unable to get host information"
	}
	
	if strings.Contains(upperCommand, "SESSION") {
		// Show session information
		result := fmt.Sprintf("Current keyspace: %s\n", h.session.CurrentKeyspace())
		result += fmt.Sprintf("Consistency: %s\n", h.session.Consistency())
		result += fmt.Sprintf("Page size: %d\n", h.session.PageSize())
		result += fmt.Sprintf("Tracing: %v\n", h.session.Tracing())
		result += fmt.Sprintf("Expand mode: %v", h.expandMode)
		return result
	}
	
	return "Usage: SHOW VERSION | SHOW HOST | SHOW SESSION"
}

// handleTracing handles TRACING command
func (h *MetaCommandHandler) handleTracing(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))
	
	if len(parts) < 2 {
		if h.session.Tracing() {
			return "Tracing is currently ON"
		}
		return "Tracing is currently OFF"
	}
	
	switch parts[1] {
	case "ON":
		h.session.SetTracing(true)
		return "Tracing turned ON"
	case "OFF":
		h.session.SetTracing(false)
		return "Tracing turned OFF"
	default:
		return "Usage: TRACING ON | OFF"
	}
}

// handlePaging handles PAGING command
func (h *MetaCommandHandler) handlePaging(command string) interface{} {
	parts := strings.Fields(command)
	
	if len(parts) == 1 {
		// Show current page size
		return fmt.Sprintf("Current page size: %d", h.session.PageSize())
	}
	
	if len(parts) == 2 {
		// Try to parse the page size
		var pageSize int
		if _, err := fmt.Sscanf(parts[1], "%d", &pageSize); err != nil {
			return fmt.Sprintf("Invalid page size: %s", parts[1])
		}
		
		if pageSize < 1 {
			return "Page size must be at least 1"
		}
		
		h.session.SetPageSize(pageSize)
		return fmt.Sprintf("Page size set to %d", pageSize)
	}
	
	if len(parts) == 3 && strings.ToUpper(parts[1]) == "OFF" {
		// Disable paging (set to very large number)
		h.session.SetPageSize(10000)
		return "Paging disabled (set to 10000)"
	}
	
	return "Usage: PAGING [size] | PAGING OFF"
}

// handleExpand handles EXPAND command for vertical output
func (h *MetaCommandHandler) handleExpand(command string) interface{} {
	parts := strings.Fields(strings.ToUpper(command))
	
	if len(parts) == 1 {
		if h.expandMode {
			return "Expand mode is currently ON (vertical output)"
		}
		return "Expand mode is currently OFF (table output)"
	}
	
	switch parts[1] {
	case "ON":
		h.expandMode = true
		return "Expand mode turned ON - results will be shown vertically"
	case "OFF":
		h.expandMode = false
		return "Expand mode turned OFF - results will be shown as tables"
	default:
		return "Usage: EXPAND ON | OFF"
	}
}

// handleSource handles SOURCE command to execute CQL from file
func (h *MetaCommandHandler) handleSource(command string) interface{} {
	parts := strings.Fields(command)
	
	if len(parts) < 2 {
		return "Usage: SOURCE 'filename'"
	}
	
	// Extract filename (remove quotes if present)
	filename := strings.Join(parts[1:], " ")
	filename = strings.Trim(filename, "'\"")
	
	// Expand home directory if needed
	if strings.HasPrefix(filename, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			filename = filepath.Join(home, filename[2:])
		}
	}
	
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	
	// Split into statements (simple split by semicolon)
	statements := strings.Split(string(content), ";")
	
	results := []string{}
	successCount := 0
	errorCount := 0
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		// Execute the statement
		result := ProcessCommand(stmt + ";", h.session)
		
		// Check if it's an error
		if err, ok := result.(error); ok {
			errorCount++
			results = append(results, fmt.Sprintf("Error in statement: %v", err))
		} else {
			successCount++
		}
	}
	
	summary := fmt.Sprintf("\nExecuted %d statements from %s", successCount+errorCount, filename)
	if errorCount > 0 {
		summary += fmt.Sprintf(" (%d successful, %d failed)", successCount, errorCount)
	}
	
	if len(results) > 0 {
		return strings.Join(results, "\n") + "\n" + summary
	}
	
	return summary
}

// handleCapture handles CAPTURE command to save output to file
func (h *MetaCommandHandler) handleCapture(command string) interface{} {
	parts := strings.Fields(command)
	
	if len(parts) == 1 {
		// Show current capture status
		if h.captureFile != "" {
			return fmt.Sprintf("Currently capturing to: %s (format: %s)", h.captureFile, h.captureFormat)
		}
		return "Not currently capturing output"
	}
	
	if len(parts) >= 2 && strings.ToUpper(parts[1]) == "OFF" {
		// Stop capturing
		if h.captureOutput != nil {
			// If JSON format, properly close the array
			if h.captureFormat == "json" {
				// Seek back to remove trailing comma if exists
				info, _ := h.captureOutput.Stat()
				if info.Size() > 2 {
					h.captureOutput.Seek(-2, 2) // Go to end minus 2 chars
					h.captureOutput.WriteString("\n]\n")
				} else {
					h.captureOutput.WriteString("]\n")
				}
			} else if h.captureFormat == "csv" && h.csvWriter != nil {
				// Flush CSV writer
				h.csvWriter.Flush()
				h.csvWriter = nil
			}
			
			h.captureOutput.Close()
			h.captureOutput = nil
			result := fmt.Sprintf("Stopped capturing to %s", h.captureFile)
			h.captureFile = ""
			h.captureFormat = "text"
			return result
		}
		return "Not currently capturing"
	}
	
	// Parse capture command: CAPTURE [JSON|CSV] 'filename'
	format := "text"
	filenameStart := 1
	
	if len(parts) >= 2 {
		upperFormat := strings.ToUpper(parts[1])
		if upperFormat == "JSON" {
			format = "json"
			filenameStart = 2
		} else if upperFormat == "CSV" {
			format = "csv"
			filenameStart = 2
		}
	}
	
	if len(parts) <= filenameStart {
		return "Usage: CAPTURE [JSON|CSV] 'filename' | CAPTURE OFF"
	}
	
	// Get filename
	filename := strings.Join(parts[filenameStart:], " ")
	filename = strings.Trim(filename, "'\"")
	
	// Expand home directory if needed
	if strings.HasPrefix(filename, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			filename = filepath.Join(home, filename[2:])
		}
	}
	
	// Add appropriate extension if not provided
	if format == "json" && !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	} else if format == "csv" && !strings.HasSuffix(filename, ".csv") {
		filename = filename + ".csv"
	}
	
	// Close existing capture file if any
	if h.captureOutput != nil {
		if h.csvWriter != nil {
			h.csvWriter.Flush()
			h.csvWriter = nil
		}
		h.captureOutput.Close()
	}
	
	// Open new capture file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Sprintf("Error opening capture file: %v", err)
	}
	
	h.captureOutput = file
	h.captureFile = filename
	h.captureFormat = format
	
	if format == "json" {
		// Write opening bracket for JSON array
		file.WriteString("[\n")
	} else if format == "csv" {
		// Create CSV writer
		h.csvWriter = csv.NewWriter(file)
	}
	
	return fmt.Sprintf("Now capturing query output to %s (format: %s). Use 'CAPTURE OFF' to stop.", filename, format)
}

// handleHelp handles HELP command
func (h *MetaCommandHandler) handleHelp() interface{} {
	help := [][]string{
		{"Command", "Description"},
		{"--------", "-----------"},
		{"DESCRIBE KEYSPACES", "List all keyspaces"},
		{"DESCRIBE TABLES", "List tables in current keyspace"},
		{"DESCRIBE TABLE <name>", "Show table details"},
		{"DESCRIBE CLUSTER", "Show cluster information"},
		{"USE <keyspace>", "Switch to a keyspace"},
		{"CONSISTENCY [level]", "Show/set consistency level"},
		{"SHOW VERSION", "Show Cassandra version"},
		{"SHOW HOST", "Show connection details"},
		{"SHOW SESSION", "Show session information"},
		{"TRACING ON|OFF", "Enable/disable query tracing"},
		{"PAGING [size]", "Set result page size"},
		{"EXPAND ON|OFF", "Toggle vertical output"},
		{"SOURCE 'file'", "Execute CQL from file"},
		{"CAPTURE [JSON] 'file'", "Capture output to file (optionally as JSON)"},
		{"CAPTURE OFF", "Stop capturing output"},
		{"CLEAR", "Clear the screen"},
		{"EXIT/QUIT", "Exit cqlai"},
	}
	
	return help
}

// IsExpandMode returns whether expand mode is on
func (h *MetaCommandHandler) IsExpandMode() bool {
	return h.expandMode
}

// GetCaptureFile returns the current capture file if any
func (h *MetaCommandHandler) GetCaptureFile() *os.File {
	return h.captureOutput
}

// GetCaptureFormat returns the current capture format ("text" or "json")
func (h *MetaCommandHandler) GetCaptureFormat() string {
	return h.captureFormat
}

// IsCapturing returns true if currently capturing output
func (h *MetaCommandHandler) IsCapturing() bool {
	return h.captureOutput != nil
}

// WriteToCapture writes data to the capture file if active
func (h *MetaCommandHandler) WriteToCapture(data string) error {
	if h.captureOutput != nil {
		_, err := h.captureOutput.WriteString(data)
		return err
	}
	return nil
}

// FormatResultAsJSON formats query results as JSON
func FormatResultAsJSON(headers []string, rows [][]string) (string, error) {
	type QueryResult struct {
		Query   string                   `json:"query,omitempty"`
		Columns []string                 `json:"columns"`
		Rows    []map[string]interface{} `json:"rows"`
		Count   int                      `json:"row_count"`
	}
	
	result := QueryResult{
		Columns: headers,
		Rows:    make([]map[string]interface{}, 0, len(rows)),
		Count:   len(rows),
	}
	
	for _, row := range rows {
		rowMap := make(map[string]interface{})
		for i, col := range headers {
			if i < len(row) {
				// Try to parse as number or boolean
				value := row[i]
				if value == "null" {
					rowMap[col] = nil
				} else if value == "true" || value == "false" {
					rowMap[col] = value == "true"
				} else if num, err := json.Number(value).Float64(); err == nil {
					rowMap[col] = num
				} else {
					rowMap[col] = value
				}
			}
		}
		result.Rows = append(result.Rows, rowMap)
	}
	
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(jsonBytes), nil
}

// WriteCaptureResult writes query results to the capture file
func (h *MetaCommandHandler) WriteCaptureResult(command string, headers []string, rows [][]string) error {
	if h.captureOutput == nil {
		return nil
	}
	
	switch h.captureFormat {
	case "csv":
		// Write headers as first row (with query command as comment)
		// Write comment with the query
		h.csvWriter.Write([]string{"# Query: " + command})
		
		// Write headers
		if err := h.csvWriter.Write(headers); err != nil {
			return err
		}
		
		// Write data rows
		for _, row := range rows {
			if err := h.csvWriter.Write(row); err != nil {
				return err
			}
		}
		
		// Add empty row to separate queries
		h.csvWriter.Write([]string{})
		
		// Flush to ensure data is written
		h.csvWriter.Flush()
		return h.csvWriter.Error()
		
	case "json":
		// Format as JSON
		type QueryResult struct {
			Query   string                   `json:"query"`
			Columns []string                 `json:"columns"`
			Rows    []map[string]interface{} `json:"rows"`
			Count   int                      `json:"row_count"`
		}
		
		result := QueryResult{
			Query:   command,
			Columns: headers,
			Rows:    make([]map[string]interface{}, 0, len(rows)),
			Count:   len(rows),
		}
		
		for _, row := range rows {
			rowMap := make(map[string]interface{})
			for i, col := range headers {
				if i < len(row) {
					// Try to parse as number or boolean
					value := row[i]
					if value == "null" {
						rowMap[col] = nil
					} else if value == "true" || value == "false" {
						rowMap[col] = value == "true"
					} else if num, err := json.Number(value).Float64(); err == nil {
						rowMap[col] = num
					} else {
						rowMap[col] = value
					}
				}
			}
			result.Rows = append(result.Rows, rowMap)
		}
		
		jsonBytes, err := json.MarshalIndent(result, "  ", "  ")
		if err != nil {
			return err
		}
		
		// Check if this is the first entry
		info, _ := h.captureOutput.Stat()
		if info.Size() > 2 { // More than just "[\n"
			h.captureOutput.WriteString(",\n")
		}
		h.captureOutput.WriteString("  ")
		h.captureOutput.Write(jsonBytes)
		
	default:
		// Text format - write the command and a simple table representation
		h.captureOutput.WriteString(fmt.Sprintf("\n> %s\n", command))
		h.captureOutput.WriteString(strings.Repeat("-", 50) + "\n")
		
		// Write headers
		h.captureOutput.WriteString(strings.Join(headers, "\t") + "\n")
		
		// Write rows
		for _, row := range rows {
			h.captureOutput.WriteString(strings.Join(row, "\t") + "\n")
		}
		
		h.captureOutput.WriteString(fmt.Sprintf("\n(%d rows)\n\n", len(rows)))
	}
	
	return nil
}

// Close closes any open resources
func (h *MetaCommandHandler) Close() {
	if h.captureOutput != nil {
		// If JSON format, close the array
		if h.captureFormat == "json" {
			// Remove trailing comma and newline if present, then close array
			h.captureOutput.Seek(-2, 1) // Go back 2 chars (comma and newline)
			h.captureOutput.WriteString("\n]\n")
		} else if h.captureFormat == "csv" && h.csvWriter != nil {
			// Flush CSV writer
			h.csvWriter.Flush()
			h.csvWriter = nil
		}
		h.captureOutput.Close()
		h.captureOutput = nil
		h.captureFile = ""
		h.captureFormat = "text"
	}
}