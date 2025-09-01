-- ============================================================================
-- Comprehensive CQL Test Suite
-- Tests all major aspects of CQL with SimpleStrategy (RF=1)
-- ============================================================================

-- ============================================================================
-- SECTION 1: KEYSPACE OPERATIONS
-- ============================================================================

-- Drop keyspace if exists (for clean testing)
DROP KEYSPACE IF EXISTS cql_test;

-- Create keyspace with SimpleStrategy
CREATE KEYSPACE cql_test
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
}
AND durable_writes = true;

-- Use the test keyspace
USE cql_test;

-- ============================================================================
-- SECTION 2: DATA TYPES AND TABLE CREATION
-- ============================================================================

-- 2.1: Basic data types table
CREATE TABLE basic_types (
    id UUID PRIMARY KEY,
    text_col text,
    int_col int,
    bigint_col bigint,
    float_col float,
    double_col double,
    boolean_col boolean,
    timestamp_col timestamp,
    date_col date,
    time_col time,
    blob_col blob,
    inet_col inet,
    varchar_col varchar,
    varint_col varint,
    decimal_col decimal
);

-- 2.2: Collection types table
CREATE TABLE collection_types (
    id UUID PRIMARY KEY,
    list_col list<text>,
    set_col set<int>,
    map_col map<text, int>,
    frozen_list frozen<list<text>>,
    frozen_set frozen<set<int>>,
    frozen_map frozen<map<text, int>>
);

-- 2.3: Counter table
CREATE TABLE counters (
    id text PRIMARY KEY,
    counter_value counter
);

-- 2.4: Composite primary key table
CREATE TABLE composite_key (
    partition_key text,
    clustering_key int,
    value text,
    PRIMARY KEY (partition_key, clustering_key)
) WITH CLUSTERING ORDER BY (clustering_key DESC);

-- 2.5: Multiple clustering columns
CREATE TABLE multi_clustering (
    pk text,
    ck1 int,
    ck2 text,
    ck3 timestamp,
    data text,
    PRIMARY KEY (pk, ck1, ck2, ck3)
) WITH CLUSTERING ORDER BY (ck1 ASC, ck2 DESC, ck3 ASC);

-- 2.6: Static columns table
CREATE TABLE static_columns (
    partition_key text,
    clustering_key int,
    static_col text STATIC,
    regular_col text,
    PRIMARY KEY (partition_key, clustering_key)
);

-- 2.7: Table with TTL support
CREATE TABLE ttl_table (
    id UUID PRIMARY KEY,
    data text,
    created_at timestamp
);

-- 2.8: Table with custom properties
CREATE TABLE custom_properties (
    id UUID PRIMARY KEY,
    data text
) WITH 
    comment = 'Table with custom properties for testing'
    AND gc_grace_seconds = 864000
    AND default_time_to_live = 3600
    AND memtable_flush_period_in_ms = 3600000
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'class': 'LZ4Compressor'};

-- ============================================================================
-- SECTION 3: USER-DEFINED TYPES (UDT)
-- ============================================================================

-- Create user-defined types
CREATE TYPE address (
    street text,
    city text,
    state text,
    zip_code text,
    country text
);

CREATE TYPE phone (
    country_code int,
    number text,
    type text
);

-- Table using UDTs
CREATE TABLE users_with_udt (
    user_id UUID PRIMARY KEY,
    username text,
    email text,
    home_address frozen<address>,
    work_address frozen<address>,
    phones list<frozen<phone>>
);

-- ============================================================================
-- SECTION 4: INDEXES
-- ============================================================================

-- Create table for indexing examples
CREATE TABLE indexed_table (
    id UUID PRIMARY KEY,
    name text,
    age int,
    email text,
    status text,
    tags set<text>,
    attributes map<text, text>
);

-- Secondary index on regular column
CREATE INDEX idx_name ON indexed_table (name);

-- Secondary index on collection values
CREATE INDEX idx_tags ON indexed_table (VALUES(tags));

-- Secondary index on map keys
CREATE INDEX idx_attr_keys ON indexed_table (KEYS(attributes));

-- Secondary index on map entries
CREATE INDEX idx_attr_entries ON indexed_table (ENTRIES(attributes));

-- SAI indexes (if supported)
-- CREATE CUSTOM INDEX idx_sai_age ON indexed_table (age) USING 'org.apache.cassandra.index.sai.StorageAttachedIndex';

-- ============================================================================
-- SECTION 5: MATERIALIZED VIEWS
-- ============================================================================

