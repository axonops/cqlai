package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The FILE menu at the left-hand end of the tab line.
//
// Save and Capture are the same operation - write results to a file - and have
// shared one window since #129. They were in two different corners, one on the
// tab line and one on the status line, which was the odd part rather than the
// grouping.
//
// Capture's state stays on the status line, in beside Fetch (#149). It is the
// one setting that keeps doing something after you have stopped thinking about
// it, and a menu would hide that. What moved here is the action.
//
// It is not a commandList or a settingChooser. Those pick a value for something
// and mark the one already set; this picks what to do next, and there is
// nothing in it to be current. What it shares with them is the rule that
// drawing and hit testing come from one description of the layout.

// fileMenuItem is one entry.
//
// An entry either opens the two-step window Save and Capture share, or a form.
// Which one it is depends on how many questions the command asks: a format and
// a path can be asked one at a time, a table and a path and an option cannot.
type fileMenuItem struct {
	label string
	kind  fileKind   // for the two-step window
	form  formAction // for a form
	asks  bool       // true if this entry opens a form
	prefs bool       // true for PREFERENCES, which opens its own window
	quits bool       // true for QUIT, which is the end of it

	// rule is a line across the menu rather than an entry. It cannot be
	// selected, the arrow keys step over it, and a click on it does nothing.
	rule bool
}

// fileMenu is the open menu, if any.
type fileMenu struct {
	active   bool
	selected int
	anchorX  int // column FILE starts at, so the menu hangs under it
}

// fileMenuItems is what the menu offers, in the order it offers them.
//
// AutoSave first: it is a decision about everything that comes next, and the
// one you are most likely to be here to make.
func fileMenuItems() []fileMenuItem {
	return []fileMenuItem{
		{label: "AUTOSAVE", kind: capturing},
		{label: "SOURCE", form: sourcing, asks: true},
		{label: "COPY TO", form: copyingTo, asks: true},
		{label: "COPY FROM", form: copyingFrom, asks: true},
		{rule: true},
		// Below the line because it is not a file operation. The four above
		// move data in or out; this one changes how cqlai starts.
		{label: "PREFERENCES", prefs: true},
		{rule: true},
		// And last, on its own, because leaving is not like either group above
		// it and is the one entry here you cannot undo.
		{label: "QUIT", quits: true},
	}
}

// available reports whether an item can be picked.
//
// SAVE RESULTS was here and is not any more. Everything left is a file
// operation that stands on its own; saving what is on screen depends on there
// being something on screen, which made it the one entry that was sometimes
// dimmed. The SAVE command is unchanged and still opens the same window.
func (m *MainModel) fileItemAvailable(item fileMenuItem) bool {
	if item.rule {
		return false
	}

	// COPY names a keyspace and a table, and without a session there is nothing
	// to name and nothing to complete against. SOURCE reads a file and
	// AUTOSAVE writes them, so neither needs one.
	if item.asks && item.form != sourcing {
		return m.session != nil
	}
	return true
}

// openFileMenu opens the menu under the FILE button.
func (m *MainModel) openFileMenu(anchorX int) (*MainModel, tea.Cmd) {
	if m.fileMenu.active {
		m.fileMenu = fileMenu{}
		return m, nil
	}

	m.fileMenu = fileMenu{active: true, anchorX: anchorX}

	// Open on something you can actually pick, rather than on a dimmed row that
	// does nothing when you press Enter.
	items := fileMenuItems()
	for i, item := range items {
		if m.fileItemAvailable(item) {
			m.fileMenu.selected = i
			break
		}
	}
	return m, nil
}

// closeFileMenu dismisses it without doing anything.
func (m *MainModel) closeFileMenu() {
	m.fileMenu = fileMenu{}
}

