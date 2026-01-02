# gocql Driver Bug: DELETE doesn't work after INSERT IF NOT EXISTS

## Summary

DELETE statements via the gocql driver return success (nil error) but do not actually delete rows from Cassandra when those rows were inserted using `IF NOT EXISTS`.

Manual DELETE via cqlsh works correctly on the same rows.

## Environment

- **Cassandra Version:** 5.0.6
- **gocql Version:** github.com/apache/cassandra-gocql-driver/v2 v2.0.0
- **Go Version:** 1.23
- **Cassandra Setup:** Single node, LocalOne consistency
- **Authentication:** Username/password (cassandra/cassandra)

## Reproduction

### Standalone Go Program

```go
package main

import (
	"fmt"
	"log"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func main() {
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Consistency = gocql.LocalOne

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	ks := "bug_test"

	// Setup
	session.Query("DROP KEYSPACE IF EXISTS bug_test").Exec()
	session.Query(`CREATE KEYSPACE bug_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`).Exec()
	session.Query(`CREATE TABLE bug_test.test (id int PRIMARY KEY, data text)`).Exec()

	// INSERT with IF NOT EXISTS
	err = session.Query("INSERT INTO bug_test.test (id, data) VALUES (?, ?) IF NOT EXISTS", 100, "test").Exec()
	if err != nil {
		log.Fatal("INSERT failed:", err)
	}
	fmt.Println("✅ INSERT IF NOT EXISTS succeeded")

	// Verify INSERT
	var id int
	var data string
	iter := session.Query("SELECT id, data FROM bug_test.test WHERE id = ?", 100).Iter()
	found := iter.Scan(&id, &data)
	iter.Close()
	if !found {
		log.Fatal("Row not found after INSERT")
	}
	fmt.Printf("✅ Row exists: id=%d, data='%s'\n", id, data)

	// DELETE
	err = session.Query("DELETE FROM bug_test.test WHERE id = ?", 100).Exec()
	if err != nil {
		log.Fatal("DELETE failed:", err)
	}
	fmt.Println("✅ DELETE succeeded (no error)")

	// Verify DELETE
	iter = session.Query("SELECT id FROM bug_test.test WHERE id = ?", 100).Iter()
	found = iter.Scan(&id)
	iter.Close()

	if found {
		fmt.Printf("\n❌ BUG: Row STILL EXISTS after DELETE! id=%d\n", id)
		fmt.Println("\nVerify manually via cqlsh:")
		fmt.Println("  cqlsh> SELECT * FROM bug_test.test WHERE id = 100;")
		fmt.Println("  cqlsh> DELETE FROM bug_test.test WHERE id = 100;")
		fmt.Println("  cqlsh> SELECT * FROM bug_test.test WHERE id = 100;")
	} else {
		fmt.Println("✅ DELETE worked correctly")
	}
}
```

### Steps to Reproduce

1. Run the program above
2. Observe: DELETE returns nil error
3. Observe: Row still exists when queried via gocql
4. Verify via cqlsh: Row EXISTS in Cassandra
5. Manual DELETE via cqlsh: Works immediately

## Expected Behavior

After `session.Query("DELETE FROM table WHERE id = ?", id).Exec()` returns nil error, the row should be deleted from Cassandra.

## Actual Behavior

- `Exec()` returns `nil` (no error)
- Subsequent SELECT via gocql shows row still exists
- Querying Cassandra via cqlsh confirms row still exists
- Manual DELETE via cqlsh works correctly

## Additional Testing

### Python cassandra-driver

Tested with Python's official cassandra-driver - **SAME BUG**:

```python
from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider

auth = PlainTextAuthProvider(username='cassandra', password='cassandra')
cluster = Cluster(['127.0.0.1'], auth_provider=auth)
session = cluster.connect()

session.execute("CREATE KEYSPACE test ...")
session.execute("CREATE TABLE test.t (id int PRIMARY KEY, data text)")

# INSERT with IF NOT EXISTS
session.execute("INSERT INTO test.t (id, data) VALUES (1, 'data') IF NOT EXISTS")

# DELETE
session.execute("DELETE FROM test.t WHERE id = 1")

# Verify - row STILL EXISTS
rows = session.execute("SELECT * FROM test.t WHERE id = 1")
# Row found!
```

**Both official drivers (Go and Python) have the same bug.**

## Workaround

None via driver. Manual DELETE via cqlsh works.

## Impact

- Cannot reliably delete rows that were inserted with IF NOT EXISTS
- Silent failure (no error returned)
- Critical for applications using LWT (Lightweight Transactions)

## Questions

1. Is this a known issue with LWT in Cassandra 5.0.6?
2. Is there special handling required for DELETE after IF NOT EXISTS?
3. Why does cqlsh DELETE work but driver DELETE doesn't?

## Files

- Standalone Go reproduction: `/tmp/test_gocql_delete_bug.go`
- Python reproduction: `/tmp/test_python_delete.py`
- Test with fresh session: `/tmp/test_delete_new_session.go`

All reproductions are self-contained and demonstrate the bug.