-- Base table for materialized view
CREATE TABLE base_table (
    pk text,
    ck int,
    val1 text,
    val2 int,
    PRIMARY KEY (pk, ck)
);

-- Materialized view with different partition key
CREATE MATERIALIZED VIEW mv_by_val1 AS
    SELECT * FROM base_table
    WHERE val1 IS NOT NULL AND pk IS NOT NULL AND ck IS NOT NULL
    PRIMARY KEY (val1, pk, ck);

-- ============================================================================
-- SECTION 6: DATA MANIPULATION - INSERT
-- ============================================================================

-- 6.1: Insert into basic types
INSERT INTO basic_types (
    id, text_col, int_col, bigint_col, float_col, double_col, 
    boolean_col, timestamp_col, date_col, time_col, inet_col,
    varchar_col, varint_col, decimal_col
) VALUES (
    uuid(), 'sample text', 42, 9223372036854775807, 3.14, 2.71828,
    true, '2024-01-15T10:30:00Z', '2024-01-15', '10:30:00',
    '192.168.1.1', 'varchar value', 123456789, 123.456
);

-- 6.2: Insert with TTL
INSERT INTO ttl_table (id, data, created_at) 
VALUES (uuid(), 'This will expire', toTimestamp(now()))
USING TTL 3600;

-- 6.3: Insert with timestamp
INSERT INTO basic_types (id, text_col) 
VALUES (uuid(), 'With custom timestamp')
USING TIMESTAMP 1234567890123456;

-- 6.4: Insert collections
INSERT INTO collection_types (
    id, list_col, set_col, map_col
) VALUES (
    uuid(),
    ['item1', 'item2', 'item3'],
    {1, 2, 3, 4, 5},
    {'key1': 100, 'key2': 200, 'key3': 300}
);

-- 6.5: Insert JSON
INSERT INTO basic_types JSON '{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "text_col": "JSON insert",
    "int_col": 100,
    "boolean_col": false
}';

-- 6.6: Insert with IF NOT EXISTS
INSERT INTO basic_types (id, text_col)
VALUES (550e8400-e29b-41d4-a716-446655440001, 'Conditional insert')
IF NOT EXISTS;

-- 6.7: Batch insert into composite key table
BEGIN BATCH
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk1', 1, 'value1');
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk1', 2, 'value2');
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk1', 3, 'value3');
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk2', 1, 'value4');
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk2', 2, 'value5');
APPLY BATCH;

-- 6.8: Insert with static columns
INSERT INTO static_columns (partition_key, static_col) VALUES ('pk1', 'static_value');
INSERT INTO static_columns (partition_key, clustering_key, regular_col) VALUES ('pk1', 1, 'regular1');
INSERT INTO static_columns (partition_key, clustering_key, regular_col) VALUES ('pk1', 2, 'regular2');

-- 6.9: Insert UDT data
INSERT INTO users_with_udt (
    user_id, username, email,
    home_address,
    work_address,
    phones
) VALUES (
    uuid(), 'john_doe', 'john@example.com',
    {street: '123 Main St', city: 'New York', state: 'NY', zip_code: '10001', country: 'USA'},
    {street: '456 Work Ave', city: 'New York', state: 'NY', zip_code: '10002', country: 'USA'},
    [{country_code: 1, number: '555-0100', type: 'mobile'}, {country_code: 1, number: '555-0101', type: 'work'}]
);

-- ============================================================================
-- SECTION 7: DATA MANIPULATION - UPDATE
-- ============================================================================

-- 7.1: Basic update
UPDATE basic_types 
SET text_col = 'Updated text', int_col = 99
WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 7.2: Update with TTL
UPDATE ttl_table USING TTL 7200
SET data = 'Updated with new TTL'
WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 7.3: Counter update
UPDATE counters SET counter_value = counter_value + 1 WHERE id = 'counter1';
UPDATE counters SET counter_value = counter_value + 5 WHERE id = 'counter1';
UPDATE counters SET counter_value = counter_value - 2 WHERE id = 'counter1';

-- 7.4: Collection updates
-- List operations
UPDATE collection_types SET list_col = list_col + ['item4'] WHERE id = 550e8400-e29b-41d4-a716-446655440000;
UPDATE collection_types SET list_col[0] = 'updated_item' WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- Set operations
UPDATE collection_types SET set_col = set_col + {6, 7} WHERE id = 550e8400-e29b-41d4-a716-446655440000;
UPDATE collection_types SET set_col = set_col - {1, 2} WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- Map operations
UPDATE collection_types SET map_col['key4'] = 400 WHERE id = 550e8400-e29b-41d4-a716-446655440000;
UPDATE collection_types SET map_col = map_col + {'key5': 500} WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 7.5: Conditional update
UPDATE basic_types
SET text_col = 'Conditionally updated'
WHERE id = 550e8400-e29b-41d4-a716-446655440000
IF int_col = 99;

