package completion

import "strings"

// Completing the WITH clause of a table.
//
// The option names were in constants.go and read by nothing: a list kept up to
// date for nobody, while `CREATE TABLE ... WITH ` completed to silence. The
// three options anyone edits by hand - compaction, compression, caching - take
// maps, so stopping at the opening brace is stopping where the question starts.
//
// This works on the text as typed rather than on the word positions the rest of
// this package uses. A map is punctuation, and splitting on spaces loses it:
// `{'class':` is one word to strings.Fields, and where the cursor is inside the
// braces is the whole question here.

// mapOptions are the table options that take a map, and the keys each one takes.
var mapOptions = map[string][]string{
	"compaction":  CompactionOptions,
	"compression": CompressionOptions,
	"caching":     CachingOptions,
}

// whatGoesHere is the answer for a place a value goes.
//
// Some options are one of a few words Cassandra fixes, and those can be
// offered. The rest are a number of seconds, a fraction, a name - answers only
// the user has - and for those the note says what to type. A few are both: the
// rows cached per partition is ALL, NONE, or a number.
type whatGoesHere struct {
	values []string
	note   string
}

// suggestions is the answer as a list, the note last where there is one.
func (w whatGoesHere) suggestions() []string {
	out := make([]string, 0, len(w.values)+1)
	out = append(out, w.values...)
	if w.note != "" {
		out = append(out, w.note)
	}
	return out
}

// optionValues is what each table option can be set to, outside a map.
var optionValues = map[string]whatGoesHere{
	"additional_write_policy":     {values: quoted(SpeculativeRetryValues), note: Hint("'99p' or '50ms'")},
	"allow_auto_snapshot":         {values: BooleanValues},
	"bloom_filter_fp_chance":      {note: Hint("0.0 to 1.0")},
	"cdc":                         {values: BooleanValues},
	"comment":                     {note: Hint("'text'")},
	"crc_check_chance":            {note: Hint("0.0 to 1.0")},
	"default_time_to_live":        {note: Hint("seconds")},
	"extensions":                  {note: Hint("{'name': 0x00}")},
	"gc_grace_seconds":            {note: Hint("seconds")},
	"incremental_backups":         {values: BooleanValues},
	"max_index_interval":          {note: Hint("rows")},
	"memtable":                    {note: Hint("'configuration'")},
	"memtable_flush_period_in_ms": {note: Hint("milliseconds")},
	"min_index_interval":          {note: Hint("rows")},
	"read_repair":                 {values: quoted(ReadRepairValues)},
	"speculative_retry":           {values: quoted(SpeculativeRetryValues), note: Hint("'99p' or '50ms'")},
}

// mapValues is what each key inside a map can be set to, under `option.key`.
// Everything inside one of these maps is written as a quoted string, numbers
// included.
var mapValues = map[string]whatGoesHere{
	"caching.keys":               {values: quoted(CachingValues)},
	"caching.rows_per_partition": {values: quoted(CachingValues), note: Hint("rows")},

	"compaction.class":                          {values: quoted(CompactionStrategies)},
	"compaction.enabled":                        {values: quoted(BooleanValues)},
	"compaction.max_threshold":                  {note: Hint("sstables")},
	"compaction.min_threshold":                  {note: Hint("sstables")},
	"compaction.only_purge_repaired_tombstones": {values: quoted(BooleanValues)},
	"compaction.provide_overlapping_tombstones": {values: quoted(TombstoneValues)},
	"compaction.tombstone_compaction_interval":  {note: Hint("seconds")},
	"compaction.tombstone_threshold":            {note: Hint("0.0 to 1.0")},
	"compaction.unchecked_tombstone_compaction": {values: quoted(BooleanValues)},

	"compaction.bucket_high":            {note: Hint("number")},
	"compaction.bucket_low":             {note: Hint("number")},
	"compaction.min_sstable_size":       {note: Hint("size, like 100MiB")},
	"compaction.fanout_size":            {note: Hint("number")},
	"compaction.sstable_size_in_mb":     {note: Hint("megabytes")},
	"compaction.single_sstable_uplevel": {values: quoted(BooleanValues)},

	"compaction.compaction_window_size":                  {note: Hint("number of units")},
	"compaction.compaction_window_unit":                  {values: quoted(CompactionWindowUnits)},
	"compaction.timestamp_resolution":                    {values: quoted(TimestampResolutions)},
	"compaction.expired_sstable_check_frequency_seconds": {note: Hint("seconds")},
	"compaction.unsafe_aggressive_sstable_expiration":    {values: quoted(BooleanValues)},

	"compaction.base_shard_count":        {note: Hint("number")},
	"compaction.max_sstables_to_compact": {note: Hint("number")},
	"compaction.scaling_parameters":      {note: Hint("'T4', 'L4' or 'N'")},
	"compaction.sstable_growth":          {note: Hint("0.0 to 1.0")},
	"compaction.target_sstable_size":     {note: Hint("size, like 1GiB")},

	"compression.chunk_length_in_kb": {note: Hint("kibibytes")},
	"compression.class":              {values: quoted(Compressors)},
	"compression.enabled":            {values: quoted(BooleanValues)},
	"compression.min_compress_ratio": {note: Hint("number")},
}

