package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/router"
)

// The documented facts, checked against the code that decides them.
//
// The README was updated in every commit that changed the UI and still fell
// behind, because the same fact is written in more than one place in it: the
// FILE menu is drawn in one section and described in another, and each time the
// menu changed the section being written got the update and the other did not.
// That is the defect this project keeps finding in its own code - a rule
// written twice with only one copy maintained - and the answer is the same one:
// check the copy against the thing it describes rather than keeping them in
// step by hand.
//
// These tests fail when the code says something the documentation does not.
// They do not check prose, only the facts a reader would act on: which views
// there are and what reaches them, and what the FILE menu offers.

func readDoc(t *testing.T, name string) string {
	t.Helper()

	text, err := os.ReadFile("../../" + name) //nolint:gosec // a file in this repository
	require.NoError(t, err, "reading %s", name)
	return string(text)
}

// TestTheReadmeSaysWhichKeyReachesWhichView.
func TestTheReadmeSaysWhichKeyReachesWhichView(t *testing.T) {
	readme := readDoc(t, "README.md")

	for _, tab := range modeTabs {
		// The label as the README writes it: CONSOLE in the tab line, Console
		// in the table of keys.
		label := strings.ToUpper(tab.label[:1]) + strings.ToLower(tab.label[1:])

		assert.Contains(t, readme, "`"+tab.key+"` | "+label,
			"the key table should say %s reaches %s", tab.key, label)
		assert.Contains(t, readme, tab.label+" ("+tab.key+")",
			"the tab line should show %s", tab.label)
	}
}

// TestTheReadmeListsWhatTheFileMenuOffers.
//
// This is the one that went stale: the menu grew SAVE RESULTS, PREFERENCES and
// QUIT, and the section drawing it kept showing four entries.
func TestTheReadmeListsWhatTheFileMenuOffers(t *testing.T) {
	readme := readDoc(t, "README.md")

	for _, item := range fileMenuItems() {
		if item.rule {
			continue
		}
		assert.Contains(t, readme, item.label,
			"the FILE menu section should list %s", item.label)
	}
}

// TestTheTranslationsSayWhichKeyReachesWhichView.
//
// Only the keys. What a view is called in another language is the translator's
// business; which key reaches it is not, and that is the part that went stale -
// every translation still said F5 was the AI view long after it was not.
func TestTheTranslationsSayWhichKeyReachesWhichView(t *testing.T) {
	for _, name := range []string{"README_jp.md", "README_es.md", "README_gl.md"} {
		doc := readDoc(t, name)
		for _, tab := range modeTabs {
			assert.Contains(t, doc, "`"+tab.key+"`", "%s should mention %s", name, tab.key)
		}
	}
}

// TestTheHelpInTheAppNamesEveryView, since it is what F1 shows and the only
// documentation anyone reads while they are using cqlai.
func TestTheHelpInTheAppNamesEveryView(t *testing.T) {
	var help strings.Builder
	for _, row := range router.HelpRows() {
		help.WriteString(strings.Join(row, " ") + "\n")
	}

	for _, tab := range modeTabs {
		assert.Contains(t, help.String(), tab.key, "the help should name %s", tab.key)
	}
}

// TestTheDocsHaveNoDeadLinks.
//
// Six of them pointed at pages that were never written - a COPY reference, a
// data types guide, a performance page - which is worse than no link: it reads
// as somewhere to go and is not.
func TestTheDocsHaveNoDeadLinks(t *testing.T) {
	docs, err := filepath.Glob("../../docs/*.md")
	require.NoError(t, err)
	docs = append(docs, "../../README.md")

	link := regexp.MustCompile(`\]\((\./|\.\./)?([A-Za-z0-9_./-]+\.md)(#[^)]*)?\)`)
	for _, doc := range docs {
		text, err := os.ReadFile(doc) //nolint:gosec // a file in this repository
		require.NoError(t, err)

		for _, found := range link.FindAllStringSubmatch(string(text), -1) {
			target := filepath.Join(filepath.Dir(doc), found[1]+found[2])
			assert.FileExists(t, target, "%s links to %s, which is not there", doc, found[0])
		}
	}
}