-- ============================================================================
-- SECTION 8: DATA MANIPULATION - DELETE
-- ============================================================================

-- 8.1: Delete entire row
DELETE FROM basic_types WHERE id = 550e8400-e29b-41d4-a716-446655440001;

-- 8.2: Delete specific columns
DELETE text_col, int_col FROM basic_types WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 8.3: Delete with timestamp
DELETE FROM ttl_table USING TIMESTAMP 1234567890123456 WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 8.4: Delete from collection
DELETE list_col[0] FROM collection_types WHERE id = 550e8400-e29b-41d4-a716-446655440000;
DELETE map_col['key1'] FROM collection_types WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 8.5: Conditional delete
DELETE FROM basic_types
WHERE id = 550e8400-e29b-41d4-a716-446655440000
IF EXISTS;

-- 8.6: Delete with clustering key range
DELETE FROM composite_key WHERE partition_key = 'pk1' AND clustering_key > 1 AND clustering_key < 3;

-- ============================================================================
-- SECTION 9: QUERIES - SELECT
-- ============================================================================

-- 9.1: Basic select
SELECT * FROM basic_types;
SELECT id, text_col, int_col FROM basic_types;

-- 9.2: Select with WHERE clause
SELECT * FROM basic_types WHERE id = 550e8400-e29b-41d4-a716-446655440000;
SELECT * FROM composite_key WHERE partition_key = 'pk1';
SELECT * FROM composite_key WHERE partition_key = 'pk1' AND clustering_key > 1;
SELECT * FROM composite_key WHERE partition_key = 'pk1' AND clustering_key >= 1 AND clustering_key <= 3;

-- 9.3: Select with ORDER BY
SELECT * FROM composite_key WHERE partition_key = 'pk1' ORDER BY clustering_key ASC;
SELECT * FROM composite_key WHERE partition_key = 'pk1' ORDER BY clustering_key DESC;

-- 9.4: Select with LIMIT
SELECT * FROM basic_types LIMIT 10;
SELECT * FROM composite_key WHERE partition_key = 'pk1' LIMIT 2;

-- 9.5: Select with ALLOW FILTERING
SELECT * FROM basic_types WHERE text_col = 'sample text' ALLOW FILTERING;
SELECT * FROM composite_key WHERE value = 'value1' ALLOW FILTERING;

-- 9.6: Select JSON
SELECT JSON * FROM basic_types;
SELECT JSON id, text_col, int_col FROM basic_types WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- 9.7: Select with functions
SELECT id, toJson(text_col) AS json_text, writetime(text_col) AS write_time, ttl(text_col) AS time_to_live FROM basic_types;
SELECT token(id) FROM basic_types;
-- dateOf() is deprecated, use toTimestamp() for modern Cassandra versions
SELECT toTimestamp(now()) AS current_timestamp FROM basic_types LIMIT 1;

-- 9.8: Select DISTINCT
SELECT DISTINCT partition_key FROM composite_key;

-- 9.9: Select with IN clause
SELECT * FROM composite_key WHERE partition_key IN ('pk1', 'pk2');
SELECT * FROM composite_key WHERE partition_key = 'pk1' AND clustering_key IN (1, 2, 3);

-- 9.10: Select static columns
SELECT static_col FROM static_columns WHERE partition_key = 'pk1';
SELECT partition_key, static_col FROM static_columns WHERE partition_key = 'pk1' LIMIT 1;

-- 9.11: Select with secondary index
SELECT * FROM indexed_table WHERE name = 'John Doe';
SELECT * FROM indexed_table WHERE tags CONTAINS 'important';
SELECT * FROM indexed_table WHERE attributes CONTAINS KEY 'color';
SELECT * FROM indexed_table WHERE attributes CONTAINS 'blue' ALLOW FILTERING;

-- 9.12: Aggregate functions
SELECT COUNT(*) FROM basic_types;
SELECT COUNT(*) FROM composite_key WHERE partition_key = 'pk1';
SELECT MIN(clustering_key), MAX(clustering_key) FROM composite_key WHERE partition_key = 'pk1';

-- 9.13: Select from materialized view
SELECT * FROM mv_by_val1 WHERE val1 = 'some_value';

-- 9.14: Token range queries (for pagination)
SELECT token(id), id FROM basic_types WHERE token(id) > -9223372036854775808 LIMIT 100;

