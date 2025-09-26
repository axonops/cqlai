package router

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// handleCapture handles CAPTURE command to save output to file
func (h *MetaCommandHandler) handleCapture(command string) interface{} {
	// Parse the command to extract format, filename, and options
	upperCommand := strings.ToUpper(command)
	logger.DebugfToFile("Capture", "Command: %s", command)

	// Check for OFF first
	if strings.Contains(upperCommand, " OFF") {
		return h.stopCapture()
	}

	// Show current capture status if no arguments
	parts := strings.Fields(command)
	if len(parts) == 1 {
		if h.captureFile != "" {
			status := fmt.Sprintf("Currently capturing to: %s (format: %s)", h.captureFile, h.captureFormat)
			if h.partitionedWriter != nil {
				status += " [partitioned]"
			}
			return status
		}
		return "Not currently capturing output"
	}

	// Parse CAPTURE command with options
	// Format: CAPTURE [JSON|CSV|PARQUET] 'filename' [WITH option=value AND ...]
	format := "text"
	filenameStart := 1
	filenameEnd := len(parts)
	var options map[string]string

	// Check for format specifier
	if len(parts) >= 2 {
		upperFormat := strings.ToUpper(parts[1])
		switch upperFormat {
		case "JSON", "CSV", "PARQUET":
			format = strings.ToLower(upperFormat)
			filenameStart = 2
		}
	}

	// Find WITH clause if present
	withIndex := -1
	for i := filenameStart; i < len(parts); i++ {
		if strings.ToUpper(parts[i]) == "WITH" {
			withIndex = i
			filenameEnd = i
			break
		}
	}

	// Extract filename
	if filenameStart >= filenameEnd {
		return "Usage: CAPTURE [JSON|CSV|PARQUET] 'filename' [WITH option=value AND ...] | CAPTURE OFF"
	}

	filename := strings.Join(parts[filenameStart:filenameEnd], " ")
	filename = strings.Trim(filename, "'\"")

	// Expand home directory if needed
	if strings.HasPrefix(filename, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			filename = filepath.Join(home, filename[2:])
		}
	}

	// Parse options if WITH clause is present
	if withIndex >= 0 && withIndex < len(parts)-1 {
		options = h.parseWithOptions(strings.Join(parts[withIndex+1:], " "))
	}

	// Check for PARTITION option (only valid for PARQUET format)
	partitionColumns := ""
	if options != nil {
		if partition, ok := options["PARTITION"]; ok {
			if format != "parquet" {
				return "PARTITION option is only supported for PARQUET format"
			}
			partitionColumns = partition
			// For partitioned output, filename becomes a directory
			if !strings.HasSuffix(filename, "/") {
				filename += "/"
			}
		}
	}

	// Add appropriate extension if not provided (for non-partitioned files)
	if partitionColumns == "" {
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
	}

	// Close existing capture if any
	h.stopCapture()

	// Initialize new capture
	h.captureFile = filename
	h.captureFormat = format
	h.captureOptions = options

	// For partitioned Parquet, we'll create the writer when we receive the first result
	if format == "parquet" && partitionColumns != "" {
		h.capturePartitionColumns = strings.Split(partitionColumns, ",")
		for i := range h.capturePartitionColumns {
			h.capturePartitionColumns[i] = strings.TrimSpace(h.capturePartitionColumns[i])
		}
		logger.DebugfToFile("Capture", "Initialized partitioned capture: file=%s, format=%s, partitions=%v",
			filename, format, h.capturePartitionColumns)
		return fmt.Sprintf("Now capturing query output to %s (format: %s, partitioned by: %s). Use 'CAPTURE OFF' to stop.",
			filename, format, partitionColumns)
	}

	// For non-partitioned formats, create the file immediately
	if partitionColumns == "" {
		// Use parquet.CreateWriter which handles both local files and cloud URL error messages
		writer, err := parquet.CreateWriter(context.Background(), filename)
		if err != nil {
			return fmt.Sprintf("Error opening capture file: %v", err)
		}

		h.captureOutput = writer

		switch format {
		case "json":
			// Write opening bracket for JSON array
			_, _ = writer.Write([]byte("[\n"))
		case "csv":
			// Create CSV writer
			h.csvWriter = csv.NewWriter(writer)
		case "parquet":
			// Parquet writer will be created when we know the schema
			h.parquetWriter = nil
			h.captureHeaders = nil
		}
	}

	return fmt.Sprintf("Now capturing query output to %s (format: %s). Use 'CAPTURE OFF' to stop.", filename, format)
}

