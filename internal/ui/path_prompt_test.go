package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func promptModel(t *testing.T) *MainModel {
	t.Helper()

	return &MainModel{
		styles:          DefaultStyles(),
		windowWidth:     120,
		windowHeight:    30,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
}

// TestTabCompletesAPathAtThePrompt. Completion knew CAPTURE as a keyword and
// offered nothing after it, so paths had to be typed out in full.
func TestTabCompletesAPathAtThePrompt(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	for _, command := range []string{
		"CAPTURE JSON ",
		"CAPTURE CSV ",
		"SOURCE ",
		"SAVE ",
		"COPY users TO ",
		"COPY users FROM ",
	} {
		m := promptModel(t)
		m.input.SetValue(command + filepath.Join(dir, "res"))

		m.handleTabKey()

		assert.Equal(t, command+filepath.Join(dir, "results.csv"), m.input.Value(),
			"after %q", command)
	}
}

// TestTheQuoteIsKept, since these paths are usually written inside one.
func TestTheQuoteIsKept(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := promptModel(t)
	m.input.SetValue("CAPTURE '" + filepath.Join(dir, "res"))

	m.handleTabKey()

	assert.Equal(t, "CAPTURE '"+filepath.Join(dir, "results.csv"), m.input.Value())
}

// TestTheCandidatesAreOfferedWhenFillingInIsNotEnough.
func TestTheCandidatesAreOfferedWhenFillingInIsNotEnough(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"report.csv", "report.json"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}

	m := promptModel(t)
	m.input.SetValue("CAPTURE " + filepath.Join(dir, "rep"))

	m.handleTabKey()

	assert.True(t, strings.HasSuffix(m.input.Value(), "report."), "got %q", m.input.Value())
	assert.True(t, m.showCompletions)
	assert.Equal(t, []string{"report.csv", "report.json"}, m.completions)
}

// TestOneCandidateIsJustCompleted, with nothing to choose from.
func TestOneCandidateIsJustCompleted(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "only.csv"), nil, 0o600))

	m := promptModel(t)
	m.input.SetValue("SOURCE " + filepath.Join(dir, "on"))

	m.handleTabKey()

	assert.False(t, m.showCompletions)
	assert.True(t, strings.HasSuffix(m.input.Value(), "only.csv"))
}

// TestSelectIsNotAPath: COPY's FROM takes a file, SELECT's does not, and Tab
// there must still complete schema.
func TestSelectIsNotAPath(t *testing.T) {
	m := promptModel(t)
	m.input.SetValue("SELECT * FROM sys")

	_, _, handled := m.completePathAtPrompt(m.input.Value())

	assert.False(t, handled)
}

// TestBareCaptureOpensTheWindow rather than reporting a status the bottom line
// already shows.
func TestBareCaptureOpensTheWindow(t *testing.T) {
	m := promptModel(t)
	m.statusBar = testStatusBar()

	for _, command := range []string{"CAPTURE", "capture", "Capture"} {
		m.capture = capturePanel{}
		_, _, handled := m.handleSpecialCommands(command)

		assert.True(t, handled, "%q should be taken here", command)
		assert.True(t, m.capture.active, "%q should open the window", command)
		assert.Equal(t, captureChooseFormat, m.capture.step)
	}
}

// TestCaptureWithArgumentsStillGoesToTheCommand.
func TestCaptureWithArgumentsStillGoesToTheCommand(t *testing.T) {
	m := promptModel(t)

	for _, command := range []string{"CAPTURE OFF", "CAPTURE 'out.txt'", "CAPTURE JSON 'a.json'"} {
		_, _, handled := m.handleSpecialCommands(command)

		assert.False(t, handled, "%q belongs to the command, not the window", command)
		assert.False(t, m.capture.active)
	}
}

// TestChoosingAFilenameKeepsTheDirectory.
//
// The candidates are bare names, and applying one the way a CQL word is
// applied replaces the last word - which for "/tmp/rep" would lose the "/tmp".
func TestChoosingAFilenameKeepsTheDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"report.csv", "report.json"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV " + filepath.Join(dir, "rep"))
	m.handleTabKey()
	require.True(t, m.completingPath, "these are filenames")
	require.Len(t, m.completions, 2)

	m.completionIndex = 1 // report.json
	m.handleCompletionSelection()

	assert.Equal(t, "CAPTURE CSV "+filepath.Join(dir, "report.json"), m.input.Value())
	assert.False(t, m.showCompletions, "choosing one should put the list away")
}

// TestChoosingAFilenameKeepsTheQuote too.
func TestChoosingAFilenameKeepsTheQuote(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"report.csv", "report.json"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV '" + filepath.Join(dir, "rep"))
	m.handleTabKey()
	require.Len(t, m.completions, 2)

	m.handleCompletionSelection()

	assert.Equal(t, "CAPTURE CSV '"+filepath.Join(dir, "report.csv"), m.input.Value())
}

// TestChoosingADirectoryKeepsThePathGoing.
func TestChoosingADirectoryKeepsThePathGoing(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "exports"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "exported"), 0o750))

	m := promptModel(t)
	m.input.SetValue("SOURCE " + filepath.Join(dir, "exp"))
	m.handleTabKey()
	require.Len(t, m.completions, 2)

	// The first candidate alphabetically.
	require.Equal(t, "exported"+string(filepath.Separator), m.completions[0])
	m.handleCompletionSelection()

	assert.Equal(t, "SOURCE "+filepath.Join(dir, "exported")+string(filepath.Separator),
		m.input.Value())
}

