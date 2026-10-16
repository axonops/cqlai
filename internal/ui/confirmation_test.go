package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// TestAFileWithoutTheKeyStillConfirms: confirmation is on unless the file
// turns it off. CONNECT and PREFERENCES read the file into a zero Config, so a
// file without "requireConfirmation" turned it off - and DROP ran unasked.
func TestAFileWithoutTheKeyStillConfirms(t *testing.T) {
	for _, tt := range []struct {
		file string
		want bool
	}{
		{`{"host": "127.0.0.1"}`, true},
		{`{"host": "127.0.0.1", "requireConfirmation": true}`, true},
		{`{"host": "127.0.0.1", "requireConfirmation": false}`, false},
		{`{"port": "not a number"}`, true},
	} {
		path := filepath.Join(t.TempDir(), "cqlai.json")
		require.NoError(t, os.WriteFile(path, []byte(tt.file), 0o600))
		m := helpModel()
		m.config = &config.Config{SourcePath: path}

		cfg := m.prefsFile()
		assert.Equal(t, tt.want, cfg.RequireConfirmation, tt.file)

		m.adoptSession(&db.Session{}, cfg)
		assert.Equal(t, tt.want, m.sessionManager.RequireConfirmation(), "after CONNECT, %s", tt.file)
	}

	path := filepath.Join(t.TempDir(), "missing.json")
	m := helpModel()
	m.config = &config.Config{SourcePath: path}
	assert.True(t, m.prefsFile().RequireConfirmation, "no file at all")
}
