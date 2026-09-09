package ui

import "fmt"

// Alternate scroll mode (DECSET 1007).
//
// This is the one terminal mode Bubble Tea does not manage for us. The rest -
// the alternate screen, and whether the mouse is reported at all - are fields
// on the View returned each render; see newView.
//
// It matters only while mouse reporting is off, which MOUSE OFF does. With no
// mouse events arriving, alternate scroll mode is what still lets the wheel
// scroll: on the alternate screen the terminal turns a wheel spin into Up and
// Down key presses and sends those instead. See handleUpArrow/handleDownArrow
// for the receiving end.
const (
	seqEnableAlternateScroll  = "\x1b[?1007h"
	seqDisableAlternateScroll = "\x1b[?1007l"
)

// EnableAlternateScroll turns on alternate scroll mode.
func EnableAlternateScroll() {
	fmt.Print(seqEnableAlternateScroll)
}

// DisableAlternateScroll gives the wheel back to the terminal.
func DisableAlternateScroll() {
	fmt.Print(seqDisableAlternateScroll)
}

// ResetMouseReporting turns every mouse reporting mode off.
//
// Bubble Tea already does this as it tears the screen down, so this is
// insurance for the paths that leave without it - a panic, or a signal - where
// the alternative is a terminal that keeps printing escape sequences at the
// shell every time the mouse moves.
func ResetMouseReporting() {
	fmt.Print("\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l")
}

// viewportOwnsArrows reports whether a bare Up/Down should scroll the current
// view instead of walking back through command history.
//
// This matters because of alternate scroll mode: a wheel spin reaches us as an
// Up or Down key press that is indistinguishable from a real one. Scrolling is
// the useful reading of both while a result set or a trace is on screen. In the
// normal view there is nothing to scroll through, so Up keeps recalling the last
// command, and Ctrl+P still recalls it in every view.
//
// The AI conversation view is not listed here because handleAIConversationInput
// consumes Up/Down before they reach handleUpArrow and does its own scrolling.
//
// Alt+Up/Down, PgUp/PgDn and navigation mode are unaffected and still scroll
// everywhere.
func (m *MainModel) viewportOwnsArrows() bool {
	switch m.viewMode {
	case "table":
		return m.hasTable
	case "trace":
		return m.hasTrace
	default:
		return false
	}
}
