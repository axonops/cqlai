package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Classifying a statement: which CQL command it is, what kind of thing that
// command does, and which keyspaces and tables it names.
//
// The MCP server runs statements a model wrote, and a model can be steered by
// text it read. So what may run is decided here, from the statement itself,
// and not from what the model says the statement is. Anything this cannot read
// with certainty is refused: a statement it does not recognise, a character it
// does not expect, a name it cannot find. Being wrong in the other direction -
// letting through something it misread - is the failure that matters.

// Kind is what a command does.
type Kind int

const (
	// KindNever is a statement that is never run for a model: GRANT, REVOKE,
	// roles and users, USE, and anything not recognised. It is the zero value,
	// so a statement nothing classified is refused.
	KindNever Kind = iota
	KindRead
	KindWrite
	KindSchema
)

func (k Kind) String() string {
	switch k {
	case KindRead:
		return "read"
	case KindWrite:
		return "write"
	case KindSchema:
		return "schema"
	}
	return "never"
}

// Command is one CQL command that can be permitted or not.
type Command struct {
	Name string
	Kind Kind
	// Default is whether it is permitted when nothing says otherwise.
	Default bool
	// Covers is what permitting it allows, in words, for the preferences
	// window.
	Covers string
}

// Commands is every command that can be permitted. The statement gate
// classifies by it and the preferences window draws its rows from it, so the
// window cannot offer a command the gate does not know.
var Commands = []Command{
	{Name: "SELECT", Kind: KindRead, Default: true, Covers: "read rows"},
	{Name: "DESCRIBE", Kind: KindRead, Default: true, Covers: "see the schema"},
	{Name: "LIST", Kind: KindRead, Default: true, Covers: "list roles, users and permissions"},
	{Name: "INSERT", Kind: KindWrite, Covers: "add rows"},
	{Name: "UPDATE", Kind: KindWrite, Covers: "change rows"},
	{Name: "DELETE", Kind: KindWrite, Covers: "remove rows"},
	{Name: "BATCH", Kind: KindWrite, Covers: "group writes, each of which has to be permitted too"},
	{Name: "CREATE", Kind: KindSchema, Covers: "create keyspaces, tables, indexes, types, functions, views"},
	{Name: "ALTER", Kind: KindSchema, Covers: "change keyspaces, tables, types, views"},
	{Name: "DROP", Kind: KindSchema, Covers: "drop any of those"},
	{Name: "TRUNCATE", Kind: KindSchema, Covers: "remove every row of a table"},
}

// NeverPermitted is the statements no setting permits, in words.
const NeverPermitted = "GRANT, REVOKE, roles, users and USE are never permitted"

// CommandNamed is the command of this name, ignoring case.
func CommandNamed(name string) (Command, bool) {
	for _, c := range Commands {
		if strings.EqualFold(c.Name, strings.TrimSpace(name)) {
			return c, true
		}
	}
	return Command{}, false
}

// DefaultCommands is the names of the commands permitted when nothing says
// otherwise.
func DefaultCommands() []string {
	var names []string
	for _, c := range Commands {
		if c.Default {
			names = append(names, c.Name)
		}
	}
	return names
}

// Name is a keyspace, or a table (or other object) in one, as a statement
// names it. Unquoted names are folded to lower case and quoted ones are kept as
// written, which is how Cassandra resolves them.
type Name struct {
	Keyspace string
	Table    string // empty when the statement names a keyspace alone
	// Qualified says the statement named the keyspace. An unqualified table
	// is resolved against whatever keyspace the session is in.
	Qualified bool
}

