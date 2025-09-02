-- Cassandra 2.1 Test Suite
-- Compatible with Cassandra 2.1.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_21
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_21;

-- Basic table with simple types (2.1 compatible)
CREATE TABLE IF NOT EXISTS users (
    user_id uuid PRIMARY KEY,
    username text,
    email text,
    created_at timestamp,
    age int,
    active boolean
);

-- Insert test data
INSERT INTO users (user_id, username, email, created_at, age, active) 
VALUES (uuid(), 'john_doe', 'john@example.com', dateOf(now()), 30, true);

INSERT INTO users (user_id, username, email, created_at, age, active) 
VALUES (uuid(), 'jane_smith', 'jane@example.com', dateOf(now()), 25, false);

-- Basic SELECT queries
SELECT * FROM users;
SELECT username, email FROM users WHERE active = true ALLOW FILTERING;
SELECT COUNT(*) FROM users;

-- Table with compound primary key
CREATE TABLE IF NOT EXISTS events (
    event_id uuid,
    event_date timestamp,
    event_type text,
    event_data text,
    PRIMARY KEY (event_id, event_date)
) WITH CLUSTERING ORDER BY (event_date DESC);

-- Insert events
INSERT INTO events (event_id, event_date, event_type, event_data) 
VALUES (uuid(), dateOf(now()), 'login', 'User logged in');

INSERT INTO events (event_id, event_date, event_type, event_data) 
VALUES (uuid(), dateOf(now()), 'logout', 'User logged out');

-- Collections (list, set, map) - available in 2.1
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id uuid PRIMARY KEY,
    emails set<text>,
    phone_numbers list<text>,
    properties map<text, text>
);

INSERT INTO user_profiles (user_id, emails, phone_numbers, properties)
VALUES (uuid(), 
    {'email1@test.com', 'email2@test.com'},
    ['555-0100', '555-0101'],
    {'location': 'NYC', 'department': 'Engineering'}
);

SELECT * FROM user_profiles;

-- Counter table
CREATE TABLE IF NOT EXISTS page_views (
    page_id uuid,
    view_date date,
    views counter,
    PRIMARY KEY (page_id, view_date)
);

UPDATE page_views SET views = views + 1 WHERE page_id = uuid() AND view_date = '2024-01-01';
UPDATE page_views SET views = views + 5 WHERE page_id = uuid() AND view_date = '2024-01-01';

SELECT * FROM page_views;

-- TTL test
INSERT INTO users (user_id, username, email, created_at, age, active) 
VALUES (uuid(), 'temp_user', 'temp@example.com', dateOf(now()), 20, true)
USING TTL 3600;


-- Batch operations
BEGIN BATCH
    INSERT INTO users (user_id, username, email, created_at, age, active) 
    VALUES (uuid(), 'batch_user1', 'batch1@example.com', dateOf(now()), 35, true);
    INSERT INTO users (user_id, username, email, created_at, age, active) 
    VALUES (uuid(), 'batch_user2', 'batch2@example.com', dateOf(now()), 40, false);
APPLY BATCH;

-- Static columns (available in 2.1)
CREATE TABLE IF NOT EXISTS sensor_data (
    sensor_id uuid,
    sensor_location text STATIC,
    reading_time timestamp,
    temperature double,
    humidity double,
    PRIMARY KEY (sensor_id, reading_time)
);

INSERT INTO sensor_data (sensor_id, sensor_location, reading_time, temperature, humidity)
VALUES (uuid(), 'Building A - Floor 1', dateOf(now()), 22.5, 45.0);

-- Lightweight transactions (available in 2.1)
INSERT INTO users (user_id, username, email, created_at, age, active) 
VALUES (uuid(), 'unique_user', 'unique@example.com', dateOf(now()), 28, true)
IF NOT EXISTS;

-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_21;