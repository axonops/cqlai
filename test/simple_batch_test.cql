-- Simple test file to verify statement splitting
DROP KEYSPACE IF EXISTS test_batch;

CREATE KEYSPACE test_batch 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE test_batch;

CREATE TABLE test_table (
    id int PRIMARY KEY,
    data text
);

-- Test regular inserts
INSERT INTO test_table (id, data) VALUES (1, 'first');
INSERT INTO test_table (id, data) VALUES (2, 'second');

-- Test batch statement
BEGIN BATCH
    INSERT INTO test_table (id, data) VALUES (3, 'batch1');
    INSERT INTO test_table (id, data) VALUES (4, 'batch2');
    INSERT INTO test_table (id, data) VALUES (5, 'batch3');
APPLY BATCH;

-- Query the data
SELECT * FROM test_table;