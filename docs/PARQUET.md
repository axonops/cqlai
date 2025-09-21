# Parquet Support in CQLAI

CQLAI provides comprehensive support for Apache Parquet format, enabling efficient data import and export between Cassandra and Parquet files. This feature is particularly useful for data analytics, machine learning workflows, and data archival.

## Overview

Parquet is a columnar storage format that provides excellent compression and encoding schemes, making it ideal for storing and processing large datasets. CQLAI's Parquet integration allows you to:

- Export Cassandra table data to Parquet files
- Import Parquet files into Cassandra tables
- Handle complex Cassandra data types including collections, UDTs, and vectors
- Optimize storage with various compression algorithms

## COPY TO Parquet

Export data from Cassandra tables to Parquet format.

### Basic Usage

```sql
-- Export entire table to Parquet
COPY users TO '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- Export specific columns
COPY users (id, name, email) TO '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- Export with WHERE clause (if supported by your Cassandra version)
COPY users TO '/path/to/active_users.parquet' WITH FORMAT='PARQUET' WHERE status='active';
```

### Partitioned Datasets

Export data into partitioned directory structures for better organization and query performance:

```sql
-- Export with single partition column
COPY events TO '/data/events/' WITH FORMAT='PARQUET' AND PARTITION='date';

-- Export with multiple partition columns
COPY metrics TO '/data/metrics/' WITH FORMAT='PARQUET' AND PARTITION='year,month,day';

-- Result directory structure:
-- /data/metrics/
-- ├── year=2024/
-- │   ├── month=01/
-- │   │   ├── day=01/
-- │   │   │   └── part-00000.parquet
-- │   │   └── day=02/
-- │   │       └── part-00000.parquet
-- │   └── month=02/
-- │       └── day=01/
-- │           └── part-00000.parquet
```

Partitioning benefits:
- **Partition pruning**: Read only relevant partitions when filtering by partition columns
- **Parallel processing**: Different partitions can be processed concurrently
- **Incremental updates**: Add new partitions without rewriting existing data
- **Storage optimization**: Archive old partitions separately

### Compression Options

Parquet supports multiple compression algorithms:

```sql
-- Using Snappy compression (default, best balance)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY';

-- Using GZIP compression (better compression ratio)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='GZIP';

-- Using ZSTD compression (best compression ratio)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='ZSTD';

-- Using LZ4 compression (fastest)
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='LZ4';

-- No compression
COPY users TO 'users.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='NONE';
```

### Performance Optimization

Control chunk size for better performance with large datasets:

```sql
-- Set chunk size to 50,000 rows
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE=50000;

-- Use shorthand notation
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE='50K';
COPY huge_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE='1M';
```

## COPY FROM Parquet

Import data from Parquet files into Cassandra tables.

### Basic Usage

```sql
-- Import entire Parquet file
COPY users FROM '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- Import specific columns
COPY users (id, name, email) FROM '/path/to/users.parquet' WITH FORMAT='PARQUET';

-- Import with column mapping
COPY users (user_id, full_name) FROM 'data.parquet' WITH FORMAT='PARQUET';
```

### Importing Partitioned Datasets

Import data from partitioned Parquet datasets:

```sql
-- Import from partitioned directory
COPY events FROM '/data/events/' WITH FORMAT='PARQUET';

-- Import with partition filter (reduces data scanned)
COPY events FROM '/data/events/' WITH FORMAT='PARQUET' AND PARTITION_FILTER='year=2024,month=01';

-- Import specific partitions with pattern
COPY metrics FROM '/data/metrics/year=2024/' WITH FORMAT='PARQUET';
```

When importing partitioned datasets:
- Partition columns are automatically detected from directory structure
- Partition values are included as columns in the imported data
- Supports Hive-style partitioning convention (key=value)
- Handles special characters and NULL values appropriately

### Import Options

```sql
-- Skip header rows
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND SKIPROWS=1;

-- Limit number of rows to import
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND MAXROWS=10000;

-- Combine options
COPY users FROM 'users.parquet' WITH FORMAT='PARQUET' AND SKIPROWS=1 AND MAXROWS=5000;

-- Set batch size for partitioned datasets
COPY events FROM '/data/events/' WITH FORMAT='PARQUET' AND CHUNKSIZE=5000;
```

## Data Type Support

CQLAI's Parquet integration supports all major Cassandra data types:

### Basic Types