// Statement is what Classify found.
type Statement struct {
	Command string // "SELECT", "DROP", ...; empty for KindNever
	Kind    Kind
	Object  string // for CREATE, ALTER and DROP: "TABLE", "KEYSPACE", ...
	Names   []Name // every keyspace and table the statement names

	// For a SELECT.
	AllowFiltering bool
	Aggregate      bool     // COUNT, SUM, AVG, MIN or MAX in the selection
	Restricted     []string // columns the WHERE clause fixes with = or IN

	// PlainSelection says the SELECT returns columns under their own names:
	// * or a list of column names, with no JSON, alias or function. Only then
	// does a result column's name say which column its values came from.
	PlainSelection bool
	// WhereNames is every name the WHERE clause mentions. A condition on a
	// column can test guesses at its values without returning them.
	WhereNames []string

	// Inner is a BATCH's statements.
	Inner []Statement

	// What a change does, for saying what it will do before anyone runs it.
	Conditional bool     // IF: a lightweight transaction
	TTL         bool     // USING TTL: what is written expires
	Action      string   // ALTER TABLE: ADD, DROP, ALTER, RENAME or WITH; ALTER KEYSPACE: WITH
	Options     []string // the options a WITH clause sets: replication, compaction, ...
	Columns     bool     // DELETE names columns: it deletes values, not rows
	BatchType   string   // BATCH: LOGGED, UNLOGGED or COUNTER
}

// Classify reads one CQL statement. The error says why it cannot be: more
// than one statement, one not recognised, one not read with certainty.
func Classify(text string) (Statement, error) {
	tokens, err := lex(text)
	if err != nil {
		return Statement{}, err
	}

	statements := splitTokens(tokens)
	switch {
	case len(statements) == 0:
		return Statement{}, fmt.Errorf("there is no statement")
	case len(statements) > 1 && !isWord(statements[0], 0, "begin"):
		return Statement{}, fmt.Errorf("one statement at a time: this has %d", len(statements))
	}

	if isWord(tokens, 0, "begin") {
		return classifyBatch(tokens)
	}
	return classifyOne(statements[0])
}

// classifyOne classifies a statement that is not a BATCH.
func classifyOne(t []token) (Statement, error) {
	if len(t) == 0 || t[0].kind != tokWord {
		return Statement{}, fmt.Errorf("the statement does not start with a CQL command")
	}

	switch t[0].text {
	case "select":
		return classifySelect(t)
	case "insert":
		if !isWord(t, 1, "into") {
			return Statement{}, fmt.Errorf("INSERT has to be followed by INTO")
		}
		name, next, err := readName(t, 2)
		if err != nil {
			return Statement{}, err
		}
		s := Statement{Command: "INSERT", Kind: KindWrite, Names: []Name{name}}
		s.Conditional = findWord(t, next, "exists") >= 0 && findWord(t, next, "if") >= 0
		s.TTL = usesTTL(t[next:])
		return s, nil
	case "update":
		name, next, err := readName(t, 1)
		if err != nil {
			return Statement{}, err
		}
		s := Statement{Command: "UPDATE", Kind: KindWrite, Names: []Name{name}}
		s.TTL = usesTTL(t[next:])
		addWhere(&s, t, next)
		return s, nil
	case "delete":
		from := findWord(t, 1, "from")
		if from < 0 {
			return Statement{}, fmt.Errorf("DELETE has to name a table with FROM")
		}
		name, next, err := readName(t, from+1)
		if err != nil {
			return Statement{}, err
		}
		s := Statement{Command: "DELETE", Kind: KindWrite, Names: []Name{name}}
		s.Columns = from > 1
		addWhere(&s, t, next)
		return s, nil
	case "truncate":
		i := 1
		if isWord(t, i, "table") || isWord(t, i, "columnfamily") {
			i++
		}
		name, _, err := readName(t, i)
		if err != nil {
			return Statement{}, err
		}
		return Statement{Command: "TRUNCATE", Kind: KindSchema, Object: "TABLE", Names: []Name{name}}, nil
	case "create", "alter", "drop":
		return classifyDDL(t)
	case "describe", "desc":
		return Statement{Command: "DESCRIBE", Kind: KindRead}, nil
	case "list":
		return Statement{Command: "LIST", Kind: KindRead}, nil
	case "grant", "revoke":
		return Statement{}, fmt.Errorf("%s is never permitted", strings.ToUpper(t[0].text))
	case "use":
		return Statement{}, fmt.Errorf("USE is never permitted: name the keyspace in the statement, as keyspace.table")
	}
	return Statement{}, fmt.Errorf("%s is not a CQL command this server runs", strings.ToUpper(t[0].text))
}

