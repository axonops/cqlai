#!/bin/bash

# Integration test runner for MCP (Model Context Protocol) integration
# These tests require a running Cassandra instance

set +e  # Don't exit on error, handle it ourselves

echo "========================================="
echo "Running MCP Integration Tests"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Change to project root
cd "$(dirname "$0")/../../.."

# Check if Cassandra is running
echo "Checking for Cassandra..."
if command -v cqlsh &> /dev/null; then
    if cqlsh -e "SELECT release_version FROM system.local" &> /dev/null; then
        echo -e "${GREEN}✓ Cassandra is running${NC}"
    else
        echo -e "${YELLOW}⚠ Cassandra is not running.${NC}"
        echo "  To run these tests, start Cassandra with:"
        echo "    podman run -d -p 9042:9042 cassandra:latest"
        echo "  or"
        echo "    docker run -d -p 9042:9042 cassandra:latest"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ cqlsh not found. Attempting to connect directly...${NC}"
fi

echo
echo "Running MCP integration tests..."
echo "-----------------------------------------"

# Run integration tests with build tag
if go test -tags=integration -v -timeout 120s ./test/integration/mcp/... 2>&1 | tee /tmp/mcp_integration_test_output.log; then
    echo -e "${GREEN}✓ All MCP integration tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some MCP integration tests failed!${NC}"
    echo "Check /tmp/mcp_integration_test_output.log for details"
    exit 1
fi
