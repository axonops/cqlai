package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every position in the statements that read and write rows.
//
// What these had was a list of the keywords that can appear anywhere in the
// statement, offered at every position: after a column in a WHERE clause it
// offered AND, GROUP, ORDER, LIMIT and ALLOW, where what goes there is an
// operator, and it offered WHERE after LIMIT, where nothing goes at all.

// TestEveryPositionInAQuery.
func TestEveryPositionInAQuery(t *testing.T) {
	offered(t, map[string][]string{
		"SELECT a, ":       {columnHint},
		"SELECT a ":        {"FROM"},
		"SELECT * FROM ":   {tableHint},
		"SELECT * FROM t ": {"WHERE", "GROUP BY", "ORDER BY", "PER PARTITION LIMIT", "LIMIT", "ALLOW FILTERING"},

		"SELECT * FROM t WHERE ":             {columnHint},
		"SELECT * FROM t WHERE a ":           {"=", "!=", "IN", "CONTAINS KEY", "BETWEEN", "LIKE", "IS NOT NULL"},
		"SELECT * FROM t WHERE a = 1 ":       {"AND", "LIMIT"},
		"SELECT * FROM t WHERE a = 1 AND ":   {columnHint},
		"SELECT * FROM t WHERE a IN ":        {valueHint},
		"SELECT * FROM t WHERE a CONTAINS ":  {"KEY", valueHint},
		"SELECT * FROM t WHERE a BETWEEN ":   {valueHint},
		"SELECT * FROM t WHERE a BETWEEN 1 ": {"AND"},
		"SELECT * FROM t WHERE a IS ":        {"NOT NULL"},
		"SELECT * FROM t WHERE TOKEN(":       {columnHint},

		"SELECT * FROM t GROUP ":               {"BY"},
		"SELECT * FROM t GROUP BY ":            {columnHint},
		"SELECT * FROM t GROUP BY a ":          {"ORDER BY", "LIMIT"},
		"SELECT * FROM t ORDER ":               {"BY"},
		"SELECT * FROM t ORDER BY ":            {columnHint},
		"SELECT * FROM t ORDER BY a ":          {"ASC", "DESC", "ANN OF"},
		"SELECT * FROM t ORDER BY a ANN OF ":   {valueHint},
		"SELECT * FROM t PER ":                 {"PARTITION LIMIT"},
		"SELECT * FROM t PER PARTITION LIMIT ": {Hint("rows")},
		"SELECT * FROM t LIMIT ":               {Hint("rows")},
		"SELECT * FROM t ALLOW ":               {"FILTERING"},
	})
}

// TestAQueryThatIsFinishedIsFinished: it offered WHERE after LIMIT, and LIMIT
// after ALLOW FILTERING.
func TestAQueryThatIsFinishedIsFinished(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{"ALLOW FILTERING"}, ce.Complete("SELECT * FROM t LIMIT 10 "))
	assert.Empty(t, ce.Complete("SELECT * FROM t ALLOW FILTERING "))
}

// TestEveryPositionInAnInsert, given as columns and values or as one document.
func TestEveryPositionInAnInsert(t *testing.T) {
	offered(t, map[string][]string{
		"INSERT ":                                    {"INTO"},
		"INSERT INTO ":                               {tableHint},
		"INSERT INTO t ":                             {"("},
		"INSERT INTO t (":                            {columnHint},
		"INSERT INTO t (a, ":                         {columnHint},
		"INSERT INTO t (a) ":                         {"VALUES"},
		"INSERT INTO t (a) VALUES (":                 {valueHint},
		"INSERT INTO t (a) VALUES (1) ":              {"IF NOT EXISTS", "USING"},
		"INSERT INTO t (a) VALUES (1) IF ":           {"NOT EXISTS"},
		"INSERT INTO t (a) VALUES (1) USING ":        {"TTL", "TIMESTAMP"},
		"INSERT INTO t (a) VALUES (1) USING TTL ":    {Hint("seconds")},
		"INSERT INTO t (a) VALUES (1) USING TTL 60 ": {"AND"},

		"INSERT INTO t JSON ":              {Hint("'json'")},
		"INSERT INTO t JSON '{}' ":         {"DEFAULT", "IF NOT EXISTS", "USING"},
		"INSERT INTO t JSON '{}' DEFAULT ": {"NULL", "UNSET"},
	})
}

