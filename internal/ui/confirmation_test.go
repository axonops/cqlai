package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
)

// TestNoConfirmCountsWhateverElseIsGiven: it used to count only beside a
// connection flag.
func TestNoConfirmCountsWhateverElseIsGiven(t *testing.T) {
	cfg := &config.Config{RequireConfirmation: true}
	applyConfirmation(cfg, ConnectionOptions{RequireConfirmation: false})
	assert.False(t, cfg.RequireConfirmation, "--no-confirm on its own")

	cfg = &config.Config{RequireConfirmation: false}
	applyConfirmation(cfg, ConnectionOptions{RequireConfirmation: true, Host: "10.0.0.5"})
	assert.False(t, cfg.RequireConfirmation, "a host flag no longer overrides the file")

	cfg = &config.Config{RequireConfirmation: true}
	applyConfirmation(cfg, ConnectionOptions{RequireConfirmation: true})
	assert.True(t, cfg.RequireConfirmation)
}

// TestNoConfirmSurvivesConnecting: CONNECT builds the session manager from the
// file again.
func TestNoConfirmSurvivesConnecting(t *testing.T) {
	m := helpModel()
	m.noConfirm = true
	m.adoptSession(&db.Session{}, &config.Config{RequireConfirmation: true})
	assert.False(t, m.sessionManager.RequireConfirmation())

	m = helpModel()
	m.adoptSession(&db.Session{}, &config.Config{RequireConfirmation: true})
	assert.True(t, m.sessionManager.RequireConfirmation())
}
