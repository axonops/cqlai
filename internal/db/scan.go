package db

import (
	"reflect"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// Scanning a row without losing NULLs.
//
// The obvious destination for a column of unknown type is *interface{}, and it
// does not work. When gocql unmarshals a NULL into one it takes the type
// currently inside the interface and stores that type's zero value:
//
//   - on the first row the interface is empty, there is no type to take, and
//     gocql panics with "reflect: call of reflect.Value.Type on zero Value"
//   - on later rows the interface still holds the previous row's value, so the
//     NULL is silently recorded as ""  or 0 or false
//
// Reusing one destination across rows makes both happen in the same scan, which
// is why a table whose first row is NULL crashes while the same table with a
// populated row first exports quietly wrong data.
//
// Scanning into a pointer to a pointer avoids it. gocql sets the inner pointer
// to nil for a NULL and allocates a value otherwise, so the two are told apart
// without guessing.

// NewScanDest returns a scan destination for a column that can represent NULL.
//
// Pass the column's TypeInfo from Iter.Columns(). The result goes straight into
// Iter.Scan, and ScanValue reads it back.
func NewScanDest(info gocql.TypeInfo) interface{} {
	if info == nil {
		return new(interface{})
	}

	// gocql populates a UDT only when the destination is a map, and it already
	// leaves that map nil for a NULL, so there is nothing to fix here.
	if info.Type() == gocql.TypeUDT {
		return new(map[string]interface{})
	}

	zero := info.Zero()
	if zero == nil {
		return new(interface{})
	}

	// **T: gocql nils the inner pointer for a NULL.
	return reflect.New(reflect.PointerTo(reflect.TypeOf(zero))).Interface()
}

// ScanValue reads back a destination from NewScanDest, returning nil for a NULL.
func ScanValue(dest interface{}) interface{} {
	if dest == nil {
		return nil
	}

	if udt, ok := dest.(*map[string]interface{}); ok {
		if udt == nil || *udt == nil {
			return nil
		}
		return *udt
	}

	ref := reflect.ValueOf(dest)
	if ref.Kind() != reflect.Pointer || ref.IsNil() {
		return nil
	}

	inner := ref.Elem()
	if inner.Kind() == reflect.Pointer {
		if inner.IsNil() {
			return nil // NULL
		}
		return inner.Elem().Interface()
	}

	// Fell back to *interface{} for a column with no usable type information.
	return inner.Interface()
}
