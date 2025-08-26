package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AISelectionModal represents a modal for user to select from AI-provided options
type AISelectionModal struct {
	Active       bool
	Title        string   // e.g., "Select a keyspace"
	Message      string   // e.g., "Multiple keyspaces found. Please select one:"
	Options      []string // The list of options to choose from
	Selected     int      // Currently selected option index
	Width        int
	Height       int
	AllowCancel  bool     // Show cancel button
	AllowCustom  bool     // Allow user to enter custom text
	CustomInput  string   // Custom input text
	InputMode    bool     // Whether we're in custom input mode
}

// NewAISelectionModal creates a new selection modal
func NewAISelectionModal(selectionType string, options []string) *AISelectionModal {
	title := fmt.Sprintf("Select %s", selectionType)
	message := fmt.Sprintf("Multiple options found for %s. Please select one:", selectionType)
	
	return &AISelectionModal{
		Active:      true,
		Title:       title,
		Message:     message,
		Options:     options,
		Selected:    0,
		Width:       60,
		Height:      20,
		AllowCancel: true,
		AllowCustom: true,
		InputMode:   false,
	}
}

// NextOption moves to the next option
func (m *AISelectionModal) NextOption() {
	if !m.InputMode && len(m.Options) > 0 {
		m.Selected = (m.Selected + 1) % len(m.Options)
	}
}

// PrevOption moves to the previous option
func (m *AISelectionModal) PrevOption() {
	if !m.InputMode && len(m.Options) > 0 {
		m.Selected--
		if m.Selected < 0 {
			m.Selected = len(m.Options) - 1
		}
	}
}

// ToggleInputMode toggles between selection and custom input mode
func (m *AISelectionModal) ToggleInputMode() {
	if m.AllowCustom {
		m.InputMode = !m.InputMode
	}
}

// GetSelection returns the selected value or custom input
func (m *AISelectionModal) GetSelection() string {
	if m.InputMode && m.CustomInput != "" {
		return m.CustomInput
	}
	if m.Selected >= 0 && m.Selected < len(m.Options) {
		return m.Options[m.Selected]
	}
	return ""
}

// Render renders the selection modal
func (m *AISelectionModal) Render(screenWidth, screenHeight int, styles *Styles) string {
	if !m.Active {
		return ""
	}

	// Create modal box style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		BorderBackground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#000000")).
		Padding(1, 2).
		Width(m.Width).
		MaxWidth(screenWidth - 4)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 4)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Width(m.Width - 4).
		Align(lipgloss.Center).
		MarginTop(1)

	// Options list style
	optionStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Width(m.Width - 8)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(styles.Accent).
		Bold(true).
		Padding(0, 2).
		Width(m.Width - 8)

	// Build options list
	var optionsList []string
	for i, opt := range m.Options {
		var optText string
		if i == m.Selected && !m.InputMode {
			// Selected option
			optText = selectedStyle.Render(fmt.Sprintf("▶ %s", opt))
		} else {
			// Normal option
			optText = optionStyle.Render(fmt.Sprintf("  %s", opt))
		}
		optionsList = append(optionsList, optText)
	}

	// Options container
	optionsContainer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Padding(1).
		Width(m.Width - 6).
		Height(10).
		MarginTop(1).
		Render(strings.Join(optionsList, "\n"))

	// Custom input field (if enabled)
	customInputSection := ""
	if m.AllowCustom {
		inputLabel := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			MarginTop(1).
			Render("Or enter custom value:")

		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border).
			Padding(0, 1).
			Width(m.Width - 8).
			MarginTop(1)

		if m.InputMode {
			// Active input mode
			inputStyle = inputStyle.BorderForeground(styles.Accent)
			input := m.CustomInput
			if input == "" {
				input = "Type your custom value..."
			}
			customInputSection = lipgloss.JoinVertical(
				lipgloss.Left,
				inputLabel,
				inputStyle.Render(input),
			)
		} else {
			// Inactive input mode
			customInputSection = lipgloss.JoinVertical(
				lipgloss.Left,
				inputLabel,
				inputStyle.Render("Press 'i' to enter custom input"),
			)
		}
	}

	// Buttons
	buttonStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.Width - 4).
		MarginTop(1)

	var buttons string
	if m.InputMode {
		buttons = buttonStyle.Render("Enter: Confirm  •  Esc: Cancel Input Mode")
	} else {
		buttons = buttonStyle.Render("↑↓: Navigate  •  Enter: Select  •  i: Custom Input  •  Esc: Cancel")
	}

	// Build the complete modal
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(m.Title),
		messageStyle.Render(m.Message),
		optionsContainer,
		customInputSection,
		"",
		buttons,
	)

	modalBox := modalStyle.Render(content)

	// Center the modal on screen
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#000000")),
	)
}