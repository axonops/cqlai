package db

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"net"
	"reflect"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"gopkg.in/inf.v0"
)

// BinaryDecoder handles decoding of Cassandra binary protocol data
type BinaryDecoder struct {
	registry *UDTRegistry
}

// NewBinaryDecoder creates a new binary decoder with the given UDT registry
func NewBinaryDecoder(registry *UDTRegistry) *BinaryDecoder {
	return &BinaryDecoder{
		registry: registry,
	}
}

// Decode decodes binary data based on the type information
func (d *BinaryDecoder) Decode(data []byte, typeInfo *CQLTypeInfo, keyspace string) (interface{}, error) {
	// Handle empty data - for text/blob types, empty data is a valid empty value, not null
	// Null values in Cassandra are represented by a -1 length prefix, not empty data
	if len(data) == 0 {
		switch typeInfo.BaseType {
		case "ascii", "text", "varchar":
			return "", nil // Empty string, not null
		case "blob":
			return []byte{}, nil // Empty blob, not null
		default:
			// For other types, empty data likely means null or invalid
			return nil, nil
		}
	}

	switch typeInfo.BaseType {
	// String types
	case "ascii", "text", "varchar":
		return d.decodeText(data)

	// Integer types
	case "tinyint":
		return d.decodeTinyInt(data)
	case "smallint":
		return d.decodeSmallInt(data)
	case "int":
		return d.decodeInt(data)
	case "bigint":
		return d.decodeBigInt(data)
	case "varint":
		return d.decodeVarInt(data)
	case "counter":
		return d.decodeBigInt(data) // Counter is encoded as bigint

	// Floating point types
	case "float":
		return d.decodeFloat(data)
	case "double":
		return d.decodeDouble(data)
	case "decimal":
		return d.decodeDecimal(data)

	// Boolean type
	case "boolean":
		return d.decodeBoolean(data)

	// UUID types
	case "uuid", "timeuuid":
		return d.decodeUUID(data)

	// Time types
	case "timestamp":
		return d.decodeTimestamp(data)
	case "date":
		return d.decodeDate(data)
	case "time":
		return d.decodeTime(data)
	case "duration":
		return d.decodeDuration(data)

	// Binary type
	case "blob":
		return d.decodeBlob(data)

	// Network type
	case "inet":
		return d.decodeInet(data)

	// Collection types
	case "list", "set":
		return d.decodeList(data, typeInfo.Parameters[0], keyspace)
	case "map":
		return d.decodeMap(data, typeInfo.Parameters[0], typeInfo.Parameters[1], keyspace)
	case "tuple":
		return d.decodeTuple(data, typeInfo.Parameters, keyspace)

	// UDT type
	case "udt":
		return d.decodeUDT(data, typeInfo, keyspace)

	default:
		return nil, fmt.Errorf("unsupported type: %s", typeInfo.BaseType)
	}
}

// Primitive type decoders

func (d *BinaryDecoder) decodeText(data []byte) (string, error) {
	return string(data), nil
}

func (d *BinaryDecoder) decodeTinyInt(data []byte) (int8, error) {
	if len(data) != 1 {
		return 0, fmt.Errorf("invalid tinyint data length: %d", len(data))
	}
	return int8(data[0]), nil //nolint:gosec // G115: intentional byte-to-int8 conversion for CQL tinyint
}

func (d *BinaryDecoder) decodeSmallInt(data []byte) (int16, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid smallint data length: %d", len(data))
	}
	// Cassandra smallint is signed 16-bit, read as unsigned and cast to preserve two's complement
	return int16(binary.BigEndian.Uint16(data)), nil //nolint:gosec // CQL smallint is two's-complement; conversion is intentional
}

func (d *BinaryDecoder) decodeInt(data []byte) (int32, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid int data length: %d", len(data))
	}
	// Cassandra int is signed 32-bit, read as unsigned and cast to preserve two's complement
	return int32(binary.BigEndian.Uint32(data)), nil //nolint:gosec // CQL int is two's-complement; conversion is intentional
}

func (d *BinaryDecoder) decodeBigInt(data []byte) (int64, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid bigint data length: %d", len(data))
	}
	// Cassandra bigint is signed 64-bit, read as unsigned and cast to preserve two's complement
	return int64(binary.BigEndian.Uint64(data)), nil //nolint:gosec // CQL bigint is two's-complement; conversion is intentional
}

