package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// Picking a completion with the mouse.

// completionOverlay builds the completion list and says where it goes. The
// view draws it here and clicks are tested against it here, so a click cannot
// land on a different row from the one drawn there.
func (m *MainModel) completionOverlay(screenHeight int) (CompletionModal, Layer, bool) {
	if !m.showCompletions || len(m.completions) == 0 {
		return CompletionModal{}, Layer{}, false
	}
	list := NewCompletionModal(m.completions, m.completionIndex)
	list.scrollOffset = m.completionScrollOffset

	// Above the prompt rather than over it: the list narrows as you type, so
	// what is typed has to stay in sight. The two bars are below the prompt.
	layer := overlayLayer(list.RenderContent(m.styles), screenHeight)
	layer.Y = max(screenHeight-layer.Height-3, 0)
	return list, layer, true
}

// completionAt is the completion under a press, and whether the press was
// inside the list at all.
func (m *MainModel) completionAt(col, row int) (index int, onItem, inside bool) {
	list, layer, ok := m.completionOverlay(m.windowHeight)
	if !ok || col < layer.X || col >= layer.X+layer.Width || row < layer.Y || row >= layer.Y+layer.Height {
		return 0, false, false
	}

	// The border, the title, and the arrow when the list is scrolled.
	first := layer.Y + 2
	if list.scrollOffset > 0 {
		first++
	}
	shown := min(list.maxShow, len(list.items)-list.scrollOffset)
	if row < first || row >= first+shown {
		return 0, false, true
	}
	return list.scrollOffset + row - first, true, true
}

// clickCompletion answers a press while the completion list is showing. A
// press on a completion uses it, the same as Enter on it. A press elsewhere in
// the box does nothing. A press outside it closes the list and carries on.
func (m *MainModel) clickCompletion(col, row int) (*MainModel, tea.Cmd, bool) {
	i, onItem, inside := m.completionAt(col, row)
	switch {
	case onItem && !completion.IsHint(m.completions[i]):
		m.completionIndex = i
		updated, cmd := m.handleCompletionSelection()
		return updated, cmd, true
	case inside:
		// A note about what to type, the title or the key help.
		return m, nil, true
	}
	m.clearCompletions()
	return m, nil, false
}
