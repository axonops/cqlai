package db

import (
	"fmt"
	"sort"
	"strings"
)

// Virtual keyspaces and tables.
//
// Cassandra 4.0 added tables that are not stored but computed on the node that
// is asked - system_views.clients, .settings, .thread_pools and forty-odd more.
// They are listed in system_virtual_schema rather than system_schema, so every
// question cqlai asked about what exists - for USE, for DESCRIBE KEYSPACES, for
// the schema browser, for the keyspace chooser, for completion - was asked of
// the one place they are not. You could SELECT from system_views.clients and
// not USE system_views, or see it in a list of keyspaces.
//
// These are the one place that asks the other. Before 4.0 there is no
// system_virtual_schema, the question fails, and the answer is that there are
// none - which is true.

// VirtualKeyspaces is the keyspaces the node computes rather than stores, by
// name, sorted. None before Cassandra 4.0.
func (s *Session) VirtualKeyspaces() []string {
	return s.virtualNames("SELECT keyspace_name FROM system_virtual_schema.keyspaces")
}

// VirtualTables is the tables of a virtual keyspace, by name, sorted. None for
// a keyspace that is not virtual.
func (s *Session) VirtualTables(keyspace string) []string {
	return s.virtualNames(
		"SELECT table_name FROM system_virtual_schema.tables WHERE keyspace_name = ?", keyspace)
}

// VirtualColumns is the columns of a virtual table, by name.
func (s *Session) VirtualColumns(keyspace, table string) []string {
	return s.virtualNames(
		"SELECT column_name FROM system_virtual_schema.columns WHERE keyspace_name = ? AND table_name = ?",
		keyspace, table)
}

// virtualTableList is the tables of a virtual keyspace as DESCRIBE TABLES
// lists them, with their keys.
//
// The keys matter more here than for a stored table: a virtual table can only
// be filtered on its key, so the key is what says how to ask it anything. One
// query for the keyspace rather than one per table - there are forty-odd.
func (s *Session) virtualTableList(keyspace string) []TableListInfo {
	type keys struct{ partition, clustering []string }
	byTable := map[string]*keys{}

	iter := s.Query(`SELECT table_name, column_name, kind, position
	                 FROM system_virtual_schema.columns WHERE keyspace_name = ?`, keyspace).Iter()
	var table, column, kind string
	var position int
	for iter.Scan(&table, &column, &kind, &position) {
		k := byTable[table]
		if k == nil {
			k = &keys{}
			byTable[table] = k
		}
		switch kind {
		case "partition_key":
			k.partition = placeKey(k.partition, position, column)
		case "clustering":
			k.clustering = placeKey(k.clustering, position, column)
		}
	}
	_ = iter.Close()

	var tables []TableListInfo
	for _, name := range s.VirtualTables(keyspace) {
		t := TableListInfo{Name: name, Keyspace: keyspace, Virtual: true}
		if k := byTable[name]; k != nil {
			t.PartitionKeys, t.ClusteringKeys = k.partition, k.clustering
		}
		tables = append(tables, t)
	}
	return tables
}

// placeKey puts a key column at its position, growing the list to fit.
func placeKey(keys []string, position int, column string) []string {
	for len(keys) <= position {
		keys = append(keys, "")
	}
	if position >= 0 {
		keys[position] = column
	}
	return keys
}

// IsVirtualKeyspace reports whether a keyspace is one of the virtual ones.
func (s *Session) IsVirtualKeyspace(keyspace string) bool {
	for _, name := range s.VirtualKeyspaces() {
		if name == keyspace {
			return true
		}
	}
	return false
}

// virtualNames is the first column of a query against system_virtual_schema.
//
// An error is the answer "none", not a failure to report: it is what a cluster
// older than 4.0 says, having no system_virtual_schema to ask.
func (s *Session) virtualNames(query string, values ...interface{}) []string {
	if s == nil || s.Session == nil {
		return nil
	}

	iter := s.Query(query, values...).Iter()
	var names []string
	var name string
	for iter.Scan(&name) {
		names = append(names, name)
	}
	if err := iter.Close(); err != nil {
		return nil
	}

	sort.Strings(names)
	return names
}

