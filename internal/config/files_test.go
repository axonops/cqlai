package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMoveInMovesAnOldFile with what it holds, and makes the directory.
func TestMoveInMovesAnOldFile(t *testing.T) {
	home := t.TempDir()
	old := filepath.Join(home, ".cqlai_mcp_token")
	require.NoError(t, os.WriteFile(old, []byte("secret\n"), 0o600))
	path := filepath.Join(home, ".cassandra", "cqlai_mcp_token")

	require.NoError(t, MoveIn(path, old))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "secret\n", string(got))
	assert.NoFileExists(t, old)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "it keeps its permissions")
}

// TestMoveInLeavesBothWhenBothAreThere: the new one is used, and the old one
// is not thrown away.
func TestMoveInLeavesBothWhenBothAreThere(t *testing.T) {
	home := t.TempDir()
	old := filepath.Join(home, ".cqlai_mcp_token")
	path := filepath.Join(home, ".cassandra", "cqlai_mcp_token")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(old, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(path, []byte("new"), 0o600))

	require.NoError(t, MoveIn(path, old))

	got, _ := os.ReadFile(path)
	assert.Equal(t, "new", string(got))
	assert.FileExists(t, old)
}

// TestMoveInMakesTheDirectoryWhenThereIsNothingToMove.
func TestMoveInMakesTheDirectoryWhenThereIsNothingToMove(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".cassandra", "cqlai_mcp_audit.log")

	require.NoError(t, MoveIn(path, filepath.Join(home, ".cqlai_mcp_audit.log")))

	assert.DirExists(t, filepath.Dir(path))
	assert.NoFileExists(t, path, "nothing is made in it")
}

// TestDirIsCassandraInTheHomeDirectory, where cqlai.json and cqlshrc are.
func TestDirIsCassandraInTheHomeDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	assert.Equal(t, filepath.Join(home, ".cassandra"), Dir())
	assert.Equal(t, filepath.Dir(DefaultConfigPath()), Dir())
}
