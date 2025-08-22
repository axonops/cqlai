package ui

import (
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// Layer represents a renderable layer with position and size
type Layer struct {
	Content string
	X       int
	Y       int
	Width   int
	Height  int
	ZIndex  int
}

// LayerManager manages multiple layers for rendering
type LayerManager struct {
	layers []Layer
	width  int
	height int
}

// NewLayerManager creates a new layer manager
func NewLayerManager(width, height int) *LayerManager {
	return &LayerManager{
		layers: []Layer{},
		width:  width,
		height: height,
	}
}

// AddLayer adds a new layer to the manager
func (lm *LayerManager) AddLayer(layer Layer) {
	lm.layers = append(lm.layers, layer)
}

// Clear removes all layers
func (lm *LayerManager) Clear() {
	lm.layers = []Layer{}
}

// Render composites all layers into a single view
func (lm *LayerManager) Render(base string) string {
	// Start with the base content
	lines := strings.Split(base, "\n")
	
	// Ensure we have enough lines
	for len(lines) < lm.height {
		lines = append(lines, "")
	}
	
	// Sort layers by z-index (higher z-index on top)
	// For now, we'll just render in order since we typically only have one modal
	
	// Apply each layer
	for _, layer := range lm.layers {
		lines = lm.applyLayer(lines, layer)
	}
	
	return strings.Join(lines, "\n")
}

// applyLayer applies a single layer to the view
func (lm *LayerManager) applyLayer(lines []string, layer Layer) []string {
	contentLines := strings.Split(layer.Content, "\n")
	
	for i, contentLine := range contentLines {
		lineIdx := layer.Y + i
		if lineIdx >= 0 && lineIdx < len(lines) {
			bgLine := lines[lineIdx]
			
			// Calculate the actual visual width of this specific content line
			contentLineWidth := lipgloss.Width(contentLine)
			
			// Simply replace the line with modal content + remaining background
			// The modal should be self-contained with its own background
			result := contentLine
			
			// Get the visual width of the background line
			bgWidth := lipgloss.Width(bgLine)
			
			// If the background extends beyond the modal, preserve it
			if layer.X + contentLineWidth < bgWidth {
				// We need to extract the part of background that's after the modal
				// This is complex with ANSI codes, so for now we'll pad with spaces
				remainingWidth := bgWidth - (layer.X + contentLineWidth)
				if remainingWidth > 0 {
					// Extract the visual content after the modal position
					bgPlain := stripAnsi(bgLine)
					bgRunes := []rune(bgPlain)
					startPos := layer.X + contentLineWidth
					if startPos < len(bgRunes) {
						result += string(bgRunes[startPos:])
					}
				}
			}
			
			lines[lineIdx] = result
		}
	}
	
	return lines
}

// RenderModal renders a modal as a centered overlay
func RenderModal(content string, width, height int) Layer {
	// Calculate position to center the modal
	modalLines := strings.Split(content, "\n")
	modalHeight := len(modalLines)
	modalWidth := 0
	
	for _, line := range modalLines {
		w := lipgloss.Width(line)
		if w > modalWidth {
			modalWidth = w
		}
	}
	
	// Center the modal
	x := (width - modalWidth) / 2
	if x < 0 {
		x = 0
	}
	
	y := (height - modalHeight) / 2
	if y < 0 {
		y = 0
	}
	
	return Layer{
		Content: content,
		X:       x,
		Y:       y,
		Width:   modalWidth,
		Height:  modalHeight,
		ZIndex:  100,
	}
}