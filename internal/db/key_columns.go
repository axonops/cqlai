package db

import (
	"fmt"
	"regexp"
	"strings"
)

// The markers appended to result headers for the columns that make up a
// table's primary key.
//
// One place builds them and one place strips them. They used to be written out
// by hand wherever they were needed - ten files strip them before writing CSV,
// Parquet or a capture - so changing what a marker looks like meant finding
// every one of those and would silently break exports if any were missed.

// KeyColumns describes which columns of a table form its primary key.
type KeyColumns map[string]KeyColumnInfo

// Marker is the suffix to append to a column's header, or "" if the column is
// not part of the key.
//
// The component number appears only when there is more than one component to
// tell apart. A composite partition key on (tenant, region) reads "(PK1)" and
// "(PK2)", because the order is part of the key and cannot be read off the
// column order on screen. A key with one partition column reads "(PK)", since
// a number there says nothing.
func (k KeyColumns) Marker(name string) string {
	info, ok := k[name]
	if !ok {
		return ""
	}

	var label string
	switch info.Kind {
	case "partition_key":
		label = "PK"
	case "clustering":
		label = "C"
	default:
		return ""
	}

	if k.count(info.Kind) > 1 {
		// system_schema counts components from zero; people count from one.
		return fmt.Sprintf(" (%s%d)", label, info.Position+1)
	}
	return " (" + label + ")"
}

// count is how many columns are of a kind.
func (k KeyColumns) count(kind string) int {
	n := 0
	for _, info := range k {
		if info.Kind == kind {
			n++
		}
	}
	return n
}

// keyMarker matches a marker at the end of a header: (PK), (C), or either with
// a component number. The label is captured so KeyLabel can read it back.
var keyMarker = regexp.MustCompile(`\s+\(((?:PK|C)\d*)\)$`)

// KeyLabel returns the key marker of a header without its brackets - PK, PK2,
// C, C1 - or "" for a column that is neither.
//
// The header row shows the name alone and the row under it shows the label
// beside the type, so the two are drawn apart. Reading the label back out of
// the header keeps the bracket format in this file, next to Marker, which is
// what writes it: two places knowing that shape is how the ten hand-rolled
// strippers in #118 happened.
func KeyLabel(header string) string {
	m := keyMarker.FindStringSubmatch(header)
	if m == nil {
		return ""
	}
	return m[1]
}

// StripKeyMarker removes the key marker from a header, leaving the column name
// as Cassandra knows it.
//
// Anything writing a column name back out - a CSV header, a Parquet field, a
// capture file, a COPY - goes through here, so an export can be read back in.
func StripKeyMarker(header string) string {
	return keyMarker.ReplaceAllString(header, "")
}

// StripKeyMarkers removes the markers from a row of headers.
func StripKeyMarkers(headers []string) []string {
	clean := make([]string, len(headers))
	for i, header := range headers {
		clean[i] = StripKeyMarker(header)
	}
	return clean
}

// fromClause picks the keyspace and table out of a SELECT. Either part may be
// quoted, which is how a name keeps its capitals in Cassandra.
var fromClause = regexp.MustCompile(`(?i)FROM\s+(?:("[^"]+"|[a-zA-Z_][a-zA-Z0-9_]*)\.)?("[^"]+"|[a-zA-Z_][a-zA-Z0-9_]*)`)

// cqlIdentifier resolves a name the way Cassandra does.
//
// An unquoted name is folded to lower case, so SELECT * FROM Users is the table
// "users"; a quoted one keeps exactly the case it was created with. Looking the
// name up as it was typed meant a mixed-case query found nothing in
// system_schema, and the markers quietly did not appear.
func cqlIdentifier(name string) string {
	if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
		return name[1 : len(name)-1]
	}
	return strings.ToLower(name)
}
