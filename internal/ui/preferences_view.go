package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Placing the PREFERENCES window and drawing it.
//
// One description of the layout, used to draw and to work out what a click
// landed on. Two copies of it is how a click ends up on the row above the one
// pointed at, which is the trap the status line, the settings chooser and the
// COPY forms all avoid the same way.

// prefValueWidth is how wide a setting's box is, which is a path worth reading.
const prefValueWidth = 34

// prefChrome is every row of the window that is neither a setting nor a
// candidate: the title, what it says and a blank; a blank above the hint and
// the hint; a blank below the candidates; the line saying where Save writes;
// the buttons; the keys; and the border top and bottom.
const prefChrome = 3 + 1 + 1 + 1 + 1 + 1 + 1 + 2

// The candidate rows Tab fills in, reserved whether or not there are any, so
// the window is one size for as long as it is open.
const (
	prefMatchRowsMax = 6
	prefMatchRowsMin = 3
)

// prefKeyLines is every line of keys the window can show. The widest sets the
// width, so moving between settings does not change it.
var prefKeyLines = []string{
	"↑↓: Setting   Tab: Complete   Esc: Close",
	"↑↓: Setting   Tab: Values   Esc: Close",
	"↑↓: Setting   Space: Change   Esc: Close",
	"↑↓: Setting   Esc: Close",
	"↑↓: Move   Enter: Press   Esc: Close",
	"↑↓/wheel: Move   Enter: Use   Esc: Back",
}

const prefDescribe = "What cqlai starts with. The status line changes this session."

// fitPrefHeights works out how many settings and how many candidates this
// screen has room for. Both are settled when the window opens.
func (m *MainModel) fitPrefHeights(screenHeight int) (rows, matchRows int) {
	matchRows = prefMatchRowsMax
	room := screenHeight - 1 - prefChrome - matchRows

	// Below the minimum the candidates give up their rows first: a settings
	// area of three lines is a window you cannot find anything in.
	for matchRows > prefMatchRowsMin && room < 12 {
		matchRows--
		room++
	}
	return max(room, 3), matchRows
}

// labelWidth is the width of the label column, from every label rather than
// from the ones on screen: the column cannot move as the list scrolls.
func (p preferences) labelWidth() int {
	width := 0
	for _, field := range p.fields {
		width = max(width, lipgloss.Width(field.spec.label))
	}
	return width
}

// fixedWidth is the window's content width, from the parts of it that do not
// change while it is open.
func (p preferences) fixedWidth() int {
	width := max(lipgloss.Width("Preferences"), lipgloss.Width(prefDescribe))
	for _, keys := range prefKeyLines {
		width = max(width, lipgloss.Width(keys))
	}
	width = max(width, lipgloss.Width(prefSavedTo(p.path)))

	// The marker, the labels, the gap and a box; and a column at the edge for
	// the scrollbar.
	return max(width, 2+p.labelWidth()+2+prefValueWidth) + 1
}

// prefSavedTo is the line above the buttons, saying which file Save writes.
//
// Beside the button rather than at the top: cqlai reads three possible files
// and it is worth being able to see which one this is before pressing it.
func prefSavedTo(path string) string {
	return "Save writes " + path
}

// prefLineText draws one line of the settings area.
func (m *MainModel) prefLineText(line prefLine) string {
	if line.field < 0 {
		return line.heading
	}

	p := m.preferences
	field := p.fields[line.field]

	marker := "  "
	switch {
	case line.field == p.focus && p.onButton == onNoButton:
		marker = "> "
	case m.prefWrong(line.field):
		// Marked even when you are somewhere else, so a window that will not
		// save says where rather than only that.
		marker = "! "
	}

	return marker + pad(field.spec.label, p.labelWidth()) + "  " + field.display()
}

// prefWindow is the part of the settings list that is showing.
func (p preferences) prefWindow() (first, last int) {
	lines := p.lines()
	if len(lines) <= p.rows {
		return 0, len(lines)
	}
	first = min(max(p.scroll, 0), len(lines)-p.rows)
	return first, first + p.rows
}

// prefMatchWindow is the part of the candidate list that is showing.
func (p preferences) prefMatchWindow() (first, last int) {
	if len(p.matches) <= p.matchRows {
		return 0, len(p.matches)
	}
	first = min(max(p.scrollTop, 0), len(p.matches)-p.matchRows)
	return first, first + p.matchRows
}

// prefHint is the line under the settings saying what the one you are on is
// for, or what is wrong with it.
func (m *MainModel) prefHint() string {
	field := m.preferences.current()
	if field == nil {
		return ""
	}
	if wrong, bad := m.preferenceErrors()[m.preferences.focus]; bad {
		return wrong
	}
	if field.spec.hint != "" {
		return field.spec.label + ": " + field.spec.hint
	}
	return ""
}

