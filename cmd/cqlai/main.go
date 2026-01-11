package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/axonops/cqlai/internal/batch"
	"github.com/axonops/cqlai/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// Version is set via ldflags at build time
var Version = "dev"

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
		configFile     string
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
	pflag.StringVar(&configFile, "config-file", "", "Path to config file (overrides default locations)")

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
		fmt.Printf("cqlai version %s\n", Version)
		os.Exit(0)
	}

	// Override with environment variables if command-line flags not set
	// This allows users to set CQLAI_* env vars as an alternative to flags
	if configFile == "" {
		if envConfigFile := os.Getenv("CQLAI_CONFIG_FILE"); envConfigFile != "" {
			configFile = envConfigFile
		}
	}
	if host == "" {
		if envHost := os.Getenv("CQLAI_HOST"); envHost != "" {
			host = envHost
		}
	}
	if port == 0 {
		if envPort := os.Getenv("CQLAI_PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil {
				port = p
			}
		}
	}
	if keyspace == "" {
		if envKeyspace := os.Getenv("CQLAI_KEYSPACE"); envKeyspace != "" {
			keyspace = envKeyspace
		}
	}
	if username == "" {
		if envUsername := os.Getenv("CQLAI_USERNAME"); envUsername != "" {
			username = envUsername
		}
	}
	if !debug {
		if envDebug := os.Getenv("CQLAI_DEBUG"); envDebug != "" {
			debug = envDebug == "true" || envDebug == "1"
		}
	}
	if connectTimeout == 10 { // Check if still at default
		if envTimeout := os.Getenv("CQLAI_CONNECT_TIMEOUT"); envTimeout != "" {
			if t, err := strconv.Atoi(envTimeout); err == nil {
				connectTimeout = t
			}
		}
	}
	if requestTimeout == 10 { // Check if still at default
		if envTimeout := os.Getenv("CQLAI_REQUEST_TIMEOUT"); envTimeout != "" {
			if t, err := strconv.Atoi(envTimeout); err == nil {
				requestTimeout = t
			}
		}
	}
	if !noConfirm {
		if envNoConfirm := os.Getenv("CQLAI_NO_CONFIRM"); envNoConfirm != "" {
			noConfirm = envNoConfirm == "true" || envNoConfirm == "1"
		}
	}
	// Batch mode environment variables
	if execute == "" {
		if envExecute := os.Getenv("CQLAI_EXECUTE"); envExecute != "" {
			execute = envExecute
		}
	}
	if executeFile == "" {
		if envFile := os.Getenv("CQLAI_FILE"); envFile != "" {
			executeFile = envFile
		}
	}
	if format == "ascii" { // Check if still at default
		if envFormat := os.Getenv("CQLAI_FORMAT"); envFormat != "" {
			format = envFormat
		}
	}
	if !noHeader {
		if envNoHeader := os.Getenv("CQLAI_NO_HEADER"); envNoHeader != "" {
			noHeader = envNoHeader == "true" || envNoHeader == "1"
		}
	}
	if fieldSep == "," { // Check if still at default
		if envFieldSep := os.Getenv("CQLAI_FIELD_SEPARATOR"); envFieldSep != "" {
			fieldSep = envFieldSep
		}
	}
	if pageSize == 100 { // Check if still at default
		if envPageSize := os.Getenv("CQLAI_PAGE_SIZE"); envPageSize != "" {
			if ps, err := strconv.Atoi(envPageSize); err == nil {
				pageSize = ps
			}
		}
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
		ConfigFile:          configFile,
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
			// Execute command from -e flag (with multi-statement support)
			err = executor.ExecuteMulti(execute)
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

		// Create program with alternate screen buffer (like less) and mouse support
		// This hides the terminal scrollbar and provides a clean full-screen experience
		p := tea.NewProgram(m,
			tea.WithAltScreen(),
			// Don't use WithMouseCellMotion - we'll enable mouse manually in Init
		)

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
