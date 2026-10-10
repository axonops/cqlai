package db

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// What a keyspace holds besides its tables, for the SCHEMA view's tree: its
// materialized views, indexes, types, functions, aggregates and triggers.

// The kinds of schema object, as the SCHEMA view groups them.
const (
	KindTables     = "tables"
	KindViews      = "views"
	KindIndexes    = "indexes"
	KindTypes      = "types"
	KindFunctions  = "functions"
	KindAggregates = "aggregates"
	KindTriggers   = "triggers"
)

// SchemaObjectKinds is the kinds after tables, in the order the tree lists
// them.
var SchemaObjectKinds = []string{KindViews, KindIndexes, KindTypes, KindFunctions, KindAggregates, KindTriggers}

// KeyspaceObjects is the names of what a keyspace holds besides its tables, by
// kind, each sorted. A function or aggregate overloaded for several argument
// types is one name. A trigger is named table.trigger: its name is the
// table's. A kind the cluster cannot list - system_schema is from 3.0 - is
// left empty.
func (s *Session) KeyspaceObjects(keyspace string) map[string][]string {
	objects := map[string][]string{}
	if s == nil || s.Session == nil || s.IsVirtualKeyspace(keyspace) {
		return objects
	}

	list := func(kind, query string, scan func(scan func(...any) bool) []string) {
		iter := s.Query(query, keyspace).Iter()
		names := scan(iter.Scan)
		if err := iter.Close(); err != nil {
			return
		}
		objects[kind] = uniqueSorted(names)
	}
	// names reads a query of one column.
	names := func(scan func(...any) bool) []string {
		var out []string
		var name string
		for scan(&name) {
			out = append(out, name)
		}
		return out
	}

	list(KindViews, `SELECT view_name FROM system_schema.views WHERE keyspace_name = ?`, names)
	list(KindIndexes, `SELECT index_name FROM system_schema.indexes WHERE keyspace_name = ?`, names)
	list(KindTypes, `SELECT type_name FROM system_schema.types WHERE keyspace_name = ?`, names)
	list(KindFunctions, `SELECT function_name FROM system_schema.functions WHERE keyspace_name = ?`, names)
	list(KindAggregates, `SELECT aggregate_name FROM system_schema.aggregates WHERE keyspace_name = ?`, names)
	list(KindTriggers, `SELECT table_name, trigger_name FROM system_schema.triggers WHERE keyspace_name = ?`,
		func(scan func(...any) bool) []string {
			var out []string
			var table, trigger string
			for scan(&table, &trigger) {
				out = append(out, table+"."+trigger)
			}
			return out
		})
	return objects
}

// AllSchemaNames is the names of every keyspace's tables and other objects,
// as TableNames and KeyspaceObjects give them for one: read with one query
// for each kind across all keyspaces. Filtering the SCHEMA view searches
// every keyspace, and asking each in turn took seven queries a keyspace and
// one more a table, with the screen frozen until they were done.
func (s *Session) AllSchemaNames() (tables map[string][]string, objects map[string]map[string][]string) {
	tables = map[string][]string{}
	objects = map[string]map[string][]string{}
	if s == nil || s.Session == nil {
		return tables, objects
	}

	// pairs reads (keyspace, name) rows into byKeyspace; ok is false when the
	// cluster cannot answer the query at all.
	pairs := func(query string, byKeyspace map[string][]string, name func(a, b string) string, three bool) bool {
		iter := s.Query(query).Iter()
		var ks, a, b string
		dest := []any{&ks, &a}
		if three {
			dest = append(dest, &b)
		}
		for iter.Scan(dest...) {
			byKeyspace[ks] = append(byKeyspace[ks], name(a, b))
		}
		return iter.Close() == nil
	}
	plain := func(a, _ string) string { return a }

	pairs(`SELECT keyspace_name, table_name FROM system_schema.tables`, tables, plain, false)
	// Virtual keyspaces, from 4.0; an older cluster has none to list.
	pairs(`SELECT keyspace_name, table_name FROM system_virtual_schema.tables`, tables, plain, false)
	for ks := range tables {
		tables[ks] = uniqueSorted(tables[ks])
	}

	kinds := []struct {
		kind, query string
		three       bool
		name        func(a, b string) string
	}{
		{KindViews, `SELECT keyspace_name, view_name FROM system_schema.views`, false, plain},
		{KindIndexes, `SELECT keyspace_name, index_name FROM system_schema.indexes`, false, plain},
		{KindTypes, `SELECT keyspace_name, type_name FROM system_schema.types`, false, plain},
		{KindFunctions, `SELECT keyspace_name, function_name FROM system_schema.functions`, false, plain},
		{KindAggregates, `SELECT keyspace_name, aggregate_name FROM system_schema.aggregates`, false, plain},
		{KindTriggers, `SELECT keyspace_name, table_name, trigger_name FROM system_schema.triggers`, true,
			func(table, trigger string) string { return table + "." + trigger }},
	}
	for ks := range tables {
		objects[ks] = map[string][]string{}
	}
	for _, k := range kinds {
		byKeyspace := map[string][]string{}
		if !pairs(k.query, byKeyspace, k.name, k.three) {
			continue
		}
		for ks, names := range byKeyspace {
			if objects[ks] == nil {
				objects[ks] = map[string][]string{}
			}
			objects[ks][k.kind] = uniqueSorted(names)
		}
	}
	return tables, objects
}

