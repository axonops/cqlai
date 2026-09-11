package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// The confirmation dialog: a question with two answers, over everything else.
//
// It is drawn from one description of its layout, the same as the menus and the
// windows, so a click lands on the button drawn under the pointer. It had no
// mouse handling at all before: the buttons were there to be pressed with Enter
// and did nothing when clicked, which is no way to answer a question raised by
// something that was clicked.

// ModalType represents the type of modal
type ModalType int

const (
	ModalNone ModalType = iota
	ModalConfirmDangerous
	ModalConfirmQuit
)

// Modal represents a modal dialog
type Modal struct {
	Type     ModalType
	Title    string
	Message  string
	Command  string // the statement in question, where there is one
	Choices  []string
	Selected int
	Width    int
}

// NewConfirmationModal creates a new confirmation modal for dangerous commands
func NewConfirmationModal(command string) Modal {
	return Modal{
		Type:     ModalConfirmDangerous,
		Title:    "Confirm destructive command",
		Message:  "This command may permanently modify or delete data:",
		Command:  command,
		Choices:  []string{"Cancel", "Execute"},
		Selected: 0, // Default to Cancel for safety
		Width:    60,
	}
}

// NewQuitModal asks whether to leave.
//
// Only the FILE menu raises it. Ctrl+C and Ctrl+D ask their own way - press it
// again - because they are single keys next to several others, and typing EXIT
// does not ask at all, because you typed it. What the menu adds is a question
// that can be answered with the mouse, since that is how its entry was picked.
func NewQuitModal() Modal {
	return Modal{
		Type:     ModalConfirmQuit,
		Title:    "Quit cqlai?",
		Message:  "The connection closes and anything on screen is lost.",
		Choices:  []string{"Cancel", "Quit"},
		Selected: 0, // Default to Cancel: the question is asked to be answered, not waved past
		Width:    50,
	}
}

// NextChoice moves to the next choice
func (m *Modal) NextChoice() {
	m.Selected = (m.Selected + 1) % len(m.Choices)
}

// PrevChoice moves to the previous choice
func (m *Modal) PrevChoice() {
	m.Selected--
	if m.Selected < 0 {
		m.Selected = len(m.Choices) - 1
	}
}

// modalKeys is the line under the buttons.
const modalKeys = "← →/Tab: Move   Enter: Choose   Esc: Cancel"

// modalButtonGap separates the two answers.
const modalButtonGap = "     "

// modalGeometry is where the dialog sits and what is on each of its rows.
// Drawing and clicking both come through here.
type modalGeometry struct {
	x, y          int
	width, height int
	inner         int
	rows          []string
	buttonRow     int
}

// inner is the dialog's content width, inside the border and the padding.
func (m Modal) innerWidth() int {
	width := max(m.Width-4, 20)
	for _, line := range []string{m.Title, m.Message, modalKeys, m.buttons()} {
		width = max(width, lipgloss.Width(line))
	}
	return width
}

// buttonLabel is one answer, with the two columns in front of it that carry the
// marker - the same shape the buttons on every other window have.
func buttonLabel(choice string) string {
	return "  [ " + choice + " ]"
}

// buttons is the row of answers, as it is drawn.
func (m Modal) buttons() string {
	labels := make([]string, 0, len(m.Choices))
	for _, choice := range m.Choices {
		labels = append(labels, buttonLabel(choice))
	}
	return strings.Join(labels, modalButtonGap)
}

// buttonStarts is where each answer begins, measured from the left of the
// content. Drawing and clicking take the same numbers from here.
func (m Modal) buttonStarts(inner int) []int {
	row := m.buttons()
	left := max((inner-lipgloss.Width(row))/2, 0)

	starts := make([]int, 0, len(m.Choices))
	at := left
	for _, choice := range m.Choices {
		starts = append(starts, at)
		at += lipgloss.Width(buttonLabel(choice)) + len(modalButtonGap)
	}
	return starts
}

// geometry places the dialog in the middle of the screen.
func (m Modal) geometry(screenWidth, screenHeight int) (modalGeometry, bool) {
	if m.Type == ModalNone {
		return modalGeometry{}, false
	}

	inner := min(m.innerWidth(), max(screenWidth-6, 1))

	rows := []string{"", m.Title, "", m.Message}
	if m.Command != "" {
		rows = append(rows, "", m.Command)
	}
	rows = append(rows, "")
	buttonRow := len(rows)
	rows = append(rows, m.buttons(), "", modalKeys, "")

	width := inner + 4 + 2 // the padding either side, and the border
	height := len(rows) + 2
	if width > screenWidth || height > screenHeight {
		return modalGeometry{}, false
	}

	return modalGeometry{
		x:         max((screenWidth-width)/2, 0),
		y:         max((screenHeight-height)/2, 0),
		width:     width,
		height:    height,
		inner:     inner,
		rows:      rows,
		buttonRow: buttonRow,
	}, true
}

// modalBackground is the destructive-command dialog's own ground, so what is
// behind it does not show through the gaps between its words.
var modalBackground = lipgloss.Color("#1A1A1A")

