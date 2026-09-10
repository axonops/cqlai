package ui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wideTrace is a trace far wider than the screen, which is what a real one is:
// the activity column alone runs past eighty columns.
func wideTrace(t *testing.T) *MainModel {
	t.Helper()

	rows := [][]string{{"activity", "timestamp", "source", "source_elapsed", "client", "thread"}}
	for i := range 5 {
		rows = append(rows, []string{
			fmt.Sprintf("Parsing SELECT * FROM my_keyspace.users WHERE id = %d LIMIT 100", i),
			"2026-09-10 17:20:00.000", "10.0.0.1", "1234", "10.0.0.9", "Native-Transport-Requests-1",
		})
	}

	m := &MainModel{
		styles:        DefaultStyles(),
		viewMode:      "trace",
		hasTrace:      true,
		traceData:     rows,
		traceHeaders:  rows[0],
		traceViewport: viewport.New(viewport.WithWidth(80), viewport.WithHeight(10)),
		tableViewport: viewport.New(viewport.WithWidth(80), viewport.WithHeight(10)),
		windowWidth:   80,
		windowHeight:  24,
		input:         newTestInput(),
	}
	m.refreshTraceView()

	require.Greater(t, m.traceTableWidth, m.traceViewport.Width(),
		"the fixture has to be wider than the screen or there is nothing to test")
	return m
}

// TestEveryWayOfScrollingATraceSideways.
//
// The keyboard worked and the mouse did not: scrollHorizontally, which the
// wheel calls, returned without doing anything for anything but the Results
// view, while the trace had four hand-rolled copies of the same clamp that the
// keys reached instead.
func TestEveryWayOfScrollingATraceSideways(t *testing.T) {
	for name, scroll := range map[string]func(*MainModel){
		"Alt+Right": func(m *MainModel) {
			m.handleRightArrow(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModAlt})
		},
		"Alt+PageDown":         func(m *MainModel) { m.handlePageRightScroll() },
		"> in navigation mode": func(m *MainModel) { m.handleHorizontalScrollRight() },
		"the wheel sideways":   func(m *MainModel) { m.handleMouseWheelRight() },
		"a modifier and the wheel": func(m *MainModel) {
			m.handleMouseWheel(tea.Mouse{Button: tea.MouseWheelDown, Mod: tea.ModShift})
		},
	} {
		m := wideTrace(t)
		before := traceRow(m)

		scroll(m)

		assert.Positive(t, m.traceHorizontalOffset, "%s should move it", name)
		assert.NotEqual(t, before, traceRow(m), "%s should change what is drawn", name)
	}
}

// TestAModifierAndTheWheelDoesNotScrollATraceUpAndDown.
//
// The gate named the Results view, so on a trace it fell through to the
// vertical wheel - on the one view where reaching sideways matters most.
func TestAModifierAndTheWheelDoesNotScrollATraceUpAndDown(t *testing.T) {
	m := wideTrace(t)
	before := m.traceViewport.YOffset()

	m.handleMouseWheel(tea.Mouse{Button: tea.MouseWheelDown, Mod: tea.ModShift})

	assert.Equal(t, before, m.traceViewport.YOffset(), "it should not have moved down")
	assert.Positive(t, m.traceHorizontalOffset, "it should have moved across")
}

// TestScrollingATraceBackToTheLeftStopsAtTheEdge.
func TestScrollingATraceBackToTheLeftStopsAtTheEdge(t *testing.T) {
	m := wideTrace(t)
	leftmost := traceRow(m)

	for range 5 {
		m.handleHorizontalScrollRight()
	}
	require.Positive(t, m.traceHorizontalOffset)

	for range 20 {
		m.handleHorizontalScrollLeft()
	}

	assert.Zero(t, m.traceHorizontalOffset, "it should stop at the left-hand edge")
	assert.Equal(t, leftmost, traceRow(m), "and be showing what it started with")
}

// TestATraceStopsAtItsRightHandEdge rather than scrolling into empty space.
func TestATraceStopsAtItsRightHandEdge(t *testing.T) {
	m := wideTrace(t)

	for range 200 {
		m.handleHorizontalScrollRight()
	}

	assert.LessOrEqual(t, m.traceHorizontalOffset,
		m.traceTableWidth-m.traceViewport.Width()+horizontalStep)
	assert.NotEmpty(t, strings.TrimSpace(traceRow(m)), "there should still be something drawn")
}

// TestANarrowTraceDoesNotScroll, because there is nothing off the edge.
func TestANarrowTraceDoesNotScroll(t *testing.T) {
	m := wideTrace(t)
	m.traceViewport.SetWidth(m.traceTableWidth + horizontalStep)

	m.handleHorizontalScrollRight()

	assert.Zero(t, m.traceHorizontalOffset)
}

// traceRow is the first row of the trace as drawn, for telling whether the view
// actually moved rather than only the number behind it.
func traceRow(m *MainModel) string {
	rows := strings.Split(stripAnsiForTest(m.traceViewport.View()), "\n")
	if len(rows) < 2 {
		return ""
	}
	return rows[1]
}

// TestATableIsMeasuredInColumnsNotBytes.
//
// A boxed table is mostly box-drawing characters and each of those is three
// bytes, so len made the table about two and a half times as wide as it is.
// Everything that reads the width believed it: scrolling ran hundreds of
// columns past the end into blank screen, and the Results view had it too.
func TestATableIsMeasuredInColumnsNotBytes(t *testing.T) {
	m := wideTrace(t)

	// The table as it is built, before the viewport clips it.
	m.horizontalOffset = 0
	m.cachedTableLines = nil
	m.initialColumnWidths = nil
	drawn := m.formatTableForViewport(m.traceData)

	want := 0
	bytes := 0
	for _, line := range strings.Split(drawn, "\n") {
		want = max(want, lipgloss.Width(line))
		bytes = max(bytes, len(stripAnsiForTest(line)))
	}

	require.Greater(t, bytes, want, "the fixture has to contain box drawing to be a test")
	assert.Equal(t, want, m.tableWidth, "the width is what it takes on screen")
	assert.NotEqual(t, bytes, m.tableWidth, "not what it takes in memory")
}

// TestScrollingRightLeavesSomethingOnScreen.
//
// With the width in bytes there was room to scroll a hundred columns past the
// last one, and the view went blank rather than stopping at the edge.
func TestScrollingRightLeavesSomethingOnScreen(t *testing.T) {
	m := wideTrace(t)

	for range 200 {
		m.handleHorizontalScrollRight()
	}

	assert.Contains(t, traceRow(m), "thread", "the last column should still be there")
}
