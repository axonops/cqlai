package completion

import "strings"

// Completing the CREATE statements that are a sequence of words.
//
// A trigger, a type, a function, an aggregate and a materialized view are each
// written in a fixed order - a name, a keyword, what the keyword takes, the
// next keyword - and completion offered none of it. Everything after the name
// was silence, which for a function means eight keywords in a row that have to
// be remembered exactly.
//
// The order is written down here once, as the steps a statement is made of, and
// one walk through the steps answers for all of them.

// valueKind is how to tell that a step's value has been given.
type valueKind int

const (
	noValue    valueKind = iota // the keyword stands on its own
	oneWord                     // a name, or a type
	inBrackets                  // a group in brackets
	quotedText                  // a string in quotes
	untilWords                  // anything, up to the words that end it
	toTheEnd                    // anything at all: the statement ends here
)

// step is one part of a statement: the keyword that introduces it, and what
// that keyword takes.
type step struct {
	keyword  string
	kind     valueKind
	until    []string // for untilWords: whatever ends the part, first one wins
	offers   []string // what to offer for the value
	optional bool     // a part the statement can do without

	// answer is what to offer inside a bracketed part, or partway through one
	// that runs until a keyword. It is given what has been typed of it.
	answer func(typed string) []string
}

// createShapes are the statements this walks, by the word that names them.
var createShapes = map[string][]step{
	"TRIGGER": {
		{kind: oneWord},
		{keyword: "ON", kind: oneWord},
		{keyword: "USING", kind: quotedText, offers: []string{Hint("'trigger class'")}},
	},
	"TYPE": {
		{kind: oneWord},
		{kind: inBrackets, answer: fieldInAType},
	},
	"FUNCTION": {
		{kind: oneWord},
		{kind: inBrackets, answer: argumentOfAFunction},
		{kind: untilWords, until: []string{"ON NULL INPUT"}, answer: nullInputOfAFunction},
		{keyword: "RETURNS", kind: oneWord, offers: CQLDataTypes},
		{keyword: "LANGUAGE", kind: oneWord, offers: UDFLanguages},
		{keyword: "AS", kind: toTheEnd, offers: []string{Hint("$$ body $$")}},
	},
	"AGGREGATE": {
		{kind: oneWord},
		{kind: inBrackets, answer: argumentOfAnAggregate},
		{keyword: "SFUNC", kind: oneWord, offers: []string{Hint("function name")}},
		{keyword: "STYPE", kind: oneWord, offers: CQLDataTypes},
		{keyword: "FINALFUNC", kind: oneWord, optional: true, offers: []string{Hint("function name")}},
		{keyword: "INITCOND", kind: toTheEnd, optional: true, offers: []string{Hint("value")}},
	},
	"VIEW": {
		{kind: oneWord},
		{keyword: "AS", kind: noValue},
		{keyword: "SELECT", kind: untilWords, until: []string{"FROM"}, answer: selectorOfAView},
		{keyword: "FROM", kind: oneWord}, // the table it selects from
		{keyword: "WHERE", kind: untilWords, until: []string{"PRIMARY KEY"}, answer: conditionOfAView},
		{keyword: "PRIMARY KEY", kind: inBrackets, answer: columnOfAViewKey},
		{keyword: "WITH", kind: toTheEnd, optional: true},
	},
}

// createStatementCompletions is what to offer in a CREATE statement written as
// a sequence of words, and whether the statement is one of them.
func createStatementCompletions(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) < 2 || words[0] != "CREATE" {
		return nil, false
	}

	// CREATE OR REPLACE, before the word it replaces.
	if words[1] == "OR" {
		if len(words) == 2 {
			return []string{"REPLACE"}, true
		}
		if len(words) == 3 && words[2] == "REPLACE" && strings.HasSuffix(input, " ") {
			return []string{"FUNCTION", "AGGREGATE"}, true
		}
	}

	// MATERIALIZED VIEW is named in two words, and is walked as VIEW.
	if words[1] == "MATERIALIZED" && len(words) == 2 {
		return []string{"VIEW"}, true
	}

	for name, shape := range createShapes {
		// The word has to be the one that says what the CREATE makes, near the
		// front, rather than the same word turning up inside the statement: a
		// comment can say anything at all.
		if !startsTheStatement(words, name) {
			continue
		}
		after, found := afterWords(input, name)
		if !found {
			continue
		}
		return walkShape(skipIfNotExists(after), shape)
	}
	return nil, false
}

// startsTheStatement reports whether this is the word that says what the CREATE
// makes. It comes after CREATE, and after at most OR REPLACE or MATERIALIZED.
func startsTheStatement(words []string, name string) bool {
	for i, word := range words {
		if i > 3 {
			return false
		}
		if word == name {
			return true
		}
	}
	return false
}

