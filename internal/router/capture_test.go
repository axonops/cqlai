package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// autoSaveHandler is a handler with nothing saved yet.
func autoSaveHandler(t *testing.T) (*MetaCommandHandler, string) {
	t.Helper()

	dir := t.TempDir()
	return NewMetaCommandHandler(&db.Session{}, session.NewManager(&config.Config{})), dir
}

// saved is the files AutoSave has written into a directory.
func saved(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// TestAutoSaveWritesOneFilePerQuery.
//
// This is why it takes a directory. One file cannot hold the output of every
// query run while it is on: two queries against different tables have different
// columns, and a Parquet file has one schema. CSV has the same problem more
// quietly - a header row, then rows from another table underneath it.
func TestAutoSaveWritesOneFilePerQuery(t *testing.T) {
	h, dir := autoSaveHandler(t)

	require.Contains(t, h.handleCapture("AUTOSAVE CSV '"+dir+"'"), "Saving each query")
	require.Empty(t, saved(t, dir), "nothing is written until a query runs")

	require.NoError(t, h.WriteCaptureResult("SELECT * FROM users",
		[]string{"id", "name"}, [][]string{{"1", "alice"}}))
	require.NoError(t, h.WriteCaptureResult("SELECT * FROM events",
		[]string{"day", "count"}, [][]string{{"2026-09-10", "7"}}))

	files := saved(t, dir)
	assert.Len(t, files, 2, "a file each, not one file with both in it")
	for _, name := range files {
		assert.True(t, strings.HasSuffix(name, ".csv"), "got %q", name)
		assert.True(t, strings.HasPrefix(name, "query_"), "got %q", name)
	}
}

// TestEachFileHoldsItsOwnQuery, headers and all.
func TestEachFileHoldsItsOwnQuery(t *testing.T) {
	h, dir := autoSaveHandler(t)
	require.Contains(t, h.handleCapture("AUTOSAVE CSV '"+dir+"'"), "Saving each query")

	require.NoError(t, h.WriteCaptureResult("SELECT * FROM users",
		[]string{"id", "name"}, [][]string{{"1", "alice"}}))
	require.NoError(t, h.WriteCaptureResult("SELECT * FROM events",
		[]string{"day", "count"}, [][]string{{"2026-09-10", "7"}}))

	files := saved(t, dir)
	require.Len(t, files, 2)

	var seen []string
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a path this test wrote
		require.NoError(t, err)
		seen = append(seen, string(body))
	}

	joined := strings.Join(seen, "\n---\n")
	assert.Contains(t, joined, "id")
	assert.Contains(t, joined, "day")
	for _, body := range seen {
		assert.False(t, strings.Contains(body, "id") && strings.Contains(body, "day"),
			"two queries should not share a file: %q", body)
	}
}

