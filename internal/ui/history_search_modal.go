package ui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// HistorySearchModal represents a modal for searching command history
type HistorySearchModal struct {
	query        string
	results      []string
	selected     int
	maxShow      int
	scrollOffset int // Tracks the first visible item index
	maxWidth     int // Maximum width for display
}

// NewHistorySearchModal creates a new history search modal
func NewHistorySearchModal(query string, results []string, selectedIndex int, screenWidth int) HistorySearchModal {
	// Calculate reasonable max width (80% of screen width, max 100 chars)
	maxWidth := int(float64(screenWidth) * 0.8)
	if maxWidth > 100 {
		maxWidth = 100
	}
	if maxWidth < 40 {
		maxWidth = 40
	}
	
	return HistorySearchModal{
		query:        query,
		results:      results,
		selected:     selectedIndex,
		maxShow:      10,
		scrollOffset: 0,
		maxWidth:     maxWidth,
	}
}

// truncateWithEllipsis truncates a string to maxWidth and adds ellipsis if needed
func (hsm HistorySearchModal) truncateWithEllipsis(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth < 4 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

// RenderContent renders just the modal content without positioning
func (hsm HistorySearchModal) RenderContent(styles *Styles) string {
	// First, determine the actual content width we need
	minWidth := 40 // Minimum width for readability
	maxContentWidth := 0
	
	// Check title width
	titleWidth := len("History Search (Ctrl+R)") + 4 // +4 for padding
	if titleWidth > maxContentWidth {
		maxContentWidth = titleWidth
	}
	
	// Check search query width
	queryWidth := len("Search: ") + len(hsm.query) + 4
	if hsm.query == "" {
		queryWidth = len("Search: (type to search)") + 4
	}
	if queryWidth > maxContentWidth {
		maxContentWidth = queryWidth
	}
	
	// Check results width
	for _, result := range hsm.results {
		resultWidth := len(result) + 6 // +6 for arrow/spaces and padding
		if resultWidth > maxContentWidth {
			maxContentWidth = resultWidth
		}
	}
	
	// Check instruction width
	instructionWidth := len("↑↓: Navigate • Enter: Select • Esc: Cancel") + 4
	if instructionWidth > maxContentWidth {
		maxContentWidth = instructionWidth
	}
	
	// Use the calculated width, but respect min/max bounds
	boxWidth := maxContentWidth
	if boxWidth < minWidth {
		boxWidth = minWidth
	}
	if boxWidth > hsm.maxWidth {
		boxWidth = hsm.maxWidth
	}
	
	boxContentWidth := boxWidth - 6 // Account for border and arrow
	
	// Create the modal style WITHOUT width or background - let content determine size
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent)
	
	// Build the content
	var content []string
	
	// Title
	titleText := "History Search (Ctrl+R)"
	// Pad title to desired width
	padding := (boxWidth - len(titleText)) / 2
	if padding > 0 {
		titleText = strings.Repeat(" ", padding) + titleText + strings.Repeat(" ", padding)
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Accent).
		Bold(true)
	content = append(content, titleStyle.Render(titleText))
	
	// Search query
	queryStyle := lipgloss.NewStyle().
		Foreground(styles.AccentText.GetForeground())
	queryDisplay := hsm.query
	if queryDisplay == "" {
		queryDisplay = "(type to search)"
	}
	queryText := " Search: " + queryDisplay
	// Pad to width
	if len(queryText) < boxWidth {
		queryText += strings.Repeat(" ", boxWidth - len(queryText))
	}
	content = append(content, queryStyle.Render(queryText))
	
	// Results section
	if len(hsm.results) == 0 {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(styles.MutedText.GetForeground()).
			Italic(true)
		noResultsText := "No matching commands"
		padding := (boxWidth - len(noResultsText)) / 2
		if padding > 0 {
			noResultsText = strings.Repeat(" ", padding) + noResultsText + strings.Repeat(" ", padding)
		}
		content = append(content, noResultsStyle.Render(noResultsText))
	} else {
		// Determine visible range
		endIndex := hsm.scrollOffset + hsm.maxShow
		if endIndex > len(hsm.results) {
			endIndex = len(hsm.results)
		}
		displayResults := hsm.results[hsm.scrollOffset:endIndex]
		
		// Show scroll up indicator if not at top
		if hsm.scrollOffset > 0 {
			scrollStyle := lipgloss.NewStyle().
				Foreground(styles.MutedText.GetForeground())
			scrollText := "▲ (more)"
			padding := (boxWidth - len(scrollText)) / 2
			if padding > 0 {
				scrollText = strings.Repeat(" ", padding) + scrollText + strings.Repeat(" ", padding)
			}
			content = append(content, scrollStyle.Render(scrollText))
		}
		
		// Results count indicator
		if len(hsm.results) > hsm.maxShow {
			countStyle := lipgloss.NewStyle().
				Foreground(styles.MutedText.GetForeground()).
				Italic(true)
			countText := fmt.Sprintf("Showing %d-%d of %d matches", 
				hsm.scrollOffset+1, endIndex, len(hsm.results))
			padding := (boxWidth - len(countText)) / 2
			if padding > 0 {
				countText = strings.Repeat(" ", padding) + countText + strings.Repeat(" ", padding)
			}
			content = append(content, countStyle.Render(countText))
		}
		
		// Show results
		for i, result := range displayResults {
			actualIndex := hsm.scrollOffset + i
			// Truncate the result if it's too long
			displayText := hsm.truncateWithEllipsis(result, boxContentWidth)
			
			var line string
			if actualIndex == hsm.selected {
				// Selected item with arrow
				itemStyle := lipgloss.NewStyle().
					Foreground(styles.Accent).
					Bold(true)
				itemText := " → " + displayText
				// Pad to width
				if len(itemText) < boxWidth {
					itemText += strings.Repeat(" ", boxWidth - len(itemText))
				}
				line = itemStyle.Render(itemText)
			} else {
				// Regular item
				itemStyle := lipgloss.NewStyle().
					Foreground(styles.MutedText.GetForeground())
				itemText := "   " + displayText  // 3 spaces to align with arrow
				// Pad to width
				if len(itemText) < boxWidth {
					itemText += strings.Repeat(" ", boxWidth - len(itemText))
				}
				line = itemStyle.Render(itemText)
			}
			content = append(content, line)
		}
		
		// Show scroll down indicator if not at bottom
		if endIndex < len(hsm.results) {
			scrollDownStyle := lipgloss.NewStyle().
				Foreground(styles.MutedText.GetForeground())
			scrollDownText := "▼ (more)"
			downPadding := (boxWidth - len(scrollDownText)) / 2
			if downPadding > 0 {
				scrollDownText = strings.Repeat(" ", downPadding) + scrollDownText + strings.Repeat(" ", downPadding)
			}
			content = append(content, scrollDownStyle.Render(scrollDownText))
		}
	}
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MutedText.GetForeground()).
		Italic(true)
	instructionText := "↑↓: Navigate • Enter: Select • Esc: Cancel"
	instrPadding := (boxWidth - len(instructionText)) / 2
	if instrPadding > 0 {
		instructionText = strings.Repeat(" ", instrPadding) + instructionText + strings.Repeat(" ", instrPadding)
	}
	content = append(content, instructionStyle.Render(instructionText))
	
	// Join all content
	modalContent := strings.Join(content, "\n")
	return modalStyle.Render(modalContent)
}

// GetOverlay returns the modal content and positioning information
func (hsm HistorySearchModal) GetOverlay(screenWidth, screenHeight int, styles *Styles) ModalOverlay {
	content := hsm.RenderContent(styles)
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