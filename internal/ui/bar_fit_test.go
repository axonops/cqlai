package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// busyStatusBar is a session with something in every field: a keyspace with a
// real name and a consistency level longer than the default.
func busyStatusBar() StatusBarModel {
	m := NewStatusBarModel()
	m.Keyspace = "my_keyspace"
	m.Consistency = "LOCAL_QUORUM"
	m.OutputFormat = "TABLE"
	m.PagingSize = 100
	return m
}

// barWidths is the number of rows a rendered bar takes and how wide its widest
// row is.
func barWidths(rendered string) (rows, width int) {
	for _, line := range strings.Split(rendered, "\n") {
		rows++
		width = max(width, lipgloss.Width(line))
	}
	return rows, width
}

// testWidths covers a 24x80 default terminal, a split pane, and something wide.
var testWidths = []int{20, 30, 40, 60, 80, 100, 120, 200}

// TestTheStatusLineIsAlwaysOneRow is the defect behind #137.
//
// The bar laid itself out at whatever width it wanted and was rendered through
// lipgloss Width, which wraps rather than truncates. Its fields come to 115
// columns with a real keyspace name, so below about a hundred columns it became
// two rows and the whole view was a row taller than the terminal - and 80
// columns is the default width of a great many terminals, so this was the
// common case rather than the edge one.
func TestTheStatusLineIsAlwaysOneRow(t *testing.T) {
	m := busyStatusBar()

	for _, width := range testWidths {
		rows, got := barWidths(m.View(width, DefaultStyles(), "history"))

		assert.Equal(t, 1, rows, "at %d columns the status line took %d rows", width, rows)
		assert.Equal(t, width, got, "and it fills the width exactly")
	}
}

// TestTheInfoBarIsAlwaysOneRow. The same defect one row up, found by the same
// test: at 30 columns it wrapped where the status line did not.
func TestTheInfoBarIsAlwaysOneRow(t *testing.T) {
	m := TopBarModel{
		LastCommand:  "SELECT id, name, status FROM users WHERE status = 'active' ALLOW FILTERING",
		HasQueryData: true,
		RowCount:     1234,
	}

	for _, width := range testWidths {
		rows, got := barWidths(m.View(width, DefaultStyles(), "history"))

		assert.Equal(t, 1, rows, "at %d columns the info bar took %d rows", width, rows)
		assert.Equal(t, width, got, "and it fills the width exactly")
	}
}

// TestTheWholeViewFitsTheTerminal, which is what the wrapping cost: 25 rows for
// a 24-row terminal at 80 columns, 26 at 40.
func TestTheWholeViewFitsTheTerminal(t *testing.T) {
	for _, width := range testWidths {
		m := widthModel(t, width)
		m.statusBar = busyStatusBar()

		rows := strings.Count(m.View().Content, "\n") + 1

		assert.LessOrEqual(t, rows, 24, "at %d columns the view is %d rows", width, rows)
	}
}

// TestTheKeyspaceAndConsistencyAreTheLastToGo. They decide what a query does
// and where it goes; the rest are settings you can look up.
func TestTheKeyspaceAndConsistencyAreTheLastToGo(t *testing.T) {
	m := busyStatusBar()

	kept := func(width int) []string {
		var settings []string
		for _, seg := range placeSegments(m.segments(), width) {
			settings = append(settings, seg.setting)
		}
		return settings
	}

	assert.Contains(t, kept(60), settingKeyspace)
	assert.Contains(t, kept(60), settingConsistency)
	assert.NotContains(t, kept(60), settingAutoFetch, "the least useful goes first")

	full := kept(200)
	for _, setting := range []string{
		settingConnection, settingKeyspace, settingConsistency,
		settingOutput, settingPaging, settingTracing, settingAutoFetch, settingCapture,
	} {
		assert.Contains(t, full, setting, "a wide terminal shows everything")
	}
}

// TestTheLabelsShortenBeforeAFieldIsDropped. Losing three characters of a label
// costs nothing; losing a field costs what it said.
func TestTheLabelsShortenBeforeAFieldIsDropped(t *testing.T) {
	m := busyStatusBar()

	atWidth := func(width int) string {
		return stripAnsiForTest(m.View(width, DefaultStyles(), "history"))
	}

	assert.Contains(t, atWidth(120), "Connection", "there is room for the full labels")
	assert.Contains(t, atWidth(120), "Fetch: ")

	narrow := atWidth(100)
	assert.Contains(t, narrow, "Conn", "shortened rather than dropped")
	assert.NotContains(t, narrow, "Connection")
	assert.Contains(t, narrow, "F: ", "and the field is still there")
}

// TestAClickLandsOnWhatIsDrawn at any width. Rendering and hit testing both go
// through placeSegments, so what has been shortened or dropped to fit cannot
// make them disagree.
func TestAClickLandsOnWhatIsDrawn(t *testing.T) {
	m := busyStatusBar()

	for _, width := range testWidths {
		for _, seg := range placeSegments(m.segments(), width) {
			if !seg.clickable() {
				continue
			}
			mid := (seg.start + seg.end) / 2

			setting, _, ok := m.settingAt(width, mid)

			require.True(t, ok, "%s at %d columns, column %d", seg.setting, width, mid)
			assert.Equal(t, seg.setting, setting, "at %d columns", width)
		}
	}
}

// TestTheLastCommandIsCutRatherThanDropped. It is the only field that can be
// any length, and a shortened command still tells you which one it was.
func TestTheLastCommandIsCutRatherThanDropped(t *testing.T) {
	m := TopBarModel{
		LastCommand:  "SELECT id, name, status FROM users WHERE status = 'active'",
		HasQueryData: true,
	}

	drawn := stripAnsiForTest(m.View(60, DefaultStyles(), "history"))

	assert.Contains(t, drawn, "History: SELECT", "the start of it survives")
	assert.Contains(t, drawn, "…", "and it says it was cut")
	assert.Contains(t, drawn, "Rows: ", "the short fields stay")
}

// TestTruncateToWidthNeverOverruns, whatever it is given.
func TestTruncateToWidthNeverOverruns(t *testing.T) {
	for _, room := range []int{0, 1, 2, 5, 20, 80} {
		for _, value := range []string{"", "a", "SELECT * FROM users", strings.Repeat("x", 200)} {
			got := truncateToWidth(value, room)
			assert.LessOrEqual(t, lipgloss.Width(got), max(room, 1),
				"%q in %d columns gave %q", value, room, got)
		}
	}
}
