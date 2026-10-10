package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestADurationKeepsItsMonthsAndDays(t *testing.T) {
	for in, want := range map[string]gocql.Duration{
		"1mo2d3ns":   {Months: 1, Days: 2, Nanoseconds: 3},
		"1h30m":      {Nanoseconds: 5_400_000_000_000},
		"1y2w":       {Months: 12, Days: 14},
		"2s500ms7us": {Nanoseconds: 2_500_007_000},
		"3µs":        {Nanoseconds: 3_000},
		"-1mo2d":     {Months: -1, Days: -2},
		"0mo0d0ns":   {},
		"14MO1D":     {Months: 14, Days: 1},
	} {
		got, err := ParseCQLDuration(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, bad := range []string{"", "-", "1x", "h", "1h 2m", "1.5h", "2147483647mo1mo", "200000000y"} {
		_, err := ParseCQLDuration(bad)
		assert.Error(t, err, bad)
	}

	d := gocql.Duration{Months: 14, Days: 3, Nanoseconds: 5_400_000_000_000}
	back, err := ParseCQLDuration(FormatCQLDuration(d))
	require.NoError(t, err)
	assert.Equal(t, d, back)
}

func TestADurationIsWrittenAsCassandraWritesIt(t *testing.T) {
	for want, d := range map[string]gocql.Duration{
		"1mo2d1h30m": {Months: 1, Days: 2, Nanoseconds: 5_400_000_000_000},
		"1y2mo":      {Months: 14},
		"-3d1ns":     {Days: -3, Nanoseconds: -1},
		"2s500ms7us": {Nanoseconds: 2_500_007_000},
		"0s":         {},
	} {
		assert.Equal(t, want, FormatCQLDuration(d))
		back, err := ParseCQLDuration(want)
		require.NoError(t, err, want)
		assert.Equal(t, d, back, want)
	}
}
