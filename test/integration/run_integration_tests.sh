#!/bin/bash

# Integration test runner for COPY TO/FROM with Cassandra
# These tests require a running Cassandra instance

set +e  # Don't exit on error, handle it ourselves

echo "========================================="
echo "Running CQLAI Cassandra Integration Tests"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Change to project root
cd "$(dirname "$0")/../.."

# Check if Cassandra is running
echo "Checking for Cassandra..."
if command -v cqlsh &> /dev/null; then
    if cqlsh -e "SELECT release_version FROM system.local" &> /dev/null; then
        echo -e "${GREEN}✓ Cassandra is running${NC}"
    else
        echo -e "${YELLOW}⚠ Cassandra is not running. Starting Cassandra may be required.${NC}"
        echo "  To run these tests, start Cassandra with:"
        echo "    docker run -d -p 9042:9042 cassandra:latest"
        echo "  or"
        echo "    cassandra -f"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ cqlsh not found. These tests require Cassandra.${NC}"
    echo "  Install Cassandra or run with Docker:"
    echo "    docker run -d -p 9042:9042 cassandra:latest"
    exit 1
fi

echo
echo "Running integration tests..."
echo "-----------------------------------------"

# Run MCP integration tests first
echo
echo "=== MCP Integration Tests ==="
if ./test/integration/mcp/run_mcp_tests.sh; then
    echo -e "${GREEN}✓ MCP tests passed${NC}"
    MCP_PASSED=true
else
    echo -e "${RED}✗ MCP tests failed${NC}"
    MCP_PASSED=false
fi

# Run other integration tests
echo
echo "=== COPY TO/FROM Tests ==="
if go test -v -timeout 120s ./test/integration/... 2>&1 | tee /tmp/integration_test_output.log; then
    echo -e "${GREEN}✓ COPY tests passed${NC}"
    COPY_PASSED=true
else
    echo -e "${RED}✗ COPY tests failed${NC}"
    echo "Check /tmp/integration_test_output.log for details"
    COPY_PASSED=false
fi

echo
echo "========================================="
echo "Integration Test Summary"
echo "========================================="
if [ "$MCP_PASSED" = true ]; then
    echo -e "${GREEN}✓ MCP Integration Tests: PASSED${NC}"
else
    echo -e "${RED}✗ MCP Integration Tests: FAILED${NC}"
fi

if [ "$COPY_PASSED" = true ]; then
    echo -e "${GREEN}✓ COPY TO/FROM Tests: PASSED${NC}"
else
    echo -e "${RED}✗ COPY TO/FROM Tests: FAILED${NC}"
fi

# Exit with failure if any test suite failed
if [ "$MCP_PASSED" = true ] && [ "$COPY_PASSED" = true ]; then
    echo
    echo -e "${GREEN}All integration tests passed!${NC}"
    exit 0
else
    echo
    echo -e "${RED}Some integration tests failed!${NC}"
    exit 1
fi