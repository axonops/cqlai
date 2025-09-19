package router

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axonops/cqlai/internal/parquet"
)

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
			switch h.captureFormat {
			case "json":
				// Seek back to remove trailing comma if exists
				info, _ := h.captureOutput.Stat()
				if info.Size() > 2 {
					_, _ = h.captureOutput.Seek(-2, 2) // Go to end minus 2 chars
					_, _ = h.captureOutput.WriteString("\n]\n")
				} else {
					_, _ = h.captureOutput.WriteString("]\n")
				}
			case "csv":
				if h.csvWriter != nil {
					// Flush CSV writer
					h.csvWriter.Flush()
					h.csvWriter = nil
				}
			case "parquet":
				if h.parquetWriter != nil {
					// Close Parquet writer
					_ = h.parquetWriter.Close()
					h.parquetWriter = nil
					h.captureHeaders = nil
				}
			}

			_ = h.captureOutput.Close()
			h.captureOutput = nil
			result := fmt.Sprintf("Stopped capturing to %s", h.captureFile)
			h.captureFile = ""
			h.captureFormat = "text"
			return result
		}
		return "Not currently capturing"
	}

	// Parse capture command: CAPTURE [JSON|CSV|PARQUET] 'filename'
	format := "text"
	filenameStart := 1

	if len(parts) >= 2 {
		upperFormat := strings.ToUpper(parts[1])
		switch upperFormat {
		case "JSON":
			format = "json"
			filenameStart = 2
		case "CSV":
			format = "csv"
			filenameStart = 2
		case "PARQUET":
			format = "parquet"
			filenameStart = 2
		}
	}

	if len(parts) <= filenameStart {
		return "Usage: CAPTURE [JSON|CSV|PARQUET] 'filename' | CAPTURE OFF"
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
	switch format {
	case "json":
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}
	case "csv":
		if !strings.HasSuffix(filename, ".csv") {
			filename += ".csv"
		}
	case "parquet":
		if !strings.HasSuffix(filename, ".parquet") {
			filename += ".parquet"
		}
	}

	// Close existing capture file if any
	if h.captureOutput != nil {
		if h.csvWriter != nil {
			h.csvWriter.Flush()
			h.csvWriter = nil
		}
		if h.parquetWriter != nil {
			_ = h.parquetWriter.Close()
			h.parquetWriter = nil
			h.captureHeaders = nil
		}
		_ = h.captureOutput.Close()
	}

	// Open new capture file
	file, err := os.Create(filename) // #nosec G304 - User-provided capture filename
	if err != nil {
		return fmt.Sprintf("Error opening capture file: %v", err)
	}

	h.captureOutput = file
	h.captureFile = filename
	h.captureFormat = format

	switch format {
	case "json":
		// Write opening bracket for JSON array
		_, _ = file.WriteString("[\n")
	case "csv":
		// Create CSV writer
		h.csvWriter = csv.NewWriter(file)
	case "parquet":
		// Parquet writer will be created when we know the schema
		// (on the first result set)
		h.parquetWriter = nil
		h.captureHeaders = nil
	}

	return fmt.Sprintf("Now capturing query output to %s (format: %s). Use 'CAPTURE OFF' to stop.", filename, format)
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

// WriteCaptureText writes text output (like DESCRIBE results) to the capture file
func (h *MetaCommandHandler) WriteCaptureText(command string, output string) error {
	if h.captureOutput == nil {
		return nil
	}

	switch h.captureFormat {
	case "csv":
		// For CSV format, write as a single column with the output
		// Write command as comment
		_ = h.csvWriter.Write([]string{"# Command: " + command})

		// Split output by lines and write each as a row
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if err := h.csvWriter.Write([]string{line}); err != nil {
				return err
			}
		}

		// Add empty row to separate commands
		_ = h.csvWriter.Write([]string{})

		// Flush to ensure data is written
		h.csvWriter.Flush()
		return h.csvWriter.Error()

	case "json":
		// For JSON format, create a text result object
		type TextResult struct {
			Command string `json:"command"`
			Output  string `json:"output"`
			Type    string `json:"type"`
		}

		result := TextResult{
			Command: command,
			Output:  output,
			Type:    "text",
		}

		jsonBytes, err := json.MarshalIndent(result, "  ", "  ")
		if err != nil {
			return err
		}

		// Check if this is the first entry
		info, _ := h.captureOutput.Stat()
		if info.Size() > 2 { // More than just "[\n"
			_, _ = h.captureOutput.WriteString(",\n")
		}
		_, _ = h.captureOutput.WriteString("  ")
		_, _ = h.captureOutput.Write(jsonBytes)

	default:
		// Text format - write the command and output
		_, _ = fmt.Fprintf(h.captureOutput, "\n> %s\n", command)
		_, _ = h.captureOutput.WriteString(strings.Repeat("-", 50) + "\n")
		_, _ = h.captureOutput.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			_, _ = h.captureOutput.WriteString("\n")
		}
		// Add just a blank line for separation
		_, _ = h.captureOutput.WriteString("\n")
	}

	return nil
}

// FormatResultAsJSON formats query results as JSON
func FormatResultAsJSON(headers []string, rows [][]string) (string, error) {
	return FormatResultAsJSONWithRawData(headers, rows, nil)
}

