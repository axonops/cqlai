package db

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
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
	// Handle null values
	if len(data) == 0 {
		return nil, nil
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
	return int8(data[0]), nil
}

func (d *BinaryDecoder) decodeSmallInt(data []byte) (int16, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid smallint data length: %d", len(data))
	}
	val := binary.BigEndian.Uint16(data)
	if val > math.MaxInt16 {
		return 0, fmt.Errorf("smallint value %d exceeds int16 range", val)
	}
	return int16(val), nil
}

func (d *BinaryDecoder) decodeInt(data []byte) (int32, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid int data length: %d", len(data))
	}
	val := binary.BigEndian.Uint32(data)
	if val > math.MaxInt32 {
		return 0, fmt.Errorf("int value %d exceeds int32 range", val)
	}
	return int32(val), nil
}

func (d *BinaryDecoder) decodeBigInt(data []byte) (int64, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid bigint data length: %d", len(data))
	}
	val := binary.BigEndian.Uint64(data)
	if val > math.MaxInt64 {
		return 0, fmt.Errorf("bigint value %d exceeds int64 range", val)
	}
	return int64(val), nil
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

	scaleVal := binary.BigEndian.Uint32(data[:4])
	if scaleVal > math.MaxInt32 {
		return "", fmt.Errorf("decimal scale %d exceeds int32 range", scaleVal)
	}
	scale := int32(scaleVal)
	unscaled := new(big.Int)
	unscaled.SetBytes(data[4:])

	// Handle sign bit
	if len(data) > 4 && data[4]&0x80 != 0 {
		// Negative number
		bytes := make([]byte, len(data)-4)
		copy(bytes, data[4:])
		for i := range bytes {
			bytes[i] = ^bytes[i]
		}
		temp := new(big.Int)
		temp.SetBytes(bytes)
		unscaled = temp.Add(temp, big.NewInt(1))
		unscaled = unscaled.Neg(unscaled)
	}

	// Format with scale
	str := unscaled.String()
	if scale > 0 {
		if len(str) <= int(scale) {
			// Need to pad with zeros
			zeros := int(scale) - len(str) + 1
			for i := 0; i < zeros; i++ {
				str = "0" + str
			}
		}
		// Insert decimal point
		pos := len(str) - int(scale)
		str = str[:pos] + "." + str[pos:]
	}

	return str, nil
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
	val := binary.BigEndian.Uint64(data)
	// Check for overflow before converting to int64
	if val > math.MaxInt64 {
		return time.Time{}, fmt.Errorf("timestamp value %d exceeds int64 range", val)
	}
	millis := int64(val)
	return time.Unix(0, millis*int64(time.Millisecond)), nil
}

func (d *BinaryDecoder) decodeDate(data []byte) (time.Time, error) {
	if len(data) != 4 {
		return time.Time{}, fmt.Errorf("invalid date data length: %d", len(data))
	}
	val := binary.BigEndian.Uint32(data)
	// Check for overflow before converting to int32
	if val > math.MaxInt32 {
		return time.Time{}, fmt.Errorf("date value %d exceeds int32 range", val)
	}
	days := int32(val)
	// Cassandra date epoch is 1970-01-01 with day precision
	// Days are stored as days since epoch (can be negative)
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	return epoch.AddDate(0, 0, int(days)), nil
}

func (d *BinaryDecoder) decodeTime(data []byte) (time.Duration, error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid time data length: %d", len(data))
	}
	val := binary.BigEndian.Uint64(data)
	// Check for overflow before converting to int64
	if val > math.MaxInt64 {
		return 0, fmt.Errorf("time value %d exceeds int64 range", val)
	}
	nanos := int64(val)
	return time.Duration(nanos), nil
}

func (d *BinaryDecoder) decodeDuration(data []byte) (map[string]interface{}, error) {
	// Duration is encoded as vint months, vint days, vint nanoseconds
	pos := 0

	months, bytesRead := d.readVInt(data[pos:])
	pos += bytesRead

	days, bytesRead := d.readVInt(data[pos:])
	pos += bytesRead

	nanos, _ := d.readVInt(data[pos:])

	return map[string]interface{}{
		"months": months,
		"days":   days,
		"nanos":  nanos,
	}, nil
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

		// Convert key to string for string keys (common case)
		if keyStr, ok := key.(string); ok {
			result[keyStr] = value
		} else {
			result[key] = value
		}
	}

	return result, nil
}

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

func (d *BinaryDecoder) readVInt(data []byte) (int64, int) {
	if len(data) == 0 {
		return 0, 0
	}

	firstByte := data[0]

	// Check if it's a single byte vint
	if firstByte&0x80 == 0 {
		// Positive single byte
		return int64(firstByte), 1
	}

	// Multi-byte vint
	// Count leading ones to determine length
	length := 0
	for i := 7; i >= 0; i-- {
		if firstByte&(1<<uint(i)) == 0 {
			break
		}
		length++
	}

	if length == 0 || length > 8 || length > len(data) {
		return 0, 0
	}

	// Extract value
	// Safely calculate the shift amount
	if length > 8 || length < 1 {
		return 0, 0
	}
	shiftAmount := 8 - length
	result := int64(firstByte & ((1 << shiftAmount) - 1))
	for i := 1; i < length; i++ {
		result = (result << 8) | int64(data[i])
	}

	// Handle negative numbers
	if length == 8 {
		// Special case for maximum negative value
		result = -result
	}

	return result, length
}