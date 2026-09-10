package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheBackgroundKeepsItsColourBesideALayer.
//
// It used to be stripped of ANSI and pasted back as plain text, so anything
// coloured beside a modal lost its colour for as long as the modal was up -
// most visibly the Console scrollbar, which turned white for exactly the rows
// the modal covered.
func TestTheBackgroundKeepsItsColourBesideALayer(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	background := "left" + strings.Repeat(" ", 16) + red.Render("█")
	require.Equal(t, 21, lipgloss.Width(background))

	lm := NewLayerManager(21, 1)
	lm.AddLayer(Layer{Content: "MODAL", X: 6, Y: 0, Width: 5, Height: 1})

	got := lm.Render(background)

	assert.Equal(t, "left  MODAL         █", stripAnsiForTest(got), "the row should be intact")
	assert.Contains(t, got, red.Render("█"), "the scrollbar keeps its colour")
}

// TestTheLayerIsNotColouredByWhatIsBehindIt.
func TestTheLayerIsNotColouredByWhatIsBehindIt(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	background := green.Render("coloured background text here")

	lm := NewLayerManager(29, 1)
	lm.AddLayer(Layer{Content: "PLAIN", X: 4, Y: 0, Width: 5, Height: 1})

	got := lm.Render(background)

	before, _, found := strings.Cut(got, "PLAIN")
	require.True(t, found)
	assert.True(t, strings.HasSuffix(before, ansiReset),
		"the background's colour should be closed before the layer: %q", before)
}

// TestALayerPastTheEndOfTheLineIsPadded rather than jammed against the text.
func TestALayerPastTheEndOfTheLineIsPadded(t *testing.T) {
	lm := NewLayerManager(20, 1)
	lm.AddLayer(Layer{Content: "X", X: 10, Y: 0, Width: 1, Height: 1})

	got := stripAnsiForTest(lm.Render("short"))

	assert.Equal(t, "short     X", got)
}

// TestTheWidthIsUnchangedByALayer, so nothing is pushed off the screen.
func TestTheWidthIsUnchangedByALayer(t *testing.T) {
	background := strings.Repeat("x", 40)

	lm := NewLayerManager(40, 1)
	lm.AddLayer(Layer{Content: "MODAL", X: 10, Y: 0, Width: 5, Height: 1})

	assert.Equal(t, 40, lipgloss.Width(stripAnsiForTest(lm.Render(background))))
}

// TestALayerOnlyTouchesItsOwnRows.
func TestALayerOnlyTouchesItsOwnRows(t *testing.T) {
	background := strings.Join([]string{"row0", "row1", "row2", "row3"}, "\n")

	lm := NewLayerManager(10, 4)
	lm.AddLayer(Layer{Content: "AA", X: 0, Y: 1, Width: 2, Height: 1})

	rows := strings.Split(stripAnsiForTest(lm.Render(background)), "\n")
	assert.Equal(t, "row0", rows[0])
	assert.Equal(t, "AAw1", rows[1])
	assert.Equal(t, "row2", rows[2])
	assert.Equal(t, "row3", rows[3])
}

// TestTheConsoleScrollbarSurvivesAModal, which is the case that started this.
func TestTheConsoleScrollbarSurvivesAModal(t *testing.T) {
	m := consoleModel(200, 6)
	drawn := m.consoleScrollbar(m.historyViewport.View())

	lm := NewLayerManager(41, 6)
	lm.AddLayer(Layer{Content: strings.Join([]string{"┌────┐", "│ hi │", "└────┘"}, "\n"),
		X: 5, Y: 1, Width: 6, Height: 3})

	got := lm.Render(drawn)

	thumb := lipgloss.NewStyle().Foreground(m.styles.Accent).Render("█")
	track := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a")).Render("░")

	for i, row := range strings.Split(got, "\n") {
		plain := []rune(stripAnsiForTest(row))
		require.NotEmpty(t, plain)

		last := plain[len(plain)-1]
		assert.True(t, last == '█' || last == '░', "row %d lost its bar: %q", i, stripAnsiForTest(row))

		// And it is still the coloured one, not a bare character that renders
		// white - which is what the modal used to leave behind.
		assert.True(t, strings.HasSuffix(row, thumb) || strings.HasSuffix(row, track),
			"row %d lost the bar's colour: %q", i, row)
	}
}