// prefKeys is the line of keys, which changes with where you are.
func (m *MainModel) prefKeys() string {
	if len(m.preferences.matches) > 0 {
		return "↑↓/wheel: Move   Enter: Use   Esc: Back"
	}
	if m.preferences.onButton != onNoButton {
		return "↑↓: Move   Enter: Press   Esc: Close"
	}

	keys := "↑↓: Setting   Esc: Close"
	if field := m.preferences.current(); field != nil {
		switch field.spec.kind {
		case prefPath:
			keys = "↑↓: Setting   Tab: Complete   Esc: Close"
		case prefChoice:
			keys = "↑↓: Setting   Tab: Values   Esc: Close"
		case prefYesNo:
			keys = "↑↓: Setting   Space: Change   Esc: Close"
		}
	}
	return keys
}

// prefGeometry is where the window sits and what is in it.
type prefGeometry struct {
	x, y          int
	width, height int
	rows          []string

	listRow   int // the row the settings start on, inside the border
	matchRow  int // the row the candidates start on
	buttonRow int

	// bars is the scrollbar character for the rows that have one, against the
	// right-hand edge of the window rather than beside the list.
	bars map[int]rune
}

// prefRows is the window's content, one line each.
func (m *MainModel) prefRows() (rows []string, listRow, matchRow int, bars map[int]rune) {
	p := m.preferences
	bars = map[int]rune{}

	rows = []string{"Preferences", prefDescribe, ""}
	listRow = len(rows)

	lines := p.lines()
	first, last := p.prefWindow()
	thumb := scrollbarColumn(last-first, first, len(lines))
	scrolls := len(lines) > p.rows

	for i := first; i < last; i++ {
		rows = append(rows, m.prefLineText(lines[i]))
		if scrolls {
			bar := '░'
			if thumb[i-first] {
				bar = '█'
			}
			bars[len(rows)-1] = bar
		}
	}
	// A short list still takes the whole area, so the window is one size.
	for range p.rows - (last - first) {
		rows = append(rows, "")
	}

	rows = append(rows, "", m.prefHint())
	matchRow = len(rows)

	if len(p.matches) == 0 {
		for range p.matchRows {
			rows = append(rows, "")
		}
		return append(rows, "", prefSavedTo(p.path), m.prefButtons(), m.prefKeys()), listRow, -1, bars
	}

	firstMatch, lastMatch := p.prefMatchWindow()
	width := 0
	for _, match := range p.matches {
		width = max(width, lipgloss.Width(match))
	}
	matchThumb := scrollbarColumn(lastMatch-firstMatch, firstMatch, len(p.matches))
	matchScrolls := len(p.matches) > p.matchRows

	for i := firstMatch; i < lastMatch; i++ {
		marker := "  "
		if i == p.match {
			marker = "> "
		}
		rows = append(rows, marker+pad(p.matches[i], width))

		bar := ' '
		if matchScrolls {
			bar = '░'
			if matchThumb[i-firstMatch] {
				bar = '█'
			}
		}
		bars[len(rows)-1] = bar
	}
	for range p.matchRows - (lastMatch - firstMatch) {
		rows = append(rows, "")
	}

	return append(rows, "", prefSavedTo(p.path), m.prefButtons(), m.prefKeys()), listRow, matchRow, bars
}

// The buttons at the foot. Saving from the keyboard alone would make this the
// one window in cqlai that cannot be finished with the mouse.
const (
	prefSave   = "Save"
	prefCancel = "Cancel"
)

// prefButtons is the button row, at its full width so it sets the window's.
func (m *MainModel) prefButtons() string {
	return "  [ " + prefSave + " ]" + formButtonGap + "  [ " + prefCancel + " ]"
}

// prefButtonSpans is where the two buttons sit inside the row, so drawing and
// clicking cannot disagree about which one was pressed.
func prefButtonSpans() (saveEnd, cancelStart, cancelEnd int) {
	saveEnd = 2 + lipgloss.Width("[ "+prefSave+" ]")
	cancelStart = saveEnd + len(formButtonGap)
	return saveEnd, cancelStart, cancelStart + 2 + lipgloss.Width("[ "+prefCancel+" ]")
}

// renderPrefButtons draws the button row, with Save dimmed while something in
// the window cannot be written.
func (m *MainModel) renderPrefButtons(inner int) string {
	ready := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))
	plain := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())

	mark := func(label string, on bool) string {
		if on {
			return "> " + label
		}
		return "  " + label
	}

	save := mark("[ "+prefSave+" ]", m.preferences.onButton == onRun)
	if m.preferencesReady() {
		save = ready.Render(save)
	} else {
		save = dim.Render(save)
	}

	cancel := plain.Render(mark("[ "+prefCancel+" ]", m.preferences.onButton == onCancel))
	row := " " + save + formButtonGap + cancel
	if gap := inner - lipgloss.Width(stripAnsi(row)); gap > 0 {
		row += strings.Repeat(" ", gap)
	}
	return row
}

