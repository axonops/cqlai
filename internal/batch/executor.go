package batch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
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

// ExecuteFile executes CQL from a file
func (e *Executor) ExecuteFile(filename string) error {
	// Clean the filename to prevent path traversal
	cleanPath := filepath.Clean(filename)
	content, err := os.ReadFile(cleanPath) // #nosec G304 - file path is user input but cleaned
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Strip comments from the content
	content = []byte(stripComments(string(content)))

	// Split into individual statements
	statements := splitStatements(string(content))

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Execute the statement
		if err := e.Execute(stmt); err != nil {
			return err
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