package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTheWithClauseOffersTheOptions.
//
// The list was in the package and read by nothing: `CREATE TABLE ... WITH `
// completed to silence while a list of option names sat beside it.
func TestTheWithClauseOffersTheOptions(t *testing.T) {
	got := tableOptionCompletions("CREATE TABLE users (id uuid PRIMARY KEY) WITH ")

	assert.Contains(t, got, "compaction = ")
	assert.Contains(t, got, "default_time_to_live = ")
	assert.Contains(t, got, "cdc = ")

	// And again after AND, which is how the second option is reached.
	got = tableOptionCompletions("ALTER TABLE users WITH comment = 'x' AND ")
	assert.Contains(t, got, "compression = ")
}

// TestTheOptionsAreCassandras: the ones it has, and not the ones it dropped.
func TestTheOptionsAreCassandras(t *testing.T) {
	got := tableOptionCompletions("CREATE TABLE users (id uuid PRIMARY KEY) WITH ")

	for _, gone := range []string{"read_repair_chance = ", "dclocal_read_repair_chance = "} {
		assert.NotContains(t, got, gone, "removed in Cassandra 4.0")
	}
	for _, wanted := range []string{"allow_auto_snapshot = ", "incremental_backups = ",
		"additional_write_policy = ", "memtable = ", "read_repair = "} {
		assert.Contains(t, got, wanted)
	}
}

// TestAMapOptionOpensItsBrace, rather than leaving you to remember the shape.
func TestAMapOptionOpensItsBrace(t *testing.T) {
	assert.Equal(t, []string{"{'"},
		tableOptionCompletions("CREATE TABLE users (id uuid PRIMARY KEY) WITH compaction = "))
}

