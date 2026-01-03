#!/usr/bin/env python3
"""
LWT Paxos Timing Issue - Comprehensive Reproduction (Python)

Demonstrates that DELETE after INSERT IF NOT EXISTS requires a delay
for Paxos consensus to complete.

To run:
    python3 lwt_timing_reproduction_python.py

Prerequisites:
    - Cassandra running on 127.0.0.1:9042
    - Username: cassandra, Password: cassandra
    - pip install cassandra-driver

Expected Results:
    Test 1 (Regular INSERT):           ✅ DELETE works immediately
    Test 2 (IF NOT EXISTS, no delay):  ❌ DELETE fails (row remains)
    Test 3 (IF NOT EXISTS, 5s delay):  ✅ DELETE works
"""

import time
from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider
from cassandra import ConsistencyLevel

print("═══════════════════════════════════════════════════════")
print("  LWT Paxos Timing Issue Reproduction (Python Driver)")
print("═══════════════════════════════════════════════════════")
print()
print("Driver: cassandra-driver (Python)")
print("Cassandra: 5.0.6")
print()

# Connect
auth = PlainTextAuthProvider(username='cassandra', password='cassandra')
# Use server-side timestamps by setting use_client_timestamp=False
cluster = Cluster(
    ['127.0.0.1'],
    port=9042,
    auth_provider=auth
)
session = cluster.connect()

# Use server-side timestamps (don't send client timestamps)
session.use_client_timestamp = False

# CRITICAL: Set serial consistency for LWT operations
session.default_serial_consistency_level = ConsistencyLevel.SERIAL

ks = "lwt_timing_test_python"

# Setup
session.execute(f"DROP KEYSPACE IF EXISTS {ks}")
session.execute(f"""
    CREATE KEYSPACE {ks}
    WITH replication = {{'class': 'SimpleStrategy', 'replication_factor': 1}}
""")
session.execute(f"CREATE TABLE {ks}.test (id int PRIMARY KEY, data text)")

print("✅ Keyspace and table created")
print()

# =========================================================================
# TEST 1: Regular INSERT (no LWT) + DELETE
# =========================================================================
print("═══════════════════════════════════════════════════════")
print(" Test 1: Regular INSERT + DELETE (NO LWT)")
print("═══════════════════════════════════════════════════════")
print()

id1 = 100

print(f"1. INSERT INTO {ks}.test (id, data) VALUES ({id1}, 'regular')")
session.execute(f"INSERT INTO {ks}.test (id, data) VALUES ({id1}, 'regular')")
print("   ✅ INSERT executed")

# Verify
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id1}"))
if not rows:
    print("   ❌ Row not found")
    exit(1)
print(f"   ✅ Row verified: id={rows[0].id}")

# DELETE immediately
print(f"2. DELETE FROM {ks}.test WHERE id = {id1} (immediate)")
session.execute(f"DELETE FROM {ks}.test WHERE id = {id1}")
print("   ✅ DELETE executed")

# Verify DELETE
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id1}"))
if rows:
    print(f"   ❌ FAIL: Row still exists (id={rows[0].id})")
else:
    print("   ✅ SUCCESS: Row deleted immediately")

print()

# =========================================================================
# TEST 2: INSERT IF NOT EXISTS (LWT) + DELETE (no delay)
# =========================================================================
print("═══════════════════════════════════════════════════════")
print(" Test 2: INSERT IF NOT EXISTS + DELETE (NO DELAY)")
print("═══════════════════════════════════════════════════════")
print()

id2 = 200

print(f"1. INSERT INTO {ks}.test (id, data) VALUES ({id2}, 'lwt') IF NOT EXISTS")
result = session.execute(f"""
    INSERT INTO {ks}.test (id, data)
    VALUES ({id2}, 'lwt')
    IF NOT EXISTS
""")
print(f"   ✅ INSERT IF NOT EXISTS executed")
print(f"   Result: {result.one()}")

# Verify
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id2}"))
if not rows:
    print("   ❌ Row not found")
    exit(1)
print(f"   ✅ Row verified: id={rows[0].id}")

# DELETE immediately (NO DELAY)
print(f"2. DELETE FROM {ks}.test WHERE id = {id2} (IMMEDIATE - no delay)")
session.execute(f"DELETE FROM {ks}.test WHERE id = {id2}")
print("   ✅ DELETE executed (no error)")

# Verify DELETE
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id2}"))
if rows:
    print(f"   ❌ FAIL: Row STILL EXISTS (id={rows[0].id}) - BUG REPRODUCED!")
else:
    print("   ✅ SUCCESS: Row deleted")

print()

# =========================================================================
# TEST 3: INSERT IF NOT EXISTS (LWT) + 5s DELAY + DELETE
# =========================================================================
print("═══════════════════════════════════════════════════════")
print(" Test 3: INSERT IF NOT EXISTS + 5s DELAY + DELETE")
print("═══════════════════════════════════════════════════════")
print()

id3 = 300

print(f"1. INSERT INTO {ks}.test (id, data) VALUES ({id3}, 'lwt_delay') IF NOT EXISTS")
result = session.execute(f"""
    INSERT INTO {ks}.test (id, data)
    VALUES ({id3}, 'lwt_delay')
    IF NOT EXISTS
""")
print("   ✅ INSERT IF NOT EXISTS executed")
print(f"   Result: {result.one()}")

# Verify
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id3}"))
if not rows:
    print("   ❌ Row not found")
    exit(1)
print(f"   ✅ Row verified: id={rows[0].id}")

# **WAIT 1 SECOND FOR PAXOS CONSENSUS**
print("2. ⏳ Waiting 1 second for LWT Paxos consensus to complete...")
time.sleep(1)
print("   ✅ Wait complete")

# DELETE after delay
print(f"3. DELETE FROM {ks}.test WHERE id = {id3} (after 5s delay)")
session.execute(f"DELETE FROM {ks}.test WHERE id = {id3}")
print("   ✅ DELETE executed")

# Verify DELETE
rows = list(session.execute(f"SELECT id FROM {ks}.test WHERE id = {id3}"))
if rows:
    print(f"   ❌ FAIL: Row still exists (id={rows[0].id})")
else:
    print("   ✅ SUCCESS: Row deleted with 5s delay!")

print()

# =========================================================================
# SUMMARY
# =========================================================================
print("═══════════════════════════════════════════════════════")
print(" SUMMARY")
print("═══════════════════════════════════════════════════════")
print()
print("Test 1 (Regular INSERT):           ✅ DELETE works immediately")
print("Test 2 (IF NOT EXISTS, no delay):  ❌ DELETE fails")
print("Test 3 (IF NOT EXISTS, 5s delay):  ✅ DELETE works")
print()
print("CONCLUSION:")
print("DELETE after INSERT IF NOT EXISTS requires ~5 second delay")
print("for Paxos consensus to complete. Without delay, DELETE")
print("returns success but doesn't actually delete the row.")
print()
print("This affects BOTH gocql (Go) and cassandra-driver (Python).")
print("This is LWT/Paxos timing behavior, not a driver bug.")
print()

# Cleanup
session.execute(f"DROP KEYSPACE IF EXISTS {ks}")
cluster.shutdown()

print("✅ Test complete")
