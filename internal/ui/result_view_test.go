package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/viewport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
)

// shownModel is a result on screen, drawn in the given format, with the router
// pointed at the same session manager the model holds - which is what decides
// what OUTPUT does.
func shownModel(t *testing.T, format config.OutputFormat) *MainModel {
	t.Helper()

	manager := session.NewManager(&config.Config{})
	require.NoError(t, manager.SetOutputFormat(format))
	router.InitRouter(manager)
	t.Cleanup(func() { router.InitRouter(nil) })

	m := &MainModel{
		styles:          DefaultStyles(),
		sessionManager:  manager,
		windowWidth:     100,
		windowHeight:    30,
		viewMode:        "history",
		input:           newTestInput(),
		historyViewport: viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		tableViewport:   viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}
	m.showQueryResult([][]string{
		{"id", "name"},
		{"1", "alice"},
		{"2", "bob"},
	}, []string{"int", "text"})
	return m
}

func shown(m *MainModel) string {
	return stripAnsiForTest(m.tableViewport.GetContent())
}

// TestSwitchingOutputRedrawsWhatIsOnScreen, which is the point: the rows are
// already in hand, so looking at them another way is a redraw rather than a
// reason to ask Cassandra the same question again.
func TestSwitchingOutputRedrawsWhatIsOnScreen(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)
	require.Contains(t, shown(m), "│", "the boxed table should be boxed")

	m.processCommand("OUTPUT JSON")

	assert.Equal(t, config.OutputFormatJSON, m.resultFormat)
	assert.Contains(t, shown(m), `{"id":"1","name":"alice"}`)
	assert.NotContains(t, shown(m), "│")

	// The same rows throughout - nothing was re-queried, and nothing was lost.
	for _, value := range []string{"alice", "bob"} {
		assert.Contains(t, shown(m), value)
	}
}

// TestEveryFormatCanBeSwitchedToAndBack.
func TestEveryFormatCanBeSwitchedToAndBack(t *testing.T) {
	formats := config.OutputFormats()

	for _, from := range formats {
		for _, to := range formats {
			parsedTo, err := config.ParseOutputFormat(to)
			require.NoError(t, err)
			parsedFrom, err := config.ParseOutputFormat(from)
			require.NoError(t, err)

			m := shownModel(t, parsedFrom)
			m.processCommand("OUTPUT " + to)

			assert.Equal(t, parsedTo, m.resultFormat, "%s to %s", from, to)
			assert.Contains(t, shown(m), "alice", "%s to %s: the rows should still be there", from, to)
			assert.Positive(t, m.tableWidth, "%s to %s: scrolling sideways needs a width", from, to)
		}
	}
}

// TestSwitchingOutputMovesTheWidthWithIt.
//
// The width is what sideways scrolling clamps against, and it was only ever set
// when a result first arrived. JSON lines are far wider than the boxed table of
// the same rows, so a switch left it describing the wrong thing.
func TestSwitchingOutputMovesTheWidthWithIt(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)
	boxed := m.tableWidth

	m.processCommand("OUTPUT JSON")
	assert.NotEqual(t, boxed, m.tableWidth)
	assert.Equal(t, widestLine(shown(m)), m.tableWidth)

	m.processCommand("OUTPUT TABLE")
	assert.Equal(t, boxed, m.tableWidth, "back where it started")
}

// TestSwitchingOutputStartsAtTheTopLeft: row 500 of a boxed table is not line
// 500 of JSON, and column 40 is not column 40.
func TestSwitchingOutputStartsAtTheTopLeft(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)
	m.horizontalOffset = 20
	m.tableViewport.SetYOffset(2)

	m.processCommand("OUTPUT ASCII")

	assert.Zero(t, m.horizontalOffset)
	assert.Zero(t, m.tableViewport.XOffset())
	assert.Zero(t, m.tableViewport.YOffset())
}

// TestExpandOnAndOffRedrawTheResultToo.
//
// The redraw comes from the format changing rather than from the command
// saying OUTPUT, so EXPAND - which sets the same setting by another name - gets
// it without knowing anything about the view.
func TestExpandOnAndOffRedrawTheResultToo(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)

	m.processCommand("EXPAND ON")
	assert.Equal(t, config.OutputFormatExpand, m.resultFormat)
	assert.Contains(t, shown(m), "@ Row 1")

	m.processCommand("EXPAND OFF")
	assert.Equal(t, config.OutputFormatTable, m.resultFormat)
	assert.NotContains(t, shown(m), "@ Row 1")
}

