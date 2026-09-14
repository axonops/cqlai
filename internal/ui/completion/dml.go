package completion

import "strings"

// Completing the statements that read and write rows.
//
// These have the same shape as the CREATE statements - a sequence of keywords
// in a fixed order, some of them optional - and are walked the same way. What
// they had before was a list of the keywords that can appear anywhere in the
// statement, offered at every position: after a column in a WHERE clause it
// offered AND, GROUP, ORDER, LIMIT and ALLOW, where what goes there is an
// operator, and it offered WHERE after LIMIT, where nothing goes at all.

// selectTail is what can follow the WHERE clause of a query, in the order the
// grammar has them.
var selectTail = []string{"GROUP BY", "ORDER BY", "PER PARTITION LIMIT", "LIMIT", "ALLOW FILTERING"}

// dmlCompletions is what to offer in a statement that reads or writes rows,
// and whether the statement is one.
func (ce *CompletionEngine) dmlCompletions(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) == 0 {
		return nil, false
	}

	switch words[0] {
	case "SELECT":
		after, _ := afterWords(input, "SELECT")
		return walkShape(after, ce.selectShape())

	case "INSERT":
		after, found := afterWords(input, "INSERT INTO")
		if !found {
			return []string{"INTO"}, len(words) > 1 || strings.HasSuffix(input, " ")
		}
		if containsWord(input, "JSON") {
			return walkShape(after, insertJSONShape)
		}
		return walkShape(after, insertShape)

	case "UPDATE":
		after, _ := afterWords(input, "UPDATE")
		return walkShape(after, updateShape)

	case "DELETE":
		after, _ := afterWords(input, "DELETE")
		return walkShape(after, deleteShape)

	case "BEGIN", "APPLY":
		return batchCompletions(input, words)
	}
	return nil, false
}

// selectShape is a query: what to read, where from, and then the clauses that
// narrow it down, each of which can be left out.
func (ce *CompletionEngine) selectShape() []step {
	return []step{
		{kind: untilWords, until: []string{"FROM"}, answer: ce.selectorOfAQuery},
		{keyword: "FROM", kind: oneWord},
		{keyword: "WHERE", kind: untilWords, until: selectTail,
			answer: conditionThenAnyOf(selectTail...), optional: true},
		{keyword: "GROUP BY", kind: untilWords, until: selectTail[1:],
			answer: namesThenAnyOf(selectTail[1:]...), optional: true},
		{keyword: "ORDER BY", kind: untilWords, until: selectTail[2:],
			answer: orderOfAQuery, optional: true},
		{keyword: "PER PARTITION LIMIT", kind: oneWord, offers: []string{Hint("rows")}, optional: true},
		{keyword: "LIMIT", kind: oneWord, offers: []string{Hint("rows")}, optional: true},
		{keyword: "ALLOW FILTERING", kind: noValue, optional: true},
	}
}

// insertShape is the columns and the values given for them.
var insertShape = []step{
	{kind: oneWord}, // the table
	{kind: inBrackets, answer: columnOfAnInsert},
	{keyword: "VALUES", kind: inBrackets, answer: valueOfAnInsert},
	{keyword: "IF NOT EXISTS", kind: noValue, optional: true},
	{keyword: "USING", kind: untilWords, until: []string{";"}, answer: usingOfAWrite, optional: true},
}

// insertJSONShape is the same row given as one document.
var insertJSONShape = []step{
	{kind: oneWord},
	{keyword: "JSON", kind: quotedText, offers: []string{Hint("'json'")}},
	{keyword: "DEFAULT", kind: oneWord, offers: []string{"NULL", "UNSET"}, optional: true},
	{keyword: "IF NOT EXISTS", kind: noValue, optional: true},
	{keyword: "USING", kind: untilWords, until: []string{";"}, answer: usingOfAWrite, optional: true},
}

var updateShape = []step{
	{kind: oneWord}, // the table
	{keyword: "USING", kind: untilWords, until: []string{"SET"},
		answer: usingThenAnyOf("SET"), optional: true},
	{keyword: "SET", kind: untilWords, until: []string{"WHERE"}, answer: assignmentOfAnUpdate},
	{keyword: "WHERE", kind: untilWords, until: []string{"IF"}, answer: conditionThenAnyOf("IF")},
	{keyword: "IF", kind: untilWords, until: []string{";"}, answer: conditionOfAWrite, optional: true},
}

