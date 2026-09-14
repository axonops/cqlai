package completion

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Checking the lists against Cassandra's own.
//
// Everything this package offers is a copy of something Cassandra defines: the
// types, the permissions, the table options. The copies went stale - four
// permissions that are not permissions, two table options removed in 4.0, and
// no `vector` - because nothing compared them with the source.
//
// These tests do, when a Cassandra checkout is at hand. They skip when it is
// not, so CI and anyone without one are unaffected: this is a check that runs
// where the answer is available, not a dependency.

// cassandraSource is where to look for a checkout: the environment first, so it
// can be pointed anywhere, then the place people put it.
func cassandraSource(t *testing.T) string {
	t.Helper()

	roots := []string{os.Getenv("CASSANDRA_SOURCE")}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, "git", "cassandra"))
	}

	for _, root := range roots {
		if root == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "src", "java", "org", "apache", "cassandra")); err == nil {
			return root
		}
	}

	t.Skip("no Cassandra checkout: set CASSANDRA_SOURCE to compare against one")
	return ""
}

func javaSource(t *testing.T, root, path string) string {
	t.Helper()

	text, err := os.ReadFile(filepath.Join(root, "src", "java", "org", "apache", "cassandra", path)) //nolint:gosec // a path inside the checkout
	require.NoError(t, err)
	return string(text)
}

// TestThePermissionsAreCassandras.
//
// INSERT, UPDATE, DELETE and TRUNCATE were offered here and are not
// permissions - MODIFY is the one that covers writing - so GRANT completed to
// four statements the server refuses.
func TestThePermissionsAreCassandras(t *testing.T) {
	root := cassandraSource(t)
	source := javaSource(t, root, "auth/Permission.java")

	// The enum's constants: a name on its own line, ending the declaration or
	// carrying on to the next.
	names := regexp.MustCompile(`(?m)^\s{4}([A-Z][A-Z_]*)\s*[,;]`)

	wanted := map[string]bool{"ALL": true} // ALL is CQL's word for every permission
	for _, found := range names.FindAllStringSubmatch(source, -1) {
		wanted[found[1]] = true
	}
	require.Contains(t, wanted, "MODIFY", "the enum was not read properly")

	for _, offered := range CQLPermissions {
		assert.True(t, wanted[offered],
			"%s is offered and is not a permission in this Cassandra", offered)
	}
	for name := range wanted {
		if name == "NONE" {
			continue
		}
		assert.Contains(t, CQLPermissions, name, "%s is a permission and is not offered", name)
	}
}

// TestTheTypesAreCassandras, less the ones that are not worth completing.
func TestTheTypesAreCassandras(t *testing.T) {
	root := cassandraSource(t)
	source := javaSource(t, root, "cql3/CQL3Type.java")

	native := regexp.MustCompile(`(?m)^\s{8}([A-Z][A-Z0-9]*)\s*\(`)
	found := native.FindAllStringSubmatch(source, -1)
	require.NotEmpty(t, found, "the Native enum was not read properly")

	offered := map[string]bool{}
	for _, name := range CQLDataTypes {
		offered[name] = true
	}

	for _, match := range found {
		name := strings.ToLower(match[1])
		if name == "empty" {
			continue // a type, and not one to declare a column as
		}
		assert.True(t, offered[name], "%s is a type and is not offered", name)
	}
}

// TestTheTableOptionsAreCassandras.
//
// Not the other way round: Cassandra's list is ahead of the servers people run,
// and the options newer than 5.0 are deliberately not offered yet.
func TestTheTableOptionsAreCassandras(t *testing.T) {
	root := cassandraSource(t)
	source := javaSource(t, root, "schema/TableParams.java")

	block := regexp.MustCompile(`(?s)public enum Option\s*\{(.*?)\}`).FindStringSubmatch(source)
	require.Len(t, block, 2, "the Option enum was not read properly")

	known := map[string]bool{}
	for _, name := range regexp.MustCompile(`[A-Z][A-Z_]*`).FindAllString(block[1], -1) {
		known[strings.ToLower(name)] = true
	}
	require.True(t, known["compaction"], "the Option enum was not read properly")

	for _, offered := range TableOptions {
		assert.True(t, known[offered],
			"%s is offered and is not a table option in this Cassandra", offered)
	}
}

// TestTheCompactionStrategiesExist.
func TestTheCompactionStrategiesExist(t *testing.T) {
	root := cassandraSource(t)

	for _, strategy := range CompactionStrategies {
		path := filepath.Join(root, "src", "java", "org", "apache", "cassandra",
			"db", "compaction", strategy+".java")
		_, err := os.Stat(path)
		assert.NoError(t, err, "%s is offered and there is no such strategy", strategy)
	}
}

// TestTheCompressorsExist.
func TestTheCompressorsExist(t *testing.T) {
	root := cassandraSource(t)

	for _, compressor := range Compressors {
		path := filepath.Join(root, "src", "java", "org", "apache", "cassandra",
			"io", "compress", compressor+".java")
		_, err := os.Stat(path)
		assert.NoError(t, err, "%s is offered and there is no such compressor", compressor)
	}
}

