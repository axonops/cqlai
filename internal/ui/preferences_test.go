package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// prefModel is an open preferences window on a terminal with room for it,
// reading a configuration file written for the test.
//
// The window shows the file rather than the configuration cqlai is running on,
// so the file is where a test puts what it wants to see.
func prefModel(t *testing.T, cfg *config.Config) *MainModel {
	t.Helper()

	path := cfg.SourcePath
	if path == "" {
		path = filepath.Join(t.TempDir(), "cqlai.json")
		cfg.SourcePath = path
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 44
	m.config = &config.Config{SourcePath: path}
	m.openPreferences()
	require.True(t, m.preferences.active)
	return m
}

// prefIndex is the position of a setting by the field it edits.
func prefIndex(t *testing.T, m *MainModel, path string) int {
	t.Helper()

	for i, field := range m.preferences.fields {
		if field.spec.path == path {
			return i
		}
	}
	t.Fatalf("no setting for %q", path)
	return -1
}

func press(m *MainModel, key string) *MainModel {
	updated, _ := m.handlePreferencesKey(tea.KeyPressMsg{Code: keyCode(key), Text: key})
	return updated
}

func keyCode(key string) rune {
	if key == " " {
		return ' '
	}
	return []rune(key)[0]
}

// TestEverySettingReachesAConfigField: the specs name fields by path, and a
// path that names nothing would be a setting you could type into that never
// saved anything.
func TestEverySettingReachesAConfigField(t *testing.T) {
	cfg := &config.Config{}
	for _, spec := range prefSpecs() {
		require.NotEmpty(t, spec.label, "a setting with no label")

		err := setPrefValue(cfg, spec.path, prefTestValue(spec))
		require.NoError(t, err, "setting %s", spec.path)

		assert.Equal(t, prefTestValue(spec), prefValue(cfg, spec.path), "reading %s back", spec.path)
	}
}

// prefTestValue is something a setting will accept.
func prefTestValue(spec prefSpec) string {
	switch spec.kind {
	case prefNumber:
		return "7"
	case prefYesNo:
		return "true"
	case prefChoice:
		return spec.choices()[0]
	}
	return "x"
}

// TestTheWindowShowsWhatIsConfigured.
func TestTheWindowShowsWhatIsConfigured(t *testing.T) {
	m := prefModel(t, &config.Config{
		Host:     "cassandra.example.com",
		Port:     9142,
		Debug:    true,
		Password: "hunter2",
		SSL:      &config.SSLConfig{Enabled: true, CAPath: "/etc/ca.pem"},
		AI:       &config.AIConfig{Provider: "anthropic"},
	})

	assert.Equal(t, "cassandra.example.com", m.preferences.fields[prefIndex(t, m, "Host")].value())
	assert.Equal(t, "9142", m.preferences.fields[prefIndex(t, m, "Port")].value())
	assert.Equal(t, "true", m.preferences.fields[prefIndex(t, m, "Debug")].value())
	assert.Equal(t, "/etc/ca.pem", m.preferences.fields[prefIndex(t, m, "SSL.CAPath")].value())
	assert.Equal(t, "anthropic", m.preferences.fields[prefIndex(t, m, "AI.Provider")].value())

	// An unset number is empty rather than 0: nothing here is meaningfully
	// zero, and a box saying 0 reads as a value someone chose.
	assert.Empty(t, m.preferences.fields[prefIndex(t, m, "PageSize")].value())

	// The password is in the window but not on the screen.
	password := m.preferences.fields[prefIndex(t, m, "Password")]
	assert.Equal(t, "hunter2", password.value())
	assert.NotContains(t, password.display(), "hunter2")
}

// TestSavingWritesTheFile, which is the point of the window.
func TestSavingWritesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"host":"old","port":9042}`), 0o600))

	m := prefModel(t, &config.Config{Host: "old", Port: 9042, SourcePath: path})
	m.setPrefField(prefIndex(t, m, "Host"), "new-host")
	m.setPrefField(prefIndex(t, m, "PageSize"), "500")
	m.preferences.fields[prefIndex(t, m, "SSL.Enabled")].yes = true

	m, _ = m.savePreferences()
	assert.False(t, m.preferences.active, "the window stays open after saving")
	assert.Contains(t, m.fullHistoryContent, path, "saving says nothing about where it went")

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)

	var written map[string]any
	require.NoError(t, json.Unmarshal(data, &written))
	assert.Equal(t, "new-host", written["host"])
	assert.Equal(t, float64(500), written["pageSize"])
	assert.Equal(t, true, written["ssl"].(map[string]any)["enabled"])
}

// TestSavingAddsNoEmptySections: opening the window makes an SSL block and five
// chat blocks so every setting has somewhere to go. A configuration that had
// none of them should not gain six empty objects for having been looked at.
func TestSavingAddsNoEmptySections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"host":"h"}`), 0o600))

	m := prefModel(t, &config.Config{Host: "h", SourcePath: path})
	if _, cmd := m.savePreferences(); cmd != nil {
		t.Fatal("saving asked for something to be run")
	}

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)

	var written map[string]any
	require.NoError(t, json.Unmarshal(data, &written))
	assert.NotContains(t, written, "ssl")
	assert.NotContains(t, written, "ai")
	assert.NotContains(t, written, "authProvider")
}

