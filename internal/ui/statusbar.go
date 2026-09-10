package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var ()

// StatusBarModel is the Bubble Tea model for the status bar.
type StatusBarModel struct {
	Username     string
	Host         string
	Latency      string
	Consistency  string
	PagingSize   int
	Tracing      bool
	AutoFetch    bool
	HasTraceData bool // Whether trace data is available to view
	Keyspace     string
	Version      string
	OutputFormat string
	Capturing    bool
}

// NewStatusBarModel creates a new StatusBarModel.
func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{
		Username:     "cassandra",
		Host:         "127.0.0.1",
		Latency:      "10ms",
		Consistency:  "LOCAL_ONE",
		PagingSize:   100,
		Tracing:      false,
		AutoFetch:    false,
		OutputFormat: "TABLE",
	}
}

// View renders the status bar.
func (m StatusBarModel) View(width int, styles *Styles, currentView string) string {
	// Define component styles
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	keyspaceStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF87FF")).
		Bold(true)

	hostStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87D7FF"))

	consistencyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD787"))

	pageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FFD7"))

	outputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87D7FF"))

	tracingOnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F")).
		Bold(true)

	tracingOffStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F5F5F"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	// Render from the same segment list that hit testing uses, so a click can
	// never land on a different field from the one drawn there.
	styleFor := func(seg statusSegment) lipgloss.Style {
		switch seg.setting {
		case settingKeyspace:
			return keyspaceStyle
		case settingConsistency:
			return consistencyStyle
		case settingOutput:
			return outputStyle
		case settingPaging:
			return pageStyle
		case settingTracing, settingAutoFetch:
			if seg.value == "ON" {
				return tracingOnStyle
			}
			return tracingOffStyle
		}
		return hostStyle
	}

	segs := placeSegments(m.segments(), width)

	statusText := ""
	col := statusBarPadding
	for i, seg := range segs {
		switch {
		case seg.right:
			statusText += strings.Repeat(" ", max(seg.start-col, 0))
		case i > 0:
			statusText += separatorStyle.Render(statusSeparator)
		}
		statusText += labelStyle.Render(seg.label) + styleFor(seg).Render(seg.value)
		col = seg.end
	}

	// One row, always. Width on its own wraps rather than truncates, so a line
	// that overran turned the bar into two rows and made the whole view a row
	// taller than the terminal. placeSegments fits the fields to the width;
	// MaxHeight makes sure that a mistake there costs a field rather than the
	// layout.
	barStyle := lipgloss.NewStyle().
		Padding(0, statusBarPadding).
		Width(width).
		MaxHeight(1)

	return barStyle.Render(statusText)
}
