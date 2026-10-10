package db

import (
	"encoding/binary"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/inf.v0"
)

// Bytes from gocql's own marshalling, so these test the decoder against what
// a driver writes, not against a reading of the protocol.

func TestADecimalInAUDTKeepsItsSignAndScale(t *testing.T) {
	info := gocql.NewNativeType(4, gocql.TypeDecimal, "")
	d := NewBinaryDecoder(nil)
	for _, text := range []string{"-0.05", "-12.345", "12.345", "0", "-1", "123456789012345678901234567890.5"} {
		want, ok := new(inf.Dec).SetString(text)
		require.True(t, ok, text)
		data, err := gocql.Marshal(info, *want)
		require.NoError(t, err, text)
		got, err := d.decodeDecimal(data)
		require.NoError(t, err, text)
		assert.Equal(t, want.String(), got, text)
	}

	// A negative scale: 1 times ten to the third.
	data, err := gocql.Marshal(info, *inf.NewDec(1, -3))
	require.NoError(t, err)
	got, err := d.decodeDecimal(data)
	require.NoError(t, err)
	assert.Equal(t, "1000", got)
}

func TestADurationInAUDTKeepsItsMonthsDaysAndSign(t *testing.T) {
	info := gocql.NewNativeType(4, gocql.TypeDuration, "")
	d := NewBinaryDecoder(nil)
	for _, want := range []gocql.Duration{
		{Months: 1, Days: 2, Nanoseconds: 3},
		{Nanoseconds: 5_400_000_000_000}, // 1h30m: a vint of several bytes
		{Months: -14, Days: -3, Nanoseconds: -1},
		{Months: 1000, Days: 300, Nanoseconds: 1 << 62},
		{},
	} {
		data, err := gocql.Marshal(info, want)
		require.NoError(t, err)
		got, err := d.decodeDuration(data)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

// TestAMapKeyThatIsASliceDoesNotCrash: a blob or inet key decodes to a
// slice, which a Go map cannot have as a key.
func TestAMapKeyThatIsASliceDoesNotCrash(t *testing.T) {
	entry := func(b []byte) []byte {
		out := binary.BigEndian.AppendUint32(nil, uint32(len(b))) // #nosec G115 - test data
		return append(out, b...)
	}
	one := binary.BigEndian.AppendUint32(nil, 1)
	d := NewBinaryDecoder(nil)

	blobMap := append(binary.BigEndian.AppendUint32(nil, 1), entry([]byte{0xca, 0xfe})...)
	blobMap = append(blobMap, entry(one)...)
	var got map[interface{}]interface{}
	require.NotPanics(t, func() {
		var err error
		got, err = d.decodeMap(blobMap, &CQLTypeInfo{BaseType: "blob"}, &CQLTypeInfo{BaseType: "int"}, "")
		require.NoError(t, err)
	})
	assert.Equal(t, map[interface{}]interface{}{KeyText("0xcafe"): int32(1)}, got)

	inetMap := append(binary.BigEndian.AppendUint32(nil, 1), entry([]byte{10, 0, 0, 1})...)
	inetMap = append(inetMap, entry(one)...)
	require.NotPanics(t, func() {
		var err error
		got, err = d.decodeMap(inetMap, &CQLTypeInfo{BaseType: "inet"}, &CQLTypeInfo{BaseType: "int"}, "")
		require.NoError(t, err)
	})
	assert.Equal(t, map[interface{}]interface{}{"10.0.0.1": int32(1)}, got)
}
