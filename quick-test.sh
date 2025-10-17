#!/bin/bash

# Quick test script that works without Docker
# Compiles and runs a simple 4-node consensus test

set -e

echo "========================================"
echo "  AlterBFT Quick Test"
echo "========================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed."
    echo "Please install Go 1.16-1.18 from https://golang.org/doc/install"
    echo "Note: Go 1.19+ has dependency compatibility issues"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Found Go: $GO_VERSION"
echo ""

# Navigate to bin directory
cd bin

# Compile binaries
echo "Compiling binaries..."
./xcompile.sh

echo ""
echo "Starting 4-node consensus test..."
echo "Press Ctrl+C to stop the test"
echo ""
sleep 2

# Run the test with 4 nodes
./test.sh 4

echo ""
echo "========================================"
echo "  Test Complete!"
echo "========================================"
echo ""
echo "Results are in bin/logs/ directory:"
ls -lh logs/deliveries.* 2>/dev/null || echo "No delivery files found"
echo ""
echo "To verify consensus, check that all nodes delivered the same blocks:"
echo "  cd bin/logs && md5sum deliveries.*"

