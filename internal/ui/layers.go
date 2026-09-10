package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
// applyLayer draws one layer over the lines beneath it.
//
// The background either side of the layer keeps its styling. It used to be
// stripped of ANSI and pasted back as plain text, so anything coloured beside
// a modal lost its colour for as long as the modal was up - most visibly the
// Console scrollbar, which turned white for exactly the rows the modal
// covered.
func (lm *LayerManager) applyLayer(lines []string, layer Layer) []string {
	contentLines := strings.Split(layer.Content, "\n")

	for i, contentLine := range contentLines {
		lineIdx := layer.Y + i
		if lineIdx < 0 || lineIdx >= len(lines) {
			continue
		}

		bgLine := lines[lineIdx]
		bgWidth := lipgloss.Width(bgLine)
		contentWidth := lipgloss.Width(contentLine)

		// What is in front of the layer, styling and all. Past the end of the
		// background there is nothing to keep, so pad.
		var result string
		if layer.X > 0 {
			result = ansi.Cut(bgLine, 0, layer.X)
			if pad := layer.X - lipgloss.Width(result); pad > 0 {
				result += strings.Repeat(" ", pad)
			}
			// The layer draws its own colours; the background's must not run
			// on into it.
			result += ansiReset
		}

		result += contentLine

		// And what is behind it, from where the layer ends.
		if end := layer.X + contentWidth; end < bgWidth {
			result += ansiReset + ansi.Cut(bgLine, end, bgWidth)
		}

		lines[lineIdx] = result
	}

	return lines
}

// ansiReset closes any styling still in effect, so one piece of a composited
// line cannot colour the next.
const ansiReset = "\x1b[0m"

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
