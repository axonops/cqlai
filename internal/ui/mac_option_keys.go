package ui

import (
	tea "charm.land/bubbletea/v2"
)

// Option on a Mac.
//
// A Mac terminal sends Option+letter as the character the keyboard layout
// gives it, unless it is set to send Option as Alt: Option+F types ƒ and
// Option+H types ˙, so Alt+F and Alt+H never arrived and the FILE menu and the
// help could not be opened from the keyboard in iTerm2 or Terminal.app as they
// come.
//
// The characters below are read as the shortcuts they stand for. They are the
// ones the standard US layout gives, and none of them is written in a CQL
// statement. Option+A is not among them: it types å, a letter people write in
// Danish, Norwegian and Swedish data, and reading it as Alt+A would open the
// AI in the middle of typing a name. For Alt+A, and every other Alt shortcut,
// the terminal can be set to send Option as Alt.
var macOptionKeys = map[string]rune{
	"ƒ": 'f', // Alt+F: the FILE menu, and a word forward in a text field
	"˙": 'h', // Alt+H: the help
	"∂": 'd', // Alt+D: delete the next word
	"∫": 'b', // Alt+B: a word back in a text field
}

// macOptionKey is the Alt shortcut a Mac's Option key typed a character for,
// or the key as it came.
func macOptionKey(msg tea.KeyPressMsg) tea.KeyPressMsg {
	if msg.Mod != 0 {
		return msg
	}
	if letter, ok := macOptionKeys[msg.Text]; ok {
		return tea.KeyPressMsg{Code: letter, Mod: tea.ModAlt}
	}
	return msg
}