var deleteShape = []step{
	{kind: untilWords, until: []string{"FROM"}, answer: selectionOfADelete},
	{keyword: "FROM", kind: oneWord}, // the table
	{keyword: "USING TIMESTAMP", kind: oneWord, offers: []string{Hint("microseconds")}, optional: true},
	{keyword: "WHERE", kind: untilWords, until: []string{"IF"}, answer: conditionThenAnyOf("IF")},
	{keyword: "IF", kind: untilWords, until: []string{";"}, answer: conditionOfAWrite, optional: true},
}

// selectorOfAQuery is what a query reads: the columns, or everything.
//
// With nothing typed the answer is the one the rest of this package gives,
// which knows the functions as well as the columns.
func (ce *CompletionEngine) selectorOfAQuery(typed string) []string {
	if strings.TrimSpace(typed) == "" {
		return nil
	}

	// JSON and DISTINCT come before what is being read, not instead of it.
	if onlyModifiers(typed) {
		return without(ce.getSelectCompletions([]string{"SELECT"}, 1), finishedWords(typed))
	}
	if strings.HasSuffix(strings.TrimRight(typed, " "), ",") {
		return []string{columnHint}
	}
	return []string{"FROM"}
}

// onlyModifiers reports whether all that has been written is the words that
// come before the columns.
func onlyModifiers(typed string) bool {
	for _, word := range finishedWords(typed) {
		switch strings.ToUpper(word) {
		case "JSON", "DISTINCT":
		default:
			return false
		}
	}
	return true
}

// without is the suggestions less the words already written.
func without(suggestions []string, written []string) []string {
	kept := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		gone := false
		for _, word := range written {
			gone = gone || strings.EqualFold(suggestion, word)
		}
		if !gone {
			kept = append(kept, suggestion)
		}
	}
	return kept
}

// conditionThenAnyOf is a WHERE clause: a column, an operator, a value, and
// then either another condition or whatever the clause runs into.
func conditionThenAnyOf(next ...string) func(string) []string {
	return func(typed string) []string {
		if being := conditionBeingWritten(typed); being != nil {
			return being
		}
		return append([]string{"AND"}, next...)
	}
}

// conditionBeingWritten is what the condition under the cursor wants next, or
// nil when it is finished.
func conditionBeingWritten(typed string) []string {
	condition := typed
	if after, found := afterWords(typed, "AND"); found {
		condition = after
	}

	// Inside a bracket: the columns of a TOKEN, or the values of an IN.
	if open := strings.LastIndex(condition, "("); open >= 0 && !strings.Contains(condition[open:], ")") {
		if containsWord(condition, "IN") {
			return []string{valueHint}
		}
		return []string{columnHint}
	}

	words := finishedWords(condition)
	if len(words) == 0 {
		return []string{columnHint}
	}

	switch strings.ToUpper(words[len(words)-1]) {
	case "CONTAINS":
		return append([]string{"KEY"}, valueHint)
	case "NOT":
		return []string{"IN", "CONTAINS"}
	case "IS":
		return []string{"NOT NULL"}
	case "BETWEEN":
		return []string{valueHint}
	case "AND":
		return []string{valueHint} // the second half of a BETWEEN
	case "KEY", "IN", "LIKE":
		return []string{valueHint}
	}

	// A column on its own is waiting for the operator that compares it.
	if len(words) == 1 && !endsWithComparison(condition) {
		return comparisons()
	}
	if endsWithComparison(condition) {
		return []string{valueHint}
	}

	// A BETWEEN needs its second value, and the AND that joins them, before
	// the condition is finished.
	if containsWord(condition, "BETWEEN") && !containsWord(condition, "AND") {
		return []string{"AND"}
	}
	return nil
}

// conditionOfAWrite is what a write is conditional on: that the row exists, or
// that a column of it holds what it is expected to.
func conditionOfAWrite(typed string) []string {
	words := finishedWords(typed)
	if len(words) == 0 {
		return []string{"EXISTS", columnHint}
	}
	if strings.EqualFold(words[0], "EXISTS") {
		return []string{} // answered, and the answer is that nothing follows
	}
	return conditionBeingWritten(typed)
}

// comparisons are the operators a condition can use.
func comparisons() []string {
	return append(append([]string{}, ComparisonOperators...), "LIKE", "IS NOT NULL")
}

// namesThenAnyOf is a list of column names, and then whatever follows it.
func namesThenAnyOf(next ...string) func(string) []string {
	return func(typed string) []string {
		if len(finishedWords(typed)) == 0 || strings.HasSuffix(strings.TrimRight(typed, " "), ",") {
			return []string{columnHint}
		}
		return next
	}
}

