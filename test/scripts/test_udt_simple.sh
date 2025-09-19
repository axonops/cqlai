#!/bin/bash
# Simplified test for UDT support

echo "=== Simple UDT Test ==="

# Create test keyspace and UDT
cqlsh -e "
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

SELECT * FROM users;
"

echo "=== Exporting UDT data to Parquet ==="
./cqlai << EOF
CONNECT localhost;
USE udt_test;
SELECT * FROM users;
COPY users TO '/tmp/udt_export.parquet' WITH FORMAT='PARQUET';
EXIT;
EOF

echo "=== Checking Parquet file ==="
ls -lh /tmp/udt_export.parquet

echo "=== Reading Parquet with Python ==="
python3 -c "
import pyarrow.parquet as pq
import pandas as pd

# Read the parquet file
table = pq.read_table('/tmp/udt_export.parquet')
print('Schema:', table.schema)
print('\nData:')
df = table.to_pandas()
print(df)
print('\nColumn types:')
print(df.dtypes)
" 2>/dev/null || echo "PyArrow not installed, skipping Python validation"

echo "=== Test complete ==="