package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Drawing the SCHEMA view, and working out what a click landed on.
//
// One description of the layout: where the tree stops and the definition
// starts, and which row is which. Two copies of that is how a click selects the
// keyspace above the one pointed at.

const (
	// schemaTreeMin and schemaTreeMax bound the left-hand pane. Wide enough for
	// a keyspace name and the marker in front of it, and never so wide that the
	// definition beside it has nowhere to go.
	schemaTreeMin = 18
	schemaTreeMax = 40

	// schemaDivider separates the panes, in the same grey as the one on the
	// COPY forms.
	schemaDivider = " │ "

	// schemaIndent is how far a table sits inside its keyspace: under the name
	// above it, clear of the marker.
	schemaIndent = "    "

	// schemaHeaderRows is the heading over each pane and the rule under it.
	// Two rows of a screen, so that neither column is a list of names with
	// nothing saying what they are or what the text beside them describes.
	schemaHeaderRows = 2

	// schemaTreeHeading names the left-hand pane. The tables are inside the
	// keyspaces rather than beside them, so the keyspaces are what it lists.
	schemaTreeHeading = "KEYSPACES"
)

// schemaGeometry is where the panes are, and which rows they show.
type schemaGeometry struct {
	width, height int
	treeWidth     int // the left pane, including its scrollbar column
	detailWidth   int
	treeFirst     int // the first tree row drawn
	detailFirst   int // the first line of the definition drawn

	// The definition pane splits in two once there is a review of it: the
	// definition above, the review below, and a line between them that is
	// dragged to give either of them room.
	detailRows int
	ruleRow    int // rows from the top of the panes, or -1 while there is none
	reviewRows int
	reviewTop  int

	// The columns the button on the heading row covers, so that drawing it and
	// clicking it cannot disagree.
	buttonFrom, buttonTo int
}

// schemaGeometryFor places the panes for a screen.
func (m *MainModel) schemaGeometry(width, height int) schemaGeometry {
	tree := min(max(m.schemaTreeWidth(), schemaTreeMin), schemaTreeMax)
	tree = min(tree, max(width/2, 1))

	g := schemaGeometry{
		width:       width,
		height:      max(height-schemaHeaderRows, 1),
		treeWidth:   tree,
		detailWidth: max(width-tree-lipgloss.Width(schemaDivider), 1),
	}

	// The selection stays on screen, and the pane moves only as far as it takes
	// to keep it there: walking a long list should not jump it about.
	rows := len(m.schema.rows())
	first := min(m.schema.scroll, m.schema.selected)
	first = max(first, m.schema.selected-g.height+1)
	first = min(first, max(rows-g.height, 0))
	g.treeFirst = max(first, 0)

	// The button sits at the right of the definition pane's heading.
	button := reviewButton
	if m.schema.review.running {
		button = reviewingNow
	}
	g.buttonTo = max(width-1, 0)
	g.buttonFrom = max(g.buttonTo-lipgloss.Width(button)+1, 0)

	g.detailRows = g.height
	g.ruleRow = -1
	if m.schema.review.open {
		// What the pane was asked for is squeezed to fit rather than written
		// back: a terminal made small and large again gives it back the size
		// it had.
		room := max(g.height-reviewRuleRows, definitionMin+reviewMin)
		g.reviewRows = min(max(m.schema.review.rows, reviewMin), room-definitionMin)
		g.detailRows = room - g.reviewRows
		g.ruleRow = g.detailRows
		g.reviewTop = g.ruleRow + reviewRuleRows
	}

	g.detailFirst = min(max(m.schema.detailScroll, 0), max(len(m.schema.detail)-g.detailRows, 0))
	return g
}

// schemaTreeWidth is how wide the tree wants to be: its widest row, and a
// column for the scrollbar.
func (m *MainModel) schemaTreeWidth() int {
	width := lipgloss.Width(schemaTreeHeading)
	for _, row := range m.schema.rows() {
		width = max(width, lipgloss.Width(m.schemaLine(row)))
	}
	if m.schema.message != "" {
		width = max(width, lipgloss.Width(m.schema.message))
	}
	return width + 2
}

