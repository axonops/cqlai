package db

import (
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// RawBytes is a custom type to capture raw bytes from gocql for types like UDTs
// that gocql cannot decode without compile-time type information
type RawBytes []byte

// UnmarshalCQL implements gocql.Unmarshaler interface
// This allows us to get the raw bytes instead of gocql's automatic decoding
func (r *RawBytes) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	if data == nil {
		*r = nil
		return nil
	}
	*r = make([]byte, len(data))
	copy(*r, data)
	return nil
}

// MarshalCQL implements gocql.Marshaler interface
func (r RawBytes) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return []byte(r), nil
}