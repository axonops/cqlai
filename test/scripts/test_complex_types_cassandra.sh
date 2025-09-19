#!/bin/bash
# Test script for complex types with real Cassandra

echo "=== Testing Complex Types with Cassandra ==="

# Create test keyspace and tables with complex types
cqlsh -e "
DROP KEYSPACE IF EXISTS parquet_test;
CREATE KEYSPACE parquet_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE parquet_test;

-- Create a UDT
CREATE TYPE address (
    street text,
    city text,
    zip int
);

-- Create a table with complex types
CREATE TABLE complex_types (
    id int PRIMARY KEY,
    tags list<text>,
    scores set<int>,
    metadata map<text, text>,
    addresses frozen<list<frozen<address>>>,
    location frozen<address>,
    tuple_col tuple<text, int, boolean>
);

-- Insert test data
INSERT INTO complex_types (id, tags, scores, metadata, addresses, location, tuple_col)
VALUES (1,
    ['tag1', 'tag2', 'tag3'],
    {100, 200, 300},
    {'key1': 'value1', 'key2': 'value2'},
    [{street: '123 Main St', city: 'New York', zip: 10001}, {street: '456 Oak Ave', city: 'Boston', zip: 02101}],
    {street: '789 Pine Rd', city: 'Seattle', zip: 98101},
    ('test', 42, true)
);

INSERT INTO complex_types (id, tags, scores, metadata, addresses, location, tuple_col)
VALUES (2,
    ['tag4'],
    {400},
    {'key3': 'value3'},
    [{street: '321 Elm St', city: 'Chicago', zip: 60601}],
    {street: '654 Maple Dr', city: 'Austin', zip: 78701},
    ('another', 99, false)
);

INSERT INTO complex_types (id, tags, scores, metadata)
VALUES (3,
    null,
    {500, 600},
    {'key4': 'value4', 'key5': 'value5'}
);
"

echo "Created test keyspace and inserted data with complex types"

# Export to Parquet using COPY TO
echo "Exporting to Parquet..."
./cqlai << EOF
CONNECT localhost;
USE parquet_test;
COPY complex_types TO '/tmp/complex_types_export.parquet' WITH FORMAT='PARQUET';
DESCRIBE complex_types;
SELECT * FROM complex_types;
EXIT;
EOF

echo "=== Validating Parquet file with DuckDB ==="
duckdb -c "
SELECT * FROM read_parquet('/tmp/complex_types_export.parquet');
DESCRIBE SELECT * FROM read_parquet('/tmp/complex_types_export.parquet');
"

echo "=== Testing import back to Cassandra ==="

# Create a new table for import
cqlsh -e "
USE parquet_test;
CREATE TABLE complex_types_imported (
    id int PRIMARY KEY,
    tags list<text>,
    scores set<int>,
    metadata map<text, text>,
    addresses frozen<list<frozen<address>>>,
    location frozen<address>,
    tuple_col tuple<text, int, boolean>
);
"

# Import from Parquet
echo "Importing from Parquet..."
./cqlai << EOF
CONNECT localhost;
USE parquet_test;
COPY complex_types_imported FROM '/tmp/complex_types_export.parquet' WITH FORMAT='PARQUET';
SELECT * FROM complex_types_imported;
EXIT;
EOF

echo "=== Verifying imported data in Cassandra ==="
cqlsh -e "
USE parquet_test;
SELECT * FROM complex_types_imported;
SELECT count(*) FROM complex_types_imported;

-- Compare original and imported
SELECT 'Original:' as source, * FROM complex_types WHERE id = 1
UNION ALL
SELECT 'Imported:' as source, * FROM complex_types_imported WHERE id = 1;
"

echo "=== Test complete ==="