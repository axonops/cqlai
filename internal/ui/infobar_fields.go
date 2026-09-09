package ui

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// The query info bar, described once so that drawing it and clicking it agree.
//
// Same arrangement as the status line below it: segments carry the columns they
// occupy, and both the renderer and the hit test walk the same list. Working the
// columns out separately is how a click lands on the wrong field.

// infoHistory is the only clickable field on this line.
const infoHistory = "History"

// infoSegment is one "Label: value" pair. A segment with no field is a fact
// about the last query and ignores clicks.
type infoSegment struct {
	field      string
	label      string
	value      string
	start, end int // column range covering label and value, end exclusive
}

// infoBarPadding is the left padding lipgloss adds when rendering the line.
const infoBarPadding = 1

// maxLastCommand is how much of the last command is shown before it is cut.
const maxLastCommand = 50

// infoPlaceholder stands in for a value there is not one of yet.
const infoPlaceholder = "-"

// segments describes the info bar in order.
//
// The line used to open with a separator that separated nothing: AutoFetch sat
// to the left of it until it moved to the connection bar, and its separator
// stayed behind. Built from a list, that cannot happen - the separator goes
// between segments rather than before one.
//
// All three fields are always on the line, whether or not anything has run.
// They used to appear only once there was something to put in them, so the bar
// arrived from nowhere on the first query and the fields shifted sideways as
// each one filled - and there was nothing to click until you had already run a
// command.
func (m TopBarModel) segments() []infoSegment {
	command := infoPlaceholder
	if m.LastCommand != "" {
		command = m.LastCommand
		if len(command) > maxLastCommand {
			command = command[:maxLastCommand] + "..."
		}
	}

	// A dash rather than a stale figure: HasQueryData goes false for commands
	// that return no rows, so carrying the last query's timing over would put
	// numbers next to a command that never produced them.
	queryTime, rows := infoPlaceholder, infoPlaceholder
	if m.HasQueryData {
		queryTime = fmt.Sprintf("%v", m.QueryTime.Round(time.Millisecond))

		rows = fmt.Sprintf("%d", m.RowCount)
		if m.HasMoreData && !m.AutoFetch {
			rows += "+"
		}
	}

	// "History" rather than "Last" because clicking it opens the history. The
	// value is still the last command run, which is the one thing worth having
	// on screen without opening anything.
	segs := []infoSegment{
		{field: infoHistory, label: "History: ", value: command},
		{label: "Query: ", value: queryTime},
		{label: "Rows: ", value: rows},
	}

	// Columns, not bytes, for the same reason as the status line: the
	// separator is three columns wide and five bytes.
	col := infoBarPadding
	for i := range segs {
		if i > 0 {
			col += lipgloss.Width(statusSeparator)
		}
		segs[i].start = col
		col += lipgloss.Width(segs[i].label) + lipgloss.Width(segs[i].value)
		segs[i].end = col
	}
	return segs
}

// fieldAt returns the clickable field covering a column.
func (m TopBarModel) fieldAt(col int) (string, bool) {
	for _, seg := range m.segments() {
		if col >= seg.start && col < seg.end {
			return seg.field, seg.field != ""
		}
	}
	return "", false
}