// moveFileMenu moves the highlight, skipping what cannot be picked and stopping
// at the ends rather than wrapping: two items are not a list worth cycling.
func (m *MainModel) moveFileMenu(delta int) (*MainModel, tea.Cmd) {
	items := fileMenuItems()
	for i := m.fileMenu.selected + delta; i >= 0 && i < len(items); i += delta {
		if m.fileItemAvailable(items[i]) {
			m.fileMenu.selected = i
			break
		}
	}
	return m, nil
}

// chooseFileMenuItem opens the window for the highlighted entry.
//
// Both open in the middle of the screen. Capture's used to sit just above its
// control on the bottom line, pointing at it; opened from a menu at the top
// there is nothing down there to point at, and one placement rule is better
// than two.
func (m *MainModel) chooseFileMenuItem() (*MainModel, tea.Cmd) {
	items := fileMenuItems()
	if m.fileMenu.selected >= len(items) {
		return m, nil
	}

	item := items[m.fileMenu.selected]
	if !m.fileItemAvailable(item) {
		return m, nil
	}

	m.closeFileMenu()
	switch {
	case item.quits:
		// Asked rather than done. The other three ways out have their own
		// answer to this: Ctrl+C and Ctrl+D want the key twice, and typing EXIT
		// is typing. A menu entry is one press away from the entry above it and
		// sits under a button on the tab line, so it asks.
		m.modal = NewQuitModal()
		return m, nil
	case item.prefs:
		return m.openPreferences()
	case item.asks:
		return m.openFileForm(item.form)
	case item.kind == saving:
		return m.openSavePanel()
	}
	return m.openCapturePanelCentred()
}

// fileMenuGeometry is where the menu sits. Drawing and clicking both come
// through here, so a click cannot land on a different row from the one drawn
// there.
type fileMenuGeometry struct {
	x, y          int
	width, height int
	innerWidth    int
	items         []fileMenuItem
}

func (m *MainModel) fileMenuGeometry(screenWidth, screenHeight int) (fileMenuGeometry, bool) {
	if !m.fileMenu.active {
		return fileMenuGeometry{}, false
	}

	items := fileMenuItems()
	inner := 0
	for _, item := range items {
		inner = max(inner, lipgloss.Width(item.label))
	}
	inner += 2 // a column of padding either side of the widest label

	width := inner + 2
	height := len(items) + 2 // the items, and the border round them
	if width > screenWidth || height > screenHeight {
		return fileMenuGeometry{}, false
	}

	// Hangs under FILE, and pulled back when the button is near the right-hand
	// edge so the menu cannot run off it.
	x := min(m.fileMenu.anchorX, max(screenWidth-width, 0))

	return fileMenuGeometry{
		x:          x,
		y:          tabBarHeight,
		width:      width,
		height:     height,
		innerWidth: inner,
		items:      items,
	}, true
}

// fileMenuItemAt returns the item a screen row covers.
func (m *MainModel) fileMenuItemAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.fileMenuGeometry(screenWidth, screenHeight)
	if !ok {
		return 0, false
	}
	if col < g.x || col >= g.x+g.width {
		return 0, false
	}

	// The border comes before the first item.
	i := row - g.y - 1
	if i < 0 || i >= len(g.items) {
		return 0, false
	}
	return i, true
}

// viewFileMenu draws it.
func (m *MainModel) viewFileMenu(screenWidth, screenHeight int) (Layer, bool) {
	g, ok := m.fileMenuGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	itemStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1c1c1c")).
		Background(m.styles.Accent).
		Bold(true)

	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))

	var rows []string
	for i, item := range g.items {
		if item.rule {
			rows = append(rows, ruleStyle.Render(strings.Repeat("─", g.innerWidth)))
			continue
		}

		text := pad(" "+item.label, g.innerWidth)
		switch {
		case !m.fileItemAvailable(item):
			rows = append(rows, dimStyle.Render(text))
		case i == m.fileMenu.selected:
			rows = append(rows, selectedStyle.Render(text))
		default:
			rows = append(rows, itemStyle.Render(text))
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Accent).
		Render(strings.Join(rows, "\n"))

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  300, // over everything; it was just asked for
	}, true
}
