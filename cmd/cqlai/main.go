package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/batch"
	"github.com/axonops/cqlai/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags
	host := flag.String("host", "", "Cassandra host (overrides config)")
	port := flag.Int("port", 0, "Cassandra port (overrides config)")
	keyspace := flag.String("keyspace", "", "Default keyspace (overrides config)")
	username := flag.String("username", "", "Username for authentication (overrides config)")
	password := flag.String("password", "", "Password for authentication (overrides config)")
	noConfirm := flag.Bool("no-confirm", false, "Disable confirmation prompts for dangerous commands")
	
	// Batch mode flags (compatible with cqlsh)
	execute := flag.String("e", "", "Execute CQL statement and exit")
	executeFile := flag.String("f", "", "Execute CQL from file and exit")
	format := flag.String("format", "ascii", "Output format: ascii, json, csv, table (default: ascii)")
	noHeader := flag.Bool("no-header", false, "Don't output column headers (CSV format)")
	fieldSep := flag.String("field-separator", ",", "Field separator for CSV output")
	pageSize := flag.Int("page-size", 100, "Pagination size for batch mode (default: 100)")
	
	// Version flag
	version := flag.Bool("version", false, "Print version and exit")
	
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Println("cqlai version 1.0.0")
		os.Exit(0)
	}

	// Create connection options
	connOptions := ui.ConnectionOptions{
		Host:                *host,
		Port:                *port,
		Keyspace:            *keyspace,
		Username:            *username,
		Password:            *password,
		RequireConfirmation: !*noConfirm,
	}

	// Check if we're in batch mode
	isBatchMode := *execute != "" || *executeFile != "" || !isTerminal()

	if isBatchMode {
		// Batch mode execution
		batchOptions := &batch.Options{
			Execute:     *execute,
			File:        *executeFile,
			Format:      batch.OutputFormat(strings.ToLower(*format)),
			NoHeader:    *noHeader,
			FieldSep:    *fieldSep,
			PageSize:    *pageSize,
			ConnOptions: connOptions,
		}

		executor, err := batch.NewExecutor(batchOptions, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer executor.Close()

		// Execute based on input source
		if *execute != "" {
			// Execute command from -e flag
			if err := executor.Execute(*execute); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else if *executeFile != "" {
			// Execute from file
			if err := executor.ExecuteFile(*executeFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Execute from stdin
			if err := executor.ExecuteStdin(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Interactive mode - use Bubble Tea UI
		m, err := ui.NewMainModelWithConnectionOptions(connOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating model: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(m)

		if err := p.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting program: %v", err)
			os.Exit(1)
		}
	}
}

// isTerminal checks if stdin is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return true // Assume terminal if we can't stat
	}
	return fileInfo.Mode()&os.ModeCharDevice != 0
}
