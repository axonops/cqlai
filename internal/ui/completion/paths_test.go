package completion

import (
	"testing"

	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPathIsExpectedAfterEveryCommandThatTakesOne. Doing it for one command
// is how the next one becomes a bug.
func TestAPathIsExpectedAfterEveryCommandThatTakesOne(t *testing.T) {
	tests := []struct{ input, prefix, quote string }{
		{"CAPTURE JSON /tmp/rep", "/tmp/rep", ""},
		{"capture json /tmp/rep", "/tmp/rep", ""},
		{"CAPTURE PARQUET /tmp/", "/tmp/", ""},
		{"SOURCE /tmp/setup", "/tmp/setup", ""},
		{"SAVE /tmp/out", "/tmp/out", ""},
		{"COPY users TO /tmp/out", "/tmp/out", ""},
		{"COPY users FROM /tmp/in", "/tmp/in", ""},

		// Quoted, which is how these are usually written. A quote means a
		// path even without a format, since CAPTURE 'file' is still valid.
		{"CAPTURE '/tmp/rep", "/tmp/rep", "'"},
		{"CAPTURE /tmp/rep", "/tmp/rep", ""},
		{`SOURCE "/tmp/rep`, "/tmp/rep", `"`},
		{"COPY users TO '/tmp/out", "/tmp/out", "'"},

		// Nothing typed yet.
		{"CAPTURE JSON ", "", ""},
		{"SOURCE ", "", ""},
	}

	for _, tt := range tests {
		prefix, quote, ok := PathPrefix(tt.input)
		if !assert.True(t, ok, "%q should expect a path", tt.input) {
			continue
		}
		assert.Equal(t, tt.prefix, prefix, "prefix of %q", tt.input)
		assert.Equal(t, tt.quote, quote, "quote of %q", tt.input)
	}
}

// TestSelectFromIsNotAPath. COPY's FROM takes a file; SELECT's does not.
func TestSelectFromIsNotAPath(t *testing.T) {
	for _, input := range []string{
		"CAPTURE",
		"CAPTURE ",
		"CAPTURE JS",
		"SELECT * FROM users",
		"select id from myks.users where x = 1",
		"DESCRIBE TABLE users",
		"CONSISTENCY QUORUM",
		"",
	} {
		_, _, ok := PathPrefix(input)
		assert.False(t, ok, "%q should not expect a path", input)
	}
}

// TestOnceWITHStartsTheFilenameIsBehindUs, so what is being typed is an option
// rather than a path.
func TestOnceWITHStartsTheFilenameIsBehindUs(t *testing.T) {
	for _, input := range []string{
		"CAPTURE CSV 'out.csv' WITH ",
		"CAPTURE CSV 'out.csv' WITH COMPRESSION=",
		"COPY users TO '/tmp/a.csv' WITH HEADER=",
	} {
		_, _, ok := PathPrefix(input)
		assert.False(t, ok, "%q is past the filename", input)
	}
}

// TestAFinishedFilenameIsNotStillBeingTyped.
func TestAFinishedFilenameIsNotStillBeingTyped(t *testing.T) {
	_, _, ok := PathPrefix("CAPTURE CSV 'out.csv' ")
	assert.False(t, ok, "the filename is closed and done with")
}

// TestTheCompletionGoesBackWhereItCameFrom, keeping the command in front of it
// and the quote around it.
func TestTheCompletionGoesBackWhereItCameFrom(t *testing.T) {
	assert.Equal(t, "CAPTURE JSON '/tmp/report.json",
		ReplacePath("CAPTURE JSON '/tmp/rep", "/tmp/rep", "'", "/tmp/report.json"))

	assert.Equal(t, "COPY users TO '/tmp/out.csv",
		ReplacePath("COPY users TO '/tmp/o", "/tmp/o", "'", "/tmp/out.csv"))

	// Nothing typed yet: the completion is simply appended.
	assert.Equal(t, "CAPTURE CSV /tmp/", ReplacePath("CAPTURE CSV ", "", "", "/tmp/"))
}

// TestCaptureOffersItsFormatsBeforeAPath.
//
// The syntax is CAPTURE <format> <file>, so after the word itself the formats
// are what is wanted - completing a path there was offering the wrong thing.
func TestCaptureOffersItsFormatsBeforeAPath(t *testing.T) {
	sce := &SimpleCompletionEngine{}

	after := sce.GetTokenCompletions("CAPTURE ")
	assert.Contains(t, after, "JSON")
	assert.Contains(t, after, "CSV")
	assert.Contains(t, after, "PARQUET")
	assert.Contains(t, after, "OFF", "stopping is one of the things you can do")

	partial := sce.GetTokenCompletions("CAPTURE PAR")
	assert.Equal(t, []string{"PARQUET"}, partial)
}

// TestTheFormatsComeFromTheCommandThatParsesThem.
func TestTheFormatsComeFromTheCommandThatParsesThem(t *testing.T) {
	sce := &SimpleCompletionEngine{}

	for _, format := range router.CaptureFormats() {
		assert.Contains(t, sce.getCaptureFormats(), format)
	}
}

// TestAFileIsExpectedWithOrWithoutASpaceAfterTheFormat.
//
// Once the format is complete the next thing is a file, whether or not the
// space has been typed yet.
func TestAFileIsExpectedWithOrWithoutASpaceAfterTheFormat(t *testing.T) {
	for _, input := range []string{"CAPTURE CSV", "CAPTURE CSV ", "capture json", "CAPTURE PARQUET"} {
		prefix, _, ok := PathPrefix(input)
		assert.True(t, ok, "%q should expect a file", input)
		assert.Empty(t, prefix, "nothing has been typed of it yet")
	}
}

// TestAPartialFormatIsStillAFormat, not a file.
func TestAPartialFormatIsStillAFormat(t *testing.T) {
	for _, input := range []string{"CAPTURE CS", "CAPTURE JSO", "CAPTURE P"} {
		_, _, ok := PathPrefix(input)
		assert.False(t, ok, "%q is still being typed", input)
	}
}

// TestSomethingThatOnlyStartsLikeAFormatIsNot.
func TestSomethingThatOnlyStartsLikeAFormatIsNot(t *testing.T) {
	_, _, ok := PathPrefix("CAPTURE CSVX")
	assert.False(t, ok, "CSVX is not a format")
}

// TestCompletingBetweenAPairOfQuotes.
//
// The usage shows the filename quoted, so the pair is what gets typed. The
// closing quote must not be taken as part of the path, and must survive the
// completion.
func TestCompletingBetweenAPairOfQuotes(t *testing.T) {
	prefix, quote, ok := PathPrefix("CAPTURE CSV ''")
	require.True(t, ok)
	assert.Empty(t, prefix, "there is nothing between them yet")
	assert.Equal(t, "'", quote)

	assert.Equal(t, "CAPTURE CSV '/tmp/'",
		ReplacePath("CAPTURE CSV ''", "", "'", "/tmp/"))

	prefix, quote, ok = PathPrefix("CAPTURE CSV '/tmp/rep'")
	require.True(t, ok)
	assert.Equal(t, "/tmp/rep", prefix, "the closing quote is not part of the path")

	assert.Equal(t, "CAPTURE CSV '/tmp/report.csv'",
		ReplacePath("CAPTURE CSV '/tmp/rep'", "/tmp/rep", quote, "/tmp/report.csv"))
}

// TestAnUnclosedQuoteStillWorks, since people type one and carry on.
func TestAnUnclosedQuoteStillWorks(t *testing.T) {
	prefix, quote, ok := PathPrefix("CAPTURE CSV '/tmp/rep")
	require.True(t, ok)
	assert.Equal(t, "/tmp/rep", prefix)

	assert.Equal(t, "CAPTURE CSV '/tmp/report.csv",
		ReplacePath("CAPTURE CSV '/tmp/rep", prefix, quote, "/tmp/report.csv"))
}

// TestTheWholeCaptureGrammarCompletes:
//
//	CAPTURE [JSON|CSV|PARQUET] 'filename' [WITH option=value AND ...]
//	CAPTURE OFF
func TestTheWholeCaptureGrammarCompletes(t *testing.T) {
	sce := &SimpleCompletionEngine{}

	tests := []struct {
		input string
		want  []string
	}{
		{"CAPTURE ", append(router.CaptureFormats(), "OFF")},
		{"CAPTURE PAR", []string{"PARQUET"}},
		{"CAPTURE OF", []string{"OFF"}},

		// After the filename, the only thing that follows is WITH.
		{"CAPTURE CSV 'out.csv' ", []string{"WITH"}},

		// The options, then a value, then AND.
		{"CAPTURE CSV 'out.csv' WITH ", router.CaptureOptions()},
		{"CAPTURE CSV 'out.csv' WITH COMP", []string{"COMPRESSION"}},
		{"CAPTURE CSV 'out.csv' WITH COMPRESSION=", ParquetCompressionTypes},
		{"CAPTURE CSV 'out.csv' WITH MAX_FILE_SIZE=1000", []string{"AND"}},
		{"CAPTURE CSV 'out.csv' WITH COMPRESSION='snappy'", []string{"AND"}},
		{"CAPTURE CSV 'out.csv' WITH COMPRESSION='snappy' AND ", router.CaptureOptions()},

		// Nothing follows OFF.
		{"CAPTURE OFF ", nil},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, sce.GetTokenCompletions(tt.input), "after %q", tt.input)
	}
}

// TestTheOptionsComeFromTheCommandThatParsesThem.
func TestTheOptionsComeFromTheCommandThatParsesThem(t *testing.T) {
	assert.Equal(t, []string{"COMPRESSION", "MAX_FILE_SIZE", "PARTITION"}, router.CaptureOptions())
}