func (d *BinaryDecoder) decodeVarInt(data []byte) (*big.Int, error) {
	result := new(big.Int)
	result.SetBytes(data)
	// Handle sign bit for negative numbers
	if len(data) > 0 && data[0]&0x80 != 0 {
		// Negative number - perform two's complement
		bytes := make([]byte, len(data))
		copy(bytes, data)
		for i := range bytes {
			bytes[i] = ^bytes[i]
		}
		temp := new(big.Int)
		temp.SetBytes(bytes)
		result = temp.Add(temp, big.NewInt(1))
		result = result.Neg(result)
	}
	return result, nil
}

func (d *BinaryDecoder) decodeFloat(data []byte) (float32, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid float data length: %d", len(data))
	}
	bits := binary.BigEndian.Uint32(data)
	return math.Float32frombits(bits), nil
}

func (d *BinaryDecoder) decodeDouble(data []byte) (float64, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid double data length: %d", len(data))
	}
	bits := binary.BigEndian.Uint64(data)
	return math.Float64frombits(bits), nil
}

func (d *BinaryDecoder) decodeDecimal(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("invalid decimal data length: %d", len(data))
	}
	// A signed scale - a negative one is a power of ten - and a two's
	// complement unscaled value, as gocql reads a decimal column.
	scale := int32(binary.BigEndian.Uint32(data[:4])) // #nosec G115 - the scale is a signed int32 on the wire
	unscaled := new(big.Int).SetBytes(data[4:])
	if len(data) > 4 && data[4]&0x80 != 0 {
		unscaled.Sub(unscaled, new(big.Int).Lsh(big.NewInt(1), uint(8*(len(data)-4))))
	}
	return inf.NewDecBig(unscaled, inf.Scale(scale)).String(), nil
}

func (d *BinaryDecoder) decodeBoolean(data []byte) (bool, error) {
	if len(data) != 1 {
		return false, fmt.Errorf("invalid boolean data length: %d", len(data))
	}
	return data[0] != 0, nil
}

func (d *BinaryDecoder) decodeUUID(data []byte) (gocql.UUID, error) {
	if len(data) != 16 {
		return gocql.UUID{}, fmt.Errorf("invalid UUID data length: %d", len(data))
	}
	var uuid gocql.UUID
	copy(uuid[:], data)
	return uuid, nil
}

func (d *BinaryDecoder) decodeTimestamp(data []byte) (time.Time, error) {
	if len(data) != 8 {
		return time.Time{}, fmt.Errorf("invalid timestamp data length: %d", len(data))
	}
	// Cassandra timestamps are signed int64 milliseconds since epoch
	// Negative values represent dates before 1970-01-01
	// Read as signed int64 directly (two's complement)
	millis := int64(binary.BigEndian.Uint64(data)) //nolint:gosec // CQL timestamp is two's-complement int64; conversion is intentional
	return time.Unix(0, millis*int64(time.Millisecond)), nil
}

func (d *BinaryDecoder) decodeDate(data []byte) (time.Time, error) {
	if len(data) != 4 {
		return time.Time{}, fmt.Errorf("invalid date data length: %d", len(data))
	}
	// Cassandra DATE is stored as unsigned 32-bit integer
	// The value 2^31 (2147483648) represents 1970-01-01 (epoch)
	// Values < 2^31 represent dates before epoch
	// Values > 2^31 represent dates after epoch
	val := binary.BigEndian.Uint32(data)
	// Convert to days relative to epoch by subtracting 2^31
	days := int64(val) - (1 << 31)
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	return epoch.AddDate(0, 0, int(days)), nil
}

func (d *BinaryDecoder) decodeTime(data []byte) (time.Duration, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid time data length: %d", len(data))
	}
	// Cassandra TIME is nanoseconds since midnight (0 to 86399999999999)
	// Always positive, fits in int64
	nanos := int64(binary.BigEndian.Uint64(data)) //nolint:gosec // CQL time fits in int64 (0..86399999999999); conversion is intentional
	return time.Duration(nanos), nil
}

func (d *BinaryDecoder) decodeDuration(data []byte) (gocql.Duration, error) {
	// Three signed vints: months, days and nanoseconds. The same type a
	// duration column gives, so it is shown the same way.
	months, n1 := d.readVInt(data)
	days, n2 := d.readVInt(data[n1:])
	nanos, n3 := d.readVInt(data[n1+n2:])
	if n1 == 0 || n2 == 0 || n3 == 0 || n1+n2+n3 != len(data) {
		return gocql.Duration{}, fmt.Errorf("invalid duration data length: %d", len(data))
	}
	if months < math.MinInt32 || months > math.MaxInt32 || days < math.MinInt32 || days > math.MaxInt32 {
		return gocql.Duration{}, fmt.Errorf("duration out of range")
	}
	return gocql.Duration{Months: int32(months), Days: int32(days), Nanoseconds: nanos}, nil // #nosec G115 - range checked above
}

