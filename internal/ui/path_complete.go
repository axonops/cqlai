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

	if len(matches) == 0 {
		return none
	}
	sort.Strings(matches)

	// Fill in as far as they agree. With one match that is the whole name, and
	// a directory ends in a separator so the next Tab carries on inside it.
	return pathCompletion{
		Completed: dir + commonPrefix(matches),
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
