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
		ssl            bool
		consistency    string
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
	pflag.BoolVar(&ssl, "ssl", false, "Enable SSL/TLS connection")
	pflag.StringVar(&consistency, "consistency", "", "Default consistency level (e.g., ONE, QUORUM, LOCAL_QUORUM)")
	pflag.StringVar(&configFile, "config-file", "", "Path to config file (overrides default locations)")

	// Batch mode flags (compatible with cqlsh)
	pflag.StringVarP(&execute, "execute", "e", "", "Execute CQL statement and exit")
	pflag.StringVarP(&executeFile, "file", "f", "", "Execute CQL from file and exit")
	pflag.StringVar(&format, "format", "ascii", "Output format: ascii, json, csv, table")
	pflag.BoolVar(&noHeader, "no-header", false, "Don't output column headers (CSV format)")
	pflag.StringVar(&fieldSep, "field-separator", ",", "Field separator for CSV output")
	pflag.IntVar(&pageSize, "page-size", 100, "Pagination size for batch mode")

	// Version and help flags
	pflag.BoolVarP(&version, "version", "V", false, "Print version and exit")
	pflag.BoolVarP(&help, "help", "h", false, "Show help message")

	pflag.Parse()

	// Handle positional arguments for cqlsh compatibility (cqlai [host] [port])
	args := pflag.Args()

	// 1. Guard Clause: Fail fast
	if len(args) > 2 {
		fmt.Fprintf(os.Stderr, "Error: unexpected positional arguments: %v\nUsage: cqlai [options] [host [port]]\n", args[2:])
		os.Exit(1)
	}

	// 2. Handle Host
	if len(args) >= 1 {
		if host != "" {
			fmt.Fprintf(os.Stderr, "Warning: positional argument %q ignored because --host was specified\n", args[0])
		} else {
			host = args[0]
		}
	}

	// 3. Handle Port
	if len(args) >= 2 {
		if port != 0 {
			fmt.Fprintf(os.Stderr, "Warning: positional argument %q ignored because --port was specified\n", args[1])
		} else if p, err := strconv.Atoi(args[1]); err == nil {
			port = p
		} else {
			fmt.Fprintf(os.Stderr, "Error: invalid port number %q\n", args[1])
			os.Exit(1)
		}
	}

	// Handle help flag
	if help {
		fmt.Println("cqlai - A modern Cassandra CQL shell with AI assistance")
		fmt.Println()
		fmt.Println("Usage: cqlai [options] [host [port]]")
		fmt.Println()
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// Handle version flag
	if version {
		fmt.Printf("cqlai version %s\n", Version)
		os.Exit(0)
	}

	// Validate --format value
	switch strings.ToLower(format) {
	case "ascii", "json", "csv", "table":
		// valid
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid output format %q (valid: ascii, json, csv, table)\n", format)
		os.Exit(1)
	}

	// Validate --consistency value if provided
	if consistency != "" {
		switch strings.ToUpper(consistency) {
		case "ANY", "ONE", "TWO", "THREE", "QUORUM", "ALL",
			"LOCAL_QUORUM", "EACH_QUORUM", "LOCAL_ONE", "SERIAL", "LOCAL_SERIAL":
			consistency = strings.ToUpper(consistency)
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid consistency level %q (valid: ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE, SERIAL, LOCAL_SERIAL)\n", consistency)
			os.Exit(1)
		}
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
	if !pflag.CommandLine.Changed("connect-timeout") {
		if envTimeout := os.Getenv("CQLAI_CONNECT_TIMEOUT"); envTimeout != "" {
			if t, err := strconv.Atoi(envTimeout); err == nil {
				connectTimeout = t
			}
		}
	}
	if !pflag.CommandLine.Changed("request-timeout") {
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
	if !pflag.CommandLine.Changed("format") {
		if envFormat := os.Getenv("CQLAI_FORMAT"); envFormat != "" {
			format = envFormat
		}
	}
	if !noHeader {
		if envNoHeader := os.Getenv("CQLAI_NO_HEADER"); envNoHeader != "" {
			noHeader = envNoHeader == "true" || envNoHeader == "1"
		}
	}
	if !pflag.CommandLine.Changed("field-separator") {
		if envFieldSep := os.Getenv("CQLAI_FIELD_SEPARATOR"); envFieldSep != "" {
			fieldSep = envFieldSep
		}
	}
	if !pflag.CommandLine.Changed("page-size") {
		if envPageSize := os.Getenv("CQLAI_PAGE_SIZE"); envPageSize != "" {
			if ps, err := strconv.Atoi(envPageSize); err == nil {
				pageSize = ps
			}
		}
	}

	// Check password environment variable before interactive prompt
	// Precedence: CLI flag (-p) > env var > interactive prompt
	if password == "" {
		if envPass := os.Getenv("CQLAI_PASSWORD"); envPass != "" {
			password = envPass
		}
	}

	// Prompt for password interactively only if still empty and username was provided
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
		SSL:                 ssl,
		Consistency:         consistency,
		PageSize:            pageSize,
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
