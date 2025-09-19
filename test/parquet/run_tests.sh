#!/bin/bash

# Test runner for Parquet COPY TO/FROM functionality
# This script runs all Parquet-related tests in sequence

# Don't exit on test failures, we handle them ourselves
set +e

echo "========================================="
echo "Running CQLAI Parquet Tests"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Function to run a test
run_test() {
    local test_name=$1
    echo -n "Running $test_name... "

    if go test -v -timeout 30s ./... -run "^$test_name$" > /tmp/test_output_$$.log 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAILED${NC}"
        echo "Error output:"
        tail -20 /tmp/test_output_$$.log
        ((FAILED++))
    fi
    rm -f /tmp/test_output_$$.log
}

# Change to project root
cd "$(dirname "$0")/../.."

echo "Running tests from: $(pwd)"
echo

# Run tests in order of complexity
echo "1. Simple Types Tests"
echo "-----------------------------------------"
run_test "TestSimpleTypes"
run_test "TestNullHandling"
run_test "TestCompressionFormats"

echo
echo "2. Time Series / Wide Partition Tests"
echo "-----------------------------------------"
run_test "TestTimeSeriesData"
run_test "TestIOTDeviceMetrics"
run_test "TestEventLogData"

echo
echo "3. Collection Types Tests"
echo "-----------------------------------------"
run_test "TestListTypes"
run_test "TestSetTypes"
run_test "TestMapTypes"
run_test "TestFrozenCollections"
run_test "TestComplexNestedCollections"

echo
echo "4. UDT and Tuple Tests"
echo "-----------------------------------------"
run_test "TestUDTBasic"
run_test "TestComplexUDT"
run_test "TestTupleTypes"
run_test "TestMixedComplexTypes"

echo
echo "5. Vector Types Tests"
echo "-----------------------------------------"
run_test "TestVectorTypes"
run_test "TestHighDimensionalVectors"
run_test "TestVectorSimilarityData"

# Skip large dataset test in CI
if [ "$CI" != "true" ]; then
    echo
    echo "6. Large Dataset Test (Local Only)"
    echo "-----------------------------------------"
    run_test "TestLargeDataset"
fi

echo
echo "========================================="
echo "Test Results Summary"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi