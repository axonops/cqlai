package router

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// captureFormats are the formats CAPTURE takes in front of a filename. Without
// one it writes text.
var captureFormats = []string{"CSV", "JSON", "PARQUET"}

// captureOptions are the options CAPTURE takes after WITH.
var captureOptions = []string{"COMPRESSION", "MAX_FILE_SIZE", "PARTITION"}

// CaptureOptions names the options CAPTURE accepts after WITH.
func CaptureOptions() []string {
	return slices.Clone(captureOptions)
}

// CaptureFormats names the formats CAPTURE accepts.
//
// Anything offering a choice of format reads this rather than writing the list
// out again. The consistency levels, the output formats and the key column
// markers were each spelled out in several places and each drifted.
func CaptureFormats() []string {
	return slices.Clone(captureFormats)
}

// autoSaveCommand is what the command is called.
//
// CAPTURE is cqlsh's word and describes the mechanism; what is happening is
// that every result is saved as it arrives, without being asked. Both words
// reach here - see ParseCommand - so nobody's scripts break, and everything
// that says the name says this one.
const autoSaveCommand = "AUTOSAVE"

// handleCapture handles the AUTOSAVE command, which saves each query's output
// into a directory
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
		if h.autoSaveDir != "" {
			status := fmt.Sprintf("Saving each query to %s (%s)", h.autoSaveDir, h.captureFormat)
			if h.lastAutoSaved != "" {
				status += fmt.Sprintf("\nLast written: %s", h.lastAutoSaved)
			}
			return status
		}
		return "Not saving query output"
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
		if slices.Contains(captureFormats, upperFormat) {
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
		return fmt.Sprintf(
			"Usage: %s [JSON|CSV|PARQUET] 'directory' [WITH option=value AND ...] | %s OFF\n"+
				"A directory, not a file: each query's output is saved as its own timestamped file.",
			autoSaveCommand, autoSaveCommand)
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

	// A directory, not a file.
	//
	// AutoSave writes one file per query, so what it is given is where to put
	// them. A single file cannot hold the output of every query run while it is
	// on: two queries against different tables have different columns, and
	// Parquet has one schema per file. CSV has the same problem more quietly -
	// a header row, then rows from another table underneath it.
	dir := filename
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Sprintf("Error creating %s: %v", dir, err)
	}

	// Whatever was on stops before this starts.
	h.stopCapture()

	h.autoSaveDir = dir
	h.captureFormat = format
	h.captureOptions = options

	if format == "parquet" && partitionColumns != "" {
		h.capturePartitionColumns = strings.Split(partitionColumns, ",")
		for i := range h.capturePartitionColumns {
			h.capturePartitionColumns[i] = strings.TrimSpace(h.capturePartitionColumns[i])
		}
		return fmt.Sprintf("Saving each query to %s (%s, partitioned by %s). %s OFF stops.",
			dir, format, partitionColumns, autoSaveCommand)
	}

	return fmt.Sprintf("Saving each query to %s (%s). %s OFF stops.", dir, format, autoSaveCommand)
}

// autoSaveName is where the next query's output goes.
//
// A prefix and the time, the same shape the Save window offers as its default
// filename - so the files sort into the order the queries ran, and two queries
// a second apart do not land on the same name.
func (h *MetaCommandHandler) autoSaveName(format string) string {
	ext := "." + format
	if format == "text" {
		ext = ".txt"
	}

	// The time, and then which query it was.
	//
	// The time alone is not enough. Nobody types two queries in the same
	// second, but a SOURCE script or a batch run fires them as fast as the
	// cluster answers - which is where saving every query is most worth doing -
	// and two landing in the same second would write over each other.
	h.autoSaveCount++
	name := fmt.Sprintf("query_%s_%03d%s", time.Now().Format("20060102_150405"), h.autoSaveCount, ext)
	return filepath.Join(h.autoSaveDir, name)
}

// startAutoSaveFile makes sure the right file is open for what is about to be
// written.
//
// A query's rows do not all arrive at once: the first page comes with the
// result and the rest as they are paged in, and those later writes carry no
// command. So the file is opened when a command arrives and stays open until
// the next one does - closing it after every write gave one query a file per
// page, each holding a copy of what the last one already had.
func (h *MetaCommandHandler) startAutoSaveFile(command string) error {
	if h.autoSaveDir == "" || len(h.capturePartitionColumns) > 0 {
		return nil
	}

	// A result reaches this twice - once from WriteCaptureResultWithTypes and
	// again from the function it delegates to - so the file is opened once and
	// closed once, when the outermost of them is done.
	h.autoSaveDepth++

	if h.captureOutput != nil {
		return nil
	}
	if err := h.openAutoSaveFile(); err != nil {
		h.autoSaveDepth--
		return err
	}
	h.openForCommand = command
	return nil
}

