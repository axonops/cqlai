-- Cassandra 3.0 Test Suite
-- Compatible with Cassandra 3.0.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_30
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_30;

-- Basic table with new 3.0 types
CREATE TABLE IF NOT EXISTS users (
    user_id uuid PRIMARY KEY,
    username text,
    email text,
    created_at timestamp,
    age int,
    active boolean,
    balance decimal,
    ip_address inet,
    -- New in 3.0: smallint, tinyint, date, time
    small_number smallint,
    tiny_number tinyint,
    birth_date date,
    login_time time
);

-- Insert test data with new types
INSERT INTO users (user_id, username, email, created_at, age, active, balance, 
                  ip_address, small_number, tiny_number, birth_date, login_time) 
VALUES (uuid(), 'john_doe', 'john@example.com', toTimestamp(now()), 30, true, 
        1234.56, '192.168.1.1', 100, 10, '1990-01-15', '14:30:00');

INSERT INTO users (user_id, username, email, created_at, age, active, balance,
                  ip_address, small_number, tiny_number, birth_date, login_time) 
VALUES (uuid(), 'jane_smith', 'jane@example.com', toTimestamp(now()), 25, false,
        5678.90, '10.0.0.1', 200, 20, '1995-06-20', '09:15:30');

SELECT * FROM users;

-- JSON support (new in 3.0)
CREATE TABLE IF NOT EXISTS json_data (
    id uuid PRIMARY KEY,
    data text
);

-- Insert JSON data
INSERT INTO json_data JSON '{"id": "550e8400-e29b-41d4-a716-446655440000", "data": "{\"name\": \"Test\", \"value\": 123}"}';

-- Select as JSON
SELECT JSON * FROM users LIMIT 2;

-- User-defined types (improved in 3.0)
CREATE TYPE IF NOT EXISTS address (
    street text,
    city text,
    state text,
    zip_code text
);

CREATE TABLE IF NOT EXISTS customers (
    customer_id uuid PRIMARY KEY,
    name text,
    home_address frozen<address>,
    work_address frozen<address>
);

INSERT INTO customers (customer_id, name, home_address, work_address)
VALUES (uuid(), 'Bob Smith', 
    {street: '123 Main St', city: 'NYC', state: 'NY', zip_code: '10001'},
    {street: '456 Work Ave', city: 'NYC', state: 'NY', zip_code: '10002'});

SELECT * FROM customers;

-- Materialized views (new in 3.0)
CREATE MATERIALIZED VIEW IF NOT EXISTS users_by_age AS
    SELECT age, user_id, username, email
    FROM users
    WHERE age IS NOT NULL AND user_id IS NOT NULL
    PRIMARY KEY (age, user_id);

SELECT * FROM users_by_age WHERE age >= 25;

-- Aggregates
SELECT COUNT(*) FROM users;
SELECT MAX(age) FROM users;
SELECT MIN(age) FROM users;
SELECT AVG(age) FROM users;

-- Tuple type
CREATE TABLE IF NOT EXISTS coordinates (
    id uuid PRIMARY KEY,
    location frozen<tuple<double, double>>,
    dimensions frozen<tuple<int, int, int>>
);

INSERT INTO coordinates (id, location, dimensions)
VALUES (uuid(), (40.7128, -74.0060), (10, 20, 30));

SELECT * FROM coordinates;

-- Index on collections (improved in 3.0)
CREATE TABLE IF NOT EXISTS tags_table (
    id uuid PRIMARY KEY,
    tags set<text>
);

CREATE INDEX IF NOT EXISTS tags_idx ON tags_table (tags);

INSERT INTO tags_table (id, tags) VALUES (uuid(), {'cassandra', 'database', 'nosql'});
INSERT INTO tags_table (id, tags) VALUES (uuid(), {'cassandra', 'cql'});

SELECT * FROM tags_table WHERE tags CONTAINS 'cassandra';

-- SASI indexes (new in 3.0)
CREATE CUSTOM INDEX IF NOT EXISTS users_email_sasi ON users (email) 
USING 'org.apache.cassandra.index.sasi.SASIIndex'
WITH OPTIONS = {'mode': 'CONTAINS'};

SELECT * FROM users WHERE email LIKE '%example%';

-- Per-partition limit (new in 3.0)
CREATE TABLE IF NOT EXISTS time_series (
    device_id uuid,
    reading_time timestamp,
    value double,
    PRIMARY KEY (device_id, reading_time)
) WITH CLUSTERING ORDER BY (reading_time DESC);

INSERT INTO time_series (device_id, reading_time, value) 
VALUES (550e8400-e29b-41d4-a716-446655440000, toTimestamp(now()), 23.5);

INSERT INTO time_series (device_id, reading_time, value) 
VALUES (550e8400-e29b-41d4-a716-446655440000, toTimestamp(now()), 24.5);

INSERT INTO time_series (device_id, reading_time, value) 
VALUES (550e8400-e29b-41d4-a716-446655440000, toTimestamp(now()), 25.5);

-- PER PARTITION LIMIT syntax (new in 3.0)
SELECT * FROM time_series PER PARTITION LIMIT 2;

-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_30;