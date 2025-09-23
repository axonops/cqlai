package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/router"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderSaveModal renders the save modal dialog
func (m *MainModel) renderSaveModal(screenWidth, screenHeight int) string {
	if !m.saveModalActive {
		return ""
	}

	// Modal content
	var content string
	switch m.saveModalStep {
	case 0:
		content = m.renderSaveFormatSelection()
	case 1:
		content = m.renderSaveFilenameInput()
	}

	// Calculate the actual width needed for content
	contentLines := strings.Split(content, "\n")
	maxLineLength := 0
	for _, line := range contentLines {
		// Strip ANSI codes to get actual length
		cleanLine := stripAnsi(line)
		if len(cleanLine) > maxLineLength {
			maxLineLength = len(cleanLine)
		}
	}

	// Add padding for the border and internal padding (2 chars on each side for padding + 2 for border)
	modalWidth := maxLineLength + 6
	if modalWidth < 50 {
		modalWidth = 50 // Minimum width
	}
	if modalWidth > 70 {
		modalWidth = 70 // Maximum width
	}

	// Create a box style for the modal with explicit border and background
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.AccentText.GetForeground()).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(1, 2).
		Width(modalWidth)

	// Apply the style and render the modal
	return modalStyle.Render(content)
}

// renderSaveFormatSelection renders the format selection step
func (m *MainModel) renderSaveFormatSelection() string {
	formats := []struct {
		name string
		desc string
		ext  string
	}{
		{"CSV", "Comma-separated values", ".csv"},
		{"JSON", "JavaScript Object Notation", ".json"},
		{"ASCII", "ASCII table with borders", ".txt"},
	}

	var b strings.Builder
	b.WriteString(m.styles.AccentText.Bold(true).Render("Save Query Results") + "\n")
	b.WriteString(strings.Repeat("═", 30) + "\n\n")
	b.WriteString("Select format:\n\n")

	for i, f := range formats {
		prefix := "  "
		// Shortened line to fit within modal width
		line := fmt.Sprintf("%-7s - %s %s", f.name, f.desc, f.ext)

		if i == m.saveModalFormat {
			prefix = m.styles.AccentText.Render("▶ ")
			line = m.styles.AccentText.Render(line)
		}
		b.WriteString(prefix + line + "\n")
	}

	b.WriteString("\n" + m.styles.MutedText.Render("↑/↓: navigate, Enter: confirm, ESC: cancel"))
	return b.String()
}

// renderSaveFilenameInput renders the filename input step
func (m *MainModel) renderSaveFilenameInput() string {
	formats := []string{"CSV", "JSON", "ASCII"}
	format := formats[m.saveModalFormat]

	// Generate default filename
	timestamp := time.Now().Format("20060102_150405")
	ext := ".csv"
	switch m.saveModalFormat {
	case 1: // JSON
		ext = ".json"
	case 2: // ASCII
		ext = ".txt"
	}
	defaultName := fmt.Sprintf("query_results_%s%s", timestamp, ext)

	var b strings.Builder
	b.WriteString(m.styles.AccentText.Bold(true).Render(fmt.Sprintf("Save as %s", format)) + "\n")
	b.WriteString(strings.Repeat("═", 40) + "\n\n")
	b.WriteString("Enter file path:\n")

	// Initialize the input if needed
	if m.saveModalInput.Width == 0 {
		input := textinput.New()
		input.Placeholder = defaultName
		input.CharLimit = 256
		input.Width = 50
		input.Focus()
		m.saveModalInput = input
	}

	// Update the value if it changed
	if m.saveModalFilename != m.saveModalInput.Value() {
		m.saveModalInput.SetValue(m.saveModalFilename)
	}

	b.WriteString(m.saveModalInput.View())
	b.WriteString("\n\n" + m.styles.MutedText.Render("Default: "+defaultName))
	b.WriteString("\n\n" + m.styles.MutedText.Render("[Enter to save, ESC to cancel, ← to go back]"))

	return b.String()
}

// handleSaveModalKeyboard handles keyboard input for the save modal
func (m *MainModel) handleSaveModalKeyboard(msg tea.KeyMsg) (*MainModel, tea.Cmd) {
	if !m.saveModalActive {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		// Cancel and close modal
		m.saveModalActive = false
		m.saveModalStep = 0
		m.saveModalFormat = 0
		m.saveModalFilename = ""
		if m.saveModalInput.Width > 0 {
			m.saveModalInput.Reset()
		}
		return m, nil

	case tea.KeyEnter:
		if m.saveModalStep == 0 {
			// Move to filename input step
			m.saveModalStep = 1
			m.saveModalFilename = ""

			// Initialize the input if not already done
			if m.saveModalInput.Width == 0 {
				input := textinput.New()
				input.CharLimit = 256
				input.Width = 50
				input.Focus()
				m.saveModalInput = input
			} else {
				m.saveModalInput.Reset()
				m.saveModalInput.Focus()
			}
			return m, nil
		} else {
			// Execute save
			return m.executeSaveFromModal()
		}

	case tea.KeyUp:
		if m.saveModalStep == 0 && m.saveModalFormat > 0 {
			m.saveModalFormat--
		}
		return m, nil

	case tea.KeyDown:
		if m.saveModalStep == 0 && m.saveModalFormat < 2 {
			m.saveModalFormat++
		}
		return m, nil

	case tea.KeyLeft:
		if m.saveModalStep == 1 {
			// Go back to format selection
			m.saveModalStep = 0
			m.saveModalFilename = ""
			if m.saveModalInput.Width > 0 {
				m.saveModalInput.Reset()
			}
			return m, nil
		}

	default:
		// Handle text input for filename
		if m.saveModalStep == 1 {
			var cmd tea.Cmd
			m.saveModalInput, cmd = m.saveModalInput.Update(msg)
			m.saveModalFilename = m.saveModalInput.Value()
			return m, cmd
		}
	}

	return m, nil
}

// executeSaveFromModal executes the save operation from the modal
func (m *MainModel) executeSaveFromModal() (*MainModel, tea.Cmd) {
	formats := []string{"CSV", "JSON", "ASCII"}
	format := formats[m.saveModalFormat]

	// Get filename (use default if empty)
	filename := m.saveModalFilename
	if filename == "" {
		filename = router.GenerateDefaultFilename(format)
	}

	// Create save command
	cmd := router.SaveCommand{
		Filename: filename,
		Format:   format,
		Options:  make(map[string]interface{}),
	}

	// Add default options based on format
	if format == "JSON" {
		cmd.Options["pretty"] = true
	}

	// Execute save
	err := router.HandleSaveCommand(cmd, m.lastTableData)

	// Close modal
	m.saveModalActive = false
	m.saveModalStep = 0
	m.saveModalFormat = 0
	m.saveModalFilename = ""
	if m.saveModalInput.Width > 0 {
		m.saveModalInput.Reset()
	}

	// Show result in history
	if err != nil {
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: "+err.Error())
	} else {
		rowCount := len(m.lastTableData) - 1 // Exclude header
		successMsg := fmt.Sprintf("Successfully saved %d rows to %s", rowCount, filename)
		m.fullHistoryContent += "\n" + m.styles.SuccessText.Render(successMsg)
	}
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	return m, nil
}