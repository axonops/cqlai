package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every position in every CREATE statement.
//
// These statements are a sequence of words in a fixed order, and completion
// offered none of it: everything after the name was silence. A function is
// eight keywords in a row that had to be remembered exactly.

// offered checks what a position offers, at the top of the engine so that the
// answer is the one a user gets.
func offered(t *testing.T, positions map[string][]string) {
	t.Helper()

	ce := NewCompletionEngine(nil, nil)
	for typed, wanted := range positions {
		got := ce.Complete(typed)
		for _, one := range wanted {
			assert.Contains(t, got, one, "after %q", typed)
		}
	}
}

// TestEveryPositionInAMaterializedView.
func TestEveryPositionInAMaterializedView(t *testing.T) {
	const start = "CREATE MATERIALIZED VIEW v AS SELECT * FROM t WHERE a IS NOT NULL"

	offered(t, map[string][]string{
		"CREATE MATERIALIZED ":                                      {"VIEW"},
		"CREATE MATERIALIZED VIEW ":                                 {Hint("view name")},
		"CREATE MATERIALIZED VIEW v ":                               {"AS"},
		"CREATE MATERIALIZED VIEW v AS ":                            {"SELECT"},
		"CREATE MATERIALIZED VIEW v AS SELECT ":                     {"*", columnHint},
		"CREATE MATERIALIZED VIEW v AS SELECT a, ":                  {columnHint},
		"CREATE MATERIALIZED VIEW v AS SELECT * ":                   {"FROM"},
		"CREATE MATERIALIZED VIEW v AS SELECT * FROM ":              {tableHint},
		"CREATE MATERIALIZED VIEW v AS SELECT * FROM t ":            {"WHERE"},
		"CREATE MATERIALIZED VIEW v AS SELECT * FROM t WHERE ":      {columnHint},
		"CREATE MATERIALIZED VIEW v AS SELECT * FROM t WHERE a ":    {"IS NOT NULL"},
		"CREATE MATERIALIZED VIEW v AS SELECT * FROM t WHERE a IS ": {"NOT NULL"},
		start + " ":              {"AND", "PRIMARY KEY"},
		start + " AND ":          {columnHint},
		start + " PRIMARY ":      {"KEY"},
		start + " PRIMARY KEY ":  {"("},
		start + " PRIMARY KEY (": {columnHint},
		// A view takes the options of a table, and its clustering order.
		start + " PRIMARY KEY (a) ":      {"WITH"},
		start + " PRIMARY KEY (a) WITH ": {"gc_grace_seconds = ", "CLUSTERING ORDER BY ("},
	})
}

// TestEveryPositionInAFunction, and the phrase in the middle of it that is
// four words long and means nothing on its own.
func TestEveryPositionInAFunction(t *testing.T) {
	const args = "CREATE FUNCTION f (a int)"

	offered(t, map[string][]string{
		"CREATE OR ":                                              {"REPLACE"},
		"CREATE OR REPLACE ":                                      {"FUNCTION", "AGGREGATE"},
		"CREATE FUNCTION ":                                        {Hint("function name")},
		"CREATE FUNCTION f ":                                      {"("},
		"CREATE FUNCTION f (":                                     {Hint("argument name")},
		"CREATE FUNCTION f (a ":                                   {"int", "text"},
		"CREATE FUNCTION f (a int, b ":                            {"int"},
		args + " ":                                                {"CALLED ON NULL INPUT", "RETURNS NULL ON NULL INPUT"},
		args + " CALLED ":                                         {"ON NULL INPUT"},
		args + " RETURNS ":                                        {"NULL ON NULL INPUT"},
		args + " CALLED ON NULL INPUT ":                           {"RETURNS"},
		args + " CALLED ON NULL INPUT RETURNS ":                   {"bigint"},
		args + " CALLED ON NULL INPUT RETURNS int ":               {"LANGUAGE"},
		args + " CALLED ON NULL INPUT RETURNS int LANGUAGE ":      {"java"},
		args + " CALLED ON NULL INPUT RETURNS int LANGUAGE java ": {"AS"},
	})
}

// TestAPhraseIsNotOfferedTwice: `CALLED` and then the whole phrase again put
// it in twice - CALLED CALLED ON NULL INPUT.
func TestAPhraseIsNotOfferedTwice(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE FUNCTION f (a int) CALLED ")
	assert.Equal(t, []string{"ON NULL INPUT"}, got)

	assert.Equal(t, []string{"INPUT"}, ce.Complete("CREATE FUNCTION f (a int) CALLED ON NULL "))
}

// TestEveryPositionInAnAggregate, whose two last parts are optional.
func TestEveryPositionInAnAggregate(t *testing.T) {
	const stype = "CREATE AGGREGATE a (int) SFUNC f STYPE int"

	offered(t, map[string][]string{
		"CREATE AGGREGATE ":                       {Hint("aggregate name")},
		"CREATE AGGREGATE a ":                     {"("},
		"CREATE AGGREGATE a (":                    {"int", "text"},
		"CREATE AGGREGATE a (int) ":               {"SFUNC"},
		"CREATE AGGREGATE a (int) SFUNC ":         {Hint("function name")},
		"CREATE AGGREGATE a (int) SFUNC f ":       {"STYPE"},
		"CREATE AGGREGATE a (int) SFUNC f STYPE ": {"int"},
		stype + " ":                               {"FINALFUNC", "INITCOND"},
		stype + " FINALFUNC g ":                   {"INITCOND"},
		stype + " INITCOND ":                      {Hint("value")},
	})

	// An aggregate that skips FINALFUNC is not asked for it again.
	ce := NewCompletionEngine(nil, nil)
	assert.NotContains(t, ce.Complete(stype+" INITCOND 0 "), "FINALFUNC")
}

