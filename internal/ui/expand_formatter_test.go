package ui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripAnsiForTest removes styling so the assertions look at the text.
func stripAnsiForTest(s string) string {
	var out strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func expandTestData(rows, cols int) [][]string {
	headers := make([]string, cols)
	for c := range cols {
		headers[c] = "col" + string(rune('a'+c))
	}
	data := [][]string{headers}
	for r := range rows {
		row := make([]string, cols)
		for c := range cols {
			row[c] = "r" + string(rune('0'+r)) + "c" + string(rune('a'+c))
		}
		data = append(data, row)
	}
	return data
}

// TestExpandBoundariesPointAtRecordStarts is the guard for #85. The boundaries
// have to describe the layout actually on screen: paging snaps to them, so if
// they come from a different renderer the tail of the result is unreachable.
func TestExpandBoundariesPointAtRecordStarts(t *testing.T) {
	const rows = 5

	out, boundaries := FormatExpandTableWithBoundaries(expandTestData(rows, 4), DefaultStyles())

	// One per record, plus a final one so paging can reach the end.
	require.Len(t, boundaries, rows+1)

	lines := strings.Split(stripAnsiForTest(out), "\n")

	for i := range rows {
		at := boundaries[i]
		require.Less(t, at, len(lines), "boundary %d is past the end of the output", i)
		assert.Equal(t, "@ Row "+string(rune('1'+i)), lines[at],
			"boundary %d should land on the start of record %d", i, i+1)
	}

	last := boundaries[rows]
	assert.Less(t, last, len(lines))
	assert.Greater(t, last, boundaries[rows-1],
		"the final boundary must be past the last record start, or the last record cannot be scrolled through")
}

// TestExpandBoundariesGrowWithRecordHeight is the actual bug in miniature: a
// record here is many lines tall, so the boundaries must be spaced accordingly
// rather than tracking the row count.
func TestExpandBoundariesGrowWithRecordHeight(t *testing.T) {
	const rows = 10

	narrow, narrowBounds := FormatExpandTableWithBoundaries(expandTestData(rows, 2), DefaultStyles())
	wide, wideBounds := FormatExpandTableWithBoundaries(expandTestData(rows, 20), DefaultStyles())

	narrowLines := len(strings.Split(narrow, "\n"))
	wideLines := len(strings.Split(wide, "\n"))
	require.Greater(t, wideLines, narrowLines*2, "more columns must mean taller records")

	// Both have the same number of records...
	require.Len(t, narrowBounds, rows+1)
	require.Len(t, wideBounds, rows+1)

	// ...but the wide one spreads them much further down the output. This is
	// what boundaries taken from the boxed table renderer cannot express.
	assert.Greater(t, wideBounds[rows-1], narrowBounds[rows-1]*2,
		"boundaries must follow the rendered height, not the row count")

	assert.Greater(t, wideBounds[rows-1], rows,
		"the last record must start well below line %d, or paging stops near the row count", rows)
}

func TestExpandBoundariesEmptyResults(t *testing.T) {
	out, boundaries := FormatExpandTableWithBoundaries(nil, DefaultStyles())
	assert.Equal(t, "No results", out)
	assert.Nil(t, boundaries)

	out, boundaries = FormatExpandTableWithBoundaries([][]string{{"a", "b"}}, DefaultStyles())
	assert.Equal(t, "No results", out)
	assert.Nil(t, boundaries)
}

// TestExpandPagingReachesTheLastRecord is the behaviour reported in #85: in
// expand format, paging down could not get past roughly the row count, leaving
// most of the result unreachable.
//
// Paging clamps to the viewport's line count and then snaps to the nearest row
// boundary at or below that. When the boundaries describe a different layout,
// the snap silently undoes the clamp.
func TestExpandPagingReachesTheLastRecord(t *testing.T) {
	const (
		rows          = 52 // as reported
		cols          = 8
		viewportLines = 20
	)

	content, boundaries := FormatExpandTableWithBoundaries(expandTestData(rows, cols), DefaultStyles())
	totalLines := len(strings.Split(content, "\n"))
	require.Greater(t, totalLines, rows*8, "each record should be several lines tall")

	newModel := func(b []int) *MainModel {
		m := &MainModel{
			viewMode:           "table",
			hasTable:           true,
			tableViewport:      viewport.New(viewport.WithWidth(120), viewport.WithHeight(viewportLines)),
			tableRowBoundaries: b,
		}
		m.tableViewport.SetContent(content)
		return m
	}

	pageToBottom := func(m *MainModel) int {
		last := -1
		for range 500 {
			m.handleHalfPageDown()
			if m.tableViewport.YOffset() == last {
				break // stuck
			}
			last = m.tableViewport.YOffset()
		}
		return m.tableViewport.YOffset()
	}

	maxOffset := totalLines - viewportLines

	// Paging walks all the way to the bottom, so every record is reachable.
	m := newModel(boundaries)
	got := pageToBottom(m)
	assert.Equal(t, maxOffset, got,
		"paging must reach the bottom of the result (line %d), stopped at %d", maxOffset, got)

	// And it lands on record starts on the way, which is the point of tracking
	// boundaries at all rather than scrolling by raw lines.
	m = newModel(boundaries)
	starts := make(map[int]bool, len(boundaries))
	for _, b := range boundaries {
		starts[b] = true
	}
	offsets := []int{}
	last := -1
	for range 500 {
		m.handleHalfPageDown()
		o := m.tableViewport.YOffset()
		if o == last {
			break
		}
		last = o
		offsets = append(offsets, o)
	}
	require.NotEmpty(t, offsets)
	for _, o := range offsets[:len(offsets)-1] { // the final stop is the clamp to the bottom
		assert.True(t, starts[o], "paging stopped at line %d, which is not the start of a record", o)
	}

	t.Logf("paged to the bottom in %d steps, landing on record starts", len(offsets))
}

// TestSnapDownToBoundaryAlwaysAdvances covers the second half of #85.
//
// Snapping used to take the last boundary at or before the target and nothing
// else. When a record is taller than the scroll amount there is no such
// boundary past the current position, so the snap returned where it started and
// scrolling wedged. Boxed-table boundaries are one line apart and hid this;
// expand records are ten or more lines apart and expose it.
func TestSnapDownToBoundaryAlwaysAdvances(t *testing.T) {
	// Records 11 lines apart, half a page is 10 lines: the target never reaches
	// the next boundary.
	boundaries := []int{1, 12, 23, 34, 45}

	tests := []struct {
		name            string
		current, target int
		want            int
	}{
		{"steps to the next record when none is in range", 1, 11, 12},
		{"lands on a record start when one is in range", 1, 30, 23},
		{"advances from the top", 0, 10, 1},
		{"keeps going past the middle", 23, 33, 34},
		{"scrolls freely past the last record", 45, 55, 55},
		{"does not move when the target is behind", 23, 20, 23},
		{"does not move when the target is level", 23, 23, 23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := snapDownToBoundary(tt.current, tt.target, 20, boundaries)
			assert.Equal(t, tt.want, got)
			if tt.target > tt.current {
				assert.Greater(t, got, tt.current, "a downward scroll must make progress")
			}
		})
	}
}