// finishAutoSaveFile closes the file once the write that opened it is done.
//
// Closed at the end of the write rather than held open until the next query. A
// Parquet file is nothing until its footer is written, so one left open is a
// zero-byte file - and looking at what AutoSave has written should not require
// running another query first.
func (h *MetaCommandHandler) finishAutoSaveFile() {
	if h.autoSaveDepth == 0 {
		return
	}

	h.autoSaveDepth--
	if h.autoSaveDepth == 0 {
		h.closeAutoSaveFile()
	}
}

// openAutoSaveFile starts a file for one query's output.
//
// Opened here rather than when AUTOSAVE was switched on, because there is a
// file per query: which one is being written is only known once a query has
// run, and the schema of a Parquet file is only known once its columns are.
func (h *MetaCommandHandler) openAutoSaveFile() error {
	name := h.autoSaveName(h.captureFormat)

	writer, err := parquet.CreateWriter(context.Background(), name)
	if err != nil {
		return fmt.Errorf("opening %s: %w", name, err)
	}
	h.captureOutput = writer
	h.lastAutoSaved = name

	switch h.captureFormat {
	case "json":
		_, _ = writer.Write([]byte("[\n"))
	case "csv":
		h.csvWriter = csv.NewWriter(writer)
	case "parquet":
		// The writer is made once the columns are known.
		h.parquetWriter = nil
		h.captureHeaders = nil
	}
	return nil
}

// closeAutoSaveFile finishes the file for one query.
func (h *MetaCommandHandler) closeAutoSaveFile() {
	if h.captureOutput == nil {
		return
	}

	switch h.captureFormat {
	case "json":
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
	h.openForCommand = ""
	h.autoSaveDepth = 0
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
	if h.autoSaveDir == "" {
		return "Not saving query output"
	}

	// A query's file is closed when that query is written, so there is nothing
	// left to close but a partitioned writer, which spans the whole session.
	h.closeAutoSaveFile()
	if h.partitionedWriter != nil {
		_ = h.partitionedWriter.Close()
		h.partitionedWriter = nil
	}

	result := fmt.Sprintf("Stopped saving query output to %s", h.autoSaveDir)
	h.autoSaveDir = ""
	h.lastAutoSaved = ""
	h.autoSaveCount = 0
	h.captureFormat = "text"
	h.captureOptions = nil
	h.capturePartitionColumns = nil
	h.captureColumnTypes = nil
	return result
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
	// Whether a directory is set, not whether a file happens to be open. The
	// files come and go with each query now, and between two queries there is
	// none - which is not the same as AutoSave being off.
	return h.autoSaveDir != ""
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
	// Pages that arrive after the result get a file of their own, numbered
	// after it. Parquet cannot be appended to - a file is sealed by its footer
	// - so the choice is a second part or nothing, and a second part is the
	// one that keeps the rows.
	if err := h.startAutoSaveFile(""); err != nil {
		return err
	}
	defer h.finishAutoSaveFile()

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
	if err := h.startAutoSaveFile(command); err != nil {
		return err
	}
	defer h.finishAutoSaveFile()

	// Handle partitioned Parquet capture
	logger.DebugfToFile("WriteCaptureResultWithTypes", "Format: %s, PartitionColumns: %v, Rows: %d",
		h.captureFormat, h.capturePartitionColumns, len(rows))
	if h.captureFormat == "parquet" && h.capturePartitionColumns != nil && len(h.capturePartitionColumns) > 0 {
		// Create partitioned writer if not exists
		if h.partitionedWriter == nil {
			logger.DebugfToFile("WriteCaptureResultWithTypes", "Creating partitioned writer for path: %s", h.autoSaveDir)
			// Parquet wants the column names Cassandra knows.
			cleanHeaders := db.StripKeyMarkers(headers)

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
			writer, err := parquet.NewPartitionedParquetWriter(h.autoSaveDir, cleanHeaders, columnTypes, writerOptions)
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
		writer, err := parquet.NewParquetCaptureWriter(h.lastAutoSaved, cleanHeaders, columnTypes, options)
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
	if err := h.startAutoSaveFile(command); err != nil {
		return err
	}
	defer h.finishAutoSaveFile()

	// Skip if using partitioned writer (handled in WriteCaptureResultWithTypes)
	if h.partitionedWriter != nil {
		return nil
	}

	if h.captureOutput == nil && h.captureFormat != "parquet" {
		return nil
	}

	// A write with no command is more of the query already being written, so
	// the file gets its header row once, at the top, rather than again in the
	// middle every time another page arrives.
	continuing := command == ""

	switch h.captureFormat {
	case "csv":
		if !continuing {
			_ = h.csvWriter.Write([]string{"# Query: " + command})
			if err := h.csvWriter.Write(headers); err != nil {
				return err
			}
		}

		for _, row := range rows {
			if err := h.csvWriter.Write(row); err != nil {
				return err
			}
		}

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
			writer, err := parquet.NewParquetCaptureWriter(h.lastAutoSaved, headers, columnTypes, options)
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
