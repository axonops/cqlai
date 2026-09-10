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

	return segs
}

// placeInfoSegments fits the line to the terminal and gives each segment its
// columns.
//
// The same defect as the status line had, one row up: laid out at whatever
// width it wanted and rendered through lipgloss Width, which wraps rather than
// truncates, so on a narrow terminal the bar became two rows and the whole view
// a row taller than the screen.
//
// What gives here is the last command, because it is the only field that can be
// any length - it is cut to what is left rather than dropped, since a shortened
// command still tells you which one it was. The timing and the row count are a
// handful of columns and stay. If even that will not fit, fields drop from the
// right, and Query goes before Rows.
//
// Columns, not bytes, for the same reason as the status line: the separator is
// three columns wide and five bytes.
func placeInfoSegments(segs []infoSegment, width int) []infoSegment {
	room := width - infoBarPadding*2

	for {
		if len(segs) == 0 {
			return segs
		}

		fixed := 0
		for _, seg := range segs[1:] {
			fixed += lipgloss.Width(statusSeparator) + lipgloss.Width(seg.label) + lipgloss.Width(seg.value)
		}

		// What the first segment has left for its value.
		spare := room - fixed - infoBarPadding - lipgloss.Width(segs[0].label)
		if spare >= 1 {
			segs[0].value = truncateToWidth(segs[0].value, spare)
			break
		}
		segs = segs[:len(segs)-1]
	}

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

// truncateToWidth cuts a value to fit, with an ellipsis where anything was
// taken off, so a shortened command still reads as one that was cut.
func truncateToWidth(value string, room int) string {
	if lipgloss.Width(value) <= room {
		return value
	}
	if room <= 1 {
		return "…"
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > room {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

// placedSegments is the info bar as it is drawn at a width.
func (m TopBarModel) placedSegments(width int) []infoSegment {
	return placeInfoSegments(m.segments(), width)
}

// fieldAt returns the clickable field covering a column.
//
// It takes the width for the same reason placeSegments does: what is on the
// line depends on how much of it there is, and a click has to land on the field
// actually drawn there.
func (m TopBarModel) fieldAt(width, col int) (string, bool) {
	for _, seg := range placeInfoSegments(m.segments(), width) {
		if col >= seg.start && col < seg.end {
			return seg.field, seg.field != ""
		}
	}
	return "", false
}
