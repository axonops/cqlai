package ui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// CompletionModal represents a modal for showing completions
type CompletionModal struct {
	items        []string
	selected     int
	maxShow      int
	scrollOffset int // Tracks the first visible item index
}

// NewCompletionModal creates a new completion modal
func NewCompletionModal(completions []string, currentIndex int) CompletionModal {
	return CompletionModal{
		items:        completions,
		selected:     currentIndex,
		maxShow:      10,
		scrollOffset: 0,
	}
}

// RenderContent renders just the modal content without positioning
func (cm CompletionModal) RenderContent(styles *Styles) string {
	if len(cm.items) == 0 {
		return ""
	}
	
	// Determine visible range
	endIndex := cm.scrollOffset + cm.maxShow
	if endIndex > len(cm.items) {
		endIndex = len(cm.items)
	}
	displayItems := cm.items[cm.scrollOffset:endIndex]
	
	// Calculate the maximum width needed (check all items for consistent width)
	maxWidth := 0
	for _, item := range cm.items {
		if len(item) > maxWidth {
			maxWidth = len(item)
		}
	}
	
	// Add space for arrow and padding
	boxWidth := maxWidth + 6
	if boxWidth < 30 {
		boxWidth = 30
	}
	
	// Create the modal style with solid background
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Background(lipgloss.Color("#2D2D2D")).
		Width(boxWidth - 2)
	
	// Build the content
	var content []string
	
	// Title with scroll indicator
	titleText := "Completions"
	if len(cm.items) > cm.maxShow {
		titleText = fmt.Sprintf("Completions (%d-%d of %d)", 
			cm.scrollOffset+1, endIndex, len(cm.items))
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Width(boxWidth - 2).
		Align(lipgloss.Center)
	content = append(content, titleStyle.Render(titleText))
	
	
	// Show scroll up indicator if not at top
	if cm.scrollOffset > 0 {
		scrollStyle := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			Width(boxWidth - 2).
			Align(lipgloss.Center)
		content = append(content, scrollStyle.Render("▲ "))
	}
	
	// Items
	for i, item := range displayItems {
		actualIndex := cm.scrollOffset + i
		var line string
		if actualIndex == cm.selected {
			// Selected item with arrow
			itemStyle := lipgloss.NewStyle().
				Foreground(styles.Accent).
				Bold(true).
				Width(boxWidth - 2)
			line = itemStyle.Render("→ " + item)
		} else {
			// Regular item
			itemStyle := lipgloss.NewStyle().
				Foreground(styles.MutedText.GetForeground()).
				Width(boxWidth - 2)
			line = itemStyle.Render("  " + item)
		}
		content = append(content, line)
	}
	
	// Show scroll down indicator if not at bottom
	if endIndex < len(cm.items) {
		scrollStyle := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			Width(boxWidth - 2).
			Align(lipgloss.Center)
		content = append(content, scrollStyle.Render("▼"))
	}
	
	// Instructions
	content = append(content, strings.Repeat("─", boxWidth))
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Width(boxWidth - 2).
		Align(lipgloss.Center)
	content = append(content, instructionStyle.Render("↑↓/Tab: Navigate • Enter: Accept • Esc: Close"))
	
	// Join all content
	modalContent := strings.Join(content, "\n")
	return modalStyle.Render(modalContent)
}

// GetOverlay returns the modal content and positioning information
func (cm CompletionModal) GetOverlay(screenWidth, screenHeight int, styles *Styles) ModalOverlay {
	content := cm.RenderContent(styles)
	modalHeight := strings.Count(content, "\n") + 1
	
	// Calculate the actual width from the rendered content
	modalLines := strings.Split(content, "\n")
	modalWidth := 0
	for _, line := range modalLines {
		lineWidth := len(stripAnsi(line))
		if lineWidth > modalWidth {
			modalWidth = lineWidth
		}
	}
	
	// Left align the modal with no margin
	x := 0
	
	// Position just above the input line (prompt)
	// The last two lines are the prompt and status bar
	y := screenHeight - modalHeight - 2
	if y < 0 {
		y = 0
	}
	
	return ModalOverlay{
		Content: content,
		X:       x,
		Y:       y,
		Width:   modalWidth,
		Height:  modalHeight,
	}
}