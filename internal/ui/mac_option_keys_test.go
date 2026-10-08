package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// Option+letter on a Mac, as the terminal sends it without Option set to Alt.

func typed(s string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

// TestOptionFOpensTheFileMenu: Option+F types ƒ.
func TestOptionFOpensTheFileMenu(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m, _ = m.handleKeyboardInput(typed("ƒ"))
	assert.True(t, m.fileMenu.active)
	assert.Empty(t, m.input.Value(), "and ƒ is not typed")
}

// TestOptionHOpensTheHelp: Option+H types ˙.
func TestOptionHOpensTheHelp(t *testing.T) {
	m := helpModel()
	m.input.Focus()
	m, _ = m.handleKeyboardInput(typed("˙"))
	assert.True(t, m.help.active)
}

// TestOptionDDeletesTheNextWord: Option+D types ∂.
func TestOptionDDeletesTheNextWord(t *testing.T) {
	m := editModel("SELECT * FROM shop.orders", 0)
	m, _ = m.handleKeyboardInput(typed("∂"))
	assert.Equal(t, " * FROM shop.orders", m.input.Value())
}

// TestOptionAStillTypesALetter: å is a letter in Danish, Norwegian and
// Swedish, and typing it must not open the AI.
func TestOptionAStillTypesALetter(t *testing.T) {
	m := editModel("SELECT * FROM shop.cities WHERE name = '", len("SELECT * FROM shop.cities WHERE name = '"))
	m, _ = m.handleKeyboardInput(typed("Å"))
	m, _ = m.handleKeyboardInput(typed("l"))
	m, _ = m.handleKeyboardInput(typed("å"))
	assert.Equal(t, "SELECT * FROM shop.cities WHERE name = 'Ålå", m.input.Value())
	assert.Equal(t, ModalNone, m.modal.Type)
}

// TestOptionBMovesBackAWord: Option+B types ∫.
func TestOptionBMovesBackAWord(t *testing.T) {
	m := editModel("SELECT * FROM shop.orders", len("SELECT * FROM shop.orders"))
	m, _ = m.handleKeyboardInput(typed("∫"))
	assert.Equal(t, "SELECT * FROM shop.orders", m.input.Value(), "and ∫ is not typed")
	assert.Less(t, m.input.Position(), len("SELECT * FROM shop.orders"))
}
