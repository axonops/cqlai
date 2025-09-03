package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AISelectionModal represents a modal for user to select from AI-provided options
type AISelectionModal struct {
	Active         bool
	Title          string   // e.g., "Select a keyspace"
	Message        string   // e.g., "Multiple keyspaces found. Please select one:"
	SelectionType  string   // The type being selected (e.g., "keyspace", "table")
	Options        []string // The list of options to choose from
	Selected       int      // Currently selected option index
	ScrollOffset   int      // Scroll offset for long lists
	Width          int
	Height         int
	AllowCancel    bool   // Show cancel button
	AllowCustom    bool   // Allow user to enter custom text
	CustomInput    string // Custom input text
	InputMode      bool   // Whether we're in custom input mode
	MaxVisibleOpts int    // Maximum visible options (set during render based on screen height)
}

// NewAISelectionModal creates a new selection modal
func NewAISelectionModal(selectionType string, options []string) *AISelectionModal {
	title := fmt.Sprintf("Select %s", selectionType)
	message := fmt.Sprintf("Multiple options found for %s. Please select one:", selectionType)

	return &AISelectionModal{
		Active:         true,
		Title:          title,
		Message:        message,
		SelectionType:  selectionType,
		Options:        options,
		Selected:       0,
		Width:          60,
		Height:         20,
		AllowCancel:    true,
		AllowCustom:    true,
		InputMode:      false,
		MaxVisibleOpts: 8, // Default value, will be recalculated during render
	}
}

// NextOption moves to the next option
func (m *AISelectionModal) NextOption() {
	if !m.InputMode && len(m.Options) > 0 {
		m.Selected = (m.Selected + 1) % len(m.Options)
		// Adjust scroll offset if needed
		if m.Selected >= m.ScrollOffset+m.MaxVisibleOpts {
			m.ScrollOffset = m.Selected - m.MaxVisibleOpts + 1
		} else if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
		}
	}
}

// PrevOption moves to the previous option
func (m *AISelectionModal) PrevOption() {
	if !m.InputMode && len(m.Options) > 0 {
		m.Selected--
		if m.Selected < 0 {
			m.Selected = len(m.Options) - 1
		}
		// Adjust scroll offset if needed
		if m.Selected >= m.ScrollOffset+m.MaxVisibleOpts {
			m.ScrollOffset = m.Selected - m.MaxVisibleOpts + 1
		} else if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
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

	// Calculate available space for the modal
	// Reserve space for: title(2) + message(2) + custom input(4) + buttons(3) + borders/padding(6)
	reservedHeight := 17
	availableHeight := screenHeight - reservedHeight

	// Calculate maximum visible options based on available space
	maxVisibleOptions := availableHeight / 2 // Each option takes roughly 1 line, but we need margin
	if maxVisibleOptions > 8 {
		maxVisibleOptions = 8 // Cap at 8 for better UX
	}
	if maxVisibleOptions < 3 {
		maxVisibleOptions = 3 // Minimum of 3 visible options
	}

	// Update the MaxVisibleOpts field for use in scrolling
	m.MaxVisibleOpts = maxVisibleOptions

	numOptions := len(m.Options)
	visibleOptions := numOptions
	if visibleOptions > maxVisibleOptions {
		visibleOptions = maxVisibleOptions
	}

	// Calculate actual options container height
	optionsHeight := visibleOptions + 2 // +2 for padding

	// Create modal box style with reduced padding
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1).
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
		Align(lipgloss.Center)

	// Options list style
	optionStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Width(m.Width - 8)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(styles.Accent).
		Bold(true).
		Padding(0, 2).
		Width(m.Width - 8)

	// Build options list with scrolling
	var optionsList []string
	startIdx := m.ScrollOffset
	endIdx := startIdx + visibleOptions
	if endIdx > len(m.Options) {
		endIdx = len(m.Options)
	}

	// Add scroll indicator at top if needed
	if startIdx > 0 {
		scrollIndicator := optionStyle.Render("  ↑ (more above)")
		optionsList = append(optionsList, scrollIndicator)
	}

	for i := startIdx; i < endIdx; i++ {
		var optText string
		if i == m.Selected && !m.InputMode {
			// Selected option
			optText = selectedStyle.Render(fmt.Sprintf("▶ %s", m.Options[i]))
		} else {
			// Normal option
			optText = optionStyle.Render(fmt.Sprintf("  %s", m.Options[i]))
		}
		optionsList = append(optionsList, optText)
	}

	// Add scroll indicator at bottom if needed
	if endIdx < len(m.Options) {
		scrollIndicator := optionStyle.Render("  ↓ (more below)")
		optionsList = append(optionsList, scrollIndicator)
	}

	// Options container - use calculated height with reduced padding
	optionsContainer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Border).
		Padding(0, 1).
		Width(m.Width - 6).
		Height(optionsHeight).
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

	// Always position with a margin from top to avoid cutoff
	// Add top padding to ensure modal doesn't get cut off
	topPadding := 2
	paddedModal := lipgloss.NewStyle().
		MarginTop(topPadding).
		Render(modalBox)

	// Position the modal with top alignment to prevent cutoff
	return lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Top,
		paddedModal,
	)
}