// schemaDetailHeading names what the right-hand pane is showing: a table's
// definition, a keyspace's, or nothing yet.
func (m *MainModel) schemaDetailHeading() string {
	row, ok := m.schema.current()
	if !ok {
		return "DEFINITION"
	}
	if row.table != "" {
		return "TABLE  " + row.key()
	}
	return "KEYSPACE  " + row.keyspace
}

// schemaDetailHeadingRow is the heading over the definition, with the button
// that reviews it at the right-hand end.
func (m *MainModel) schemaDetailHeadingRow(g schemaGeometry, headingStyle lipgloss.Style) string {
	heading := m.schemaDetailHeading()

	button := reviewButton
	style := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	switch {
	case m.schema.review.running:
		button, style = reviewingNow, lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	case len(m.schema.detail) == 0:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(formPlaceholderColour))
	}

	room := g.detailWidth - lipgloss.Width(button)
	if room < lipgloss.Width(heading)+1 {
		// Too narrow for both: the heading says what the pane is, and Alt+A
		// still reviews it.
		return headingStyle.Render(pad(heading, g.detailWidth))
	}
	return headingStyle.Render(pad(heading, room)) + style.Render(button)
}

// schemaDetailRow is one row of the right-hand pane: a line of the definition,
// the line between the panes, or a line of the review under it.
//
// One description of where those are, used to draw them and to work out what a
// press landed on.
func (m *MainModel) schemaDetailRow(g schemaGeometry, row int, detailBar []bool) string {
	textStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	switch {
	case row == g.ruleRow:
		if m.schema.review.dragging {
			ruleStyle = lipgloss.NewStyle().Foreground(m.styles.Accent)
		}
		label := textStyle.Render(reviewHeading) + ruleStyle.Render(dragHint)
		line := strings.Repeat("─", max(g.detailWidth-lipgloss.Width(stripAnsi(label))-1, 0))
		return ruleStyle.Render("─") + label + ruleStyle.Render(line)

	case g.ruleRow >= 0 && row > g.ruleRow:
		return m.schemaReviewRow(g, row-g.reviewTop)
	}

	definition := ""
	if g.detailFirst+row < len(m.schema.detail) {
		definition = m.schema.detail[g.detailFirst+row]
	}

	drawn := textStyle.Render(pad(cut(definition, g.detailWidth-1), g.detailWidth-1))
	return drawn + scrollbarCell(detailBar, row, len(m.schema.detail) > g.detailRows, thumbStyle, trackStyle)
}

// schemaReviewRow is one row of the review pane.
func (m *MainModel) schemaReviewRow(g schemaGeometry, row int) string {
	textStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	if m.schema.review.failed {
		textStyle = m.styles.ErrorText
	}
	headingStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	lines := m.schema.review.lines
	first := min(max(m.schema.review.scroll, 0), max(len(lines)-g.reviewRows, 0))

	text := ""
	if first+row < len(lines) {
		text = lines[first+row]
	}

	style := textStyle
	if isHeadingLine(text) {
		style = headingStyle
	}

	bar := scrollbarColumn(min(g.reviewRows, len(lines)), first, len(lines))
	return style.Render(pad(cut(text, g.detailWidth-1), g.detailWidth-1)) +
		scrollbarCell(bar, row, len(lines) > g.reviewRows, thumbStyle, trackStyle)
}

// schemaLine is a row as it is drawn: a keyspace with the marker saying which
// way it is folded, or a table indented under the one holding it.
//
// Drawing and measuring both come from here, so the pane cannot be sized for
// one thing and drawn with another.
func (m *MainModel) schemaLine(row schemaRow) string {
	if row.table != "" {
		return schemaIndent + row.table
	}
	if m.schema.expanded[row.keyspace] {
		return "▾ " + row.keyspace
	}
	return "▸ " + row.keyspace
}