| Cassandra Type | Parquet Type | Notes |
|---------------|--------------|-------|
| text/varchar | STRING (UTF8) | Full Unicode support |
| int | INT32 | 32-bit signed integer |
| bigint | INT64 | 64-bit signed integer |
| float | FLOAT | 32-bit floating point |
| double | DOUBLE | 64-bit floating point |
| boolean | BOOLEAN | True/false values |
| timestamp | TIMESTAMP_MILLIS | Millisecond precision |
| date | DATE | Days since epoch |
| time | TIME_MILLIS | Milliseconds since midnight |
| uuid/timeuuid | STRING | Stored as formatted string |
| blob | BYTE_ARRAY | Binary data |
| decimal | DECIMAL | Arbitrary precision |

### Collection Types

```sql
-- Lists
CREATE TABLE products (
    id int PRIMARY KEY,
    tags list<text>,
    prices list<decimal>
);

-- Sets
CREATE TABLE users (
    id int PRIMARY KEY,
    emails set<text>,
    roles set<text>
);

-- Maps
CREATE TABLE settings (
    user_id int PRIMARY KEY,
    preferences map<text, text>,
    scores map<text, int>
);
```

### User-Defined Types (UDTs)

```sql
-- Define UDT
CREATE TYPE address (
    street text,
    city text,
    zip_code text,
    country text
);

-- Use in table
CREATE TABLE customers (
    id int PRIMARY KEY,
    name text,
    home_address address,
    work_address address
);

-- Export/import preserves UDT structure
COPY customers TO 'customers.parquet' WITH FORMAT='PARQUET';
COPY customers FROM 'customers.parquet' WITH FORMAT='PARQUET';
```

### Vector Types (Cassandra 5.0+)

Support for machine learning and similarity search use cases:

```sql
-- Create table with vector column
CREATE TABLE embeddings (
    id int PRIMARY KEY,
    content text,
    vector list<float>,  -- Vector embeddings
    metadata text
);

-- Export vectors to Parquet
COPY embeddings TO 'embeddings.parquet' WITH FORMAT='PARQUET';

-- Vectors are stored as LIST types in Parquet
-- Compatible with Apache Arrow and pandas
```

## Advanced Features

### Streaming Large Datasets

For very large tables, CQLAI uses streaming to minimize memory usage:

```sql
-- Export large table with optimized streaming
COPY large_events_table TO 'events.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD'
AND CHUNKSIZE='100K';
```

### Capture Mode Integration

CQLAI's CAPTURE command provides an interactive way to save query results to Parquet files, which is fundamentally different from the COPY command:

#### Partitioned Capture

Save captured query results to partitioned datasets for better organization:

```sql
-- Start partitioned capture with single partition column
CAPTURE PARQUET '/data/analysis/' WITH PARTITION='date';

-- Subsequent queries will be partitioned by date value
SELECT * FROM events WHERE date >= '2024-01-01';
-- Creates: /data/analysis/date=2024-01-01/part-00000.parquet
--          /data/analysis/date=2024-01-02/part-00000.parquet
--          etc.

-- Multi-column partitioning
CAPTURE PARQUET '/data/metrics/' WITH PARTITION='year,month,day';

SELECT * FROM metrics WHERE year = 2024;
-- Creates: /data/metrics/year=2024/month=01/day=01/part-00000.parquet
--          /data/metrics/year=2024/month=01/day=02/part-00000.parquet

CAPTURE OFF;
```

##### Virtual Column Extraction from TimeUUIDs

A powerful feature of partitioned capture is the ability to extract time components from TimeUUID columns for partitioning:

```sql
-- Table with TimeUUID primary key
CREATE TABLE events (
    event_id timeuuid PRIMARY KEY,
    event_name text,
    event_value int
);

-- Partition by year and month extracted from TimeUUID
CAPTURE PARQUET '/data/events/' WITH PARTITION='event_id.year,event_id.month';

SELECT * FROM events;
-- Creates: /data/events/event_id.year=2024/event_id.month=01/part-00000.parquet
--          /data/events/event_id.year=2024/event_id.month=02/part-00000.parquet

-- Virtual columns can be extracted:
-- .year   - Extract year from TimeUUID
-- .month  - Extract month from TimeUUID
-- .day    - Extract day from TimeUUID
-- .hour   - Extract hour from TimeUUID
-- .date   - Extract date as YYYY-MM-DD string

CAPTURE OFF;
```

