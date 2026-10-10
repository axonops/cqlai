package router

import (
	"sort"
	"strings"
)

// HELP <command>: how a statement or a shell command is written.
//
// HELP on its own is the list of everything there is. HELP INSERT is INSERT:
// what it does, the syntax with every clause it takes, and an example - what
// you look up when you know the statement and not the order its clauses go in.

// HelpTopic is one command's help.
type HelpTopic struct {
	Name    string   // as typed after HELP: "INSERT", "CREATE TABLE"
	Aliases []string // other ways of typing it: "DESC", "BEGIN BATCH"
	Summary string   // what it does, in a sentence
	Syntax  []string // how it is written, a line each; [ ] is optional, | is one of
	Notes   []string // what is worth knowing about it, a line each
	Example []string
	URL     string // where it is documented in full
}

// Lines is the topic laid out for reading.
func (t HelpTopic) Lines() []string {
	lines := []string{t.Name, ""}
	lines = append(lines, wrapWords(t.Summary, helpTextWidth)...)
	lines = append(lines, "", "Syntax:", "")
	for _, s := range t.Syntax {
		lines = append(lines, "    "+s)
	}
	if len(t.Notes) > 0 {
		lines = append(lines, "")
		for _, note := range t.Notes {
			lines = append(lines, wrapWords(note, helpTextWidth)...)
		}
	}
	if len(t.Example) > 0 {
		lines = append(lines, "", "Example:", "")
		for _, e := range t.Example {
			lines = append(lines, "    "+e)
		}
	}
	if t.URL != "" {
		// On a line of its own: a link cannot be wrapped.
		lines = append(lines, "", "More:", t.URL)
	}
	return lines
}

// helpTextWidth is how wide the sentences of a topic are wrapped: the syntax
// and examples are kept as they are written.
const helpTextWidth = 74

// wrapWords breaks a sentence into lines no wider than width, at spaces.
func wrapWords(text string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && len([]rune(line))+1+len([]rune(word)) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// HelpTopics is every topic, sorted by name.
func HelpTopics() []HelpTopic {
	topics := append(append([]HelpTopic{}, cqlHelpTopics...), shellHelpTopics...)
	sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })
	return topics
}

// FindHelpTopics is the topics for what was typed after HELP.
//
// The topic whose name, or one of whose aliases, the typed words begin with -
// the longest such, so HELP CREATE TABLE IF NOT EXISTS is CREATE TABLE. Failing
// that, every topic that begins with what was typed: HELP CREATE is the CREATE
// statements, to pick from. Nothing at all is no topic.
func FindHelpTopics(typed string) []HelpTopic {
	words := strings.Fields(strings.ToUpper(strings.TrimSuffix(strings.TrimSpace(typed), ";")))
	if len(words) == 0 {
		return nil
	}
	text := strings.Join(words, " ")

	var best HelpTopic
	bestLen := 0
	for _, t := range HelpTopics() {
		for _, name := range append([]string{t.Name}, t.Aliases...) {
			if (text == name || strings.HasPrefix(text, name+" ")) && len(name) > bestLen {
				best, bestLen = t, len(name)
			}
		}
	}
	if bestLen > 0 {
		return []HelpTopic{best}
	}

	var starting []HelpTopic
	for _, t := range HelpTopics() {
		if strings.HasPrefix(t.Name, text+" ") || strings.HasPrefix(t.Name, text) {
			starting = append(starting, t)
		}
	}
	return starting
}

// HelpTopicNames is the name of every topic, for saying what there is.
func HelpTopicNames() []string {
	var names []string
	for _, t := range HelpTopics() {
		names = append(names, t.Name)
	}
	return names
}

// HelpForTopic is HELP <typed> as lines: the topic, the topics to pick from,
// or what there is when nothing matches.
func HelpForTopic(typed string) []string {
	found := FindHelpTopics(typed)
	switch len(found) {
	case 1:
		return found[0].Lines()
	case 0:
		lines := []string{"There is no help for " + strings.TrimSpace(typed) + ". HELP has these:", ""}
		return append(lines, columns(HelpTopicNames(), 3)...)
	}
	lines := []string{"HELP " + strings.ToUpper(strings.TrimSpace(typed)) + " is one of these:", ""}
	width := 0
	for _, t := range found {
		width = max(width, len(t.Name))
	}
	for _, t := range found {
		name := "    HELP " + t.Name + strings.Repeat(" ", width-len(t.Name)) + "  "
		for i, line := range wrapWords(t.Summary, max(helpTextWidth-len(name), 30)) {
			if i > 0 {
				name = strings.Repeat(" ", len(name))
			}
			lines = append(lines, name+line)
		}
	}
	return lines
}

// columns lays names out n to a line, lined up.
func columns(names []string, n int) []string {
	width := 0
	for _, name := range names {
		width = max(width, len(name))
	}
	var lines []string
	for i := 0; i < len(names); i += n {
		var row []string
		for _, name := range names[i:min(i+n, len(names))] {
			row = append(row, name+strings.Repeat(" ", width-len(name)))
		}
		lines = append(lines, "    "+strings.TrimRight(strings.Join(row, "  "), " "))
	}
	return lines
}
