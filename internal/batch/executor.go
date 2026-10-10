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
	Execute     string       // CQL to execute directly (-e flag)
	File        string       // CQL file to execute (-f flag)
	Format      OutputFormat // Output format
	NoHeader    bool         // Skip headers in output
	FieldSep    string       // Field separator for CSV
	NoPager     bool         // Disable paging (print all results)
	PageSize    int          // Number of rows per batch for streaming
	ConnOptions ui.ConnectionOptions
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
	cfg, err := config.LoadConfig(options.ConnOptions.ConfigFile)
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
	// Override consistency from CLI flag
	if options.ConnOptions.Consistency != "" {
		cfg.Consistency = options.ConnOptions.Consistency
	}
	// Enable SSL from CLI flag (--ssl)
	if options.ConnOptions.SSL {
		if cfg.SSL == nil {
			cfg.SSL = &config.SSLConfig{}
		}
		cfg.SSL.Enabled = true
	}
	// Override SSL host verification and insecure skip verify from CLI flags
	if options.ConnOptions.SSLHostVerification != nil || options.ConnOptions.SSLInsecureSkipVerify != nil {
		if cfg.SSL == nil {
			cfg.SSL = &config.SSLConfig{}
		}
		if options.ConnOptions.SSLHostVerification != nil {
			cfg.SSL.HostVerification = *options.ConnOptions.SSLHostVerification
		}
		if options.ConnOptions.SSLInsecureSkipVerify != nil {
			cfg.SSL.InsecureSkipVerify = *options.ConnOptions.SSLInsecureSkipVerify
		}
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
		Consistency:    cfg.Consistency,
		SSL:            cfg.SSL,
		BatchMode:      true, // Disable schema caching in batch mode
		ConnectTimeout: options.ConnOptions.ConnectTimeout,
		RequestTimeout: options.ConnOptions.RequestTimeout,
		ConfigFile:     options.ConnOptions.ConfigFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	// Create session manager for tracking keyspace changes
	sessionMgr := session.NewManager(cfg)
	if cfg.Keyspace != "" {
		// SetKeyspace validates the keyspace name, but config values should already be valid
		_ = sessionMgr.SetKeyspace(cfg.Keyspace)
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

// ExecuteMulti runs multiple CQL statements using the robust CQL splitter
func (e *Executor) ExecuteMulti(cql string) error {
	// Split into individual statements using proper CQL tokenizer
	statements, err := SplitForNode(cql)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Execute each statement (SplitForNode already trims and filters empty statements)
	for _, stmt := range statements {
		if err := e.Execute(stmt); err != nil {
			return err
		}
	}

	return nil
}

// Execute runs a single CQL statement
func (e *Executor) Execute(cql string) error {
	// Set up signal handling for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C stops this statement. The handler goes when the statement does:
	// left registered, one per statement, a signal after the first statement
	// only cancelled a context nobody was using, and Ctrl+C and kill no longer
	// stopped a long -f run.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
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

			// Update the session manager (keyspace already validated by Cassandra)
			if e.sessionManager != nil {
				_ = e.sessionManager.SetKeyspace(keyspaceName)
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
		// For any other type, try to print it as a string
		if v != nil {
			fmt.Fprintf(e.writer, "%v\n", v)
		}
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

	// One statement at a time, split as the shell splits what is typed.
	for _, stmt := range splitStatements(string(content)) {
		if err := e.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteStdin executes CQL from stdin, a statement at a time: each runs once
// it is whole, by the rule the shell uses for what is typed - a semicolon
// outside any string or comment, and for a BATCH its APPLY BATCH;.
func (e *Executor) ExecuteStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var buffer strings.Builder

	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n")
		if !router.StatementComplete(buffer.String()) {
			continue
		}
		for _, stmt := range splitStatements(buffer.String()) {
			if err := e.Execute(stmt); err != nil {
				return err
			}
		}
		buffer.Reset()
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Whatever is left at the end, without its semicolon.
	for _, stmt := range splitStatements(buffer.String()) {
		if err := e.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
}
