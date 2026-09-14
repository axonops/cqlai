package ui

import (
	"strings"

	"github.com/axonops/cqlai/internal/ui/completion"
)

// Putting a chosen completion into the prompt.
//
// There were three of these: one for Enter, one for the Tab that finds a single
// match, and one for Space. They agreed about the ordinary case and differed
// everywhere else - only Space added the space that finishes a word, and only
// Enter knew what to do after an equals sign - so which key you pressed decided
// what you got.

// applyCompletion is what the prompt becomes when a completion is chosen.
func applyCompletion(input, chosen string) string {
	// A hint - `<table name>` - is the note that says what to type there. It
	// is not a word to put in the prompt, whichever key was pressed on it.
	if completion.IsHint(chosen) {
		return input
	}

	// The brace a map opens with is punctuation rather than a word: it goes on
	// the end of the option it belongs to.
	if strings.HasPrefix(chosen, "{") {
		opened := strings.TrimRight(input, " ") + " " + chosen
		if strings.HasSuffix(input, "=") {
			opened = input + chosen
		}
		return opened + spaceAfter(opened)
	}

	// A closing bracket attaches to what it closes, rather than landing after
	// the space that follows the last word: `(id uuid PRIMARY KEY )`.
	if chosen == ")" || chosen == "}" {
		closed := strings.TrimRight(input, " ") + chosen
		return closed + spaceAfter(closed)
	}

	// A key or a value inside a map is offered with the quotes it needs, and it
	// replaces what has been typed of it rather than being appended after it:
	// `compaction = {'` and 'class': is `compaction = {'class': `, not
	// `compaction = {' 'class': `.
	if inMap, ok := insideAMap(input, chosen); ok {
		return inMap + spaceAfter(inMap)
	}

	newValue := ""

	// Special case: if input ends with a dot (keyspace.), just append the table name
	if strings.HasSuffix(input, ".") { //nolint:gocritic // more readable as if
		newValue = input + chosen
	} else if strings.HasSuffix(input, "=") {
		// For option assignments (FORMAT=, COMPRESSION=), just append the value
		newValue = input + "'" + chosen + "'"
	} else if strings.HasSuffix(input, "'") || strings.HasSuffix(input, "\"") {
		// Input ends with a quote (complete quoted string), append with space
		newValue = input + " " + chosen
	} else if strings.HasSuffix(input, " ") {
		// Just append the completion
		newValue = input + chosen
	} else {
		// Check if we have a partial word to replace
		lastSpace := strings.LastIndex(input, " ")
		if lastSpace >= 0 {
			// Check if the last word is a complete token that shouldn't be replaced
			lastWord := input[lastSpace+1:]
			upperLastWord := strings.ToUpper(lastWord)

			// Check if this is a complete parameter=value assignment
			isCompleteAssignment := false
			if strings.Contains(lastWord, "=") {
				// Check various complete assignment patterns
				// Use case-insensitive check for boolean values
				lowerLastWord := strings.ToLower(lastWord)
				if strings.HasSuffix(lastWord, "'") || // String value: FORMAT='parquet'
					strings.HasSuffix(upperLastWord, "TRUE") || // Boolean: HEADER=TRUE
					strings.HasSuffix(upperLastWord, "FALSE") || // Boolean: HEADER=FALSE
					strings.HasSuffix(lowerLastWord, "true") || // Boolean: header=true (lowercase)
					strings.HasSuffix(lowerLastWord, "false") || // Boolean: header=false (lowercase)
					(len(lastWord) > 0 && lastWord[len(lastWord)-1] >= '0' && lastWord[len(lastWord)-1] <= '9') { // Number: PAGESIZE=1000
					isCompleteAssignment = true
				}
			}

			// Check if last word is a quoted string (file path or value)
			isQuotedString := (strings.HasPrefix(lastWord, "'") && strings.HasSuffix(lastWord, "'")) ||
				(strings.HasPrefix(lastWord, "\"") && strings.HasSuffix(lastWord, "\""))

			// Determine how to apply the completion based on the last word
			switch {
			case upperLastWord == "TO" || upperLastWord == "FROM":
				// Don't replace TO/FROM, just append the file path with space
				newValue = input + " " + chosen
			case isQuotedString:
				// Last word is a complete quoted string, don't replace, append with space
				newValue = input + " " + chosen
			case strings.HasSuffix(lastWord, "="):
				// Don't replace, just append the value with quotes
				newValue = input + "'" + chosen + "'"
			case isCompleteAssignment:
				// This is a complete assignment, don't replace, just append with space
				newValue = input + " " + chosen
			case strings.Contains(lastWord, "."):
				// Check if this is completing a table name or completing after table name
				// If the completion starts with "(" it's column list for INSERT, not a table name
				if strings.HasPrefix(chosen, "(") {
					// Append column list after table name
					newValue = input + " " + chosen
				} else {
					// For keyspace.table patterns, replace the part after the dot
					// The completion engine returns just the table name
					dotIndex := strings.LastIndex(input, ".")
					newValue = input[:dotIndex+1] + chosen
				}
			case strings.HasSuffix(lastWord, "("):
				// Inside a bracket that has just been opened: the word goes in
				// it rather than in place of it.
				newValue = input + chosen
			case lastWord == "*" || strings.HasSuffix(lastWord, ")"):
				// Don't replace, just append
				newValue = input + " " + chosen
			default:
				// Replace the partial word
				newValue = input[:lastSpace+1] + chosen
			}
		} else {
			// Replace the entire input (single partial word)
			newValue = chosen
		}
	}

	return newValue + spaceAfter(newValue)
}

// insideAMap is the prompt with a quoted key or value put into the map being
// typed, and whether that is what is being done.
//
// What has been typed of the key or value starts after the last thing that
// ends one: the brace the map opened with, or the comma or colon before this
// one. The quote is part of what is being typed, not something to keep.
func insideAMap(input, chosen string) (string, bool) {
	if !strings.HasPrefix(chosen, "'") && !strings.HasPrefix(chosen, ",") {
		return "", false
	}

	open := strings.LastIndex(input, "{")
	if open < 0 || strings.Contains(input[open:], "}") {
		return "", false
	}

	// A key that carries the comma before it goes after the pair it follows.
	if strings.HasPrefix(chosen, ",") {
		return strings.TrimRight(input, " ") + chosen, true
	}

	typed := input[open+1:]
	start := open + 1 + strings.LastIndexAny(typed, ",: ") + 1
	return input[:start] + chosen, true
}

// spaceAfter is the space to put after a completion that has just been applied,
// or "" where one would be in the way.
//
// A word taken from the list is a finished word, and the next thing typed is the
// next word: without the space you have to reach for one every time, which is
// the opposite of what pressing Tab was for.
//
// Three endings take nothing after them. A keyspace ends with a dot and the
// table follows it; a function ends with an open bracket and its argument
// follows that; and a quote that has not been closed is waiting for what goes
// inside it.
func spaceAfter(value string) string {
	if value == "" || strings.HasSuffix(value, " ") {
		return ""
	}

	// Inside a map the next thing is the comma before the next key, or the
	// brace that closes it, and the key carries its own comma.
	if open := strings.LastIndex(value, "{"); open >= 0 && !strings.Contains(value[open:], "}") {
		return ""
	}

	switch value[len(value)-1] {
	case '.', '(':
		return ""
	case '\'', '"':
		if strings.Count(value, string(value[len(value)-1]))%2 == 1 {
			return "" // opened and not closed: the value goes inside it
		}
	}
	return " "
}
