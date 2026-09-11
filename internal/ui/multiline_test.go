package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// promptModelFor is a console at the prompt, sized like a real terminal.
func multiLineModel(t *testing.T) *MainModel {
	t.Helper()

	m := &MainModel{
		styles:          DefaultStyles(),
		viewMode:        "history",
		windowWidth:     100,
		windowHeight:    24,
		ready:           true,
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(18)),
	}
	m.input.Focus()
	m.setInputPrompt(promptMark)
	return m
}

// typeLine types a line at the prompt and presses Enter.
func typeLine(m *MainModel, line string) *MainModel {
	m.input.SetValue(line)
	updated, _ := m.handleEnterKey()
	return updated
}

func consoleLines(m *MainModel) []string {
	return strings.Split(stripAnsiForTest(m.fullHistoryContent), "\n")
}

// TestAnUnfinishedStatementGoesIntoTheConsole.
//
// The lines were drawn above the prompt instead, which is a block that grows:
// the window has one row for the input, so the second line pushed the
// connection bar off the bottom and the third took the query line with it.
func TestAnUnfinishedStatementGoesIntoTheConsole(t *testing.T) {
	m := multiLineModel(t)

	m = typeLine(m, "SELECT *")
	require.True(t, m.multiLineMode)
	assert.Contains(t, consoleLines(m), "> SELECT *")
	assert.Empty(t, m.input.Value(), "the prompt should be clear for the next line")

	m = typeLine(m, "FROM users")
	assert.Contains(t, consoleLines(m), "... FROM users")

	// Nothing is drawn above the prompt: the view is the tabs, the content, the
	// two bars and one row of input, whatever is being typed.
	assert.NotContains(t, m.View().Content, "\n... FROM users\n> ")
}

// TestTheContinuedStatementIsWrittenOnce.
//
// Each line is in the console as it was typed, so writing the whole statement
// out again when it runs would be two of it.
func TestTheContinuedStatementIsWrittenOnce(t *testing.T) {
	m := multiLineModel(t)
	m = typeLine(m, "SELECT *")
	m = typeLine(m, "FROM users;")

	assert.False(t, m.multiLineMode, "the semicolon ends it")
	assert.Equal(t, "SELECT * FROM users;", m.lastCommand)

	console := stripAnsiForTest(m.fullHistoryContent)
	assert.Equal(t, 1, strings.Count(console, "SELECT *"), "written twice: %q", console)
	assert.NotContains(t, console, "> SELECT * FROM users;")
}

// TestThePromptSaysWhichLineYouAreOn.
func TestThePromptSaysWhichLineYouAreOn(t *testing.T) {
	m := multiLineModel(t)
	require.Contains(t, m.input.Prompt, promptMark)

	m = typeLine(m, "SELECT *")
	assert.Contains(t, m.input.Prompt, continuationMark, "a continued statement says so")

	m = typeLine(m, "FROM users;")
	assert.Contains(t, m.input.Prompt, promptMark)
	assert.NotContains(t, m.input.Prompt, continuationMark)
}

// TestTheBoxFitsBesideWhicheverPromptIsInFront.
//
// The prompt is drawn outside the width the textinput is given, so a wider one
// has to take that width off the box, or the row runs past the terminal and
// wraps onto the next.
func TestTheBoxFitsBesideWhicheverPromptIsInFront(t *testing.T) {
	m := multiLineModel(t)
	atPrompt := m.input.Width()

	m = typeLine(m, "SELECT *")
	assert.Equal(t, atPrompt-2, m.input.Width(), "... is two columns wider than >")

	m = typeLine(m, "FROM users;")
	assert.Equal(t, atPrompt, m.input.Width())
}

// TestEscapeGivesTheStatementUpAndThePromptBack.
func TestEscapeGivesTheStatementUpAndThePromptBack(t *testing.T) {
	m := multiLineModel(t)
	m = typeLine(m, "SELECT *")

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.False(t, m.multiLineMode)
	assert.Empty(t, m.multiLineBuffer)
	assert.Contains(t, m.input.Prompt, promptMark)
	assert.Contains(t, stripAnsiForTest(m.fullHistoryContent), "Multi-line mode cancelled")
}

// TestTheBarsStayPutHoweverManyLinesAreTyped, which is the bug: the layout is
// the tabs, the content, two bars and one row of input, and a statement being
// typed cannot be allowed to grow into them.
func TestTheBarsStayPutHoweverManyLinesAreTyped(t *testing.T) {
	m := multiLineModel(t)
	m = typeLine(m, "SELECT")
	require.True(t, m.multiLineMode)

	height := len(strings.Split(m.View().Content, "\n"))
	bars := stripAnsiForTest(m.View().Content)
	require.Contains(t, bars, "Connection")
	require.Contains(t, bars, "History:")

	for _, line := range []string{"id,", "name,", "created_at", "FROM users", "WHERE id = 1"} {
		m = typeLine(m, line)
	}

	assert.Len(t, strings.Split(m.View().Content, "\n"), height,
		"the view grew with the statement being typed")

	// And both bars are still on it: they were pushed off the bottom a line at
	// a time, the connection bar first and the query line after it.
	bars = stripAnsiForTest(m.View().Content)
	assert.Contains(t, bars, "Connection")
	assert.Contains(t, bars, "History:")
}

