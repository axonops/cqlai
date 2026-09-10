package ui

import (
	"time"

	"charm.land/lipgloss/v2"
)

// TopBarModel is the Bubble Tea model for the top status bar.
type TopBarModel struct {
	LastCommand  string
	QueryTime    time.Duration
	RowCount     int
	HasQueryData bool
	// Not displayed here any more - the status bar shows it with the other
	// session settings - but still needed to decide whether the row count is
	// shown as "100+" with more to fetch.
	AutoFetch   bool
	HasMoreData bool // Indicates if there's more data to fetch

	// RowsDropped is set when the sliding window has thrown away rows from the
	// start of the result to stay inside its memory limit, and FirstRow is the
	// first row still held, counting from one.
	//
	// This is the only place the program admits it has dropped anything. It was
	// computed every frame and appended to the status line past the right-hand
	// edge, where it was never once visible (#135), so scrolling up to the start
	// of a result and not finding it there looked like lost data rather than a
	// decision.
	RowsDropped bool
	FirstRow    int64
}

// NewTopBarModel creates a new TopBarModel.
func NewTopBarModel() TopBarModel {
	return TopBarModel{}
}

// View renders the query info bar.
//
// It draws the segment list, which is also what fieldAt tests clicks against,
// so the History field cannot be drawn in one place and clicked in another.
func (m TopBarModel) View(width int, styles *Styles, viewMode string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D7FF")).
		Bold(true)

	queryTimeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FF00"))

	rowCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))

	droppedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	// The mode used to be named here. The tab line above shows it now, and
	// highlights it, so repeating it would only take up room.
	styleFor := func(seg infoSegment) lipgloss.Style {
		switch {
		case seg.field == infoHistory:
			return commandStyle
		case seg.label == infoRowsLabel:
			return rowCountStyle
		case seg.label == infoDroppedLabel:
			return droppedStyle
		default:
			return queryTimeStyle
		}
	}

	content := ""
	for i, seg := range placeInfoSegments(m.segments(), width) {
		if i > 0 {
			content += separatorStyle.Render(statusSeparator)
		}
		content += labelStyle.Render(seg.label) + styleFor(seg).Render(seg.value)
	}

	// Apply style to the entire bar without forced background
	// One row, always - see placeInfoSegments. MaxHeight makes sure a mistake
	// in the fitting costs a field rather than the layout.
	barStyle := lipgloss.NewStyle().
		MaxHeight(1).
		Padding(0, infoBarPadding).
		Width(width)

	return barStyle.Render(content)
}
