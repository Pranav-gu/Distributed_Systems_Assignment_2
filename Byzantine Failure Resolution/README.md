# Byzantine Fault Tolerance (BFT) Implementation Documentation

## Overview

This project implements a Byzantine Fault Tolerance system using the Byzantine Generals Problem algorithm. The system consists of a Commander (server) and multiple Generals (clients) that must reach consensus on a binary decision (Attack or Retreat) even in the presence of Byzantine failures.

## System Architecture

### Components

1. **Commander (server.go)** - Central coordinator that sends orders to generals
2. **Generals (client.go)** - Distributed nodes that receive orders and perform cross-verification
3. **Protocol Definitions (services.proto)** - gRPC service definitions for communication

## Requirements

### System Requirements
- **Operating System**: Linux, macOS, or Windows
- **Go Version**: Go 1.21 or higher (uses `math/rand/v2`)
- **Network**: TCP/IP networking capability
- **Ports**: Multiple available ports for gRPC communication

### Dependencies
```bash
# Core dependencies
go mod init Q4
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Protocol Buffers compiler
# On Ubuntu/Debian:
sudo apt-get install protobuf-compiler

# On macOS:
brew install protobuf

# On Windows:
# Download from https://github.com/protocolbuffers/protobuf/releases

# Go plugins for protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Project Structure
```
Q4/
├── server
│   ├── main.go           # Commander implementation
├── client
│   ├── main.go           # General implementation (to be created)
├── protofiles/         # Generated protobuf files
│   ├── services.pb.go
│   └── services_grpc.pb.go
│   ├── services.proto
├── go.mod
└── go.sum
```

## Byzantine Fault Tolerance Algorithm

### The Byzantine Generals Problem

The Byzantine Generals Problem is a fundamental problem in distributed computing that deals with achieving consensus in the presence of Byzantine failures (arbitrary, potentially malicious behavior).

### Algorithm Implementation

#### Core Principles
1. **f-Byzantine Fault Tolerance**: The system can tolerate up to `t` Byzantine failures among `n` total nodes
2. **Consensus Requirement**: All honest generals must agree on the same action
3. **Cross-Verification**: Generals verify orders by communicating with each other

#### Fault Tolerance Guarantee
- **Total Nodes**: `n = 3t + 1` (minimum requirement for Byzantine fault tolerance)
- **Maximum Failures**: Can tolerate up to `t` Byzantine failures
- **Honest Majority**: Requires at least `2t + 1` honest nodes

### Algorithm Steps

1. **Registration Phase**:
   - Generals register with the Commander
   - Commander establishes connections to all generals

2. **Order Distribution Phase**:
   - Commander sends orders to all generals
   - Honest commander sends the same order to all
   - Byzantine commander may send different orders

3. **Cross-Verification Phase**:
   - Each general queries other generals about received orders
   - Generals apply majority voting to determine the true order
   - Byzantine generals may lie about received orders

4. **Consensus Phase**:
   - Each general decides based on majority of received information
   - Honest generals reach the same decision despite Byzantine failures

## File Operations and Functionalities

### server.go (Commander)

#### Core Functionalities
- **General Registration**: Accepts connections from generals
- **Order Broadcasting**: Sends attack/retreat orders to all generals
- **Byzantine Behavior Simulation**: Can act dishonestly if configured

#### Key Operations
- `RegisterGeneral()`: Handles general registration requests
- `OrderGenerals()`: Broadcasts orders to all registered generals
- Connection management with gRPC clients

#### Configuration Parameters
- `n`: Total number of generals + 1 (commander)
- `t`: Number of Byzantine nodes
- `dishonest`: Boolean flag for commander's behavior

### client.go (General) - To Be Implemented

#### Expected Functionalities
- **Commander Registration**: Register with the commander
- **Order Reception**: Receive orders from commander
- **Cross-Verification**: Query other generals for their received orders
- **Consensus Decision**: Apply majority voting to determine final action

#### Key Operations (Expected)
- `CommanderBroadcast()`: Handle orders from commander
- `CrossVerification()`: Respond to queries from other generals
- Majority voting algorithm implementation

### services.proto (Protocol Definitions)

#### Service Definitions

1. **CommanderService**:
   - `RegisterGeneral`: General registration with commander

2. **GeneralService**:
   - `CommanderBroadcast`: Receive orders from commander
   - `CrossVerification`: Cross-verify orders with other generals

#### Message Types
- `RegisterRequest/Response`: Registration messages
- `Order/OrderResponse`: Command distribution messages
- `Enquire/EnquireResponse`: Cross-verification messages

## Installation and Setup

### 1. Project Initialization
```bash
# Create project directory
mkdir byzantine-fault-tolerance
cd byzantine-fault-tolerance

# Initialize Go module
go mod init Q4

# Install dependencies
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

### 2. Generate Protocol Buffer Files
```bash
# Create protofiles directory
mkdir protofiles

# Generate Go files from proto
protoc --go_out=./protofiles --go_opt=paths=source_relative \
       --go-grpc_out=./protofiles --go-grpc_opt=paths=source_relative \
       services.proto
```

### 3. Build the Project
```bash
# Build server
go build -o server server.go

# Build client (once implemented)
go build -o client client.go
```

## Running Instructions

### Step 1: Start the Commander
```bash
# Syntax: ./server <total_nodes> <byzantine_count> <node_types...>
# Example: 4 total nodes, 1 byzantine, commander is honest (1)
./server 4 1 1

# Example: 4 total nodes, 1 byzantine, commander is dishonest (0)
./server 4 1 0
```

### Step 2: Start the Generals
```bash
# Start each general on different ports
# General 1
./client <port1>

# General 2
./client <port2>

# General 3
./client <port3>
```

### Example Execution Sequence
```bash
# Terminal 1 - Start honest commander with 4 nodes, 1 byzantine
./server 4 1 1

# Terminal 2 - Start General 1
./client 8081

# Terminal 3 - Start General 2  
./client 8082

# Terminal 4 - Start General 3
./client 8083
```

## Configuration Parameters

### Command Line Arguments

#### Server (Commander)
- `n`: Total number of nodes (commander + generals)
- `t`: Number of Byzantine nodes in the system
- `node_types`: Array indicating honest (1) or Byzantine (0) behavior

#### Client (General)
- `port`: Port number for the general to listen on

### System Limits
- **Minimum Nodes**: 4 (1 commander + 3 generals for t=1)
- **Port Range**: Dynamic allocation between available ports
- **Timeout**: No explicit timeout (can be added for production)

## Testing Scenarios

### 1. All Honest Nodes
```bash
./server 4 0 1  # No Byzantine nodes
```

### 2. Byzantine Commander
```bash
./server 4 1 0  # Commander is Byzantine
```

### 3. Maximum Byzantine Tolerance
```bash
./server 7 2 1  # 6 generals, up to 2 can be Byzantine
```

## Security Considerations

### Threats Mitigated
- **Byzantine Failures**: Arbitrary malicious behavior
- **Inconsistent Commands**: Different orders to different generals  
- **False Information**: Lying about received orders

## Performance Characteristics

### Time Complexity
- **Registration**: O(n) for n generals
- **Order Distribution**: O(n) messages from commander
- **Cross-Verification**: O(n²) messages between generals
- **Overall**: O(n²) message complexity

### Space Complexity
- **Commander**: O(n) for storing general connections
- **General**: O(n) for storing peer connections and votes