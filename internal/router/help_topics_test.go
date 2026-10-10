package router

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func topicNames(topics []HelpTopic) []string {
	var names []string
	for _, t := range topics {
		names = append(names, t.Name)
	}
	return names
}

// TestHELPFindsTheCommandTyped, the longest name the words begin with, in any
// case, with what follows it ignored.
func TestHELPFindsTheCommandTyped(t *testing.T) {
	for typed, want := range map[string]string{
		"insert":                        "INSERT",
		"INSERT;":                       "INSERT",
		"  Create   Table ":             "CREATE TABLE",
		"CREATE TABLE IF NOT EXISTS x":  "CREATE TABLE",
		"create columnfamily":           "CREATE TABLE",
		"begin batch":                   "BATCH",
		"CREATE MATERIALIZED VIEW":      "CREATE MATERIALIZED VIEW",
		"drop materialized view shop.v": "DROP MATERIALIZED VIEW",
	} {
		got := FindHelpTopics(typed)
		require.Len(t, got, 1, "%q", typed)
		assert.Equal(t, want, got[0].Name, "%q", typed)
	}
}

// TestHELPOnHalfACommandListsTheOnesItCouldBe.
func TestHELPOnHalfACommandListsTheOnesItCouldBe(t *testing.T) {
	names := topicNames(FindHelpTopics("create"))
	assert.Contains(t, names, "CREATE TABLE")
	assert.Contains(t, names, "CREATE KEYSPACE")
	assert.NotContains(t, names, "DROP TABLE")

	lines := strings.Join(HelpForTopic("create"), "\n")
	assert.Contains(t, lines, "HELP CREATE TABLE")
}

// TestHELPOnSomethingUnknownSaysWhatThereIs.
func TestHELPOnSomethingUnknownSaysWhatThereIs(t *testing.T) {
	assert.Empty(t, FindHelpTopics("frobnicate"))
	lines := strings.Join(HelpForTopic("frobnicate"), "\n")
	assert.Contains(t, lines, "There is no help for frobnicate")
	assert.Contains(t, lines, "INSERT")
	assert.Contains(t, lines, "CREATE TABLE")
}

// TestEveryTopicIsComplete: a name, what it does, and how it is written, in
// lines that fit the window.
func TestEveryTopicIsComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, topic := range HelpTopics() {
		assert.NotEmpty(t, topic.Summary, topic.Name)
		assert.NotEmpty(t, topic.Syntax, topic.Name)
		assert.False(t, seen[topic.Name], "%s twice", topic.Name)
		seen[topic.Name] = true
		for _, line := range topic.Lines() {
			if strings.HasPrefix(line, "https://") {
				continue // a link cannot be wrapped
			}
			assert.LessOrEqual(t, len([]rune(line)), 100, "%s: %q", topic.Name, line)
		}
		if topic.URL != "" {
			assert.True(t, strings.HasPrefix(topic.URL, "https://axonops.com/docs/"), topic.Name)
		}
	}
}

// TestBatchHELPPrintsTheTopic: there is no window there.
func TestBatchHELPPrintsTheTopic(t *testing.T) {
	h := &MetaCommandHandler{}
	text, ok := h.handleHelp("help insert;").(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(text, "INSERT"), text)

	_, isRows := h.handleHelp("HELP").([][]string)
	assert.True(t, isRows, "HELP on its own is the list")
}