// classifySelect reads a SELECT: its table, whether it filters or aggregates,
// and which columns its WHERE clause fixes.
func classifySelect(t []token) (Statement, error) {
	from := findWord(t, 1, "from")
	if from < 0 {
		return Statement{}, fmt.Errorf("SELECT has to name a table with FROM")
	}

	s := Statement{Command: "SELECT", Kind: KindRead}
	for i := 1; i < from; i++ {
		if t[i].kind == tokWord && aggregates[t[i].text] && isPunct(t, i+1, "(") {
			s.Aggregate = true
		}
	}
	s.PlainSelection = plainSelection(t[1:from])

	name, next, err := readName(t, from+1)
	if err != nil {
		return Statement{}, err
	}
	s.Names = []Name{name}

	for i := next; i < len(t); i++ {
		if isWord(t, i, "allow") && isWord(t, i+1, "filtering") {
			s.AllowFiltering = true
		}
	}

	if where := findWord(t, next, "where"); where >= 0 {
		s.Restricted = restrictedColumns(t[where+1:])
		s.WhereNames = whereNames(t[where+1:])
	}
	return s, nil
}

// plainSelection reports whether a selection is * or column names separated
// by commas, optionally after DISTINCT.
func plainSelection(t []token) bool {
	if isWord(t, 0, "distinct") {
		t = t[1:]
	}
	if len(t) == 1 && isPunct(t, 0, "*") {
		return true
	}
	if len(t) == 0 {
		return false
	}
	for i, tok := range t {
		if i%2 == 1 {
			if !isPunct(t, i, ",") {
				return false
			}
			continue
		}
		if tok.kind != tokQuoted && (tok.kind != tokWord || tok.text == "json") {
			return false
		}
	}
	return len(t)%2 == 1
}

// whereNames is every name in a WHERE clause up to where it ends, inside
// parentheses or not. A function name or a keyword is included too: more
// names than there are columns can only refuse more.
func whereNames(t []token) []string {
	var names []string
	depth := 0
	for i, tok := range t {
		switch {
		case isPunct(t, i, "("):
			depth++
		case isPunct(t, i, ")"):
			depth--
		case depth == 0 && tok.kind == tokWord && clauseEnds[tok.text]:
			return names
		case tok.kind == tokWord || tok.kind == tokQuoted:
			names = append(names, tok.text)
		}
	}
	return names
}

// aggregates are the functions that read every row they are given before they
// return anything.
var aggregates = map[string]bool{"count": true, "sum": true, "avg": true, "min": true, "max": true}

// restrictedColumns is the columns a WHERE clause fixes with = or IN, at the
// top level. A column inside parentheses - token(id), a tuple - is not counted:
// those are ranges or are not read here with certainty, and leaving a column
// out can only make a scan look like one.
func restrictedColumns(t []token) []string {
	var columns []string
	depth := 0
	for i := 0; i < len(t); i++ {
		switch {
		case isPunct(t, i, "("):
			depth++
			continue
		case isPunct(t, i, ")"):
			depth--
			continue
		}
		if depth != 0 {
			continue
		}
		if t[i].kind == tokWord && clauseEnds[t[i].text] {
			break
		}
		if t[i].kind != tokWord && t[i].kind != tokQuoted {
			continue
		}
		if isPunct(t, i+1, "=") || isWord(t, i+1, "in") {
			columns = append(columns, t[i].text)
		}
	}
	return columns
}

// clauseEnds are the words that end a WHERE clause.
var clauseEnds = map[string]bool{"group": true, "order": true, "limit": true, "allow": true, "per": true, "if": true}

