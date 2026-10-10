package db

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// FormatCQLDuration is a duration as Cassandra writes it: 1y2mo3d4h5m6s.
// Months and days are kept apart from the time of day, as Cassandra keeps
// them, and ParseCQLDuration reads it back.
func FormatCQLDuration(d gocql.Duration) string {
	months, days, nanos := int64(d.Months), int64(d.Days), d.Nanoseconds
	var b strings.Builder
	if months < 0 || days < 0 || nanos < 0 {
		// Cassandra keeps the three of one sign.
		b.WriteByte('-')
		months, days, nanos = -months, -days, -nanos
	}
	write := func(n int64, unit string) {
		if n != 0 {
			fmt.Fprintf(&b, "%d%s", n, unit)
		}
	}
	write(months/12, "y")
	write(months%12, "mo")
	write(days, "d")
	for _, u := range []struct {
		size int64
		name string
	}{{3_600_000_000_000, "h"}, {60_000_000_000, "m"}, {1_000_000_000, "s"}, {1_000_000, "ms"}, {1_000, "us"}, {1, "ns"}} {
		write(nanos/u.size, u.name)
		nanos %= u.size
	}
	if b.Len() == 0 || b.String() == "-" {
		return "0s"
	}
	return b.String()
}

// durationUnit is one number and unit of a CQL duration such as 1h30m.
var durationUnit = regexp.MustCompile(`(?i)(\d+)(y|mo|w|d|h|ms|us|µs|ns|m|s)`)

// ParseCQLDuration reads a CQL duration in the units Cassandra takes: y, mo,
// w, d, h, m, s, ms, us (or µs) and ns, with an optional leading minus.
func ParseCQLDuration(s string) (gocql.Duration, error) {
	text := strings.TrimSpace(s)
	negative := strings.HasPrefix(text, "-")
	text = strings.TrimPrefix(text, "-")
	if text == "" {
		return gocql.Duration{}, fmt.Errorf("not a duration: %q", s)
	}

	// Added up in int64, so a sum too large for Cassandra's int32 months or
	// days is refused rather than wrapped round.
	var months, days, nanos int64
	rest := text
	for rest != "" {
		m := durationUnit.FindStringSubmatchIndex(rest)
		if m == nil || m[0] != 0 {
			return gocql.Duration{}, fmt.Errorf("not a duration: %q", s)
		}
		n, err := strconv.ParseInt(rest[m[2]:m[3]], 10, 64)
		if err != nil {
			return gocql.Duration{}, fmt.Errorf("not a duration: %q", s)
		}
		unit := strings.ToLower(rest[m[4]:m[5]])
		if (unit == "y" || unit == "mo" || unit == "w" || unit == "d") && n > math.MaxInt32 {
			return gocql.Duration{}, fmt.Errorf("duration out of range: %q", s)
		}
		switch unit {
		case "y":
			months += n * 12
		case "mo":
			months += n
		case "w":
			days += n * 7
		case "d":
			days += n
		case "h":
			nanos += n * 3_600_000_000_000
		case "m":
			nanos += n * 60_000_000_000
		case "s":
			nanos += n * 1_000_000_000
		case "ms":
			nanos += n * 1_000_000
		case "us", "µs":
			nanos += n * 1_000
		case "ns":
			nanos += n
		}
		if months > math.MaxInt32 || days > math.MaxInt32 {
			return gocql.Duration{}, fmt.Errorf("duration out of range: %q", s)
		}
		rest = rest[m[1]:]
	}
	d := gocql.Duration{Months: int32(months), Days: int32(days), Nanoseconds: nanos} // #nosec G115 - checked against MaxInt32 above
	if negative {
		d.Months, d.Days, d.Nanoseconds = -d.Months, -d.Days, -d.Nanoseconds
	}
	return d, nil
}