-- ============================================================================
-- SECTION 10: BATCH OPERATIONS
-- ============================================================================

-- 10.1: Logged batch (default)
BEGIN BATCH
    INSERT INTO basic_types (id, text_col) VALUES (uuid(), 'Batch insert 1');
    INSERT INTO basic_types (id, text_col) VALUES (uuid(), 'Batch insert 2');
    UPDATE basic_types SET int_col = 100 WHERE id = 550e8400-e29b-41d4-a716-446655440000;
    DELETE FROM basic_types WHERE id = 550e8400-e29b-41d4-a716-446655440001;
APPLY BATCH;

-- 10.2: Unlogged batch
BEGIN UNLOGGED BATCH
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk3', 1, 'unlogged1');
    INSERT INTO composite_key (partition_key, clustering_key, value) VALUES ('pk3', 2, 'unlogged2');
APPLY BATCH;

-- 10.3: Counter batch
BEGIN COUNTER BATCH
    UPDATE counters SET counter_value = counter_value + 1 WHERE id = 'c1';
    UPDATE counters SET counter_value = counter_value + 2 WHERE id = 'c2';
    UPDATE counters SET counter_value = counter_value - 1 WHERE id = 'c3';
APPLY BATCH;

-- 10.4: Batch with timestamp
BEGIN BATCH USING TIMESTAMP 1234567890123456
    INSERT INTO basic_types (id, text_col) VALUES (uuid(), 'Batch with timestamp');
APPLY BATCH;

-- ============================================================================
-- SECTION 11: LIGHTWEIGHT TRANSACTIONS (LWT)
-- ============================================================================

-- 11.1: Insert if not exists
INSERT INTO basic_types (id, text_col, int_col)
VALUES (550e8400-e29b-41d4-a716-446655440010, 'LWT Insert', 1)
IF NOT EXISTS;

-- 11.2: Update if condition
UPDATE basic_types
SET text_col = 'LWT Update', int_col = 2
WHERE id = 550e8400-e29b-41d4-a716-446655440010
IF int_col = 1;

-- 11.3: Update if exists
UPDATE basic_types
SET text_col = 'Update if exists'
WHERE id = 550e8400-e29b-41d4-a716-446655440010
IF EXISTS;

-- 11.4: Delete if condition
DELETE FROM basic_types
WHERE id = 550e8400-e29b-41d4-a716-446655440010
IF text_col = 'LWT Update';

-- 11.5: Complex conditions
UPDATE basic_types
SET text_col = 'Complex LWT'
WHERE id = 550e8400-e29b-41d4-a716-446655440010
IF text_col = 'old value' AND int_col > 0 AND int_col < 100;

-- ============================================================================
-- SECTION 12: USER-DEFINED FUNCTIONS (UDF)
-- ============================================================================

-- Create a simple UDF
CREATE OR REPLACE FUNCTION multiply_int(val int, multiplier int)
RETURNS NULL ON NULL INPUT
RETURNS int
LANGUAGE java
AS 'return val * multiplier;';

-- Create a UDF for text manipulation
CREATE OR REPLACE FUNCTION reverse_text(input text)
CALLED ON NULL INPUT
RETURNS text
LANGUAGE java
AS 'return input == null ? null : new StringBuilder(input).reverse().toString();';

-- Use UDFs in queries
SELECT id, int_col, multiply_int(int_col, 2) AS doubled FROM basic_types;
SELECT id, text_col, reverse_text(text_col) AS reversed FROM basic_types;

-- ============================================================================
-- SECTION 13: USER-DEFINED AGGREGATES (UDA)
-- ============================================================================

-- Note: Java UDFs with complex types like tuples have changed significantly in Cassandra 5.0
-- For demonstration, we'll create simpler UDFs that don't use tuples

-- Create a simple sum state function
CREATE OR REPLACE FUNCTION sum_state(state bigint, val int)
RETURNS NULL ON NULL INPUT
RETURNS bigint
LANGUAGE java
AS 'return state + val;';

-- Create a count state function
CREATE OR REPLACE FUNCTION count_state(state int, val int)
RETURNS NULL ON NULL INPUT
RETURNS int
LANGUAGE java
AS 'return state + 1;';

-- Create the sum aggregate
CREATE OR REPLACE AGGREGATE sum_agg(int)
SFUNC sum_state
STYPE bigint
INITCOND 0;

-- Create the count aggregate
CREATE OR REPLACE AGGREGATE count_agg(int)
SFUNC count_state
STYPE int
INITCOND 0;