// AppendCaptureRows appends additional rows to the capture file (for paging)
func (h *MetaCommandHandler) AppendCaptureRows(rows [][]string) error {
	if h.captureOutput == nil && h.captureFormat != "parquet" {
		return nil
	}

	switch h.captureFormat {
	case "csv":
		// Write data rows only (no headers for continuation)
		for _, row := range rows {
			if err := h.csvWriter.Write(row); err != nil {
				return err
			}
		}
		// Flush to ensure data is written
		h.csvWriter.Flush()
		return h.csvWriter.Error()

	case "json":
		// For JSON, we can't easily append to an existing object
		// So we'll skip continuation rows in JSON format
		// This maintains valid JSON structure
		return nil

	case "parquet":
		// Append rows to Parquet file
		if h.parquetWriter != nil {
			// Use the headers we stored when creating the writer
			if err := h.parquetWriter.WriteStringRows(h.captureHeaders, rows); err != nil {
				return fmt.Errorf("failed to append Parquet rows: %w", err)
			}
		}
		return nil

	default:
		// Text format - just append the rows, no command header
		for _, row := range rows {
			_, _ = h.captureOutput.WriteString(strings.Join(row, "\t") + "\n")
		}
	}

	return nil
}

// WriteCaptureResult writes query results to the capture file
func (h *MetaCommandHandler) WriteCaptureResult(command string, headers []string, rows [][]string) error {
	return h.WriteCaptureResultWithRawData(command, headers, rows, nil)
}

// WriteCaptureResultWithTypes writes query results with column type information (for Parquet support)
func (h *MetaCommandHandler) WriteCaptureResultWithTypes(command string, headers []string, columnTypes []string, rows [][]string, rawData []map[string]interface{}) error {
	// Store column types for parquet writer
	if h.captureFormat == "parquet" && h.parquetWriter == nil && len(headers) > 0 {
		// Clean column names for Parquet - remove (PK) and (C) suffixes
		cleanHeaders := make([]string, len(headers))
		for i, header := range headers {
			// Remove (PK) suffix
			if idx := strings.Index(header, " (PK)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else if idx := strings.Index(header, " (C)"); idx != -1 {
				// Remove (C) suffix
				cleanHeaders[i] = header[:idx]
			} else {
				// No suffix to remove
				cleanHeaders[i] = header
			}
		}

		// Create the Parquet writer with clean column names
		options := parquet.DefaultWriterOptions()
		writer, err := parquet.NewParquetCaptureWriter(h.captureFile, cleanHeaders, columnTypes, options)
		if err != nil {
			return fmt.Errorf("failed to create Parquet writer: %w", err)
		}
		h.parquetWriter = writer
		h.captureHeaders = headers

		// Write the header (no-op for Parquet, but maintains interface consistency)
		if err := writer.WriteHeader(); err != nil {
			return fmt.Errorf("failed to write Parquet header: %w", err)
		}
	}

	// Delegate to the existing method
	return h.WriteCaptureResultWithRawData(command, headers, rows, rawData)
}

// WriteCaptureResultWithRawData writes query results to the capture file with optional raw data for JSON
func (h *MetaCommandHandler) WriteCaptureResultWithRawData(command string, headers []string, rows [][]string, rawData []map[string]interface{}) error {
	if h.captureOutput == nil && h.captureFormat != "parquet" {
		return nil
	}

	switch h.captureFormat {
	case "csv":
		// Write headers as first row (with query command as comment)
		// Write comment with the query
		_ = h.csvWriter.Write([]string{"# Query: " + command})

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
		_ = h.csvWriter.Write([]string{})

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

		// Use raw data if provided, otherwise fall back to string parsing
		if rawData != nil && len(rawData) == len(rows) {
			// Use the raw data directly - it preserves types
			result.Rows = rawData
		} else {
			// Fall back to parsing strings (backward compatibility)
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
		}

		jsonBytes, err := json.MarshalIndent(result, "  ", "  ")
		if err != nil {
			return err
		}

		// Check if this is the first entry
		info, _ := h.captureOutput.Stat()
		if info.Size() > 2 { // More than just "[\n"
			_, _ = h.captureOutput.WriteString(",\n")
		}
		_, _ = h.captureOutput.WriteString("  ")
		_, _ = h.captureOutput.Write(jsonBytes)

	case "parquet":
		if h.parquetWriter == nil {
			// If the writer wasn't created yet (shouldn't happen if WriteCaptureResultWithTypes was called)
			// We'll create a simple writer with all columns as text type
			columnTypes := make([]string, len(headers))
			for i := range columnTypes {
				columnTypes[i] = "text"
			}

			options := parquet.DefaultWriterOptions()
			writer, err := parquet.NewParquetCaptureWriter(h.captureFile, headers, columnTypes, options)
			if err != nil {
				return fmt.Errorf("failed to create Parquet writer: %w", err)
			}
			h.parquetWriter = writer
			h.captureHeaders = headers
		}

		// Write the data
		if rawData != nil && len(rawData) == len(rows) {
			// Use raw data if available (preserves types)
			for _, rowData := range rawData {
				if err := h.parquetWriter.WriteRow(rowData); err != nil {
					return fmt.Errorf("failed to write Parquet row: %w", err)
				}
			}
		} else {
			// Use string data
			if err := h.parquetWriter.WriteStringRows(headers, rows); err != nil {
				return fmt.Errorf("failed to write Parquet string rows: %w", err)
			}
		}

	default:
		// Text format - write the command and a simple table representation
		_, _ = fmt.Fprintf(h.captureOutput, "\n> %s\n", command)
		_, _ = h.captureOutput.WriteString(strings.Repeat("-", 50) + "\n")

		// Write headers
		_, _ = h.captureOutput.WriteString(strings.Join(headers, "\t") + "\n")

		// Write rows
		for _, row := range rows {
			_, _ = h.captureOutput.WriteString(strings.Join(row, "\t") + "\n")
		}

		// Add just a blank line for separation, no row count
		_, _ = h.captureOutput.WriteString("\n")
	}

	return nil
}