// TestCQLWordsAreStillAppliedAsWords, not as paths.
func TestCQLWordsAreStillAppliedAsWords(t *testing.T) {
	m := promptModel(t)
	m.input.SetValue("SELECT * FROM ")
	m.completions = []string{"users"}
	m.completionIndex = 0
	m.showCompletions = true
	m.completingPath = false

	m.handleCompletionSelection()

	assert.Equal(t, "SELECT * FROM users", m.input.Value())
}

// TestNothingTypedSaysWhatIsExpected rather than listing the whole directory.
//
// After "CAPTURE CSV " there is nothing to complete against, and the working
// directory is noise - the binary and the log file are not what you are after.
func TestNothingTypedSaysWhatIsExpected(t *testing.T) {
	for _, command := range []string{"CAPTURE CSV ", "SOURCE ", "SAVE ", "COPY users TO "} {
		m := promptModel(t)
		m.input.SetValue(command)

		m.handleTabKey()

		assert.Equal(t, []string{pathHint}, m.completions, "after %q", command)
		assert.Equal(t, command, m.input.Value(), "%q should be left alone", command)
	}
}

// TestOnceSomethingIsTypedItCompletesAgain.
func TestOnceSomethingIsTypedItCompletesAgain(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV " + filepath.Join(dir, "res"))

	m.handleTabKey()

	assert.Equal(t, "CAPTURE CSV "+filepath.Join(dir, "results.csv"), m.input.Value())
}

// TestTheHintShowsWithOrWithoutASpaceAfterTheFormat.
func TestTheHintShowsWithOrWithoutASpaceAfterTheFormat(t *testing.T) {
	for _, command := range []string{"CAPTURE CSV", "CAPTURE CSV ", "capture json", "CAPTURE PARQUET"} {
		m := promptModel(t)
		m.input.SetValue(command)

		m.handleTabKey()

		assert.Equal(t, []string{pathHint}, m.completions, "after %q", command)
		assert.Equal(t, command, m.input.Value(), "%q should be left alone", command)
	}
}

// TestAPartialFormatStillCompletesTheFormat.
func TestAPartialFormatStillCompletesTheFormat(t *testing.T) {
	m := promptModel(t)
	m.input.SetValue("CAPTURE PARQ")

	_, _, handled := m.completePathAtPrompt(m.input.Value())

	assert.False(t, handled, "PARQ is a format being typed, not a file")
}

// TestChoosingTheHintStartsTheFilename. A row you cannot select is worse than
// no row at all.
func TestChoosingTheHintStartsTheFilename(t *testing.T) {
	for _, command := range []string{"CAPTURE CSV", "CAPTURE CSV "} {
		m := promptModel(t)
		m.input.SetValue(command)
		m.handleTabKey()
		require.Equal(t, []string{pathHint}, m.completions, "after %q", command)

		m.handleCompletionSelection()

		assert.Equal(t, "CAPTURE CSV ''", m.input.Value(), "after %q", command)
		assert.Equal(t, len("CAPTURE CSV '"), m.input.Position(),
			"the cursor should be between the quotes")
		assert.False(t, m.showCompletions, "the list should have gone")
	}
}

// TestInsideTheQuoteItListsTheDirectory: by then you have committed to typing
// a path, so the hint would be in the way.
func TestInsideTheQuoteItListsTheDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.csv", "b.csv"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV '" + dir + string(filepath.Separator))

	m.handleTabKey()

	assert.Equal(t, []string{parentEntry, "a.csv", "b.csv"}, m.completions,
		"the files, and the way up above them")
	assert.NotContains(t, m.completions, pathHint)
}

// TestCompletingBetweenTheQuotesKeepsThem, so the command stays valid.
func TestCompletingBetweenTheQuotesKeepsThem(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV '" + filepath.Join(dir, "res") + "'")

	m.handleTabKey()

	assert.Equal(t, "CAPTURE CSV '"+filepath.Join(dir, "results.csv")+"'", m.input.Value())
}

// TestTheWholeFlowFromTheFormatToTheFile.
func TestTheWholeFlowFromTheFormatToTheFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.csv"), nil, 0o600))

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV")

	// The hint, then the quotes.
	m.handleTabKey()
	require.Equal(t, []string{pathHint}, m.completions)
	m.handleCompletionSelection()
	require.Equal(t, "CAPTURE CSV ''", m.input.Value())

	// Type a path between them, then complete it.
	m.input.SetValue("CAPTURE CSV '" + filepath.Join(dir, "res") + "'")
	m.handleTabKey()

	assert.Equal(t, "CAPTURE CSV '"+filepath.Join(dir, "results.csv")+"'", m.input.Value())
}

// TestTheWayUpWorksAtThePromptToo, not only in the windows.
func TestTheWayUpWorksAtThePromptToo(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "exports")
	require.NoError(t, os.Mkdir(inner, 0o750))

	m := promptModel(t)
	m.input.SetValue("CAPTURE CSV '" + inner + string(filepath.Separator))
	m.handleTabKey()

	require.Equal(t, parentEntry, m.completions[0])
	m.applyPathCompletion(parentEntry)

	assert.Contains(t, m.input.Value(), dir+string(filepath.Separator))
	assert.NotContains(t, m.input.Value(), parentEntry, "it goes up rather than appending")
}
