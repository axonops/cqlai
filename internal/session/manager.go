package session

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/axonops/cqlai/internal/config"
)

// validKeyspacePattern matches valid Cassandra keyspace names
// Keyspace names must start with a letter or underscore, followed by alphanumerics/underscores
var validKeyspacePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Manager handles application-level session state
// This is separate from the database session
type Manager struct {
	mu                  sync.RWMutex
	currentKeyspace     string
	requireConfirmation bool
	outputFormat        config.OutputFormat
}

// NewManager creates a new session manager
func NewManager(cfg *config.Config) *Manager {
	outputFormat := config.OutputFormatTable // Default

	// Read output format from config if specified
	if cfg != nil && cfg.OutputFormat != "" {
		if parsed, err := config.ParseOutputFormat(cfg.OutputFormat); err == nil {
			outputFormat = parsed
		}
	}

	keyspace := ""
	if cfg != nil && cfg.Keyspace != "" {
		keyspace = cfg.Keyspace
	}

	return &Manager{
		currentKeyspace:     keyspace,
		requireConfirmation: cfg != nil && cfg.RequireConfirmation,
		outputFormat:        outputFormat,
	}
}

// CurrentKeyspace returns the current keyspace
func (m *Manager) CurrentKeyspace() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentKeyspace
}

// SetKeyspace sets the current keyspace
// Returns an error if the keyspace name is invalid
func (m *Manager) SetKeyspace(keyspace string) error {
	// Empty keyspace is allowed (clears the current keyspace)
	if keyspace != "" && !validKeyspacePattern.MatchString(keyspace) {
		return fmt.Errorf("invalid keyspace name: %q (must start with letter/underscore, contain only alphanumerics/underscores)", keyspace)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentKeyspace = keyspace
	return nil
}

// RequireConfirmation returns whether confirmation is required for dangerous commands
func (m *Manager) RequireConfirmation() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requireConfirmation
}

// SetRequireConfirmation sets whether confirmation is required
func (m *Manager) SetRequireConfirmation(require bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requireConfirmation = require
}

// GetOutputFormat returns the current output format
func (m *Manager) GetOutputFormat() config.OutputFormat {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.outputFormat
}

// SetOutputFormat sets the output format
// Returns an error if the format is not valid
func (m *Manager) SetOutputFormat(format config.OutputFormat) error {
	// Validate output format
	switch format {
	case config.OutputFormatTable, config.OutputFormatASCII, config.OutputFormatExpand, config.OutputFormatJSON:
		// Valid format
	default:
		return fmt.Errorf("invalid output format: %q (valid formats: TABLE, ASCII, EXPAND, JSON)", format)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputFormat = format
	return nil
}
