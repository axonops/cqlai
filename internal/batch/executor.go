package batch

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
	"github.com/axonops/cqlai/internal/ui"
)

// OutputFormat represents the output format for batch mode
type OutputFormat string

const (
	OutputFormatASCII OutputFormat = "ascii"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatCSV   OutputFormat = "csv"
	OutputFormatTable OutputFormat = "table" // Default table format
)

// Options contains batch execution options
type Options struct {
	Execute      string       // CQL to execute directly (-e flag)
	File         string       // CQL file to execute (-f flag)
	Format       OutputFormat // Output format
	NoHeader     bool         // Skip headers in output
	FieldSep     string       // Field separator for CSV
	NoPager      bool         // Disable paging (print all results)
	PageSize     int          // Number of rows per batch for streaming
	ConnOptions  ui.ConnectionOptions
}

// Executor handles batch mode execution
type Executor struct {
	session        *db.Session
	sessionManager *session.Manager
	options        *Options
	writer         io.Writer
}

// NewExecutor creates a new batch executor
func NewExecutor(options *Options, writer io.Writer) (*Executor, error) {
	// Create database session
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{
			Host: "127.0.0.1",
			Port: 9042,
		}
	}
	
	// Enable debug logging if configured (from config file or command-line)
	if cfg.Debug || options.ConnOptions.Debug {
		logger.SetDebugEnabled(true)
	}

	// Override with connection options
	if options.ConnOptions.Host != "" {
		cfg.Host = options.ConnOptions.Host
	}
	if options.ConnOptions.Port != 0 {
		cfg.Port = options.ConnOptions.Port
	}
	if options.ConnOptions.Keyspace != "" {
		cfg.Keyspace = options.ConnOptions.Keyspace
	}
	if options.ConnOptions.Username != "" {
		cfg.Username = options.ConnOptions.Username
	}
	if options.ConnOptions.Password != "" {
		cfg.Password = options.ConnOptions.Password
	}
	
	// Use config PageSize if not specified on command line
	if options.PageSize == 0 && cfg.PageSize > 0 {
		options.PageSize = cfg.PageSize
	}
	// Default to 100 if still not set
	if options.PageSize == 0 {
		options.PageSize = 100
	}

	dbSession, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Keyspace:       cfg.Keyspace,
		Username:       cfg.Username,
		Password:       cfg.Password,
		SSL:            cfg.SSL,
		BatchMode:      true, // Disable schema caching in batch mode
		ConnectTimeout: options.ConnOptions.ConnectTimeout,
		RequestTimeout: options.ConnOptions.RequestTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	// Create session manager for tracking keyspace changes
	sessionMgr := session.NewManager(cfg)
	if cfg.Keyspace != "" {
		sessionMgr.SetKeyspace(cfg.Keyspace)
	}
	
	// Initialize router with session manager
	router.InitRouter(sessionMgr)

	return &Executor{
		session:        dbSession,
		sessionManager: sessionMgr,
		options:        options,
		writer:         writer,
	}, nil
}

// Close closes the executor and its resources
func (e *Executor) Close() error {
	if e.session != nil {
		e.session.Close()
	}
	return nil
}

// Execute runs CQL in batch mode
func (e *Executor) Execute(cql string) error {
	// Set up signal handling for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Process the CQL command
	result := router.ProcessCommand(cql, e.session, e.sessionManager)

	// Handle the result based on type
	var err error
	switch v := result.(type) {
	case db.StreamingQueryResult:
		err = e.handleStreamingResult(ctx, v)
		// Check for tracing data after streaming result
		if err == nil && e.session.Tracing() {
			e.printTraceData()
		}
		return err
	case db.QueryResult:
		err = e.handleQueryResult(v)
		// Check for tracing data after query result
		if err == nil && e.session.Tracing() {
			e.printTraceData()
		}
		return err
	case [][]string:
		err = e.outputTable(v)
		// Check for tracing data after table output
		if err == nil && e.session.Tracing() {
			e.printTraceData()
		}
		return err
	case string:
		// Check if this is a USE command result and update the keyspace
		if strings.HasPrefix(v, "Now using keyspace ") {
			// Extract the keyspace name
			keyspaceName := strings.TrimPrefix(v, "Now using keyspace ")
			keyspaceName = strings.TrimSpace(keyspaceName)
			
			// Update the session manager
			if e.sessionManager != nil {
				e.sessionManager.SetKeyspace(keyspaceName)
			}
			
			// Update the database session's keyspace
			if err := e.session.SetKeyspace(keyspaceName); err != nil {
				return fmt.Errorf("failed to change keyspace: %w", err)
			}
		}
		fmt.Fprintln(e.writer, v)
		return nil
	case error:
		return v
	default:
		return nil
	}
}

