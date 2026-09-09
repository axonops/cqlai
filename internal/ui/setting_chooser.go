package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/router"
	"github.com/charmbracelet/x/ansi"
)

// The list that opens when you click a setting on the status line.
//
// The status line is the last row, so the list opens upward, anchored to the
// column of the field it belongs to. Anywhere else on screen and it would be
// unclear which setting it was offering.

// settingChooser is the open list, if any.
type settingChooser struct {
	active   bool
	setting  string // which status field this belongs to
	label    string // shown above the list, e.g. "CL"
	choices  []string
	selected int    // the value already set, which the list opens centred on
	current  string // marked in the list
	anchorX  int    // column the field starts at

	// Where the visible window starts, once the wheel has moved it. Until then
	// window() centres on selected instead, so a long list opens showing where
	// you already are rather than its alphabetical start.
	scroll   int
	scrolled bool
}

// openSettingChooser opens the list for a status field.
func (m *MainModel) openSettingChooser(setting string, anchorX int, choices []string, current string) {
	if len(choices) == 0 {
		return
	}

	selected := 0
	for i, c := range choices {
		if c == current {
			selected = i
		}
	}

	m.chooser = settingChooser{
		active:   true,
		setting:  setting,
		label:    setting,
		choices:  choices,
		selected: selected,
		current:  current,
		anchorX:  anchorX,
	}
}

const (
	// chooserMinHeight keeps the list usable on a terminal too short for half
	// of it to be worth having.
	chooserMinHeight = 5

	// chooserScrollStep is how far one wheel notch moves the list. The same as
	// the viewports scroll by, so the wheel feels the same wherever it is.
	chooserScrollStep = wheelLines
)

// closeSettingChooser dismisses the list without changing anything.
func (m *MainModel) closeSettingChooser() {
	m.chooser = settingChooser{}
}

// window returns the slice of choices to draw, so a list longer than there is
// room for scrolls rather than losing its tail.
func (c settingChooser) window(height int) (start, end int) {
	if height <= 0 || height >= len(c.choices) {
		return 0, len(c.choices)
	}

	start = c.scroll
	if !c.scrolled {
		// Roughly centred on the value already set.
		start = c.selected - height/2
	}
	start = min(max(start, 0), len(c.choices)-height)
	return start, start + height
}

// scrollChooser moves the open list by n rows.
//
// Without this the wheel went to the view behind the list, so on a cluster with
// more keyspaces than the box has rows the rest of them could not be reached at
// all: there is no keyboard navigation here by design, and nothing else moved
// the window.
func (m *MainModel) scrollChooser(n int) (*MainModel, tea.Cmd) {
	g, ok := m.chooserGeometry(m.windowWidth, m.windowHeight)
	if !ok {
		return m, nil
	}

	// g.first is where the window sits now, so this moves from wherever the
	// list happened to open. window() does the clamping.
	m.chooser.scroll = g.first + n
	m.chooser.scrolled = true
	return m, nil
}

// scrollbarColumn returns one entry per visible row, true where the thumb is.
//
// The thumb's size says how much of the list is showing and its position says
// where in the list you are, which is the thing a long keyspace list needs and
// a plain box cannot tell you. It is empty when the whole list fits, since a
// bar that is always full says nothing.
func scrollbarColumn(rows, first, total int) []bool {
	thumb := make([]bool, max(rows, 0))
	if rows <= 0 || total <= rows {
		return thumb
	}

	size := max(rows*rows/total, 1)

	// Scaled so the thumb reaches the bottom exactly when the last choice is
	// showing, rather than stopping a row short.
	start := 0
	if lastFirst := total - rows; lastFirst > 0 {
		start = first * (rows - size) / lastFirst
	}

	for i := start; i < start+size && i < rows; i++ {
		thumb[i] = true
	}
	return thumb
}

// chooserGeometry works out where the list sits and which choices it shows.
//
// Drawing and clicking both go through this. Working the position out twice is
// how a click ends up applying the row above the one you pointed at.
type chooserGeometry struct {
	x, y          int // top-left of the box, including its border
	width, height int
	first, last   int // range of choices drawn, last exclusive
	innerWidth    int
}

func (m *MainModel) chooserGeometry(screenWidth, screenHeight int) (chooserGeometry, bool) {
	c := m.chooser
	if !c.active || len(c.choices) == 0 {
		return chooserGeometry{}, false
	}

	// Rows above the status line, less the box's own border.
	//
	// Half of that, at most. A fixed cap was wrong in both directions: ten rows
	// wasted a tall terminal, and no cap at all let a long keyspace list cover
	// everything but the status line. Half leaves you able to see what is
	// behind the list whatever size the terminal is, and short lists still take
	// only the rows they need.
	available := screenHeight - 1 - 2
	rows := min(len(c.choices), min(max(available/2, chooserMinHeight), available))
	if rows <= 0 {
		return chooserGeometry{}, false
	}

	first, last := c.window(rows)

	inner := lipgloss.Width(c.label)
	for _, choice := range c.choices {
		if w := lipgloss.Width(choice); w > inner {
			inner = w
		}
	}
	// A leading space, a gap, the current-value marker, and the scrollbar.
	inner += 4

	// A keyspace name can be long enough to push the box off the screen, so the
	// box is capped and the names inside it are truncated to fit.
	if inner > screenWidth-2 {
		inner = max(screenWidth-2, 0)
	}

	width := inner + 2 // borders
	height := rows + 2

	x := c.anchorX
	if x+width > screenWidth {
		x = screenWidth - width
	}
	if x < 0 {
		x = 0
	}

	y := screenHeight - 1 - height
	if y < 0 {
		y = 0
	}

	return chooserGeometry{
		x: x, y: y, width: width, height: height,
		first: first, last: last, innerWidth: inner,
	}, true
}