// orderOfAQuery is a column and the direction to read it in.
func orderOfAQuery(typed string) []string {
	column := typed
	if comma := strings.LastIndex(typed, ","); comma >= 0 {
		column = typed[comma+1:]
	}

	words := finishedWords(column)
	switch len(words) {
	case 0:
		return []string{columnHint}
	case 1:
		// A vector column is read in order of how near it is to a value.
		return []string{"ASC", "DESC", "ANN OF"}
	}

	if strings.EqualFold(words[len(words)-1], "OF") {
		return []string{valueHint}
	}
	if strings.EqualFold(words[1], "ANN") && len(words) == 3 {
		return []string{"ASC", "DESC"}
	}
	return selectTail[2:]
}

// usingOfAWrite is the TTL and the timestamp a write can be given.
func usingOfAWrite(typed string) []string {
	return usingThenAnyOf()(typed)
}

// usingThenAnyOf is a USING clause, and then whatever follows it.
func usingThenAnyOf(next ...string) func(string) []string {
	return func(typed string) []string {
		option := typed
		if after, found := afterWords(typed, "AND"); found {
			option = after
		}

		words := finishedWords(option)
		switch len(words) {
		case 0:
			return []string{"TTL", "TIMESTAMP"}
		case 1:
			if strings.EqualFold(words[0], "TTL") {
				return []string{Hint("seconds")}
			}
			return []string{Hint("microseconds")}
		}
		return append([]string{"AND"}, next...)
	}
}

// columnOfAnInsert is a column of the table the row is going into.
func columnOfAnInsert(typed string) []string {
	if len(finishedWords(typed)) == 0 {
		return []string{columnHint}
	}
	return []string{")"}
}

// valueOfAnInsert is one of the values given for those columns.
func valueOfAnInsert(typed string) []string {
	if len(finishedWords(typed)) == 0 {
		return []string{valueHint}
	}
	return []string{")"}
}

// selectionOfADelete is the columns a delete removes, which it can leave out
// to remove the whole row.
func selectionOfADelete(typed string) []string {
	if strings.HasSuffix(strings.TrimRight(typed, " "), ",") {
		return []string{columnHint}
	}
	if len(finishedWords(typed)) == 0 {
		return []string{"FROM", columnHint}
	}
	return []string{"FROM"}
}

// assignmentOfAnUpdate is a column and what to set it to.
func assignmentOfAnUpdate(typed string) []string {
	assignment := typed
	if comma := strings.LastIndex(typed, ","); comma >= 0 {
		assignment = typed[comma+1:]
	}

	if len(finishedWords(assignment)) == 0 {
		return []string{columnHint}
	}
	if endsWithComparison(assignment) {
		return []string{valueHint}
	}
	if len(finishedWords(assignment)) == 1 {
		// A collection is added to and taken from as well as set.
		return []string{"= ", "+= ", "-= "}
	}
	return []string{"WHERE"}
}

// batchCompletions is what to offer in a batch: the statements it is made of,
// and the words that open and close it.
func batchCompletions(input string, words []string) ([]string, bool) {
	if words[0] == "APPLY" {
		return []string{"BATCH"}, true
	}

	switch {
	case len(words) == 1:
		return []string{"BATCH", "UNLOGGED", "COUNTER"}, strings.HasSuffix(input, " ")
	case !containsWord(input, "BATCH"):
		return []string{"BATCH"}, true
	}

	// Inside the batch: each statement stands on its own, and the one being
	// typed is the text after the last semicolon.
	after, _ := afterWords(input, "BATCH")
	if semicolon := strings.LastIndex(after, ";"); semicolon >= 0 {
		after = after[semicolon+1:]
	}

	if containsWord(after, "USING") && !strings.Contains(after, ";") {
		using, _ := afterWords(after, "USING")
		return usingThenAnyOf("INSERT", "UPDATE", "DELETE")(using), true
	}
	if strings.TrimSpace(after) != "" {
		return nil, false // a statement of its own, answered as one
	}
	if !containsWord(input, "USING") && !strings.Contains(input, ";") {
		return []string{"USING TIMESTAMP", "INSERT", "UPDATE", "DELETE", "APPLY BATCH"}, true
	}
	return []string{"INSERT", "UPDATE", "DELETE", "APPLY BATCH"}, true
}

// statementInABatch is the statement being typed inside a batch, and whether
// the text is one.
//
// A batch holds whole statements, and each of them completes the way it does
// on its own: what comes after the last semicolon is a statement being typed.
func statementInABatch(input string) (string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) == 0 || words[0] != "BEGIN" || !containsWord(input, "BATCH") {
		return "", false
	}

	after, _ := afterWords(input, "BATCH")
	semicolon := strings.LastIndex(after, ";")
	if semicolon < 0 {
		return "", false
	}

	typed := after[semicolon+1:]
	if strings.TrimSpace(typed) == "" {
		return "", false
	}
	return typed, true
}