// tableOptionCompletions is what to offer in the WITH clause of a CREATE TABLE
// or ALTER TABLE, or nothing when the text is not in one.
func tableOptionCompletions(input string) []string {
	// CLUSTERING ORDER BY is written in the WITH clause and is not one of the
	// options: it has its own brackets rather than a value.
	if order, inOrder := clusteringOrderCompletions(input); inOrder {
		return order
	}

	option, inside, inMap := optionUnderCursor(input)

	if !inMap {
		// Between options: after WITH, or after the AND that joins them.
		// AND joins the options of a WITH clause, and also the conditions of a
		// materialized view's WHERE, which is not one.
		if endsWithWord(input, "WITH") || (endsWithWord(input, "AND") && containsWord(input, "WITH")) {
			options := withEquals(optionsLeft(input))

			// The clustering order is written like an option and is not one.
			// It belongs to something being created: the order of a table that
			// exists cannot be altered.
			if orderCanBeGiven(input) && !containsWord(input, "CLUSTERING") {
				options = append(options, "CLUSTERING ORDER BY (")
			}
			return options
		}
		if !awaitingValue(input) {
			// An option named but not yet given its equals sign. Typing the
			// name rather than taking it from the list left the clause with
			// nothing to offer at all.
			if named := endsWithOptionName(input); named {
				return []string{"= "}
			}
			// An option that has been given its value is followed by the AND
			// that joins the next one, and so is a clustering order.
			given := lastOptionName(input) != "" || containsWord(input, "CLUSTERING")
			if strings.HasSuffix(input, " ") && given {
				return []string{"AND"}
			}
			return nil
		}
		// A map option opens its brace; the rest take a value.
		if _, takesMap := mapOptions[option]; takesMap {
			return []string{"{'"}
		}
		return optionValues[option].suggestions()
	}

	// Inside the braces. A value is being given for a key if the text ends with
	// a colon, or a colon and an opening quote.
	if key := keyAwaitingValue(input); key != "" {
		return valueInMap(option, key).suggestions()
	}
	return keysLeft(option, inside)
}

// valueInMap is what a key inside a map can be set to. A key nobody has
// written down still takes something, so say so rather than nothing.
func valueInMap(option, key string) whatGoesHere {
	if answer, known := mapValues[option+"."+key]; known {
		return answer
	}
	return whatGoesHere{note: Hint("value")}
}

// keysLeft is the keys a map can still be given.
//
// The keys already in it are not among them - offering a second `class` is
// offering a mistake - and a compaction map takes the keys of the strategy it
// names on top of the ones every strategy takes.
func keysLeft(option, inside string) []string {
	keys, takesMap := mapOptions[option]
	if !takesMap {
		return nil
	}

	all := make([]string, 0, len(keys)+8)
	all = append(all, keys...)
	if option == "compaction" {
		all = append(all, strategyOptionsFor(classIn(inside))...)
	}

	return keysLeftOf(all, inside)
}

// keysLeftOf is the keys a map can still be given: the ones it has already are
// not among them, and a map with a pair in it can be closed.
func keysLeftOf(keys []string, inside string) []string {
	written := keysIn(inside)

	left := make([]string, 0, len(keys))
	for _, key := range keys {
		if !written[key] {
			left = append(left, key)
		}
	}

	if !needsComma(inside) {
		return withColon(left, false)
	}

	// A map with a pair in it can be closed. The brace comes last: another key
	// is the likelier answer, and Tab starts at the top of the list.
	return append(withColon(left, true), "}")
}