// addWhere reads the WHERE clause of an UPDATE or a DELETE, and its IF.
func addWhere(s *Statement, t []token, from int) {
	where := findWord(t, from, "where")
	if where < 0 {
		return
	}
	s.Restricted = restrictedColumns(t[where+1:])
	s.WhereNames = whereNames(t[where+1:])
	s.Conditional = topLevelWord(t, where, "if")
}

// usesTTL reports whether a statement sets a time to live.
func usesTTL(t []token) bool {
	for i := range t {
		if isWord(t, i, "ttl") && (isWord(t, i-1, "using") || isWord(t, i-1, "and")) {
			return true
		}
	}
	return false
}

// topLevelWord reports whether a word is in the statement outside
// parentheses, brackets and braces, at or after from.
func topLevelWord(t []token, from int, word string) bool {
	depth := 0
	for i := from; i < len(t); i++ {
		switch {
		case isPunct(t, i, "("), isPunct(t, i, "["), isPunct(t, i, "{"):
			depth++
		case isPunct(t, i, ")"), isPunct(t, i, "]"), isPunct(t, i, "}"):
			depth--
		case depth == 0 && isWord(t, i, word):
			return true
		}
	}
	return false
}

// withOptions is the options a WITH clause sets: the names before = outside
// any brackets, after the first WITH outside them. A map such as
// replication = {...} is one option.
func withOptions(t []token, from int) []string {
	var options []string
	depth, with := 0, false
	for i := from; i < len(t); i++ {
		switch {
		case isPunct(t, i, "("), isPunct(t, i, "["), isPunct(t, i, "{"):
			depth++
		case isPunct(t, i, ")"), isPunct(t, i, "]"), isPunct(t, i, "}"):
			depth--
		case depth != 0:
		case isWord(t, i, "with"):
			with = true
		case with && t[i].kind == tokWord && isPunct(t, i+1, "="):
			options = append(options, t[i].text)
		}
	}
	return options
}

// classifyDDL reads CREATE, ALTER and DROP: what kind of object, and its name.
func classifyDDL(t []token) (Statement, error) {
	command := strings.ToUpper(t[0].text)
	i := 1
	if command == "CREATE" && isWord(t, i, "or") && isWord(t, i+1, "replace") {
		i += 2
	}
	if command == "CREATE" && isWord(t, i, "custom") {
		i++
	}
	if i >= len(t) || t[i].kind != tokWord {
		return Statement{}, fmt.Errorf("%s has to say what it %ss", command, strings.ToLower(command))
	}

	object := t[i].text
	i++
	switch object {
	case "role", "user", "identity":
		return Statement{}, fmt.Errorf("%s %s is never permitted", command, strings.ToUpper(object))
	case "schema":
		object = "keyspace"
	case "columnfamily":
		object = "table"
	case "materialized":
		if !isWord(t, i, "view") {
			return Statement{}, fmt.Errorf("MATERIALIZED has to be followed by VIEW")
		}
		object = "materialized view"
		i++
	case "keyspace", "table", "index", "type", "function", "aggregate", "trigger":
	default:
		return Statement{}, fmt.Errorf("%s %s is not something this server runs", command, strings.ToUpper(object))
	}

	// IF NOT EXISTS, IF EXISTS.
	if isWord(t, i, "if") {
		i++
		if isWord(t, i, "not") {
			i++
		}
		if !isWord(t, i, "exists") {
			return Statement{}, fmt.Errorf("IF has to be followed by EXISTS or NOT EXISTS")
		}
		i++
	}

	s := Statement{Command: command, Kind: KindSchema, Object: strings.ToUpper(object)}
	s.Options = withOptions(t, i)

	switch {
	case object == "keyspace":
		if i >= len(t) || (t[i].kind != tokWord && t[i].kind != tokQuoted) {
			return Statement{}, fmt.Errorf("%s KEYSPACE has to name the keyspace", command)
		}
		s.Names = []Name{{Keyspace: t[i].text, Qualified: true}}
		if command == "ALTER" && isWord(t, i+1, "with") {
			s.Action = "WITH"
		}

	case object == "index" && command == "CREATE", object == "trigger":
		// CREATE INDEX [name] ON keyspace.table, and a trigger is named the
		// same way: what matters is the table it is on.
		on := findWord(t, i, "on")
		if on < 0 {
			return Statement{}, fmt.Errorf("%s %s has to name its table with ON", command, s.Object)
		}
		name, _, err := readName(t, on+1)
		if err != nil {
			return Statement{}, err
		}
		s.Names = []Name{name}

	default:
		name, next, err := readName(t, i)
		if err != nil {
			return Statement{}, err
		}
		if command == "ALTER" && next < len(t) && t[next].kind == tokWord {
			s.Action = strings.ToUpper(t[next].text)
		}
		if object == "index" {
			// DROP INDEX keyspace.index names the keyspace, not a table.
			name = Name{Keyspace: name.Keyspace, Qualified: name.Qualified}
		}
		s.Names = []Name{name}

		// A view is built from a table, which has to be visible too.
		if object == "materialized view" && command == "CREATE" {
			from := findWord(t, next, "from")
			if from < 0 {
				return Statement{}, fmt.Errorf("CREATE MATERIALIZED VIEW has to name its table with FROM")
			}
			base, _, err := readName(t, from+1)
			if err != nil {
				return Statement{}, err
			}
			s.Names = append(s.Names, base)
		}
	}
	return s, nil
}

