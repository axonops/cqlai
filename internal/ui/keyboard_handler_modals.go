package ui

import (
	tea "charm.land/bubbletea/v2"
)

// handleAICQLModal handles keyboard input for the AI CQL execution modal
func (m *MainModel) handleAICQLModal(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.aiCQLModal.PrevChoice()
		return m, nil
	case "right", "l":
		m.aiCQLModal.NextChoice()
		return m, nil
	case "enter":
		switch m.aiCQLModal.Selected {
		case 0:
			// Execute the CQL
			cql := m.aiCQLModal.CQL
			m.aiCQLModal.Active = false
			m.aiCQLModal = nil

			// Exit AI conversation view
			m.aiConversationActive = false
			m.viewMode = "history"

			// Execute the CQL command
			m.input.SetValue(cql)
			// Trigger the enter key handler to execute
			return m.handleEnterKey()
		case 1:
			// Edit - put the CQL in the input field and close modal
			cql := m.aiCQLModal.CQL
			m.aiCQLModal.Active = false
			m.aiCQLModal = nil

			// Exit AI conversation view and go to query view
			m.aiConversationActive = false
			m.viewMode = "history"

			// Put the CQL in the input field for editing
			m.input.SetValue(cql)
			m.input.CursorEnd()
			return m, nil
		case 2:
			// Cancel - close modal and stay in AI conversation
			m.aiCQLModal.Active = false
			m.aiCQLModal = nil
			return m, nil
		}
	case "esc", "ctrl+c":
		// Cancel - close modal and stay in AI conversation
		m.aiCQLModal.Active = false
		m.aiCQLModal = nil
		return m, nil
	}
	return m, nil
}

// handleAISelectionModal handles keyboard input for the AI selection modal
func (m *MainModel) handleAISelectionModal(msg tea.KeyPressMsg) (*MainModel, tea.Cmd) {
	if m.aiSelectionModal.InputMode {
		// In custom input mode
		switch msg.String() {
		case "esc":
			// Exit input mode
			m.aiSelectionModal.InputMode = false
			m.aiSelectionModal.CustomInput = ""
			return m, nil
		case "enter":
			// Submit custom input
			if m.aiSelectionModal.CustomInput != "" {
				return m, func() tea.Msg {
					return AISelectionResultMsg{
						Selection:     m.aiSelectionModal.CustomInput,
						SelectionType: m.aiSelectionModal.SelectionType,
						Cancelled:     false,
					}
				}
			}
			return m, nil
		case "backspace":
			if len(m.aiSelectionModal.CustomInput) > 0 {
				m.aiSelectionModal.CustomInput = m.aiSelectionModal.CustomInput[:len(m.aiSelectionModal.CustomInput)-1]
			}
			return m, nil
		default:
			// Add character to custom input
			if msg.Text != "" {
				m.aiSelectionModal.CustomInput += string(msg.Text)
			}
			return m, nil
		}
	} else {
		// In selection mode
		switch msg.String() {
		case "up", "k":
			m.aiSelectionModal.PrevOption()
			return m, nil
		case "down", "j":
			m.aiSelectionModal.NextOption()
			return m, nil
		case "i":
			// Enter custom input mode
			m.aiSelectionModal.InputMode = true
			m.aiSelectionModal.CustomInput = ""
			return m, nil
		case "enter":
			// Select current option
			return m, func() tea.Msg {
				return AISelectionResultMsg{
					Selection:     m.aiSelectionModal.GetSelection(),
					SelectionType: m.aiSelectionModal.SelectionType,
					Cancelled:     false,
				}
			}
		case "esc", "ctrl+c":
			// Cancel selection
			return m, func() tea.Msg {
				return AISelectionResultMsg{
					Cancelled: true,
				}
			}
		}
	}
	return m, nil
}
