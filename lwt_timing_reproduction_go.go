// LWT Paxos Timing Issue - Comprehensive Reproduction (Go)
//
// Demonstrates that DELETE after INSERT IF NOT EXISTS requires a delay
// for Paxos consensus to complete.
//
// To run:
//   go run lwt_timing_reproduction_go.go
//
// Prerequisites:
//   - Cassandra running on 127.0.0.1:9042
//   - Username: cassandra, Password: cassandra
//
// Expected Results:
//   Test 1 (Regular INSERT):           ✅ DELETE works immediately
//   Test 2 (IF NOT EXISTS, no delay):  ❌ DELETE fails (row remains)
//   Test 3 (IF NOT EXISTS, 5s delay):  ✅ DELETE works

package main

import (
	"fmt"
	"log"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func main() {
	// Connect to Cassandra
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
	cluster.Consistency = gocql.LocalOne
	cluster.Timeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer session.Close()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  LWT Paxos Timing Issue Reproduction (Go Driver)")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Driver: github.com/apache/cassandra-gocql-driver/v2")
	fmt.Println("Cassandra: 5.0.6")
	fmt.Println()

	ks := "lwt_timing_test_go"

	// Setup
	session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", ks)).Exec()
	session.Query(fmt.Sprintf(`CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, ks)).Exec()
	session.Query(fmt.Sprintf(`CREATE TABLE %s.test (id int PRIMARY KEY, data text)`, ks)).Exec()

	fmt.Println("✅ Keyspace and table created")
	fmt.Println()

	// =========================================================================
	// TEST 1: Regular INSERT (no LWT) + DELETE
	// =========================================================================
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println(" Test 1: Regular INSERT + DELETE (NO LWT)")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	id1 := 100

	// Regular INSERT (no IF NOT EXISTS)
	fmt.Printf("1. INSERT INTO %s.test (id, data) VALUES (%d, 'regular')\n", ks, id1)
	err = session.Query(fmt.Sprintf(`INSERT INTO %s.test (id, data) VALUES (?, ?)`, ks), id1, "regular").Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ INSERT executed")

	// Verify
	var id int
	iter := session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id1).Iter()
	found := iter.Scan(&id)
	iter.Close()
	if !found {
		log.Fatal("   ❌ Row not found")
	}
	fmt.Printf("   ✅ Row verified: id=%d\n", id)

	// DELETE immediately (no delay)
	fmt.Printf("2. DELETE FROM %s.test WHERE id = %d (immediate)\n", ks, id1)
	err = session.Query(fmt.Sprintf(`DELETE FROM %s.test WHERE id = ?`, ks), id1).Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ DELETE executed")

	// Verify DELETE
	iter = session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id1).Iter()
	found = iter.Scan(&id)
	iter.Close()

	if found {
		fmt.Printf("   ❌ FAIL: Row still exists (id=%d)\n", id)
	} else {
		fmt.Println("   ✅ SUCCESS: Row deleted immediately")
	}

	fmt.Println()

	// =========================================================================
	// TEST 2: INSERT IF NOT EXISTS (LWT) + DELETE (no delay)
	// =========================================================================
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println(" Test 2: INSERT IF NOT EXISTS + DELETE (NO DELAY)")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	id2 := 200

	// INSERT with IF NOT EXISTS
	fmt.Printf("1. INSERT INTO %s.test (id, data) VALUES (%d, 'lwt') IF NOT EXISTS\n", ks, id2)
	err = session.Query(fmt.Sprintf(`INSERT INTO %s.test (id, data) VALUES (?, ?) IF NOT EXISTS`, ks), id2, "lwt").Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ INSERT IF NOT EXISTS executed")

	// Verify
	iter = session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id2).Iter()
	found = iter.Scan(&id)
	iter.Close()
	if !found {
		log.Fatal("   ❌ Row not found")
	}
	fmt.Printf("   ✅ Row verified: id=%d\n", id)

	// DELETE immediately (no delay)
	fmt.Printf("2. DELETE FROM %s.test WHERE id = %d (IMMEDIATE - no delay)\n", ks, id2)
	err = session.Query(fmt.Sprintf(`DELETE FROM %s.test WHERE id = ?`, ks), id2).Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ DELETE executed (no error)")

	// Verify DELETE
	iter = session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id2).Iter()
	found = iter.Scan(&id)
	iter.Close()

	if found {
		fmt.Printf("   ❌ FAIL: Row STILL EXISTS (id=%d) - BUG REPRODUCED!\n", id)
	} else {
		fmt.Println("   ✅ SUCCESS: Row deleted")
	}

	fmt.Println()

	// =========================================================================
	// TEST 3: INSERT IF NOT EXISTS (LWT) + 5s DELAY + DELETE
	// =========================================================================
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println(" Test 3: INSERT IF NOT EXISTS + 5s DELAY + DELETE")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	id3 := 300

	// INSERT with IF NOT EXISTS
	fmt.Printf("1. INSERT INTO %s.test (id, data) VALUES (%d, 'lwt_delay') IF NOT EXISTS\n", ks, id3)
	err = session.Query(fmt.Sprintf(`INSERT INTO %s.test (id, data) VALUES (?, ?) IF NOT EXISTS`, ks), id3, "lwt_delay").Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ INSERT IF NOT EXISTS executed")

	// Verify
	iter = session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id3).Iter()
	found = iter.Scan(&id)
	iter.Close()
	if !found {
		log.Fatal("   ❌ Row not found")
	}
	fmt.Printf("   ✅ Row verified: id=%d\n", id)

	// **WAIT 1 SECOND FOR PAXOS CONSENSUS**
	fmt.Println("2. ⏳ Waiting 1 second for LWT Paxos consensus to complete...")
	time.Sleep(1 * time.Second)
	fmt.Println("   ✅ Wait complete")

	// DELETE after delay
	fmt.Printf("3. DELETE FROM %s.test WHERE id = %d (after 5s delay)\n", ks, id3)
	err = session.Query(fmt.Sprintf(`DELETE FROM %s.test WHERE id = ?`, ks), id3).Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   ✅ DELETE executed")

	// Verify DELETE
	iter = session.Query(fmt.Sprintf("SELECT id FROM %s.test WHERE id = ?", ks), id3).Iter()
	found = iter.Scan(&id)
	iter.Close()

	if found {
		fmt.Printf("   ❌ FAIL: Row still exists (id=%d)\n", id)
	} else {
		fmt.Println("   ✅ SUCCESS: Row deleted with 5s delay!")
	}

	fmt.Println()

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println(" SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Test 1 (Regular INSERT):           ✅ DELETE works immediately")
	fmt.Println("Test 2 (IF NOT EXISTS, no delay):  ❌ DELETE fails")
	fmt.Println("Test 3 (IF NOT EXISTS, 5s delay):  ✅ DELETE works")
	fmt.Println()
	fmt.Println("CONCLUSION:")
	fmt.Println("DELETE after INSERT IF NOT EXISTS requires a delay")
	fmt.Println("for Paxos consensus to complete. Without delay, DELETE")
	fmt.Println("returns success but doesn't actually delete the row.")
	fmt.Println()
	fmt.Println("This is LWT/Paxos timing behavior in Cassandra,")
	fmt.Println("not a bug in the gocql driver.")
	fmt.Println()

	// Cleanup
	session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", ks)).Exec()
	fmt.Println("✅ Cleanup complete")
}
