package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/router"
)

// The window behind Capture on the bottom line: pick a format, then type where
// to write it.
//
// Two steps, which is one more than the settings lists do, so it is its own
// window rather than another commandList. It anchors above the field that
// opened it, the way those lists do.
//
// SAVE has the same shape and its own copy of it. #129 tracks moving SAVE onto
// this one; two of them will drift, which is how five separate bugs happened
// this week.

// captureStep is which half of the window is showing.
type captureStep int

const (
	captureChooseFormat captureStep = iota
	captureEnterPath
)

// capturePanel is the open window, if any.
type capturePanel struct {
	active  bool
	step    captureStep
	formats []string
	format  int // index into formats
	anchorX int // column the field starts at, when opened from it

	// centred says to put the window in the middle of the screen rather than
	// above the field. Typing CAPTURE has no field to point at, and a window
	// hugging the bottom right corner for a command typed at the prompt looks
	// like it belongs to something you did not touch.
	centred bool

	input textinput.Model

	// matches are the candidates from the last Tab, shown when more than one
	// name fits what has been typed. They are a list you pick from, not a note
	// under the box: joined onto one line they ran off the right-hand edge with
	// no way to see or reach the rest.
	matches     []string
	match       int // which candidate is highlighted
	matchScroll int // first candidate shown
}

// captureMatchRows is the most candidates shown at once. Anything longer
// scrolls, so a directory with a hundred files does not fill the screen.
const captureMatchRows = 8

// captureStopping is the entry that turns capture off, offered only while it
// is running.
const captureStopping = "OFF"

// openCapturePanel opens the window above the Capture field, for a click on it.
func (m *MainModel) openCapturePanel(anchorX int) (*MainModel, tea.Cmd) {
	return m.showCapturePanel(anchorX, false)
}

// openCapturePanelCentred opens it in the middle of the screen, for the typed
// CAPTURE command, which has no field to point at.
func (m *MainModel) openCapturePanelCentred() (*MainModel, tea.Cmd) {
	return m.showCapturePanel(0, true)
}

func (m *MainModel) showCapturePanel(anchorX int, centred bool) (*MainModel, tea.Cmd) {
	if m.capture.active {
		m.capture = capturePanel{}
		return m, nil
	}

	// The formats the command parses, so the list cannot offer one it will
	// refuse. CAPTURE with no format at all writes text; that is still there
	// if you type it, but it is not one of the choices, because the usage does
	// not list it as one.
	formats := router.CaptureFormats()
	if handler := router.GetMetaHandler(); handler != nil && handler.IsCapturing() {
		// Stopping is the thing you want while it is running, so it goes first.
		formats = append([]string{captureStopping}, formats...)
	}

	input := textinput.New()
	input.Prompt = "> "
	input.CharLimit = 512
	input.SetWidth(48)

	m.capture = capturePanel{
		active:  true,
		formats: formats,
		anchorX: anchorX,
		centred: centred,
		input:   input,
	}
	return m, nil
}

// closeCapturePanel dismisses the window without changing anything.
func (m *MainModel) closeCapturePanel() {
	m.capture = capturePanel{}
}

// captureGeometry is where the window sits.
//
// Drawing and clicking both come through here, so a click cannot land on a
// different row from the one drawn there.
type captureGeometry struct {
	x, y          int
	width, height int
	rows          []string // the lines inside the border
}

// captureRows builds what the window shows at the step it is on.
func (m *MainModel) captureRows() []string {
	c := m.capture

	if c.step == captureChooseFormat {
		rows := make([]string, 0, len(c.formats)+2)
		rows = append(rows, "Capture output to a file", "")
		for i, format := range c.formats {
			marker := "  "
			if i == c.format {
				marker = "> "
			}
			rows = append(rows, marker+captureFormatLabel(format))
		}
		return append(rows, "", "↑↓: Move   Enter: Choose   Esc: Close")
	}

	rows := []string{
		fmt.Sprintf("Capture as %s", c.formats[c.format]),
		"",
		c.input.View(),
		"",
	}

	if len(c.matches) == 0 {
		return append(rows, "Tab: Complete   Enter: Start   Esc: Back")
	}

	// The candidates from the last Tab, as a list to pick from.
	first, last := c.matchWindow()
	width := 0
	for _, name := range c.matches {
		width = max(width, lipgloss.Width(name))
	}

	for i := first; i < last; i++ {
		marker := "  "
		if i == c.match {
			marker = "> "
		}
		rows = append(rows, marker+pad(c.matches[i], width))
	}

	return append(rows, "", "↑↓: Move   Enter: Use   Tab: Complete   Esc: Back")
}

// matchWindow is the slice of candidates showing, so a long list scrolls with
// the highlighted one in view.
func (c capturePanel) matchWindow() (first, last int) {
	if len(c.matches) <= captureMatchRows {
		return 0, len(c.matches)
	}

	first = min(max(c.matchScroll, 0), len(c.matches)-captureMatchRows)
	return first, first + captureMatchRows
}

