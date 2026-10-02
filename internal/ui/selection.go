package ui

import (
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/config"
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

	// left and right bound the selection to one pane, for a view that draws
	// two of them side by side. right is exclusive; both zero means the whole
	// line, which is every view but the schema browser.
	//
	// Without it, dragging down the definition took the tree with it: the rule
	// for a flowing selection gives the first line from its start column, the
	// last line up to its end column, and every line between whole - and in
	// this view a whole line is "tree │ definition".
	left, right int

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

// selectionSource is what a selection is anchored to: the lines on screen, and
// which of them is drawn at the top.
//
// Most views are a viewport, and for those it is the viewport's content and
// scroll offset. The schema browser is not - it is two panes drawn side by side,
// each scrolling on its own - so what it offers is the block of lines it just
// drew. The selection does not care which it is: a span is lines and columns of
// what is on screen either way.
type selectionSource struct {
	// vp is the viewport behind the lines, where there is one. Dragging past
	// the top or bottom edge scrolls it; a view without one stops at the edge.
	vp *viewport.Model

	view   string // names it, so a span cannot be painted onto another view
	lines  []string
	offset int // content line drawn on the first row
	height int // rows on screen
}

// selectionTarget returns what a selection applies to, matching what View
// draws. Both come from here so a selection cannot end up recorded against one
// view and painted onto another.
//
// It returns nothing where the view has nothing to select: the "no table data"
// and "no trace data" messages are drawn through a throwaway viewport, so there
// is no content to anchor to.
func (m *MainModel) selectionTarget() (selectionSource, bool) {
	fromViewport := func(vp *viewport.Model, view string) (selectionSource, bool) {
		return selectionSource{
			vp:     vp,
			view:   view,
			lines:  strings.Split(vp.GetContent(), "\n"),
			offset: vp.YOffset(),
			height: vp.Height(),
		}, true
	}

	switch {
	case m.viewMode == "schema":
		// What it drew last, which is what is on screen: both panes, the
		// headings and the divider between them. A definition is text like any
		// other view's, and copying it out of here is the point of having it.
		if len(m.schema.drawn) == 0 {
			return selectionSource{}, false
		}
		return selectionSource{
			view:   "schema",
			lines:  m.schema.drawn,
			height: len(m.schema.drawn),
		}, true
	case m.viewMode == "ai" && m.aiConversationActive:
		return fromViewport(&m.aiConversationViewport, "ai")
	case m.viewMode == "trace":
		// What it drew last, which is what is on screen: the trace, and the
		// analysis under it. Both are worth copying out, and the view is two
		// panes and a rule rather than the one viewport it used to be.
		if len(m.trace.drawn) == 0 {
			return selectionSource{}, false
		}
		return selectionSource{
			view:   "trace",
			lines:  m.trace.drawn,
			height: len(m.trace.drawn),
		}, true
	case m.viewMode == "table" && m.hasTable:
		return fromViewport(&m.tableViewport, "table")
	case m.viewMode == "table":
		return selectionSource{}, false
	default:
		return fromViewport(&m.historyViewport, "history")
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
	// Only a boxed table has a header to freeze. ASCII art and JSON lines are
	// text, and drawing a table header over them puts columns on screen that
	// nothing below them lines up with.
	if m.resultFormat != config.OutputFormatTable {
		return 0
	}
	if m.viewMode == "table" && m.hasTable && m.tableViewport.YOffset() > 0 &&
		len(m.tableHeaders) > 0 && m.lastTableData != nil && m.columnWidths != nil {
		// headerBlock decides how many rows it draws - four with a detail row
		// under the names, three without - and this has to be the same number
		// or the frozen header covers the wrong rows.
		return m.headerRowCount(m.tableHeaders)
	}
	return 0
}

// docPosition converts a screen position into a line and column of the active
// viewport's content.
func (m *MainModel) docPosition(col, row int) (line, column int, ok bool) {
	source, found := m.selectionTarget()
	if !found || col < 0 {
		return 0, 0, false
	}

	top := m.viewTop() + m.stickyHeaderRows()
	bottom := m.viewTop() + source.height
	if row < top || row >= bottom {
		return 0, 0, false
	}

	return source.offset + row - m.viewTop(), col, true
}

// beginSelection starts a drag, or picks out a word or a line if this press
// follows others on the same spot.
func (m *MainModel) beginSelection(col, row int) (*MainModel, tea.Cmd) {
	line, column, ok := m.docPosition(col, row)
	if !ok {
		m.clearSelection()
		return m, nil
	}
	source, _ := m.selectionTarget()
	view := source.view

	// A view drawn as two panes side by side bounds the selection to the one
	// the drag began in. Everywhere else a line is a line.
	left, right := 0, 0
	if view == "schema" {
		left, right = m.selectionPane(col)
	}

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
		left:       left,
		right:      right,
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
	source, found := m.selectionTarget()
	if !found || source.view != m.selection.view {
		return m, nil
	}

	top := m.viewTop() + m.stickyHeaderRows()
	bottom := m.viewTop() + source.height
	switch {
	case row < top:
		m.scrollUnderDrag(source, -1)
		row = top
	case row >= bottom:
		m.scrollUnderDrag(source, 1)
		row = bottom - 1
	}

	line, column, ok := m.docPosition(col, row)
	if !ok {
		return m, nil
	}
	m.selection.headLine, m.selection.headCol = line, column
	return m, nil
}

// scrollUnderDrag moves the content along when a drag has run off an edge, so
// a selection can be longer than the screen.
//
// One function for both edges. The guard against a view with no viewport was
// on the top edge only, so dragging off the bottom of the schema browser - the
// view most likely to hold more than fits, and the one someone is most likely
// to be copying out of - dereferenced nil and took cqlai down with it.
func (m *MainModel) scrollUnderDrag(source selectionSource, by int) {
	if source.vp != nil {
		furthest := max(0, source.vp.TotalLineCount()-source.vp.Height())
		source.vp.SetYOffset(min(furthest, max(0, source.vp.YOffset()+by)))
		return
	}

	// Nothing else scrolls under a drag. The schema browser and the trace draw
	// as blocks, and their selection is over what was drawn rather than over a
	// document with a scroll offset - so scrolling would move the text out from
	// under the span and copy whatever happened to land there instead. The drag
	// stops at the edge, which means a definition longer than the pane cannot
	// yet be selected whole.
}

// endSelection finishes a drag and puts the text on the system clipboard.
//
// It goes out twice. Bubble Tea writes it with OSC 52, which is the one way a
// terminal application can reach the clipboard of whatever machine the terminal
// is on - it works through ssh and through tmux, where shelling out to xclip
// does not. But terminals are free to ignore OSC 52 and several do, so the
// machine cqlai runs on is told as well, through pbcopy or its equivalent.
// Locally one of the two lands; remotely OSC 52 is the one that can.
//
// Reading the clipboard back is a different matter and most terminals refuse
// it, which is why there is still no right-click paste.
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

	// Kept as well as sent. A right click pastes this when the terminal will
	// not say what its clipboard holds, which is most of them.
	m.lastCopied = text

	// Both routes. OSC 52 is the one that reaches the terminal's own machine
	// through ssh or tmux; the command line is the one that works on the
	// terminals which ignore OSC 52, macOS Terminal.app among them.
	return m, tea.Batch(tea.SetClipboard(text), writeSystemClipboard(text))
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

// documentLines is the content a selection is over, by line.
func (m *MainModel) documentLines() ([]string, bool) {
	source, found := m.selectionTarget()
	if !found || source.view != m.selection.view {
		return nil, false
	}
	return source.lines, true
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
// columnsOn is the part of a line the selection covers.
//
// A flowing selection, clamped to the pane it began in: the first line from
// where the drag started, the last line to where it ended, and the lines
// between from edge to edge of that pane. The copy and the highlight both ask
// here, so what is painted is what is copied.
func (s textSelection) columnsOn(line, width int) (from, to int) {
	startLine, startCol, endLine, endCol := s.span()

	left, right := s.left, s.right
	if right == 0 {
		right = width
	}
	right = min(right, width)

	from, to = left, right
	if line == startLine {
		from = max(startCol, left)
	}
	if line == endLine {
		to = min(endCol, right)
	}
	return from, to
}

func (m *MainModel) selectedText() string {
	if !m.selection.active || m.selection.empty() {
		return ""
	}
	lines, ok := m.documentLines()
	if !ok {
		return ""
	}

	startLine, _, endLine, _ := m.selection.span()
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
		from, to := m.selection.columnsOn(line, ansi.StringWidth(text))
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
	source, found := m.selectionTarget()
	if !found || source.view != m.selection.view {
		return section
	}

	startLine, _, endLine, _ := m.selection.span()
	offset := source.offset
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

		from, to := m.selection.columnsOn(line, ansi.StringWidth(rows[i]))
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