-- Use the custom aggregates
SELECT sum_agg(int_col) AS sum_val, count_agg(int_col) AS count_val FROM basic_types;

-- ============================================================================
-- SECTION 14: TRIGGERS (if supported)
-- ============================================================================

-- Note: Triggers require external implementation
-- CREATE TRIGGER audit_trigger ON basic_types USING 'com.example.AuditTrigger';

-- ============================================================================
-- SECTION 15: PERMISSIONS AND SECURITY
-- ============================================================================

-- Create roles
CREATE ROLE IF NOT EXISTS app_read WITH PASSWORD = 'read_password' AND LOGIN = true;
CREATE ROLE IF NOT EXISTS app_write WITH PASSWORD = 'write_password' AND LOGIN = true;
CREATE ROLE IF NOT EXISTS app_admin WITH PASSWORD = 'admin_password' AND LOGIN = true AND SUPERUSER = true;

-- Grant permissions
GRANT SELECT ON KEYSPACE cql_test TO app_read;
GRANT SELECT, MODIFY ON KEYSPACE cql_test TO app_write;
GRANT ALL ON KEYSPACE cql_test TO app_admin;

-- Grant specific table permissions
GRANT SELECT ON TABLE cql_test.basic_types TO app_read;
GRANT SELECT, MODIFY ON TABLE cql_test.basic_types TO app_write;

-- List permissions
LIST ALL PERMISSIONS OF app_read;
LIST ALL PERMISSIONS ON cql_test.basic_types;

-- ============================================================================
-- SECTION 16: SYSTEM QUERIES
-- ============================================================================

-- Query system tables
SELECT * FROM system.local;
SELECT * FROM system.peers;
SELECT * FROM system_schema.keyspaces WHERE keyspace_name = 'cql_test';
SELECT * FROM system_schema.tables WHERE keyspace_name = 'cql_test';
SELECT * FROM system_schema.columns WHERE keyspace_name = 'cql_test' AND table_name = 'basic_types';

-- ============================================================================
-- SECTION 17: TRACING
-- ============================================================================

-- Enable tracing for a query
TRACING ON;
SELECT * FROM basic_types LIMIT 5;
TRACING OFF;

-- Query trace information
-- SELECT * FROM system_traces.sessions;
-- SELECT * FROM system_traces.events;

-- ============================================================================
-- SECTION 18: DESCRIBE COMMANDS (Meta-commands)
-- ============================================================================

-- These are typically client-side commands
-- DESCRIBE KEYSPACES;
-- DESCRIBE KEYSPACE cql_test;
-- DESCRIBE TABLES;
-- DESCRIBE TABLE basic_types;
-- DESCRIBE TYPE address;
-- DESCRIBE INDEX idx_name;
-- DESCRIBE MATERIALIZED VIEW mv_by_val1;

-- ============================================================================
-- SECTION 19: CLEANUP (Optional - uncomment to clean up after testing)
-- ============================================================================

-- DROP AGGREGATE IF EXISTS custom_avg;
-- DROP FUNCTION IF EXISTS multiply_int;
-- DROP FUNCTION IF EXISTS reverse_text;
-- DROP FUNCTION IF EXISTS avg_state;
-- DROP FUNCTION IF EXISTS avg_final;
-- DROP MATERIALIZED VIEW IF EXISTS mv_by_val1;
-- DROP INDEX IF EXISTS idx_name;
-- DROP INDEX IF EXISTS idx_tags;
-- DROP INDEX IF EXISTS idx_attr_keys;
-- DROP INDEX IF EXISTS idx_attr_entries;
-- DROP TABLE IF EXISTS basic_types;
-- DROP TABLE IF EXISTS collection_types;
-- DROP TABLE IF EXISTS counters;
-- DROP TABLE IF EXISTS composite_key;
-- DROP TABLE IF EXISTS multi_clustering;
-- DROP TABLE IF EXISTS static_columns;
-- DROP TABLE IF EXISTS ttl_table;
-- DROP TABLE IF EXISTS custom_properties;
-- DROP TABLE IF EXISTS users_with_udt;
-- DROP TABLE IF EXISTS indexed_table;
-- DROP TABLE IF EXISTS base_table;
-- DROP TYPE IF EXISTS address;
-- DROP TYPE IF EXISTS phone;
-- DROP ROLE IF EXISTS app_read;
-- DROP ROLE IF EXISTS app_write;
-- DROP ROLE IF EXISTS app_admin;
-- DROP KEYSPACE IF EXISTS cql_test;

-- ============================================================================
-- END OF TEST SUITE
-- ============================================================================