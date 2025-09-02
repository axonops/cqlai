-- Cassandra 5.0 Test Suite
-- Compatible with Cassandra 5.0.x

-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS cqlai_test_50
WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE cqlai_test_50;

-- Vector type (new in 5.0)
CREATE TABLE IF NOT EXISTS vector_test (
    id uuid PRIMARY KEY,
    embedding vector<float, 3>,
    metadata text
);

INSERT INTO vector_test (id, embedding, metadata)
VALUES (uuid(), [0.1, 0.2, 0.3], 'First vector');

INSERT INTO vector_test (id, embedding, metadata)
VALUES (uuid(), [0.4, 0.5, 0.6], 'Second vector');

SELECT * FROM vector_test;

-- Storage-attached indexes (SAI) - improved in 5.0
CREATE TABLE IF NOT EXISTS products_sai (
    product_id uuid PRIMARY KEY,
    name text,
    description text,
    price decimal,
    category text,
    tags set<text>,
    in_stock boolean,
    rating float
);

-- Create SAI indexes (improved performance in 5.0)
CREATE INDEX IF NOT EXISTS products_name_sai ON products_sai (name) 
USING 'sai';

CREATE INDEX IF NOT EXISTS products_price_sai ON products_sai (price) 
USING 'sai';

CREATE INDEX IF NOT EXISTS products_category_sai ON products_sai (category) 
USING 'sai';

INSERT INTO products_sai (product_id, name, description, price, category, tags, in_stock, rating)
VALUES (uuid(), 'Laptop Pro', 'High-performance laptop', 1999.99, 'Electronics', 
        {'computer', 'portable', 'professional'}, true, 4.8);

INSERT INTO products_sai (product_id, name, description, price, category, tags, in_stock, rating)
VALUES (uuid(), 'Wireless Mouse', 'Ergonomic wireless mouse', 49.99, 'Electronics',
        {'computer', 'accessory', 'wireless'}, true, 4.5);

INSERT INTO products_sai (product_id, name, description, price, category, tags, in_stock, rating)
VALUES (uuid(), 'Programming Book', 'Learn Cassandra 5.0', 59.99, 'Books',
        {'education', 'database', 'cassandra'}, false, 4.9);

-- Complex SAI queries (improved in 5.0)
-- SELECT * FROM products_sai WHERE category = 'Electronics' AND in_stock = true;

-- Accord transactions (production-ready in 5.0)
-- Note: Syntax may vary based on final implementation
-- BEGIN TRANSACTION
--     INSERT INTO products_sai (product_id, name, description, price, category)
--     VALUES (uuid(), 'Transactional Product', 'Test product', 99.99, 'Test');
--     
--     UPDATE products_sai SET in_stock = false 
--     WHERE product_id = 550e8400-e29b-41d4-a716-446655440000;
-- COMMIT;

-- Unified Compaction Strategy (default in 5.0)
CREATE TABLE IF NOT EXISTS ucs_table (
    id uuid PRIMARY KEY,
    data text,
    timestamp timestamp
) WITH compaction = {
    'class': 'UnifiedCompactionStrategy',
    'scaling_parameters': 'T4'
};

INSERT INTO ucs_table (id, data, timestamp)
VALUES (uuid(), 'Data with UCS', toTimestamp(now()));

-- Improved guardrails configuration
CREATE TABLE IF NOT EXISTS guardrails_test (
    id uuid PRIMARY KEY,
    data1 text,
    data2 text,
    data3 text,
    data4 text,
    data5 text,
    data6 text,
    data7 text,
    data8 text,
    data9 text,
    data10 text
);

-- Standard table (trie memtables are used by default in 5.0)
CREATE TABLE IF NOT EXISTS trie_improvements (
    id uuid PRIMARY KEY,
    indexed_data text,
    value int
);

CREATE INDEX IF NOT EXISTS trie_indexed_data ON trie_improvements (indexed_data);

INSERT INTO trie_improvements (id, indexed_data, value)
VALUES (uuid(), 'searchable text', 100);

-- Dynamic data masking (new in 5.0)
CREATE TABLE IF NOT EXISTS sensitive_data (
    user_id uuid PRIMARY KEY,
    username text,
    email text,
    ssn text,
    credit_card text,
    phone text
);

-- Masking would be configured at the schema level
-- ALTER TABLE sensitive_data ALTER ssn MASK WITH DEFAULT;
-- ALTER TABLE sensitive_data ALTER credit_card MASK WITH partial(4);

INSERT INTO sensitive_data (user_id, username, email, ssn, credit_card, phone)
VALUES (uuid(), 'john_doe', 'john@example.com', '123-45-6789', '1234-5678-9012-3456', '555-0123');

-- Improved streaming with zero-copy (backend improvement)
CREATE TABLE IF NOT EXISTS streaming_optimized (
    id uuid PRIMARY KEY,
    large_blob blob,
    metadata map<text, text>
);

-- Native support for audit logging
CREATE TABLE IF NOT EXISTS audit_events (
    event_id timeuuid PRIMARY KEY,
    user text,
    action text,
    resource text,
    timestamp timestamp,
    details text
);

INSERT INTO audit_events (event_id, user, action, resource, timestamp, details)
VALUES (now(), 'admin', 'CREATE_TABLE', 'test_table', toTimestamp(now()), 'Table created');

-- Improved virtual tables
SELECT * FROM system_views.clients LIMIT 5;
SELECT * FROM system_views.sstable_tasks;

-- New metrics and monitoring views
SELECT * FROM system_views.coordinator_read_latency LIMIT 5;
SELECT * FROM system_views.local_read_latency LIMIT 5;

-- Improved prepared statement caching
CREATE TABLE IF NOT EXISTS prepared_test (
    id uuid PRIMARY KEY,
    data text,
    counter int
);

-- Better support for large partitions
CREATE TABLE IF NOT EXISTS large_partition_test (
    partition_key int,
    clustering_key int,
    data text,
    PRIMARY KEY (partition_key, clustering_key)
) WITH compaction = {
    'class': 'UnifiedCompactionStrategy'
};

-- Insert data for large partition handling
INSERT INTO large_partition_test (partition_key, clustering_key, data)
VALUES (1, 1, 'Data row 1');

INSERT INTO large_partition_test (partition_key, clustering_key, data)
VALUES (1, 2, 'Data row 2');

INSERT INTO large_partition_test (partition_key, clustering_key, data)
VALUES (1, 3, 'Data row 3');

-- Improved JMX-free operations (using virtual tables instead)
SELECT * FROM system_views.thread_pools;
SELECT * FROM system_views.caches;

-- Enhanced rate limiting
CREATE TABLE IF NOT EXISTS rate_limited_ops (
    id uuid PRIMARY KEY,
    operation_count counter
);

UPDATE rate_limited_ops SET operation_count = operation_count + 1
WHERE id = 550e8400-e29b-41d4-a716-446655440000;

-- Clean up
DROP KEYSPACE IF EXISTS cqlai_test_50;