func (d *BinaryDecoder) decodeBlob(data []byte) ([]byte, error) {
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func (d *BinaryDecoder) decodeInet(data []byte) (net.IP, error) {
	if len(data) != 4 && len(data) != 16 {
		return nil, fmt.Errorf("invalid inet data length: %d", len(data))
	}
	return net.IP(data), nil
}

// Collection type decoders

func (d *BinaryDecoder) decodeList(data []byte, elementType *CQLTypeInfo, keyspace string) ([]interface{}, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid list data")
	}

	// Direct conversion to int32
	count := int32(binary.BigEndian.Uint32(data[:4])) // #nosec G115 - Safe conversion, handles -1 for null
	if count < 0 {
		return nil, fmt.Errorf("invalid collection count: %d", count)
	}
	pos := 4

	result := make([]interface{}, 0, count)

	for i := int32(0); i < count; i++ {
		if pos+4 > len(data) {
			return nil, fmt.Errorf("invalid list element at index %d", i)
		}

		// Direct conversion to int32 - negative values indicate null
		elementLen := int32(binary.BigEndian.Uint32(data[pos : pos+4])) // #nosec G115 - Safe conversion, -1 indicates null
		pos += 4

		if elementLen < 0 {
			// Null element
			result = append(result, nil)
			continue
		}

		if pos+int(elementLen) > len(data) {
			return nil, fmt.Errorf("invalid list element data at index %d", i)
		}

		elementData := data[pos : pos+int(elementLen)]
		pos += int(elementLen)

		element, err := d.Decode(elementData, elementType, keyspace)
		if err != nil {
			return nil, fmt.Errorf("failed to decode list element at index %d: %w", i, err)
		}

		result = append(result, element)
	}

	return result, nil
}

func (d *BinaryDecoder) decodeMap(data []byte, keyType, valueType *CQLTypeInfo, keyspace string) (map[interface{}]interface{}, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid map data")
	}

	// Direct conversion to int32
	count := int32(binary.BigEndian.Uint32(data[:4])) // #nosec G115 - Safe conversion, handles -1 for null
	if count < 0 {
		return nil, fmt.Errorf("invalid collection count: %d", count)
	}
	pos := 4

	result := make(map[interface{}]interface{}, count)

	for i := int32(0); i < count; i++ {
		// Decode key
		if pos+4 > len(data) {
			return nil, fmt.Errorf("invalid map key at index %d", i)
		}

		keyLenVal := binary.BigEndian.Uint32(data[pos : pos+4])
		if keyLenVal > math.MaxInt32 {
			return nil, fmt.Errorf("key length %d exceeds int32 range", keyLenVal)
		}
		keyLen := int32(keyLenVal)
		pos += 4

		if keyLen < 0 {
			return nil, fmt.Errorf("null key in map at index %d", i)
		}

		if pos+int(keyLen) > len(data) {
			return nil, fmt.Errorf("invalid map key data at index %d", i)
		}

		keyData := data[pos : pos+int(keyLen)]
		pos += int(keyLen)

		key, err := d.Decode(keyData, keyType, keyspace)
		if err != nil {
			return nil, fmt.Errorf("failed to decode map key at index %d: %w", i, err)
		}

		// Decode value
		if pos+4 > len(data) {
			return nil, fmt.Errorf("invalid map value at index %d", i)
		}

		valueLenVal := binary.BigEndian.Uint32(data[pos : pos+4])
		if valueLenVal > math.MaxInt32 {
			return nil, fmt.Errorf("value length %d exceeds int32 range", valueLenVal)
		}
		valueLen := int32(valueLenVal)
		pos += 4

		var value interface{}
		if valueLen < 0 {
			// Null value
			value = nil
		} else {
			if pos+int(valueLen) > len(data) {
				return nil, fmt.Errorf("invalid map value data at index %d", i)
			}

			valueData := data[pos : pos+int(valueLen)]
			pos += int(valueLen)

			value, err = d.Decode(valueData, valueType, keyspace)
			if err != nil {
				return nil, fmt.Errorf("failed to decode map value at index %d: %w", i, err)
			}
		}

		result[mapKey(key)] = value
	}

	return result, nil
}