Virtual columns are only used for directory partitioning and are not stored in the Parquet files themselves. When querying with tools like DuckDB or Apache Spark, these partition values are automatically available as columns based on the Hive-style directory structure.

##### Compression and Performance Options

Control compression and file sizes for optimal performance:

```sql
-- Use ZSTD compression for better compression ratio
CAPTURE PARQUET '/data/compressed/' WITH COMPRESSION='ZSTD' AND PARTITION='date';

-- Use LZ4 for fastest compression
CAPTURE PARQUET '/data/fast/' WITH COMPRESSION='LZ4';

-- Control maximum file size (useful for partitioned datasets)
CAPTURE PARQUET '/data/sized/' WITH MAX_FILE_SIZE='500MB' AND PARTITION='date';
-- When a partition file exceeds 500MB, a new file (part-00001.parquet) is created

-- Combine all options
CAPTURE PARQUET '/data/optimized/' WITH
    PARTITION='event_id.year,event_id.month'
    AND COMPRESSION='ZSTD'
    AND MAX_FILE_SIZE='1GB';

CAPTURE OFF;
```

Benefits of partitioned capture:
- Organize large datasets by time or category
- Enable efficient data lifecycle management
- Support incremental processing pipelines
- Optimize downstream analytics queries
- Automatic virtual column extraction from TimeUUIDs
- Compatible with Hive-style partitioned datasets

#### Key Differences from COPY

| Aspect | COPY | CAPTURE |
|--------|------|---------|
| **Purpose** | Bulk export/import of entire tables | Save results of ad-hoc queries |
| **Scope** | Single table operation | Multiple queries across any tables |
| **Use Case** | Data migration, backup, ETL | Exploratory analysis, reporting |
| **Execution** | Immediate, single operation | Session-based, continuous |
| **Data Source** | Table data with optional filters | Any SELECT query results |

#### How Capture Works

```sql
-- Start capturing - subsequent query results will be saved
CAPTURE PARQUET '/tmp/analysis_results.parquet';

-- Run a query - results are written to the Parquet file
SELECT * FROM users WHERE country='US';
-- This creates a Parquet file with columns: id, name, email, country, etc.

-- IMPORTANT: Subsequent queries must have the SAME schema
SELECT * FROM users WHERE country='UK';  -- ✓ Works - same columns
SELECT * FROM users WHERE age > 18;      -- ✓ Works - same columns

-- This would FAIL or create issues - different columns!
-- SELECT id, order_total FROM orders;   -- ✗ Different schema

-- Stop capturing
CAPTURE OFF;
```

**Schema Limitation**: When capturing to Parquet, all queries in a capture session must return the same columns in the same order. This is because Parquet files have a fixed schema that cannot change mid-file.

#### Paging Behavior

When capturing large result sets, CQLAI automatically handles paging:

```sql
-- Start capture with Parquet format
CAPTURE PARQUET '/tmp/large_results.parquet';

-- This query might return millions of rows
SELECT * FROM events WHERE date >= '2024-01-01';
-- CQLAI will automatically page through results:
-- - Fetches data in chunks (default 5000 rows per page)
-- - Writes each page to the Parquet file
-- - Shows progress: "Page 1 of 1000..."
-- - Continues until all data is captured
-- - Memory efficient - only one page in memory at a time

CAPTURE OFF;
```

#### Capture Syntax Examples

```sql
-- Basic capture to Parquet
CAPTURE PARQUET '/tmp/results.parquet';

-- Capture with compression
CAPTURE PARQUET '/tmp/compressed.parquet' WITH COMPRESSION='ZSTD';

-- Capture with partitioning
CAPTURE PARQUET '/tmp/partitioned/' WITH PARTITION='date';

-- Capture with all options
CAPTURE PARQUET '/data/output/' WITH
    PARTITION='year,month'
    AND COMPRESSION='LZ4'
    AND MAX_FILE_SIZE='100MB';

-- Check current capture status
CAPTURE;

-- Stop capturing
CAPTURE OFF;
```

#### Use Cases for Capture with Parquet

1. **Filtering and Combining Same-Schema Results**
   ```sql
   -- Capture filtered results from the same table
   CAPTURE '/tmp/filtered_users.parquet' FORMAT='PARQUET';
   SELECT * FROM users WHERE country='US' AND status='active';
   SELECT * FROM users WHERE country='UK' AND status='active';
   SELECT * FROM users WHERE country='CA' AND status='active';
   CAPTURE OFF;
   -- All queries have the same schema, so they append correctly
   ```