// endsWithOptionName reports whether the text ends with the name of an option
// that has not been given its equals sign.
func endsWithOptionName(input string) bool {
	if !strings.HasSuffix(input, " ") {
		return false
	}

	fields := strings.Fields(input)
	if len(fields) == 0 {
		return false
	}

	last := strings.ToLower(fields[len(fields)-1])
	for _, option := range TableOptions {
		if option == last {
			return true
		}
	}
	return false
}

// optionsLeft is the options the clause can still be given: an option it names
// already is not one of them, since a statement takes each of them once.
func optionsLeft(input string) []string {
	given := optionsGiven(input)

	left := make([]string, 0, len(TableOptions))
	for _, option := range TableOptions {
		if !given[option] {
			left = append(left, option)
		}
	}
	return left
}

// optionsGiven are the options the clause names already.
//
// An option is a name before an equals sign, outside the braces of a map:
// inside one, `'class': 'x'` names a key rather than an option.
func optionsGiven(input string) map[string]bool {
	given := make(map[string]bool)

	fields := strings.Fields(withoutMaps(input))
	for i, field := range fields {
		// Written either way round: `comment = 'x'` or `comment='x'`.
		if name, _, assigned := strings.Cut(field, "="); assigned && name != "" {
			given[strings.ToLower(name)] = true
		}
		if strings.HasPrefix(field, "=") && i > 0 {
			given[strings.ToLower(fields[i-1])] = true
		}
	}
	return given
}

// withoutMaps is the text with what is inside any braces taken out.
func withoutMaps(input string) string {
	var outside strings.Builder
	depth := 0

	for _, letter := range input {
		switch letter {
		case '{':
			depth++
			outside.WriteRune(' ') // the option before it keeps its equals sign
		case '}':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				outside.WriteRune(letter)
			}
		}
	}
	return outside.String()
}

// needsComma reports whether the pair before this one is finished, so that the
// next key follows a comma.
func needsComma(inside string) bool {
	last := inside
	if comma := strings.LastIndex(inside, ","); comma >= 0 {
		last = inside[comma+1:]
	}

	_, value, given := strings.Cut(last, ":")
	return given && strings.TrimSpace(value) != ""
}

// strategyOptionsFor is the keys a compaction strategy takes of its own, found
// by a name written in any case.
func strategyOptionsFor(class string) []string {
	for strategy, options := range StrategyOptions {
		if strings.EqualFold(strategy, class) {
			return options
		}
	}
	return nil
}

// keysIn are the keys the map being typed already has.
func keysIn(inside string) map[string]bool {
	written := make(map[string]bool)
	for _, pair := range strings.Split(inside, ",") {
		name, _, given := strings.Cut(pair, ":")
		if given {
			written[strings.ToLower(bare(name))] = true
		}
	}
	return written
}

// classIn is the compaction strategy the map names, by its short name: it can
// be written in full, as org.apache.cassandra.db.compaction.SomeStrategy.
func classIn(inside string) string {
	for _, pair := range strings.Split(inside, ",") {
		name, value, given := strings.Cut(pair, ":")
		if !given || !strings.EqualFold(bare(name), "class") {
			continue
		}
		class := bare(value)
		if dot := strings.LastIndex(class, "."); dot >= 0 {
			class = class[dot+1:]
		}
		return class
	}
	return ""
}

// bare is a written key or value as it is meant: trimmed of the space and the
// quotes around it. The case is left as it was, since a compaction class is
// written in the case its own name has.
func bare(written string) string {
	return strings.Trim(strings.TrimSpace(written), "'")
}

// optionUnderCursor is the option the cursor is inside the map of, and whether
// it is inside one at all.
//
// Which option it belongs to is the last `name =` before the brace that is
// still open.
func optionUnderCursor(input string) (option, inside string, inMap bool) {
	open := strings.LastIndex(input, "{")
	if open < 0 || strings.Contains(input[open:], "}") {
		return lastOptionName(input), "", false
	}
	return lastOptionName(input[:open]), input[open+1:], true
}

// lastOptionName is the name in the last `name =` of the text.
func lastOptionName(input string) string {
	equals := strings.LastIndex(input, "=")
	if equals < 0 {
		return ""
	}

	fields := strings.Fields(input[:equals])
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[len(fields)-1])
}

// keyAwaitingValue is the key whose value is being typed, or "".
func keyAwaitingValue(input string) string {
	trimmed := strings.TrimRight(input, " '")
	if !strings.HasSuffix(trimmed, ":") {
		return ""
	}

	trimmed = strings.TrimSuffix(trimmed, ":")
	trimmed = strings.TrimRight(trimmed, " ")
	trimmed = strings.Trim(trimmed, "'")

	i := strings.LastIndexAny(trimmed, "{,' ")
	return strings.ToLower(strings.Trim(trimmed[i+1:], "'"))
}

