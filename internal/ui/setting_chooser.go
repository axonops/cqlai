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
	selected int    // only used to centre the scroll window on the current value
	current  string // marked in the list
	anchorX  int    // column the field starts at
}

// chooserMaxHeight caps the list so it cannot cover the whole screen. Anything
// longer scrolls.
const chooserMaxHeight = 10

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

// closeSettingChooser dismisses the list without changing anything.
func (m *MainModel) closeSettingChooser() {
	m.chooser = settingChooser{}
}

// chooserWindow returns the slice of choices to draw, and where the selection
// sits within it, so a list longer than the space available scrolls.
func (c settingChooser) window(height int) (start, end int) {
	if height <= 0 || height >= len(c.choices) {
		return 0, len(c.choices)
	}

	// Keep the current value in view, roughly centred, so a long list opens
	// showing where you already are.
	start = c.selected - height/2
	if start < 0 {
		start = 0
	}
	if start+height > len(c.choices) {
		start = len(c.choices) - height
	}
	return start, start + height
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
	available := screenHeight - 1 - 2
	rows := min(len(c.choices), min(chooserMaxHeight, available))
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
	// A leading space, a gap, and the current-value marker.
	inner += 3

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
		// then the marker. Anything else and the box's borders do not line up
		// with the width chooserGeometry reports, which is what choiceAt tests
		// clicks against.
		text := ansi.Truncate(choice, max(g.innerWidth-2, 0), "…")
		fill := max(g.innerWidth-lipgloss.Width(text)-2, 0)
		line := " " + text + strings.Repeat(" ", fill) + marker
		b.WriteString(style.Render(line))
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

	// Same path as typing it, so the transcript shows what changed and the
	// status line updates the way it always has.
	result := router.ProcessCommand(command, m.session, m.sessionManager)

	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	if text, ok := result.(string); ok && text != "" {
		m.fullHistoryContent += "\n" + text

		// A USE leaves three places tracking the keyspace to be told, the same
		// as when one is typed at the prompt. Not switching to the console
		// view, though: you clicked a setting, so whatever you were reading
		// should still be in front of you.
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