2. **Time-Series Data Collection**
   ```sql
   -- Capture hourly snapshots with same schema
   CAPTURE '/tmp/metrics_snapshot.parquet' FORMAT='PARQUET';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 00:00:00';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 01:00:00';
   SELECT hour, metric_name, value FROM metrics WHERE hour='2024-01-01 02:00:00';
   CAPTURE OFF;
   ```

3. **Paged Export of Large Tables**
   ```sql
   -- Export a large table in manageable chunks
   CAPTURE '/tmp/large_export.parquet' FORMAT='PARQUET';
   SELECT * FROM events WHERE date='2024-01-01' LIMIT 100000;
   SELECT * FROM events WHERE date='2024-01-02' LIMIT 100000;
   SELECT * FROM events WHERE date='2024-01-03' LIMIT 100000;
   CAPTURE OFF;
   ```

**Note**: For capturing results from different queries with different schemas, consider using JSON or CSV format instead:
```sql
-- JSON format can handle different schemas
CAPTURE '/tmp/mixed_results.json' FORMAT='JSON';
SELECT COUNT(*) as user_count FROM users;
SELECT id, name, email FROM users LIMIT 10;
SELECT order_id, total FROM orders LIMIT 10;
CAPTURE OFF;
```

#### Important Notes

- **Schema Consistency**: All queries in a Parquet capture session must have identical schemas
- **Append Behavior**: Each query result appends rows to the same Parquet file (same schema required)
- **Memory Efficiency**: Large results are paged automatically, keeping memory usage constant
- **Progress Indication**: Shows current page number for large result sets
- **Format Alternative**: Use JSON or CSV formats for capturing queries with different schemas

### File Detection

CQLAI automatically detects Parquet format from file extension:

```sql
-- Automatic format detection
COPY users TO 'users.parquet';  -- Automatically uses PARQUET format
COPY users FROM 'data.parquet'; -- Automatically detected as PARQUET
```

## Use Cases

### 1. Data Analytics Pipeline

Export Cassandra data for analysis in Apache Spark, pandas, or other analytics tools:

```sql
-- Export for Spark processing
COPY events TO 's3://bucket/events.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY';

-- Python pandas can read directly:
-- df = pd.read_parquet('events.parquet')
```

### 2. Data Archival

Archive historical data with excellent compression:

```sql
-- Archive old data with maximum compression
COPY historical_data TO '/archive/data_2023.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD'
WHERE year=2023;
```

### 3. Machine Learning Workflows

Export vectors and features for ML training:

```sql
-- Export embeddings and features
COPY ml_features TO 'training_data.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='SNAPPY';

-- Load in Python for training:
-- features = pd.read_parquet('training_data.parquet')
-- X = np.stack(features['vector'].values)
```

### 4. Data Migration

Migrate data between Cassandra clusters:

```sql
-- Source cluster: Export
COPY users TO 'users_backup.parquet' WITH FORMAT='PARQUET';

-- Target cluster: Import
COPY users FROM 'users_backup.parquet' WITH FORMAT='PARQUET';
```

## Performance Considerations

### Chunk Size

- Default: 10,000 rows per chunk
- For large rows: Decrease chunk size (e.g., 1000-5000)
- For small rows: Increase chunk size (e.g., 50000-100000)

### Compression Trade-offs

| Compression | Speed | Ratio | Use Case |
|------------|-------|-------|----------|
| NONE | Fastest | None | Temporary files, fast I/O |
| SNAPPY | Fast | Good | Default, balanced performance |
| LZ4 | Very Fast | Good | Real-time processing |
| GZIP | Slow | Better | Network transfer |
| ZSTD | Slower | Best | Long-term storage |

### Memory Usage

- Streaming mode minimizes memory footprint
- Chunk size affects memory usage: `memory ≈ chunk_size × avg_row_size`
- For very wide tables, reduce chunk size

## Limitations

1. **Nested Collections**: Deeply nested collections (e.g., `list<map<text, set<int>>>`) may have limited support
2. **Custom Types**: Custom Cassandra types may be converted to strings
3. **Streaming Only**: COPY FROM uses streaming - entire file is not loaded into memory
4. **Schema Matching**: Column names and types must be compatible between Parquet and Cassandra

## Troubleshooting

### Common Issues