func TestSnapDownToBoundaryWithoutBoundaries(t *testing.T) {
	assert.Equal(t, 10, snapDownToBoundary(0, 10, 20, nil))
}

// TestScrollingThroughATallRecord covers the case where one record is taller
// than the screen, which happens in EXPAND as soon as a table has more columns
// than the terminal has rows.
//
// Snapping to record starts is right only while a record fits. When it does
// not, jumping to the next record start steps over every line that did not fit,
// and those lines cannot be reached any other way - the columns past the bottom
// of the screen simply never appear.
func TestScrollingThroughATallRecord(t *testing.T) {
	const viewportHeight = 20

	// One 60-line record, then another. Far taller than the screen.
	boundaries := []int{1, 61, 121}

	t.Run("half page scrolls within the record", func(t *testing.T) {
		got := snapDownToBoundary(1, 11, viewportHeight, boundaries)
		assert.Equal(t, 11, got,
			"a record taller than the screen must be scrolled through, not skipped")
		assert.NotEqual(t, 61, got, "jumping to the next record hides the rest of this one")
	})

	t.Run("line scroll advances one line", func(t *testing.T) {
		assert.Equal(t, 2, snapLineDown(1, viewportHeight, boundaries))
		assert.Equal(t, 31, snapLineDown(30, viewportHeight, boundaries))
	})

	t.Run("paging still reaches every line of the record", func(t *testing.T) {
		at := 1
		seen := map[int]bool{at: true}
		for range 200 {
			next := snapDownToBoundary(at, at+10, viewportHeight, boundaries)
			if next == at {
				break
			}
			at = next
			seen[at] = true
			if at >= 61 {
				break
			}
		}
		assert.GreaterOrEqual(t, at, 61, "should be able to page out of a tall record")
		assert.Greater(t, len(seen), 2,
			"a tall record should take several steps to read, not one jump")
	})

	t.Run("short records still snap to their start", func(t *testing.T) {
		short := []int{1, 12, 23, 34}
		assert.Equal(t, 12, snapDownToBoundary(1, 11, viewportHeight, short),
			"a record that fits should still be stepped over whole")
		assert.Equal(t, 12, snapLineDown(1, viewportHeight, short))
	})
}

