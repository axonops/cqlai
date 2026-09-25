package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
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

	// The RESULTS view has two tabs of its own, and they have keys like any
	// other view. A reader who cannot find the trace in the README because it
	// is one level down is a reader who thinks it is gone.
	for _, tab := range resultTabs {
		assert.Contains(t, readme, tab.label+" ("+tab.key+")",
			"the tabs inside RESULTS should show %s", tab.label)
	}
	assert.Contains(t, readme, "`F5` | Trace", "the key table should still say F5 reaches the trace")
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

// TestEveryReadmeSaysWhatToTypeWhereCompletionStops.
//
// Completion answers a name it cannot look up with a note - `<column name>` -
// and the notation is worth nothing if the reader has to work out that it is
// not something to press Enter on.
func TestEveryReadmeSaysWhatToTypeWhereCompletionStops(t *testing.T) {
	for _, name := range []string{"README.md", "README_jp.md", "README_es.md", "README_gl.md"} {
		doc := readDoc(t, name)
		assert.Contains(t, doc, "`<table name>`", "%s should show the notation", name)
		assert.Contains(t, doc, "`<column name>`", "%s should show the notation", name)
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
	for _, tab := range resultTabs {
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

// TestTheDocsHaveNoDeadAnchors.
//
// The table of contents is a list of links into the same page, and stripping
// the emoji out of the headings changed every anchor they pointed at.
func TestTheDocsHaveNoDeadAnchors(t *testing.T) {
	docs, err := filepath.Glob("../../docs/*.md")
	require.NoError(t, err)
	docs = append(docs, "../../README.md", "../../README_jp.md",
		"../../README_es.md", "../../README_gl.md", "../../CONTRIBUTING.md")

	anchor := regexp.MustCompile(`\]\(#([^)]+)\)`)
	heading := regexp.MustCompile(`(?m)^#{1,6} +(.+)$`)

	for _, doc := range docs {
		text, err := os.ReadFile(doc) //nolint:gosec // a file in this repository
		if err != nil {
			continue // a file someone keeps locally rather than in the repository
		}

		headings := map[string]bool{}
		for _, found := range heading.FindAllStringSubmatch(string(text), -1) {
			headings[slug(found[1])] = true
		}

		for _, found := range anchor.FindAllStringSubmatch(string(text), -1) {
			assert.True(t, headings[found[1]],
				"%s links to %s, and no heading there makes that anchor", doc, found[0])
		}
	}
}

// slug turns a heading into the anchor GitHub gives it: lower case, spaces to
// dashes, and anything else that is not a letter, a digit or a dash dropped.
func slug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(heading)) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TestTheReadmeSaysConnectionsAreKept.
//
// The list of saved connections is the answer to "do I have to type this host
// again", and a reader who does not know it is there will.
func TestTheReadmeSaysConnectionsAreKept(t *testing.T) {
	for _, name := range []string{"README.md", "README_jp.md"} {
		doc := readDoc(t, name)
		assert.Contains(t, doc, createConnection, "%s should show the button", name)
		assert.Contains(t, doc, `"connections"`, "%s should show what the file holds", name)
	}
}

// TestTheDocsListTheConfigurationFiles.
//
// Where cqlai looks for its configuration is written out in eight documents in
// four languages, and the code that looks is one function. Changing the code
// and half the documents is how a reader ends up creating a file cqlai will
// never read.
func TestTheDocsListTheConfigurationFiles(t *testing.T) {
	// ConfigPaths builds them from the home directory, and the documents write
	// that as a tilde.
	t.Setenv("HOME", "~")

	for _, name := range []string{
		"README.md", "README_jp.md", "README_es.md", "README_gl.md",
		"docs/INSTALLATION.md", "docs/INSTALLATION_jp.md",
		"docs/CQLSHRC_SUPPORT.md", "docs/CQLSHRC_SUPPORT_jp.md",
	} {
		doc := readDoc(t, name)
		for _, path := range config.ConfigPaths() {
			// The one in the working directory is written either way round:
			// `cqlai.json`, or `./cqlai.json` to say which directory.
			said := strings.Contains(doc, "`"+path+"`") || strings.Contains(doc, "`./"+path+"`")
			assert.True(t, said, "%s does not say cqlai reads %s", name, path)
		}
	}
}

// TestEveryReadmeNamesTheQuitKey, and the help in the app does too.
//
// A shortcut only exists as far as someone can find it. The menu shows it
// beside QUIT, which is where you look while you are in there; these are where
// you look before you have opened it.
func TestEveryReadmeNamesTheQuitKey(t *testing.T) {
	for _, name := range []string{"README.md", "README_jp.md", "README_es.md", "README_gl.md"} {
		assert.Contains(t, readDoc(t, name), "`"+quitKeyLabel+"`",
			"%s should say which key quits", name)
	}

	var help strings.Builder
	for _, row := range router.HelpRows() {
		help.WriteString(strings.Join(row, " ") + "\n")
	}
	assert.Contains(t, help.String(), quitKeyLabel, "the help in the app should name it too")
}

// TestTheDocsDoNotPromiseAKeyTheTerminalTakes.
//
// Command+Q never reaches a terminal application - the terminal quits itself -
// so naming it beside the quit row would send a Mac user to a key that closes
// the wrong thing.
func TestTheDocsDoNotPromiseAKeyTheTerminalTakes(t *testing.T) {
	for _, name := range []string{"README.md", "README_jp.md", "README_es.md", "README_gl.md"} {
		found := false
		for _, line := range strings.Split(readDoc(t, name), "\n") {
			cells := strings.Split(line, "|")
			if len(cells) < 4 || strings.TrimSpace(cells[1]) != "`"+quitKeyLabel+"`" {
				continue
			}
			found = true

			// The last column is what macOS is told to press. It may go on to
			// explain why Command+Q is not it; what it must not do is offer it.
			assert.True(t, strings.HasPrefix(strings.TrimSpace(cells[3]), "`"+quitKeyLabel+"`"),
				"%s tells macOS to press %q", name, strings.TrimSpace(cells[3]))
		}
		assert.True(t, found, "%s has no row for %s", name, quitKeyLabel)
	}
}

// TestTheDocsNameEveryClipboardToolCqlaiLooksFor.
//
// Copying on a stock Linux desktop reaches no further than cqlai itself: no
// VTE terminal supports OSC 52, and wl-clipboard, xclip and xsel are none of
// them installed by default. Someone has to be told which to install, and the
// list they are told is only right if it is the list the code actually walks.
func TestTheDocsNameEveryClipboardToolCqlaiLooksFor(t *testing.T) {
	readme := readDoc(t, "README.md")

	for _, table := range [][][]string{clipboardWriters(), clipboardReaders()} {
		for _, command := range table {
			assert.Contains(t, readme, command[0],
				"the README should name %s, which is a tool cqlai tries", command[0])
		}
	}
}

// TestTheDocsSayHowToInstallAClipboardTool: naming the tool is not the same as
// saying how to get it, and the person reading this has just watched a copy do
// nothing.
func TestTheDocsSayHowToInstallAClipboardTool(t *testing.T) {
	for _, name := range []string{"README.md", "docs/INSTALLATION.md"} {
		doc := readDoc(t, name)

		assert.Contains(t, doc, "apt install wl-clipboard", "%s should say how to install it", name)
		assert.Contains(t, doc, "apt install xclip", "%s should say what X11 needs instead", name)
		assert.Contains(t, doc, "XDG_SESSION_TYPE", "%s should say how to tell which", name)
	}
}
