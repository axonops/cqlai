package completion

import "strings"

// Saying what to type where completion has nothing to offer.
//
// Completion stops wherever the next word is the user's to invent - the name of
// a keyspace, a table, a column, or a value to compare against. An empty list
// looks the same as completion having no idea, so there was no way to tell
// "type a name here" from "this is not understood".
//
// A hint is a note, not a candidate. It is written `<table name>` and it is
// never put into the prompt.

const (
	hintOpen  = "<"
	hintClose = ">"
)

// Hint is how a note about what to type is written.
func Hint(what string) string {
	return hintOpen + what + hintClose
}

// IsHint reports whether a row in the completion list is a note about what to
// type rather than something to insert.
func IsHint(suggestion string) bool {
	return strings.HasPrefix(suggestion, hintOpen) && strings.HasSuffix(suggestion, hintClose)
}

var (
	keyspaceHint = Hint("keyspace name")
	tableHint    = Hint("table name")
	columnHint   = Hint("column name")
	valueHint    = Hint("value")
)

// thingsYouName is what CREATE makes, and what its name is called.
var thingsYouName = map[string]string{
	"KEYSPACE":     keyspaceHint,
	"TABLE":        tableHint,
	"COLUMNFAMILY": tableHint,
	"TYPE":         Hint("type name"),
	"INDEX":        Hint("index name"),
	"VIEW":         Hint("view name"),
	"ROLE":         Hint("role name"),
	"USER":         Hint("user name"),
	"TRIGGER":      Hint("trigger name"),
	"FUNCTION":     Hint("function name"),
	"AGGREGATE":    Hint("aggregate name"),
}

// hintAfter is what the word before says about the word being typed, where the
// word being typed is a name.
var hintAfter = map[string]string{
	"USE":      keyspaceHint,
	"KEYSPACE": keyspaceHint,
	"FROM":     tableHint,
	"INTO":     tableHint,
	"TABLE":    tableHint,
	"UPDATE":   tableHint,
	"TRUNCATE": tableHint,
	"COPY":     tableHint,
	"ON":       tableHint,
	"WHERE":    columnHint,
	"SET":      columnHint,
	"BY":       columnHint,
}

// numberAfter is what the word before takes, where what it takes is a number
// and nothing else. Nothing else is right there, so these replace whatever the
// parser had to say: it answered LIMIT with the keywords that come before it.
var numberAfter = map[string]string{
	"LIMIT":     Hint("rows"),
	"TTL":       Hint("seconds"),
	"TIMESTAMP": Hint("microseconds"),
}

// hintPlacement says where a note belongs in the list of suggestions.
type hintPlacement int

const (
	noHint hintPlacement = iota

	// hintWhenNothingElse is for a name that could have been looked up - a
	// table, or a column of one. The names themselves are better than a note
	// saying a name goes here, so the note is what is left when there are
	// none, which is what happens with no cluster connected.
	hintWhenNothingElse

	// hintAlongside is for a name you invent. After `CREATE TABLE` the list
	// offers IF, and the name is still the likelier thing to type.
	hintAlongside

	// hintInstead is for the places where the note is the whole answer and the
	// suggestions are wrong: inside the brackets of a VALUES what comes next
	// is a value, not the IF and USING that follow the closing bracket.
	hintInstead
)

// hintFor is the note to show for what is being typed, and where it belongs.
func hintFor(input string) (string, hintPlacement) {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "", noHint // an empty prompt has the commands to offer
	}

	if hint, placement := hintInsideBrackets(input, words); placement != noHint {
		return hint, placement
	}

	if name := nameBeingCreated(input, words); name != "" {
		return name, hintAlongside
	}

	if endsWithComparison(input) {
		return valueHint, hintWhenNothingElse
	}

	finished := finishedWords(input)
	if len(finished) == 0 {
		return "", noHint
	}
	last := strings.ToUpper(finished[len(finished)-1])

	// What an ALTER TABLE changes is a column, named rather than defined.
	if strings.EqualFold(words[0], "ALTER") {
		switch last {
		case "ADD", "TO":
			// A column that does not exist yet: the one being added, and the
			// new name of one being renamed.
			return columnHint, hintAlongside
		case "DROP", "RENAME":
			// One the table has, which can be looked up.
			return columnHint, hintWhenNothingElse
		}
	}

	// AND joins conditions in a WHERE, and options in a WITH. Only the first
	// kind is followed by a column.
	if last == "AND" && containsWord(input, "WHERE") && !containsWord(input, "WITH") {
		return columnHint, hintWhenNothingElse
	}

	if hint, takesANumber := numberAfter[last]; takesANumber {
		return hint, hintInstead
	}
	if hint, expected := hintAfter[last]; expected {
		return hint, hintWhenNothingElse
	}
	return "", noHint
}

