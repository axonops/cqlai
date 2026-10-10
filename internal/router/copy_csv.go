package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

// CSV for COPY, read and written by Cassandra's own JSON.
//
// COPY TO reads each row with SELECT JSON and writes each column's JSON value:
// a string as it is, a number or boolean as written, and a collection, UDT or
// tuple as its JSON. COPY FROM gives each field back to INSERT JSON as a
// string, which Cassandra reads as the column's type. Nothing is guessed from
// the text: 02134 in a text column stays text, and a timestamp, blob, uuid,
// duration or collection is read by Cassandra as it reads any value.

// tableColumn is one column of a table.
type tableColumn struct {
	name     string
	kind     string // partition_key, clustering, static or regular
	position int
	cqlType  string
}

// tableColumnsInOrder is a table's columns in the order SELECT * returns
// them: the partition key and the clustering columns by position, then the
// static and the regular columns by name.
func (h *MetaCommandHandler) tableColumnsInOrder(table string) ([]tableColumn, error) {
	keyspace, name := h.splitTableName(table)
	if keyspace == "" {
		return nil, fmt.Errorf("no keyspace for %s: name it as keyspace.table, or USE one", table)
	}
	if h.session == nil || h.session.Session == nil {
		return nil, fmt.Errorf("not connected")
	}
	iter := h.session.Query(`SELECT column_name, kind, position, type FROM system_schema.columns
		WHERE keyspace_name = ? AND table_name = ?`, keyspace, name).Iter()
	var columns []tableColumn
	var c tableColumn
	for iter.Scan(&c.name, &c.kind, &c.position, &c.cqlType) {
		columns = append(columns, c)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	rank := map[string]int{"partition_key": 0, "clustering": 1, "static": 2, "regular": 3}
	sort.SliceStable(columns, func(i, j int) bool {
		a, b := columns[i], columns[j]
		if rank[a.kind] != rank[b.kind] {
			return rank[a.kind] < rank[b.kind]
		}
		if a.kind == "partition_key" || a.kind == "clustering" {
			return a.position < b.position
		}
		return a.name < b.name
	})
	return columns, nil
}

// splitTableName is keyspace.table, or table in the current keyspace, each
// as Cassandra stores it: a quoted name as written, any other in lower case.
func (h *MetaCommandHandler) splitTableName(table string) (string, string) {
	keyspace, name, qualified := cutName(table)
	if !qualified {
		name, keyspace = keyspace, ""
		if h.sessionManager != nil {
			keyspace = h.sessionManager.CurrentKeyspace()
		}
		if keyspace == "" && h.session != nil {
			keyspace = h.session.Keyspace()
		}
		return keyspace, storedName(name)
	}
	return storedName(keyspace), storedName(name)
}

// cutName splits keyspace.name at the first dot outside double quotes.
func cutName(s string) (string, string, bool) {
	quoted := false
	for i, c := range s {
		switch {
		case c == '"':
			quoted = !quoted
		case c == '.' && !quoted:
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}

// storedName is a name as Cassandra stores it.
func storedName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) >= 2 && strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
		return strings.ReplaceAll(name[1:len(name)-1], `""`, `"`)
	}
	return strings.ToLower(name)
}

// jsonFields is a SELECT JSON row's values, in the order of the selection.
func jsonFields(doc string) ([]json.RawMessage, error) {
	dec := json.NewDecoder(strings.NewReader(doc))
	dec.UseNumber()
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("not a JSON object: %.40s", doc)
	}
	var fields []json.RawMessage
	for dec.More() {
		if _, err := dec.Token(); err != nil { // the column's name
			return nil, err
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		fields = append(fields, raw)
	}
	return fields, nil
}

// csvField is a column's CSV field from its JSON value. A timestamp is
// written as the shell shows it, with its milliseconds and zone.
func csvField(raw json.RawMessage, cqlType, nullVal string) string {
	raw = bytes.TrimSpace(raw)
	switch {
	case len(raw) == 0 || string(raw) == "null":
		return nullVal
	case raw[0] == '"':
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return string(raw)
		}
		if cqlType == "timestamp" {
			if t, err := time.Parse("2006-01-02 15:04:05.999Z07:00", s); err == nil {
				return t.UTC().Format(db.TimestampLayout)
			}
		}
		return s
	default:
		// A number, a boolean, or a collection, UDT or tuple as JSON.
		return string(raw)
	}
}

// jsonRow is an INSERT JSON document for one CSV record: each field as a
// string, for Cassandra to read as its column's type, and nullVal as NULL.
func jsonRow(columns, record []string, nullVal string) (string, error) {
	doc := make(map[string]interface{}, len(columns))
	for i, column := range columns {
		// A quoted name keeps its quotes: INSERT JSON reads a key in quotes as
		// case-sensitive, and any other in lower case.
		key := strings.TrimSpace(column)
		if record[i] == nullVal {
			doc[key] = nil
		} else {
			doc[key] = record[i]
		}
	}
	b, err := json.Marshal(doc)
	return string(b), err
}
