# AlterBFT: Byzantine Fault-Tolerant Consensus Algorithm Implementation

This repository contains two main components:

1. **AlterBFT**: A Byzantine fault-tolerant consensus protocol implementation
2. **Delta Protocol**: A measurement tool for collecting message latencies between distributed machines

> **📖 Open Source Research Code**
> 
> This implementation is open-sourced to enable **experimentation, learning, and further research**. Feel free to use it, modify it, and build upon it for your own research or educational purposes. We provide an easy Docker-based setup so you can quickly run experiments on your local machine.
>
> **Note**: This is a research prototype. While it demonstrates the core protocol correctly, it is not production-ready. Use it for research, education, and experimentation.

## Table of Contents
- [Overview](#overview)
  - [System Model](#system-model)
- [Quick Start with Docker](#quick-start-with-docker)
- [Manual Setup](#manual-setup)
- [Running Experiments](#running-experiments)
- [Output and Results](#output-and-results)
- [Configuration Options](#configuration-options)
- [Use Cases and Limitations](#use-cases-and-limitations)

## Overview

This repository provides two key implementations written in Go:

1. **AlterBFT Protocol**: A Byzantine fault-tolerant consensus protocol that allows you to experiment with BFT consensus, run tests, and build upon the implementation.

2. **Delta Protocol**: A measurement tool we used to collect network latencies between machines in our distributed experiments. This helps characterize real-world network delays for consensus protocol evaluation.

### System Model

AlterBFT operates under a **hybrid synchrony** system model with a key distinction:
- **Small messages** (votes, certificates) are assumed to be **timely** and delivered within known bounds
- **Large messages** (proposed blocks with transaction data) can be **delayed arbitrarily** and are expected to be eventually timely

This model reflects real-world networks where small control messages typically have predictable latency, while large data transfers may experience variable delays. See our paper for detailed analysis of this system model.

**Components included:**

- **AlterBFT protocol implementation**:
  - Complete core consensus algorithm with fast-alter optimization
  - Byzantine fault tolerance with configurable Byzantine nodes
  - Agent nodes that participate in consensus
  
- **Delta protocol implementation**:
  - Measures message latencies between distributed machines
  - Used for network characterization in our experiments
  
- **Infrastructure**:
  - Rendezvous server for peer-to-peer discovery
  - Performance monitoring and comprehensive logging
  - Docker setup for easy local experimentation
  
**Key features:**
- Run either AlterBFT consensus or Delta measurement protocol
- Highly configurable for different experimental scenarios
- Support for Byzantine fault injection (silence, equivocation attacks)
- Detailed logging and performance metrics
- Easy-to-use Docker setup requiring no manual dependency management

## Quick Start with Docker

The easiest way to try AlterBFT is using Docker. Just run one command and see the consensus protocol in action!

### Prerequisites
- Docker (version 20.10 or later)
- Docker Compose (version 1.29 or later)

### Running the Demo

1. **Clone or extract the repository**:
   ```bash
   cd alterbft-impl
   ```

2. **Run the AlterBFT consensus demonstration**:
   ```bash
   ./run-demo.sh
   ```

   This will:
   - Build the Docker image
   - Start 4 consensus nodes running AlterBFT
   - Run 100 epochs of consensus
   - Store results in the `results/` directory

3. **Run with custom number of nodes**:
   ```bash
   ./run-demo.sh 7   # Runs with 7 nodes
   ```

### Understanding the Output

After the demo completes, you'll find the following in the `results/` directory:

- `deliveries.0`, `deliveries.1`, etc.: Delivered blocks for each node (one block per line), each node prints block for height when it was the proposer
- `a.0`, `a.1`, etc.: Detailed execution logs for each node

## Manual Setup (Without Docker)

**Note**: Manual setup can be problematic due to Go version dependencies. This codebase requires Go 1.16-1.18, but most modern systems have Go 1.19+ installed, which is incompatible with some of our dependencies (specifically `quic-go`). **We provide Docker to avoid these dependency issues** and ensure the code works consistently across different machines.

If you still want to build manually and have the correct Go version:

### Prerequisites
- Go 1.16 to 1.18 (**NOT 1.19+** - incompatible dependencies)
- Git

### Check Your Go Version
```bash
go version
# If you see go1.19 or higher, use Docker instead to avoid dependency issues
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

You can run both AlterBFT consensus experiments and Delta latency measurements using `./run-demo.sh` with Docker (recommended).

### AlterBFT Consensus Experiments

**1. Different system sizes**:
```bash
./run-demo.sh 4    # 4 nodes (f=1)
./run-demo.sh 7    # 7 nodes (f=2)
./run-demo.sh 10   # 10 nodes (f=3)
```

**2. Byzantine nodes**:
```bash
./run-demo.sh 4 -byz 1 -mod silence   # 1 Byzantine node with silence attack
./run-demo.sh 7 -byz 2 -mod equiv     # 2 Byzantine nodes with equivocation attack
```

**3. Fast Alter optimization**:
```bash            
./run-demo.sh 4 -mod alter -fast    # AlterBFT with fast-alter optimization
```
**4. Custom timeouts**:
```bash
./run-demo.sh 4 -s-delta 200 -b-delta 1500   # Small delta: 200ms, Big delta: 1500ms
```

### Delta Protocol (Latency Measurement)

To run the Delta protocol for measuring network latencies between machines:

```bash
./run-demo.sh 4 -mod delta              # Run Delta protocol to collect message delays
```

```bash
./run-demo.sh 4 -mod delta-chunk              # Run Delta protocol where each message is chunked in X small messages
```

The Delta protocol is used to characterize network behavior rather than reaching consensus. It measures how long messages take to travel between nodes.



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
- `-s-delta <MS>`: Small delta timeout (milliseconds), used for small messages
- `-b-delta <MS>`: Big delta timeout (milliseconds), used for large messages
- `-maxEpoch <N>`: Number of consensus epochs to run
- `-mod <MODEL>`: Consensus model (alter, delta, silence, equiv)
- `-fast`: Enable FastAlter optimization

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

## Configuration Options

### Protocol Models (`-mod` parameter)

- **alter**: AlterBFT consensus protocol (default)
  - Use with `-fast true` for fast-alter optimization
- **delta**: Delta protocol for measuring message delays between distributed nodes
  - Used for network characterization experiments
  - Not a consensus protocol, but a measurement tool
- **alter with byzantine nodes**: Alter with Byzantine nodes
   - When running with `-byz N` to specify N Byzantine nodes:
   - **silence**: Byzantine leaders remain silent (do not propose blocks)
   - **equiv**: Byzantine leaders send equivocating proposals (different proposals to different nodes) 


## Use Cases and Limitations

**What you can do with this code:**
- 🔬 **Experiment** with Byzantine consensus protocols
- 📚 **Learn** how BFT consensus works in practice
- 🔧 **Build** your own consensus protocol variations
- 📊 **Benchmark** different configurations and parameters
- 🎓 **Teach** distributed systems and consensus concepts
- 💻 **Run local tests** on your laptop with Docker

**Important limitations:**
- This is a **research prototype**, not production-ready software
- Intended for **controlled experimental environments**
- This repository contains code for **local testing** and experimentation

**Large-scale distributed experiments:**

The paper presents results from large-scale experiments on AWS EC2 instances. The orchestration scripts for deploying and running experiments across multiple EC2 instances are **not currently included in this repository** but may be added in the future. Those scripts essentially:
- Upload compiled binaries to EC2 instances
- Configure appropriate parameters for distributed settings
- Coordinate experiment execution across nodes

For now, this repository focuses on making the core protocol accessible for local experimentation and learning.

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
[nele.milosevic95@gmail.com]

## Acknowledgments

This implementation uses:
- [libp2p](https://libp2p.io/) for peer-to-peer networking
- [Tendermint](https://tendermint.com/) libraries for cryptographic operations
