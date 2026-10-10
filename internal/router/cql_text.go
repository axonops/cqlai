package router

import (
	"regexp"
	"strings"
)

// Reading CQL as text: where its comments are, and where a statement ends.
//
// Both have to skip what is inside a string ('it''s'), a quoted name
// ("My Table") and a $$ body - a function's Java, which has comments and
// semicolons of its own.

// StripComments removes the comments from CQL: -- and // to the end of their
// line, and /* */ blocks. The end of the line stays, so what follows a
// comment is still read: a comment on the first line of a statement typed
// over several lines must not take the lines after it with it.
func StripComments(input string) string {
	stripped, _ := scanCQL(input)
	return strings.TrimSpace(stripped)
}

// StatementComplete reports whether text is a whole statement: it ends with
// a semicolon that is not inside a string, a quoted name, a $$ body or a
// comment. A BATCH is whole only at its APPLY BATCH;, since the statements
// inside it end with semicolons too.
func StatementComplete(text string) bool {
	stripped, open := scanCQL(text)
	if open {
		return false
	}
	s := strings.TrimSpace(stripped)
	if !strings.HasSuffix(s, ";") {
		return false
	}
	if beginsBatch(s) {
		return applyBatch.MatchString(s)
	}
	return true
}

// applyBatch is the end of a BATCH.
var applyBatch = regexp.MustCompile(`(?i)\bAPPLY\s+BATCH\s*;$`)

// beginsBatch reports whether a statement is a BATCH: BEGIN BATCH, or BEGIN
// UNLOGGED or COUNTER BATCH.
func beginsBatch(s string) bool {
	words := strings.Fields(strings.ToUpper(s))
	switch {
	case len(words) < 2 || words[0] != "BEGIN":
		return false
	case strings.HasPrefix(words[1], "BATCH"):
		return true
	}
	return len(words) > 2 && (words[1] == "UNLOGGED" || words[1] == "COUNTER") && strings.HasPrefix(words[2], "BATCH")
}

// scanCQL is the text without its comments, and whether it ends inside a
// string, a quoted name, a $$ body or a /* comment.
func scanCQL(input string) (string, bool) {
	stripped, open, _ := scanCQLEnds(input)
	return stripped, open
}

// semicolon is where a semicolon outside any string or comment is: in the
// input, and in the text without its comments.
type semicolon struct{ in, out int }

// scanCQLEnds is scanCQL, with where each semicolon that can end a statement
// is.
func scanCQLEnds(input string) (string, bool, []semicolon) {
	var ends []semicolon
	var b strings.Builder
	b.Grow(len(input))

	n := len(input)
	for i := 0; i < n; {
		c := input[i]
		switch {
		case c == '\'' || c == '"':
			// To the closing quote; two of them together are one inside.
			j := i + 1
			closed := false
			for j < n {
				if input[j] == c {
					if j+1 < n && input[j+1] == c {
						j += 2
						continue
					}
					j++
					closed = true
					break
				}
				j++
			}
			b.WriteString(input[i:j])
			if !closed {
				return b.String(), true, ends
			}
			i = j

		case c == '$' && i+1 < n && input[i+1] == '$':
			end := strings.Index(input[i+2:], "$$")
			if end < 0 {
				b.WriteString(input[i:])
				return b.String(), true, ends
			}
			j := i + 2 + end + 2
			b.WriteString(input[i:j])
			i = j

		case (c == '-' || c == '/') && i+1 < n && input[i+1] == c:
			// To the end of the line, which is kept. Cassandra ends a line
			// comment at a carriage return as well as a newline.
			for i < n && input[i] != '\n' && input[i] != '\r' {
				i++
			}

		case c == '/' && i+1 < n && input[i+1] == '*':
			end := strings.Index(input[i+2:], "*/")
			if end < 0 {
				return b.String(), true, ends
			}
			// A space where it was, so the words either side stay apart.
			b.WriteByte(' ')
			i += 2 + end + 2

		default:
			if c == ';' {
				ends = append(ends, semicolon{in: i, out: b.Len()})
			}
			b.WriteByte(c)
			i++
		}
	}
	return b.String(), false, ends
}

// SplitStatements splits CQL text - a file for -f or SOURCE - into its
// statements, each as written, comments and all. A statement ends at a
// semicolon outside a string, a quoted name, a $$ body or a comment, and a
// BATCH only at its APPLY BATCH;. Text after the last semicolon is a
// statement of its own unless it is only space and comments.
//
// It reads the text once. The splitter it replaces upper-cased the rest of
// the file at every B and A, which took minutes on a large file.
func SplitStatements(text string) []string {
	stripped, _, ends := scanCQLEnds(text)
	var statements []string
	inStart, outStart := 0, 0
	batch, known := false, false
	for _, e := range ends {
		body := stripped[outStart : e.out+1]
		if !known {
			batch, known = beginsBatch(strings.TrimSpace(body)), true
		}
		if batch {
			tail := strings.TrimSpace(body)
			if len(tail) > 64 {
				tail = tail[len(tail)-64:]
			}
			if !applyBatch.MatchString(tail) {
				continue
			}
		}
		if strings.TrimSpace(body) != ";" {
			statements = append(statements, strings.TrimSpace(text[inStart:e.in+1]))
		}
		inStart, outStart = e.in+1, e.out+1
		known = false
	}
	if strings.TrimSpace(stripped[outStart:]) != "" {
		statements = append(statements, strings.TrimSpace(text[inStart:]))
	}
	return statements
}