**Issue**: "Cannot convert type X to Parquet"
```sql
-- Solution: Check data type compatibility
DESCRIBE TABLE your_table;
-- Ensure all types are supported
```

**Issue**: Out of memory errors
```sql
-- Solution: Reduce chunk size
COPY large_table TO 'data.parquet' WITH FORMAT='PARQUET' AND CHUNKSIZE=1000;
```

**Issue**: Slow export performance
```sql
-- Solution: Increase chunk size and use faster compression
COPY table TO 'data.parquet' WITH FORMAT='PARQUET'
AND COMPRESSION='LZ4'
AND CHUNKSIZE='50K';
```

## Examples

### Complete Export/Import Workflow

```sql
-- 1. Create source table
CREATE KEYSPACE IF NOT EXISTS analytics
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE analytics;

CREATE TABLE IF NOT EXISTS user_events (
    user_id uuid,
    event_time timestamp,
    event_type text,
    properties map<text, text>,
    vector list<float>,
    PRIMARY KEY (user_id, event_time)
) WITH CLUSTERING ORDER BY (event_time DESC);

-- 2. Insert sample data
INSERT INTO user_events (user_id, event_time, event_type, properties, vector)
VALUES (uuid(), toTimestamp(now()), 'click',
        {'page': 'home', 'button': 'signup'},
        [0.1, 0.2, 0.3, 0.4, 0.5]);

-- 3. Export to Parquet with compression
COPY user_events TO '/tmp/events.parquet'
WITH FORMAT='PARQUET'
AND COMPRESSION='ZSTD';

-- 4. Create destination table
CREATE TABLE IF NOT EXISTS user_events_archive (
    user_id uuid,
    event_time timestamp,
    event_type text,
    properties map<text, text>,
    vector list<float>,
    PRIMARY KEY (user_id, event_time)
);

-- 5. Import from Parquet
COPY user_events_archive FROM '/tmp/events.parquet'
WITH FORMAT='PARQUET';

-- 6. Verify import
SELECT COUNT(*) FROM user_events_archive;
```

### Complete Partitioned Capture Workflow

```sql
-- 1. Create a table with TimeUUID
CREATE TABLE IF NOT EXISTS events (
    event_id timeuuid PRIMARY KEY,
    event_type text,
    user_id int,
    data text
);

-- 2. Insert sample data
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'login', 123, 'user logged in');
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'purchase', 123, 'bought item ABC');
INSERT INTO events (event_id, event_type, user_id, data)
VALUES (now(), 'logout', 123, 'user logged out');

-- 3. Start partitioned capture by year and month from TimeUUID
CAPTURE PARQUET '/data/events/' WITH
    PARTITION='event_id.year,event_id.month'
    AND COMPRESSION='ZSTD';

-- 4. Execute queries - results are partitioned automatically
SELECT * FROM events WHERE event_type = 'login';
SELECT * FROM events WHERE event_type = 'purchase';
SELECT * FROM events WHERE user_id = 123;

-- 5. Stop capturing
CAPTURE OFF;

-- 6. Query the partitioned data with DuckDB
-- Files are organized as:
-- /data/events/event_id.year=2024/event_id.month=01/part-00000.parquet
-- /data/events/event_id.year=2024/event_id.month=02/part-00000.parquet

-- 7. Query with partition pruning in DuckDB
-- Only reads files from January 2024
-- duckdb -c "SELECT * FROM '/data/events/**/*.parquet' WHERE \"event_id.year\" = 2024 AND \"event_id.month\" = 1;"
```

## Best Practices

1. **Always specify FORMAT='PARQUET'** for clarity, even when using .parquet extension
2. **Use compression** for production exports (SNAPPY or ZSTD recommended)
3. **Test chunk sizes** with your specific data patterns
4. **Monitor memory usage** during large exports
5. **Validate data** after import with COUNT(*) and sample queries
6. **Use appropriate compression** based on use case (storage vs. speed)
7. **Consider partitioning** large exports by time or other dimensions

## Planned Features

The following features are planned for future releases:

### Near-term (v0.1.x)

1. **Schema Evolution Support**
   - Automatic schema mapping for column additions/removals
   - Type conversion warnings and options
   - Schema validation before import

2. **Parallel Processing**
   - Multi-threaded export for large tables
   - Concurrent chunk processing
   - Parallel file writing for partitions

