package router

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/logger"
)

// SaveCommand represents a parsed SAVE command
type SaveCommand struct {
	Interactive bool                   // true for "SAVE" without arguments
	Filename    string                 // target filename
	Format      string                 // CSV, JSON, or ASCII
	Options     map[string]interface{} // additional options like header=false, pretty=true
}

// SaveModalTrigger is a message to trigger the save modal in the UI
type SaveModalTrigger struct{}

// ParseSaveCommand parses SAVE command without ANTLR (exported for testing)
func ParseSaveCommand(input string) (*SaveCommand, error) {
	input = strings.TrimSpace(input)
	upperInput := strings.ToUpper(input)

	logger.DebugfToFile("ParseSaveCommand", "Input: '%s', upperInput: '%s'", input, upperInput)

	// Check for simple "SAVE" (interactive mode)
	if upperInput == "SAVE" {
		logger.DebugToFile("ParseSaveCommand", "Interactive mode triggered")
		return &SaveCommand{Interactive: true}, nil
	}

	// Parse "SAVE TO 'filename' [AS format] [WITH options]"
	if !strings.HasPrefix(upperInput, "SAVE TO ") {
		return nil, fmt.Errorf("invalid SAVE syntax. Use: SAVE TO 'filename' [AS format] [WITH options]")
	}

	// Extract the part after "SAVE TO "
	remainder := strings.TrimSpace(input[8:])

	// Parse filename (handle quoted strings)
	filename, remainder, err := parseQuotedString(remainder)
	if err != nil {
		return nil, fmt.Errorf("invalid filename: %v", err)
	}

	cmd := &SaveCommand{
		Filename: filename,
		Options:  make(map[string]interface{}),
	}

	// Parse optional AS clause
	remainder = strings.TrimSpace(remainder)
	upperRemainder := strings.ToUpper(remainder)
	if strings.HasPrefix(upperRemainder, "AS ") {
		parts := strings.Fields(remainder[3:])
		if len(parts) > 0 {
			cmd.Format = strings.ToUpper(parts[0])
			remainder = strings.TrimSpace(remainder[3+len(parts[0]):])
		}
	}

	// Parse optional WITH clause
	upperRemainder = strings.ToUpper(remainder)
	if strings.HasPrefix(upperRemainder, "WITH ") {
		optionsStr := strings.TrimSpace(remainder[5:])
		cmd.Options = parseSaveOptions(optionsStr)
	}

	// Auto-detect format from extension if not specified
	if cmd.Format == "" {
		cmd.Format = detectFormatFromExtension(cmd.Filename)
	}

	// Validate format
	switch cmd.Format {
	case "CSV", "JSON", "ASCII", "TXT", "TEXT":
		// Valid formats
		if cmd.Format == "TXT" || cmd.Format == "TEXT" {
			cmd.Format = "ASCII"
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s. Supported formats: CSV, JSON, ASCII", cmd.Format)
	}

	return cmd, nil
}

// parseQuotedString parses a quoted or unquoted string from input
func parseQuotedString(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return "", "", fmt.Errorf("empty string")
	}

	// Handle single-quoted strings
	if input[0] == '\'' {
		end := strings.Index(input[1:], "'")
		if end == -1 {
			return "", "", fmt.Errorf("unclosed quote")
		}
		return input[1 : end+1], input[end+2:], nil
	}

	// Handle double-quoted strings
	if input[0] == '"' {
		end := strings.Index(input[1:], "\"")
		if end == -1 {
			return "", "", fmt.Errorf("unclosed quote")
		}
		return input[1 : end+1], input[end+2:], nil
	}

	// Handle unquoted strings (stop at space or AS/WITH)
	// Look for AS or WITH keywords
	upperInput := strings.ToUpper(input)
	asIndex := strings.Index(upperInput, " AS ")
	withIndex := strings.Index(upperInput, " WITH ")

	endIndex := len(input)
	if asIndex > 0 && asIndex < endIndex {
		endIndex = asIndex
	}
	if withIndex > 0 && withIndex < endIndex {
		endIndex = withIndex
	}

	// Also stop at first space if no keywords found
	spaceIndex := strings.Index(input, " ")
	if spaceIndex > 0 && spaceIndex < endIndex && asIndex == -1 && withIndex == -1 {
		endIndex = spaceIndex
	}

	if endIndex > 0 {
		return input[:endIndex], input[endIndex:], nil
	}

	return input, "", nil
}

// parseSaveOptions parses WITH clause options
func parseSaveOptions(optionsStr string) map[string]interface{} {
	options := make(map[string]interface{})

	// Simple parsing for key=value pairs
	pairs := strings.Split(optionsStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '\'' && value[len(value)-1] == '\'') ||
				(value[0] == '"' && value[len(value)-1] == '"') {
				value = value[1 : len(value)-1]
			}
		}

		// Convert to appropriate type
		switch strings.ToLower(value) {
		case "true":
			options[key] = true
		case "false":
			options[key] = false
		default:
			options[key] = value
		}
	}

	return options
}

// detectFormatFromExtension auto-detects format from filename extension
func detectFormatFromExtension(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		return "CSV"
	case strings.HasSuffix(lower, ".json"):
		return "JSON"
	case strings.HasSuffix(lower, ".txt") || strings.HasSuffix(lower, ".text"):
		return "ASCII"
	default:
		return "CSV" // Default to CSV
	}
}