// TestTheMapKeysAreCassandras.
//
// The keys of compaction, compression and caching, and the ones each strategy
// takes of its own. max_compressed_length was offered here and is in no version
// of the server - cqlsh offers it too - so the compression map completed to a
// key Cassandra refuses.
func TestTheMapKeysAreCassandras(t *testing.T) {
	root := cassandraSource(t)

	for _, this := range []struct {
		keys  []string
		files []string
	}{
		{CompactionOptions, []string{
			"schema/CompactionParams.java",
			"db/compaction/AbstractCompactionStrategy.java",
		}},
		{CompressionOptions, []string{"schema/CompressionParams.java"}},
		{CachingOptions, []string{"schema/CachingParams.java"}},
		{StrategyOptions["SizeTieredCompactionStrategy"], []string{
			"db/compaction/SizeTieredCompactionStrategyOptions.java",
		}},
		{StrategyOptions["LeveledCompactionStrategy"], []string{
			"db/compaction/LeveledCompactionStrategy.java",
		}},
		{StrategyOptions["TimeWindowCompactionStrategy"], []string{
			"db/compaction/TimeWindowCompactionStrategyOptions.java",
		}},
		{StrategyOptions["UnifiedCompactionStrategy"], []string{
			"db/compaction/unified/Controller.java",
		}},
	} {
		var source strings.Builder
		for _, file := range this.files {
			source.WriteString(javaSource(t, root, file))
		}

		for _, key := range this.keys {
			assert.True(t, definedIn(source.String(), key),
				"%s is offered and this Cassandra does not take it", key)
		}
	}
}

// TestTheOptionValuesAreCassandras: the values offered for the options that
// have a fixed few, rather than a number only the user knows.
func TestTheOptionValuesAreCassandras(t *testing.T) {
	root := cassandraSource(t)

	for _, this := range []struct {
		values []string
		file   string
	}{
		{ReadRepairValues, "service/reads/repair/ReadRepairStrategy.java"},
		{TombstoneValues, "schema/CompactionParams.java"},
		{CachingValues, "schema/CachingParams.java"},
		{CompactionWindowUnits, "db/compaction/TimeWindowCompactionStrategyOptions.java"},
		{TimestampResolutions, "db/compaction/TimeWindowCompactionStrategyOptions.java"},
	} {
		source := javaSource(t, root, this.file)
		for _, value := range this.values {
			assert.Contains(t, source, value,
				"%s is offered and is not in %s", value, this.file)
		}
	}
}

// definedIn reports whether a source file defines an option by this name,
// written either as the string itself or as the constant of an enum whose
// toString lowercases it.
func definedIn(source, name string) bool {
	return strings.Contains(source, `"`+name+`"`) ||
		regexp.MustCompile(`\b`+strings.ToUpper(name)+`\b`).MatchString(source)
}

// TestTheIndexOptionsAreCassandras.
//
// What a storage attached index takes, from the set the server validates
// against: the vector options, and the three an analyzer takes.
func TestTheIndexOptionsAreCassandras(t *testing.T) {
	root := cassandraSource(t)

	source := javaSource(t, root, "index/sai/disk/v1/IndexWriterConfig.java") +
		javaSource(t, root, "index/sai/analyzer/NonTokenizingOptions.java")

	for _, option := range IndexOptions {
		assert.Contains(t, source, `"`+option+`"`,
			"%s is offered and this Cassandra does not take it", option)
	}

	valid := javaSource(t, root, "index/sai/StorageAttachedIndex.java")
	require.Contains(t, valid, "VALID_OPTIONS", "the option set was not read properly")
}

// TestTheIndexValuesAreCassandras: what an index is built on, what it is built
// with, and the values its options take.
func TestTheIndexValuesAreCassandras(t *testing.T) {
	root := cassandraSource(t)

	targets := javaSource(t, root, "cql3/statements/schema/IndexTarget.java")
	for _, target := range IndexTargets {
		name := strings.TrimSuffix(target, "(")
		assert.Contains(t, targets, `"`+name+`"`,
			"%s is offered and is not a target in this Cassandra", name)
	}

	names := javaSource(t, root, "index/sai/StorageAttachedIndex.java") +
		javaSource(t, root, "index/internal/CassandraIndex.java") +
		javaSource(t, root, "schema/IndexMetadata.java")
	for _, class := range []string{"sai", "legacy_local_table"} {
		assert.Contains(t, names, `"`+class+`"`,
			"%s is offered and this Cassandra does not know it", class)
	}

	optimize := javaSource(t, root, "index/sai/disk/v1/vector/OptimizeFor.java")
	for _, value := range OptimizeForValues {
		assert.Contains(t, optimize, value, "%s is offered and is not an OptimizeFor", value)
	}
}