// modalPalette is how a dialog is coloured.
//
// The destructive-command dialog keeps the colours it has always had: a warning
// border, and answers filled green and red, because the two answers are not
// alike and the difference is the point of asking. Everything else is drawn
// like the rest of cqlai - accent border, accent on whichever answer the cursor
// is on - rather than borrowing that alarm for a question about leaving.
type modalPalette struct {
	border     color.Color
	background color.Color // nil for a dialog drawn like the other windows
	fill       bool        // the chosen answer is filled in rather than marked
}

func (m Modal) palette(styles *Styles) modalPalette {
	if m.Type == ModalConfirmDangerous {
		return modalPalette{border: styles.Warn, background: modalBackground, fill: true}
	}
	return modalPalette{border: styles.Accent}
}

// ground is the style everything in the dialog is drawn on.
func (p modalPalette) ground() lipgloss.Style {
	style := lipgloss.NewStyle()
	if p.background != nil {
		style = style.Background(p.background)
	}
	return style
}

// chosenStyle is how the answer the cursor is on is drawn.
func (p modalPalette) chosenStyle(choice int, styles *Styles) lipgloss.Style {
	if !p.fill {
		return p.ground().Foreground(styles.Accent).Bold(true)
	}

	// Filled: the safe answer in green, the one that does something in red.
	colour := styles.Ok
	if choice > 0 {
		colour = styles.Error
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(colour).
		Bold(true)
}

// Layer draws the dialog over the view.
func (m Modal) Layer(screenWidth, screenHeight int, styles *Styles) (Layer, bool) {
	g, ok := m.geometry(screenWidth, screenHeight)
	if !ok {
		return Layer{}, false
	}

	palette := m.palette(styles)
	base := palette.ground()
	titleStyle := base.Foreground(palette.border).Bold(true)
	textStyle := base.Foreground(styles.MutedText.GetForeground())
	commandStyle := base.Foreground(styles.AccentText.GetForeground())
	if palette.background != nil {
		// The statement in question, on a band of its own: it is what is being
		// agreed to, rather than part of the sentence above it.
		commandStyle = commandStyle.Background(lipgloss.Color("#2D2D2D"))
	}
	keyStyle := base.Foreground(styles.MutedText.GetForeground()).Italic(true)

	centre := func(text string) string {
		gap := max(g.inner-lipgloss.Width(text), 0)
		left := gap / 2
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", gap-left)
	}

	lines := make([]string, 0, len(g.rows))
	for i, row := range g.rows {
		switch {
		case i == g.buttonRow:
			lines = append(lines, m.renderButtons(g, styles))
		case row == m.Title && row != "":
			lines = append(lines, titleStyle.Render(centre(row)))
		case row == m.Command && row != "":
			lines = append(lines, commandStyle.Render(centre(row)))
		case row == modalKeys:
			lines = append(lines, keyStyle.Render(centre(row)))
		default:
			lines = append(lines, textStyle.Render(centre(row)))
		}
	}

	box := base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Padding(0, 2).
		Render(strings.Join(lines, "\n"))
	if palette.background != nil {
		box = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(palette.border).
			BorderBackground(palette.background).
			Background(palette.background).
			Padding(0, 2).
			Render(strings.Join(lines, "\n"))
	}

	return Layer{
		Content: box,
		X:       g.x,
		Y:       g.y,
		Width:   g.width,
		Height:  g.height,
		ZIndex:  400, // over every window: it is a question, and it is in the way on purpose
	}, true
}

// renderButtons draws the answers, the chosen one filled in.
//
// The safe answer is filled in green and the one that does something in red,
// which is the difference worth seeing before pressing Enter.
func (m Modal) renderButtons(g modalGeometry, styles *Styles) string {
	palette := m.palette(styles)
	base := palette.ground()
	plain := base.Foreground(styles.MutedText.GetForeground())

	row := strings.Repeat(" ", g.inner)
	drawn := ""
	at := 0
	for i, start := range m.buttonStarts(g.inner) {
		marker, style := "  ", plain
		if i == m.Selected {
			marker, style = "> ", palette.chosenStyle(i, styles)
		}

		// The marker sits outside a filled answer rather than inside it, so the
		// fill is the same size whichever answer the cursor is on.
		label := "[ " + m.Choices[i] + " ]"
		drawn += base.Render(row[at:start])
		if palette.fill {
			drawn += plain.Render(marker) + style.Render(label)
		} else {
			drawn += style.Render(marker + label)
		}
		at = start + lipgloss.Width(buttonLabel(m.Choices[i]))
	}
	return drawn + base.Render(row[at:])
}

// buttonAt returns the answer a press landed on.
func (m Modal) buttonAt(screenWidth, screenHeight, col, row int) (int, bool) {
	g, ok := m.geometry(screenWidth, screenHeight)
	if !ok || row != g.y+1+g.buttonRow {
		return 0, false
	}

	// One column for the border, two for the padding.
	at := col - g.x - 3
	for i, start := range m.buttonStarts(g.inner) {
		if at >= start && at < start+lipgloss.Width("[ "+m.Choices[i]+" ]") {
			return i, true
		}
	}
	return 0, false
}
