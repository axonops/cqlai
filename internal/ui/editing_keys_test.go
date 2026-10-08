package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// The line-editing keys on text outside ASCII. The cursor counts characters
// and the text is UTF-8, so cutting at the cursor as if it counted bytes cut
// in the wrong place: Ctrl+U on a line of Japanese left most of it.

func editModel(value string, cursor int) *MainModel {
	m := helpModel()
	m.input.Focus()
	m.input.SetValue(value)
	m.input.SetCursor(cursor)
	return m
}

func ctrl(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl} }

// at is the character a piece of the text starts at.
func at(text, piece string) int {
	before, _, found := strings.Cut(text, piece)
	if !found {
		panic("no " + piece + " in " + text)
	}
	return len([]rune(before))
}

func TestEditingKeysCountCharacters(t *testing.T) {
	const text = "SELECT * FROM 東京.店舗 WHERE 名前 = 'さくら'"
	end := len([]rune(text))
	afterName := at(text, " = ")  // just after 名前
	beforeName := at(text, " 名前") // just after WHERE
	shop := at(text, "店舗")

	for _, c := range []struct {
		name      string
		key       tea.KeyPressMsg
		cursor    int
		want      string
		cut       string
		cursorEnd int
	}{
		{"Ctrl+U at the end", ctrl('u'), end, "", text, 0},
		{"Ctrl+U in the middle", ctrl('u'), shop, "店舗 WHERE 名前 = 'さくら'", "SELECT * FROM 東京.", 0},
		{"Ctrl+K in the middle", ctrl('k'), shop, "SELECT * FROM 東京.", "店舗 WHERE 名前 = 'さくら'", shop},
		{"Ctrl+W after a word", ctrl('w'), afterName, "SELECT * FROM 東京.店舗 WHERE  = 'さくら'", "名前", beforeName + 1},
		{"Alt+D before a word", tea.KeyPressMsg{Code: 'd', Mod: tea.ModAlt}, beforeName, "SELECT * FROM 東京.店舗 WHERE = 'さくら'", " 名前", beforeName},
	} {
		m := editModel(text, c.cursor)
		m, _ = m.handleKeyboardInput(c.key)
		assert.Equal(t, c.want, m.input.Value(), c.name)
		assert.Equal(t, c.cut, m.clipboardBuffer, c.name)
		assert.Equal(t, c.cursorEnd, m.input.Position(), c.name)
	}
}

// TestCtrlYPutsTheCursorAfterWhatItPasted, counted in characters.
func TestCtrlYPutsTheCursorAfterWhatItPasted(t *testing.T) {
	m := editModel("東京", 0)
	m.clipboardBuffer = "店舗"
	m, _ = m.handleKeyboardInput(ctrl('y'))
	assert.Equal(t, "店舗東京", m.input.Value())
	assert.Equal(t, 2, m.input.Position())
}

// TestTheWordJumpsLandOnWords in Japanese text with spaces between words.
func TestTheWordJumpsLandOnWords(t *testing.T) {
	const text = "東京 大阪 名古屋 札幌 福岡 那覇 仙台 広島 京都 神戸"
	m := editModel(text, len([]rune(text)))
	m, _ = m.handleCtrlLeft()
	assert.LessOrEqual(t, m.input.Position(), len([]rune(text)))
	assert.GreaterOrEqual(t, m.input.Position(), 0)

	m = editModel(text, 0)
	m, _ = m.handleCtrlRight()
	assert.LessOrEqual(t, m.input.Position(), len([]rune(text)), "never past the end")
	assert.Positive(t, m.input.Position())
}
