package ui

import (
	"fmt"
	"slices"

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
	short      string // the label used when the full set will not fit
	value      string
	start, end int  // column range covering label and value, end exclusive
	keep       bool // never dropped, however narrow the line
}

// width is the columns this segment takes as it will be drawn.
func (s statusSegment) width() int {
	return lipgloss.Width(s.label) + lipgloss.Width(s.value)
}

// shorten swaps in the abbreviated label, where there is one.
func (s statusSegment) shorten() statusSegment {
	if s.short != "" {
		s.label = s.short
	}
	return s
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
	// In order of what you would rather keep when the line will not hold all of
	// it: the keyspace and the consistency level decide what a query does and
	// where it goes, and the rest are settings you can look up. placeSegments
	// drops from the right of this list.
	segs := []statusSegment{
		{setting: settingConnection, label: "Connection", short: "Conn", value: ""},
		{setting: settingKeyspace, label: "KS: ", value: keyspace},
		{setting: settingConsistency, label: "CL: ", value: m.Consistency},
		{setting: settingOutput, label: "Output: ", short: "Out: ", value: m.OutputFormat},
		{setting: settingPaging, label: "Pg: ", value: fmt.Sprintf("%d", m.PagingSize)},
		{setting: settingTracing, label: "Trace: ", short: "Tr: ", value: onOff(m.Tracing)},
		{setting: settingAutoFetch, label: "Fetch: ", short: "F: ", value: onOff(m.AutoFetch)},
	}

	// Capture was against the right-hand edge, apart from the settings, because
	// it was put there as a control - a thing you press. It is both, and the
	// fact is the more important half: it is the one setting that keeps doing
	// something after you have stopped thinking about it, and quietly writing
	// to a file you have forgotten is exactly what wants to be on screen. As a
	// fact it belongs with the other facts.
	//
	// A capture that is running is never dropped, however narrow the line, for
	// that same reason. One that is off takes its turn with the rest:
	// "Capture: OFF" says nothing you had not already assumed.
	segs = append(segs, statusSegment{
		setting: settingCapture,
		label:   "Capture: ",
		short:   "Cap: ",
		value:   onOff(m.Capturing),
		keep:    m.Capturing,
	})

	return segs
}

// layOutFlow gives the left-hand segments their columns, in order.
//
// Columns, not bytes. The separator is three columns wide but five bytes,
// because of the box-drawing character, and lipgloss.Width also gets
// double-width characters right, which a keyspace name can contain.
func layOutFlow(segs []statusSegment) []statusSegment {
	col := statusBarPadding
	for i := range segs {
		if i > 0 {
			col += lipgloss.Width(statusSeparator)
		}
		segs[i].start = col
		col += segs[i].width()
		segs[i].end = col
	}
	return segs
}

// flowEnd is the column the laid-out flow reaches.
func flowEnd(segs []statusSegment) int {
	end := statusBarPadding
	for _, seg := range segs {
		end = max(end, seg.end)
	}
	return end
}

// placeSegments fits the line to the terminal and gives every segment its
// columns.
//
// The bar used to lay itself out at whatever width it wanted and be rendered
// through lipgloss Width, which wraps rather than truncates: below about a
// hundred columns the fields ran past the end and the bar became two rows, so
// the whole view was a row taller than the terminal. It is one row at every
// width now, and what does not fit is dropped rather than wrapped.
//
// What gives, in order:
//
//  1. the labels shorten - Connection to Conn, Output to Out, Trace to Tr
//  2. fields drop from the right of the flow, so the keyspace and consistency
//     level are the last to go: they decide what a query does and where it
//     goes, and the rest are settings you can look up
//
// Capture keeps its place through both. It is a thing you do rather than a fact
// about the session, and a control that moves or vanishes as the terminal is
// resized is worse than a fact you have to go and look up. Its own label
// shortens with the rest.
//
// Rendering and hit testing both call this, so a click lands on the field drawn
// there however wide the terminal is and whatever has been dropped to fit.
func placeSegments(segs []statusSegment, width int) []statusSegment {
	room := width - statusBarPadding

	if fitted, ok := fitFlow(segs, room); ok {
		return fitted
	}

	shortened := make([]statusSegment, len(segs))
	for i, seg := range segs {
		shortened[i] = seg.shorten()
	}

	// Drop from the right until what is left fits, passing over anything
	// marked keep - a capture that is running, which is the whole reason the
	// field is on this line. The first field always stays: a bar with nothing
	// on it says less than a crowded one.
	for {
		if fitted, ok := fitFlow(shortened, room); ok {
			return fitted
		}

		last := -1
		for i := len(shortened) - 1; i > 0; i-- {
			if !shortened[i].keep {
				last = i
				break
			}
		}
		if last < 0 {
			// Only the first field and whatever must stay are left, and they
			// still overrun. Drawing them cut off says more than drawing none.
			return layOutFlow(shortened)
		}
		shortened = append(shortened[:last:last], shortened[last+1:]...)
	}
}

// fitFlow lays the segments out and reports whether they stay inside room.
func fitFlow(segs []statusSegment, room int) ([]statusSegment, bool) {
	laid := layOutFlow(slices.Clone(segs))
	return laid, flowEnd(laid) <= room
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