// captureMatchesStart is how many rows sit above the candidates inside the
// border: the title, a blank line, the path box, and another blank.
const captureMatchesStart = 4

// matchAt returns the candidate under a screen position.
func (m *MainModel) matchAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.captureGeometry(screenWidth, screenHeight)
	if !ok || m.capture.step != captureEnterPath || len(m.capture.matches) == 0 {
		return 0, false
	}
	if col <= g.x || col >= g.x+g.width-1 {
		return 0, false
	}

	first, last := m.capture.matchWindow()
	i := first + (row - g.y - 1 - captureMatchesStart)
	if i < first || i >= last {
		return 0, false
	}
	return i, true
}

// showMatch highlights a candidate and keeps it in view.
func (m *MainModel) showMatch(i int) {
	m.capture.match = min(max(i, 0), len(m.capture.matches)-1)

	m.capture.matchScroll = min(m.capture.matchScroll, m.capture.match)
	m.capture.matchScroll = max(m.capture.matchScroll, m.capture.match-captureMatchRows+1)
	m.capture.matchScroll = max(m.capture.matchScroll, 0)
}

// useMatch puts the highlighted candidate into the path.
//
// A directory keeps the list open on what is inside it, so you can walk down a
// tree without retyping any of it.
func (m *MainModel) useMatch() (*MainModel, tea.Cmd) {
	if len(m.capture.matches) == 0 {
		return m, nil
	}

	dir, _ := splitPath(m.capture.input.Value())
	chosen := dir + m.capture.matches[m.capture.match]

	m.capture.input.SetValue(chosen)
	m.capture.input.CursorEnd()
	m.clearMatches()

	if strings.HasSuffix(chosen, string(filepath.Separator)) {
		return m.completeCapturePath()
	}
	return m, nil
}

// clearMatches takes the list down.
func (m *MainModel) clearMatches() {
	m.capture.matches = nil
	m.capture.match = 0
	m.capture.matchScroll = 0
}

// captureHeaderRows is how many rows sit above the format entries inside the
// border: the title and the blank line under it.
const captureHeaderRows = 2

// formatAt returns the format under a screen position, if the click is on one.
//
// It counts from the same rows captureRows builds, so a click cannot land on
// the row above the one you pointed at. There is a test that the row it names
// really does hold that format.
func (m *MainModel) formatAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.captureGeometry(screenWidth, screenHeight)
	if !ok || m.capture.step != captureChooseFormat {
		return 0, false
	}

	// Inside the border, not on it.
	if col <= g.x || col >= g.x+g.width-1 {
		return 0, false
	}

	i := row - g.y - 1 - captureHeaderRows
	if i < 0 || i >= len(m.capture.formats) {
		return 0, false
	}
	return i, true
}

// inCapturePanel reports whether a position is inside the window.
func (m *MainModel) inCapturePanel(screenWidth, screenHeight, col, row int) bool {
	g, ok := m.captureGeometry(screenWidth, screenHeight)
	return ok && col >= g.x && col < g.x+g.width && row >= g.y && row < g.y+g.height
}

// captureFormatLabel says what each entry does, since OFF and TEXT are not
// self-explanatory next to CSV and JSON.
func captureFormatLabel(format string) string {
	switch format {
	case captureStopping:
		return "OFF       stop capturing"
	case "CSV":
		return "CSV       comma separated"
	case "JSON":
		return "JSON      one object per row"
	case "PARQUET":
		return "PARQUET   columnar, for analysis"
	}
	return format
}

func (m *MainModel) captureGeometry(screenWidth, screenHeight int) (captureGeometry, bool) {
	if !m.capture.active {
		return captureGeometry{}, false
	}

	rows := m.captureRows()

	inner := 0
	for _, row := range rows {
		if w := lipgloss.Width(row); w > inner {
			inner = w
		}
	}
	inner = min(inner+2, max(screenWidth-2, 0))

	width := inner + 2
	height := len(rows) + 2
	if width < 20 || height > screenHeight-1 {
		return captureGeometry{}, false
	}

	x := min(m.capture.anchorX, max(screenWidth-width, 0))
	y := max(screenHeight-1-height, 0)
	if m.capture.centred {
		x = max((screenWidth-width)/2, 0)
		y = max((screenHeight-height)/2, 0)
	}

	return captureGeometry{x: x, y: y, width: width, height: height, rows: rows}, true
}