// skipIfNotExists steps over the IF NOT EXISTS that can come before the name.
func skipIfNotExists(input string) string {
	trimmed := strings.TrimLeft(input, " ")
	for _, word := range []string{"IF", "NOT", "EXISTS"} {
		after, found := afterWords(trimmed, word)
		if !found {
			break
		}
		trimmed = strings.TrimLeft(after, " ")
	}
	return " " + trimmed
}

// walkShape is what to offer at the end of a statement written in this shape.
//
// The walk takes the text apart step by step: each keyword is found and each
// value is stepped over, until it reaches the one being typed. Whatever is
// left when the text runs out is what the statement wants next.
func walkShape(input string, shape []step) ([]string, bool) {
	rest := input

	for i, part := range shape {
		if part.keyword != "" {
			after, found := afterWords(rest, part.keyword)
			if !found {
				if part.optional && laterKeyword(rest, shape[i+1:]) {
					continue
				}
				wanted := keywordsFrom(shape[i:])
				// Part of the keyword may be typed already, and offering the
				// whole of it again would write it twice.
				if left := restOfAPhrase(rest, wanted...); len(left) > 0 {
					return left, true
				}
				return wanted, true
			}
			rest = after
		}

		switch part.kind {
		case noValue:
			continue

		case inBrackets:
			open := strings.Index(rest, "(")
			if open < 0 {
				return []string{"("}, true
			}
			closed := closingBracket(rest, open)
			if closed < 0 {
				suggestions := part.answer(itemBeingTyped(rest[open+1:]))
				return suggestions, suggestions != nil
			}
			rest = rest[closed+1:]

		case untilWords:
			// The words that end the part are left where they are: they
			// introduce the part that follows, and that part reads them.
			after, found := beforeAnyWords(rest, part.until)
			if !found {
				if part.answer == nil {
					return part.offers, len(part.offers) > 0
				}
				// An answer of nothing is not an answer: the statement is left
				// to whatever else can complete it.
				suggestions := part.answer(rest)
				return suggestions, suggestions != nil
			}
			rest = after

		case toTheEnd:
			return part.offers, len(part.offers) > 0

		default: // oneWord and quotedText
			after, given := valueGiven(rest, part.kind)
			if !given {
				return part.offers, len(part.offers) > 0
			}
			rest = after
		}
	}

	// Every part is in place: the statement is finished, and saying so is an
	// answer of its own.
	return nil, true
}

// beforeAnyWords is the text from whichever of these comes first, the words
// themselves included.
func beforeAnyWords(input string, words []string) (from string, found bool) {
	longest := ""
	for _, phrase := range words {
		rest, ok := afterWords(input, phrase)
		if !ok {
			continue
		}
		// The phrase and what follows it, which is where the next part starts.
		at := len(input) - len(rest) - len(phrase)
		if at < 0 {
			at = 0
		}
		if !found || len(input[at:]) > len(longest) {
			longest, found = input[at:], true
		}
	}
	return longest, found
}

// keywordsFrom is the keyword the statement wants next, and the optional ones
// that could come instead of it.
func keywordsFrom(shape []step) []string {
	keywords := []string{shape[0].keyword}
	if !shape[0].optional {
		return keywords
	}

	// Everything that can be written instead: the optional parts that follow,
	// and the first part that is not optional, which is what the statement
	// comes to if all of them are left out.
	for _, part := range shape[1:] {
		if part.keyword == "" {
			break
		}
		keywords = append(keywords, part.keyword)
		if !part.optional {
			break
		}
	}
	return keywords
}

// laterKeyword reports whether a part further on has been written, which is
// what says an optional part was passed over.
func laterKeyword(rest string, later []step) bool {
	for _, part := range later {
		if part.keyword == "" {
			continue
		}
		if _, found := afterWords(rest, part.keyword); found {
			return true
		}
	}
	return false
}

// valueGiven reports whether a value of this kind has been written, and what
// text follows it.
func valueGiven(rest string, kind valueKind) (after string, given bool) {
	if kind == quotedText {
		open := strings.IndexAny(rest, "'\"")
		if open < 0 {
			return rest, false
		}
		closed := strings.IndexAny(rest[open+1:], "'\"")
		if closed < 0 {
			return rest, false
		}
		return rest[open+1+closed+1:], true
	}

	// One word, which is finished once something that is not part of it has
	// been typed: a space, or the bracket of a function's arguments.
	trimmed := strings.TrimLeft(rest, " ")
	end := strings.IndexAny(trimmed, " (")
	if end < 0 {
		return rest, false
	}
	return trimmed[end:], true
}

// afterWords is the text after these words, and whether they are there.
//
// The words are matched as words - `AS` is not the end of `ALIAS` - and the
// match is the first one, since these are the keywords of a statement written
// in order.
func afterWords(input, words string) (after string, found bool) {
	wanted := strings.Fields(strings.ToUpper(words))
	if len(wanted) == 0 {
		return input, true
	}

	upper := strings.ToUpper(input)
	at := 0
	for {
		found := strings.Index(upper[at:], wanted[0])
		if found < 0 {
			return input, false
		}
		start := at + found
		end := start + len(wanted[0])
		at = end

		if !standsAlone(upper, start, end) {
			continue
		}

		// The words after the first, each in turn.
		matched := true
		for _, word := range wanted[1:] {
			next := strings.TrimLeft(upper[end:], " ")
			if !strings.HasPrefix(next, word) {
				matched = false
				break
			}
			end = len(upper) - len(next) + len(word)
			if !standsAlone(upper, end-len(word), end) {
				matched = false
				break
			}
		}
		if matched {
			return input[end:], true
		}
	}
}