// TestSaveWaitsForSomethingItCanWrite.
func TestSaveWaitsForSomethingItCanWrite(t *testing.T) {
	m := prefModel(t, &config.Config{})
	path := m.preferences.path
	require.NoError(t, os.Remove(path))

	m.setPrefField(prefIndex(t, m, "Port"), "nine thousand")
	assert.False(t, m.preferencesReady())
	assert.True(t, m.prefWrong(prefIndex(t, m, "Port")))

	m, _ = m.savePreferences()
	assert.True(t, m.preferences.active, "the window closed on a value it could not write")
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "a file was written from a window that was not ready")

	m.setPrefField(prefIndex(t, m, "Port"), "9042")
	assert.True(t, m.preferencesReady())

	// A value that is not one of the ones offered is wrong in the same way.
	m.setPrefField(prefIndex(t, m, "Consistency"), "MAYBE")
	assert.False(t, m.preferencesReady())
	m.setPrefField(prefIndex(t, m, "Consistency"), "local_quorum")
	assert.True(t, m.preferencesReady(), "the values are not case sensitive")
}

// TestEverySettingCanBeClicked: every setting, on every section, focusable by
// clicking the row it is drawn on.
//
// This is the test the COPY FROM crash (#165) was missing. Reaching a field
// only through the fields a test happens to fill in leaves the ones nothing
// else touches untested, and one of those is where the crash was.
func TestEverySettingCanBeClicked(t *testing.T) {
	m := prefModel(t, &config.Config{})

	for i := range m.preferences.fields {
		// Bring it into view the way the keyboard would, then click where it
		// is drawn.
		m.preferences.focusField(i)

		g, ok := m.prefGeometry(m.windowWidth, m.windowHeight)
		require.True(t, ok, "no window")

		first, _ := m.preferences.prefWindow()
		line := m.preferences.lineOf(i)
		row := g.y + 1 + g.listRow + line - first
		col := g.x + 2

		got, hit := m.prefFieldAt(m.windowWidth, m.windowHeight, col, row)
		require.True(t, hit, "nothing at the row %s is drawn on", m.preferences.fields[i].spec.path)
		assert.Equal(t, i, got, "clicking %s landed on %s",
			m.preferences.fields[i].spec.path, m.preferences.fields[got].spec.path)

		m, _ = m.handlePreferencesClick(col, row)
	}
}

// TestClickingAYesNoSettingChangesIt, since there is nothing to type into one.
func TestClickingAYesNoSettingChangesIt(t *testing.T) {
	m := prefModel(t, &config.Config{})
	i := prefIndex(t, m, "SSL.Enabled")
	m.preferences.focusField(i)

	g, _ := m.prefGeometry(m.windowWidth, m.windowHeight)
	first, _ := m.preferences.prefWindow()
	row := g.y + 1 + g.listRow + m.preferences.lineOf(i) - first

	m, _ = m.handlePreferencesClick(g.x+2, row)
	assert.True(t, m.preferences.fields[i].yes)

	m, _ = m.handlePreferencesClick(g.x+2, row)
	assert.False(t, m.preferences.fields[i].yes)
}

// TestSpaceChangesAYesNoSettingAndTypesIntoTheRest.
func TestSpaceChangesAYesNoSettingAndTypesIntoTheRest(t *testing.T) {
	m := prefModel(t, &config.Config{})

	m.preferences.focusField(prefIndex(t, m, "Debug"))
	m = press(m, " ")
	assert.True(t, m.preferences.fields[prefIndex(t, m, "Debug")].yes)

	m.preferences.focusField(prefIndex(t, m, "Keyspace"))
	m = press(m, "a")
	m = press(m, " ")
	m = press(m, "b")
	assert.Equal(t, "a b", m.preferences.fields[prefIndex(t, m, "Keyspace")].input.Value())
}

