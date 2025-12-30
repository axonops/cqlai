#!/bin/bash
# Benchmark script for COPY FROM performance (issue #63)
# Tests various batch sizes and concurrency levels with 1 million rows

set -e

CQLAI=${CQLAI:-./cqlai}
ROWS=${ROWS:-1000000}
CSV_FILE="/tmp/benchmark_data.csv"
KEYSPACE="benchmark_copy"

echo "========================================"
echo "COPY FROM Benchmark - ${ROWS} rows"
echo "========================================"

# Create keyspace and table
echo "Setting up test keyspace and table..."
$CQLAI -e "DROP KEYSPACE IF EXISTS ${KEYSPACE};"
$CQLAI -e "CREATE KEYSPACE ${KEYSPACE} WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};"
$CQLAI -k ${KEYSPACE} -e "CREATE TABLE bulk_data (id int PRIMARY KEY, name text, value int, category text);"

# Generate CSV with 1 million rows using awk (much faster than shell loop)
echo "Generating ${ROWS} row CSV file..."
START_GEN=$(date +%s.%N)

echo "id,name,value,category" > ${CSV_FILE}
awk -v rows=${ROWS} 'BEGIN {
    categories[0]="alpha"; categories[1]="beta"; categories[2]="gamma";
    categories[3]="delta"; categories[4]="epsilon";
    for (i = 1; i <= rows; i++) {
        printf "%d,name_%d,%d,%s\n", i, i, i * 10, categories[i % 5]
    }
}' >> ${CSV_FILE}

END_GEN=$(date +%s.%N)
GEN_TIME=$(echo "$END_GEN - $START_GEN" | bc)
FILE_SIZE=$(du -h ${CSV_FILE} | cut -f1)

echo "CSV generated: ${FILE_SIZE} in ${GEN_TIME}s"
echo ""

# Test configurations (4 combinations covering range of settings)
# Format: "batch_size:concurrency"
CONFIGS="20:4 50:8 100:12 200:16"

echo "| Batch Size | Concurrency | Time (s) | Rows/sec |"
echo "|------------|-------------|----------|----------|"

for config in ${CONFIGS}; do
    batch=${config%:*}
    conc=${config#*:}

    # Truncate table for fresh import
    $CQLAI -k ${KEYSPACE} -e "TRUNCATE bulk_data;" 2>/dev/null

    # Run import and measure time
    START_TIME=$(date +%s.%N)
    $CQLAI -k ${KEYSPACE} -e "COPY bulk_data FROM '${CSV_FILE}' WITH HEADER=true AND MAXBATCHSIZE=${batch} AND MAXREQUESTS=${conc}" 2>/dev/null
    END_TIME=$(date +%s.%N)

    DURATION=$(echo "$END_TIME - $START_TIME" | bc)

    # Verify row count (extract count value from table output)
    COUNT_OUTPUT=$($CQLAI -k ${KEYSPACE} -e "SELECT COUNT(*) FROM bulk_data;" 2>/dev/null)
    # Extract the number between | characters (the actual count value)
    COUNT=$(echo "$COUNT_OUTPUT" | grep -E '^\| [0-9]' | grep -oE '[0-9]+' | head -1)

    if [ "$COUNT" = "${ROWS}" ]; then
        RATE=$(echo "scale=0; ${ROWS} / $DURATION" | bc)
        printf "| %10d | %11d | %8.2f | %8d |\n" $batch $conc $DURATION $RATE
    else
        printf "| %10d | %11d | FAILED (got %s rows) |\n" $batch $conc "$COUNT"
    fi
done

echo ""
echo "========================================"

# Cleanup
echo "Cleaning up..."
$CQLAI -e "DROP KEYSPACE ${KEYSPACE};"
rm -f ${CSV_FILE}

echo "Benchmark complete!"
