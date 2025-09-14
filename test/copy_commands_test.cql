-- COPY Command Tests for CQLAI
-- Tests both COPY TO and COPY FROM functionality

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS copy_test
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE copy_test;

-- Create test table
CREATE TABLE IF NOT EXISTS test_data (
    id int PRIMARY KEY,
    name text,
    age int,
    active boolean,
    created_at timestamp,
    data blob
);

-- Insert test data
INSERT INTO test_data (id, name, age, active, created_at, data)
VALUES (1, 'Alice', 25, true, '2024-01-01 10:00:00+0000', 0x48656c6c6f);

INSERT INTO test_data (id, name, age, active, created_at, data)
VALUES (2, 'Bob', 30, false, '2024-01-02 11:00:00+0000', 0x576f726c64);

INSERT INTO test_data (id, name, age, active, created_at, data)
VALUES (3, 'Charlie', 35, true, '2024-01-03 12:00:00+0000', 0x54657374);

INSERT INTO test_data (id, name, age, active, created_at, data)
VALUES (4, 'Diana', 28, true, '2024-01-04 13:00:00+0000', null);

INSERT INTO test_data (id, name, age, active, created_at, data)
VALUES (5, 'Eve', 22, false, '2024-01-05 14:00:00+0000', 0x42696e617279);

-- Test COPY TO with various options
-- Basic export
COPY test_data TO '/tmp/test_export_basic.csv';

-- Export with header
COPY test_data TO '/tmp/test_export_header.csv' WITH HEADER=true;

-- Export specific columns
COPY test_data (id, name, age) TO '/tmp/test_export_columns.csv' WITH HEADER=true;

-- Export with custom delimiter
COPY test_data TO '/tmp/test_export_pipe.csv' WITH HEADER=true AND DELIMITER='|';

-- Export with custom null value
COPY test_data TO '/tmp/test_export_null.csv' WITH HEADER=true AND NULLVAL='N/A';

-- Verify exported data exists
SELECT COUNT(*) FROM test_data;

-- Clear table for import test
TRUNCATE test_data;

-- Test COPY FROM
-- Import basic CSV (assuming the file was created by COPY TO above)
COPY test_data FROM '/tmp/test_export_header.csv' WITH HEADER=true;

-- Verify imported data
SELECT COUNT(*) FROM test_data;
SELECT * FROM test_data WHERE id = 1;

-- Test import with limited rows
TRUNCATE test_data;
COPY test_data FROM '/tmp/test_export_header.csv' WITH HEADER=true AND MAXROWS=3;
SELECT COUNT(*) FROM test_data;  -- Should be 3

-- Test import with skipped rows
TRUNCATE test_data;
COPY test_data FROM '/tmp/test_export_header.csv' WITH HEADER=true AND SKIPROWS=1;
SELECT COUNT(*) FROM test_data;  -- Should be less than original

-- Clean up
DROP TABLE IF EXISTS test_data;
DROP KEYSPACE IF EXISTS copy_test;