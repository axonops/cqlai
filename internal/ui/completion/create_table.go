package completion

import "strings"

// Completing the shape of a CREATE TABLE.
//
// A table statement is a name, a bracket, a list of columns with their types, a
// primary key, the closing bracket and then the options. Positions do not
// describe it: `IF NOT EXISTS` moves everything along by three, a keyspace
// qualifier is one word or two depending on how it is written, and a column
// definition can be any number of words. Counting words got it wrong in every
// one of those places - nothing was offered after the table name, nor after a
// column name, nor after PRIMARY - so this reads the text instead.
//
// What is being typed is decided by two things: whether the bracket is open,
// and what the words are since whatever separator came last.

// createTableCompletions is what to offer in a CREATE TABLE, or nothing when
// the statement is not one or has nothing obvious to offer.
func createTableCompletions(input string) []string {
	definition, inColumns := columnUnderCursor(input)

	if !inColumns {
		if strings.Contains(input, ")") {
			// The columns are done. Options come next, and they have their own
			// completion; this only has to offer the word that starts them.
			if !containsWord(input, "WITH") {
				return []string{"WITH"}
			}
			return nil
		}

		// A name, and then the bracket its columns go in.
		if named(input) {
			return []string{"("}
		}
		return nil
	}

	// A type with a parameter - map<text, text> - is not finished until its
	// angle bracket closes, and what goes inside it is another type.
	if strings.Count(definition, "<") > strings.Count(definition, ">") {
		return CQLDataTypes
	}

	words := finishedWords(withoutTypeParameters(definition))
	switch len(words) {
	case 0:
		return nil // the column's name is yours to invent
	case 1:
		return CQLDataTypes
	}

	// After a type: what closes the definition off, or what qualifies it.
	switch strings.ToUpper(words[len(words)-1]) {
	case "PRIMARY":
		return []string{"KEY"}
	case "KEY":
		if len(words) == 2 {
			return []string{"("} // the PRIMARY KEY clause of its own, and its columns
		}
		return []string{")"} // the column is finished, and so is the list
	case "STATIC":
		return []string{"PRIMARY KEY"}
	}

	// A definition that is finished can be followed by the bracket that closes
	// the list, whatever else it can be followed by.
	if len(words) == 2 {
		return []string{"PRIMARY KEY", "STATIC", ")"}
	}
	return []string{")"}
}

// withoutTypeParameters is the definition with what is inside a type's angle
// brackets taken out, so that a type counts as the one word it is:
// `a tuple<int, text>` is a name and a type, not four words.
func withoutTypeParameters(definition string) string {
	var kept strings.Builder
	depth := 0

	for i := 0; i < len(definition); i++ {
		switch definition[i] {
		case '<':
			depth++
		case '>':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				kept.WriteByte(definition[i])
			}
		}
	}
	return kept.String()
}

// finishedWords is the words of a column definition that have been finished.
//
// The last word is only finished once something that is not part of it has
// been typed. Until then it is the word being typed, and it does not move the
// definition on: in `(id te` the type is still what comes next, and counting
// `te` as a word would offer what follows a type instead.
func finishedWords(definition string) []string {
	words := strings.Fields(definition)
	if partialWord(definition) != "" && len(words) > 0 {
		words = words[:len(words)-1]
	}
	return words
}

// columnUnderCursor is the column definition being typed, and whether the
// cursor is inside the column list at all.
//
// The list is the bracket that opens after the table name, and it runs to the
// bracket that closes it - not to the end of the text. A CREATE TABLE has
// brackets after that one: `PRIMARY KEY (a, b)` has its own, and so does
// `WITH CLUSTERING ORDER BY (b DESC)`, and taking the last open bracket for
// the column list offered types inside both of them.
//
// The definition is what has been typed since the comma that started it, which
// is what says whether a name, a type or a qualifier comes next. Commas inside
// a type - `map<text, text>` - or inside the primary key do not start one.
func columnUnderCursor(input string) (definition string, inColumns bool) {
	open := strings.Index(input, "(")
	if open < 0 {
		return "", false
	}

	brackets, angles, start := 1, 0, open+1
	for i := open + 1; i < len(input); i++ {
		switch input[i] {
		case '(':
			brackets++
		case ')':
			brackets--
			if brackets == 0 {
				return "", false // the column list is closed
			}
		case '<':
			angles++
		case '>':
			if angles > 0 {
				angles--
			}
		case ',':
			if brackets == 1 && angles == 0 {
				start = i + 1
			}
		}
	}

	if brackets > 1 {
		// Inside the brackets of a primary key, where a column is named rather
		// than defined.
		return "", true
	}
	return input[start:], true
}

