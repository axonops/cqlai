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
	st.Accent = lipgloss.Color("#00BFFF") // DeepSkyBlue
	st.Ok = lipgloss.Color("#00FF00")     // Lime
	st.Warn = lipgloss.Color("#FFFF00")   // Yellow
	st.Error = lipgloss.Color("#FF0000")   // Red
	st.Muted = lipgloss.Color("#808080")   // Gray
	st.Border = lipgloss.Color("#444444")  // Dark Gray

	st.AccentText = lipgloss.NewStyle().Foreground(st.Accent)
	st.MutedText = lipgloss.NewStyle().Foreground(st.Muted)
	st.ErrorText = lipgloss.NewStyle().Foreground(st.Error)
	st.SuccessText = lipgloss.NewStyle().Foreground(st.Ok)
	st.WarnText = lipgloss.NewStyle().Foreground(st.Warn)

	return st
}