// mapKey is a decoded key as a Go map can hold it. A blob, an inet or a
// frozen collection decodes to a slice or a map, which Go cannot use as a
// key - assigning one panicked - so it is held as the text it is shown as.
func mapKey(key interface{}) interface{} {
	switch k := key.(type) {
	case nil:
		return nil
	case []byte:
		return KeyText(fmt.Sprintf("0x%x", k))
	case net.IP:
		return k.String() // text, quoted as cqlsh quotes an inet
	}
	if !reflect.TypeOf(key).Comparable() {
		return KeyText(NewCQLTypeHandler().formatValueInCollection(key))
	}
	return key
}

// KeyText is a map key held as the text it is shown as, written without
// quotes: 0xcafe or [1, 2], as cqlsh writes them.
type KeyText string

func (d *BinaryDecoder) decodeTuple(data []byte, elementTypes []*CQLTypeInfo, keyspace string) ([]interface{}, error) {
	result := make([]interface{}, len(elementTypes))
	pos := 0

	for i, elementType := range elementTypes {
		if pos+4 > len(data) {
			// Not enough data for this element - rest are null
			for j := i; j < len(elementTypes); j++ {
				result[j] = nil
			}
			break
		}

		// Direct conversion to int32 - negative values indicate null
		elementLen := int32(binary.BigEndian.Uint32(data[pos : pos+4])) // #nosec G115 - Safe conversion, -1 indicates null
		pos += 4

		if elementLen < 0 {
			// Null element
			result[i] = nil
			continue
		}

		if pos+int(elementLen) > len(data) {
			return nil, fmt.Errorf("invalid tuple element data at index %d", i)
		}

		elementData := data[pos : pos+int(elementLen)]
		pos += int(elementLen)

		element, err := d.Decode(elementData, elementType, keyspace)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tuple element at index %d: %w", i, err)
		}

		result[i] = element
	}

	return result, nil
}

// UDT decoder

func (d *BinaryDecoder) decodeUDT(data []byte, typeInfo *CQLTypeInfo, keyspace string) (map[string]interface{}, error) {
	// Determine the keyspace to use
	ks := keyspace
	if typeInfo.Keyspace != "" {
		ks = typeInfo.Keyspace
	}

	// Get the UDT definition
	udtDef, err := d.registry.GetUDTDefinitionOrLoad(ks, typeInfo.UDTName)
	if err != nil {
		return nil, fmt.Errorf("failed to get UDT definition for %s.%s: %w", ks, typeInfo.UDTName, err)
	}

	result := make(map[string]interface{})
	pos := 0

	for _, field := range udtDef.Fields {
		if pos+4 > len(data) {
			// Not enough data for this field - rest are null
			result[field.Name] = nil
			continue
		}

		// Direct conversion to int32 - negative values indicate null
		fieldLen := int32(binary.BigEndian.Uint32(data[pos : pos+4])) // #nosec G115 - Safe conversion, -1 indicates null
		pos += 4

		if fieldLen < 0 {
			// Null field
			result[field.Name] = nil
			continue
		}

		if pos+int(fieldLen) > len(data) {
			return nil, fmt.Errorf("invalid UDT field data for %s", field.Name)
		}

		fieldData := data[pos : pos+int(fieldLen)]
		pos += int(fieldLen)

		fieldValue, err := d.Decode(fieldData, field.TypeInfo, ks)
		if err != nil {
			return nil, fmt.Errorf("failed to decode UDT field %s: %w", field.Name, err)
		}

		result[field.Name] = fieldValue
	}

	return result, nil
}

// Helper functions

// readVInt reads one of Cassandra's signed vints: the leading one bits of
// the first byte say how many bytes follow, and the value is zig-zag encoded,
// so a sign costs one bit. It returns the value and the bytes read, 0 when
// there are not enough.
func (d *BinaryDecoder) readVInt(data []byte) (int64, int) {
	if len(data) == 0 {
		return 0, 0
	}
	first := data[0]
	extra := bits.LeadingZeros8(^first)
	if 1+extra > len(data) {
		return 0, 0
	}
	var u uint64
	if extra < 8 {
		u = uint64(first & (0xff >> (extra + 1)))
	}
	for i := 1; i <= extra; i++ {
		u = u<<8 | uint64(data[i])
	}
	return int64(u>>1) ^ -int64(u&1), 1 + extra // #nosec G115 - zig-zag decoding
}
