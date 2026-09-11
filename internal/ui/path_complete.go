package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Tab completion for file paths.
//
// There was none anywhere in cqlai: every path - SAVE, CAPTURE, COPY, SOURCE -
// had to be typed out in full and correctly, with no way to find out what was
// there without leaving the program.
//
// It behaves the way a shell does. The first Tab fills in as far as the
// candidates agree; if that is not far enough to pick one, the candidates are
// listed and a second Tab does nothing new.

// pathCompletion is what Tab found for a partially typed path.
type pathCompletion struct {
	// Completed is the path with as much filled in as the candidates agree on.
	// It is the input unchanged when nothing matched.
	Completed string

	// Matches are the candidates, for showing when there is more than one.
	// Directories carry a trailing separator.
	Matches []string
}

// completePath completes a partially typed file path.
//
// It never fails: a directory that cannot be read, or does not exist, simply
// offers nothing. Tab is not the place to learn that a path is wrong.
func completePath(input string) pathCompletion {
	none := pathCompletion{Completed: input}

	dir, prefix := splitPath(input)

	// Nothing typed at all: start at the root rather than in whatever directory
	// cqlai was started from, which is rarely where the file is and is not
	// somewhere you can see - the field looks empty and Tab produces a listing
	// you had no reason to expect. It is also the one case with no directory to
	// offer a way up out of.
	//
	// A name with no directory in front of it - report.csv - is a different
	// thing, and still completes here, which is what typing a bare name means.
	if dir == "" && prefix == "" {
		dir = pathRoot
	}

	entries, err := os.ReadDir(expandHome(dir))
	if err != nil {
		return none
	}

	matches := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// Hidden files only when they were asked for, as a shell does.
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if entry.IsDir() {
			name += string(filepath.Separator)
		}
		matches = append(matches, name)
	}

	// The way up counts as something to offer, so an empty directory is not a
	// dead end: with nothing in it and nothing listed, the only way back out
	// was to edit the path by hand.
	canGoUp := prefix == "" && parentDir(dir) != ""
	if len(matches) == 0 && !canGoUp {
		return none
	}
	sort.Strings(matches)

	// Fill in as far as they agree. With one match that is the whole name, and
	// a directory ends in a separator so the next Tab carries on inside it.
	completed := dir + commonPrefix(matches)

	// The way back up, at the top of the list, so browsing is not one-way: with
	// only the contents of a directory to pick from, a wrong turn meant editing
	// the path by hand.
	//
	// Added after the common prefix is worked out rather than before. In the
	// list it is one more thing to pick; in the prefix it agrees with nothing,
	// and would stop "/b" filling itself in to "/boot/" the moment the parent
	// directory existed.
	if canGoUp {
		matches = append([]string{parentEntry}, matches...)
	}

	return pathCompletion{
		Completed: completed,
		Matches:   matches,
	}
}

// splitPath divides what has been typed into the directory to look in and the
// prefix to match, keeping the directory exactly as it was written so
// completing does not rewrite the rest of the path.
func splitPath(input string) (dir, prefix string) {
	i := strings.LastIndex(input, string(filepath.Separator))
	if i < 0 {
		return "", input
	}
	return input[:i+1], input[i+1:]
}

// worthListing reports whether the candidates are worth showing.
//
// More than one, as a shell does - or exactly one that is the way up, which is
// what an empty directory offers. That one is not a completion to fill in: it
// is the only thing you can do from there, and applying it silently would walk
// you out of the directory you had just walked into.
func (p pathCompletion) worthListing() bool {
	return len(p.Matches) > 1 || (len(p.Matches) == 1 && p.Matches[0] == parentEntry)
}

// parentEntry is the way up, as it appears in the list.
const parentEntry = ".."

// parentDir is the directory above one, or "" when there is nowhere above.
//
// Takes the path as typed, so it works before ~ is expanded: what goes back in
// the field should read the way the rest of it does.
func parentDir(dir string) string {
	sep := string(filepath.Separator)

	trimmed := strings.TrimSuffix(dir, sep)
	if trimmed == "" || trimmed == "~" {
		return "" // the root, or home with nothing above it worth offering
	}

	i := strings.LastIndex(trimmed, sep)
	if i < 0 {
		// A bare name with no separator: above it is where we started.
		return ""
	}
	return trimmed[:i+1]
}

// expandHome turns a leading ~ into the home directory.
func expandHome(dir string) string {
	switch {
	case dir == "":
		return "."
	case dir == "~"+string(filepath.Separator), dir == "~":
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	case strings.HasPrefix(dir, "~"+string(filepath.Separator)):
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, dir[2:])
		}
	}
	return dir
}

// commonPrefix is the longest start the names share.
func commonPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}

	prefix := names[0]
	for _, name := range names[1:] {
		for !strings.HasPrefix(name, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}