// TestTabOffersTheValuesForASetting.
func TestTabOffersTheValuesForASetting(t *testing.T) {
	m := prefModel(t, &config.Config{})
	m.preferences.focusField(prefIndex(t, m, "OutputFormat"))

	m = press(m, "tab")
	assert.Equal(t, config.OutputFormats(), m.preferences.matches)

	// Picking one puts it in the setting and the list goes away.
	m, _ = m.movePrefMatch(1)
	m, _ = m.usePrefMatch()
	assert.Equal(t, config.OutputFormats()[1], m.preferences.fields[prefIndex(t, m, "OutputFormat")].value())
	assert.Empty(t, m.preferences.matches)
}

// TestTheCursorStopsAtTheEndsAndReachesTheButtons.
func TestTheCursorStopsAtTheEndsAndReachesTheButtons(t *testing.T) {
	m := prefModel(t, &config.Config{})

	m, _ = m.movePrefFocus(-1)
	assert.Equal(t, 0, m.preferences.focus, "moving up from the first setting went somewhere")

	m.preferences.focusField(len(m.preferences.fields) - 1)
	m, _ = m.movePrefFocus(1)
	assert.Equal(t, onRun, m.preferences.onButton)
	m, _ = m.movePrefFocus(1)
	assert.Equal(t, onCancel, m.preferences.onButton)
	m, _ = m.movePrefFocus(1)
	assert.Equal(t, onCancel, m.preferences.onButton, "moving past Cancel went somewhere")

	// And back off them to the last setting.
	m, _ = m.movePrefFocus(-1)
	m, _ = m.movePrefFocus(-1)
	assert.Equal(t, onNoButton, m.preferences.onButton)
	assert.Equal(t, len(m.preferences.fields)-1, m.preferences.focus)
}

// TestTheButtonsAreWhereTheyAreDrawn.
func TestTheButtonsAreWhereTheyAreDrawn(t *testing.T) {
	m := prefModel(t, &config.Config{SourcePath: filepath.Join(t.TempDir(), "cqlai.json")})
	g, ok := m.prefGeometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	row := g.y + 1 + g.buttonRow
	saveEnd, cancelStart, _ := prefButtonSpans()

	button, hit := m.prefButtonAt(m.windowWidth, m.windowHeight, g.x+2, row)
	require.True(t, hit)
	assert.Equal(t, "save", button)

	button, hit = m.prefButtonAt(m.windowWidth, m.windowHeight, g.x+2+cancelStart, row)
	require.True(t, hit)
	assert.Equal(t, "cancel", button)

	_, hit = m.prefButtonAt(m.windowWidth, m.windowHeight, g.x+2+saveEnd, row)
	assert.False(t, hit, "the gap between the buttons is a button")

	m, _ = m.handlePreferencesClick(g.x+2+cancelStart, row)
	assert.False(t, m.preferences.active, "Cancel left the window open")
}

// TestAPressOutsideTheWindowClosesIt, the same as every other window.
func TestAPressOutsideTheWindowClosesIt(t *testing.T) {
	m := prefModel(t, &config.Config{Host: "h"})
	path := m.preferences.path
	m.setPrefField(prefIndex(t, m, "Host"), "typed-but-not-saved")

	updated, _ := m.handleMousePress(tea.Mouse{X: 0, Y: m.windowHeight - 1, Button: tea.MouseLeft})
	assert.False(t, updated.preferences.active)

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)
	assert.NotContains(t, string(data), "typed-but-not-saved", "closing the window wrote the file anyway")
}

