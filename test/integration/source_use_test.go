//go:build integration
// +build integration

package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAUseInASourceFileHolds: the USE was checked and forgotten, so every
// statement after it that named no keyspace failed. -f worked, because batch
// mode moved the session itself; SOURCE did not.
func TestAUseInASourceFileHolds(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()
	mgr := session.NewManager(&config.Config{})
	defer func() { _ = sess.Query(`DROP KEYSPACE IF EXISTS src_use`).Exec() }()

	file := filepath.Join(t.TempDir(), "load.cql")
	require.NoError(t, os.WriteFile(file, []byte(`DROP KEYSPACE IF EXISTS src_use;
CREATE KEYSPACE src_use WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE src_use;
CREATE TABLE talks (id int PRIMARY KEY, title text);
INSERT INTO talks (id, title) VALUES (1, 'Cassandra 5');
`), 0o600))

	result := router.ProcessCommand("SOURCE '"+file+"'", sess, mgr)
	_, failed := result.(error)
	require.False(t, failed, "%v", result)

	assert.Equal(t, "src_use", mgr.CurrentKeyspace(), "the USE holds after the file, as in cqlsh")
	var n int
	require.NoError(t, sess.Query(`SELECT COUNT(*) FROM src_use.talks`).Scan(&n))
	assert.Equal(t, 1, n)

	// The session itself is in it, not just the manager's note of it.
	_, failed = router.ProcessCommand("SELECT * FROM talks", sess, mgr).(error)
	assert.False(t, failed)
}

// TestDroppingTheCurrentKeyspaceLeavesIt: the status line went on naming a
// keyspace that was gone.
func TestDroppingTheCurrentKeyspaceLeavesIt(t *testing.T) {
	sess, _, cleanup := getTestSession(t)
	defer cleanup()
	mgr := session.NewManager(&config.Config{})

	require.NoError(t, sess.Query(`CREATE KEYSPACE IF NOT EXISTS src_drop
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec())
	router.ProcessCommand("USE src_drop", sess, mgr)
	require.Equal(t, "src_drop", mgr.CurrentKeyspace())

	// Another keyspace leaves it alone.
	require.NoError(t, sess.Query(`CREATE KEYSPACE IF NOT EXISTS src_other
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec())
	router.ProcessCommand("DROP KEYSPACE src_other", sess, mgr)
	assert.Equal(t, "src_drop", mgr.CurrentKeyspace())

	router.ProcessCommand("DROP KEYSPACE src_drop", sess, mgr)
	assert.Empty(t, mgr.CurrentKeyspace())
	_, failed := router.ProcessCommand("CREATE TABLE t (id int PRIMARY KEY)", sess, mgr).(error)
	assert.True(t, failed, "no keyspace to put it in")
}
