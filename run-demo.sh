#!/bin/bash

# Demo script for AlterBFT artifact evaluation
# This script runs a simple demonstration of the consensus protocol

set -e

echo "========================================"
echo "  AlterBFT Consensus Demo"
echo "========================================"
echo ""

# Default number of nodes
N=${1:-4}
shift 2>/dev/null || true

echo "Configuration:"
echo "  Number of nodes: $N"
if [ -z "$*" ]; then
    echo "  Parameters: (defaults)"
else
    echo "  Parameters: $*"
fi
echo ""

# Check if Docker is available
if command -v docker &> /dev/null && command -v docker-compose &> /dev/null; then
    echo "Running with Docker..."
    echo ""
    
    # Clean up previous results
    rm -rf results
    mkdir -p results
    
    # Run with docker-compose
    docker-compose build
    docker-compose run --rm consensus-test ./test.sh $N "$@"
    
    echo ""
    echo "========================================"
    echo "  Demo Complete!"
    echo "========================================"
    echo ""
    echo "Results are available in the 'results' directory:"
    echo "  - deliveries.* files contain the delivered blocks for each node"
    echo "  - a.* files contain the detailed logs for each node"
    echo ""
    
    # Show summary
    if [ -d "results" ]; then
        echo "Summary of delivered blocks:"
        for i in $(seq 0 $((N-1))); do
            if [ -f "results/deliveries.$i" ]; then
                COUNT=$(wc -l < "results/deliveries.$i")
                echo "  Node $i: $COUNT blocks"
            fi
        done
    fi
    
elif command -v go &> /dev/null; then
    echo "Docker not found. Running directly with Go..."
    echo ""
    
    cd bin
    ./test.sh $N "$@"
    
    echo ""
    echo "========================================"
    echo "  Demo Complete!"
    echo "========================================"
    echo ""
    echo "Results are available in the 'bin/logs' directory"
    
else
    echo "ERROR: Neither Docker nor Go is installed on this system."
    echo "Please install either Docker or Go (1.16-1.18) to run this demo."
    exit 1
fi

