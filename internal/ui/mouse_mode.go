package ui

import (
	"fmt"
)

// Alternate scroll mode (DECSET 1007).
//
// cqlai runs on the alternate screen and only ever wanted the scroll wheel.
// Asking for that with mouse reporting (DECSET 1000) is too blunt: the terminal
// then hands us every click, so its own text selection and right-click paste
// stop working and users have to hold Shift to select anything.
//
// Alternate scroll mode gets the wheel without the buttons. While the alternate
// screen is active and mouse reporting is off, the terminal turns wheel events
// into Up/Down key presses and sends those instead, so it keeps the buttons and
// we still get to scroll. See handleUpArrow/handleDownArrow for the receiving
// end.
const (
	seqEnableAlternateScroll  = "\x1b[?1007h"
	seqDisableAlternateScroll = "\x1b[?1007l"
)

// EnableAlternateScroll turns on alternate scroll mode.
//
// Bubble Tea has no command for this mode, so it is written straight to stdout
// the same way the mouse sequences used to be.
func EnableAlternateScroll() {
	fmt.Print(seqEnableAlternateScroll)
}

// DisableAlternateScroll gives the wheel back to the terminal.
func DisableAlternateScroll() {
	fmt.Print(seqDisableAlternateScroll)
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
