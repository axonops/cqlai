package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
	HasTraceData bool // Whether trace data is available to view
	Keyspace     string
	Version      string
	OutputFormat string
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

	tracingOnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F")).
		Bold(true)

	tracingOffStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F5F5F"))

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	// Format values
	keyspaceDisplay := m.Keyspace
	if keyspaceDisplay == "" {
		keyspaceDisplay = "(none)"
	}

	usernameDisplay := m.Username
	if usernameDisplay == "" {
		usernameDisplay = "(anonymous)"
	}

	tracingState := "OFF"
	tracingStyle := tracingOffStyle
	if m.Tracing {
		tracingState = "ON"
		tracingStyle = tracingOnStyle
	}

	// Version style
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B8B8B8"))

	// Build the status text with colors in the requested order:
	// Cassandra version, User, Host, KS, Cons, Pg, Trace
	statusText := ""

	// Start with version if available
	if m.Version != "" {
		statusText = labelStyle.Render("v") + versionStyle.Render(m.Version) +
			separatorStyle.Render(" │ ")
	}

	// Then add the rest in order
	statusText += labelStyle.Render("User: ") + hostStyle.Render(usernameDisplay) +
		separatorStyle.Render(" │ ") +
		labelStyle.Render("Host: ") + hostStyle.Render(m.Host) +
		separatorStyle.Render(" │ ") +
		labelStyle.Render("KS: ") + keyspaceStyle.Render(keyspaceDisplay) +
		separatorStyle.Render(" │ ") +
		labelStyle.Render("Cons: ") + consistencyStyle.Render(m.Consistency) +
		separatorStyle.Render(" │ ") +
		labelStyle.Render("Pg: ") + pageStyle.Render(fmt.Sprintf("%d", m.PagingSize)) +
		separatorStyle.Render(" │ ") +
		labelStyle.Render("Trace: ") + tracingStyle.Render(tracingState)


	// Apply style to the entire bar without forced background
	barStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(width)

	return barStyle.Render(statusText)
}