// classifyBatch reads BEGIN [UNLOGGED | COUNTER | LOGGED] BATCH ... APPLY
// BATCH, and every statement in it.
func classifyBatch(t []token) (Statement, error) {
	i := 1
	batchType := "LOGGED"
	if isWord(t, i, "unlogged") || isWord(t, i, "counter") || isWord(t, i, "logged") {
		batchType = strings.ToUpper(t[i].text)
		i++
	}
	if !isWord(t, i, "batch") {
		return Statement{}, fmt.Errorf("BEGIN has to start a BATCH")
	}
	i++
	// USING TIMESTAMP n.
	if isWord(t, i, "using") {
		i += 3
	}

	apply := -1
	for j := i; j < len(t)-1; j++ {
		if isWord(t, j, "apply") && isWord(t, j+1, "batch") {
			apply = j
			break
		}
	}
	if apply < 0 {
		return Statement{}, fmt.Errorf("the BATCH has no APPLY BATCH")
	}
	for j := apply + 2; j < len(t); j++ {
		if !isPunct(t, j, ";") {
			return Statement{}, fmt.Errorf("one statement at a time: there is more after APPLY BATCH")
		}
	}

	batch := Statement{Command: "BATCH", Kind: KindWrite, BatchType: batchType}
	for _, inner := range splitTokens(t[i:apply]) {
		s, err := classifyOne(inner)
		if err != nil {
			return Statement{}, fmt.Errorf("in the BATCH: %w", err)
		}
		if s.Command != "INSERT" && s.Command != "UPDATE" && s.Command != "DELETE" {
			return Statement{}, fmt.Errorf("a BATCH can only hold INSERT, UPDATE and DELETE, not %s", s.Command)
		}
		batch.Inner = append(batch.Inner, s)
		batch.Names = append(batch.Names, s.Names...)
	}
	if len(batch.Inner) == 0 {
		return Statement{}, fmt.Errorf("the BATCH is empty")
	}
	return batch, nil
}

// readName reads keyspace.table, or a table alone, at t[i]. It reports the
// index after it.
func readName(t []token, i int) (Name, int, error) {
	if i >= len(t) || (t[i].kind != tokWord && t[i].kind != tokQuoted) {
		return Name{}, i, fmt.Errorf("the statement does not name its table")
	}
	if isPunct(t, i+1, ".") {
		if i+2 >= len(t) || (t[i+2].kind != tokWord && t[i+2].kind != tokQuoted) {
			return Name{}, i, fmt.Errorf("%s. has to be followed by a name", t[i].text)
		}
		return Name{Keyspace: t[i].text, Table: t[i+2].text, Qualified: true}, i + 3, nil
	}
	return Name{Table: t[i].text}, i + 1, nil
}

