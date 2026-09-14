package completion

import "strings"

// Completing the WITH clause of a keyspace, and of a role.
//
// Both are options joined by AND, the same shape as a table's, and both were
// answered with one thing and then silence: a keyspace offered REPLICATION and
// then the map to put after it, and a role offered nothing at all.

// replicationValues is what each key of the replication map can be set to.
var replicationValues = map[string]whatGoesHere{
	"class":              {values: quoted(ReplicationStrategies)},
	"replication_factor": {note: Hint("number of copies")},
}

// keyspaceOptionCompletions is what to offer in the WITH clause of a keyspace,
// and whether the statement is one.
func keyspaceOptionCompletions(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) < 2 || words[1] != "KEYSPACE" {
		return nil, false
	}
	if words[0] != "CREATE" && words[0] != "ALTER" {
		return nil, false
	}

	// Inside the replication map.
	if open := strings.LastIndex(input, "{"); open >= 0 && !strings.Contains(input[open:], "}") {
		if key := keyAwaitingValue(input); key != "" {
			return replicationValue(key).suggestions(), true
		}
		return replicationKeys(input[open+1:]), true
	}

	switch {
	case endsWithWord(input, "WITH"), endsWithWord(input, "AND"):
		return withEquals(keyspaceOptionsLeft(input)), true

	case awaitingValue(input):
		switch lastOptionName(input) {
		case "replication":
			// The map written out is what everyone wants; the brace is for
			// anyone building one of their own.
			return append(append([]string{}, ReplicationTemplates...), "{'"), true
		case "durable_writes":
			return BooleanValues, true
		}
		return nil, false
	}

	// An option that has been given its value is followed by the AND that
	// joins the next one.
	if strings.HasSuffix(input, " ") && lastOptionName(input) != "" && len(keyspaceOptionsLeft(input)) > 0 {
		return []string{"AND"}, true
	}
	return nil, false
}

// keyspaceOptionsLeft is the options the clause has not been given already.
func keyspaceOptionsLeft(input string) []string {
	given := optionsGiven(input)

	left := make([]string, 0, len(KeyspaceOptions))
	for _, option := range KeyspaceOptions {
		if !given[option] {
			left = append(left, option)
		}
	}
	return left
}

// replicationKeys is what the replication map can still be given.
//
// A simple strategy takes the number of copies; a network topology strategy
// takes one for each datacenter, and those are named after the datacenters -
// names this cannot know, so it says what they are instead.
func replicationKeys(inside string) []string {
	written := keysIn(inside)

	if !written["class"] {
		return withColon([]string{"class"}, needsComma(inside))
	}

	var keys []string
	if strings.EqualFold(classIn(inside), "SimpleStrategy") {
		if !written["replication_factor"] {
			keys = append(keys, "replication_factor")
		}
	}

	offered := withColon(keys, needsComma(inside))
	if strings.EqualFold(classIn(inside), "NetworkTopologyStrategy") {
		offered = append(offered, Hint("'datacenter': copies"))
	}
	if needsComma(inside) {
		offered = append(offered, "}")
	}
	return offered
}

// replicationValue is what a key of the replication map takes. A key that is
// not one of the two named ones is a datacenter, and takes a number.
func replicationValue(key string) whatGoesHere {
	if answer, known := replicationValues[key]; known {
		return answer
	}
	return whatGoesHere{note: Hint("number of copies")}
}

// roleOptionCompletions is what to offer in the WITH clause of a role or a
// user, and whether the statement is one.
func roleOptionCompletions(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) < 2 {
		return nil, false
	}
	if words[0] != "CREATE" && words[0] != "ALTER" {
		return nil, false
	}

	switch words[1] {
	case "ROLE":
		return roleClause(input)
	case "USER":
		return userClause(input)
	}
	return nil, false
}

