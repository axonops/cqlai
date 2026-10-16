package completion

import (
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// Names that need double quotes. CQL folds an unquoted name to lower case, so a
// keyspace called MyKeyspace has to be written "MyKeyspace" - and completion
// put it in as it is, which ran as mykeyspace and was not found.
//
// The engine works on names without their quotes: what is typed has them taken
// off on the way in, and the names it offers have them put on, where they are
// needed, on the way out. In between, a name is looked up in the cache as it is
// stored, whatever case it was typed in.

// unquoteIdentifiers is the input with the double quotes around names taken
// off, so the engine sees "MyKeyspace". as MyKeyspace. and "My as My.
//
// A string in single quotes is left as it is: a double quote in one is part of
// the string. A doubled double quote inside a name is one double quote.
func unquoteIdentifiers(input string) string {
	if !strings.Contains(input, `"`) {
		return input
	}

	var b strings.Builder
	inString, inName := false, false
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case inString:
			b.WriteByte(c)
			if c == '\'' {
				inString = false
			}
		case c == '"' && inName && i+1 < len(input) && input[i+1] == '"':
			b.WriteByte('"')
			i++
		case c == '"':
			inName = !inName
		case c == '\'' && !inName:
			inString = true
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// openNameAt is where a name in double quotes that has not been closed starts,
// or -1 when there is none. It is what a completion replaces: `FROM "My` takes
// "MyKeyspace" in place of `"My`, quote and all.
func openNameAt(input string) int {
	at, inString := -1, false
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case inString:
			if c == '\'' {
				inString = false
			}
		case at >= 0:
			if c == '"' {
				if i+1 < len(input) && input[i+1] == '"' {
					i++ // a doubled quote is part of the name
					continue
				}
				at = -1
			}
		case c == '"':
			at = i
		case c == '\'':
			inString = true
		}
	}
	return at
}

// OpenNameAt is openNameAt for the prompt.
func OpenNameAt(input string) int { return openNameAt(input) }

// nameIn is the name in names that typed means.
//
// The name as it is, when there is one: a quoted name is exact. Then the name in
// lower case, which is what CQL makes of one typed without quotes. Then any
// that differs only in case, since the engine has often upper-cased the input
// by the time it looks a name up. Lower case when nothing matches, as CQL would
// read it.
func nameIn(names []string, typed string) string {
	for _, name := range names {
		if name == typed {
			return name
		}
	}
	lower := strings.ToLower(typed)
	for _, name := range names {
		if name == lower {
			return name
		}
	}
	for _, name := range names {
		if strings.EqualFold(name, typed) {
			return name
		}
	}
	return lower
}

// keyspaceNamed is the keyspace a typed name means.
func (ce *CompletionEngine) keyspaceNamed(typed string) string {
	return nameIn(ce.getKeyspaceNames(), typed)
}

// tableNamed is the keyspace and table a typed name means, qualified or not.
// The keyspace is "" when the name is not qualified and none is in use.
func (ce *CompletionEngine) tableNamed(typed string) (keyspace, table string) {
	if ce.sessionManager != nil {
		keyspace = ce.sessionManager.CurrentKeyspace()
	}
	if ks, rest, qualified := strings.Cut(typed, "."); qualified {
		keyspace, typed = ce.keyspaceNamed(ks), rest
	}
	if keyspace == "" {
		return "", ""
	}
	return keyspace, nameIn(ce.getTablesForKeyspace(keyspace), typed)
}

// quoteNames puts double quotes on the names offered that need them: a
// keyspace, table or column the cache holds, alone, with the dot a keyspace is
// offered with, or as keyspace.table.
//
// Only a name the cache holds: SELECT, offered as a keyword, has capitals too.
func (ce *CompletionEngine) quoteNames(suggestions []string) []string {
	known := ce.knownNames()
	if len(known) == 0 {
		return suggestions
	}

	quote := func(name string) (string, bool) {
		if !known[name] {
			return name, false
		}
		return db.QuoteName(name), true
	}

	out := make([]string, len(suggestions))
	for i, s := range suggestions {
		out[i] = s
		if IsHint(s) {
			continue
		}
		if q, ok := quote(s); ok {
			out[i] = q
			continue
		}
		if ks, rest, dotted := strings.Cut(s, "."); dotted {
			qks, ok := quote(ks)
			if !ok {
				continue
			}
			if rest == "" {
				out[i] = qks + "."
				continue
			}
			if qrest, ok := quote(rest); ok {
				out[i] = qks + "." + qrest
			}
		}
	}
	return out
}

// knownNames is every keyspace, table and column the cache has read, without
// reading any more.
func (ce *CompletionEngine) knownNames() map[string]bool {
	known := map[string]bool{}
	for _, name := range ce.cache.keyspaces {
		known[name] = true
	}
	for _, names := range ce.cache.tables {
		for _, name := range names {
			known[name] = true
		}
	}
	for _, names := range ce.cache.columns {
		for _, name := range names {
			known[name] = true
		}
	}
	return known
}

// quoteEach is names written as CQL reads them back.
func quoteEach(names []string) []string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = db.QuoteName(name)
	}
	return out
}
