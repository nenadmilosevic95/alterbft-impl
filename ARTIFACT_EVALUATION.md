# Artifact Evaluation Guide

This guide is specifically designed for reviewers evaluating this artifact for reproducibility and validation of the paper's claims.

## Overview

This artifact contains the implementation of AlterBFT, a Byzantine fault-tolerant consensus protocol. The artifact is fully functional and can be tested using Docker (recommended) or directly with Go.

## Quick Evaluation (5 minutes)

For a quick verification that the artifact is functional:

1. **Run the demo**:
   ```bash
   ./run-demo.sh
   ```

2. **Verify output**:
   ```bash
   cd results
   md5sum deliveries.*
   ```
   All checksums should be identical, demonstrating that all nodes agreed on the same sequence of blocks.

## Detailed Evaluation (30 minutes)

### Test 1: Basic Consensus (4 nodes)

**Objective**: Verify that multiple nodes can reach consensus.

```bash
./run-demo.sh 4
```

**Expected result**: 
- All nodes complete 100 epochs
- All `deliveries.*` files contain identical content
- Each node delivers the same sequence of blocks

**Verification**:
```bash
cd results
# All files should have same line count
wc -l deliveries.*

# All files should be identical
for i in 1 2 3; do diff deliveries.0 deliveries.$i && echo "Node 0 and $i agree"; done
```

### Test 2: Scalability (7 nodes)

**Objective**: Verify that the system scales to larger deployments.

```bash
./run-demo.sh 7
```

**Expected result**:
- All 7 nodes reach consensus
- Agreement maintained across all nodes
- Similar throughput to 4-node setup

### Test 3: Byzantine Fault Tolerance

**Objective**: Verify that the system tolerates Byzantine failures.

```bash
cd bin
./test.sh 4 -byz 1 -attack silence
```

**Expected result**:
- System continues to operate despite 1 Byzantine node
- All honest nodes (3 out of 4) agree on the same blocks
- Byzantine node may have different or incomplete delivery log

**Verification**:
```bash
cd logs
# Check that honest nodes agree
diff deliveries.0 deliveries.2
diff deliveries.0 deliveries.3
```

### Test 4: Different Models

**Objective**: Verify that different consensus models work correctly.

```bash
cd bin
./test.sh 4 -mod alter
./test.sh 4 -mod alter -fast true
```

**Expected result**:
- Both configurations reach consensus
- FastAlter may show improved performance in logs

## Claims Validation

This section maps specific claims from the paper to reproducible experiments:

### Claim 1: Correctness
**Paper claim**: AlterBFT ensures agreement among all honest nodes.

**Validation**: Run any of the tests above and verify that all honest nodes produce identical `deliveries.*` files.

### Claim 2: Byzantine Fault Tolerance
**Paper claim**: AlterBFT tolerates up to f Byzantine nodes where 3f+1 ≤ n.

**Validation**: 
```bash
cd bin
./test.sh 4 -byz 1     # Tolerates 1 Byzantine (f=1, n=4)
./test.sh 7 -byz 2     # Tolerates 2 Byzantine (f=2, n=7)
```

### Claim 3: Liveness
**Paper claim**: The protocol makes progress under specified timing assumptions.

**Validation**: Run any test and verify that blocks are continuously delivered (check increasing line counts in `deliveries.*` files and timestamps in logs).

## Performance Benchmarks

For performance evaluation, run with different configurations:

### Throughput Test
```bash
cd bin
./test.sh 4 -maxEpoch 1000   # Run for 1000 epochs
```

Check logs for throughput statistics (blocks per second).

### Latency Test
```bash
cd bin
./test.sh 4 -s-delta 150 -b-delta 1000   # Standard deltas
./test.sh 4 -s-delta 100 -b-delta 500    # Aggressive deltas
```

Compare block delivery times in the logs.

## Output Interpretation

### Delivery Files Format

Each line in `deliveries.<node-id>` represents a delivered block:
```
epoch:block-hash:value-hash:timestamp
```

Example:
```
0:ABC123:DEF456:1634567890
1:GHI789:JKL012:1634567892
```

### Log Files

The `a.<node-id>` files contain detailed execution logs:
- **Peer discovery**: Shows how nodes find each other
- **Consensus phases**: Shows proposal, voting, and commit phases
- **Statistics**: Periodic performance metrics
- **Deliveries**: When blocks are finalized

Key log markers:
- `Bootstrap`: Initialization complete
- `Connected to X peers`: Network setup complete
- `Delivered block`: A block was finalized
- `Stats`: Performance statistics

## Common Issues

### Issue 1: Docker not available
If Docker is not installed, you can run directly with Go:
```bash
cd bin
./xcompile.sh   # Compile binaries
./test.sh 4     # Run test
```

### Issue 2: Ports already in use
If you see "address already in use" errors:
```bash
# Kill any existing processes
killall agent rendezvous
# Or reboot the system
```

### Issue 3: Slow performance in VM
If running in a virtual machine, performance may be reduced. Increase the timeout parameters:
```bash
cd bin
./test.sh 4 -s-delta 300 -b-delta 2000
```

## Hardware Requirements

**Minimum**:
- 2 CPU cores
- 4 GB RAM
- 1 GB disk space

**Recommended**:
- 4 CPU cores
- 8 GB RAM
- 10 GB disk space

## Estimated Time

- Initial setup (Docker): ~5 minutes
- Quick test (4 nodes, 100 epochs): ~30 seconds
- Full evaluation: ~30-60 minutes
- Performance benchmarks: ~2-3 hours

## File Structure

```
alterbft-impl/
├── README.md                    # Main documentation
├── ARTIFACT_EVALUATION.md       # This file
├── Dockerfile                   # Docker build configuration
├── docker-compose.yml           # Multi-container orchestration
├── run-demo.sh                 # Quick demo script
├── go.mod, go.sum              # Go dependencies
├── bin/
│   ├── test.sh                 # Test runner script
│   ├── xcompile.sh             # Compilation script
│   ├── agent/                  # Consensus node implementation
│   ├── rendezvous/             # Discovery server implementation
│   └── client/                 # Workload generator
├── consensus/                  # Core consensus logic
├── crypto/                     # Cryptographic operations
├── net/                        # Networking layer
├── types/                      # Data structures
└── workload/                   # Workload generation
```

## Reproducibility Checklist

- [ ] System builds successfully (Docker or Go)
- [ ] 4-node test completes without errors
- [ ] All nodes produce identical delivery files
- [ ] 7-node test completes successfully
- [ ] Byzantine fault tolerance test works (3/4 honest nodes agree)
- [ ] Different models (alter, fast-alter) both work
- [ ] Log files show expected consensus phases
- [ ] Performance is within expected range

## Support

If you encounter any issues during evaluation:

1. Check the Troubleshooting section in README.md
2. Examine the log files in `results/` or `bin/logs/` for error messages
3. Ensure system requirements are met
4. Try running with increased timeout values

For additional help, please contact: [Your email]

## Expected Evaluation Results

After completing the evaluation, you should observe:

1. **Functional correctness**: All honest nodes agree on the same block sequence
2. **Byzantine tolerance**: System operates correctly with f Byzantine nodes (f < n/3)
3. **Liveness**: System continuously makes progress and delivers blocks
4. **Performance**: Throughput comparable to values reported in the paper (accounting for hardware differences)

## Validation Statement

Upon successful completion of the evaluation, the artifact demonstrates:

- ✅ The implementation is functional and reproducible
- ✅ The core protocol properties (safety, liveness) are satisfied
- ✅ The system tolerates Byzantine faults as claimed
- ✅ Different configurations and models work as described in the paper