// roleClause is the options of a role: each takes an equals sign, except the
// ones that stand on their own.
func roleClause(input string) ([]string, bool) {
	// Inside the map of options the role manager takes, whose names are its
	// own rather than Cassandra's.
	if open := strings.LastIndex(input, "{"); open >= 0 && !strings.Contains(input[open:], "}") {
		// The braces hold a map of options, or a list of the places a role may
		// be used from.
		note := Hint("'name': 'value'")
		switch {
		case containsWord(input, "DATACENTERS"):
			note = Hint("'datacenter'")
		case containsWord(input, "CIDRS"):
			note = Hint("'cidr'")
		}

		if strings.TrimSpace(input[open+1:]) != "" {
			return []string{note, "}"}, true
		}
		return []string{note}, true
	}

	switch {
	case endsWithWord(input, "WITH"), endsWithWord(input, "AND"):
		return roleOptionsLeft(input), true

	// An option named by hand rather than taken from the list, which brings
	// its equals sign with it.
	case endsWithWord(input, "PASSWORD"), endsWithWord(input, "LOGIN"),
		endsWithWord(input, "SUPERUSER"), endsWithWord(input, "OPTIONS"):
		if !containsWord(input, "GENERATED") {
			return []string{"= "}, true
		}

	case awaitingValue(input):
		switch lastOptionName(input) {
		case "password", "hashed":
			return []string{Hint("'password'")}, true
		case "login", "superuser":
			return BooleanValues, true
		case "options":
			return []string{"{'"}, true
		}
		return nil, false
	}

	if !strings.HasSuffix(input, " ") {
		return nil, false
	}
	if roleOptionGiven(input) {
		return []string{"AND"}, true
	}

	// The name has been given and nothing else has: the options follow WITH.
	if len(strings.Fields(input)) > 2 && !containsWord(input, "WITH") {
		return []string{"WITH"}, true
	}
	return nil, false
}

// roleOptionsLeft is the options the clause has not been given already.
//
// Each is told by the words that distinguish it: a password is given one way
// or another, and access is granted to datacenters and from CIDRs separately.
func roleOptionsLeft(input string) []string {
	left := make([]string, 0, len(RoleOptions))

	for _, option := range RoleOptions {
		if _, given := afterWords(input, distinguishing(option)); !given {
			left = append(left, option)
		}
	}
	return left
}

// distinguishing is the words of an option that say which one it is.
func distinguishing(option string) string {
	words := strings.Fields(option)
	if strings.EqualFold(words[0], "ACCESS") {
		return strings.Join(words[:2], " ") // ACCESS TO, or ACCESS FROM
	}
	// A password is given one way or another and never two ways at once, so
	// any of them being written rules out the rest.
	if strings.Contains(strings.ToUpper(option), "PASSWORD") {
		return "PASSWORD"
	}
	return words[0]
}

// userClause is the password of a user, written without an equals sign, and
// whether it is a superuser.
func userClause(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))

	switch {
	case endsWithWord(input, "WITH"):
		return UserPasswords, true

	case endsWithWord(input, "PASSWORD"):
		return []string{Hint("'password'")}, true

	case len(words) > 2 && strings.HasSuffix(input, " "):
		// A user is a superuser or is not, and that is all there is after the
		// password.
		if containsWord(input, "SUPERUSER") || containsWord(input, "NOSUPERUSER") {
			return nil, false
		}
		if containsWord(input, "PASSWORD") && !quoteOpen(input) {
			return []string{"SUPERUSER", "NOSUPERUSER"}, true
		}
		return []string{"WITH", "SUPERUSER", "NOSUPERUSER"}, true
	}
	return nil, false
}

// roleOptionGiven reports whether the clause has an option with a value in it,
// which is what AND joins the next one to.
func roleOptionGiven(input string) bool {
	if lastOptionName(input) != "" {
		return true
	}
	return containsWord(input, "PASSWORD") || containsWord(input, "DATACENTERS") || containsWord(input, "CIDRS")
}

// quoteOpen reports whether a quote has been opened and not closed.
func quoteOpen(input string) bool {
	return strings.Count(input, "'")%2 == 1
}
