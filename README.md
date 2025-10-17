# AlterBFT: Byzantine Fault-Tolerant Consensus Algorithm Implementation

This repository contains an open-source implementation of AlterBFT, a Byzantine fault-tolerant consensus protocol.

> **📖 Open Source Research Code**
> 
> This implementation is open-sourced to enable **experimentation, learning, and further research**. Feel free to use it, modify it, and build upon it for your own research or educational purposes. We provide an easy Docker-based setup so you can quickly run experiments on your local machine.
>
> **Note**: This is a research prototype. While it demonstrates the core protocol correctly, it is not production-ready. Use it for research, education, and experimentation.

## Table of Contents
- [Overview](#overview)
- [Quick Start with Docker](#quick-start-with-docker)
- [Manual Setup](#manual-setup)
- [Running Experiments](#running-experiments)
- [Output and Results](#output-and-results)
- [Configuration Options](#configuration-options)
- [System Requirements](#system-requirements)
- [Use Cases and Limitations](#use-cases-and-limitations)

## Overview

AlterBFT is a BFT consensus protocol implementation written in Go. This open-source implementation allows you to experiment with Byzantine consensus, run your own tests, and build upon the protocol.

**What's included:**

- **Agent nodes**: Consensus participants that propose and agree on blocks
- **Rendezvous server**: A discovery service for peer-to-peer connectivity
- **Docker setup**: Run experiments instantly without worrying about dependencies

**Key features:**
- Byzantine fault tolerance with configurable number of Byzantine nodes
- Complete AlterBFT implementation (alter, fast-alter)
- Performance monitoring and comprehensive logging
- Highly configurable for different experimental scenarios
- Easy-to-use Docker setup for quick experimentation

## Quick Start with Docker (Recommended)

The easiest way to try AlterBFT is using Docker. Just run one command and see the consensus protocol in action!

### Prerequisites
- Docker (version 20.10 or later)
- Docker Compose (version 1.29 or later)

### Running the Demo

1. **Clone or extract the repository**:
   ```bash
   cd alterbft-impl
   ```

2. **Run the demonstration**:
   ```bash
   ./run-demo.sh
   ```

   This will:
   - Build the Docker image
   - Start 4 consensus nodes
   - Run 100 epochs of consensus
   - Store results in the `results/` directory

3. **Run with custom number of nodes**:
   ```bash
   ./run-demo.sh 7   # Runs with 7 nodes
   ```

### Understanding the Output

After the demo completes, you'll find the following in the `results/` directory:

- `deliveries.0`, `deliveries.1`, etc.: Delivered blocks for each node (one block per line)
- `a.0`, `a.1`, etc.: Detailed execution logs for each node

**Success criteria**: All nodes should deliver the same sequence of blocks. You can verify this by checking that all `deliveries.*` files have the same content.

Example verification:
```bash
# Check that all nodes delivered the same blocks
cd results
md5sum deliveries.*   # All checksums should be identical
```

## Manual Setup (Advanced)

⚠️ **WARNING**: Manual setup requires Go 1.16-1.18. If you have Go 1.19+ installed, **use Docker instead** (recommended above) or downgrade Go to 1.18.

Most systems now have Go 1.19+ by default, which is **incompatible** with this codebase due to dependency constraints. **We strongly recommend using Docker** unless you specifically need to modify the code.

### Prerequisites
- Go 1.16 to 1.18 (**NOT 1.19+**)
- Git

### Check Your Go Version
```bash
go version
# If you see go1.19 or higher, use Docker instead!
```

### Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd alterbft-impl
   ```

2. **Download dependencies**:
   ```bash
   go mod download
   ```

3. **Compile the binaries**:
   ```bash
   cd bin
   ./xcompile.sh
   ```

   This creates three binaries:
   - `bin/agent/agent`: Consensus node
   - `bin/rendezvous/rendezvous`: Discovery server
   - `bin/client/client`: Workload generator

### Running Manually

1. **Navigate to the bin directory**:
   ```bash
   cd bin
   ```

2. **Run N nodes**:
   ```bash
   ./test.sh 4   # Runs 4 consensus nodes
   ```

3. **View results**:
   ```bash
   cat logs/deliveries.0   # View blocks delivered by node 0
   wc -l logs/deliveries.*  # Count blocks delivered by each node
   ```

## Running Experiments

All experiments can be run using `./run-demo.sh` with Docker (recommended).

### Basic Experiments

**1. Different system sizes**:
```bash
./run-demo.sh 4    # 4 nodes (f=1)
./run-demo.sh 7    # 7 nodes (f=2)
./run-demo.sh 10   # 10 nodes (f=3)
```

**2. Byzantine nodes**:
```bash
./run-demo.sh 4 -byz 1 -attack silence   # 1 Byzantine node with silence attack
./run-demo.sh 7 -byz 2 -attack equiv     # 2 Byzantine nodes with equivocation attack
```

**3. Network models**:
```bash
./run-demo.sh 4 -mod alter              # Standard AlterBFT
./run-demo.sh 4 -mod alter -fast true   # FastAlter optimization
```

**4. Custom timeouts**:
```bash
./run-demo.sh 4 -s-delta 200 -b-delta 1500   # Small delta: 200ms, Big delta: 1500ms
```

### Manual Mode (Advanced - Requires Go 1.16-1.18)

If you're running without Docker and have Go 1.16-1.18 installed:

```bash
cd bin
./test.sh 4 -byz 1 -attack silence
./test.sh 7 -s-delta 200 -b-delta 1500
```

### Advanced Configuration

The agent binary supports many configuration options:

```bash
cd bin/agent
./agent -h   # Show all available options
```

Key parameters:
- `-n <N>`: Total number of nodes
- `-i <ID>`: Node ID (0 to N-1)
- `-byz <F>`: Number of Byzantine nodes
- `-attack <TYPE>`: Byzantine behavior (silence, equiv)
- `-s-delta <MS>`: Small delta timeout (milliseconds)
- `-b-delta <MS>`: Big delta timeout (milliseconds)
- `-maxEpoch <N>`: Number of consensus epochs to run
- `-mod <MODEL>`: Consensus model (alter, fast-alter)
- `-fast`: Enable FastAlter optimization
- `-topology <TYPE>`: Network topology (full, gossip, star)

## Output and Results

### Delivery Files

Each node produces a `deliveries.<node-id>` file containing one delivered block per line. Format:
```
<epoch>:<block-hash>:<block-value-hash>:<timestamp>
```

### Log Files

Each node produces a log file `a.<node-id>` with detailed execution information including:
- Peer discovery and connection
- Consensus rounds and phases
- Message send/receive events
- Performance statistics

### Verification

To verify correct consensus:

1. **Check all nodes delivered blocks**:
   ```bash
   wc -l results/deliveries.*
   ```
   All files should have the same line count.

2. **Verify agreement**:
   ```bash
   cd results
   diff deliveries.0 deliveries.1   # Should show no differences
   diff deliveries.0 deliveries.2   # Should show no differences
   ```

3. **Check for Byzantine tolerance** (when using `-byz`):
   Even with Byzantine nodes, all honest nodes should agree on the same sequence.

## Configuration Options

### Network Topologies

- **full**: All-to-all connectivity (default)
- **gossip**: Random gossip overlay (-k parameter sets neighbors)
- **star**: Star topology with node 0 as hub

### Consensus Models

- **alter**: Standard AlterBFT protocol
- **fast-alter**: With FastAlter optimization

### Byzantine Attacks

- **silence**: Byzantine leaders remain silent
- **equiv**: Byzantine leaders send conflicting messages

## System Requirements

### Recommended Setup (Docker)
- CPU: 2+ cores
- RAM: 4+ GB
- Disk: 1 GB free space
- Docker Engine: 20.10+
- Docker Compose: 1.29+

### Alternative Setup (Manual Build) - Advanced Users Only
- CPU: 2+ cores
- RAM: 4+ GB
- Disk: 1 GB free space
- **Go: 1.16 to 1.18 (NOT 1.19+)** ⚠️
- Git

**Note**: Most modern systems have Go 1.19+ which is incompatible. Use Docker unless you specifically need to modify code.

## Troubleshooting

### Docker Issues

**Problem**: Docker build fails with "go: module not found"
```bash
# Solution: Clean Docker cache and rebuild
docker-compose build --no-cache
```

**Problem**: Permission denied on results directory
```bash
# Solution: Fix permissions
sudo chown -R $USER:$USER results/
```

### Manual Build Issues

**Problem**: "command not found: go"
```bash
# Solution: Install Go 1.16 to 1.18
# See: https://golang.org/doc/install
# Note: Use Go 1.18 or earlier due to dependency compatibility
```

**Problem**: Build fails with quic-go error on Go 1.19+
```
Error: "The version of quic-go you're using can't be built on Go 1.19 yet"
```
```bash
# Solution: Use Go 1.18 or earlier
# The project dependencies require Go 1.18 or earlier
# Download Go 1.18 from: https://go.dev/dl/
```

**Problem**: "cannot find package"
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

### Runtime Issues

**Problem**: Nodes can't find each other
- Check that all nodes use the same rendezvous server
- Verify network connectivity between nodes
- Increase discovery timeout with longer initial wait

**Problem**: Consensus stalls
- Check logs for error messages
- Verify all nodes have started successfully
- Ensure sufficient time deltas for network conditions

## Expected Performance

On a modern laptop/workstation with Docker:

- **4 nodes**: ~50-100 blocks in 100 epochs (~20-30 seconds)
- **7 nodes**: ~40-80 blocks in 100 epochs (~30-40 seconds)
- **10 nodes**: ~30-60 blocks in 100 epochs (~40-60 seconds)

Actual performance depends on:
- CPU speed and core count
- Available memory
- Network latency (especially important for distributed deployments)
- Configured timeout values (s-delta, b-delta)

## Use Cases and Limitations

**What you can do with this code:**
- 🔬 **Experiment** with Byzantine consensus protocols
- 📚 **Learn** how BFT consensus works in practice
- 🔧 **Build** your own consensus protocol variations
- 📊 **Benchmark** different configurations and parameters
- 🎓 **Teach** distributed systems and consensus concepts

**Important limitations:**
- This is a **research prototype**, not production-ready software
- Intended for **controlled experimental environments** and local testing
- Optimized for **clarity and experimental flexibility** over production performance
- Best suited for **trusted, controlled networks** (like your laptop or test cluster)

If you're building a production system, treat this as a reference implementation and starting point, but plan for significant additional engineering, security auditing, and testing.

## Contributing

We welcome contributions, improvements, and feedback! Feel free to:
- 🐛 Report bugs or issues
- 💡 Suggest new features or optimizations
- 🔀 Submit pull requests
- 📖 Improve documentation
- 💬 Share your experiments and results

## Citation

If you use this code in your research or build upon it, please cite our paper:

```
[Your paper citation here]
```

## License

[Specify your license here]

## Contact

Questions, suggestions, or want to discuss the implementation?
[Your contact information]

## Acknowledgments

This implementation uses:
- [libp2p](https://libp2p.io/) for peer-to-peer networking
- [Tendermint](https://tendermint.com/) libraries for cryptographic operations
