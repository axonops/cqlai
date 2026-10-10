package router

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"gopkg.in/inf.v0"

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

// tableColumn is one column of a table, with its type as the driver has it.
type tableColumn struct {
	name string
	info gocql.TypeInfo
}

// tableColumnsInOrder is a table's columns in the order SELECT * returns
// them, as the cluster itself describes its result: it works on every
// version, where system_schema is from 3.0.
func (h *MetaCommandHandler) tableColumnsInOrder(table string) ([]tableColumn, error) {
	if h.session == nil || h.session.Session == nil {
		return nil, fmt.Errorf("not connected")
	}
	iter := h.session.Query("SELECT * FROM " + table + " LIMIT 1").Iter()
	var columns []tableColumn
	for _, c := range iter.Columns() {
		columns = append(columns, tableColumn{name: c.Name, info: c.TypeInfo})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return columns, nil
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
func csvField(raw json.RawMessage, info gocql.TypeInfo, nullVal string) string {
	raw = bytes.TrimSpace(raw)
	switch {
	case len(raw) == 0 || string(raw) == "null":
		return nullVal
	case raw[0] == '"':
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return string(raw)
		}
		if info != nil && info.Type() == gocql.TypeTimestamp {
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

// fromJSONValues is a CSV record's values for an INSERT whose values are
// fromJson(?): each field as a JSON string, which Cassandra reads as its
// column's type, and nullVal as NULL. Only the columns named are written, so
// those the file does not have are left as they are.
func fromJSONValues(record []string, nullVal string) ([]interface{}, error) {
	values := make([]interface{}, len(record))
	for i, field := range record {
		if field == nullVal {
			values[i] = "null"
			continue
		}
		b, err := json.Marshal(field)
		if err != nil {
			return nil, err
		}
		values[i] = string(b)
	}
	return values, nil
}

// fromJSONInsert is the INSERT for fromJSONValues.
func fromJSONInsert(table string, columns []string) string {
	values := make([]string, len(columns))
	for i := range values {
		values[i] = "fromJson(?)"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(values, ", "))
}

// typedValues is a CSV record's values for a cluster before Cassandra 2.2,
// which reads no JSON: each field made the Go value of its column's type.
func typedValues(record []string, types []gocql.TypeInfo, nullVal string) ([]interface{}, error) {
	values := make([]interface{}, len(record))
	for i, field := range record {
		if field == nullVal {
			continue
		}
		var info gocql.TypeInfo
		if i < len(types) {
			info = types[i]
		}
		v, err := typedValue(field, info)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

// typedValue is a field as the Go value gocql binds to a column of type info.
func typedValue(field string, info gocql.TypeInfo) (interface{}, error) {
	if info == nil {
		return field, nil
	}
	f := strings.TrimSpace(field)
	switch info.Type() {
	case gocql.TypeInt:
		n, err := strconv.ParseInt(f, 10, 32)
		return int32(n), err
	case gocql.TypeBigInt, gocql.TypeCounter:
		return strconv.ParseInt(f, 10, 64)
	case gocql.TypeSmallInt:
		n, err := strconv.ParseInt(f, 10, 16)
		return int16(n), err
	case gocql.TypeTinyInt:
		n, err := strconv.ParseInt(f, 10, 8)
		return int8(n), err
	case gocql.TypeFloat:
		n, err := strconv.ParseFloat(f, 32)
		return float32(n), err
	case gocql.TypeDouble:
		return strconv.ParseFloat(f, 64)
	case gocql.TypeBoolean:
		return strconv.ParseBool(f)
	case gocql.TypeUUID, gocql.TypeTimeUUID:
		return gocql.ParseUUID(f)
	case gocql.TypeDecimal:
		d, ok := new(inf.Dec).SetString(f)
		if !ok {
			return nil, fmt.Errorf("not a decimal: %q", field)
		}
		return *d, nil
	case gocql.TypeVarint:
		n, ok := new(big.Int).SetString(f, 10)
		if !ok {
			return nil, fmt.Errorf("not a varint: %q", field)
		}
		return n, nil
	case gocql.TypeBlob:
		return hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(f, "0x"), "0X"))
	case gocql.TypeInet:
		ip := net.ParseIP(f)
		if ip == nil {
			return nil, fmt.Errorf("not an address: %q", field)
		}
		return ip, nil
	case gocql.TypeTimestamp:
		for _, layout := range []string{db.TimestampLayout, "2006-01-02 15:04:05.999999999-0700", "2006-01-02 15:04:05-0700",
			time.RFC3339Nano, "2006-01-02 15:04:05.999Z07:00", "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, f); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("not a timestamp: %q", field)
	}
	return field, nil
}

// csvValue is a value read by the driver as a CSV field, for a cluster
// before 2.2: as Cassandra writes it in SELECT JSON, near enough to be read
// back the same way.
func csvValue(v interface{}, nullVal string) string {
	switch t := v.(type) {
	case nil:
		return nullVal
	case string:
		return t
	case time.Time:
		return t.UTC().Format(db.TimestampLayout)
	case []byte:
		return fmt.Sprintf("0x%x", t)
	case gocql.Duration:
		return db.FormatCQLDuration(t)
	case []interface{}, map[string]interface{}, map[interface{}]interface{}:
		b, err := json.Marshal(db.JSONValue(t))
		if err == nil {
			return string(b)
		}
	}
	if j := db.JSONValue(v); j != nil {
		if kind := reflect.ValueOf(j).Kind(); kind == reflect.Slice || kind == reflect.Map {
			if b, err := json.Marshal(j); err == nil {
				return string(b)
			}
		}
		return fmt.Sprint(j)
	}
	return nullVal
}
