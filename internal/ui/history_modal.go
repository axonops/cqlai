package ui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// HistoryModal represents a modal for showing command history
type HistoryModal struct {
	items        []string
	selected     int
	maxShow      int
	scrollOffset int // Tracks the first visible item index
	maxWidth     int // Maximum width for display
}

// NewHistoryModal creates a new history modal
func NewHistoryModal(history []string, currentIndex int, screenWidth int) HistoryModal {
	// Calculate reasonable max width (80% of screen width, max 100 chars)
	maxWidth := int(float64(screenWidth) * 0.8)
	if maxWidth > 100 {
		maxWidth = 100
	}
	if maxWidth < 40 {
		maxWidth = 40
	}
	
	return HistoryModal{
		items:        history,
		selected:     currentIndex,
		maxShow:      10,
		scrollOffset: 0,
		maxWidth:     maxWidth,
	}
}

// truncateWithEllipsis truncates a string to maxWidth and adds ellipsis if needed
func (hm HistoryModal) truncateWithEllipsis(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth < 4 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

// RenderContent renders just the modal content without positioning
func (hm HistoryModal) RenderContent(styles *Styles) string {
	if len(hm.items) == 0 {
		return ""
	}
	
	// Show most recent commands first - reverse the order
	reversedItems := make([]string, len(hm.items))
	for i, item := range hm.items {
		reversedItems[len(hm.items)-1-i] = item
	}

	// Convert the selected index from original order to reversed order
	// If selected is at end of original array (newest), it should be at start of reversed (index 0)
	reversedSelected := len(hm.items) - 1 - hm.selected

	// Determine visible range
	endIndex := hm.scrollOffset + hm.maxShow
	if endIndex > len(reversedItems) {
		endIndex = len(reversedItems)
	}
	displayItems := reversedItems[hm.scrollOffset:endIndex]

	// Calculate the box width based on content (with a reasonable max)
	boxContentWidth := hm.maxWidth - 6 // Account for border and arrow
	boxWidth := hm.maxWidth

	// Create the modal style without forced background
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent)

	// Build the content
	var content []string

	// Title with scroll indicator
	titleText := "Command History"
	if len(reversedItems) > hm.maxShow {
		titleText = fmt.Sprintf("Command History (%d-%d of %d)",
			hm.scrollOffset+1, endIndex, len(reversedItems))
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true).
		Align(lipgloss.Center)
	content = append(content, titleStyle.Render(titleText))


	// Show scroll up indicator if not at top
	if hm.scrollOffset > 0 {
		scrollStyle := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			Align(lipgloss.Center)
		content = append(content, scrollStyle.Render("▲ (older)"))
	}

	// Items
	for i, item := range displayItems {
		actualIndex := hm.scrollOffset + i
		// Truncate the item if it's too long
		displayText := hm.truncateWithEllipsis(item, boxContentWidth)

		var line string
		// Compare with the reversed selected index
		if actualIndex == reversedSelected {
			// Selected item with arrow
			itemStyle := lipgloss.NewStyle().
				Foreground(styles.Accent).
				Bold(true)
			line = itemStyle.Render("→ " + displayText)
		} else {
			// Regular item
			itemStyle := lipgloss.NewStyle().
				Foreground(styles.MutedText.GetForeground())
			line = itemStyle.Render("  " + displayText)
		}
		content = append(content, line)
	}
	
	// Show scroll down indicator if not at bottom
	if endIndex < len(reversedItems) {
		scrollStyle := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			Align(lipgloss.Center)
		content = append(content, scrollStyle.Render("▼ (newer)"))
	}
	
	// Instructions - add separator
	content = append(content, strings.Repeat("─", boxWidth - 2))
	
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true).
		Align(lipgloss.Center)
	content = append(content, instructionStyle.Render("↑↓: Navigate • Enter: Select • Esc: Close"))
	
	// Join all content
	modalContent := strings.Join(content, "\n")
	return modalStyle.Render(modalContent)
}

// GetOverlay returns the modal content and positioning information
func (hm HistoryModal) GetOverlay(screenWidth, screenHeight int, styles *Styles) ModalOverlay {
	content := hm.RenderContent(styles)
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