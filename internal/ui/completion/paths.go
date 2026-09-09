package completion

import (
	"regexp"
	"strings"
)

// Completing a file path at the prompt.
//
// Every command that takes a file had the same gap: completion knew the
// keyword and offered nothing after it, so paths had to be typed out in full
// and correctly. One rule covers all of them, because doing it for one command
// is how the next one becomes a bug.

// pathContexts match a command up to the point where a file path begins.
//
// The path itself is left out of the match deliberately: what has been typed so
// far is whatever follows, quoted or not.
var pathContexts = []*regexp.Regexp{
	// CAPTURE takes a format and then a file. Without the format there is
	// nothing to complete a path against yet - the formats are what is wanted
	// there, and getCaptureCompletions offers them.
	// The space after the format is optional: once the format is complete the
	// next thing is a file, whether or not you have pressed space yet.
	regexp.MustCompile(`(?i)^\s*CAPTURE\s+(?:CSV|JSON|PARQUET)\b\s*`),
	regexp.MustCompile(`(?i)^\s*SOURCE\s+`),
	regexp.MustCompile(`(?i)^\s*SAVE\s+`),
	// COPY takes a table first, then TO or FROM, then the file.
	regexp.MustCompile(`(?i)\s(?:TO|FROM)\s+`),
}

// withClause matches the start of a WITH clause, after which the filename is
// finished and options are being typed.
var withClause = regexp.MustCompile(`(?i)\sWITH\b`)

// captureWord matches CAPTURE and the space after it, with nothing said about
// what follows.
var captureWord = regexp.MustCompile(`(?i)^\s*CAPTURE\s+`)

// pathStart are the characters that mean what is being typed is a path rather
// than a keyword.
const pathStart = `'"/~.`

// PathPrefix reports whether the input is at a point where a file path is
// expected, and returns what has been typed of it.
//
// The quote is reported separately so the completion can be put back inside it
// rather than completed against a leading apostrophe.
func PathPrefix(input string) (prefix, quote string, ok bool) {
	// Once WITH appears the filename is behind us and the options are what is
	// being typed, so the path completer has no business here.
	if withClause.MatchString(input) {
		return "", "", false
	}

	// CAPTURE takes a format and then a file, so after the word itself the
	// formats are what is wanted - unless what is being typed already looks
	// like a path, since CAPTURE 'file' with no format is still valid.
	if loc := captureWord.FindStringIndex(input); loc != nil {
		rest := input[loc[1]:]
		if rest != "" && strings.ContainsRune(pathStart, rune(rest[0])) {
			return unquote(rest)
		}
	}

	for _, context := range pathContexts {
		loc := context.FindStringIndex(input)
		if loc == nil {
			continue
		}

		// COPY's TO/FROM has to belong to a COPY, not to a SELECT ... FROM.
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input)), "COPY") &&
			context == pathContexts[len(pathContexts)-1] {
			continue
		}

		rest := input[loc[1]:]

		// Only the last one counts: COPY t TO 'a' has a TO before the file.
		if next := lastContext(rest); next != "" {
			rest = next
		}

		if finishedPath(rest) {
			return "", "", false
		}
		return unquote(rest)
	}
	return "", "", false
}

// finishedPath reports whether the quoted filename is closed and something
// follows it, which means what is being typed is no longer the path.
func finishedPath(rest string) bool {
	for _, q := range []string{"'", `"`} {
		if !strings.HasPrefix(rest, q) {
			continue
		}
		if i := strings.Index(rest[1:], q); i >= 0 {
			return strings.TrimSpace(rest[i+2:]) != "" || strings.HasSuffix(rest, " ")
		}
	}
	return false
}

// unquote separates a path from the quotes it was typed inside, so the
// completion can be put back between them.
//
// A closing quote is stripped as well as an opening one: the pair is what the
// usage shows, and without this the closing quote would be taken as part of
// the path and nothing would ever match.
func unquote(rest string) (prefix, quote string, ok bool) {
	for _, q := range []string{"'", `"`} {
		if !strings.HasPrefix(rest, q) {
			continue
		}
		return strings.TrimSuffix(rest[1:], q), q, true
	}
	return rest, "", true
}

// lastContext handles a second TO or FROM further along the line, so the path
// being completed is the one under the cursor rather than an earlier one.
func lastContext(rest string) string {
	last := pathContexts[len(pathContexts)-1].FindAllStringIndex(rest, -1)
	if len(last) == 0 {
		return ""
	}
	return rest[last[len(last)-1][1]:]
}

// ReplacePath puts a completed path back into the input, keeping whatever came
// before it and the quotes it was between.
func ReplacePath(input, prefix, quote, completed string) string {
	if quote == "" {
		i := strings.LastIndex(input, prefix)
		if i < 0 {
			return input
		}
		return input[:i] + completed
	}

	// The closing quote, when there is one, has to survive: the completion
	// goes between the pair rather than replacing the end of the line.
	tail, closing := quote+prefix, ""
	if strings.HasSuffix(input, quote+prefix+quote) {
		tail, closing = quote+prefix+quote, quote
	}

	i := strings.LastIndex(input, tail)
	if i < 0 {
		return input
	}
	return input[:i] + quote + completed + closing
}