// TestPageKeysScrollLikeTheHalfPageKey guards against the two page handlers
// drifting apart from the half-page one again.
//
// PgDn kept its own copy of the snapping logic and so kept the bug #85 fixed
// in the d key: with no boundary between the current offset and the target it
// returned the current offset, and the key did nothing at all. Records taller
// than the screen make that the normal case, which is why PgUp and PgDn stopped
// working in EXPAND while the arrows and the wheel carried on.
func TestPageKeysScrollLikeTheHalfPageKey(t *testing.T) {
	const viewportHeight = 20

	t.Run("tall records: page down moves within the record", func(t *testing.T) {
		boundaries := []int{1, 61, 121} // 60-line records

		got := snapDownToBoundary(1, 17, viewportHeight, boundaries)
		assert.Greater(t, got, 1, "PgDn must move; standing still is the bug")
		assert.Less(t, got, 61, "and must not skip the rest of the record")
	})

	t.Run("tall records: page up moves within the record", func(t *testing.T) {
		// One record 300 lines tall, far more than the screen holds.
		boundaries := []int{1, 301}

		// Deep inside it, snapping to the record start would jump 180 lines and
		// skip everything between - unreadable on the way back up.
		got := snapUpToBoundary(200, 184, viewportHeight, boundaries)
		assert.Equal(t, 184, got, "PgUp must scroll within a record taller than the screen")

		// Near the top of the record, the start is within a screen, so aligning
		// to it hides nothing: the lines we were on are still visible.
		got = snapUpToBoundary(15, 1, viewportHeight, boundaries)
		assert.Equal(t, 1, got, "close to the record start, aligning to it is safe")
	})

	t.Run("short records still align to record starts", func(t *testing.T) {
		boundaries := []int{1, 12, 23, 34}

		assert.Equal(t, 23, snapDownToBoundary(12, 25, viewportHeight, boundaries))
		assert.Equal(t, 12, snapUpToBoundary(23, 15, viewportHeight, boundaries))
	})

	t.Run("upward scrolling never moves down or past the top", func(t *testing.T) {
		boundaries := []int{1, 12, 23}

		assert.Equal(t, 23, snapUpToBoundary(23, 30, viewportHeight, boundaries),
			"a target below the current offset must not move it")
		assert.GreaterOrEqual(t, snapUpToBoundary(12, 0, viewportHeight, boundaries), 0)
	})
}