// stripComments removes SQL-style comments from CQL statements while preserving newlines
func stripComments(input string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")
	
	inBlockComment := false
	for _, line := range lines {
		// Handle block comments /* ... */
		for {
			if inBlockComment {
				endIdx := strings.Index(line, "*/")
				if endIdx >= 0 {
					line = line[endIdx+2:]
					inBlockComment = false
				} else {
					// Entire line is within block comment
					line = ""
					break
				}
			}
			
			startIdx := strings.Index(line, "/*")
			if startIdx >= 0 {
				endIdx := strings.Index(line[startIdx:], "*/")
				if endIdx >= 0 {
					// Block comment on same line
					line = line[:startIdx] + line[startIdx+endIdx+2:]
				} else {
					// Block comment starts but doesn't end on this line
					line = line[:startIdx]
					inBlockComment = true
					break
				}
			} else {
				break
			}
		}
		
		// Handle line comments -- and //
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		
		// Trim trailing whitespace but preserve the line structure
		line = strings.TrimRight(line, " \t\r")
		
		// Add the line (even if empty) to preserve line breaks
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(line)
	}
	
	return result.String()
}

// splitStatements intelligently splits CQL statements, handling BATCH blocks
func splitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder
	
	// Track if we're inside a BATCH statement
	inBatch := false
	
	// Track if we're accumulating a non-empty statement
	hasContent := false
	
	// Process line by line to handle BATCH blocks and regular statements
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines unless we're in a batch
		if trimmedLine == "" && !inBatch && !hasContent {
			continue
		}
		
		upperLine := strings.ToUpper(trimmedLine)
		
		// Check for BATCH start
		if strings.HasPrefix(upperLine, "BEGIN BATCH") || 
		   strings.HasPrefix(upperLine, "BEGIN UNLOGGED BATCH") ||
		   strings.HasPrefix(upperLine, "BEGIN COUNTER BATCH") {
			inBatch = true
		}
		
		// Add line to current statement
		if currentStmt.Len() > 0 && trimmedLine != "" {
			currentStmt.WriteString(" ")
		}
		if trimmedLine != "" {
			currentStmt.WriteString(trimmedLine)
			hasContent = true
		}
		
		// Check for statement end
		if strings.HasSuffix(trimmedLine, ";") && hasContent {
			if inBatch {
				// Check if this ends the batch
				if strings.HasPrefix(upperLine, "APPLY BATCH") {
					inBatch = false
					// Complete batch statement
					statements = append(statements, currentStmt.String())
					currentStmt.Reset()
					hasContent = false
				}
				// Otherwise, continue accumulating the batch
			} else {
				// Regular statement ended
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
				hasContent = false
			}
		}
	}
	
	// Add any remaining statement
	if currentStmt.Len() > 0 && hasContent {
		statements = append(statements, currentStmt.String())
	}
	
	return statements
}

// ExecuteFile executes CQL from a file
func (e *Executor) ExecuteFile(filename string) error {
	file, err := os.Open(filename) // #nosec G304 - User-provided file path is expected
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Strip comments before processing
	cleanContent := stripComments(string(content))
	
	// Split statements intelligently (handling BATCH blocks)
	statements := splitStatements(cleanContent)
	
	// Debug: log the number of statements found
	if len(statements) == 1 && len(statements[0]) > 1000 {
		// Likely the entire file was treated as one statement
		fmt.Fprintf(os.Stderr, "Warning: Found only 1 very large statement (%d chars). Statement splitting may have failed.\n", len(statements[0]))
		previewLen := 200
		if len(statements[0]) < previewLen {
			previewLen = len(statements[0])
		}
		fmt.Fprintf(os.Stderr, "First %d chars: %s...\n", previewLen, statements[0][:previewLen])
	}
	
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := e.Execute(stmt); err != nil {
			return fmt.Errorf("error executing statement %d: %w", i+1, err)
		}
	}

	return nil
}

