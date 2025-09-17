package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/batch"
	"github.com/axonops/cqlai/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

func main() {
	// Parse command-line flags using pflag for POSIX/GNU-style flags
	var (
		host           string
		port           int
		keyspace       string
		username       string
		password       string
		noConfirm      bool
		connectTimeout int
		requestTimeout int
		debug          bool
		execute        string
		executeFile    string
		format         string
		noHeader       bool
		fieldSep       string
		pageSize       int
		version        bool
		help           bool
	)

	// Connection flags
	pflag.StringVar(&host, "host", "", "Cassandra host (overrides config)")
	pflag.IntVar(&port, "port", 0, "Cassandra port (overrides config)")
	pflag.StringVarP(&keyspace, "keyspace", "k", "", "Default keyspace (overrides config)")
	pflag.StringVarP(&username, "username", "u", "", "Username for authentication (overrides config)")
	pflag.StringVarP(&password, "password", "p", "", "Password for authentication (overrides config)")
	pflag.BoolVar(&noConfirm, "no-confirm", false, "Disable confirmation prompts for dangerous commands")
	pflag.IntVar(&connectTimeout, "connect-timeout", 10, "Connection timeout in seconds")
	pflag.IntVar(&requestTimeout, "request-timeout", 10, "Request timeout in seconds")
	pflag.BoolVar(&debug, "debug", false, "Enable debug logging")

	// Batch mode flags (compatible with cqlsh)
	pflag.StringVarP(&execute, "execute", "e", "", "Execute CQL statement and exit")
	pflag.StringVarP(&executeFile, "file", "f", "", "Execute CQL from file and exit")
	pflag.StringVar(&format, "format", "ascii", "Output format: ascii, json, csv, table")
	pflag.BoolVar(&noHeader, "no-header", false, "Don't output column headers (CSV format)")
	pflag.StringVar(&fieldSep, "field-separator", ",", "Field separator for CSV output")
	pflag.IntVar(&pageSize, "page-size", 100, "Pagination size for batch mode")

	// Version and help flags
	pflag.BoolVarP(&version, "version", "v", false, "Print version and exit")
	pflag.BoolVarP(&help, "help", "h", false, "Show help message")

	pflag.Parse()

	// Handle help flag
	if help {
		fmt.Println("cqlai - A modern Cassandra CQL shell with AI assistance")
		fmt.Println()
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// Handle version flag
	if version {
		fmt.Println("cqlai version 0.0.6")
		os.Exit(0)
	}

	// Handle password prompting if username provided without password
	if username != "" && password == "" && isTerminal() {
		fmt.Fprintf(os.Stderr, "Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // Print newline after password input
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}
		password = string(passwordBytes)
	}

	// Also check environment variable as fallback
	if password == "" {
		if envPass := os.Getenv("CQLAI_PASSWORD"); envPass != "" {
			password = envPass
		}
	}

	// Create connection options
	connOptions := ui.ConnectionOptions{
		Host:                host,
		Port:                port,
		Keyspace:            keyspace,
		Username:            username,
		Password:            password,
		RequireConfirmation: !noConfirm,
		ConnectTimeout:      connectTimeout,
		RequestTimeout:      requestTimeout,
		Debug:               debug,
	}

	// Check if we're in batch mode
	isBatchMode := execute != "" || executeFile != "" || !isTerminal()

	if isBatchMode {
		// Batch mode execution
		batchOptions := &batch.Options{
			Execute:     execute,
			File:        executeFile,
			Format:      batch.OutputFormat(strings.ToLower(format)),
			NoHeader:    noHeader,
			FieldSep:    fieldSep,
			PageSize:    pageSize,
			ConnOptions: connOptions,
		}

		executor, err := batch.NewExecutor(batchOptions, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		// Execute based on input source
		if execute != "" { //nolint:gocritic // more readable as if
			// Execute command from -e flag
			err = executor.Execute(execute)
		} else if executeFile != "" {
			// Execute from file
			err = executor.ExecuteFile(executeFile)
		} else {
			// Execute from stdin
			err = executor.ExecuteStdin()
		}

		_ = executor.Close() // Error already handled in defer
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Interactive mode - use Bubble Tea UI
		m, err := ui.NewMainModelWithConnectionOptions(connOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating model: %v\n", err)
			os.Exit(1)
		}

		p := tea.NewProgram(m)

		if _, err := p.Run(); err != nil {
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
