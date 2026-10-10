package db

import (
	"encoding/binary"
	"fmt"
	"reflect"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// Decoding the values gocql cannot.
//
// gocql builds a Go value for each column from its type, and for a map that
// is a Go map with the key's Go type. A blob or inet key is a []byte and a
// collection key a slice or a map, and Go maps cannot have those as keys:
// gocql panics building the type. Such a column is read as bytes and decoded
// here, with the key held as the text it is shown as.

// gocqlCanHold reports whether gocql can make a Go value for info and every
// type inside it.
func gocqlCanHold(info gocql.TypeInfo) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	switch t := info.(type) {
	case gocql.CollectionType:
		if t.Key != nil && (!gocqlCanHold(t.Key) || !comparable(t.Key)) {
			return false
		}
		if t.Elem != nil && !gocqlCanHold(t.Elem) {
			return false
		}
	case gocql.TupleTypeInfo:
		for _, e := range t.Elems {
			if !gocqlCanHold(e) {
				return false
			}
		}
	case gocql.UDTTypeInfo:
		for _, f := range t.Elements {
			if !gocqlCanHold(f.Type) {
				return false
			}
		}
	}
	_ = info.Zero()
	return true
}

// comparable reports whether a type's Go value can be a map key.
func comparable(info gocql.TypeInfo) bool {
	zero := info.Zero()
	return zero == nil || reflect.TypeOf(zero).Comparable()
}

// rawValue is a scan destination that keeps a column's bytes and its type,
// for decodeTyped.
type rawValue struct {
	info gocql.TypeInfo
	data []byte
}

// UnmarshalCQL keeps the bytes: nil for a NULL.
func (r *rawValue) UnmarshalCQL(_ gocql.TypeInfo, data []byte) error {
	if data == nil {
		r.data = nil
		return nil
	}
	r.data = append([]byte{}, data...)
	return nil
}

// decodeTyped decodes a value of type info: a collection, tuple or UDT here,
// and anything else by gocql, which can.
func decodeTyped(info gocql.TypeInfo, data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	switch t := info.(type) {
	case gocql.CollectionType:
		items, err := readItems(data, t.Type() == gocql.TypeMap)
		if err != nil {
			return nil, err
		}
		if t.Type() != gocql.TypeMap {
			list := make([]interface{}, len(items))
			for i, item := range items {
				if list[i], err = decodeTyped(t.Elem, item); err != nil {
					return nil, err
				}
			}
			return list, nil
		}
		m := make(map[interface{}]interface{}, len(items)/2)
		for i := 0; i+1 < len(items); i += 2 {
			key, err := decodeTyped(t.Key, items[i])
			if err != nil {
				return nil, err
			}
			value, err := decodeTyped(t.Elem, items[i+1])
			if err != nil {
				return nil, err
			}
			m[mapKey(key)] = value
		}
		return m, nil
	case gocql.TupleTypeInfo:
		fields, err := readFields(data, len(t.Elems))
		if err != nil {
			return nil, err
		}
		tuple := make([]interface{}, len(t.Elems))
		for i, e := range t.Elems {
			if tuple[i], err = decodeTyped(e, fields[i]); err != nil {
				return nil, err
			}
		}
		return tuple, nil
	case gocql.UDTTypeInfo:
		fields, err := readFields(data, len(t.Elements))
		if err != nil {
			return nil, err
		}
		udt := make(map[string]interface{}, len(t.Elements))
		for i, f := range t.Elements {
			if udt[f.Name], err = decodeTyped(f.Type, fields[i]); err != nil {
				return nil, err
			}
		}
		return udt, nil
	}

	dest := NewScanDest(info)
	if err := gocql.Unmarshal(info, data, dest); err != nil {
		return nil, err
	}
	return ScanValue(dest), nil
}

// readItems reads a collection's items, each [int32 length][bytes] with -1 for
// NULL; a map has a key and a value for each of its count.
func readItems(data []byte, isMap bool) ([][]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid collection data")
	}
	count := int(int32(binary.BigEndian.Uint32(data))) // #nosec G115 - a signed count on the wire
	if count < 0 {
		return nil, fmt.Errorf("invalid collection count: %d", count)
	}
	if isMap {
		count *= 2
	}
	return readValues(data[4:], count, false)
}

// readFields reads up to n fields of a tuple or UDT. A UDT written before a
// field was added has fewer: the rest are NULL.
func readFields(data []byte, n int) ([][]byte, error) {
	return readValues(data, n, true)
}

func readValues(data []byte, n int, mayEnd bool) ([][]byte, error) {
	values := make([][]byte, n)
	pos := 0
	for i := 0; i < n; i++ {
		if pos == len(data) && mayEnd {
			break
		}
		if pos+4 > len(data) {
			return nil, fmt.Errorf("invalid data at item %d", i)
		}
		size := int(int32(binary.BigEndian.Uint32(data[pos:]))) // #nosec G115 - a signed length on the wire
		pos += 4
		if size < 0 {
			continue // NULL
		}
		if pos+size > len(data) {
			return nil, fmt.Errorf("invalid data at item %d", i)
		}
		values[i] = data[pos : pos+size]
		pos += size
	}
	return values, nil
}