// ExecuteStdin executes CQL from stdin
func (e *Executor) ExecuteStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	var buffer strings.Builder
	inBatch := false
	
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		upperLine := strings.ToUpper(trimmedLine)
		
		// Check for BATCH start
		if strings.HasPrefix(upperLine, "BEGIN BATCH") || 
		   strings.HasPrefix(upperLine, "BEGIN UNLOGGED BATCH") ||
		   strings.HasPrefix(upperLine, "BEGIN COUNTER BATCH") {
			inBatch = true
		}
		
		buffer.WriteString(line)
		buffer.WriteString("\n")
		
		// Check if we have a complete statement
		if strings.HasSuffix(trimmedLine, ";") {
			if inBatch {
				// Check if this ends the batch
				if strings.HasPrefix(upperLine, "APPLY BATCH") {
					inBatch = false
					stmt := strings.TrimSpace(buffer.String())
					// Strip comments before executing
					stmt = stripComments(stmt)
					stmt = strings.TrimSpace(stmt)
					if stmt != "" {
						if err := e.Execute(stmt); err != nil {
							return err
						}
					}
					buffer.Reset()
				}
				// Otherwise, continue accumulating the batch
			} else {
				// Regular statement ended
				stmt := strings.TrimSpace(buffer.String())
				// Strip comments before executing
				stmt = stripComments(stmt)
				stmt = strings.TrimSpace(stmt)
				if stmt != "" {
					if err := e.Execute(stmt); err != nil {
						return err
					}
				}
				buffer.Reset()
			}
		}
	}

	// Execute any remaining statement
	if buffer.Len() > 0 {
		stmt := strings.TrimSpace(buffer.String())
		// Strip comments before executing
		stmt = stripComments(stmt)
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			return e.Execute(stmt)
		}
	}

	return scanner.Err()
}

// handleStreamingResult handles streaming query results with automatic pagination
func (e *Executor) handleStreamingResult(ctx context.Context, result db.StreamingQueryResult) error {
	defer result.Iterator.Close()

	// For CSV and JSON, we need to handle differently
	switch e.options.Format {
	case OutputFormatCSV:
		return e.outputStreamingCSV(ctx, result)
	case OutputFormatJSON:
		return e.outputStreamingJSON(ctx, result)
	}

	// For ASCII and Table formats, stream with pagination
	// First, print the table header
	headerData := [][]string{result.Headers}
	headerOutput := ui.FormatASCIITableHeader(headerData)
	fmt.Fprint(e.writer, headerOutput)
	
	rowCount := 0
	var rows [][]string
	
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(e.writer, "\n\nQuery interrupted by user")
			return nil
		default:
			rowMap := make(map[string]interface{})
			if !result.Iterator.MapScan(rowMap) {
				// Check for errors
				if err := result.Iterator.Close(); err != nil {
					return fmt.Errorf("iterator error: %w", err)
				}
				
				// Output final batch if we have rows
				if len(rows) > 0 {
					if err := e.outputStreamingRows(rows, result.Headers); err != nil {
						return err
					}
				}
				
				// Print bottom border and row count
				e.printTableBottom(result.Headers)
				fmt.Fprintf(e.writer, "\n(%d rows)\n", rowCount)
				return nil
			}

			// Convert row to string array
			row := make([]string, len(result.ColumnNames))
			for i, colName := range result.ColumnNames {
				if val, ok := rowMap[colName]; ok {
					if val == nil {
						row[i] = "null"
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				} else {
					row[i] = "null"
				}
			}

			rows = append(rows, row)
			rowCount++

			// Output batch based on configured page size
			batchSize := e.options.PageSize
			if batchSize <= 0 {
				batchSize = 100 // Default to 100 if not set
			}
			if len(rows) >= batchSize {
				if err := e.outputStreamingRows(rows, result.Headers); err != nil {
					return err
				}
				// Clear rows for next batch
				rows = [][]string{}
			}
		}
	}
}

// printTraceData prints tracing information if available
func (e *Executor) printTraceData() {
	// Get trace data from the session
	traceData, headers, traceInfo, err := e.session.GetTraceData()
	if err != nil {
		// Silently ignore if no trace data is available
		return
	}

	// Print a separator
	fmt.Fprintln(e.writer, "\nTracing session:")
	
	// Print trace session info
	if traceInfo != nil {
		fmt.Fprintf(e.writer, "Coordinator: %s | Total Duration: %d μs\n", 
			traceInfo.Coordinator, traceInfo.Duration)
	}
	
	// Format and print the trace data as a table
	if len(traceData) > 0 {
		// Combine headers and data
		fullData := append([][]string{headers}, traceData...)
		// Use ASCII table format for trace data
		output := ui.FormatASCIITable(fullData)
		fmt.Fprint(e.writer, output)
	}
}

// handleQueryResult handles non-streaming query results
func (e *Executor) handleQueryResult(result db.QueryResult) error {
	if len(result.Data) == 0 {
		fmt.Fprintln(e.writer, "(0 rows)")
		return nil
	}

	switch e.options.Format {
	case OutputFormatJSON:
		return e.outputJSONWithRawData(result)
	case OutputFormatCSV:
		return e.outputCSV(result.Data)
	default:
		// outputTable calls FormatASCIITable which already includes the row count
		if err := e.outputTable(result.Data); err != nil {
			return err
		}
		return nil
	}
}