// TestTheWindowFitsTheScreen: it is longer than any terminal, so what it draws
// is a window onto the settings with a scrollbar beside it.
func TestTheWindowFitsTheScreen(t *testing.T) {
	m := prefModel(t, &config.Config{})

	layer, ok := m.viewPreferences(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	assert.LessOrEqual(t, layer.Height, m.windowHeight-1)
	assert.LessOrEqual(t, layer.Width, m.windowWidth)

	lines := strings.Split(layer.Content, "\n")
	assert.Equal(t, layer.Height, len(lines))
	for i, line := range lines {
		assert.Equal(t, layer.Width, lipglossWidth(line), "line %d is a different width", i)
	}

	// The last setting is not on screen when the window opens, and is once the
	// cursor gets to it.
	last := len(m.preferences.fields) - 1
	assert.Greater(t, m.preferences.lineOf(last), m.preferences.rows-1)

	m.preferences.focusField(last)
	first, end := m.preferences.prefWindow()
	assert.GreaterOrEqual(t, m.preferences.lineOf(last), first)
	assert.Less(t, m.preferences.lineOf(last), end)
}

// TestTheWindowIsOneSize: it does not grow when Tab offers a list, or shrink
// when the list goes away. A window that moves under the pointer is one you
// have to find again.
func TestTheWindowIsOneSize(t *testing.T) {
	m := prefModel(t, &config.Config{})
	before, ok := m.prefGeometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)

	m.preferences.focusField(prefIndex(t, m, "Consistency"))
	m = press(m, "tab")
	require.NotEmpty(t, m.preferences.matches)

	after, ok := m.prefGeometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	assert.Equal(t, before.width, after.width)
	assert.Equal(t, before.height, after.height)
	assert.Equal(t, before.x, after.x)
	assert.Equal(t, before.y, after.y)
}

// lipglossWidth is the drawn width of a line, ignoring the colours in it.
func lipglossWidth(line string) int {
	return len([]rune(stripAnsi(line)))
}

// TestTheWindowShowsTheFileRatherThanTheRunningConfiguration.
//
// What cqlai is running on is the file plus a cqlshrc, plus the environment,
// plus the command line. Shown that, Save would write all of it back: a
// password that was only ever in $CASSANDRA_PASSWORD would land in the file,
// put there by someone who opened the window to change the port.
func TestTheWindowShowsTheFileRatherThanTheRunningConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"host":"from-the-file"}`), 0o600))

	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 44

	// The running configuration, as LoadConfig leaves it: the file, with the
	// environment applied over the top.
	m.config = &config.Config{Host: "from-the-environment", Password: "from-the-environment", SourcePath: path}
	m.openPreferences()

	assert.Equal(t, "from-the-file", m.preferences.fields[prefIndex(t, m, "Host")].value())
	assert.Empty(t, m.preferences.fields[prefIndex(t, m, "Password")].value())

	// And saving it back writes what was in the file, not what was around it.
	saved, _ := m.savePreferences()
	require.False(t, saved.preferences.active)

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)
	assert.NotContains(t, string(data), "from-the-environment")

	var written map[string]any
	require.NoError(t, json.Unmarshal(data, &written))
	assert.Equal(t, "from-the-file", written["host"])
	assert.NotContains(t, written, "password")
}

// TestOpeningAndSavingChangesNothing: the window is what the file says, so
// pressing Save without touching anything leaves it as it was.
func TestOpeningAndSavingChangesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	before := `{
  "host": "cassandra-1",
  "port": 9142,
  "consistency": "LOCAL_QUORUM",
  "requireConfirmation": true,
  "ssl": {
    "enabled": true,
    "caPath": "/etc/ca.pem"
  },
  "ai": {
    "provider": "anthropic",
    "apiKey": "k",
    "model": "",
    "openai": {
      "apiKey": "j",
      "model": "gpt-4"
    }
  }
}`
	require.NoError(t, os.WriteFile(path, []byte(before), 0o600))

	m := helpModel()
	m.windowWidth = 120
	m.windowHeight = 44
	m.config = &config.Config{SourcePath: path}
	m.openPreferences()
	m, _ = m.savePreferences()
	require.False(t, m.preferences.active)

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)

	var was, now map[string]any
	require.NoError(t, json.Unmarshal([]byte(before), &was))
	require.NoError(t, json.Unmarshal(data, &now))
	assert.Equal(t, was, now)
}

// TestATerminalTooSmallGetsNoPreferences, rather than a window that is not
// drawn and takes the keyboard anyway.
func TestATerminalTooSmallGetsNoPreferences(t *testing.T) {
	m := helpModel()
	m.windowWidth = 40
	m.windowHeight = 12
	m.config = &config.Config{SourcePath: filepath.Join(t.TempDir(), "cqlai.json")}

	m, _ = m.openPreferences()
	assert.False(t, m.preferences.active)
	assert.Contains(t, m.fullHistoryContent, "too small")
}
