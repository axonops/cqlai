package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAFileCqlaiCannotReadIsNotSavedOver: one value of the wrong type meant
// cqlai started on its defaults, and saving those over the file deleted every
// setting in it - the password and the saved connections with them.
func TestAFileCqlaiCannotReadIsNotSavedOver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	original := `{"host": "10.0.0.5", "port": "9042", "password": "s3cret", "connections": [{"name": "prod", "host": "10.0.0.6"}]}`
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	_, err := (&Config{SourcePath: path, Host: "127.0.0.1"}).Save()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "was not saved over")
	got, _ := os.ReadFile(path)
	assert.Equal(t, original, string(got), "the file is as it was")
}

// TestSavingMakesTheFileReadableOnlyByItsOwner, whatever it was: it holds
// passwords, and os.WriteFile leaves an existing file's mode alone.
func TestSavingMakesTheFileReadableOnlyByItsOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"host": "10.0.0.5", "custom": "keep"}`), 0o644))

	_, err := (&Config{SourcePath: path, Host: "10.0.0.5", Password: "s3cret"}).Save()
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	got, _ := os.ReadFile(path)
	assert.Contains(t, string(got), `"custom": "keep"`, "a key cqlai does not know survives")
	assert.Contains(t, string(got), `"password": "s3cret"`)

	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".cqlai.json.*.tmp"))
	assert.Empty(t, leftovers, "the file it was written through is gone")
}

// TestSavingALinkWritesWhatItLinksTo, and keeps the link.
func TestSavingALinkWritesWhatItLinksTo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dotfiles-cqlai.json")
	link := filepath.Join(dir, "cqlai.json")
	require.NoError(t, os.WriteFile(target, []byte(`{"host": "a"}`), 0o600))
	require.NoError(t, os.Symlink(target, link))

	_, err := (&Config{SourcePath: link, Host: "b"}).Save()
	require.NoError(t, err)

	info, err := os.Lstat(link)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink, "still a link")
	got, _ := os.ReadFile(target)
	assert.Contains(t, string(got), `"host": "b"`)
}