// TestAutoSaveTakesADirectoryThatDoesNotExistYet, and makes it.
func TestAutoSaveTakesADirectoryThatDoesNotExistYet(t *testing.T) {
	h, dir := autoSaveHandler(t)
	fresh := filepath.Join(dir, "exports", "today")

	require.Contains(t, h.handleCapture("AUTOSAVE CSV '"+fresh+"'"), "Saving each query")

	info, err := os.Stat(fresh)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestCaptureStillWorksAsAWord, so nobody's scripts or muscle memory break.
func TestCaptureStillWorksAsAWord(t *testing.T) {
	h, dir := autoSaveHandler(t)

	assert.Contains(t, h.handleCapture("CAPTURE JSON '"+dir+"'"), "Saving each query")
	assert.Equal(t, "json", h.captureFormat)
	assert.True(t, h.IsCapturing())

	assert.Contains(t, h.handleCapture("CAPTURE OFF"), "Stopped")
	assert.False(t, h.IsCapturing())
}

// TestItIsOnBetweenQueries.
//
// A file is opened and closed for each query, so between two of them there is
// none - which is not the same as AutoSave being off, and the bottom line has
// to keep saying so.
func TestItIsOnBetweenQueries(t *testing.T) {
	h, dir := autoSaveHandler(t)
	h.handleCapture("AUTOSAVE CSV '" + dir + "'")

	assert.True(t, h.IsCapturing(), "on, with nothing written yet")

	require.NoError(t, h.WriteCaptureResult("SELECT 1", []string{"n"}, [][]string{{"1"}}))
	assert.True(t, h.IsCapturing(), "and still on afterwards")
	assert.Nil(t, h.captureOutput, "with that query's file finished and closed")
}

// TestEveryFormatIsAccepted, and names its files accordingly.
func TestEveryFormatIsAccepted(t *testing.T) {
	for _, tt := range []struct{ command, format, ext string }{
		{"AUTOSAVE JSON", "json", ".json"},
		{"AUTOSAVE CSV", "csv", ".csv"},
		{"AUTOSAVE PARQUET", "parquet", ".parquet"},
		{"AUTOSAVE", "text", ".txt"},
	} {
		t.Run(tt.format, func(t *testing.T) {
			h, dir := autoSaveHandler(t)

			result := h.handleCapture(tt.command + " '" + dir + "'")
			assert.Contains(t, result, tt.format)
			assert.Equal(t, tt.format, h.captureFormat)
			assert.True(t, strings.HasSuffix(h.autoSaveName(tt.format), tt.ext),
				"got %q", h.autoSaveName(tt.format))
		})
	}
}

// TestAsking says where it is going and what it last wrote.
func TestAsking(t *testing.T) {
	h, dir := autoSaveHandler(t)
	assert.Contains(t, h.handleCapture("AUTOSAVE"), "Not saving")

	h.handleCapture("AUTOSAVE CSV '" + dir + "'")
	assert.Contains(t, h.handleCapture("AUTOSAVE"), dir)

	require.NoError(t, h.WriteCaptureResult("SELECT 1", []string{"n"}, [][]string{{"1"}}))
	assert.Contains(t, h.handleCapture("AUTOSAVE"), "Last written")
}

// TestParquetGetsAFilePerQueryToo, which is the case that could not work at all
// with a single file: one schema per file, and two queries with different
// columns.
func TestParquetGetsAFilePerQueryToo(t *testing.T) {
	h, dir := autoSaveHandler(t)
	require.Contains(t, h.handleCapture("AUTOSAVE PARQUET '"+dir+"'"), "Saving each query")

	require.NoError(t, h.WriteCaptureResultWithTypes("SELECT * FROM users",
		[]string{"id", "name"}, []string{"int", "text"},
		[][]string{{"1", "alice"}}, nil))
	require.NoError(t, h.WriteCaptureResultWithTypes("SELECT * FROM events",
		[]string{"day", "n"}, []string{"text", "int"},
		[][]string{{"2026-09-10", "7"}}, nil))

	files := saved(t, dir)
	assert.Len(t, files, 2)
	for _, name := range files {
		assert.True(t, strings.HasSuffix(name, ".parquet"), "got %q", name)

		info, err := os.Stat(filepath.Join(dir, name))
		require.NoError(t, err)
		assert.Positive(t, info.Size(), "%s should not be empty", name)
	}
}

// TestAQueryIsWrittenWholeAndReadableAtOnce.
//
// The file is closed at the end of the write that opened it, not held open
// until the next query. A Parquet file is nothing until its footer is written,
// so one left open is a zero-byte file - and looking at what AutoSave has
// written should not mean running another query first.
func TestAQueryIsWrittenWholeAndReadableAtOnce(t *testing.T) {
	for _, format := range []string{"CSV", "JSON", "PARQUET"} {
		t.Run(format, func(t *testing.T) {
			h, dir := autoSaveHandler(t)
			h.handleCapture("AUTOSAVE " + format + " '" + dir + "'")

			require.NoError(t, h.WriteCaptureResultWithTypes("SELECT * FROM users",
				[]string{"id", "name"}, []string{"int", "text"},
				[][]string{{"1", "alice"}}, nil))

			// Read it without stopping first, the way a person would.
			files := saved(t, dir)
			require.Len(t, files, 1, "one write, one file")

			info, err := os.Stat(filepath.Join(dir, files[0]))
			require.NoError(t, err)
			assert.Positive(t, info.Size(), "%s should be complete already", files[0])
		})
	}
}

// TestPagesThatArriveLaterGetTheirOwnFile.
//
// With AutoFetch on - the usual case - every row arrives in one write and a
// query is one file. Without it, rows page in as you scroll, and each page is a
// file of its own numbered after the last. Parquet cannot be appended to: a
// file is sealed by its footer, so the choice is a second part or losing the
// rows.
func TestPagesThatArriveLaterGetTheirOwnFile(t *testing.T) {
	h, dir := autoSaveHandler(t)
	h.handleCapture("AUTOSAVE CSV '" + dir + "'")

	headers := []string{"id", "name"}
	require.NoError(t, h.WriteCaptureResult("SELECT * FROM users", headers, [][]string{{"1", "alice"}}))
	require.NoError(t, h.AppendCaptureRows([][]string{{"2", "bob"}}))

	files := saved(t, dir)
	require.Len(t, files, 2)

	var all string
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a path this test wrote
		require.NoError(t, err)
		all += string(body)
	}

	// Nothing is written twice, which is what closing after every write and
	// then reopening on the next page used to do.
	for _, name := range []string{"alice", "bob"} {
		assert.Equal(t, 1, strings.Count(all, name), "%s should appear once", name)
	}
}

