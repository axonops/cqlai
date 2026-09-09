package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
)

// Clickable settings on the status line.
//
// The line is built from segments so that rendering and hit testing come from
// one description of the layout. Working the columns out separately from the
// drawing is how a click ends up opening the wrong thing the moment either side
// changes - the same trap the tab bar avoids by sharing layoutTabs.

// Setting keys, used to tie a click to what it changes.
const (
	settingConnection  = "Connection"
	settingKeyspace    = "KS"
	settingConsistency = "CL"
	settingOutput      = "Output"
	settingPaging      = "Pg"
	settingTracing     = "Trace"
	settingAutoFetch   = "Fetch"
	settingCapture     = "Capture"
)

// statusSegment is one "Label: value" pair on the status line.
//
// A segment with an empty setting is information rather than a control and
// ignores clicks.
type statusSegment struct {
	setting    string // one of the setting constants, or "" if not clickable
	label      string
	value      string
	start, end int  // column range covering label and value, end exclusive
	right      bool // placed against the right-hand edge rather than in the flow
}

// clickable reports whether this segment changes something.
func (s statusSegment) clickable() bool { return s.setting != "" }

// statusBarPadding is the left padding lipgloss adds when rendering the line.
// Click columns are screen columns, so they have to account for it.
const statusBarPadding = 1

// statusSeparator sits between segments.
const statusSeparator = " │ "

// segments describes the status line in order, with the column each piece
// occupies. Rendering walks the same list, so the two cannot disagree.
func (m StatusBarModel) segments() []statusSegment {
	keyspace := m.Keyspace
	if keyspace == "" {
		keyspace = "(none)"
	}

	onOff := func(b bool) string {
		if b {
			return "ON"
		}
		return "OFF"
	}

	// The version, the user and the host were three fields taking a third of
	// the line to say things that rarely change. They are one field now, and
	// clicking it opens the lot, encryption included.
	segs := []statusSegment{
		{setting: settingConnection, label: "Connection", value: ""},
		{setting: settingKeyspace, label: "KS: ", value: keyspace},
		{setting: settingConsistency, label: "CL: ", value: m.Consistency},
		{setting: settingOutput, label: "Output: ", value: m.OutputFormat},
		{setting: settingPaging, label: "Pg: ", value: fmt.Sprintf("%d", m.PagingSize)},
		{setting: settingTracing, label: "Trace: ", value: onOff(m.Tracing)},
		{setting: settingAutoFetch, label: "Fetch: ", value: onOff(m.AutoFetch)},
	}

	// Capture sits against the right-hand edge rather than in the flow, like
	// the Help button on the tab line: it is a thing you do, not a fact about
	// the session, and the line is already busy on the left.
	segs = append(segs, statusSegment{
		setting: settingCapture,
		label:   "Capture: ",
		value:   onOff(m.Capturing),
		right:   true,
	})

	// Columns, not bytes. The separator is three columns wide but five bytes,
	// because of the box-drawing character, and lipgloss.Width also gets
	// double-width characters right, which a keyspace name can contain.
	col := statusBarPadding
	for i := range segs {
		if segs[i].right {
			continue
		}
		if i > 0 {
			col += lipgloss.Width(statusSeparator)
		}
		segs[i].start = col
		col += lipgloss.Width(segs[i].label) + lipgloss.Width(segs[i].value)
		segs[i].end = col
	}
	return segs
}

// place gives the right-anchored segments their columns, once the width of the
// line is known.
//
// Rendering and hit testing both call this, so a click on Capture lands on
// Capture however wide the terminal is. It is dropped rather than overlapped
// when the line is too full, for the same reason the Help button is.
func placeSegments(segs []statusSegment, width int) []statusSegment {
	end := statusBarPadding
	for _, seg := range segs {
		if !seg.right {
			end = max(end, seg.end)
		}
	}

	placed := make([]statusSegment, 0, len(segs))
	for _, seg := range segs {
		if !seg.right {
			placed = append(placed, seg)
			continue
		}

		w := lipgloss.Width(seg.label) + lipgloss.Width(seg.value)
		seg.start = width - w - statusBarPadding
		seg.end = seg.start + w
		if seg.start <= end+lipgloss.Width(statusSeparator) {
			continue // no room; leave it off rather than on top of a field
		}
		placed = append(placed, seg)
	}
	return placed
}

// settingAt returns the setting whose segment covers a column, and where that
// segment starts, so a chooser can be anchored to it.
func (m StatusBarModel) settingAt(width, col int) (setting string, at int, ok bool) {
	for _, seg := range placeSegments(m.segments(), width) {
		if col >= seg.start && col < seg.end {
			if !seg.clickable() {
				return "", 0, false
			}
			return seg.setting, seg.start, true
		}
	}
	return "", 0, false
}

// settingChoices lists the values a setting can take, and which is current.
//
// Keyspaces are not here: they come from the cluster, so the caller supplies
// them. Neither is Connection, which is not a setting at all - it opens a panel
// of facts rather than a list of values.
func settingChoices(setting string, current string) (choices []string, selected int) {
	switch setting {
	case settingConsistency:
		// From the same table CONSISTENCY applies, so the list cannot offer a
		// level that then fails to set.
		choices = db.ConsistencyLevels()
	case settingOutput:
		// From config, which is where ParseOutputFormat reads them, so the list
		// cannot offer a format OUTPUT then refuses.
		choices = config.OutputFormats()
	case settingPaging:
		// The values PAGING accepts, plus OFF, which is how it is turned off.
		choices = []string{"OFF", "50", "100", "500", "1000", "5000"}
	case settingTracing, settingAutoFetch:
		choices = []string{"ON", "OFF"}
	default:
		return nil, 0
	}

	for i, c := range choices {
		if c == current {
			return choices, i
		}
	}
	return choices, 0
}