// viewCapturePanel renders the window and says where to put it.
func (m *MainModel) viewCapturePanel(screenWidth, screenHeight int) (Layer, bool) {
	g, ok := m.captureGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	titleStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	chosenStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D0D0D0"))
	quietStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())

	inner := g.width - 2

	// The scrollbar goes against the right-hand edge, which is only known once
	// the box has been sized from its widest row. Appending it to the row text
	// left it floating in the middle of the box.
	first, last := m.capture.matchWindow()
	scrolls := m.capture.step == captureEnterPath && len(m.capture.matches) > captureMatchRows
	thumb := scrollbarColumn(last-first, first, len(m.capture.matches))
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	var b strings.Builder
	for i, row := range g.rows {
		style := textStyle
		switch {
		case i == 0:
			style = titleStyle
		case i == len(g.rows)-1:
			style = quietStyle
		case strings.HasPrefix(row, "> "):
			style = chosenStyle
		}

		candidate := i - captureMatchesStart
		if scrolls && candidate >= 0 && candidate < last-first {
			b.WriteString(style.Render(pad(" "+row, inner-1)))
			if thumb[candidate] {
				b.WriteString(thumbStyle.Render("█"))
			} else {
				b.WriteString(trackStyle.Render("░"))
			}
		} else {
			b.WriteString(style.Render(pad(" "+row, inner)))
		}

		if i < len(g.rows)-1 {
			b.WriteString("\n")
		}
	}

	return Layer{
		Content: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.styles.Accent).
			Render(b.String()),
		X:      g.x,
		Y:      g.y,
		Width:  g.width,
		Height: g.height,
		ZIndex: 200,
	}, true
}

// handleCaptureKey takes a keypress while the window is open.
func (m *MainModel) handleCaptureKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	switch m.capture.step {
	case captureChooseFormat:
		switch msg.String() {
		case "esc":
			m.closeCapturePanel()
		case "up":
			m.capture.format = max(m.capture.format-1, 0)
		case "down":
			m.capture.format = min(m.capture.format+1, len(m.capture.formats)-1)
		case "enter":
			return m.chooseCaptureFormat()
		}
		return m, nil

	default:
		// While the candidates are showing they are what the arrows and Enter
		// are aimed at; Escape puts them away and leaves the path alone.
		if len(m.capture.matches) > 0 {
			switch msg.String() {
			case "up":
				m.showMatch(m.capture.match - 1)
				return m, nil
			case "down":
				m.showMatch(m.capture.match + 1)
				return m, nil
			case "enter":
				return m.useMatch()
			case "esc":
				m.clearMatches()
				return m, nil
			}
		}

		switch msg.String() {
		case "esc":
			// Back a step rather than closing, so a mistyped path does not
			// cost you the format as well.
			m.capture.step = captureChooseFormat
			m.clearMatches()
			return m, nil
		case "tab":
			return m.completeCapturePath()
		case "enter":
			return m.startCapture()
		}

		var cmd tea.Cmd
		m.capture.input, cmd = m.capture.input.Update(msg)
		m.clearMatches() // what was listed no longer describes what is typed
		return m, cmd
	}
}

// chooseCaptureFormat moves on to the path, or stops capturing if that is what
// was picked.
func (m *MainModel) chooseCaptureFormat() (*MainModel, tea.Cmd) {
	if m.capture.formats[m.capture.format] == captureStopping {
		m.closeCapturePanel()
		return m.runCommand("CAPTURE OFF")
	}

	m.capture.step = captureEnterPath
	m.capture.input.SetValue(m.defaultCaptureName())
	m.capture.input.CursorEnd()
	m.capture.input.Focus()
	return m, nil
}

// defaultCaptureName is a filename to start from, so there is something to edit
// rather than an empty box.
func (m *MainModel) defaultCaptureName() string {
	ext := ".txt"
	switch m.capture.formats[m.capture.format] {
	case "CSV":
		ext = ".csv"
	case "JSON":
		ext = ".json"
	case "PARQUET":
		ext = ".parquet"
	}

	return fmt.Sprintf("capture_%s%s", time.Now().Format("20060102_150405"), ext)
}

// completeCapturePath fills in as much of the path as the candidates agree on.
func (m *MainModel) completeCapturePath() (*MainModel, tea.Cmd) {
	got := completePath(m.capture.input.Value())

	m.capture.input.SetValue(got.Completed)
	m.capture.input.CursorEnd()

	// List them only when filling in was not enough to pick one.
	m.clearMatches()
	if len(got.Matches) > 1 {
		m.capture.matches = got.Matches
	}
	return m, nil
}

// startCapture runs the CAPTURE command the window describes.
func (m *MainModel) startCapture() (*MainModel, tea.Cmd) {
	path := strings.TrimSpace(m.capture.input.Value())
	if path == "" {
		return m, nil
	}

	command := captureCommand(m.capture.formats[m.capture.format], path)
	m.closeCapturePanel()
	return m.runCommand(command)
}

// captureCommand is the CAPTURE the window describes.
//
// It goes through the same command typing it would, so both routes behave the
// same.
func captureCommand(format, path string) string {
	return "CAPTURE " + format + " '" + filepath.Clean(path) + "'"
}
