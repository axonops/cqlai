-- Cassandra 3.11 Test Suite
-- Compatible with Cassandra 3.11.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_311
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_311;

-- Duration type (new in 3.11)
CREATE TABLE IF NOT EXISTS durations_test (
    id uuid PRIMARY KEY,
    name text,
    time_spent duration
);

INSERT INTO durations_test (id, name, time_spent) 
VALUES (uuid(), 'Task 1', 2h30m);

INSERT INTO durations_test (id, name, time_spent) 
VALUES (uuid(), 'Task 2', 1d12h);

INSERT INTO durations_test (id, name, time_spent) 
VALUES (uuid(), 'Task 3', 3mo2d);

SELECT * FROM durations_test;

-- Group by support (improved in 3.11)
CREATE TABLE IF NOT EXISTS sales (
    store_id int,
    sale_date date,
    product_id uuid,
    quantity int,
    price decimal,
    PRIMARY KEY ((store_id), sale_date, product_id)
);

INSERT INTO sales (store_id, sale_date, product_id, quantity, price)
VALUES (1, '2024-01-01', uuid(), 5, 29.99);

INSERT INTO sales (store_id, sale_date, product_id, quantity, price)
VALUES (1, '2024-01-01', uuid(), 3, 19.99);

INSERT INTO sales (store_id, sale_date, product_id, quantity, price)
VALUES (1, '2024-01-02', uuid(), 7, 39.99);

INSERT INTO sales (store_id, sale_date, product_id, quantity, price)
VALUES (2, '2024-01-01', uuid(), 10, 49.99);

-- GROUP BY queries
SELECT store_id, sale_date, SUM(quantity) as total_quantity, SUM(price) as total_price
FROM sales
GROUP BY store_id, sale_date;

SELECT store_id, COUNT(*) as transaction_count
FROM sales
GROUP BY store_id;

-- Improved materialized views
CREATE TABLE IF NOT EXISTS products (
    product_id uuid,
    category text,
    product_name text,
    price decimal,
    in_stock boolean,
    PRIMARY KEY (product_id)
);

INSERT INTO products (product_id, category, product_name, price, in_stock)
VALUES (uuid(), 'Electronics', 'Laptop', 999.99, true);

INSERT INTO products (product_id, category, product_name, price, in_stock)
VALUES (uuid(), 'Electronics', 'Phone', 599.99, true);

INSERT INTO products (product_id, category, product_name, price, in_stock)
VALUES (uuid(), 'Books', 'CQL Guide', 39.99, false);

CREATE MATERIALIZED VIEW IF NOT EXISTS products_by_category AS
    SELECT * FROM products
    WHERE category IS NOT NULL AND product_id IS NOT NULL
    PRIMARY KEY (category, product_id);

SELECT * FROM products_by_category WHERE category = 'Electronics';

-- CDC (Change Data Capture) table property (new in 3.11)
CREATE TABLE IF NOT EXISTS audit_log (
    id timeuuid PRIMARY KEY,
    user_id uuid,
    action text,
    timestamp timestamp,
    details text
) WITH cdc = true;

INSERT INTO audit_log (id, user_id, action, timestamp, details)
VALUES (now(), uuid(), 'LOGIN', toTimestamp(now()), 'User logged in successfully');

INSERT INTO audit_log (id, user_id, action, timestamp, details)
VALUES (now(), uuid(), 'UPDATE', toTimestamp(now()), 'Profile updated');

SELECT * FROM audit_log;

-- Improved SASI indexes
CREATE TABLE IF NOT EXISTS documents (
    doc_id uuid PRIMARY KEY,
    title text,
    content text,
    tags set<text>,
    created_date date
);

CREATE CUSTOM INDEX IF NOT EXISTS docs_title_idx ON documents (title)
USING 'org.apache.cassandra.index.sasi.SASIIndex'
WITH OPTIONS = {
    'mode': 'PREFIX',
    'case_sensitive': 'false'
};

CREATE CUSTOM INDEX IF NOT EXISTS docs_content_idx ON documents (content)
USING 'org.apache.cassandra.index.sasi.SASIIndex'
WITH OPTIONS = {
    'mode': 'CONTAINS',
    'case_sensitive': 'false',
    'analyzed': 'true',
    'analyzer_class': 'org.apache.cassandra.index.sasi.analyzer.StandardAnalyzer'
};

INSERT INTO documents (doc_id, title, content, tags, created_date)
VALUES (uuid(), 'Introduction to CQL', 'This is a comprehensive guide to CQL...', 
        {'cql', 'cassandra', 'tutorial'}, '2024-01-15');

INSERT INTO documents (doc_id, title, content, tags, created_date)
VALUES (uuid(), 'Advanced Cassandra', 'Deep dive into Cassandra internals...', 
        {'cassandra', 'advanced', 'internals'}, '2024-01-20');

-- SASI queries
SELECT * FROM documents WHERE title LIKE 'Intro%';
SELECT * FROM documents WHERE content LIKE '%comprehensive%';

-- Cast functions (improved in 3.11)
CREATE TABLE IF NOT EXISTS type_conversions (
    id uuid PRIMARY KEY,
    int_value int,
    text_value text,
    bigint_value bigint
);

INSERT INTO type_conversions (id, int_value, text_value, bigint_value)
VALUES (uuid(), 123, '456', 789);



-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_311;