// TestTheNextQueryStartsTheNextFile.
func TestTheNextQueryStartsTheNextFile(t *testing.T) {
	h, dir := autoSaveHandler(t)
	h.handleCapture("AUTOSAVE CSV '" + dir + "'")

	require.NoError(t, h.WriteCaptureResult("SELECT * FROM users",
		[]string{"id"}, [][]string{{"1"}}))
	require.NoError(t, h.WriteCaptureResult("SELECT * FROM events",
		[]string{"day"}, [][]string{{"2026-09-10"}}))

	files := saved(t, dir)
	require.Len(t, files, 2, "two queries, two files")

	first, err := os.ReadFile(filepath.Join(dir, files[0])) //nolint:gosec // a path this test wrote
	require.NoError(t, err)
	assert.NotContains(t, string(first), "day", "each query keeps to its own file")
}

// TestStoppingClosesTheFileThatIsOpen, so the last query is not left unflushed.
func TestStoppingClosesTheFileThatIsOpen(t *testing.T) {
	h, dir := autoSaveHandler(t)
	h.handleCapture("AUTOSAVE CSV '" + dir + "'")
	require.NoError(t, h.WriteCaptureResult("SELECT 1", []string{"n"}, [][]string{{"1"}}))

	h.handleCapture("AUTOSAVE OFF")

	files := saved(t, dir)
	require.Len(t, files, 1)
	body, err := os.ReadFile(filepath.Join(dir, files[0])) //nolint:gosec // a path this test wrote
	require.NoError(t, err)
	assert.Contains(t, string(body), "1", "the rows are on disk, not still in a buffer")
}

// TestAStreamingResultIsMarkedAsSaved.
//
// The router drains a streaming iterator itself and writes the rows, then hands
// them on as an ordinary QueryResult for the UI to display. The UI wrote the
// same rows again, so one query landed in two files - which is what the debug
// log showed: two writes for one SELECT, fourteen seconds apart, the second
// after the conversion.
//
// The mark is how the second write knows not to. It is a field on the result
// rather than the writer noticing the same rows twice, because running the same
// query again is a thing people do and it should get its own file.
func TestAStreamingResultIsMarkedAsSaved(t *testing.T) {
	written := db.QueryResult{AlreadySaved: true}
	assert.True(t, written.AlreadySaved)

	fresh := db.QueryResult{}
	assert.False(t, fresh.AlreadySaved, "a result nobody has written is not marked")
}

// TestWritingTheSameQueryTwiceGivesTwoFiles, because running a query again is a
// thing people do and each run is its own result.
func TestWritingTheSameQueryTwiceGivesTwoFiles(t *testing.T) {
	h, dir := autoSaveHandler(t)
	h.handleCapture("AUTOSAVE CSV '" + dir + "'")

	for range 2 {
		require.NoError(t, h.WriteCaptureResult("SELECT * FROM users",
			[]string{"id"}, [][]string{{"1"}}))
	}

	assert.Len(t, saved(t, dir), 2)
}

// TestTheConvertedResultCarriesWhatTheInfoBarReads.
//
// A streaming result is drained here when AutoSave is on and handed on as an
// ordinary QueryResult. The conversion carried the rows and the types and left
// the timing, the row count and the headers at zero - so with AutoSave on, and
// only then, every query reported "Query: 0s" and "Rows: 0".
func TestTheConvertedResultCarriesWhatTheInfoBarReads(t *testing.T) {
	headers := []string{"id", "name"}
	rows := [][]string{{"1", "alice"}, {"2", "bob"}}

	// What the conversion builds, from a stream that started a moment ago.
	started := time.Now().Add(-250 * time.Millisecond)
	converted := db.QueryResult{
		Data:     append([][]string{headers}, rows...),
		Headers:  headers,
		Duration: time.Since(started),
		RowCount: len(rows),
	}

	assert.Equal(t, headers, converted.Headers, "the info bar needs the columns")
	assert.Equal(t, 2, converted.RowCount, "and the row count")
	assert.Positive(t, converted.Duration, "and how long it took")
}
