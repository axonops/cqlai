package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pathFixture builds a directory to complete against.
func pathFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, name := range []string{"report.csv", "report.json", "results.csv", ".hidden"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}
	require.NoError(t, os.Mkdir(filepath.Join(dir, "exports"), 0o750))
	return dir
}

// TestOneMatchIsCompletedWhole.
func TestOneMatchIsCompletedWhole(t *testing.T) {
	dir := pathFixture(t)

	got := completePath(filepath.Join(dir, "results"))

	assert.Equal(t, filepath.Join(dir, "results.csv"), got.Completed)
	assert.Equal(t, []string{"results.csv"}, got.Matches)
}

// TestSeveralMatchesFillInAsFarAsTheyAgree, which is what a shell does.
func TestSeveralMatchesFillInAsFarAsTheyAgree(t *testing.T) {
	dir := pathFixture(t)

	got := completePath(filepath.Join(dir, "rep"))

	assert.Equal(t, filepath.Join(dir, "report."), got.Completed,
		"it should fill in up to where the names differ")
	assert.Equal(t, []string{"report.csv", "report.json"}, got.Matches)
}

// TestADirectoryEndsWithASeparator, so the next Tab carries on inside it.
func TestADirectoryEndsWithASeparator(t *testing.T) {
	dir := pathFixture(t)

	got := completePath(filepath.Join(dir, "exp"))

	assert.Equal(t, filepath.Join(dir, "exports")+string(filepath.Separator), got.Completed)
	assert.Equal(t, []string{"exports" + string(filepath.Separator)}, got.Matches)
}

// TestEverythingInADirectoryIsOffered when nothing has been typed after it.
func TestEverythingInADirectoryIsOffered(t *testing.T) {
	dir := pathFixture(t)

	got := completePath(dir + string(filepath.Separator))

	assert.Len(t, got.Matches, 4, "three files and a directory, not the hidden one: %v", got.Matches)
	assert.Contains(t, got.Matches, "exports"+string(filepath.Separator))
}

// TestHiddenFilesOnlyWhenAskedFor, as a shell does.
func TestHiddenFilesOnlyWhenAskedFor(t *testing.T) {
	dir := pathFixture(t)

	assert.NotContains(t, completePath(dir+string(filepath.Separator)).Matches, ".hidden")

	// Built by hand: filepath.Join would clean the "." away.
	got := completePath(dir + string(filepath.Separator) + ".")
	assert.Equal(t, []string{".hidden"}, got.Matches)
}

// TestNothingMatchingLeavesTheInputAlone rather than emptying what was typed.
func TestNothingMatchingLeavesTheInputAlone(t *testing.T) {
	dir := pathFixture(t)
	input := filepath.Join(dir, "nothing-like-this")

	got := completePath(input)

	assert.Equal(t, input, got.Completed)
	assert.Empty(t, got.Matches)
}

// TestAnUnreadableDirectoryOffersNothing rather than failing. Tab is not the
// place to learn that a path is wrong.
func TestAnUnreadableDirectoryOffersNothing(t *testing.T) {
	input := filepath.Join(t.TempDir(), "no", "such", "dir", "x")

	got := completePath(input)

	assert.Equal(t, input, got.Completed)
	assert.Empty(t, got.Matches)
}

// TestTheRestOfThePathIsNotRewritten: completing the last part must leave the
// directory in front of it exactly as it was typed.
func TestTheRestOfThePathIsNotRewritten(t *testing.T) {
	dir := pathFixture(t)
	messy := dir + string(filepath.Separator) + "." + string(filepath.Separator) + "results"

	got := completePath(messy)

	assert.True(t, len(got.Completed) > len(messy), "it should have completed something")
	assert.Contains(t, got.Completed, "."+string(filepath.Separator),
		"the path in front should be untouched: %q", got.Completed)
}

// TestHomeIsExpanded, so ~/ completes against the home directory.
func TestHomeIsExpanded(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, home, expandHome("~"))
	assert.Equal(t, home, expandHome("~"+string(filepath.Separator)))
	assert.Equal(t, filepath.Join(home, "docs"), expandHome("~"+string(filepath.Separator)+"docs"))

	// Everything else is left alone.
	assert.Equal(t, ".", expandHome(""))
	assert.Equal(t, "/etc", expandHome("/etc"))
	assert.Equal(t, "~notauser", expandHome("~notauser"))
}

func TestCommonPrefix(t *testing.T) {
	assert.Equal(t, "report.", commonPrefix([]string{"report.csv", "report.json"}))
	assert.Equal(t, "only", commonPrefix([]string{"only"}))
	assert.Equal(t, "", commonPrefix([]string{"alpha", "beta"}))
	assert.Equal(t, "", commonPrefix(nil))
}