// TestThePromptRowIsNotSharedWithTheMoreRowsHint.
//
// The hint was written over the input's placeholder, so with a partly loaded
// result on screen it landed on the row a statement was being typed on: the
// continuation prompt and "MORE DATA AVAILABLE" side by side, and no sign of
// what ends the statement.
func TestThePromptRowIsNotSharedWithTheMoreRowsHint(t *testing.T) {
	m := multiLineModel(t)
	m.viewMode = "table"
	m.hasTable = true
	m.tableViewport = viewport.New(viewport.WithWidth(100), viewport.WithHeight(14))
	m.session = &db.Session{}
	m.slidingWindow = NewSlidingWindowTable(1000, 10)
	m.slidingWindow.Headers = []string{"id"}
	m.slidingWindow.hasMoreData = true

	m = typeLine(m, "SELECT *")

	rows := strings.Split(stripAnsiForTest(m.View().Content), "\n")
	prompt := ""
	for _, row := range rows {
		if strings.HasPrefix(row, continuationMark) {
			prompt = row
		}
	}
	require.NotEmpty(t, prompt, "the continuation prompt should be on screen")

	assert.Contains(t, prompt, "end the statement with ;")
	assert.NotContains(t, prompt, "MORE DATA")
	assert.NotContains(t, prompt, "load more")
}

// TestSpaceStartsALineRatherThanPagingWhileAStatementIsBeingTyped.
//
// Space pages a partly loaded result when the prompt is empty - and the prompt
// is empty at the start of every continuation line, so a statement could not
// begin a line with a space without turning the page instead.
func TestSpaceStartsALineRatherThanPagingWhileAStatementIsBeingTyped(t *testing.T) {
	m := multiLineModel(t)
	m.viewMode = "table"
	m.hasTable = true
	m.tableViewport = viewport.New(viewport.WithWidth(100), viewport.WithHeight(14))
	m.session = &db.Session{}
	m.slidingWindow = NewSlidingWindowTable(1000, 10)
	m.slidingWindow.Headers = []string{"id"}
	m.slidingWindow.hasMoreData = true
	m.slidingWindow.AddRow([]string{"1"}, nil)

	m = typeLine(m, "SELECT *")
	rows := len(m.slidingWindow.Rows)

	m, _ = m.handleKeyboardInput(tea.KeyPressMsg{Code: ' ', Text: " "})

	assert.Equal(t, " ", m.input.Value(), "the space belongs to the statement")
	assert.Len(t, m.slidingWindow.Rows, rows, "and should not have turned the page")
}

// TestTheConsoleFillsFromTheBottom.
//
// A viewport draws its content from the top, so a session that had said little
// so far left the welcome message at the top of the screen and the command you
// had just run below it, with blank space between that and the prompt it was
// typed at. The answer belongs where the question was asked.
func TestTheConsoleFillsFromTheBottom(t *testing.T) {
	m := multiLineModel(t)
	m.historyViewport.SetHeight(8)
	m.fullHistoryContent = "Welcome to cqlai"
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	m.fullHistoryContent += "\n" + promptMark + "HELP\nsome answer"
	m.updateHistoryWrapping()
	m.historyViewport.GotoBottom()

	rows := strings.Split(stripAnsiForTest(m.historyViewport.View()), "\n")
	require.Len(t, rows, 8)

	// The last thing said is on the last row, against the prompt.
	assert.Equal(t, "some answer", strings.TrimRight(rows[len(rows)-1], " "))
	assert.Contains(t, rows[len(rows)-2], "> HELP")
	assert.Contains(t, rows[len(rows)-3], "Welcome to cqlai")

	// And what is above it is empty rather than the transcript.
	for _, row := range rows[:len(rows)-3] {
		assert.Empty(t, strings.TrimSpace(row))
	}
}

// TestALongTranscriptIsNotPadded: once it fills the view there is nothing to
// push down, and padding it would scroll the top of it off for nothing.
func TestALongTranscriptIsNotPadded(t *testing.T) {
	m := multiLineModel(t)
	m.historyViewport.SetHeight(4)
	m.fullHistoryContent = strings.TrimSuffix(strings.Repeat("line\n", 20), "\n")
	m.updateHistoryWrapping()

	assert.Equal(t, 20, len(strings.Split(m.historyViewport.GetContent(), "\n")))
}
