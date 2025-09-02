package ui

import (
	"os"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Styles contains the styles for the application.
type Styles struct {
	Accent    lipgloss.Color
	Ok        lipgloss.Color
	Warn      lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color
	Border    lipgloss.Color

	AccentText   lipgloss.Style
	MutedText    lipgloss.Style
	ErrorText    lipgloss.Style
	SuccessText  lipgloss.Style
	WarnText     lipgloss.Style
}

// DefaultStyles returns the default styles for the application.
func DefaultStyles() *Styles {
	st := &Styles{}

	// Allow users to override color mode if needed
	// but don't force it by default - let lipgloss auto-detect
	colorMode := os.Getenv("CQLAI_COLOR_MODE")
	switch colorMode {
	case "ascii":
		lipgloss.SetColorProfile(termenv.Ascii)
	case "ansi":
		lipgloss.SetColorProfile(termenv.ANSI)
	case "256":
		lipgloss.SetColorProfile(termenv.ANSI256)
	case "truecolor":
		lipgloss.SetColorProfile(termenv.TrueColor)
	// default: let lipgloss auto-detect the best color mode
	}

	// Use hex colors for better consistency across terminals
	// These will be automatically adapted to the terminal's capabilities
	// Using brighter colors for better visibility in terminals with dark backgrounds
	st.Accent = lipgloss.Color("#5FAFFF") // Bright Sky Blue (brighter than before)
	st.Ok = lipgloss.Color("#5FFF5F")     // Bright Green (more visible than pure lime)
	st.Warn = lipgloss.Color("#FFFF5F")   // Bright Yellow
	st.Error = lipgloss.Color("#FF5F5F")   // Bright Red (softer than pure red)
	st.Muted = lipgloss.Color("#9E9E9E")   // Light Gray (brighter than 808080)
	st.Border = lipgloss.Color("#626262")  // Medium Gray (brighter than 444444)

	st.AccentText = lipgloss.NewStyle().Foreground(st.Accent)
	st.MutedText = lipgloss.NewStyle().Foreground(st.Muted)
	st.ErrorText = lipgloss.NewStyle().Foreground(st.Error)
	st.SuccessText = lipgloss.NewStyle().Foreground(st.Ok)
	st.WarnText = lipgloss.NewStyle().Foreground(st.Warn)

	return st
}
