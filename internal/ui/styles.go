package ui

import (
	"os"

	"image/color"

	"charm.land/lipgloss/v2"
)

// Styles contains the styles for the application.
type Styles struct {
	Accent color.Color
	Ok     color.Color
	Warn   color.Color
	Error  color.Color
	Muted  color.Color
	Border color.Color

	AccentText  lipgloss.Style
	MutedText   lipgloss.Style
	ErrorText   lipgloss.Style
	SuccessText lipgloss.Style
	WarnText    lipgloss.Style
}

// DefaultStyles returns the default styles for the application.
func DefaultStyles() *Styles {
	st := &Styles{}

	// CQLAI_COLOR_MODE used to call lipgloss.SetColorProfile, which v2 removed:
	// colour handling moved into the renderer, which negotiates with the
	// terminal itself. Honour the variable by mapping it onto the environment
	// lipgloss and termenv already read, rather than reaching into the library.
	switch os.Getenv("CQLAI_COLOR_MODE") {
	case "ascii":
		_ = os.Setenv("NO_COLOR", "1")
	case "ansi":
		_ = os.Setenv("TERM", "xterm")
	case "256":
		_ = os.Setenv("TERM", "xterm-256color")
	case "truecolor":
		_ = os.Setenv("COLORTERM", "truecolor")
		// default: let the renderer detect what the terminal supports
	}

	// Use hex colors for better consistency across terminals
	// These will be automatically adapted to the terminal's capabilities
	// Using brighter colors for better visibility in terminals with dark backgrounds
	st.Accent = lipgloss.Color("#5FAFFF") // Bright Sky Blue (brighter than before)
	st.Ok = lipgloss.Color("#5FFF5F")     // Bright Green (more visible than pure lime)
	st.Warn = lipgloss.Color("#FFFF5F")   // Bright Yellow
	st.Error = lipgloss.Color("#FF5F5F")  // Bright Red (softer than pure red)
	st.Muted = lipgloss.Color("#9E9E9E")  // Light Gray (brighter than 808080)
	st.Border = lipgloss.Color("#626262") // Medium Gray (brighter than 444444)

	st.AccentText = lipgloss.NewStyle().Foreground(st.Accent)
	st.MutedText = lipgloss.NewStyle().Foreground(st.Muted)
	st.ErrorText = lipgloss.NewStyle().Foreground(st.Error)
	st.SuccessText = lipgloss.NewStyle().Foreground(st.Ok)
	st.WarnText = lipgloss.NewStyle().Foreground(st.Warn)

	return st
}

// dimmedColour is what something that is there but has nothing to offer yet is
// drawn in: a tab with no results behind it, a menu entry that cannot be picked.
//
// Dark enough to read as unavailable beside the ones that are, and light enough
// to read at all. At #585858 a dimmed entry looked like one that was not there,
// which is the opposite of what dimming rather than hiding is for.
var dimmedColour = lipgloss.Color("#7A7A7A")