// standsAlone reports whether the text between these two points is a word of
// its own rather than part of a longer one.
func standsAlone(text string, start, end int) bool {
	if start > 0 && isWordCharacter(text[start-1]) {
		return false
	}
	return end >= len(text) || !isWordCharacter(text[end])
}

// closingBracket is where the bracket opened at this point closes, or -1 while
// it is still open.
func closingBracket(text string, open int) int {
	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// itemBeingTyped is what has been typed since the comma that started it.
func itemBeingTyped(inside string) string {
	depth := 0
	start := 0

	for i := 0; i < len(inside); i++ {
		switch inside[i] {
		case '(', '<':
			depth++
		case ')', '>':
			depth--
		case ',':
			if depth == 0 {
				start = i + 1
			}
		}
	}
	return inside[start:]
}

// lastWordOf is the last finished word of the text, or "".
func lastWordOf(input string) string {
	words := finishedWords(input)
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

// What goes inside the brackets of each statement.

// fieldInAType: a name of your own, and then its type.
func fieldInAType(typed string) []string {
	return nameThenType(typed, Hint("field name"))
}

// nameThenType is a name of your own followed by its type, which is how a
// user defined type's fields and a function's arguments are both written.
func nameThenType(typed, name string) []string {
	// A type with a parameter is not finished until its angle bracket closes,
	// and what goes inside it is another type.
	if strings.Count(typed, "<") > strings.Count(typed, ">") {
		return CQLDataTypes
	}

	switch len(finishedWords(withoutTypeParameters(typed))) {
	case 0:
		return []string{name}
	case 1:
		return CQLDataTypes
	}
	return []string{")"}
}

// nullInputOfAFunction is what a function says about being given null: one of
// the two phrases, or the rest of the one being written.
func nullInputOfAFunction(typed string) []string {
	return restOfAPhrase(typed, "CALLED ON NULL INPUT", "RETURNS NULL ON NULL INPUT")
}

// restOfAPhrase is what is left of each phrase that carries on from the words
// already typed.
//
// Offering the whole phrase again once part of it is written puts it in twice:
// `CALLED CALLED ON NULL INPUT`.
func restOfAPhrase(typed string, phrases ...string) []string {
	written := upper(finishedWords(typed))

	left := make([]string, 0, len(phrases))
	for _, phrase := range phrases {
		words := strings.Fields(phrase)
		if len(written) > len(words) || !sameWords(words[:len(written)], written) {
			continue
		}
		if rest := strings.Join(words[len(written):], " "); rest != "" {
			left = append(left, rest)
		}
	}
	return left
}

// sameWords reports whether two runs of words are the same.
func sameWords(one, other []string) bool {
	if len(one) != len(other) {
		return false
	}
	for i := range one {
		if !strings.EqualFold(one[i], other[i]) {
			return false
		}
	}
	return true
}

// argumentOfAFunction: a name of your own, and then its type.
func argumentOfAFunction(typed string) []string {
	return nameThenType(typed, Hint("argument name"))
}

// argumentOfAnAggregate: the types it takes, without names.
func argumentOfAnAggregate(typed string) []string {
	if len(finishedWords(withoutTypeParameters(typed))) == 0 {
		return CQLDataTypes
	}
	return []string{")"}
}

// selectorOfAView: the columns it takes from the table it is a view of.
func selectorOfAView(typed string) []string {
	if strings.TrimSpace(typed) == "" {
		return []string{"*", columnHint}
	}
	if strings.HasSuffix(strings.TrimRight(typed, " "), ",") {
		return []string{columnHint}
	}
	return []string{"FROM"}
}

// conditionOfAView: every column of its key has to be known not to be null.
func conditionOfAView(typed string) []string {
	condition := typed
	if after, found := afterWords(typed, "AND"); found {
		condition = after
	}

	switch strings.ToUpper(lastWordOf(condition)) {
	case "":
		return []string{columnHint}
	case "IS":
		return []string{"NOT NULL"}
	case "NOT":
		return []string{"NULL"}
	case "NULL":
		return []string{"AND", "PRIMARY KEY"}
	case "PRIMARY":
		return []string{"KEY"}
	}
	return []string{"IS NOT NULL"}
}

// columnOfAViewKey: the columns of the key, which the view names rather than
// declares.
func columnOfAViewKey(typed string) []string {
	if len(finishedWords(typed)) == 0 {
		return []string{columnHint}
	}
	return []string{")"}
}
