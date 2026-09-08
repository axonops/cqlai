package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewScanDestIsNotABareInterface is the guard for this bug. A *interface{}
// destination cannot represent NULL: gocql panics when the interface is empty
// and silently stores the previous row's zero value once it is not.
func TestNewScanDestIsNotABareInterface(t *testing.T) {
	types := map[string]gocql.Type{
		"text": gocql.TypeText, "int": gocql.TypeInt, "bigint": gocql.TypeBigInt,
		"boolean": gocql.TypeBoolean, "double": gocql.TypeDouble,
		"timestamp": gocql.TypeTimestamp, "uuid": gocql.TypeUUID,
		"timeuuid": gocql.TypeTimeUUID,
	}

	for name, ty := range types {
		t.Run(name, func(t *testing.T) {
			dest := NewScanDest(gocql.NewNativeType(0x04, ty, ""))

			_, isBare := dest.(*interface{})
			assert.False(t, isBare, "%s must not scan into *interface{}", name)
		})
	}
}

// TestScanValueDistinguishesNullFromZero is the behaviour that was missing: an
// unwritten column and a column holding a zero must not look the same.
func TestScanValueDistinguishesNullFromZero(t *testing.T) {
	// What gocql leaves behind for a NULL: the inner pointer stays nil.
	var nullText *string
	assert.Nil(t, ScanValue(&nullText), "a nil inner pointer is a NULL")

	// What it leaves for a real empty string: an allocated pointer to "".
	empty := ""
	notNull := &empty
	assert.Equal(t, "", ScanValue(&notNull), "an allocated pointer to \"\" is not a NULL")

	var zeroInt *int
	assert.Nil(t, ScanValue(&zeroInt))

	n := 0
	zero := &n
	assert.Equal(t, 0, ScanValue(&zero), "a real zero is not a NULL")

	var f *bool
	assert.Nil(t, ScanValue(&f))

	b := false
	falsePtr := &b
	assert.Equal(t, false, ScanValue(&falsePtr), "a real false is not a NULL")
}

func TestScanValueHandlesUDTs(t *testing.T) {
	var missing map[string]interface{}
	assert.Nil(t, ScanValue(&missing), "a nil UDT map is a NULL")

	present := map[string]interface{}{"nick": "alice"}
	assert.Equal(t, present, ScanValue(&present))
}

func TestScanValueHandlesNilAndFallback(t *testing.T) {
	assert.Nil(t, ScanValue(nil))

	// The *interface{} fallback, used when a column has no usable type info.
	var iface interface{} = "hello"
	assert.Equal(t, "hello", ScanValue(&iface))
}

func TestNewScanDestFallsBackWithoutTypeInfo(t *testing.T) {
	dest := NewScanDest(nil)
	_, ok := dest.(*interface{})
	require.True(t, ok, "no type info leaves nothing to build a typed destination from")
}