// TestInsideAMapItsKeysAreOffered, each with the colon that takes it to its
// value, so the next Tab offers the value.
func TestInsideAMapItsKeysAreOffered(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH compaction = {")
	assert.Contains(t, got, "'class': ")
	assert.Contains(t, got, "'tombstone_threshold': ")
	assert.NotContains(t, got, "'chunk_length_in_kb': ", "that is compression's")

	got = tableOptionCompletions("ALTER TABLE users WITH compression = {")
	assert.Contains(t, got, "'chunk_length_in_kb': ")

	got = tableOptionCompletions("ALTER TABLE users WITH caching = {")
	assert.Contains(t, got, "'keys': ")
	assert.Contains(t, got, "'rows_per_partition': ")
}

// TestAKeyIsOfferedOnce, since a map that gives one twice is a mistake.
func TestAKeyIsOfferedOnce(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH caching = {'keys': 'ALL', ")

	assert.NotContains(t, got, "'keys': ")
	assert.Contains(t, got, "'rows_per_partition': ")
}

// TestAnOptionIsOfferedOnce, for the same reason.
func TestAnOptionIsOfferedOnce(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH gc_grace_seconds = 100 AND ")

	assert.NotContains(t, got, "gc_grace_seconds = ")
	assert.Contains(t, got, "comment = ")
}

// TestTheKeyAfterAFinishedPairCarriesTheComma.
//
// The comma cannot go after the value instead: a map may end at any pair, and
// `{'keys': 'ALL', }` is not a statement.
func TestTheKeyAfterAFinishedPairCarriesTheComma(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH caching = {'keys': 'ALL'")

	assert.Contains(t, got, ", 'rows_per_partition': ")
}

// TestAStrategyBringsItsOwnKeys, once the map says which one it is.
func TestAStrategyBringsItsOwnKeys(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH compaction = {'class': 'LeveledCompactionStrategy', ")
	assert.Contains(t, got, "'sstable_size_in_mb': ")
	assert.NotContains(t, got, "'compaction_window_unit': ", "that is the time window strategy's")

	// Written in full, as Cassandra writes it back.
	got = tableOptionCompletions("ALTER TABLE users WITH compaction = {'class': 'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy', ")
	assert.Contains(t, got, "'compaction_window_unit': ")

	// And before the class is known, only the keys every strategy takes.
	got = tableOptionCompletions("ALTER TABLE users WITH compaction = {")
	assert.NotContains(t, got, "'sstable_size_in_mb': ")
}

// TestAnOptionWithAFixedFewValuesOffersThem.
func TestAnOptionWithAFixedFewValuesOffersThem(t *testing.T) {
	for typed, wanted := range map[string]string{
		"ALTER TABLE users WITH cdc = ":                                                                            "true",
		"ALTER TABLE users WITH read_repair = ":                                                                    "'BLOCKING'",
		"ALTER TABLE users WITH speculative_retry = ":                                                              "'ALWAYS'",
		"ALTER TABLE users WITH allow_auto_snapshot = ":                                                            "false",
		"ALTER TABLE users WITH caching = {'keys': ":                                                               "'ALL'",
		"ALTER TABLE users WITH compaction = {'enabled': ":                                                         "'true'",
		"ALTER TABLE users WITH compression = {'class': ":                                                          "'ZstdCompressor'",
		"ALTER TABLE users WITH compaction = {'provide_overlapping_tombstones': ":                                  "'CELL'",
		"ALTER TABLE users WITH compaction = {'class': 'TimeWindowCompactionStrategy', 'compaction_window_unit': ": "'DAYS'",
	} {
		assert.Contains(t, tableOptionCompletions(typed), wanted, "after %q", typed)
	}
}

// TestAnOptionWithNoFixedValuesSaysWhatToType.
//
// A number of seconds is not something completion can know. Saying nothing
// leaves the user to guess whether the option takes a number, a word, or a
// string in quotes.
func TestAnOptionWithNoFixedValuesSaysWhatToType(t *testing.T) {
	for typed, note := range map[string]string{
		"ALTER TABLE users WITH gc_grace_seconds = ":                                                        Hint("seconds"),
		"ALTER TABLE users WITH memtable_flush_period_in_ms = ":                                             Hint("milliseconds"),
		"ALTER TABLE users WITH bloom_filter_fp_chance = ":                                                  Hint("0.0 to 1.0"),
		"ALTER TABLE users WITH comment = ":                                                                 Hint("'text'"),
		"ALTER TABLE users WITH compaction = {'min_threshold': ":                                            Hint("sstables"),
		"ALTER TABLE users WITH compaction = {'class': 'UnifiedCompactionStrategy', 'scaling_parameters': ": Hint("'T4', 'L4' or 'N'"),
	} {
		assert.Equal(t, []string{note}, tableOptionCompletions(typed), "after %q", typed)
	}
}

// TestAValueCanBeBothAndTheNoteComesLast: the rows cached per partition are
// ALL, NONE, or a number.
func TestAValueCanBeBothAndTheNoteComesLast(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH caching = {'rows_per_partition': ")

	assert.Equal(t, []string{"'ALL'", "'NONE'", Hint("rows")}, got)
}

// TestAValueBeingTypedIsStillTheValuePosition, so it can be filtered to what is
// being typed.
func TestAValueBeingTypedIsStillTheValuePosition(t *testing.T) {
	assert.Contains(t, tableOptionCompletions("ALTER TABLE users WITH read_repair = 'BLO"), "'BLOCKING'")
	assert.Equal(t, []string{"AND"}, tableOptionCompletions("ALTER TABLE users WITH read_repair = 'BLOCKING' "),
		"a value that is finished is not the value position")
}

// TestAFinishedOptionIsFollowedByAnd.
func TestAFinishedOptionIsFollowedByAnd(t *testing.T) {
	assert.Equal(t, []string{"AND"},
		tableOptionCompletions("ALTER TABLE users WITH gc_grace_seconds = 100 "))
	assert.Equal(t, []string{"AND"},
		tableOptionCompletions("ALTER TABLE users WITH compaction = {'class': 'LeveledCompactionStrategy'} "))
}

// TestTheClassKeyOffersTheClasses, which is the one value nobody remembers in
// full, in the quotes it is written in.
func TestTheClassKeyOffersTheClasses(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH compaction = {'class': '")
	assert.Contains(t, got, "'LeveledCompactionStrategy'")
	assert.Contains(t, got, "'UnifiedCompactionStrategy'")

	got = tableOptionCompletions("ALTER TABLE users WITH compression = {'class': '")
	assert.Contains(t, got, "'ZstdCompressor'")
	assert.NotContains(t, got, "'LeveledCompactionStrategy'")
}

// TestAClosedMapIsBackToTheOptions.
func TestAClosedMapIsBackToTheOptions(t *testing.T) {
	got := tableOptionCompletions("ALTER TABLE users WITH compaction = {'class': 'LeveledCompactionStrategy'} AND ")
	assert.Contains(t, got, "gc_grace_seconds = ")
	assert.NotContains(t, got, "'class': ")
}

// TestSomethingThatIsNotAWithClauseIsLeftAlone, so the statement's own
// completion answers for it.
func TestSomethingThatIsNotAWithClauseIsLeftAlone(t *testing.T) {
	assert.Empty(t, tableOptionCompletions("CREATE TABLE users ("))
	assert.Empty(t, tableOptionCompletions("SELECT * FROM users WHERE "))
}

// TestTheClusteringOrderIsOfferedWithTheOptions.
//
// It is written in the WITH clause and is not one of the options: it has its
// own brackets rather than a value, and it belongs to a table being created -
// the order of a table that exists cannot be altered.
func TestTheClusteringOrderIsOfferedWithTheOptions(t *testing.T) {
	got := tableOptionCompletions("CREATE TABLE t (a int, b int, PRIMARY KEY (a, b)) WITH ")
	assert.Contains(t, got, "CLUSTERING ORDER BY (")

	assert.NotContains(t, tableOptionCompletions("ALTER TABLE t WITH "), "CLUSTERING ORDER BY (")

	// And once, like every other thing in the clause.
	got = tableOptionCompletions("CREATE TABLE t (a int PRIMARY KEY) WITH CLUSTERING ORDER BY (b DESC) AND ")
	assert.NotContains(t, got, "CLUSTERING ORDER BY (")
}

// TestTheClusteringClauseIsFinishedForYou, whichever word of it was typed.
func TestTheClusteringClauseIsFinishedForYou(t *testing.T) {
	assert.Equal(t, []string{"ORDER BY ("},
		tableOptionCompletions("CREATE TABLE t (a int PRIMARY KEY) WITH CLUSTERING "))
	assert.Equal(t, []string{"BY ("},
		tableOptionCompletions("CREATE TABLE t (a int PRIMARY KEY) WITH CLUSTERING ORDER "))
	assert.Equal(t, []string{"("},
		tableOptionCompletions("CREATE TABLE t (a int PRIMARY KEY) WITH CLUSTERING ORDER BY "))
}

// TestInsideTheClusteringOrder: a column, then the direction, then the bracket
// that closes the list.
func TestInsideTheClusteringOrder(t *testing.T) {
	const start = "CREATE TABLE t (a int, b int, PRIMARY KEY (a, b)) WITH CLUSTERING ORDER BY ("

	assert.Equal(t, []string{columnHint}, tableOptionCompletions(start))
	assert.Equal(t, []string{"ASC", "DESC"}, tableOptionCompletions(start+"b "))
	assert.Equal(t, []string{")"}, tableOptionCompletions(start+"b DESC "))
	assert.Equal(t, []string{columnHint}, tableOptionCompletions(start+"b DESC, "))

	// And once it is closed, the clause carries on.
	assert.Equal(t, []string{"AND"}, tableOptionCompletions(start+"b DESC) "))
}
