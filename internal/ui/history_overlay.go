package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Placing the two history lists, and working out which command a click landed
// on.
//
// View draws them and the mouse tests clicks against them, both through the
// functions here. Two copies of the placement is two chances for a click to
// apply the row above the one you pointed at - the trap the status line and the
// settings chooser both avoid the same way.
//
// Neither list is all commands: there is a border, a title, sometimes a query
// line or a scroll arrow, then the commands, then a rule and a line of hints.
// So a screen row only means something once those are counted, which is what
// commandList.firstItemRow does.

const (
	// historyModalRows is how many commands either list shows at once. The
	// keyboard handlers scroll by the same figure.
	historyModalRows = 10

	// searchPrompt stands in for an empty Ctrl+R query.
	searchPrompt = "(type to search)"

	historyHint = "Type to search • ↑↓ PgUp/PgDn: Move • Enter or click: Use • Esc: Close"
)

// historyOverlay builds the list of commands and says where it goes.
//
// One list, whether it was opened by clicking History or with Ctrl+R. They were
// two separate pieces of state doing the same job, and keeping both meant the
// clicked one could not be typed into to narrow it down - the one thing the
// Ctrl+R one was for.
func (m *MainModel) historyOverlay(screenWidth, screenHeight int) (commandList, Layer, bool) {
	if !m.historySearchMode {
		return commandList{}, Layer{}, false
	}

	query := m.historySearchQuery
	if query == "" {
		query = searchPrompt
	}

	list := commandList{
		title:    "Command History",
		search:   query,
		items:    m.historySearchResults,
		selected: m.historySearchIndex,
		scroll:   m.historySearchScrollOffset,
		rows:     historyModalRows,
		width:    listWidth(screenWidth, append(slices.Clone(m.historySearchResults), historyHint, " Search: "+query)...),
		hint:     historyHint,
	}
	return list, overlayLayer(list.render(m.styles), screenHeight), true
}

// overlayLayer places a rendered box at the left, just above the prompt.
func overlayLayer(content string, screenHeight int) Layer {
	lines := strings.Split(content, "\n")

	width := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}

	return Layer{
		Content: content,
		X:       0,
		Y:       max(screenHeight-len(lines)-2, 0),
		Width:   width,
		Height:  len(lines),
		ZIndex:  100,
	}
}

// commandAt returns the command under a screen position, if the click is on one
// of the rows showing a command.
//
// first is the index of the topmost command shown and count how many are shown,
// so this works the same for both lists.
func commandAt(layer Layer, firstRow, first, count, row, col int) (int, bool) {
	if count <= 0 {
		return 0, false
	}
	if col < layer.X || col >= layer.X+layer.Width {
		return 0, false
	}

	offset := row - layer.Y - firstRow
	if offset < 0 || offset >= count {
		return 0, false
	}
	return first + offset, true
}

// clickHistoryCommand puts the command under the pointer into the prompt, the
// same as pressing Enter on it. It reports whether the click was on a command.
//
// Nothing runs: the command lands in the prompt and waits, so a mis-click
// cannot execute anything.
func (m *MainModel) clickHistoryCommand(col, row int) (*MainModel, tea.Cmd, bool) {
	list, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	if !ok {
		return m, nil, false
	}

	i, hit := commandAt(layer, list.firstItemRow(), list.scroll, list.visibleItems(), row, col)
	if !hit {
		return m, nil, false
	}

	m.historySearchIndex = i
	updated, cmd := m.handleHistorySearchSelect()
	return updated, cmd, true
}

// moveHistorySelection moves the highlighted command by n and keeps it in view.
//
// The wheel and PgUp/PgDn both come through here. They used to reach the view
// behind the list instead, so a list longer than the box could not be got
// through without the arrow keys.
func (m *MainModel) moveHistorySelection(n int) (*MainModel, tea.Cmd) {
	if !m.historySearchMode || len(m.historySearchResults) == 0 {
		return m, nil
	}

	m.historySearchIndex = min(max(m.historySearchIndex+n, 0), len(m.historySearchResults)-1)

	// Move the window only as far as it takes to keep the selection showing,
	// so paging back and forth does not jump the list about.
	m.historySearchScrollOffset = min(m.historySearchScrollOffset, m.historySearchIndex)
	m.historySearchScrollOffset = max(m.historySearchScrollOffset, m.historySearchIndex-historyModalRows+1)
	m.historySearchScrollOffset = max(m.historySearchScrollOffset, 0)
	return m, nil
}

// latestCommand is what the History field shows: the command run in this
// session, or failing that the newest one in the history file.
//
// The file is read at startup, so there is almost always something to show. The
// field said "-" until the first command of the session, which was not that
// there was no history - only that this session had not added to it yet.
func (m *MainModel) latestCommand() string {
	if m.lastCommand != "" {
		return m.lastCommand
	}
	if len(m.commandHistory) > 0 {
		return m.commandHistory[len(m.commandHistory)-1]
	}
	return ""
}

// openHistoryList shows the commands, opened on the most recent.
//
// It is the Ctrl+R list: the same box, already showing everything, and typing
// narrows it. Clicking History again closes it, like the lists on the status
// line below.
func (m *MainModel) openHistoryList() (*MainModel, tea.Cmd) {
	if !m.historySearchMode && len(m.commandHistory) == 0 {
		return m, nil
	}
	return m.handleCtrlR()
}

// closeHistorySearch puts the command list away.
//
// It was five assignments written out six times across five files, and two of
// those six left out the scroll offset - so closing a scrolled list with Escape
// and opening it again showed the middle of the history with the selection
// somewhere above it.
func (m *MainModel) closeHistorySearch() {
	m.historySearchMode = false
	m.historySearchQuery = ""
	m.historySearchResults = []string{}
	m.historySearchIndex = 0
	m.historySearchScrollOffset = 0
}

// inHistoryOverlay reports whether a press landed inside the command list.
func (m *MainModel) inHistoryOverlay(col, row int) bool {
	_, layer, ok := m.historyOverlay(m.windowWidth, m.windowHeight)
	if !ok {
		return false
	}
	return col >= layer.X && col < layer.X+layer.Width &&
		row >= layer.Y && row < layer.Y+layer.Height
}
