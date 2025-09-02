-- Cassandra 4.1 Test Suite
-- Compatible with Cassandra 4.1.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_41
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_41;

-- Guardrails (new in 4.1)
-- Test guardrail limits (these would fail if guardrails are configured strictly)
CREATE TABLE IF NOT EXISTS guardrail_test (
    id uuid PRIMARY KEY,
    data1 text,
    data2 text,
    data3 text,
    data4 text,
    data5 text
);

-- Trie-based memtables (new in 4.1, configured at table level)
CREATE TABLE IF NOT EXISTS trie_memtable_test (
    id uuid PRIMARY KEY,
    data text,
    value int
) WITH memtable = 'trie';

INSERT INTO trie_memtable_test (id, data, value)
VALUES (uuid(), 'Test data with trie memtable', 100);

SELECT * FROM trie_memtable_test;

-- Improved secondary indexes with WHERE clause support
CREATE TABLE IF NOT EXISTS products (
    product_id uuid PRIMARY KEY,
    name text,
    category text,
    price decimal,
    in_stock boolean,
    tags set<text>
);

-- Create index with WHERE clause (new in 4.1)
CREATE INDEX IF NOT EXISTS products_in_stock_idx ON products (price) 
WHERE in_stock = true;

INSERT INTO products (product_id, name, category, price, in_stock, tags)
VALUES (uuid(), 'Laptop', 'Electronics', 999.99, true, {'computer', 'portable'});

INSERT INTO products (product_id, name, category, price, in_stock, tags)
VALUES (uuid(), 'Mouse', 'Electronics', 29.99, true, {'computer', 'accessory'});

INSERT INTO products (product_id, name, category, price, in_stock, tags)
VALUES (uuid(), 'Keyboard', 'Electronics', 79.99, false, {'computer', 'accessory'});

-- Query using the conditional index
SELECT * FROM products WHERE price < 100 AND in_stock = true;

-- Partition tombstone limits (new guardrail in 4.1)
CREATE TABLE IF NOT EXISTS tombstone_test (
    partition_key int,
    clustering_key int,
    data text,
    PRIMARY KEY (partition_key, clustering_key)
);

INSERT INTO tombstone_test (partition_key, clustering_key, data)
VALUES (1, 1, 'data1');

INSERT INTO tombstone_test (partition_key, clustering_key, data)
VALUES (1, 2, 'data2');

DELETE FROM tombstone_test WHERE partition_key = 1 AND clustering_key = 1;

-- Improved CQL native functions
CREATE TABLE IF NOT EXISTS function_test (
    id uuid PRIMARY KEY,
    text_data text,
    numeric_data int,
    timestamp_data timestamp,
    list_data list<int>
);

INSERT INTO function_test (id, text_data, numeric_data, timestamp_data, list_data)
VALUES (uuid(), 'Hello World', 42, toTimestamp(now()), [1, 2, 3, 4, 5]);

-- New/improved functions in 4.1
SELECT id,
       -- String functions
       upper(text_data) as upper_text,
       lower(text_data) as lower_text,
       -- Math functions  
       abs(-1 * numeric_data) as abs_value,
       -- Collection functions
       list_data[0] as first_element
FROM function_test;

-- Improved permissions and RBAC
-- CREATE ROLE IF NOT EXISTS app_user WITH PASSWORD = 'password' AND LOGIN = true;
-- GRANT SELECT ON KEYSPACE cqlai_test_41 TO app_user;
-- GRANT MODIFY ON TABLE cqlai_test_41.products TO app_user;

-- Concurrent schema modifications (improved in 4.1)
-- These operations are now safer to run concurrently
CREATE TABLE IF NOT EXISTS schema_test1 (
    id uuid PRIMARY KEY,
    data text
);

CREATE TABLE IF NOT EXISTS schema_test2 (
    id uuid PRIMARY KEY,
    data text
);

-- Improved streaming performance settings
-- New table with specific streaming configuration
CREATE TABLE IF NOT EXISTS streaming_test (
    id uuid PRIMARY KEY,
    large_data text,
    binary_data blob
) WITH compression = {
    'class': 'LZ4Compressor',
    'chunk_length_in_kb': 64
};

-- Insert some data for streaming
INSERT INTO streaming_test (id, large_data, binary_data)
VALUES (uuid(), 'Large text data for streaming test', textAsBlob('Binary data'));

-- Diagnostic events (new in 4.1)
-- These are available through virtual tables
SELECT * FROM system_views.local_read_latency LIMIT 5;
SELECT * FROM system_views.coordinator_read_latency LIMIT 5;

-- Improved rate limiting for operations
-- This is configured at the server level but can be tested
CREATE TABLE IF NOT EXISTS rate_limit_test (
    id uuid PRIMARY KEY,
    counter_value counter
);

-- Multiple updates to test rate limiting
UPDATE rate_limit_test SET counter_value = counter_value + 1 WHERE id = 550e8400-e29b-41d4-a716-446655440000;
UPDATE rate_limit_test SET counter_value = counter_value + 1 WHERE id = 550e8400-e29b-41d4-a716-446655440000;
UPDATE rate_limit_test SET counter_value = counter_value + 1 WHERE id = 550e8400-e29b-41d4-a716-446655440000;

SELECT * FROM rate_limit_test;

-- Improved hint support
CREATE TABLE IF NOT EXISTS hints_test (
    id uuid PRIMARY KEY,
    data text,
    last_updated timestamp
);

-- Operations that might generate hints if a node is down
INSERT INTO hints_test (id, data, last_updated)
VALUES (uuid(), 'Test hint storage', toTimestamp(now()));

-- Check hint metrics through virtual tables
SELECT * FROM system_views.pending_hints;

-- SSTable format improvements (BTI format)
CREATE TABLE IF NOT EXISTS bti_format_test (
    id uuid PRIMARY KEY,
    data text
) WITH compaction = {
    'class': 'UnifiedCompactionStrategy'
};

INSERT INTO bti_format_test (id, data)
VALUES (uuid(), 'Data in BTI format SSTable');

-- Improved garbage collection settings per table
CREATE TABLE IF NOT EXISTS gc_test (
    id uuid PRIMARY KEY,
    data text
) WITH gc_grace_seconds = 3600
  AND memtable_flush_period_in_ms = 60000;

-- Accord transactions (experimental in 4.1+)
-- BEGIN TRANSACTION
--     INSERT INTO products (product_id, name, category, price, in_stock)
--     VALUES (uuid(), 'Transactional Product', 'Electronics', 199.99, true);
--     
--     UPDATE products SET in_stock = false 
--     WHERE product_id = 550e8400-e29b-41d4-a716-446655440000;
-- COMMIT TRANSACTION;

-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_41;