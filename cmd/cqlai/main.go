package main

import (
	"flag"
	"fmt"
	"os"

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
	flag.Parse()

	// Create options from flags
	options := ui.ConnectionOptions{
		Host:                *host,
		Port:                *port,
		Keyspace:            *keyspace,
		Username:            *username,
		Password:            *password,
		RequireConfirmation: !*noConfirm,
	}

	m, err := ui.NewMainModelWithConnectionOptions(options)
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