// Reading the statement.

type tokenKind int

const (
	tokWord   tokenKind = iota // an unquoted name or keyword, folded to lower case
	tokQuoted                  // a "quoted" name, as written
	tokString                  // a 'string' or $$string$$
	tokNumber
	tokPunct
)

type token struct {
	kind tokenKind
	text string
}

// lex splits a statement into tokens, dropping comments. It refuses anything it
// does not expect rather than guessing: a character outside ASCII that is not
// inside a string or a quoted name is one a lookalike could hide behind.
func lex(text string) ([]token, error) {
	var tokens []token
	r := []rune(text)
	for i := 0; i < len(r); {
		c := r[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++

		case c == '-' && i+1 < len(r) && r[i+1] == '-', c == '/' && i+1 < len(r) && r[i+1] == '/':
			// Cassandra ends a line comment at a carriage return as well as a
			// newline, so must this: what follows a lone \r is run.
			for i < len(r) && r[i] != '\n' && r[i] != '\r' {
				i++
			}

		case c == '/' && i+1 < len(r) && r[i+1] == '*':
			end := strings.Index(string(r[i+2:]), "*/")
			if end < 0 {
				return nil, fmt.Errorf("a /* comment is not closed")
			}
			i += 2 + len([]rune(string(r[i+2:])[:end])) + 2

		case c == '\'':
			s, n, err := closeQuote(r[i:], '\'')
			if err != nil {
				return nil, fmt.Errorf("a string is not closed")
			}
			tokens = append(tokens, token{tokString, s})
			i += n

		case c == '"':
			s, n, err := closeQuote(r[i:], '"')
			if err != nil {
				return nil, fmt.Errorf("a quoted name is not closed")
			}
			if s == "" {
				return nil, fmt.Errorf("a quoted name is empty")
			}
			tokens = append(tokens, token{tokQuoted, s})
			i += n

		case c == '$' && i+1 < len(r) && r[i+1] == '$':
			end := strings.Index(string(r[i+2:]), "$$")
			if end < 0 {
				return nil, fmt.Errorf("a $$ string is not closed")
			}
			body := string(r[i+2:])[:end]
			tokens = append(tokens, token{tokString, body})
			i += 2 + len([]rune(body)) + 2

		case isLetter(c):
			start := i
			for i < len(r) && (isLetter(r[i]) || isDigit(r[i]) || r[i] == '_') {
				i++
			}
			tokens = append(tokens, token{tokWord, strings.ToLower(string(r[start:i]))})

		case isDigit(c):
			start := i
			for i < len(r) && (isLetter(r[i]) || isDigit(r[i]) || r[i] == '_') {
				i++
			}
			tokens = append(tokens, token{tokNumber, string(r[start:i])})

		case strings.ContainsRune("(),;.=<>!*+-?:[]{}%/", c):
			if (c == '<' || c == '>' || c == '!') && i+1 < len(r) && r[i+1] == '=' {
				tokens = append(tokens, token{tokPunct, string(r[i : i+2])})
				i += 2
				continue
			}
			tokens = append(tokens, token{tokPunct, string(c)})
			i++

		default:
			return nil, fmt.Errorf("the statement has a character this server does not read: %q", c)
		}
	}
	return tokens, nil
}

// closeQuote reads a quoted run starting at r[0], where a doubled quote stands
// for one. It reports the text inside and how many runes the whole run took.
func closeQuote(r []rune, quote rune) (string, int, error) {
	var b strings.Builder
	for i := 1; i < len(r); i++ {
		if r[i] != quote {
			b.WriteRune(r[i])
			continue
		}
		if i+1 < len(r) && r[i+1] == quote {
			b.WriteRune(quote)
			i++
			continue
		}
		return b.String(), i + 1, nil
	}
	return "", 0, fmt.Errorf("not closed")
}