// TestEveryPositionInAnUpdate.
func TestEveryPositionInAnUpdate(t *testing.T) {
	offered(t, map[string][]string{
		"UPDATE ":                            {tableHint},
		"UPDATE t ":                          {"USING", "SET"},
		"UPDATE t USING ":                    {"TTL", "TIMESTAMP"},
		"UPDATE t USING TTL 60 ":             {"AND", "SET"},
		"UPDATE t SET ":                      {columnHint},
		"UPDATE t SET a ":                    {"= ", "+= ", "-= "},
		"UPDATE t SET a = ":                  {valueHint},
		"UPDATE t SET a = 1 ":                {"WHERE"},
		"UPDATE t SET a = 1 WHERE ":          {columnHint},
		"UPDATE t SET a = 1 WHERE b = 2 ":    {"AND", "IF"},
		"UPDATE t SET a = 1 WHERE b = 2 IF ": {"EXISTS", columnHint},
	})

	// A write conditional on the row existing takes nothing after it.
	ce := NewCompletionEngine(nil, nil)
	assert.Empty(t, ce.Complete("UPDATE t SET a = 1 WHERE b = 2 IF EXISTS "))
}

// TestEveryPositionInADelete, which can take the columns to remove or leave
// them out to remove the row.
func TestEveryPositionInADelete(t *testing.T) {
	offered(t, map[string][]string{
		"DELETE ":                        {"FROM", columnHint},
		"DELETE a ":                      {"FROM"},
		"DELETE a, ":                     {columnHint},
		"DELETE FROM ":                   {tableHint},
		"DELETE FROM t ":                 {"USING TIMESTAMP", "WHERE"},
		"DELETE FROM t USING ":           {"TIMESTAMP"},
		"DELETE FROM t USING TIMESTAMP ": {Hint("microseconds")},
		"DELETE FROM t WHERE ":           {columnHint},
		"DELETE FROM t WHERE a = 1 ":     {"AND", "IF"},
		"DELETE FROM t WHERE a = 1 IF ":  {"EXISTS", columnHint},
	})
}

// TestEveryPositionInABatch.
//
// A batch holds whole statements, and each of them completes the way it does
// on its own: what comes after the last semicolon is a statement being typed.
func TestEveryPositionInABatch(t *testing.T) {
	offered(t, map[string][]string{
		"BEGIN ":             {"BATCH", "UNLOGGED", "COUNTER"},
		"BEGIN UNLOGGED ":    {"BATCH"},
		"BEGIN BATCH ":       {"USING TIMESTAMP", "INSERT", "UPDATE", "DELETE", "APPLY BATCH"},
		"BEGIN BATCH USING ": {"TIMESTAMP"},
		"BEGIN BATCH INSERT INTO t (a) VALUES (1); ":                          {"INSERT", "UPDATE", "DELETE", "APPLY BATCH"},
		"BEGIN BATCH INSERT INTO t (a) VALUES (1); UPDATE t ":                 {"SET"},
		"BEGIN BATCH INSERT INTO t (a) VALUES (1); UPDATE t SET a = 1 WHERE ": {columnHint},
		"APPLY ": {"BATCH"},
	})
}

// TestAKeywordHalfTypedIsFinishedRatherThanRepeated.
func TestAKeywordHalfTypedIsFinishedRatherThanRepeated(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	assert.Equal(t, []string{"PARTITION LIMIT"}, ce.Complete("SELECT * FROM t PER "))
	assert.Equal(t, []string{"FILTERING"}, ce.Complete("SELECT * FROM t ALLOW "))
	assert.Equal(t, []string{"NOT EXISTS"}, ce.Complete("INSERT INTO t (a) VALUES (1) IF "))
	assert.Equal(t, []string{"TIMESTAMP"}, ce.Complete("DELETE FROM t USING "))
}