// TestEveryPositionInATypeATriggerAndAKeyspace.
func TestEveryPositionInATypeATriggerAndAKeyspace(t *testing.T) {
	offered(t, map[string][]string{
		"CREATE TYPE ":                 {Hint("type name")},
		"CREATE TYPE t ":               {"("},
		"CREATE TYPE t (":              {Hint("field name")},
		"CREATE TYPE t (a ":            {"int"},
		"CREATE TYPE t (a int, ":       {Hint("field name")},
		"CREATE TYPE IF NOT EXISTS t ": {"("},

		"CREATE TRIGGER ":               {Hint("trigger name")},
		"CREATE TRIGGER tr ":            {"ON"},
		"CREATE TRIGGER tr ON ":         {tableHint},
		"CREATE TRIGGER tr ON t ":       {"USING"},
		"CREATE TRIGGER tr ON t USING ": {Hint("'trigger class'")},

		"CREATE KEYSPACE ":                                                           {Hint("keyspace name")},
		"CREATE KEYSPACE k ":                                                         {"WITH"},
		"CREATE KEYSPACE k WITH ":                                                    {"replication = ", "durable_writes = "},
		"CREATE KEYSPACE k WITH replication = ":                                      {"{'class': 'SimpleStrategy', 'replication_factor': 1}", "{'"},
		"CREATE KEYSPACE k WITH replication = {":                                     {"'class': "},
		"CREATE KEYSPACE k WITH replication = {'class': ":                            {"'SimpleStrategy'", "'NetworkTopologyStrategy'"},
		"CREATE KEYSPACE k WITH replication = {'class': 'SimpleStrategy', ":          {"'replication_factor': "},
		"CREATE KEYSPACE k WITH replication = {'class': 'NetworkTopologyStrategy', ": {Hint("'datacenter': copies")},
		"CREATE KEYSPACE k WITH durable_writes = ":                                   {"true", "false"},
		"CREATE KEYSPACE k WITH replication = {'class': 'SimpleStrategy'} ":          {"AND"},
		"CREATE KEYSPACE k WITH replication = {'class': 'SimpleStrategy'} AND ":      {"durable_writes = "},
	})

	// And the option it has been given is not offered again.
	ce := NewCompletionEngine(nil, nil)
	assert.NotContains(t, ce.Complete("CREATE KEYSPACE k WITH durable_writes = true AND "), "durable_writes = ")
}

// TestEveryPositionInARoleAndAUser.
func TestEveryPositionInARoleAndAUser(t *testing.T) {
	offered(t, map[string][]string{
		"CREATE ROLE ":                               {Hint("role name")},
		"CREATE ROLE r ":                             {"WITH"},
		"CREATE ROLE r WITH ":                        {"PASSWORD = ", "LOGIN = ", "SUPERUSER = ", "ACCESS TO ALL DATACENTERS"},
		"CREATE ROLE r WITH PASSWORD ":               {"= "},
		"CREATE ROLE r WITH PASSWORD = ":             {Hint("'password'")},
		"CREATE ROLE r WITH LOGIN = ":                {"true", "false"},
		"CREATE ROLE r WITH PASSWORD = 'x' ":         {"AND"},
		"CREATE ROLE r WITH OPTIONS = {":             {Hint("'name': 'value'")},
		"CREATE ROLE r WITH ACCESS TO DATACENTERS {": {Hint("'datacenter'")},
		"CREATE ROLE r WITH ACCESS FROM CIDRS {":     {Hint("'cidr'")},
		"ALTER ROLE r WITH ":                         {"PASSWORD = "},

		// A user is a role written the old way: its password has no equals.
		"CREATE USER ":                     {Hint("user name")},
		"CREATE USER u ":                   {"WITH", "SUPERUSER", "NOSUPERUSER"},
		"CREATE USER u WITH ":              {"PASSWORD ", "HASHED PASSWORD ", "GENERATED PASSWORD"},
		"CREATE USER u WITH PASSWORD ":     {Hint("'password'")},
		"CREATE USER u WITH PASSWORD 'x' ": {"SUPERUSER", "NOSUPERUSER"},
		"ALTER USER u WITH ":               {"PASSWORD "},
	})

	// A password is given one way or another, never two ways at once.
	ce := NewCompletionEngine(nil, nil)
	got := ce.Complete("CREATE ROLE r WITH PASSWORD = 'x' AND ")
	assert.NotContains(t, got, "PASSWORD = ")
	assert.NotContains(t, got, "HASHED PASSWORD = ")
	assert.Contains(t, got, "LOGIN = ")
}

// TestAWordInsideAStatementIsNotTheStatement.
//
// The walk finds the word that says what the CREATE makes. A comment can say
// anything at all, and a view selects FROM a table.
func TestAWordInsideAStatementIsNotTheStatement(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE TABLE t (a int PRIMARY KEY) WITH comment = 'a view of things' ")
	assert.Equal(t, []string{"AND"}, got)

	_, handled := createStatementCompletions("CREATE TABLE t (a int PRIMARY KEY)")
	assert.False(t, handled, "a table is not one of these shapes")
}

// TestTheConditionsOfAViewAreNotOptions: AND joins the options of a WITH
// clause, and also the conditions of a view's WHERE, which is not one.
func TestTheConditionsOfAViewAreNotOptions(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)

	got := ce.Complete("CREATE MATERIALIZED VIEW v AS SELECT a FROM t WHERE a IS NOT NULL AND ")

	assert.Equal(t, []string{columnHint}, got)
}