3. **S3/Cloud Storage Integration**
   ```sql
   -- Direct S3 export
   COPY users TO 's3://bucket/path/users.parquet'
   WITH FORMAT='PARQUET'
   AND AWS_PROFILE='default';

   -- Azure Blob Storage
   COPY users TO 'azure://container/users.parquet'
   WITH FORMAT='PARQUET';
   ```

4. **Statistics and Metadata**
   - Row group statistics for query optimization
   - Column statistics (min, max, null count)
   - Bloom filters for efficient filtering
   - Custom metadata in Parquet files

### Medium-term (v0.2.x)

1. **Advanced Data Types**
   - Full nested collection support (arbitrary depth)
   - Geometry types for spatial data
   - JSON column type mapping
   - Custom type handlers

2. **Incremental Export/Import**
   ```sql
   -- Export only changes since last export
   COPY users TO 'users_delta.parquet'
   WITH FORMAT='PARQUET'
   AND SINCE='2024-01-01 00:00:00';

   -- Merge import (upsert)
   COPY users FROM 'users_update.parquet'
   WITH FORMAT='PARQUET'
   AND MODE='UPSERT';
   ```

3. **Data Transformation**
   ```sql
   -- Transform during export
   COPY users TO 'users.parquet'
   WITH FORMAT='PARQUET'
   AND TRANSFORM='{"email": "LOWER", "created_at": "DATE_ONLY"}';
   ```

4. **Compression Profiles**
   ```sql
   -- Predefined optimization profiles
   COPY large_table TO 'data.parquet'
   WITH FORMAT='PARQUET'
   AND PROFILE='ANALYTICS';  -- Optimized for Spark/Presto

   COPY ml_data TO 'features.parquet'
   WITH FORMAT='PARQUET'
   AND PROFILE='ML';  -- Optimized for Python/Arrow
   ```

5. **Progress Monitoring**
   - Real-time progress bars
   - ETA calculation
   - Detailed statistics during operation
   - Resumable operations

### Long-term (v0.3.x+)

1. **Apache Arrow Integration**
   - Zero-copy data transfer
   - Arrow Flight protocol support
   - Direct memory format compatibility
   - Improved Python/pandas interoperability

2. **Delta Lake Format**
   ```sql
   -- Export as Delta table
   COPY users TO '/delta/users'
   WITH FORMAT='DELTA';

   -- Time travel queries
   COPY users FROM '/delta/users'
   WITH FORMAT='DELTA'
   AND VERSION='2024-01-01';
   ```

3. **Streaming CDC (Change Data Capture)**
   ```sql
   -- Continuous export of changes
   CAPTURE STREAM changes TO 'kafka://topic'
   FROM users
   WITH FORMAT='PARQUET'
   AND MODE='CDC';
   ```

4. **Query Pushdown**
   - Parquet predicate pushdown
   - Column pruning optimization
   - Row group filtering
   - Smart data skipping

5. **Data Quality Features**
   - Data validation rules
   - Automatic data cleansing
   - Duplicate detection
   - Data profiling reports

6. **Integration with ML Frameworks**
   ```sql
   -- Direct export to ML formats
   COPY features TO 'model_data.tfrecord'
   WITH FORMAT='TENSORFLOW';

   COPY embeddings TO 'vectors.lance'
   WITH FORMAT='LANCE';  -- Optimized for vector search
   ```

7. **Distributed Operations**
   - Coordinator-worker architecture
   - Distributed export across nodes
   - Load balancing for import
   - Fault tolerance and retry logic

8. **Advanced Security**
   - Column-level encryption in Parquet
   - Field-level masking during export
   - Audit logging for all operations
   - Role-based access control for exports

## Feature Requests

We welcome feature requests and contributions! Please submit ideas through:
- GitHub Issues: [github.com/axonops/cqlai/issues](https://github.com/axonops/cqlai/issues)
- Discussions: [github.com/axonops/cqlai/discussions](https://github.com/axonops/cqlai/discussions)

Priority is given to features that:
1. Improve performance for large-scale operations
2. Enhance compatibility with data analytics ecosystems
3. Support common enterprise use cases
4. Maintain backward compatibility

## Related Documentation

- [COPY Command Reference](./COPY.md)
- [Data Types Guide](./DATA_TYPES.md)
- [Performance Tuning](./PERFORMANCE.md)
- [Apache Parquet Format](https://parquet.apache.org/docs/)
- [Apache Arrow](https://arrow.apache.org/)
- [Delta Lake](https://delta.io/)