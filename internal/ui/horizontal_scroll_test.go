package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/axonops/cqlai/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// altKey is an arrow with Alt held, which is how a wide table is scrolled.
func altKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Mod: tea.ModAlt}
}

// wideResults puts a result too wide for the screen into the Results view.
func wideResults(t *testing.T, format config.OutputFormat) *MainModel {
	t.Helper()

	m := queryResultModel(t)
	m.tableViewport.SetWidth(40)
	m.columnWidths = []int{8, 8}
	m.showQueryResult([][]string{
		{"id", "payload"},
		{"1", strings.Repeat("a", 300)},
		{"2", strings.Repeat("b", 300)},
	}, nil, format)
	return m
}

// scrollRight names the four ways of scrolling right, so each is covered.
var scrollRight = map[string]func(*MainModel){
	"Alt+Right": func(m *MainModel) { m.handleRightArrow(altKey(tea.KeyRight)) },
	"l":         func(m *MainModel) { m.handleHorizontalScrollRight() },
	">":         func(m *MainModel) { m.handlePageRightScroll() },
	"wheel":     func(m *MainModel) { m.handleMouseWheelRight() },
}

var scrollLeft = map[string]func(*MainModel){
	"Alt+Left": func(m *MainModel) { m.handleLeftArrow(altKey(tea.KeyLeft)) },
	"h":        func(m *MainModel) { m.handleHorizontalScrollLeft() },
	"<":        func(m *MainModel) { m.handlePageLeftScroll() },
	"wheel":    func(m *MainModel) { m.handleMouseWheelLeft() },
}

// TestScrollingSidewaysLeavesTextAlone, by every route.
//
// #115 fixed two of the four and left the other two rebuilding ASCII output as
// a boxed table - including Alt+arrows, which is what the README tells you to
// use.
func TestScrollingSidewaysLeavesTextAlone(t *testing.T) {
	for _, format := range []config.OutputFormat{config.OutputFormatASCII, config.OutputFormatJSON} {
		for name, scroll := range scrollRight {
			m := wideResults(t, format)
			before := m.tableViewport.GetContent()

			scroll(m)

			assert.Equal(t, before, m.tableViewport.GetContent(),
				"%s by %s rebuilt the content", format, name)
			assert.Positive(t, m.tableViewport.XOffset(),
				"%s by %s did not scroll", format, name)
		}
	}
}

// TestABoxedTableIsRebuiltWhenItScrolls, because it drops whole columns rather
// than cutting one down the middle.
func TestABoxedTableIsRebuiltWhenItScrolls(t *testing.T) {
	for name, scroll := range scrollRight {
		m := wideResults(t, config.OutputFormatTable)

		scroll(m)

		assert.Zero(t, m.tableViewport.XOffset(),
			"%s: a table scrolls by rebuilding, not by sliding", name)
		assert.Positive(t, m.horizontalOffset, "%s did not scroll", name)
	}
}

// TestEveryRouteStopsAtTheSamePlace.
func TestEveryRouteStopsAtTheSamePlace(t *testing.T) {
	limits := map[string]int{}
	for name, scroll := range scrollRight {
		m := wideResults(t, config.OutputFormatASCII)
		for range 200 {
			scroll(m)
		}
		limits[name] = m.horizontalOffset
	}

	var want int
	for _, got := range limits {
		if want == 0 {
			want = got
		}
		assert.Equal(t, want, got, "the routes stop in different places: %v", limits)
	}
	assert.Positive(t, want)
}

// TestScrollingBackReachesTheLeftEdge, by every route.
func TestScrollingBackReachesTheLeftEdge(t *testing.T) {
	for name, back := range scrollLeft {
		m := wideResults(t, config.OutputFormatASCII)
		for range 20 {
			m.handleHorizontalScrollRight()
		}
		require.Positive(t, m.horizontalOffset)

		for range 200 {
			back(m)
		}
		assert.Zero(t, m.horizontalOffset, "%s did not reach the left edge", name)
		assert.Zero(t, m.tableViewport.XOffset(), "%s left the viewport scrolled", name)
	}
}

// TestANewResultStartsAtTheLeftEdge. Resetting only horizontalOffset left the
// viewport's own offset behind, so ASCII output arrived already scrolled.
func TestANewResultStartsAtTheLeftEdge(t *testing.T) {
	m := wideResults(t, config.OutputFormatASCII)
	for range 20 {
		m.handleHorizontalScrollRight()
	}
	require.Positive(t, m.tableViewport.XOffset())

	m.showQueryResult([][]string{{"id"}, {"1"}}, nil, config.OutputFormatASCII)

	assert.Zero(t, m.horizontalOffset)
	assert.Zero(t, m.tableViewport.XOffset())
}

// TestNarrowContentDoesNotScroll: there is nothing out there to see.
func TestNarrowContentDoesNotScroll(t *testing.T) {
	m := queryResultModel(t)
	m.tableViewport.SetWidth(80)
	m.showQueryResult([][]string{{"id"}, {"1"}}, nil, config.OutputFormatASCII)

	for name, scroll := range scrollRight {
		scroll(m)
		assert.Zero(t, m.horizontalOffset, "%s scrolled past the end of the content", name)
	}
}
