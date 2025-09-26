#!/bin/bash

# Production build script for cqlai
# This script builds an optimized binary for production use

set -e

echo "Building cqlai production binary..."

# Clean previous builds
rm -f cqlai bin/cqlai

# Build with optimizations
# -s: Strip symbol table
# -w: Strip DWARF debug info
# This reduces binary size significantly
echo "Building with optimizations..."
go build -ldflags="-s -w" -o cqlai cmd/cqlai/main.go

# Check if build succeeded
if [ ! -f cqlai ]; then
    echo "Error: Build failed"
    exit 1
fi

# Get binary size
SIZE=$(du -h cqlai | cut -f1)
echo "Build successful! Binary size: $SIZE"

# Optional: Copy to bin directory
if [ ! -d bin ]; then
    mkdir -p bin
fi
cp cqlai bin/

echo "Production binary available at:"
echo "  ./cqlai"
echo "  ./bin/cqlai"