// outputTable outputs data in ASCII table format
func (e *Executor) outputTable(data [][]string) error {
	if len(data) == 0 {
		return nil
	}

	if e.options.Format == OutputFormatASCII || e.options.Format == OutputFormatTable {
		output := ui.FormatASCIITable(data)
		fmt.Fprint(e.writer, output)
	}
	return nil
}

// outputStreamingRows outputs rows without headers for streaming results
func (e *Executor) outputStreamingRows(rows [][]string, headers []string) error {
	if len(rows) == 0 {
		return nil
	}

	// Calculate column widths based on headers and data
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Output each row
	for _, row := range rows {
		fmt.Fprint(e.writer, "|")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(e.writer, " %-*s |", widths[i], cell)
			}
		}
		fmt.Fprintln(e.writer)
	}
	
	return nil
}

// printTableBottom prints the bottom border of the table
func (e *Executor) printTableBottom(headers []string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	// Print bottom border
	fmt.Fprint(e.writer, "+")
	for _, w := range widths {
		fmt.Fprint(e.writer, strings.Repeat("-", w+2))
		fmt.Fprint(e.writer, "+")
	}
	fmt.Fprintln(e.writer)
}

// outputCSV outputs data in CSV format
func (e *Executor) outputCSV(data [][]string) error {
	if len(data) == 0 {
		return nil
	}

	csvWriter := csv.NewWriter(e.writer)
	if e.options.FieldSep != "" {
		csvWriter.Comma = rune(e.options.FieldSep[0])
	}

	startIdx := 0
	if e.options.NoHeader && len(data) > 1 {
		startIdx = 1
	}

	for i := startIdx; i < len(data); i++ {
		if err := csvWriter.Write(data[i]); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

// outputStreamingCSV outputs streaming data in CSV format
func (e *Executor) outputStreamingCSV(ctx context.Context, result db.StreamingQueryResult) error {
	csvWriter := csv.NewWriter(e.writer)
	if e.options.FieldSep != "" {
		csvWriter.Comma = rune(e.options.FieldSep[0])
	}

	// Write header unless suppressed
	if !e.options.NoHeader {
		if err := csvWriter.Write(result.Headers); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	rowCount := 0
	for {
		select {
		case <-ctx.Done():
			csvWriter.Flush()
			return nil
		default:
			rowMap := make(map[string]interface{})
			if !result.Iterator.MapScan(rowMap) {
				csvWriter.Flush()
				return result.Iterator.Close()
			}

			// Convert row to string array
			row := make([]string, len(result.ColumnNames))
			for i, colName := range result.ColumnNames {
				if val, ok := rowMap[colName]; ok {
					if val == nil {
						row[i] = ""
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				} else {
					row[i] = ""
				}
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}

			rowCount++
			// Flush periodically for streaming output
			if rowCount%100 == 0 {
				csvWriter.Flush()
			}
		}
	}
}

// outputJSON outputs data in JSON format
func (e *Executor) outputJSON(data [][]string) error {
	if len(data) <= 1 {
		fmt.Fprintln(e.writer, "[]")
		return nil
	}

	headers := data[0]
	var jsonRows []map[string]string

	for i := 1; i < len(data); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(data[i]) {
				row[header] = data[i][j]
			}
		}
		jsonRows = append(jsonRows, row)
	}

	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonRows)
}

// outputJSONWithRawData outputs data in JSON format using raw data if available
func (e *Executor) outputJSONWithRawData(result db.QueryResult) error {
	if len(result.Data) <= 1 {
		fmt.Fprintln(e.writer, "[]")
		return nil
	}

	// Use raw data if available to preserve types
	if len(result.RawData) > 0 {
		encoder := json.NewEncoder(e.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result.RawData)
	}

	// Fall back to string conversion
	return e.outputJSON(result.Data)
}

// outputStreamingJSON outputs streaming data in JSON format
func (e *Executor) outputStreamingJSON(ctx context.Context, result db.StreamingQueryResult) error {
	fmt.Fprint(e.writer, "[")
	first := true

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(e.writer, "\n]")
			return nil
		default:
			rowMap := make(map[string]interface{})
			if !result.Iterator.MapScan(rowMap) {
				fmt.Fprintln(e.writer, "\n]")
				return result.Iterator.Close()
			}

			if !first {
				fmt.Fprint(e.writer, ",")
			}
			first = false

			// Convert to JSON
			jsonBytes, err := json.Marshal(rowMap)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}

			fmt.Fprintf(e.writer, "\n  %s", string(jsonBytes))
		}
	}
}