// prefGeometryFor places the window in the middle of the screen.
func (m *MainModel) prefGeometry(screenWidth, screenHeight int) (prefGeometry, bool) {
	if !m.preferences.active {
		return prefGeometry{}, false
	}

	rows, listRow, matchRow, bars := m.prefRows()

	inner := min(m.preferences.fixedWidth()+2, max(screenWidth-2, 0))
	width := inner + 2
	height := len(rows) + 2
	if width < 20 || height > screenHeight-1 {
		return prefGeometry{}, false
	}

	return prefGeometry{
		x:         max((screenWidth-width)/2, 0),
		y:         max((screenHeight-height)/2, 0),
		width:     width,
		height:    height,
		rows:      rows,
		listRow:   listRow,
		matchRow:  matchRow,
		buttonRow: len(rows) - 2, // the keys line is last
		bars:      bars,
	}, true
}

// viewPreferences draws the window.
func (m *MainModel) viewPreferences(screenWidth, screenHeight int) (Layer, bool) {
	g, ok := m.prefGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	titleStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	headingStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	wrongStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d75f5f"))
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	lines := m.preferences.lines()
	first, _ := m.preferences.prefWindow()

	inner := g.width - 2
	drawn := make([]string, 0, len(g.rows))
	for i, row := range g.rows {
		style := textStyle
		switch {
		case i == 0:
			style = titleStyle
		case i >= g.listRow && i < g.listRow+m.preferences.rows:
			n := first + i - g.listRow
			switch {
			case n < len(lines) && lines[n].heading != "":
				style = headingStyle
			case n < len(lines) && lines[n].field >= 0 && m.prefWrong(lines[n].field):
				style = wrongStyle
			}
		case i == g.listRow+m.preferences.rows+1 && m.prefHintWrong():
			style = wrongStyle
		}

		if i == g.buttonRow {
			drawn = append(drawn, m.renderPrefButtons(inner))
			continue
		}

		if bar, ok := g.bars[i]; ok {
			// One column short, so the bar has the edge to itself.
			text := style.Render(pad(" "+row, inner-1))
			switch bar {
			case '█':
				text += thumbStyle.Render("█")
			case '░':
				text += trackStyle.Render("░")
			default:
				text += " "
			}
			drawn = append(drawn, text)
			continue
		}

		drawn = append(drawn, style.Render(pad(" "+row, inner)))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Accent).
		Render(strings.Join(drawn, "\n"))

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  260,
	}, true
}

// prefHintWrong reports whether the hint line is saying what is wrong rather
// than what the setting is for.
func (m *MainModel) prefHintWrong() bool {
	if m.preferences.onButton != onNoButton {
		return false
	}
	_, wrong := m.preferenceErrors()[m.preferences.focus]
	return wrong
}

// prefFieldAt returns the setting a press covers.
func (m *MainModel) prefFieldAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.prefGeometry(screenWidth, screenHeight)
	if !ok {
		return 0, false
	}
	if col < g.x || col >= g.x+g.width {
		return 0, false
	}

	n := row - g.y - 1 - g.listRow
	if n < 0 || n >= m.preferences.rows {
		return 0, false
	}

	first, _ := m.preferences.prefWindow()
	lines := m.preferences.lines()
	if first+n >= len(lines) {
		return 0, false
	}
	line := lines[first+n]
	if line.field < 0 {
		return 0, false // a heading, or the blank above one
	}
	return line.field, true
}

// prefMatchAt returns the candidate a press covers.
func (m *MainModel) prefMatchAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.prefGeometry(screenWidth, screenHeight)
	if !ok || g.matchRow < 0 {
		return 0, false
	}
	if col < g.x || col >= g.x+g.width {
		return 0, false
	}

	n := row - g.y - 1 - g.matchRow
	first, last := m.preferences.prefMatchWindow()
	if n < 0 || n >= last-first {
		return 0, false
	}
	return first + n, true
}

// prefButtonAt returns which button a press landed on.
func (m *MainModel) prefButtonAt(screenWidth, screenHeight, col, row int) (string, bool) {
	g, ok := m.prefGeometry(screenWidth, screenHeight)
	if !ok || row != g.y+1+g.buttonRow {
		return "", false
	}

	// One for the border, one for the row's own padding.
	at := col - g.x - 2
	saveEnd, cancelStart, cancelEnd := prefButtonSpans()

	switch {
	case at >= 0 && at < saveEnd:
		return "save", true
	case at >= cancelStart && at < cancelEnd:
		return "cancel", true
	}
	return "", false
}

// inPreferences reports whether a press landed inside the window.
func (m *MainModel) inPreferences(screenWidth, screenHeight, col, row int) bool {
	g, ok := m.prefGeometry(screenWidth, screenHeight)
	if !ok {
		return false
	}
	return col >= g.x && col < g.x+g.width && row >= g.y && row < g.y+g.height
}
