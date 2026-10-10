package policy

import (
	"fmt"
	"slices"
	"strings"
)

// Rows of a system table that describe another keyspace.
//
// system_schema, system.size_estimates and many of the virtual tables hold a
// row for every keyspace and table: its name, its columns, its size. Seeing
// such a table is not seeing what is in it, so each row is filtered by what it
// is about, and a row about a hidden keyspace or table is left out.
//
// That needs the row to say what it is about. A SELECT on such a table has to
// return the columns that name the keyspace and the table, under their own
// names: left out, or given another column's name with an alias, the filter
// would have nothing to go on.

// keyspaceColumn names the keyspace a row is about.
const keyspaceColumn = "keyspace_name"

// tableColumns name the table a row is about, in the tables that have them.
var tableColumns = []string{"table_name", "view_name", "base_table_name", "columnfamily_name"}

// namingColumns is the columns a SELECT on keyspace.table has to return for
// its rows to be filtered: none for a table outside the system keyspaces, or
// one whose rows are not about other keyspaces.
//
// A table's rows are about a keyspace when keyspace_name is in its partition
// key: system_schema's tables, size_estimates, and the virtual tables kept by
// table. In another system table, such as system_views.clients, it is only
// something a row says, and a hidden one is blanked rather than the row
// dropped. A system table whose key cannot be found is taken to be about
// keyspaces.
func (p Policy) namingColumns(keyspace, table string) []string {
	if !IsSystemKeyspace(keyspace) || !p.hidesKeyspaces() {
		return nil
	}
	var key, columns []string
	if p.partitionKey != nil {
		key = p.partitionKey(keyspace, table)
	}
	if p.columns != nil {
		columns = p.columns(keyspace, table)
	}
	if len(key) == 0 || len(columns) == 0 {
		return []string{keyspaceColumn}
	}
	inKey := false
	for _, c := range key {
		inKey = inKey || strings.EqualFold(c, keyspaceColumn)
	}
	if !inKey {
		return nil
	}
	has := map[string]bool{}
	for _, c := range columns {
		has[strings.ToLower(c)] = true
	}
	need := []string{keyspaceColumn}
	for _, c := range tableColumns {
		if has[c] {
			need = append(need, c)
		}
	}
	return need
}

// hidesKeyspaces reports whether any keyspace or table is hidden by the
// settings: a keyspace list, a deny list, or the system keyspaces left out.
// When nothing is, a system table's rows are read as they are.
func (p Policy) hidesKeyspaces() bool {
	return !p.allKeyspaces || !p.systemVisible || len(p.deny) > 0
}

// describedVisible reports whether a row about a keyspace may be shown.
// system_auth's schema is the same in every cluster, so describing it hides
// nothing; its data stays hidden, because system_auth is never readable.
func (p Policy) describedVisible(keyspace string) bool {
	if strings.EqualFold(keyspace, "system_auth") {
		return !p.userDenied(keyspace, "")
	}
	return p.Visible(keyspace)
}

// userDenied reports whether the settings' own deny list hides it.
func (p Policy) userDenied(keyspace, table string) bool {
	return Policy{deny: p.deny}.deniedBySettings(keyspace, table)
}

// MissingNamingColumns is the columns a result from keyspace.table lacks
// that its rows have to carry to be filtered. A result that lacks any is not
// returned.
func (p Policy) MissingNamingColumns(keyspace, table string, result []string) []string {
	got := map[string]bool{}
	for _, c := range result {
		got[strings.ToLower(c)] = true
	}
	var missing []string
	for _, c := range p.namingColumns(keyspace, table) {
		if !got[c] {
			missing = append(missing, c)
		}
	}
	return missing
}

// RowVisible reports whether a row of keyspace.table is about something
// visible. Every row of a table outside the system keyspaces is.
func (p Policy) RowVisible(keyspace, table string, row map[string]any) bool {
	if len(p.namingColumns(keyspace, table)) == 0 {
		return true
	}
	ks, _ := row[keyspaceColumn].(string)
	if !p.describedVisible(ks) {
		return false
	}
	for _, c := range tableColumns {
		if name, _ := row[c].(string); name != "" && p.userDenied(ks, name) {
			return false
		}
	}
	return true
}

// NameHidden reports whether a value of a system table names a hidden keyspace
// or table, in a row that is not dropped for it: system_views.clients says
// which keyspace each client is using. The value is blanked.
func (p Policy) NameHidden(keyspace, table, column string, row map[string]any) bool {
	if !IsSystemKeyspace(keyspace) || len(p.namingColumns(keyspace, table)) > 0 {
		return false
	}
	if !p.hidesKeyspaces() {
		return false
	}
	ks, _ := row[keyspaceColumn].(string)
	switch {
	case strings.EqualFold(column, keyspaceColumn):
		return ks != "" && !p.describedVisible(ks)
	case slices.Contains(tableColumns, strings.ToLower(column)):
		name, _ := row[column].(string)
		return name != "" && (ks == "" || !p.describedVisible(ks) || p.userDenied(ks, name))
	}
	return false
}

// checkNamingColumns refuses a SELECT on a system table whose rows describe
// other keyspaces, unless it returns columns under their own names.
func (p Policy) checkNamingColumns(name string, keyspace, table string, plain bool) error {
	need := p.namingColumns(keyspace, table)
	if len(need) == 0 || plain {
		return nil
	}
	return refused(fmt.Sprintf("%s describes other keyspaces, and its rows are filtered by %s: select with * or by column name, including those, with no JSON, alias or function",
		name, strings.Join(need, " and ")))
}

// NamingColumnsRefusal is why a result that lacks the naming columns is not
// returned.
func NamingColumnsRefusal(keyspace, table string, missing []string) Refusal {
	return Refusal{Reason: fmt.Sprintf("%s.%s describes other keyspaces, and its rows are filtered by what they are about: select %s too",
		keyspace, table, strings.Join(missing, " and "))}
}