func uniqueSorted(names []string) []string {
	seen := map[string]bool{}
	out := names[:0]
	for _, n := range names {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// DescribeSchemaObject is the CREATE statement for one object of a keyspace,
// as DESCRIBE gives it. From Cassandra 4.0 the server writes it, as it does
// for DESCRIBE KEYSPACE; before that, or when the server cannot, it is built
// from system_schema the way DESCRIBE KEYSPACE builds it. A function or
// aggregate with several overloads is all of them.
func (s *Session) DescribeSchemaObject(kind, keyspace, name string) (string, error) {
	if s == nil || s.Session == nil {
		return "", fmt.Errorf("not connected")
	}

	if statement := describeStatement(kind, keyspace, name); statement != "" && s.IsVersion4OrHigher() {
		if text, err := s.describeOnServer(statement); err == nil {
			return text, nil
		}
	}

	switch kind {
	case KindViews:
		info, err := s.DescribeMaterializedViewQuery(keyspace, name)
		if err != nil {
			return "", err
		}
		return formatMaterializedViewCreateStatement(keyspace, info), nil

	case KindIndexes:
		info, err := s.DescribeIndexQuery(keyspace, name)
		if err != nil {
			return "", err
		}
		return formatIndexCreateStatement(keyspace, info), nil

	case KindTypes:
		info, err := s.DescribeTypeQuery(keyspace, name)
		if err != nil {
			return "", err
		}
		return formatTypeCreateStatement(keyspace, info), nil

	case KindFunctions:
		details, err := s.DescribeFunctionQuery(keyspace, name)
		if err != nil {
			return "", err
		}
		parts := make([]string, 0, len(details))
		for i := range details {
			parts = append(parts, formatFunctionCreateStatement(keyspace, &details[i]))
		}
		if len(parts) == 0 {
			return "", fmt.Errorf("function %s.%s not found", keyspace, name)
		}
		return strings.Join(parts, "\n\n"), nil

	case KindAggregates:
		info, err := s.DescribeAggregateQuery(keyspace, name)
		if err != nil {
			return "", err
		}
		return formatAggregateCreateStatement(keyspace, info), nil

	case KindTriggers:
		return s.describeTrigger(keyspace, name)
	}
	return "", fmt.Errorf("%s is not a kind of schema object", kind)
}

// describeStatement is the DESCRIBE the server answers for an object, or ""
// for a kind it has no DESCRIBE of.
func describeStatement(kind, keyspace, name string) string {
	qualified := QuoteName(keyspace) + "." + QuoteName(name)
	switch kind {
	case KindViews:
		return "DESCRIBE MATERIALIZED VIEW " + qualified
	case KindIndexes:
		return "DESCRIBE INDEX " + qualified
	case KindTypes:
		return "DESCRIBE TYPE " + qualified
	case KindFunctions:
		return "DESCRIBE FUNCTION " + qualified
	case KindAggregates:
		return "DESCRIBE AGGREGATE " + qualified
	}
	return ""
}

// describeTrigger is a trigger's CREATE TRIGGER, from system_schema: there is
// no DESCRIBE TRIGGER. name is table.trigger.
func (s *Session) describeTrigger(keyspace, name string) (string, error) {
	table, trigger, ok := strings.Cut(name, ".")
	if !ok {
		return "", fmt.Errorf("trigger %s has to be named table.trigger", name)
	}
	var options map[string]string
	if err := s.Query(`SELECT options FROM system_schema.triggers WHERE keyspace_name = ? AND table_name = ? AND trigger_name = ?`,
		keyspace, table, trigger).Scan(&options); err != nil {
		return "", fmt.Errorf("trigger %s on %s.%s: %w", trigger, keyspace, table, err)
	}
	return fmt.Sprintf("CREATE TRIGGER %s ON %s.%s USING '%s';",
		QuoteName(trigger), QuoteName(keyspace), QuoteName(table), strings.ReplaceAll(options["class"], "'", "''")), nil
}

// plainName is a name CQL takes as it is: lower case, digits and underscores,
// not starting with a digit.
var plainName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// QuoteName writes a name so CQL reads it back as it is: in double quotes when
// it has capitals or anything but letters, digits and underscores.
func QuoteName(name string) string {
	if plainName.MatchString(name) {
		return name
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