// TestACommandThatChangesNothingRedrawsNothing: a redraw rebuilds the table and
// throws the scroll position away, so it has to be the format changing that
// causes one, not any command at all.
func TestACommandThatChangesNothingRedrawsNothing(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)

	// Somewhere to scroll to: a viewport shorter than the result.
	m.tableViewport.SetHeight(3)
	m.tableViewport.SetYOffset(2)
	require.Equal(t, 2, m.tableViewport.YOffset())

	m.processCommand("OUTPUT") // asks what the format is
	assert.Equal(t, 2, m.tableViewport.YOffset())

	m.processCommand("OUTPUT TABLE") // already TABLE
	assert.Equal(t, 2, m.tableViewport.YOffset())
}

// TestSwitchingOutputWithNothingOnScreen still sets the format for next time.
func TestSwitchingOutputWithNothingOnScreen(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)
	m.hasTable = false
	m.lastTableData = nil
	m.tableViewport.SetContent("")

	m.processCommand("OUTPUT JSON")

	assert.Equal(t, config.OutputFormatJSON, m.outputFormat())
	assert.Empty(t, shown(m))
}

// TestADescribeGridFollowsTheFormatToo.
//
// These grids - DESCRIBE TABLES and the listings - used to force the boxed
// table whatever OUTPUT said. That was a rule nothing else followed, and one a
// redraw would have broken anyway: drawn one way, redrawn another.
func TestADescribeGridFollowsTheFormatToo(t *testing.T) {
	m := shownModel(t, config.OutputFormatJSON)
	m.processTableResult("DESCRIBE TABLES", [][]string{{"keyspace_name"}, {"system"}})

	assert.Equal(t, config.OutputFormatJSON, m.resultFormat)
	assert.Contains(t, shown(m), `{"keyspace_name":"system"}`)

	m.processCommand("OUTPUT TABLE")
	assert.Contains(t, shown(m), "│")
	assert.Contains(t, shown(m), "system")
}

// TestMoreRowsArrivingMovesTheWidthWithThem.
//
// The width is set from the content, so it has to be set again when the content
// changes. It was only ever set when a result first arrived: a long value in a
// later page widened the text and left the sideways scrolling clamped to the
// first page's width, with the end of the line out of reach.
func TestMoreRowsArrivingMovesTheWidthWithThem(t *testing.T) {
	m := outputModel(t, false, config.OutputFormatASCII)
	m.displayResult(m.slidingWindow.Headers, nil)
	first := m.tableWidth

	// The second page holds a value far wider than anything on the first.
	wide := []string{"3", strings.Repeat("c", 200)}
	m.slidingWindow.streamingResult.LoadMore = func(_ context.Context, _ int) (db.Page, error) {
		return db.Page{
			Rows: [][]string{wide},
			Raw:  []map[string]interface{}{rowValues(m.slidingWindow.Headers, wide)},
		}, nil
	}
	m.loadMoreTableDataHelper()

	require.Contains(t, stripAnsiForTest(m.tableViewport.GetContent()), "ccc")
	assert.Greater(t, m.tableWidth, first, "the width should have grown with the content")
	assert.Equal(t, widestLine(stripAnsiForTest(m.tableViewport.GetContent())), m.tableWidth)
}

// TestTheNoticeFollowsTheRowsInHand: it is drawn with the content, so it says
// what is true at the time rather than what was true when the result arrived.
func TestTheNoticeFollowsTheRowsInHand(t *testing.T) {
	m := outputModel(t, false, config.OutputFormatTable)
	m.displayResult(m.slidingWindow.Headers, nil)

	// A boxed table pages in as you scroll, and the scrollbar says as much.
	assert.NotContains(t, shown(m), "More data available")

	// ASCII is drawn in one pass, so what is on screen can be a fraction of the
	// answer with nothing else to say so.
	m.processCommand("OUTPUT ASCII")
	assert.Contains(t, shown(m), "More data available")
	assert.Contains(t, shown(m), "Showing 2 rows")

	// And it goes when the rest arrives.
	m.loadMoreTableDataHelper()
	assert.NotContains(t, shown(m), "More data available")
	assert.Contains(t, shown(m), "dave")
}