// viewSchema draws the whole view: the tree, the divider and the definition.
func (m *MainModel) viewSchema(width, height int) string {
	g := m.schemaGeometry(width, height)

	rows := m.schema.rows()
	treeStyle := lipgloss.NewStyle().Foreground(m.styles.MutedText.GetForeground())
	keyspaceStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1c1c1c")).
		Background(m.styles.Accent).
		Bold(true)
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))
	thumbStyle := lipgloss.NewStyle().Foreground(m.styles.Accent)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	// The scrollbars say how much of each side you are looking at. A schema is
	// a long thing in both panes and neither has a border to hint at it.
	treeBar := scrollbarColumn(min(g.height, len(rows)), g.treeFirst, len(rows))
	detailBar := scrollbarColumn(min(g.detailRows, len(m.schema.detail)), g.detailFirst, len(m.schema.detail))

	headingStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)

	lines := make([]string, 0, g.height+schemaHeaderRows)
	lines = append(lines,
		headingStyle.Render(pad(schemaTreeHeading, g.treeWidth))+
			ruleStyle.Render(schemaDivider)+
			m.schemaDetailHeadingRow(g, headingStyle),
		ruleStyle.Render(strings.Repeat("─", g.treeWidth)+schemaDivider+strings.Repeat("─", g.detailWidth)),
	)

	for i := range g.height {
		// The left-hand pane.
		text, style := "", treeStyle
		switch {
		case i == 0 && m.schema.message != "":
			text = m.schema.message
		case g.treeFirst+i < len(rows):
			row := rows[g.treeFirst+i]
			text = m.schemaLine(row)
			switch {
			case g.treeFirst+i == m.schema.selected:
				style = selectedStyle
			case row.table == "":
				style = keyspaceStyle
			}
		}

		left := style.Render(pad(text, g.treeWidth-1))
		left += scrollbarCell(treeBar, i, len(rows) > g.height, thumbStyle, trackStyle)

		lines = append(lines, left+ruleStyle.Render(schemaDivider)+m.schemaDetailRow(g, i, detailBar))
	}

	// Kept for the selection, which is over what is on screen rather than over
	// a viewport this view does not have.
	m.schema.drawn = lines
	return strings.Join(lines, "\n")
}

// scrollbarCell is one row of a scrollbar column, or a space where there is no
// bar to draw.
func scrollbarCell(thumb []bool, i int, scrolls bool, thumbStyle, trackStyle lipgloss.Style) string {
	if !scrolls {
		return " "
	}
	if i < len(thumb) && thumb[i] {
		return thumbStyle.Render("█")
	}
	return trackStyle.Render("░")
}

// cut trims a line to the width it is drawn in, so a wide definition does not
// push the panes apart.
func cut(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width || width <= 0 {
		return line
	}
	return string(runes[:width])
}

// schemaRowAt returns the tree row a press covers.
func (m *MainModel) schemaRowAt(col, row int) (int, bool) {
	if m.viewMode != "schema" {
		return 0, false
	}

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	if col < 0 || col >= g.treeWidth {
		return 0, false
	}

	i := g.treeFirst + row - tabBarHeight - schemaHeaderRows
	if i < g.treeFirst || i >= g.treeFirst+g.height || i >= len(m.schema.rows()) {
		return 0, false
	}
	return i, true
}

// inSchemaDetail reports whether a press landed on the right-hand pane.
func (m *MainModel) inSchemaDetail(col, row int) bool {
	if m.viewMode != "schema" {
		return false
	}

	g := m.schemaGeometry(m.windowWidth, m.schemaHeight())
	return col >= g.treeWidth &&
		row >= tabBarHeight+schemaHeaderRows &&
		row < tabBarHeight+schemaHeaderRows+g.height
}

// schemaHeight is how many rows the view has, which is what every other view
// gets: the window without the tabs, the bars and the prompt.
func (m *MainModel) schemaHeight() int {
	return m.historyViewport.Height()
}
