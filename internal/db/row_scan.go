package db

import (
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// ScanRow reads the next row as Iter.MapScan does - each column's value under
// its name - but with two differences that MapScan gets wrong for display.
//
// A NULL is nil. MapScan unmarshals into each type's zero value, so a NULL int
// came back as 0, a NULL text as "", a NULL boolean as false and a NULL uuid
// as the zero UUID: shown, exported and re-imported as values that were never
// there. The destinations are NewScanDest's, which tell the two apart.
//
// A tuple is one value under its own name, a []interface{} of its elements.
// MapScan files each element under a name of its own, "t[0]", "t[1]", so the
// column itself was never found and every tuple read as null.
//
// Every column's key is written, so a map reused across rows holds this row's
// values and no other's.
func ScanRow(iter *gocql.Iter, row map[string]interface{}) bool {
	cols := iter.Columns()
	dests := make([][]interface{}, len(cols))
	var all []interface{}
	for i, c := range cols {
		dests[i] = columnDests(c.TypeInfo)
		all = append(all, dests[i]...)
	}
	if !iter.Scan(all...) {
		return false
	}
	for i, c := range cols {
		row[c.Name] = columnValue(c.TypeInfo, dests[i])
	}
	return true
}

// columnDests is where to scan one column: one destination, or one for each
// element of a tuple, which the driver reads as that many columns.
func columnDests(info gocql.TypeInfo) []interface{} {
	if tuple, ok := info.(gocql.TupleTypeInfo); ok {
		dests := make([]interface{}, len(tuple.Elems))
		for i, elem := range tuple.Elems {
			dests[i] = NewScanDest(elem)
		}
		return dests
	}
	return []interface{}{NewScanDest(info)}
}

// columnValue is a column's value from its destinations: a tuple's elements
// as one []interface{}, and nil for a NULL - a NULL tuple comes back as every
// element NULL.
func columnValue(info gocql.TypeInfo, dests []interface{}) interface{} {
	if _, ok := info.(gocql.TupleTypeInfo); !ok {
		return ScanValue(dests[0])
	}
	elems := make([]interface{}, len(dests))
	null := true
	for i, d := range dests {
		elems[i] = ScanValue(d)
		null = null && elems[i] == nil
	}
	if null {
		return nil
	}
	return elems
}
