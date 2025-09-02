-- Cassandra 4.0 Test Suite
-- Compatible with Cassandra 4.0.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_40
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_40;

-- Virtual tables (new in 4.0)
-- Query system virtual tables
SELECT * FROM system_views.clients LIMIT 5;
SELECT * FROM system_views.sstable_tasks;

-- Transient replication (new in 4.0)
-- Note: This requires NetworkTopologyStrategy and multiple nodes
-- CREATE KEYSPACE IF NOT EXISTS transient_test
-- WITH REPLICATION = {
--     'class': 'NetworkTopologyStrategy',
--     'datacenter1': '3/1'  -- 3 replicas, 1 transient
-- };

-- Improved arithmetic operations
CREATE TABLE IF NOT EXISTS calculations (
    id uuid PRIMARY KEY,
    value1 int,
    value2 int,
    value3 decimal
);

INSERT INTO calculations (id, value1, value2, value3)
VALUES (uuid(), 10, 20, 30.5);

INSERT INTO calculations (id, value1, value2, value3)
VALUES (uuid(), 15, 25, 40.75);

-- Arithmetic in SELECT (improved in 4.0)
SELECT id, 
       value1 + value2 as sum_values,
       value1 * value2 as product_values,
       value3 / 2 as half_value3,
       value1 % 3 as modulo_result
FROM calculations;

-- Allow filtering improvements
CREATE TABLE IF NOT EXISTS events (
    partition_id int,
    cluster_id int,
    event_time timestamp,
    event_type text,
    event_data text,
    PRIMARY KEY (partition_id, cluster_id, event_time)
);

INSERT INTO events (partition_id, cluster_id, event_time, event_type, event_data)
VALUES (1, 1, toTimestamp(now()), 'login', 'user1 logged in');

INSERT INTO events (partition_id, cluster_id, event_time, event_type, event_data)
VALUES (1, 2, toTimestamp(now()), 'logout', 'user1 logged out');

INSERT INTO events (partition_id, cluster_id, event_time, event_type, event_data)
VALUES (2, 1, toTimestamp(now()), 'login', 'user2 logged in');

-- Improved ALLOW FILTERING performance
SELECT * FROM events WHERE event_type = 'login' ALLOW FILTERING;

-- Improved secondary indexes
CREATE TABLE IF NOT EXISTS users (
    user_id uuid PRIMARY KEY,
    username text,
    email text,
    age int,
    country text,
    created_at timestamp
);

CREATE INDEX IF NOT EXISTS users_age_idx ON users (age);
CREATE INDEX IF NOT EXISTS users_country_idx ON users (country);

INSERT INTO users (user_id, username, email, age, country, created_at)
VALUES (uuid(), 'alice', 'alice@example.com', 30, 'USA', toTimestamp(now()));

INSERT INTO users (user_id, username, email, age, country, created_at)
VALUES (uuid(), 'bob', 'bob@example.com', 25, 'UK', toTimestamp(now()));

INSERT INTO users (user_id, username, email, age, country, created_at)
VALUES (uuid(), 'charlie', 'charlie@example.com', 35, 'USA', toTimestamp(now()));

-- Index intersection (improved in 4.0)
SELECT * FROM users WHERE age > 25 AND country = 'USA' ALLOW FILTERING;

-- Audit logging tables (when enabled)
-- SELECT * FROM system_auth.audit_log LIMIT 10;

-- New token allocation algorithm
CREATE TABLE IF NOT EXISTS distributed_data (
    id uuid PRIMARY KEY,
    data text,
    created_at timestamp
);

-- Insert data that will be distributed across tokens
INSERT INTO distributed_data (id, data, created_at)
VALUES (uuid(), 'Data 1', toTimestamp(now()));

INSERT INTO distributed_data (id, data, created_at)
VALUES (uuid(), 'Data 2', toTimestamp(now()));

INSERT INTO distributed_data (id, data, created_at)
VALUES (uuid(), 'Data 3', toTimestamp(now()));

-- Check token distribution
SELECT token(id), id FROM distributed_data;

-- Incremental repair status
-- SELECT * FROM system.repair_history LIMIT 5;

-- Improved batch operations
BEGIN BATCH
    INSERT INTO users (user_id, username, email, age, country, created_at)
    VALUES (uuid(), 'batch_user1', 'batch1@example.com', 28, 'Canada', toTimestamp(now()));
    
    INSERT INTO users (user_id, username, email, age, country, created_at)
    VALUES (uuid(), 'batch_user2', 'batch2@example.com', 32, 'Australia', toTimestamp(now()));
    
    UPDATE users SET age = 31 WHERE user_id = 550e8400-e29b-41d4-a716-446655440000;
APPLY BATCH;

-- Improved TTL handling
CREATE TABLE IF NOT EXISTS ttl_test (
    id uuid PRIMARY KEY,
    data text,
    created_at timestamp
);

-- Insert with TTL
INSERT INTO ttl_test (id, data, created_at)
VALUES (uuid(), 'Temporary data', toTimestamp(now()))
USING TTL 86400;

-- Check TTL
SELECT id, data, TTL(data) FROM ttl_test;

-- Improved compaction strategies
CREATE TABLE IF NOT EXISTS compaction_test (
    id uuid PRIMARY KEY,
    data text
) WITH compaction = {
    'class': 'SizeTieredCompactionStrategy',
    'min_threshold': '4',
    'max_threshold': '32'
};

-- Zero-copy streaming improvements (backend feature, not directly testable via CQL)

-- Improved prepared statements (backend improvement)
-- The driver and server handle prepared statements more efficiently

-- Improved messaging/internode communication (backend feature)

-- Full query logging capability (when enabled)
-- Check if FQL is enabled
-- SELECT * FROM system_views.settings WHERE name = 'full_query_logging_enabled';

-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_40;