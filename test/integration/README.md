# Integration Tests for COPY TO/FROM with Parquet

## Current Status

The tests in `test/parquet/` directory validate Parquet file read/write operations but **do not** test the actual COPY TO/FROM integration with Cassandra tables.

## What's Tested Currently

1. **Parquet File Operations** (`test/parquet/*.go`)
   - Writing various Cassandra data types to Parquet files
   - Reading data back from Parquet files
   - Verifying data integrity in the Parquet files
   - Testing collections, UDTs, vectors, and time series patterns

## What's NOT Tested (Yet)

1. **COPY FROM Parquet to Cassandra**
   - Creating Cassandra tables
   - Importing Parquet files into Cassandra using COPY FROM
   - Querying Cassandra to verify imported data

2. **COPY TO from Cassandra to Parquet**
   - Populating Cassandra tables with test data
   - Exporting to Parquet using COPY TO
   - Verifying exported Parquet matches source data

3. **Round-trip Testing**
   - Cassandra → Parquet → Cassandra data integrity

## Running Integration Tests

Integration tests require a running Cassandra instance:

```bash
# Start Cassandra with Docker
docker run -d -p 9042:9042 --name cassandra-test cassandra:latest

# Wait for Cassandra to be ready (30-60 seconds)
docker exec cassandra-test cqlsh -e "SELECT release_version FROM system.local"

# Run integration tests
./test/integration/run_integration_tests.sh

# Clean up
docker stop cassandra-test
docker rm cassandra-test
```

## Test Implementation Notes

The integration tests in `copy_parquet_integration_test.go` are designed to test the full flow but require proper integration with the cqlai router's MetaCommandHandler. Currently these are placeholder tests that demonstrate what should be tested.

To properly test COPY TO/FROM:
1. Need to instantiate MetaCommandHandler with a valid session
2. Call the actual COPY TO/FROM methods
3. Verify data in Cassandra tables
