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

	g.detailFirst = min(max(m.schema.detailScroll, 0), max(len(m.schema.detail)-g.height, 0))
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
	detailBar := scrollbarColumn(min(g.height, len(m.schema.detail)), g.detailFirst, len(m.schema.detail))

	headingStyle := lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true)

	lines := make([]string, 0, g.height+schemaHeaderRows)
	lines = append(lines,
		headingStyle.Render(pad(schemaTreeHeading, g.treeWidth))+
			ruleStyle.Render(schemaDivider)+
			headingStyle.Render(pad(m.schemaDetailHeading(), g.detailWidth)),
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

		// The right-hand pane.
		definition := ""
		if g.detailFirst+i < len(m.schema.detail) {
			definition = m.schema.detail[g.detailFirst+i]
		}
		right := treeStyle.Render(pad(cut(definition, g.detailWidth-1), g.detailWidth-1))
		right += scrollbarCell(detailBar, i, len(m.schema.detail) > g.height, thumbStyle, trackStyle)

		lines = append(lines, left+ruleStyle.Render(schemaDivider)+right)
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
