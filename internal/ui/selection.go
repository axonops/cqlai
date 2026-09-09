package ui

import (
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Text selection drawn by cqlai rather than by the terminal.
//
// Asking the terminal for mouse events takes its own click-and-drag selection
// away, and there is no mode that gives you both. The way out is to stop
// relying on the terminal for selection at all: take the mouse, follow the drag
// ourselves, paint the highlight into what we render, and hand the text to the
// system clipboard on release. Terminal applications that manage clicking and
// selecting at the same time all do it this way.
//
// The span is stored against the active viewport's content, by line and column,
// not by screen row. Scrolling then moves the text under the highlight without
// moving the highlight off it, which matters because dragging past an edge
// scrolls on purpose.

const (
	// tabBarHeight is how many rows the mode tabs take, and so the first row
	// the viewport occupies. WindowSizeMsg reserves the same figure.
	tabBarHeight = 1

	// stickyHeaderHeight is how many rows the frozen table header covers when a
	// table is scrolled: top border, column names, separator.
	stickyHeaderHeight = 3

	// multiClickInterval is how close together presses have to be to count as a
	// double or triple click.
	multiClickInterval = 400 * time.Millisecond
)

// selectionStyle is what a selected run of text looks like. It matches the
// active tab, so the two highlights on screen are the same colour.
var selectionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#1c1c1c")).
	Background(lipgloss.Color("#87D7FF"))

// textSelection is the span being dragged out, or the one just dragged.
type textSelection struct {
	active   bool
	dragging bool

	// view names the viewport the coordinates belong to, so switching mode
	// drops the selection instead of painting it onto different content.
	view string

	anchorLine, anchorCol int
	headLine, headCol     int

	// Press tracking, for double and triple clicks.
	lastPress time.Time
	lastLine  int
	lastCol   int
	presses   int
}

// span returns the selection in reading order, whichever way it was dragged.
func (s textSelection) span() (startLine, startCol, endLine, endCol int) {
	if s.anchorLine < s.headLine || (s.anchorLine == s.headLine && s.anchorCol <= s.headCol) {
		return s.anchorLine, s.anchorCol, s.headLine, s.headCol
	}
	return s.headLine, s.headCol, s.anchorLine, s.anchorCol
}

// empty reports whether the span covers no characters, which is what a plain
// click leaves behind.
func (s textSelection) empty() bool {
	return s.anchorLine == s.headLine && s.anchorCol == s.headCol
}

// selectionTarget returns the viewport a selection applies to and a name for
// it, matching the viewport View draws. Both come from here so a selection
// cannot end up recorded against one view and painted onto another.
//
// It returns nil where the view has nothing to select: the "no table data" and
// "no trace data" messages are drawn through a throwaway viewport, so there is
// no content to anchor to.
func (m *MainModel) selectionTarget() (*viewport.Model, string) {
	switch {
	case m.viewMode == "ai" && m.aiConversationActive:
		return &m.aiConversationViewport, "ai"
	case m.viewMode == "trace" && m.hasTrace:
		return &m.traceViewport, "trace"
	case m.viewMode == "table" && m.hasTable:
		return &m.tableViewport, "table"
	case m.viewMode == "trace" || m.viewMode == "table":
		return nil, ""
	default:
		return &m.historyViewport, "history"
	}
}

// stickyHeaderRows is how many rows at the top of the viewport show the frozen
// table header rather than the lines the viewport is scrolled to.
//
// Those rows are outside the selection: they show the top of the table, not the
// content at that scroll position, so a selection anchored there would copy
// something other than what is under the pointer. View draws the header only
// when this is non-zero, so the two agree by construction.
func (m *MainModel) stickyHeaderRows() int {
	if m.viewMode == "table" && m.hasTable && m.tableViewport.YOffset() > 0 &&
		len(m.tableHeaders) > 0 && m.lastTableData != nil && m.columnWidths != nil {
		return stickyHeaderHeight
	}
	return 0
}

// docPosition converts a screen position into a line and column of the active
// viewport's content.
func (m *MainModel) docPosition(col, row int) (line, column int, ok bool) {
	vp, _ := m.selectionTarget()
	if vp == nil || col < 0 {
		return 0, 0, false
	}

	top := tabBarHeight + m.stickyHeaderRows()
	bottom := tabBarHeight + vp.Height()
	if row < top || row >= bottom {
		return 0, 0, false
	}

	return vp.YOffset() + row - tabBarHeight, col, true
}

// beginSelection starts a drag, or picks out a word or a line if this press
// follows others on the same spot.
func (m *MainModel) beginSelection(col, row int) (*MainModel, tea.Cmd) {
	line, column, ok := m.docPosition(col, row)
	if !ok {
		m.clearSelection()
		return m, nil
	}
	_, view := m.selectionTarget()

	// A press soon after one on the same cell counts up: two for a word, three
	// for a line, and a fourth starts over.
	presses := 1
	if m.selection.view == view && m.selection.lastLine == line &&
		m.selection.lastCol == column && time.Since(m.selection.lastPress) < multiClickInterval {
		presses = m.selection.presses%3 + 1
	}

	m.selection = textSelection{
		active:     true,
		dragging:   true,
		view:       view,
		anchorLine: line,
		anchorCol:  column,
		headLine:   line,
		headCol:    column,
		lastPress:  time.Now(),
		lastLine:   line,
		lastCol:    column,
		presses:    presses,
	}

	switch presses {
	case 2:
		m.selectWord(line, column)
	case 3:
		m.selectLine(line)
	}
	return m, nil
}

// extendSelection moves the far end of the selection to follow the pointer.
//
// Dragging past the top or bottom edge scrolls a line at a time, so a selection
// can run further than one screen. Because the span is held in content
// coordinates, the scroll moves the text and the highlight together.
func (m *MainModel) extendSelection(col, row int) (*MainModel, tea.Cmd) {
	if !m.selection.dragging {
		return m, nil
	}
	vp, view := m.selectionTarget()
	if vp == nil || view != m.selection.view {
		return m, nil
	}

	top := tabBarHeight + m.stickyHeaderRows()
	bottom := tabBarHeight + vp.Height()
	switch {
	case row < top:
		vp.SetYOffset(max(0, vp.YOffset()-1))
		row = top
	case row >= bottom:
		vp.SetYOffset(min(max(0, vp.TotalLineCount()-vp.Height()), vp.YOffset()+1))
		row = bottom - 1
	}

	line, column, ok := m.docPosition(col, row)
	if !ok {
		return m, nil
	}
	m.selection.headLine, m.selection.headCol = line, column
	return m, nil
}

// endSelection finishes a drag and puts the text on the system clipboard.
//
// Bubble Tea writes it with OSC 52, which is the one way a terminal application
// can reach the clipboard of whatever machine the terminal is on - it works
// through ssh and through tmux, where shelling out to xclip does not. Reading
// the clipboard back is a different matter and most terminals refuse it, which
// is why there is still no right-click paste.
//
// The highlight stays up until the next click or keypress, so you can see what
// you got.
func (m *MainModel) endSelection() (*MainModel, tea.Cmd) {
	if !m.selection.dragging {
		return m, nil
	}
	m.selection.dragging = false

	text := m.selectedText()
	if text == "" {
		m.selection.active = false
		return m, nil
	}
	return m, tea.SetClipboard(text)
}

// clearSelection takes the highlight down.
//
// The press bookkeeping is left alone: it is keyed on position and time, so a
// stale entry cannot turn an unrelated later click into a double click.
func (m *MainModel) clearSelection() {
	m.selection.active = false
	m.selection.dragging = false
}

// selectWord grows the selection to the word under a position.
//
// Underscores and digits count as part of a word, because the words worth
// double-clicking here are keyspace, table and column names.
func (m *MainModel) selectWord(line, col int) {
	cells, ok := m.documentCells(line)
	if !ok || col >= len(cells) || !wordCell(cells[col]) {
		return
	}

	start, end := col, col+1
	for start > 0 && wordCell(cells[start-1]) {
		start--
	}
	for end < len(cells) && wordCell(cells[end]) {
		end++
	}

	m.selection.anchorLine, m.selection.anchorCol = line, start
	m.selection.headLine, m.selection.headCol = line, end
}

// selectLine takes the whole line.
func (m *MainModel) selectLine(line int) {
	cells, ok := m.documentCells(line)
	if !ok {
		return
	}
	m.selection.anchorLine, m.selection.anchorCol = line, 0
	m.selection.headLine, m.selection.headCol = line, len(cells)
}

// wordCell reports whether a character belongs to a word.
func wordCell(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// documentLines splits the active viewport's content into lines.
func (m *MainModel) documentLines() ([]string, bool) {
	vp, view := m.selectionTarget()
	if vp == nil || view != m.selection.view {
		return nil, false
	}
	return strings.Split(vp.GetContent(), "\n"), true
}

// documentCells returns one line of content as one rune per column.
func (m *MainModel) documentCells(line int) ([]rune, bool) {
	lines, ok := m.documentLines()
	if !ok || line < 0 || line >= len(lines) {
		return nil, false
	}
	return lineCells(ansi.Strip(lines[line])), true
}

// lineCells lays a string out one rune per column, so a column index can be
// used as a slice index.
//
// A double-width character fills two columns; the second holds a zero rune,
// which is how the terminal shows it - one character, two cells.
func lineCells(s string) []rune {
	var cells []rune
	for _, r := range s {
		w := ansi.StringWidth(string(r))
		if w <= 0 {
			// A combining mark sits on the character before it and takes no
			// column of its own.
			continue
		}
		cells = append(cells, r)
		for range w - 1 {
			cells = append(cells, 0)
		}
	}
	return cells
}

// selectedText is the selection as plain text, the way a terminal hands it
// over: styling removed and trailing spaces dropped, so a selection dragged
// across a padded table does not paste as a field of blanks.
func (m *MainModel) selectedText() string {
	if !m.selection.active || m.selection.empty() {
		return ""
	}
	lines, ok := m.documentLines()
	if !ok {
		return ""
	}

	startLine, startCol, endLine, endCol := m.selection.span()
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	if startLine > endLine {
		return ""
	}

	out := make([]string, 0, endLine-startLine+1)
	for line := startLine; line <= endLine; line++ {
		text := ansi.Strip(lines[line])
		from, to := 0, ansi.StringWidth(text)
		if line == startLine {
			from = startCol
		}
		if line == endLine && endCol < to {
			to = endCol
		}
		if to <= from {
			out = append(out, "")
			continue
		}
		out = append(out, strings.TrimRight(ansi.Cut(text, from, to), " "))
	}
	return strings.Join(out, "\n")
}

// highlightSelection paints the selection onto the rendered viewport.
//
// Row i of the rendered viewport is content line YOffset+i, the same mapping
// docPosition recorded the span through, so the highlight stays on the
// characters the pointer covered however far the view has scrolled since.
func (m *MainModel) highlightSelection(section string) string {
	if !m.selection.active || m.selection.empty() {
		return section
	}
	vp, view := m.selectionTarget()
	if vp == nil || view != m.selection.view {
		return section
	}

	startLine, startCol, endLine, endCol := m.selection.span()
	offset := vp.YOffset()
	sticky := m.stickyHeaderRows()

	rows := strings.Split(section, "\n")
	for i := range rows {
		if i < sticky {
			continue
		}
		line := offset + i
		if line < startLine || line > endLine {
			continue
		}

		from, to := 0, ansi.StringWidth(rows[i])
		if line == startLine {
			from = startCol
		}
		if line == endLine && endCol < to {
			to = endCol
		}
		rows[i] = highlightColumns(rows[i], from, to)
	}
	return strings.Join(rows, "\n")
}

// highlightColumns marks a range of columns of an already styled line.
//
// The selected run loses its own colours first. A solid block is what a
// terminal's own selection looks like, and layering a highlight over the
// greens and blues already in a result set gives neither.
func highlightColumns(line string, from, to int) string {
	width := ansi.StringWidth(line)
	if from < 0 {
		from = 0
	}
	if to > width {
		to = width
	}
	if to <= from {
		return line
	}

	// The resets keep the styling either side of the highlight from leaking
	// into it, or the highlight's own colours from running on past it.
	return ansi.Cut(line, 0, from) + "\x1b[0m" +
		selectionStyle.Render(ansi.Strip(ansi.Cut(line, from, to))) + "\x1b[0m" +
		ansi.Cut(line, to, width)
}

// defaultMouseEnabled says whether a new session takes the mouse.
//
// It does. The tabs and the settings on the bottom line are no use behind a
// command nobody has heard of, and taking the mouse no longer costs the user
// their text selection now that cqlai draws that itself. MOUSE OFF is there for
// anyone who wants the terminal's own behaviour back.
const defaultMouseEnabled = true
