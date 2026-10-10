//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSourceRunsEachStatementWhole: SOURCE split its file at every
// semicolon, so a string or comment with one in it, or a BATCH, was cut up.
func TestSourceRunsEachStatementWhole(t *testing.T) {
	sess, handler, cleanup := getTestSession(t)
	defer cleanup()
	require.NoError(t, sess.Query(`DROP TABLE IF EXISTS src_t`).Exec())
	require.NoError(t, sess.Query(`CREATE TABLE src_t (id int PRIMARY KEY, s text)`).Exec())
	defer func() { _ = sess.Query(`DROP TABLE IF EXISTS src_t`).Exec() }()

	file := filepath.Join(t.TempDir(), "script.cql")
	require.NoError(t, os.WriteFile(file, []byte(`-- load; two rows
INSERT INTO src_t (id, s) VALUES (1, 'a;b');
BEGIN BATCH
  INSERT INTO src_t (id, s) VALUES (2, 'two');
  INSERT INTO src_t (id, s) VALUES (3, 'three');
APPLY BATCH;
`), 0o600))
	out := fmt.Sprint(handler.HandleMetaCommand("SOURCE '" + file + "'"))
	t.Log(out)

	var s string
	require.NoError(t, sess.Query(`SELECT s FROM src_t WHERE id = 1`).Scan(&s))
	assert.Equal(t, "a;b", s)
	var n int
	require.NoError(t, sess.Query(`SELECT COUNT(*) FROM src_t`).Scan(&n))
	assert.Equal(t, 3, n)
}

// TestASourceFileCannotSourceItself: it ran until the stack overflowed.
func TestASourceFileCannotSourceItself(t *testing.T) {
	_, handler, cleanup := getTestSession(t)
	defer cleanup()

	file := filepath.Join(t.TempDir(), "loop.cql")
	require.NoError(t, os.WriteFile(file, []byte("SOURCE '"+file+"';\n"), 0o600))
	var out interface{}
	require.NotPanics(t, func() { out = handler.HandleMetaCommand("SOURCE '" + file + "'") })
	assert.Contains(t, fmt.Sprint(out), "cannot source itself")
}