// parseWithOptions parses the WITH clause options
func (h *MetaCommandHandler) parseWithOptions(withClause string) map[string]string {
	options := make(map[string]string)

	// Remove 'WITH' if it's at the beginning
	withClause = strings.TrimPrefix(strings.TrimSpace(withClause), "WITH")
	withClause = strings.TrimSpace(withClause)

	// Split by AND
	parts := strings.Split(withClause, " AND ")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Find the equals sign
		eqIndex := strings.Index(part, "=")
		if eqIndex > 0 {
			key := strings.TrimSpace(part[:eqIndex])
			value := strings.TrimSpace(part[eqIndex+1:])
			// Remove quotes from value if present
			value = strings.Trim(value, "'\"")
			options[strings.ToUpper(key)] = value
		}
	}

	return options
}

// stopCapture stops the current capture and closes resources
func (h *MetaCommandHandler) stopCapture() interface{} {
	if h.captureOutput != nil {
		// If JSON format, properly close the array
		switch h.captureFormat {
		case "json":
			// Close the JSON array
			_, _ = h.captureOutput.Write([]byte("\n]\n"))
		case "csv":
			if h.csvWriter != nil {
				h.csvWriter.Flush()
				h.csvWriter = nil
			}
		case "parquet":
			if h.parquetWriter != nil {
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
		h.captureOptions = nil
		h.capturePartitionColumns = nil
		return result
	}

	// Close partitioned writer if exists
	if h.partitionedWriter != nil {
		_ = h.partitionedWriter.Close()
		result := fmt.Sprintf("Stopped capturing to %s", h.captureFile)
		h.partitionedWriter = nil
		h.captureFile = ""
		h.captureFormat = "text"
		h.captureOptions = nil
		h.capturePartitionColumns = nil
		h.captureColumnTypes = nil
		return result
	}

	return "Not currently capturing"
}

// GetCaptureFile returns the current capture file if any
func (h *MetaCommandHandler) GetCaptureFile() io.WriteCloser {
	return h.captureOutput
}

// GetCaptureFormat returns the current capture format ("text" or "json")
func (h *MetaCommandHandler) GetCaptureFormat() string {
	return h.captureFormat
}

// IsCapturing returns true if currently capturing output
func (h *MetaCommandHandler) IsCapturing() bool {
	return h.captureOutput != nil || h.partitionedWriter != nil || (h.captureFormat == "parquet" && len(h.capturePartitionColumns) > 0)
}

// WriteToCapture writes data to the capture file if active
func (h *MetaCommandHandler) WriteToCapture(data string) error {
	if h.captureOutput != nil {
		_, err := h.captureOutput.Write([]byte(data))
		return err
	}
	return nil
}

// WriteCaptureText writes text output (like DESCRIBE results) to the capture file
func (h *MetaCommandHandler) WriteCaptureText(command string, output string) error {
	// Partitioned capture doesn't support text output
	if h.partitionedWriter != nil {
		return nil
	}

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

		// Add comma separator between JSON records
		_, _ = h.captureOutput.Write([]byte(",\n  "))
		_, _ = h.captureOutput.Write(jsonBytes)

	default:
		// Text format - write the command and output
		_, _ = fmt.Fprintf(h.captureOutput, "\n> %s\n", command)
		_, _ = h.captureOutput.Write([]byte(strings.Repeat("-", 50) + "\n"))
		_, _ = h.captureOutput.Write([]byte(output))
		if !strings.HasSuffix(output, "\n") {
			_, _ = h.captureOutput.Write([]byte("\n"))
		}
		// Add just a blank line for separation
		_, _ = h.captureOutput.Write([]byte("\n"))
	}

	return nil
}

// FormatResultAsJSON formats query results as JSON
func FormatResultAsJSON(headers []string, rows [][]string) (string, error) {
	return FormatResultAsJSONWithRawData(headers, rows, nil)
}

// AppendCaptureRows appends additional rows to the capture file (for paging)
func (h *MetaCommandHandler) AppendCaptureRows(rows [][]string) error {
	// Handle partitioned writer
	if h.partitionedWriter != nil {
		// Convert string rows to map format
		rowMaps := make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			rowMap := make(map[string]interface{})
			for j, header := range h.captureHeaders {
				if j < len(row) {
					rowMap[header] = row[j]
				}
			}
			rowMaps[i] = rowMap
		}
		return h.partitionedWriter.WriteRows(rowMaps)
	}

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
			_, _ = h.captureOutput.Write([]byte(strings.Join(row, "\t") + "\n"))
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
	// Handle partitioned Parquet capture
	logger.DebugfToFile("WriteCaptureResultWithTypes", "Format: %s, PartitionColumns: %v, Rows: %d",
		h.captureFormat, h.capturePartitionColumns, len(rows))
	if h.captureFormat == "parquet" && h.capturePartitionColumns != nil && len(h.capturePartitionColumns) > 0 {
		// Create partitioned writer if not exists
		if h.partitionedWriter == nil {
			logger.DebugfToFile("WriteCaptureResultWithTypes", "Creating partitioned writer for path: %s", h.captureFile)
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

			// Parse options for compression and file size
			compressionStr := ""
			maxFileSize := int64(100 * 1024 * 1024) // 100MB default
			if h.captureOptions != nil {
				if comp, ok := h.captureOptions["COMPRESSION"]; ok {
					compressionStr = comp
				}
				if sizeStr, ok := h.captureOptions["MAX_FILE_SIZE"]; ok {
					maxFileSize = parseFileSize(sizeStr)
				}
			}

			// Create partitioned writer options
			writerOptions := parquet.PartitionedWriterOptions{
				WriterOptions: parquet.WriterOptions{
					ChunkSize:   10000,
					Compression: parquet.ParseCompression(compressionStr),
				},
				PartitionColumns: h.capturePartitionColumns,
				MaxOpenFiles:     10,
				MaxFileSize:      maxFileSize,
			}

			// Create partitioned writer
			logger.DebugfToFile("WriteCaptureResultWithTypes", "Creating writer with headers: %v, types: %v", cleanHeaders, columnTypes)
			writer, err := parquet.NewPartitionedParquetWriter(h.captureFile, cleanHeaders, columnTypes, writerOptions)
			if err != nil {
				return fmt.Errorf("failed to create partitioned Parquet writer: %w", err)
			}
			h.partitionedWriter = writer
			h.captureHeaders = cleanHeaders
			h.captureColumnTypes = columnTypes
		}

		// Write data using raw data if available
		if len(rawData) > 0 {
			logger.DebugfToFile("WriteCaptureResultWithTypes", "Writing %d rows with rawData to partitioned writer", len(rawData))
			// Log first row to see what columns are present
			if len(rawData) > 0 {
				logger.DebugfToFile("WriteCaptureResultWithTypes", "First row columns: %v", getMapKeys(rawData[0]))
			}
			// The partitioned writer handles virtual column extraction internally
			err := h.partitionedWriter.WriteRows(rawData)
			if err != nil {
				logger.DebugfToFile("WriteCaptureResultWithTypes", "Error writing rows: %v", err)
			}
			return err
		}

		// Convert string rows to map format
		logger.DebugfToFile("WriteCaptureResultWithTypes", "Converting %d string rows to map format", len(rows))
		rowMaps := make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			rowMap := make(map[string]interface{})
			for j, header := range h.captureHeaders {
				if j < len(row) {
					rowMap[header] = row[j]
				}
			}
			rowMaps[i] = rowMap
		}
		// The partitioned writer handles virtual column extraction internally
		err := h.partitionedWriter.WriteRows(rowMaps)
		if err != nil {
			logger.DebugfToFile("WriteCaptureResultWithTypes", "Error writing rows: %v", err)
		}
		return err
	}

	// Non-partitioned capture - use original logic
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
	// Skip if using partitioned writer (handled in WriteCaptureResultWithTypes)
	if h.partitionedWriter != nil {
		return nil
	}

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

		// Add comma separator between JSON records
		_, _ = h.captureOutput.Write([]byte(",\n  "))
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
		_, _ = h.captureOutput.Write([]byte(strings.Repeat("-", 50) + "\n"))

		// Write headers
		_, _ = h.captureOutput.Write([]byte(strings.Join(headers, "\t") + "\n"))

		// Write rows
		for _, row := range rows {
			_, _ = h.captureOutput.Write([]byte(strings.Join(row, "\t") + "\n"))
		}

		// Add just a blank line for separation, no row count
		_, _ = h.captureOutput.Write([]byte("\n"))
	}

	return nil
}

// getMapKeys returns the keys of a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

