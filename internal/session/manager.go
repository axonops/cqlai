package session

import (
	"sync"

	"github.com/axonops/cqlai/internal/config"
)

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
	if cfg != nil {
		// Could read from config if we add output format to config
	}

	return &Manager{
		currentKeyspace:     "",
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
func (m *Manager) SetKeyspace(keyspace string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentKeyspace = keyspace
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
func (m *Manager) SetOutputFormat(format config.OutputFormat) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputFormat = format
}