// DescribeVirtualTable is a virtual table's definition, as the server gives it:
// the structure, commented out, under a warning that it cannot be recreated
// with CQL - the form cqlsh shows.
//
// The schema browser builds a stored table's definition from system_schema,
// where a virtual table is not, so it said a table it had just listed did not
// exist.
func (s *Session) DescribeVirtualTable(keyspace, table string) (string, error) {
	return s.describeOnServer(fmt.Sprintf("DESCRIBE TABLE %s.%s", keyspace, table))
}

// describeOnServer runs a DESCRIBE statement on the node and joins the
// statements it gives back.
//
// Only the server can describe a virtual keyspace: nothing about it is in
// system_schema to build a description from. The rows carry keyspace_name,
// type, name and create_statement, and it is the last of those that is the
// description - the first is the keyspace's name, once per object.
func (s *Session) describeOnServer(statement string) (string, error) {
	iter := s.Query(statement).Iter()

	var parts []string
	row := map[string]interface{}{}
	for iter.MapScan(row) {
		if create, ok := row["create_statement"]; ok {
			parts = append(parts, strings.TrimSpace(fmt.Sprint(create)))
		}
		row = map[string]interface{}{}
	}
	if err := iter.Close(); err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("the server described nothing for %s", statement)
	}
	return strings.Join(parts, "\n\n"), nil
}

// virtualColumnInfo is a virtual table's columns with their types and what
// part of the key each is, keys first, as the schema cache keeps a stored
// table's.
func (s *Session) virtualColumnInfo(keyspace, table string) []ColumnInfo {
	if s == nil || s.Session == nil {
		return nil
	}
	iter := s.Query(`SELECT column_name, type, kind, position
	                 FROM system_virtual_schema.columns WHERE keyspace_name = ? AND table_name = ?`,
		keyspace, table).Iter()

	var keys, regular []ColumnInfo
	var name, dataType, kind string
	var position int
	for iter.Scan(&name, &dataType, &kind, &position) {
		c := ColumnInfo{Name: name, DataType: dataType, Kind: kind, Position: position}
		if kind == "regular" || kind == "static" {
			c.Position = -1
			regular = append(regular, c)
			continue
		}
		keys = append(keys, c)
	}
	if err := iter.Close(); err != nil {
		return nil
	}

	order := map[string]int{"partition_key": 0, "clustering": 1}
	sort.SliceStable(keys, func(i, j int) bool {
		if order[keys[i].Kind] != order[keys[j].Kind] {
			return order[keys[i].Kind] < order[keys[j].Kind]
		}
		return keys[i].Position < keys[j].Position
	})
	sort.SliceStable(regular, func(i, j int) bool { return regular[i].Name < regular[j].Name })
	return append(keys, regular...)
}

// TableKey is a table's partition key and all its columns, stored or virtual.
// Empty when the table cannot be found.
func (s *Session) TableKey(keyspace, table string) (partitionKey, columns []string) {
	if s == nil || s.Session == nil {
		return nil, nil
	}
	if meta, err := s.GetTableMetadata(keyspace, table); err == nil && meta != nil {
		for _, c := range meta.PartitionKey {
			partitionKey = append(partitionKey, c.Name)
		}
		for name := range meta.Columns {
			columns = append(columns, name)
		}
		sort.Strings(columns)
		return partitionKey, columns
	}
	for _, c := range s.virtualColumnInfo(keyspace, table) {
		if c.Kind == "partition_key" {
			partitionKey = append(partitionKey, c.Name)
		}
		columns = append(columns, c.Name)
	}
	return partitionKey, columns
}

// PrimaryKey is a table's partition key and clustering key, stored or
// virtual, in order. Empty when the table cannot be found.
func (s *Session) PrimaryKey(keyspace, table string) (partition, clustering []string) {
	if s == nil || s.Session == nil {
		return nil, nil
	}
	if meta, err := s.GetTableMetadata(keyspace, table); err == nil && meta != nil {
		for _, c := range meta.PartitionKey {
			partition = append(partition, c.Name)
		}
		for _, c := range meta.ClusteringColumns {
			clustering = append(clustering, c.Name)
		}
		return partition, clustering
	}
	for _, c := range s.virtualColumnInfo(keyspace, table) {
		switch c.Kind {
		case "partition_key":
			partition = append(partition, c.Name)
		case "clustering":
			clustering = append(clustering, c.Name)
		}
	}
	return partition, clustering
}
