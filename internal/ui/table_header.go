package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/axonops/cqlai/internal/db"
)

// The block at the top of a boxed table: the border, the column names, what
// each column is, and the separator under it.
//
// The name and what the column is are on separate rows. Putting the type beside
// the name widened every column by the length of its type name, which is why it
// used to be hidden behind F6 and so was almost never seen. A row costs one row
// for the whole table however many columns there are.
//
//	┌──────────────┬─────────────┬───────────┐
//	│ id           │ created     │ name      │
//	│ PK int       │ C timestamp │ varchar   │
//	├──────────────┼─────────────┼───────────┤
//
// The table body and the frozen header that replaces it when you scroll both
// come through here. They were two copies of this layout before, and a header
// drawn one way in the body and another way frozen puts columns on screen that
// nothing below them lines up with.

// headerDetail is what a column is: its key marker, if it has one, and its
// type. "PK int", "C1 timestamp", or "varchar" for a column that is neither.
func headerDetail(header, columnType string) string {
	label := db.KeyLabel(header)
	switch {
	case label == "":
		return columnType
	case columnType == "":
		return label
	}
	return label + " " + columnType
}

// headerDetails is the detail row for a result, or nil when there is not one to
// draw.
//
// Every column of a Cassandra table has a type, and a query carries one per
// column - executor.go fills the headers and the types in the same loop, so
// they cannot come apart. What arrives without them is a grid the shell made up
// rather than a result over a table: DESCRIBE and the listings, whose "Primary
// Key" and "GC Grace" are not columns of anything. Those get no detail row, the
// same as they get no key markers.
func (m *MainModel) headerDetails(headers []string) []string {
	if len(m.columnTypes) != len(headers) {
		return nil
	}

	details := make([]string, len(headers))
	for i, header := range headers {
		details[i] = headerDetail(header, m.columnTypes[i])
	}
	return details
}

// headerWidths returns the width each column needs for its name and detail, so
// the widths are worked out from what will be drawn rather than from the name
// alone.
func (m *MainModel) headerWidths(headers []string) []int {
	widths := make([]int, len(headers))
	details := m.headerDetails(headers)

	for i, header := range headers {
		widths[i] = runeWidth(db.StripKeyMarker(stripAnsi(header)))
		if details != nil {
			widths[i] = max(widths[i], runeWidth(details[i]))
		}
	}
	return widths
}

// headerBlock is the border, the names, the details and the separator.
func (m *MainModel) headerBlock(headers []string, colWidths []int) []string {
	lines := []string{borderRow("┌", "┬", "┐", colWidths)}

	names := make([]string, len(headers))
	for i, header := range headers {
		// The marker is drawn on the detail row, so the name row shows the name
		// as Cassandra knows it.
		names[i] = db.StripKeyMarker(stripAnsi(header))
	}
	lines = append(lines, m.headerTextRow(names, colWidths, m.styles.AccentText.Bold(true)))

	if details := m.headerDetails(headers); details != nil {
		lines = append(lines, m.headerTextRow(details, colWidths, m.styles.MutedText))
	}

	lines = append(lines, borderRow("├", "┼", "┤", colWidths))
	return lines
}

// headerRowCount is how many rows headerBlock draws. The frozen header replaces
// exactly this many rows of the table, and the selection reads it to know which
// rows on screen are the header rather than the content scrolled under it.
func (m *MainModel) headerRowCount(headers []string) int {
	if m.headerDetails(headers) == nil {
		return 3
	}
	return 4
}

// headerTextRow draws one row of header cells.
func (m *MainModel) headerTextRow(cells []string, colWidths []int, style lipgloss.Style) string {
	row := "│"
	for i, cell := range cells {
		if i >= len(colWidths) {
			break
		}
		padding := max(colWidths[i]-runeWidth(cell), 0)
		row += " " + style.Render(cell) + ansiReset + strings.Repeat(" ", padding) + " │"
	}
	return row
}

// borderRow draws one of the horizontal rules of the box.
func borderRow(left, join, right string, colWidths []int) string {
	row := left
	for i, width := range colWidths {
		row += strings.Repeat("─", width+2)
		if i < len(colWidths)-1 {
			row += join
		}
	}
	return row + right
}

// runeWidth counts the columns a string takes, ANSI stripped.
func runeWidth(s string) int {
	return len([]rune(stripAnsi(s)))
}