// endsWithWord reports whether the text ends with a word, with nothing but
// space after it.
func endsWithWord(input, word string) bool {
	if !strings.HasSuffix(input, " ") {
		return false
	}

	fields := strings.Fields(input)
	return len(fields) > 0 && strings.EqualFold(fields[len(fields)-1], word)
}

// awaitingValue reports whether the text is where an option's value goes:
// after the equals sign, or partway through typing one.
//
// Partway through counts because the value is what is being filtered against:
// `speculative_retry = 'AL` has 'ALWAYS' to offer. A value that has been
// finished - a closed quote, or a space after it - does not.
func awaitingValue(input string) bool {
	equals := strings.LastIndex(input, "=")
	if equals < 0 {
		return false
	}

	after := input[equals+1:]
	switch {
	case strings.ContainsAny(after, "{}"):
		return false // a map, which is answered inside the braces
	case strings.TrimSpace(after) == "":
		return true // nothing typed yet
	case strings.HasSuffix(after, " "):
		return false // typed, and finished
	}
	return strings.Count(after, "'")%2 == 1 || !strings.Contains(after, "'")
}

// withEquals offers an option with the equals sign it needs, since no option is
// written without one.
func withEquals(options []string) []string {
	out := make([]string, 0, len(options))
	for _, option := range options {
		out = append(out, option+" = ")
	}
	return out
}

// quoted writes a value the way it goes into the statement, in the quotes it
// needs.
func quoted(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "'"+value+"'")
	}
	return out
}

// withColon offers a map key with the colon that takes it to its value, since
// no key is written without one. The next Tab then offers the value.
//
// A key that follows a finished pair carries the comma between them as well.
// The comma cannot be put after the value instead: a map may end at any pair,
// and `{'keys': 'ALL', }` is not a statement.
func withColon(keys []string, after bool) []string {
	separator := ""
	if after {
		separator = ", "
	}

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, separator+"'"+key+"': ")
	}
	return out
}

// clusteringOrderCompletions is what to offer inside CLUSTERING ORDER BY, and
// whether the cursor is inside it.
//
// The column names are the table's own, being declared in the same statement,
// so they cannot be looked up: the note says one goes here. What follows one is
// the direction, and what follows that is the bracket that closes the list.
func clusteringOrderCompletions(input string) ([]string, bool) {
	const clause = "CLUSTERING ORDER BY"

	// The clause written as far as it has been typed, which is where the rest
	// of it is offered.
	switch {
	case endsWithWord(input, "CLUSTERING"):
		return []string{"ORDER BY ("}, true
	case containsWord(input, "CLUSTERING") && endsWithWord(input, "ORDER"):
		return []string{"BY ("}, true
	case containsWord(input, "CLUSTERING") && endsWithWord(input, "BY"):
		return []string{"("}, true
	}

	at := strings.LastIndex(strings.ToUpper(input), clause)
	if at < 0 {
		return nil, false
	}

	after := input[at+len(clause):]
	open := strings.Index(after, "(")
	if open < 0 {
		return []string{"("}, true
	}
	if strings.Contains(after[open:], ")") {
		return nil, false // closed: the rest of the clause carries on
	}

	item := after[open+1:]
	if comma := strings.LastIndex(item, ","); comma >= 0 {
		item = item[comma+1:]
	}

	words := finishedWords(item)
	switch {
	case len(words) == 0:
		return []string{columnHint}, true
	case len(words) == 1:
		return []string{"ASC", "DESC"}, true
	}
	return []string{")"}, true
}

// orderCanBeGiven reports whether the statement is one that sets the order its
// rows are stored in: a table being created, or a view, which takes a table's
// options and its order along with them.
func orderCanBeGiven(input string) bool {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) < 2 || words[0] != "CREATE" {
		return false
	}
	return words[1] == "TABLE" || words[1] == "COLUMNFAMILY" || words[1] == "MATERIALIZED"
}

// isCreateTable reports whether the statement is creating a table.
func isCreateTable(input string) bool {
	words := strings.Fields(strings.ToUpper(input))
	return len(words) > 1 && words[0] == "CREATE" &&
		(words[1] == "TABLE" || words[1] == "COLUMNFAMILY")
}