// choiceAt returns the choice under a screen position, if the click is inside
// the list.
func (m *MainModel) choiceAt(screenWidth, screenHeight, col, row int) (string, bool) {
	g, ok := m.chooserGeometry(screenWidth, screenHeight)
	if !ok {
		return "", false
	}

	// Inside the border, not on it.
	if col <= g.x || col >= g.x+g.width-1 {
		return "", false
	}
	if row <= g.y || row >= g.y+g.height-1 {
		return "", false
	}

	idx := g.first + (row - g.y - 1)
	if idx < g.first || idx >= g.last || idx >= len(m.chooser.choices) {
		return "", false
	}
	return m.chooser.choices[idx], true
}

// viewSettingChooser renders the list and says where to put it.
//
// screenHeight is the full screen; the list sits directly above the status
// line, growing upward, and is clamped to the screen if there is not room.
func (m *MainModel) viewSettingChooser(screenWidth, screenHeight int) (Layer, bool) {
	g, ok := m.chooserGeometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}
	c := m.chooser

	choiceStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D0D0D0"))
	currentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FFD7")).
		Bold(true)

	thumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87D7FF"))
	trackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3a3a3a"))

	thumb := scrollbarColumn(g.last-g.first, g.first, len(c.choices))

	var b strings.Builder
	for i := g.first; i < g.last; i++ {
		choice := c.choices[i]

		marker := " "
		style := choiceStyle
		if choice == c.current {
			marker = "\u25cf"
			style = currentStyle
		}

		// Exactly innerWidth columns: a leading space, the name padded out,
		// the marker, then the scrollbar. Anything else and the box's borders
		// do not line up with the width chooserGeometry reports, which is what
		// choiceAt tests clicks against.
		text := ansi.Truncate(choice, max(g.innerWidth-3, 0), "…")
		fill := max(g.innerWidth-lipgloss.Width(text)-3, 0)
		b.WriteString(style.Render(" " + text + strings.Repeat(" ", fill) + marker))

		switch {
		case len(thumb) == 0 || len(c.choices) <= g.last-g.first:
			b.WriteString(" ")
		case thumb[i-g.first]:
			b.WriteString(thumbStyle.Render("\u2588"))
		default:
			b.WriteString(trackStyle.Render("\u2591"))
		}

		if i < g.last-1 {
			b.WriteString("\n")
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#87D7FF")).
		Render(b.String())

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  200, // above the other overlays; it is the thing just clicked
	}, true
}

// handleSettingChooserKey dismisses the list on any keypress.
//
// There is deliberately no keyboard navigation. Anyone at the keyboard can type
// CONSISTENCY LOCAL_QUORUM directly, which is fewer keystrokes than arrowing
// through a list, so the list is for the mouse and gets out of the way as soon
// as you start typing.
func (m *MainModel) handleSettingChooserKey(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	m.closeSettingChooser()

	// Escape means "just close it". Anything else was aimed at the prompt, so
	// pass it on rather than swallowing a keystroke.
	if msg.String() == "esc" {
		return m, nil
	}
	return m.handleKeyboardInput(msg)
}

// applySettingChoice runs the command the chosen value corresponds to.
//
// It goes through the same meta-command path as typing it, so everything that
// normally follows - the confirmation in the transcript, the status line
// updating - happens the same way. Duplicating the effects here is how the two
// routes end up behaving differently.
func (m *MainModel) applySettingChoice(choice string) (*MainModel, tea.Cmd) {
	c := m.chooser
	m.closeSettingChooser()

	if choice == "" {
		return m, nil
	}

	if choice == c.current {
		// Picking what is already set should do nothing at all, rather than
		// logging a change that did not happen.
		return m, nil
	}

	command := settingCommand(c.setting, choice)
	if command == "" {
		return m, nil
	}

	return m.runCommand(command)
}

// runCommand runs a meta-command the way typing it would, for the controls on
// the bottom line.
//
// Same path as the prompt, so the transcript records what happened and the
// status line updates as it always has. Doing the effects here instead is how
// the two routes end up behaving differently.
//
// It does not switch view: you clicked a control, so whatever you were reading
// should still be in front of you.
func (m *MainModel) runCommand(command string) (*MainModel, tea.Cmd) {
	result := router.ProcessCommand(command, m.session, m.sessionManager)

	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	if text, ok := result.(string); ok && text != "" {
		m.fullHistoryContent += "\n" + text

		// A USE leaves three places tracking the keyspace to be told, the same
		// as when one is typed at the prompt.
		m.adoptKeyspace(text)
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	return m, nil
}

// settingCommand returns the meta-command that sets a value.
func settingCommand(setting, choice string) string {
	switch setting {
	case settingConsistency:
		return "CONSISTENCY " + choice
	case settingPaging:
		return "PAGING " + choice
	case settingOutput:
		return "OUTPUT " + choice
	case settingTracing:
		return "TRACING " + choice
	case settingAutoFetch:
		return "AUTOFETCH " + choice
	case settingKeyspace:
		// Quoted, so a keyspace created with a mixed-case name works. The USE
		// handler strips the quotes before looking the name up, so this is
		// harmless for the ordinary lowercase ones.
		return `USE "` + choice + `"`
	}
	return ""
}