// HandleSaveCommand executes the save operation
func HandleSaveCommand(cmd SaveCommand, tableData [][]string) error {
	// Validate data availability
	if len(tableData) == 0 {
		return fmt.Errorf("no data to export")
	}

	// Expand home directory if needed
	if strings.HasPrefix(cmd.Filename, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			cmd.Filename = filepath.Join(home, cmd.Filename[2:])
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(cmd.Filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Export based on format
	switch cmd.Format {
	case "CSV":
		return exportToCSV(cmd.Filename, tableData, cmd.Options)
	case "JSON":
		return exportToJSON(cmd.Filename, tableData, cmd.Options)
	case "ASCII":
		return exportToASCII(cmd.Filename, tableData, cmd.Options)
	default:
		return fmt.Errorf("unsupported format: %s", cmd.Format)
	}
}

// exportToCSV exports data to CSV format
func exportToCSV(filename string, data [][]string, options map[string]interface{}) error {
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check if header should be included (default: true)
	includeHeader := true
	if val, ok := options["header"]; ok {
		if b, ok := val.(bool); ok {
			includeHeader = b
		}
	}

	startRow := 0
	if !includeHeader && len(data) > 0 {
		startRow = 1 // Skip header row
	}

	for i := startRow; i < len(data); i++ {
		// Clean row - remove ANSI codes and clean headers
		cleanRow := make([]string, len(data[i]))
		for j, cell := range data[i] {
			cleanCell := stripAnsi(cell)
			// For header row, remove (PK) and (C) indicators
			if i == 0 {
				cleanCell = strings.TrimSuffix(cleanCell, " (PK)")
				cleanCell = strings.TrimSuffix(cleanCell, " (C)")
			}
			cleanRow[j] = cleanCell
		}
		if err := writer.Write(cleanRow); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	return nil
}

// exportToJSON exports data to JSON format
func exportToJSON(filename string, data [][]string, options map[string]interface{}) error {
	if len(data) < 2 {
		// Just headers, no data
		if len(data) == 1 {
			// Write empty array
			return os.WriteFile(filename, []byte("[]"), 0600)
		}
		return fmt.Errorf("insufficient data for JSON export")
	}

	headers := data[0]
	rows := data[1:]

	// Clean headers - remove (PK) and (C) indicators
	cleanHeaders := make([]string, len(headers))
	for i, header := range headers {
		cleanHeader := stripAnsi(header)
		cleanHeader = strings.TrimSuffix(cleanHeader, " (PK)")
		cleanHeader = strings.TrimSuffix(cleanHeader, " (C)")
		cleanHeaders[i] = cleanHeader
	}

	// Convert to array of objects
	jsonData := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		obj := make(map[string]string)
		for i, cell := range row {
			if i < len(cleanHeaders) {
				cleanCell := stripAnsi(cell)
				obj[cleanHeaders[i]] = cleanCell
			}
		}
		jsonData = append(jsonData, obj)
	}

	// Marshal with or without pretty printing
	var jsonBytes []byte
	var err error
	if pretty, ok := options["pretty"].(bool); ok && pretty {
		jsonBytes, err = json.MarshalIndent(jsonData, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(jsonData)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return os.WriteFile(filename, jsonBytes, 0600)
}

// exportToASCII exports data to ASCII table format
func exportToASCII(filename string, data [][]string, options map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to export")
	}

	// Calculate column widths
	colWidths := make([]int, len(data[0]))
	for rowIdx, row := range data {
		for i, cell := range row {
			cleanCell := stripAnsi(cell)
			// For headers, remove (PK) and (C) before calculating width
			if rowIdx == 0 {
				cleanCell = strings.TrimSuffix(cleanCell, " (PK)")
				cleanCell = strings.TrimSuffix(cleanCell, " (C)")
			}
			cellWidth := len(cleanCell)
			if cellWidth > colWidths[i] {
				colWidths[i] = cellWidth
			}
		}
	}

	var output strings.Builder

	// Top border
	output.WriteString("+")
	for _, width := range colWidths {
		output.WriteString(strings.Repeat("-", width+2) + "+")
	}
	output.WriteString("\n")

	// Header row
	if len(data) > 0 {
		output.WriteString("|")
		for i, header := range data[0] {
			// Clean header - remove ANSI and (PK)/(C) indicators
			cleanHeader := stripAnsi(header)
			cleanHeader = strings.TrimSuffix(cleanHeader, " (PK)")
			cleanHeader = strings.TrimSuffix(cleanHeader, " (C)")
			padding := colWidths[i] - len(cleanHeader)
			output.WriteString(" " + cleanHeader + strings.Repeat(" ", padding) + " |")
		}
		output.WriteString("\n")

		// Header separator
		output.WriteString("+")
		for _, width := range colWidths {
			output.WriteString(strings.Repeat("-", width+2) + "+")
		}
		output.WriteString("\n")
	}

	// Data rows
	for i := 1; i < len(data); i++ {
		output.WriteString("|")
		for j, cell := range data[i] {
			cleanCell := stripAnsi(cell)
			padding := colWidths[j] - len(cleanCell)
			if padding < 0 {
				padding = 0
			}
			output.WriteString(" " + cleanCell + strings.Repeat(" ", padding) + " |")
		}
		output.WriteString("\n")
	}

	// Bottom border
	output.WriteString("+")
	for _, width := range colWidths {
		output.WriteString(strings.Repeat("-", width+2) + "+")
	}
	output.WriteString("\n")

	// Add row count footer
	rowCount := len(data) - 1 // Exclude header
	output.WriteString(fmt.Sprintf("\n(%d rows)\n", rowCount))

	return os.WriteFile(filename, []byte(output.String()), 0600)
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// GenerateDefaultFilename generates a default filename with timestamp
func GenerateDefaultFilename(format string) string {
	timestamp := time.Now().Format("20060102_150405")
	ext := ".csv"
	switch format {
	case "JSON":
		ext = ".json"
	case "ASCII":
		ext = ".txt"
	}
	return fmt.Sprintf("query_results_%s%s", timestamp, ext)
}