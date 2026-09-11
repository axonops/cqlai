package ui

import (
	"charm.land/lipgloss/v2"
)

// Typing a statement over several lines.
//
// A statement without a semicolon is not finished, so cqlai keeps the prompt
// open and waits for the rest. The lines already typed used to be drawn above
// the prompt, which is a block that grows: the window is laid out with one row
// for the input, so the second line pushed the connection bar off the bottom of
// the screen, the third took the query line with it, and what was left was the
// statement sitting on top of the bars it had overwritten.
//
// They go into the console instead, a line at a time as they are typed, the way
// a shell does it. The transcript then reads as what you typed, the view
// scrolls as it always did, and nothing below the prompt moves.

const (
	// promptMark is what a line typed at the prompt is written with, and
	// continuationMark what the lines of an unfinished statement carry.
	promptMark       = "> "
	continuationMark = "... "
)

// setInputPrompt changes what the prompt says and keeps the input sized to it.
//
// The prompt and the trailing cursor cell are drawn outside the width the
// textinput is given, so the width has to be worked out from the prompt in
// front of it. It was only ever worked out on a resize, which is fine for a
// prompt that never changes and wrong for one that does.
func (m *MainModel) setInputPrompt(prompt string) {
	m.input.Prompt = m.styles.AccentText.Render(prompt)
	m.input.SetWidth(inputWidth(m.windowWidth, m.input.Prompt))
}

// inputWidth is how wide the text box can be with a given prompt in front of it.
//
// The prompt and the trailing cursor cell are drawn outside the width the
// textinput is given, so it renders wider than it is told. Measured from the
// prompt rather than written as a number, so the two cannot drift - and in one
// place, because the prompt is not the same width while a statement is being
// continued as it is at a fresh one.
//
// Clamped because v2's textinput sizes its placeholder buffer from the width
// and panics on a negative one.
func inputWidth(windowWidth int, prompt string) int {
	return max(windowWidth-lipgloss.Width(prompt)-1, 0)
}

// continueStatement puts the prompt into the state of waiting for the rest of a
// statement.
func (m *MainModel) continueStatement() {
	m.input.Placeholder = "end the statement with ;"
	m.setInputPrompt(continuationMark)
}

// endStatement puts it back to waiting for a new one.
func (m *MainModel) endStatement() {
	m.multiLineMode = false
	m.multiLineBuffer = nil
	m.input.Placeholder = "Enter CQL command..."
	m.setInputPrompt(promptMark)
}

// echoInput writes a line into the console as it was typed, under the prompt it
// was typed at.
//
// Only for the lines of a statement that is still being typed. A statement
// finished on one line is written out when it runs, which is also when anything
// it has to say arrives - and writing it twice would be two of it.
func (m *MainModel) echoInput(line string) {
	mark := promptMark
	if len(m.multiLineBuffer) > 1 {
		mark = continuationMark
	}

	m.fullHistoryContent += "\n" + m.styles.AccentText.Render(mark+line)
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()
}