// insideUnclosedBracket is what has been typed since the last bracket that is
// still open, and whether there is one.
//
// This is the bracket of an INSERT - its columns, or its values - where the
// last open bracket is the one being filled in.
func insideUnclosedBracket(input string) (item string, inBrackets bool) {
	open := strings.LastIndex(input, "(")
	if open < 0 || strings.Contains(input[open:], ")") {
		return "", false
	}

	after := input[open+1:]
	if comma := strings.LastIndex(after, ","); comma >= 0 {
		after = after[comma+1:]
	}
	return after, true
}

// named reports whether the statement has got as far as naming the table.
//
// Everything before the name is keywords - CREATE TABLE, and IF NOT EXISTS
// where it is used - so the name is the first word that is not one of those.
func named(input string) bool {
	keywords := map[string]bool{
		"CREATE": true, "TABLE": true, "COLUMNFAMILY": true,
		"IF": true, "NOT": true, "EXISTS": true,
	}

	for _, word := range strings.Fields(input) {
		if !keywords[strings.ToUpper(word)] {
			return true
		}
	}
	return false
}

// containsWord reports whether a word appears in the text as a word.
func containsWord(input, word string) bool {
	for _, typed := range strings.Fields(input) {
		if strings.EqualFold(typed, word) {
			return true
		}
	}
	return false
}

// tableStatementCompletions is what to offer in a CREATE TABLE or ALTER TABLE:
// the shape of the statement, and the options that follow it.
//
// It says whether it answered, rather than leaving an empty answer to be taken
// for "no idea". Inside the brackets of a CREATE TABLE there are places where
// nothing is the right answer - a column's name is yours to invent - and the
// parser would otherwise fill the silence with a list of types.
func tableStatementCompletions(input string) (suggestions []string, handled bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) == 0 || !isTableStatement(words) {
		return nil, false
	}

	if options := tableOptionCompletions(input); len(options) > 0 {
		return options, true
	}

	// Only a table has the shape below. A materialized view is a table
	// statement as far as its options go, and nothing like one before that.
	if words[1] != "TABLE" && words[1] != "COLUMNFAMILY" {
		return nil, false
	}
	if words[0] != "CREATE" && words[0] != "ALTER" {
		return nil, false
	}

	if words[0] == "ALTER" {
		return alterTableCompletions(words, input)
	}

	// Inside the brackets this knows the answer, whatever the answer is.
	if _, inColumns := columnUnderCursor(input); inColumns {
		return createTableCompletions(input), true
	}

	shape := createTableCompletions(input)
	return shape, len(shape) > 0
}

// filterByWord keeps the suggestions that carry on from the word being typed.
//
// The parser's answers go through the same filter further down; these arrive
// before it, so they are filtered here instead of twice.
//
// What is being typed is the run of word characters at the end, and punctuation
// ends it: after `compaction = {` there is no partial word, and filtering the
// keys by "{" would leave none of them.
func filterByWord(input string, suggestions []string) []string {
	partial := strings.ToUpper(partialWord(input))
	if partial == "" {
		return suggestions
	}

	kept := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		// A value is offered with the quotes it needs - 'ALWAYS', 'class': -
		// and the word being typed is inside them.
		candidate := strings.ToUpper(strings.TrimPrefix(suggestion, "'"))
		if strings.HasPrefix(candidate, partial) {
			kept = append(kept, suggestion)
		}
	}
	return kept
}

// partialWord is the word being typed at the end of the text, or "" when the
// text ends with something that is not part of one.
func partialWord(input string) string {
	i := len(input)
	for i > 0 && isWordCharacter(input[i-1]) {
		i--
	}
	return input[i:]
}

// isWordCharacter reports whether a byte can be part of an identifier or a
// keyword.
func isWordCharacter(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '_':
		return true
	}
	return false
}

// alterTableCompletions is what to offer in what an ALTER TABLE changes: a
// column added takes a type, and so does one whose type is being changed.
//
// The column being named is answered with a note higher up; this is the word
// after it, which nothing was offering.
func alterTableCompletions(words []string, input string) (suggestions []string, handled bool) {
	named := func(keyword string) bool {
		for i, word := range words {
			if word == keyword && i == len(words)-2 {
				return true
			}
		}
		return false
	}

	// ALTER TABLE t ADD c <type>, and ALTER TABLE t ALTER c TYPE <type>.
	if strings.HasSuffix(input, " ") && (named("ADD") || words[len(words)-1] == "TYPE") {
		return CQLDataTypes, true
	}
	return nil, false
}
