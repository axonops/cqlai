#!/bin/bash

# Test script for COPY TO and COPY FROM commands

set -e

echo "=== Testing COPY Commands ==="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Build the binary if needed
if [ ! -f ./cqlai ]; then
    echo "Building cqlai..."
    go build -o cqlai cmd/cqlai/main.go
fi

# Create test data directory
TEST_DIR="/tmp/cqlai_copy_test_$$"
mkdir -p "$TEST_DIR"

# Clean up function
cleanup() {
    rm -rf "$TEST_DIR"
    rm -f /tmp/test_export_*.csv
    rm -f /tmp/test_import_*.csv
}

# Set trap to clean up on exit
trap cleanup EXIT

# Create test CSV files for import testing
cat > "$TEST_DIR/test_import.csv" <<EOF
id,name,price,active
10,John,25.99,true
11,Jane,30.50,false
12,Jim,35.00,true
EOF

cat > "$TEST_DIR/test_import_no_header.csv" <<EOF
20,Jack,28.99,true
21,Jill,22.00,false
22,Joe,33.75,true
EOF

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"

    echo -n "Testing $test_name... "

    if output=$(./cqlai -e "$command" 2>&1); then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        echo "  Error: $output"
        return 1
    fi
}

# Initialize test keyspace
echo "Setting up test environment..."
./cqlai -e "CREATE KEYSPACE IF NOT EXISTS copy_test WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};" > /dev/null 2>&1
./cqlai -k copy_test -e "CREATE TABLE IF NOT EXISTS products (id int PRIMARY KEY, name text, price decimal, active boolean);" > /dev/null 2>&1

# Insert test data
./cqlai -k copy_test -e "TRUNCATE products;" > /dev/null 2>&1
./cqlai -k copy_test -e "INSERT INTO products (id, name, price, active) VALUES (1, 'Laptop', 999.99, true);" > /dev/null 2>&1
./cqlai -k copy_test -e "INSERT INTO products (id, name, price, active) VALUES (2, 'Mouse', 29.99, true);" > /dev/null 2>&1
./cqlai -k copy_test -e "INSERT INTO products (id, name, price, active) VALUES (3, 'Keyboard', 79.99, false);" > /dev/null 2>&1

echo ""
echo "=== COPY TO Tests ==="

# Test COPY TO commands
run_test "COPY TO basic" "COPY copy_test.products TO '$TEST_DIR/export_basic.csv'"
run_test "COPY TO with header" "COPY copy_test.products TO '$TEST_DIR/export_header.csv' WITH HEADER=true"
run_test "COPY TO specific columns" "COPY copy_test.products (id, name) TO '$TEST_DIR/export_columns.csv' WITH HEADER=true"
run_test "COPY TO with delimiter" "COPY copy_test.products TO '$TEST_DIR/export_pipe.csv' WITH HEADER=true AND DELIMITER='|'"

# Verify exported files exist
echo ""
echo "Verifying exported files..."
for file in "$TEST_DIR"/export_*.csv; do
    if [ -f "$file" ]; then
        echo -e "  ${GREEN}✓${NC} $(basename "$file") exists ($(wc -l < "$file") lines)"
    else
        echo -e "  ${RED}✗${NC} $(basename "$file") missing"
    fi
done

echo ""
echo "=== COPY FROM Tests ==="

# Test COPY FROM commands
./cqlai -k copy_test -e "TRUNCATE products;" > /dev/null 2>&1

run_test "COPY FROM with header" "COPY copy_test.products FROM '$TEST_DIR/test_import.csv' WITH HEADER=true"

# Verify import
output=$(./cqlai -k copy_test -e "SELECT COUNT(*) FROM products;" 2>/dev/null)
count=$(echo "$output" | grep "| [0-9]" | awk '{print $2}')
if [ "$count" = "3" ]; then
    echo -e "  ${GREEN}✓${NC} Imported correct number of rows: $count"
else
    echo -e "  ${RED}✗${NC} Expected 3 rows, got: $count"
    echo "  Debug output: $output"
fi

# Test COPY FROM without header
run_test "COPY FROM without header" "COPY copy_test.products (id, name, price, active) FROM '$TEST_DIR/test_import_no_header.csv'"

# Test COPY FROM with MAXROWS
./cqlai -k copy_test -e "TRUNCATE products;" > /dev/null 2>&1
run_test "COPY FROM with MAXROWS" "COPY copy_test.products FROM '$TEST_DIR/test_import.csv' WITH HEADER=true AND MAXROWS=2"

output=$(./cqlai -k copy_test -e "SELECT COUNT(*) FROM products;" 2>/dev/null)
count=$(echo "$output" | grep "| [0-9]" | awk '{print $2}')
if [ "$count" = "2" ]; then
    echo -e "  ${GREEN}✓${NC} MAXROWS working: imported $count rows"
else
    echo -e "  ${RED}✗${NC} MAXROWS failed: expected 2 rows, got: $count"
    echo "  Debug output: $output"
fi

# Test COPY FROM with SKIPROWS
./cqlai -k copy_test -e "TRUNCATE products;" > /dev/null 2>&1
run_test "COPY FROM with SKIPROWS" "COPY copy_test.products FROM '$TEST_DIR/test_import.csv' WITH HEADER=true AND SKIPROWS=1"

echo ""
echo "=== Test Summary ==="
echo -e "${GREEN}COPY command tests completed${NC}"

# Clean up test keyspace
echo ""
echo "Cleaning up test environment..."
./cqlai -k copy_test -e "DROP TABLE IF EXISTS products;" > /dev/null 2>&1
./cqlai -e "DROP KEYSPACE IF EXISTS copy_test;" > /dev/null 2>&1

echo "Done!"