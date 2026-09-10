package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// typedTable is a result with a composite partition key, a clustering column
// and two plain columns.
func typedTable() ([][]string, []string) {
	data := [][]string{
		{"id (PK1)", "region (PK2)", "created (C)", "name"},
		{"1", "eu-west", "2026-09-10T11:22:33Z", "alice"},
		{"2", "us-east", "2026-09-09T08:00:00Z", "bob"},
	}
	return data, []string{"uuid", "text", "timestamp", "varchar"}
}

// rows returns the rendered table split into lines, ANSI stripped.
func rows(rendered string) []string {
	return strings.Split(stripAnsi(rendered), "\n")
}

// TestTheDetailRowSaysWhatEachColumnIs.
//
// The type used to sit beside the name behind F6, widening every column by the
// length of its type name, which is why it was off by default and almost never
// seen. On its own row it costs one row for the whole table.
func TestTheDetailRowSaysWhatEachColumnIs(t *testing.T) {
	m := &MainModel{styles: DefaultStyles()}
	data, types := typedTable()
	m.columnTypes = types

	lines := rows(m.formatTableForViewport(data))
	require.GreaterOrEqual(t, len(lines), 4)

	names, details := lines[1], lines[2]

	for _, name := range []string{"id", "region", "created", "name"} {
		assert.Contains(t, names, name)
	}
	assert.NotContains(t, names, "PK", "the marker moved to the row below")
	assert.NotContains(t, names, "(C)")

	for _, detail := range []string{"PK1 uuid", "PK2 text", "C timestamp", "varchar"} {
		assert.Contains(t, details, detail)
	}
}

// TestAResultWithNoTypesHasNoDetailRow.
//
// Every column of a Cassandra table has a type, and a query carries one per
// column. What arrives without them is a grid the shell made up rather than a
// result over a table - DESCRIBE and the listings, whose "Primary Key" and "GC
// Grace" are not columns of anything.
func TestAResultWithNoTypesHasNoDetailRow(t *testing.T) {
	m := &MainModel{styles: DefaultStyles()}
	listing := [][]string{
		{"Table", "Primary Key", "GC Grace"},
		{"users", "id", "864000"},
	}

	lines := rows(m.formatTableForViewport(listing))

	assert.Contains(t, lines[1], "Primary Key")
	assert.Contains(t, lines[2], "─", "the separator follows the names directly")
	assert.Equal(t, 3, m.headerRowCount(listing[0]))
}

// TestATypeTooLongForItsNameWidensTheColumn, and nothing else does.
func TestATypeTooLongForItsNameWidensTheColumn(t *testing.T) {
	m := &MainModel{styles: DefaultStyles()}
	data := [][]string{{"n"}, {"1"}}
	m.columnTypes = []string{"timestamp"}

	lines := rows(m.formatTableForViewport(data))

	assert.Contains(t, lines[2], "timestamp", "the detail is not truncated")
	for _, line := range lines {
		assert.Equal(t, len([]rune(lines[0])), len([]rune(line)),
			"every row is the width the widest of them needs: %q", line)
	}
}

// TestTheFrozenHeaderIsTheSameBlock.
//
// The body draws the header and the frozen header replaces exactly those rows
// when the table is scrolled. Drawn two different ways, they put columns on
// screen that nothing below them lines up with - which is why they share one
// function now rather than being two copies of the layout.
func TestTheFrozenHeaderIsTheSameBlock(t *testing.T) {
	m := &MainModel{styles: DefaultStyles()}
	data, types := typedTable()
	m.tableHeaders, m.columnTypes = data[0], types

	body := rows(m.formatTableForViewport(data))
	frozen := rows(m.buildTableStickyHeader())

	require.Len(t, frozen, m.headerRowCount(m.tableHeaders))
	assert.Equal(t, body[:len(frozen)], frozen)
}

// TestTheFrozenHeaderCoversAsManyRowsAsItDraws. The selection reads the same
// count to know which rows on screen are the header rather than the content
// scrolled under it, so a wrong count hides or reveals a row of the result.
func TestTheFrozenHeaderCoversAsManyRowsAsItDraws(t *testing.T) {
	data, types := typedTable()

	withTypes := &MainModel{styles: DefaultStyles(), tableHeaders: data[0], columnTypes: types}
	withTypes.formatTableForViewport(data) // sets columnWidths
	assert.Equal(t, 4, withTypes.headerRowCount(withTypes.tableHeaders))
	assert.Len(t, rows(withTypes.buildTableStickyHeader()), 4)

	without := &MainModel{styles: DefaultStyles(), tableHeaders: data[0]}
	without.formatTableForViewport(data)
	assert.Equal(t, 3, without.headerRowCount(without.tableHeaders))
	assert.Len(t, rows(without.buildTableStickyHeader()), 3)
}

// TestTheDetailIsBuiltFromTheMarkerAndTheType.
func TestTheDetailIsBuiltFromTheMarkerAndTheType(t *testing.T) {
	assert.Equal(t, "PK uuid", headerDetail("id (PK)", "uuid"))
	assert.Equal(t, "PK2 text", headerDetail("region (PK2)", "text"))
	assert.Equal(t, "C1 timestamp", headerDetail("created (C1)", "timestamp"))
	assert.Equal(t, "varchar", headerDetail("name", "varchar"), "no marker, just the type")
	assert.Equal(t, "PK", headerDetail("id (PK)", ""), "no type, just the marker")
	assert.Empty(t, headerDetail("name", ""))
}

// TestTheTypeNeverReachesTheData.
//
// F6 built the type into m.lastTableData[0] - "id [int] (PK)" - and that is
// what SAVE and CAPTURE write. StripKeyMarker takes off the " (PK)" and knows
// nothing about " [int]", so a saved CSV had a header of "id [int]" and a
// Parquet file a field of that name: an export that could not be read back into
// the table it came from.
//
// Drawing the type instead of storing it means there is nothing to strip.
func TestTheTypeNeverReachesTheData(t *testing.T) {
	m := &MainModel{styles: DefaultStyles()}
	data, types := typedTable()
	m.columnTypes = types

	m.formatTableForViewport(data)

	for _, header := range data[0] {
		for _, cqlType := range types {
			assert.NotContains(t, header, cqlType,
				"the type is drawn, not written into the header")
		}
	}
	assert.Equal(t, "id (PK1)", data[0][0], "the header is untouched")
}

// TestTheASCIITableGetsTheSameDetailRow, so switching OUTPUT does not change
// what you can see about the columns.
func TestTheASCIITableGetsTheSameDetailRow(t *testing.T) {
	data, types := typedTable()

	lines := strings.Split(FormatASCIITableWithTypes(data, types), "\n")
	require.GreaterOrEqual(t, len(lines), 3)

	assert.NotContains(t, lines[1], "(PK1)", "the marker moved to the row below")
	assert.Contains(t, lines[2], "PK1 uuid")
	assert.Contains(t, lines[2], "C timestamp")
}

// TestBatchOutputKeepsOneHeaderRow. Batch output is piped somewhere and read by
// something else as often as by a person, and a second header row is one more
// line to have to skip.
func TestBatchOutputKeepsOneHeaderRow(t *testing.T) {
	data, _ := typedTable()

	lines := strings.Split(FormatASCIITable(data), "\n")

	assert.Contains(t, lines[1], "id (PK1)")
	assert.Contains(t, lines[2], "---", "the separator follows the names directly")
}
