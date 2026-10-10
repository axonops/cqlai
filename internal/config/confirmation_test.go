package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheConfirmationIsOnUnlessTheFileTurnsItOff: a file without the key -
// most of them - left it off, though it is documented as on.
func TestTheConfirmationIsOnUnlessTheFileTurnsItOff(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // no cqlshrc of the person running the tests
	dir := t.TempDir()

	without := filepath.Join(dir, "without.json")
	require.NoError(t, os.WriteFile(without, []byte(`{"host": "10.0.0.5"}`), 0o600))
	cfg, err := LoadConfig(without)
	require.NoError(t, err)
	assert.True(t, cfg.RequireConfirmation)

	off := filepath.Join(dir, "off.json")
	require.NoError(t, os.WriteFile(off, []byte(`{"host": "10.0.0.5", "requireConfirmation": false}`), 0o600))
	cfg, err = LoadConfig(off)
	require.NoError(t, err)
	assert.False(t, cfg.RequireConfirmation)

	// Saved off, it is read back off, not on again.
	_, err = cfg.Save()
	require.NoError(t, err)
	data, _ := os.ReadFile(off)
	assert.Contains(t, string(data), `"requireConfirmation": false`)
	cfg, err = LoadConfig(off)
	require.NoError(t, err)
	assert.False(t, cfg.RequireConfirmation)
}