func isLetter(c rune) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c rune) bool  { return c >= '0' && c <= '9' }

// splitTokens splits on semicolons, dropping empty statements.
func splitTokens(t []token) [][]token {
	var statements [][]token
	start := 0
	for i := 0; i <= len(t); i++ {
		if i < len(t) && !isPunct(t, i, ";") {
			continue
		}
		if i > start {
			statements = append(statements, t[start:i])
		}
		start = i + 1
	}
	return statements
}

func isWord(t []token, i int, word string) bool {
	return i >= 0 && i < len(t) && t[i].kind == tokWord && t[i].text == word
}

func isPunct(t []token, i int, p string) bool {
	return i >= 0 && i < len(t) && t[i].kind == tokPunct && t[i].text == p
}

// findWord is the index of the first unquoted word at or after from, or -1.
func findWord(t []token, from int, word string) int {
	for i := from; i < len(t); i++ {
		if isWord(t, i, word) {
			return i
		}
	}
	return -1
}

// WithoutValues is a statement with every string and number in it replaced by
// a ?, for a log that must not hold the data: an INSERT's values, or what a
// WHERE clause looked for. A statement that cannot be read is replaced whole.
func WithoutValues(text string) string {
	tokens, err := lex(uuidLiteral.ReplaceAllString(text, "0"))
	if err != nil {
		return "(a statement that could not be read)"
	}
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		switch {
		case t.kind == tokString, t.kind == tokNumber, t.kind == tokWord && valueWords[t.text]:
			parts = append(parts, "?")
		case t.kind == tokQuoted:
			parts = append(parts, `"`+strings.ReplaceAll(t.text, `"`, `""`)+`"`)
		default:
			parts = append(parts, t.text)
		}
	}
	return strings.Join(parts, " ")
}

// uuidLiteral is a uuid written without quotes. The lexer reads it as numbers
// and words, and would keep the words: e89b, a456.
var uuidLiteral = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)

// valueWords are the values written as words.
var valueWords = map[string]bool{"true": true, "false": true, "nan": true, "infinity": true}

// WithoutStatementValues is text - an error about a statement - with the
// statement's values taken out: Cassandra's errors repeat the value they
// could not use. Values of fewer than three characters are left, as they say
// little and would take out parts of the words around them.
func WithoutStatementValues(text, statement string) string {
	values := uuidLiteral.FindAllString(statement, -1)
	if tokens, err := lex(uuidLiteral.ReplaceAllString(statement, "0")); err == nil {
		for _, t := range tokens {
			if t.kind == tokString || t.kind == tokNumber {
				values = append(values, t.text)
			}
		}
	}
	// The longest first, so a value inside another does not leave the rest.
	sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	for _, v := range values {
		if len(v) >= 3 {
			text = strings.ReplaceAll(text, v, "?")
		}
	}
	return text
}

// Limit is the LIMIT a statement ends with: the rows it returns at most, not
// PER PARTITION LIMIT, which bounds each partition and leaves the result as
// large as the table. It reports false when there is none, or it is not a
// number written in the statement.
func Limit(text string) (int, bool) {
	tokens, err := lex(text)
	if err != nil {
		return 0, false
	}
	limit, found := 0, false
	for i := 0; i+1 < len(tokens); i++ {
		t := tokens[i]
		if t.kind != tokWord || t.text != "limit" {
			continue
		}
		if i > 0 && tokens[i-1].kind == tokWord && tokens[i-1].text == "partition" {
			continue
		}
		if next := tokens[i+1]; next.kind == tokNumber {
			if n, err := strconv.Atoi(next.text); err == nil {
				limit, found = n, true
			}
		}
	}
	return limit, found
}
