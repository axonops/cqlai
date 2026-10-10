package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// topicHelp is the help window open on HELP INSERT, which ends with a link.
func topicHelp(t *testing.T) (*MainModel, helpGeometry, []string) {
	t.Helper()
	m := helpModel()
	m.help = helpWindow{active: true, topic: "insert"}
	m.scrollHelp(1000) // the link is the last line
	g, lines, ok := m.helpGeometry(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	return m, g, lines
}

// linkRow is the screen row and column of the first link showing.
func linkRow(t *testing.T, g helpGeometry, lines []string) (url string, col, row int) {
	t.Helper()
	for i := g.first; i < g.last; i++ {
		if at := strings.Index(lines[i], "https://"); at >= 0 {
			return strings.TrimSpace(lines[i][at:]), g.x + 2 + ansi.StringWidth(lines[i][:at]), g.y + 2 + i - g.first
		}
	}
	t.Fatal("no link showing")
	return "", 0, 0
}

// stubBrowser records what would have been opened, and says whether it worked.
func stubBrowser(t *testing.T, works bool) *[]string {
	t.Helper()
	original := openInBrowser
	t.Cleanup(func() { openInBrowser = original })
	var opened []string
	openInBrowser = func(url string) bool {
		opened = append(opened, url)
		return works
	}
	return &opened
}

// helpDrawn is the window as drawn.
func helpDrawn(t *testing.T, m *MainModel) string {
	t.Helper()
	layer, ok := m.viewHelp(m.windowWidth, m.windowHeight)
	require.True(t, ok)
	return layer.Content
}

func ctrlPressAt(m *MainModel, col, row int) tea.Cmd {
	_, cmd := m.handleMouseInput(tea.MouseClickMsg{Button: tea.MouseLeft, Mod: tea.ModCtrl, X: col, Y: row})
	return cmd
}

// TestCtrlClickOpensTheLink: cqlai has the mouse, so the terminal never sees
// the Ctrl+click and cqlai has to open the link itself.
func TestCtrlClickOpensTheLink(t *testing.T) {
	opened := stubBrowser(t, true)
	m, g, lines := topicHelp(t)
	url, col, row := linkRow(t, g, lines)
	require.True(t, strings.HasPrefix(url, "https://axonops.com/docs/"))

	cmd := ctrlPressAt(m, col+3, row)
	require.NotNil(t, cmd)
	assert.Nil(t, cmd())

	assert.Equal(t, []string{url}, *opened)
	assert.True(t, m.help.active, "the window stays open")
	assert.Contains(t, ansi.Strip(helpDrawn(t, m)), "Opening "+url[:20])
}

// TestAPlainClickOnALinkOpensNothing: a click is how a selection starts, and
// a double click on the link should pick a word of it, not start a browser.
func TestAPlainClickOnALinkOpensNothing(t *testing.T) {
	opened := stubBrowser(t, true)
	m, g, lines := topicHelp(t)
	_, col, row := linkRow(t, g, lines)

	pressAt(m, col+3, row)

	assert.Empty(t, *opened)
	assert.True(t, m.help.active)
}

// TestWithNoBrowserTheLinkIsCopied, over ssh say, so it can be pasted into a
// browser somewhere else.
func TestWithNoBrowserTheLinkIsCopied(t *testing.T) {
	stubBrowser(t, false)
	m, g, lines := topicHelp(t)
	url, col, row := linkRow(t, g, lines)

	msg := ctrlPressAt(m, col, row)()
	require.Equal(t, linkNotOpenedMsg{url: url}, msg)
	_, cmd := m.Update(msg)

	assert.Equal(t, url, clipboardText(t, cmd))
	assert.Equal(t, url, m.lastCopied)
	assert.Contains(t, ansi.Strip(helpDrawn(t, m)), "The link is copied")
}

// TestDraggingOverTheHelpCopiesIt, which is what you do to keep a line of it.
func TestDraggingOverTheHelpCopiesIt(t *testing.T) {
	m, g, lines := topicHelp(t)
	url, col, row := linkRow(t, g, lines)

	pressAt(m, col, row)
	m.handleMouseInput(tea.MouseMotionMsg{Button: tea.MouseLeft, X: col + len(url), Y: row})
	_, cmd := m.handleMouseInput(tea.MouseReleaseMsg{Button: tea.MouseLeft, X: col + len(url), Y: row})

	assert.Equal(t, url, clipboardText(t, cmd))
	assert.True(t, m.help.active)

	// The highlight is drawn on the window, over the link.
	view := helpDrawn(t, m)
	assert.Contains(t, view, selectionStyle.Render(url))
}

// TestTheHelpSelectionGoesWithTheWindow: a span over the help is not painted
// onto the view behind it once the window closes.
func TestTheHelpSelectionGoesWithTheWindow(t *testing.T) {
	m, g, lines := topicHelp(t)
	url, col, row := linkRow(t, g, lines)
	drag(m, col, row, col+len(url), row)
	require.True(t, m.selection.active)

	m.toggleHelp()

	assert.False(t, m.selection.active)
}

func TestLinkAt(t *testing.T) {
	line := "More: https://example.com/a/b#x. And so on"
	url, ok := linkAt(line, 6)
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/a/b#x", url, "the full stop is the sentence's")
	_, ok = linkAt(line, 5)
	assert.False(t, ok, "the space before it")
	_, ok = linkAt(line, 31)
	assert.False(t, ok, "the full stop after it")
	_, ok = linkAt("no link here", 3)
	assert.False(t, ok)
}
