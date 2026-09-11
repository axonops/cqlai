package db

import (
	"fmt"
	"math/big"
	"net"
	"reflect"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"gopkg.in/inf.v0"
)

// Turning a value from the driver into one encoding/json can write.
//
// Most of them go straight through. The ones that do not are the ones Cassandra
// has and JSON does not: a UUID is sixteen bytes, which json writes as an array
// of numbers; a blob is bytes, which it writes as base64 rather than the 0x
// form every other tool prints; a decimal is a big number with a scale, and a
// timestamp is a time.
//
// This is what the Results view builds JSON from. It used to ask Cassandra for
// JSON instead, by rewriting the query as SELECT JSON, which answered the
// question but changed the result: one column of documents, in place of the
// columns that were asked for.

// JSONValue converts a driver value into something json.Marshal writes
// faithfully, walking into collections and user-defined types on the way.
func JSONValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case gocql.UUID:
		return v.String()
	case *gocql.UUID:
		if v == nil {
			return nil
		}
		return v.String()
	case []byte:
		// The form cqlsh, DESCRIBE and every driver print a blob in.
		return fmt.Sprintf("0x%x", v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case time.Duration:
		return v.String()
	case gocql.Duration:
		return fmt.Sprintf("%dmo%dd%dns", v.Months, v.Days, v.Nanoseconds)
	case net.IP:
		return v.String()
	case *big.Int:
		if v == nil {
			return nil
		}
		return v.String()
	case *inf.Dec:
		if v == nil {
			return nil
		}
		return v.String()
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	}

	return jsonCollection(value)
}

// jsonCollection walks a list, set, map or tuple, converting what is inside it.
//
// A map's keys have to be strings in JSON, and Cassandra's need not be: a
// map<uuid, text> arrives with UUID keys, which json.Marshal will not write at
// all. They are converted the same way the values are and then written as text,
// which is how the key reads everywhere else.
func jsonCollection(value interface{}) interface{} {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return JSONValue(rv.Elem().Interface())

	case reflect.Slice, reflect.Array:
		items := make([]interface{}, rv.Len())
		for i := range items {
			items[i] = JSONValue(rv.Index(i).Interface())
		}
		return items

	case reflect.Map:
		object := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			name, ok := JSONValue(key.Interface()).(string)
			if !ok {
				name = fmt.Sprint(JSONValue(key.Interface()))
			}
			object[name] = JSONValue(rv.MapIndex(key).Interface())
		}
		return object
	}

	// Anything else - a UDT the driver handed back as a struct, say - is left
	// to json.Marshal, which knows what to do with a struct and writes a string
	// for whatever it does not.
	return value
}
