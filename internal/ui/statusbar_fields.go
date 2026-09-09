package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// Clickable settings on the status line.
//
// The line is built from segments so that rendering and hit testing come from
// one description of the layout. Working the columns out separately from the
// drawing is how a click ends up opening the wrong thing the moment either side
// changes - the same trap the tab bar avoids by sharing layoutTabs.

// Setting keys, used to tie a click to what it changes.
const (
	settingKeyspace    = "KS"
	settingConsistency = "CL"
	settingPaging      = "Pg"
	settingTracing     = "Trace"
	settingAutoFetch   = "Fetch"
)

// statusSegment is one "Label: value" pair on the status line.
//
// A segment with an empty setting is information rather than a control - the
// server version, the user, the host - and ignores clicks.
type statusSegment struct {
	setting    string // one of the setting constants, or "" if not clickable
	label      string
	value      string
	start, end int // column range covering label and value, end exclusive
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

	username := m.Username
	if username == "" {
		username = "(anonymous)"
	}

	onOff := func(b bool) string {
		if b {
			return "ON"
		}
		return "OFF"
	}

	var segs []statusSegment
	if m.Version != "" {
		segs = append(segs, statusSegment{label: "v", value: m.Version})
	}
	segs = append(segs,
		statusSegment{label: "User: ", value: username},
		statusSegment{label: "Host: ", value: m.Host},
		statusSegment{setting: settingKeyspace, label: "KS: ", value: keyspace},
		statusSegment{setting: settingConsistency, label: "CL: ", value: m.Consistency},
		statusSegment{setting: settingPaging, label: "Pg: ", value: fmt.Sprintf("%d", m.PagingSize)},
		statusSegment{setting: settingTracing, label: "Trace: ", value: onOff(m.Tracing)},
		statusSegment{setting: settingAutoFetch, label: "Fetch: ", value: onOff(m.AutoFetch)},
	)

	// Columns, not bytes. The separator is three columns wide but five bytes,
	// because of the box-drawing character, and lipgloss.Width also gets
	// double-width characters right, which a keyspace name can contain.
	col := statusBarPadding
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

// settingAt returns the setting whose segment covers a column, and where that
// segment starts, so a chooser can be anchored to it.
func (m StatusBarModel) settingAt(col int) (setting string, at int, ok bool) {
	for _, seg := range m.segments() {
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
// them.
func settingChoices(setting string, current string) (choices []string, selected int) {
	switch setting {
	case settingConsistency:
		choices = []string{
			"ANY", "ONE", "TWO", "THREE", "QUORUM", "ALL",
			"LOCAL_QUORUM", "EACH_QUORUM", "LOCAL_ONE", "SERIAL", "LOCAL_SERIAL",
		}
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