// hintInsideBrackets is the note for a bracketed list: the columns of a CREATE
// TABLE, the columns an INSERT names, or the values it gives them.
func hintInsideBrackets(input string, words []string) (string, hintPlacement) {
	// An index says what it is on inside its own brackets, and answers for
	// them itself: the functions that pick a part of a collection go there as
	// well as the column.
	if isCreateIndex(upper(words)) {
		return "", noHint
	}

	// The column list of a CREATE TABLE is the bracket after the table name.
	// Every other CREATE says what goes in its own brackets - a field, an
	// argument, a column of a key - and answers for them itself.
	if strings.EqualFold(words[0], "CREATE") && !isCreateTable(strings.Join(words, " ")) {
		return "", noHint
	}
	if strings.EqualFold(words[0], "CREATE") {
		item, inColumns := columnUnderCursor(input)
		if !inColumns || len(finishedWords(item)) > 0 {
			return "", noHint // past the name: the type or the next word is offered
		}
		// A column definition names a column that does not exist yet.
		return columnHint, hintInstead
	}

	item, inBrackets := insideUnclosedBracket(input)
	if !inBrackets {
		return "", noHint
	}

	// The brackets after VALUES hold values, and so do the ones after IN.
	// The ones before VALUES hold columns.
	if containsWord(input, "VALUES") || wordBeforeBracket(input) == "IN" {
		return valueHint, hintInstead
	}
	if len(finishedWords(item)) > 0 {
		return "", noHint
	}

	// Everywhere else the columns belong to the table, and can be looked up.
	return columnHint, hintWhenNothingElse
}

// upper is the words as the keyword comparisons want them.
func upper(words []string) []string {
	out := make([]string, 0, len(words))
	for _, word := range words {
		out = append(out, strings.ToUpper(word))
	}
	return out
}

// wordBeforeBracket is the word in front of the bracket that is still open,
// which says what the bracket holds.
func wordBeforeBracket(input string) string {
	open := strings.LastIndex(input, "(")
	if open < 0 {
		return ""
	}

	fields := strings.Fields(input[:open])
	if len(fields) == 0 {
		return ""
	}
	return strings.ToUpper(fields[len(fields)-1])
}

// nameBeingCreated is the note for the name a CREATE statement is about to
// give something, or "" when the name has been typed already.
func nameBeingCreated(input string, words []string) string {
	if !strings.EqualFold(words[0], "CREATE") || len(words) < 2 {
		return ""
	}

	rest := words[1:]
	// CREATE OR REPLACE FUNCTION, and CREATE MATERIALIZED VIEW, say what they
	// are making in more than one word.
	for len(rest) > 1 && isPrelude(rest[0]) {
		rest = rest[1:]
	}

	hint, makes := thingsYouName[strings.ToUpper(rest[0])]
	if !makes {
		return ""
	}

	// IF NOT EXISTS comes before the name, and until it is finished it is what
	// is being typed rather than the name.
	if containsWord(input, "IF") && !containsWord(input, "EXISTS") {
		return ""
	}

	// Anything after it other than IF NOT EXISTS is the name, typed already -
	// unless it is the word being typed now, which the note is for.
	typed := rest[1:]
	if partialWord(input) != "" && len(typed) > 0 {
		typed = typed[:len(typed)-1]
	}
	for _, word := range typed {
		switch strings.ToUpper(word) {
		case "IF", "NOT", "EXISTS":
		default:
			return ""
		}
	}
	return hint
}

// isPrelude reports whether a word comes before the thing a CREATE makes.
func isPrelude(word string) bool {
	switch strings.ToUpper(word) {
	case "OR", "REPLACE", "MATERIALIZED", "CUSTOM":
		return true
	}
	return false
}

// endsWithComparison reports whether the text ends where a value goes: after
// an operator, or after the IN that takes a list of them.
//
// An operator stands on its own, which is what tells it from the angle bracket
// that closes a type: `tuple<int, text> ` ends with a greater-than sign and is
// a finished type, not a comparison waiting for something to compare with.
func endsWithComparison(input string) bool {
	trimmed := strings.TrimRight(input, " ")
	if trimmed == "" || partialWord(input) != "" {
		return false
	}

	start := len(trimmed)
	for start > 0 && isOperator(trimmed[start-1]) {
		start--
	}

	switch {
	case start == len(trimmed):
		// Not an operator: IN takes a list of values as well.
		words := strings.Fields(strings.ToUpper(trimmed))
		return len(words) > 0 && words[len(words)-1] == "IN"
	case trimmed[start:] == "=":
		return true // written tight as often as not: `a=`
	}
	return start == 0 || trimmed[start-1] == ' '
}

// isOperator reports whether a byte is part of a comparison operator.
func isOperator(c byte) bool {
	switch c {
	case '=', '<', '>', '!':
		return true
	}
	return false
}
