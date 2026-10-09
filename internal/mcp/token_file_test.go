package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheTokenIsKeptInCassandra, not as a hidden file in the home directory.
func TestTheTokenIsKeptInCassandra(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	token, err := readToken()
	require.NoError(t, err)

	path := filepath.Join(home, ".cassandra", "cqlai_mcp_token")
	assert.Equal(t, path, TokenFile())
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, token, strings.TrimSpace(string(got)))
	assert.NoFileExists(t, filepath.Join(home, ".cqlai_mcp_token"))
}

// TestAnOldTokenIsMovedNotReplaced: the client was set up with it.
func TestAnOldTokenIsMovedNotReplaced(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := strings.Repeat("ab", 32)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".cqlai_mcp_token"), []byte(old+"\n"), 0o600))

	token, err := readToken()
	require.NoError(t, err)

	assert.Equal(t, old, token, "the client's configuration still works")
	assert.NoFileExists(t, filepath.Join(home, ".cqlai_mcp_token"))
	assert.FileExists(t, filepath.Join(home, ".cassandra", "cqlai_mcp_token"))
}
