package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheHistoryIsKeptInCassandra, beside cqlsh_history, not in a hidden
// directory of its own.
func TestTheHistoryIsKeptInCassandra(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	hm, err := NewHistoryManager()
	require.NoError(t, err)
	require.NoError(t, hm.SaveCommand("SELECT * FROM system.local;"))
	ai, err := NewAIHistoryManager()
	require.NoError(t, err)
	require.NoError(t, ai.SaveCommand("which tables are biggest?"))

	assert.FileExists(t, filepath.Join(home, ".cassandra", "cqlai_history"))
	assert.FileExists(t, filepath.Join(home, ".cassandra", "cqlai_ai_history"))
	assert.NoDirExists(t, filepath.Join(home, ".cqlai"))
}

// TestTheOldHistoryIsMovedWithWhatItHolds, and ~/.cqlai goes once it is empty.
func TestTheOldHistoryIsMovedWithWhatItHolds(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := filepath.Join(home, ".cqlai")
	require.NoError(t, os.MkdirAll(old, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(old, "history"), []byte("SELECT 1;\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(old, "ai_history"), []byte("hello\n"), 0o600))

	hm, err := NewHistoryManager()
	require.NoError(t, err)
	assert.Equal(t, []string{"SELECT 1;"}, hm.GetHistory())
	assert.DirExists(t, old, "ai_history is still there")

	ai, err := NewAIHistoryManager()
	require.NoError(t, err)
	assert.Equal(t, []string{"hello"}, ai.GetHistory())
	assert.NoDirExists(t, old, "empty now, so it goes")
}

// TestAnotherFileInTheOldDirectoryKeepsIt: only an empty ~/.cqlai is removed.
func TestAnotherFileInTheOldDirectoryKeepsIt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := filepath.Join(home, ".cqlai")
	require.NoError(t, os.MkdirAll(old, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(old, "notes"), []byte("mine"), 0o600))

	_, err := NewHistoryManager()
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(old, "notes"))
}

// TestAHistoryFileInTheSettingsIsLeftWhereItIs.
func TestAHistoryFileInTheSettingsIsLeftWhereItIs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	custom := filepath.Join(home, "elsewhere", "history")

	hm, err := NewHistoryManagerWithPath(custom)
	require.NoError(t, err)
	require.NoError(t, hm.SaveCommand("SELECT 2;"))
	assert.FileExists(t, custom)
	assert.NoFileExists(t, filepath.Join(home, ".cassandra", "cqlai_history"))
}
