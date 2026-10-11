package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangesSchema(t *testing.T) {
	for command, want := range map[string]bool{
		"DROP KEYSPACE coc;":                  true,
		"drop table t":                        true,
		"CREATE TABLE t (id int PRIMARY KEY)": true,
		"ALTER TABLE t ADD c text":            true,
		"-- a comment\nCREATE KEYSPACE k WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}": true,
		"SOURCE 'load.cql'":             true,
		"source load.cql":               true,
		"SELECT * FROM t":               false,
		"INSERT INTO t (id) VALUES (1)": false,
		"TRUNCATE t":                    false,
		"USE coc":                       false,
		"":                              false,
	} {
		assert.Equal(t, want, ChangesSchema(command), command)
	}
}
