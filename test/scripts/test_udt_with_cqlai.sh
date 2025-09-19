#!/bin/bash
# Test UDT support using only cqlai

echo "=== Testing UDT with CQLAI ==="

# Create test keyspace and UDT using cqlai
./cqlai << 'EOF'
DROP KEYSPACE IF EXISTS udt_test;
CREATE KEYSPACE udt_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE udt_test;

-- Create a simple UDT
CREATE TYPE address (
    street text,
    city text,
    zip int
);

-- Create a table with UDT
CREATE TABLE users (
    id int PRIMARY KEY,
    name text,
    home_address frozen<address>
);

-- Insert test data
INSERT INTO users (id, name, home_address)
VALUES (1, 'Alice', {street: '123 Main St', city: 'New York', zip: 10001});

INSERT INTO users (id, name, home_address)
VALUES (2, 'Bob', {street: '456 Oak Ave', city: 'Boston', zip: 02101});

-- Query to see the data
SELECT * FROM users;

-- Export to Parquet
COPY users TO '/tmp/udt_export.parquet' WITH FORMAT='PARQUET';

EXIT;
EOF

echo "=== Checking Parquet file ==="
ls -lh /tmp/udt_export.parquet

echo "=== Validating Parquet content with cqlai import ==="
./cqlai << 'EOF'
USE udt_test;

-- Create a new table for import
CREATE TABLE users_imported (
    id int PRIMARY KEY,
    name text,
    home_address frozen<address>
);

-- Import from Parquet
COPY users_imported FROM '/tmp/udt_export.parquet' WITH FORMAT='PARQUET';

-- Verify imported data
SELECT * FROM users_imported;

-- Compare original and imported
SELECT 'original' as source, * FROM users;
SELECT 'imported' as source, * FROM users_imported;

EXIT;
EOF

echo "=== Test complete ==="