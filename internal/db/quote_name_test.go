package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQuoteName against what Cassandra 5.0.9 takes without quotes. A name that
// starts with an underscore, and a reserved word in lower case, were left bare:
// SELECT * FROM k._staging and a column called order are syntax errors.
func TestQuoteName(t *testing.T) {
	for name, want := range map[string]string{
		"users":       "users",
		"user_id2":    "user_id2",
		"key":         "key", // a keyword, but not a reserved one
		"ttl":         "ttl",
		"type":        "type",
		"MyKeyspace":  `"MyKeyspace"`,
		"_staging":    `"_staging"`,
		"2024_events": `"2024_events"`,
		"order":       `"order"`,
		"select":      `"select"`,
		"first name":  `"first name"`,
		`say "hi"`:    `"say ""hi"""`,
	} {
		assert.Equal(t, want, QuoteName(name), name)
	}
}
