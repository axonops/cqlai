#!/bin/bash

echo "=== Minimal UDT Parquet Test ==="

# Drop and recreate keyspace
echo "DROP KEYSPACE IF EXISTS test_udt_parquet; 
CREATE KEYSPACE test_udt_parquet WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE test_udt_parquet;

-- Create a simple UDT
CREATE TYPE address (
    street text,
    city text
);

-- Create table with UDT
CREATE TABLE users (
    id int PRIMARY KEY,
    name text,
    home address
);

-- Insert test data
INSERT INTO users (id, name, home) VALUES (1, 'Alice', {street: '123 Main St', city: 'NYC'});
INSERT INTO users (id, name, home) VALUES (2, 'Bob', {street: '456 Oak Ave', city: 'LA'});
" | ./cqlai

# Export to Parquet
echo "
USE test_udt_parquet;
SELECT * FROM users;
COPY users TO '/tmp/minimal_udt.parquet';
" | ./cqlai

# Check result
echo "Checking file:"
ls -la /tmp/minimal_udt.parquet 2>/dev/null || echo "File not found"

# Try to read the parquet file if it exists
if [ -f /tmp/minimal_udt.parquet ]; then
    echo "File created successfully"
    # Could use parquet-tools here if available
fi