// TestSwitchingFormatFetchesNothingExtra.
//
// AUTOFETCH decides how much is pulled in. A format switch is a redraw of the
// rows in hand, not a reason to read the rest of a table.
func TestSwitchingFormatFetchesNothingExtra(t *testing.T) {
	m := outputModel(t, false, config.OutputFormatTable)
	m.displayResult(m.slidingWindow.Headers, nil)
	require.Len(t, m.slidingWindow.Rows, 2)

	m.processCommand("OUTPUT JSON")

	assert.Len(t, m.slidingWindow.Rows, 2, "switching format fetched more rows")
	assert.True(t, m.slidingWindow.hasMoreData)
}

// TestAJSONResultGoesBackToBeingATable.
//
// OUTPUT JSON used to rewrite the query as SELECT JSON, so what came back was
// one column of documents rather than the columns asked for. Switching to TABLE
// then drew that one column: the JSON, in a table, with the columns nowhere to
// be found. The format decides how a result is drawn and nothing else, so the
// columns are still there to go back to.
func TestAJSONResultGoesBackToBeingATable(t *testing.T) {
	m := shownModel(t, config.OutputFormatJSON)
	require.Contains(t, shown(m), `"name":"alice"`)
	require.NotContains(t, shown(m), "│")

	m.processCommand("OUTPUT TABLE")

	drawn := shown(m)
	assert.Contains(t, drawn, "│", "the result should be a table again")
	assert.NotContains(t, drawn, `{"id"`, "the table is drawn from the values, not from the JSON")

	// Both columns, in their own right.
	for _, column := range []string{"id", "name"} {
		assert.Contains(t, drawn, column)
	}
	assert.Contains(t, drawn, "alice")
	assert.Contains(t, drawn, "bob")
}

// TestJSONKeepsTheTypesOfTheValues, which is what a JSON document is for.
//
// The values are kept beside the strings drawn in the cells, so a number is a
// number rather than the text of one - which is all that would be left if JSON
// were built from what is on screen.
func TestJSONKeepsTheTypesOfTheValues(t *testing.T) {
	m := shownModel(t, config.OutputFormatTable)
	m.slidingWindow = NewSlidingWindowTable(1000, 10)
	m.slidingWindow.Headers = []string{"id", "name", "active", "score", "when"}
	m.slidingWindow.AddRow(
		[]string{"7", "alice", "true", "1.5", "2026-09-11 10:00:00"},
		map[string]interface{}{
			"id":     7,
			"name":   "alice",
			"active": true,
			"score":  1.5,
			"when":   time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC),
		})
	m.tableHeaders = m.slidingWindow.Headers
	m.hasTable = true
	m.lastTableData = m.resultRows()

	m.processCommand("OUTPUT JSON")

	line := strings.TrimSpace(shown(m))
	assert.Contains(t, line, `"id":7`, "a number should not be quoted")
	assert.Contains(t, line, `"active":true`, "a boolean should not be quoted")
	assert.Contains(t, line, `"score":1.5`)
	assert.Contains(t, line, `"name":"alice"`)
	assert.Contains(t, line, `"when":"2026-09-11T10:00:00Z"`)

	// And the columns come out in the order they were asked for.
	assert.Less(t, strings.Index(line, `"id"`), strings.Index(line, `"name"`))
}

// TestJSONFallsBackToWhatIsDrawn, for a grid that never had values behind it:
// DESCRIBE and the listings are made up by the shell rather than read from a
// table.
func TestJSONFallsBackToWhatIsDrawn(t *testing.T) {
	m := shownModel(t, config.OutputFormatJSON)
	m.processTableResult("DESCRIBE TABLES", [][]string{{"keyspace_name"}, {"system"}})

	assert.Contains(t, shown(m), `{"keyspace_name":"system"}`)
}

// TestAnExplicitSelectJSONIsLeftAlone: asking Cassandra for JSON is a different
// thing from asking cqlai to draw JSON, and it still comes back as one column
// of documents.
func TestAnExplicitSelectJSONIsLeftAlone(t *testing.T) {
	m := shownModel(t, config.OutputFormatJSON)
	m.showQueryResult([][]string{{"[json]"}, {`{"id": 1, "name": "alice"}`}}, nil)

	assert.Equal(t, `{"id": 1, "name": "alice"}`, strings.TrimSpace(shown(m)))

	// And drawn as a table it is the column it is: one document per row.
	m.processCommand("OUTPUT TABLE")
	assert.Contains(t, shown(m), "[json]")
	assert.Contains(t, shown(m), `{"id": 1`)
}
