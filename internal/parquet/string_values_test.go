package parquet

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValuesConvertFromTheirDisplayedForm.
//
// WriteStringRows is the path for rows that are already strings: everything
// SAVE has, since a result is [][]string by the time it can be saved, and
// CAPTURE's fallback when it has no typed rows to hand. The converters only
// took Go values, so a string going into anything but a text column failed,
// and AppendValueToBuilder appends null on a failed conversion. The file came
// out with the right number of rows and nothing in any column that was not
// text.
func TestValuesConvertFromTheirDisplayedForm(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		name  string
		typ   arrow.DataType
		given string
		want  interface{}
	}{
		{"tinyint", arrow.PrimitiveTypes.Int8, "3", int8(3)},
		{"smallint", arrow.PrimitiveTypes.Int16, "-7", int16(-7)},
		{"int", arrow.PrimitiveTypes.Int32, "42", int32(42)},
		{"bigint", arrow.PrimitiveTypes.Int64, "9000000000", int64(9000000000)},
		{"float", arrow.PrimitiveTypes.Float32, "1.5", float32(1.5)},
		{"double", arrow.PrimitiveTypes.Float64, "2.25", 2.25},
		{"boolean", arrow.FixedWidthTypes.Boolean, "true", true},
		{"text", arrow.BinaryTypes.String, "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.ConvertValue(tt.given, tt.typ)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestTimestampsConvertFromTheirDisplayedForm. FormatValue renders a time as
// RFC3339, so that is the shape a saved timestamp arrives in.
func TestTimestampsConvertFromTheirDisplayedForm(t *testing.T) {
	tm := NewTypeMapper()
	want := time.Date(2026, 9, 10, 11, 22, 33, 0, time.UTC)

	for _, given := range []string{
		"2026-09-10T11:22:33Z",
		"2026-09-10 11:22:33 +0000 UTC",
		"2026-09-10 11:22:33",
	} {
		got, err := tm.ConvertValue(given, arrow.FixedWidthTypes.Timestamp_ms)
		require.NoError(t, err, given)
		assert.Equal(t, arrow.Timestamp(want.UnixMilli()), got, given)
	}
}

// TestAnUnparseableValueIsAnError, which the builder turns into a null. One bad
// cell should cost that cell, not the rest of the row or the rest of the file.
func TestAnUnparseableValueIsAnError(t *testing.T) {
	tm := NewTypeMapper()

	for _, given := range []string{"null", "", "not a number"} {
		_, err := tm.ConvertValue(given, arrow.PrimitiveTypes.Int32)
		assert.Error(t, err, given)
	}
}
