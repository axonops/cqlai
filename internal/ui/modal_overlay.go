package ui

import (
	"strings"
)

// ModalOverlay represents positioning information for a modal
type ModalOverlay struct {
	Content string
	X       int // Horizontal position
	Y       int // Vertical position
	Width   int // Width of the modal
	Height  int // Height of the modal
}

// OverlayOnView overlays modal content onto a view, preserving background where modal doesn't cover
func OverlayModalOnView(background string, modal ModalOverlay) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal.Content, "\n")
	
	// Ensure we have enough lines in the background
	for len(bgLines) <= modal.Y + modal.Height {
		bgLines = append(bgLines, "")
	}
	
	// Overlay the modal onto the background
	for i, modalLine := range modalLines {
		lineIdx := modal.Y + i
		if lineIdx >= 0 && lineIdx < len(bgLines) {
			bgLine := bgLines[lineIdx]
			
			// Calculate the actual width of this specific modal line
			modalLineWidth := len(stripAnsi(modalLine))
			
			// Ensure the background line is long enough
			if len(bgLine) < modal.X + modalLineWidth {
				bgLine += strings.Repeat(" ", modal.X + modalLineWidth - len(bgLine))
			}
			
			// Get the part before the modal
			before := ""
			if modal.X > 0 {
				before = bgLine[:modal.X]
			}
			
			// Get the part after this specific modal line (preserve background to the right)
			after := ""
			if modal.X + modalLineWidth < len(bgLine) {
				after = bgLine[modal.X + modalLineWidth:]
			}
			
			// Combine: before + modal + after
			bgLines[lineIdx] = before + modalLine + after
		}
	}
	
	return strings.Join(bgLines, "\n")
}