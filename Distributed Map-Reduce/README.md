# Distributed Map-Reduce Implementation Documentation

## Overview

This is a distributed MapReduce implementation in Go using gRPC for communication between master and worker nodes. The system supports two main operations:
1. **Word Count** - Counts occurrences of each word across input files
2. **Inverted Index** - Creates an index mapping words to the files they appear in

## Architecture

### Components

1. **Master Server** (`server/main.go`)
    - Coordinates the entire MapReduce operation
    - Manages worker registration
    - Distributes tasks to workers
    - Runs on port 8080

2. **Worker Nodes** (`client/main.go`)
    - Execute map and reduce tasks
    - Register with master server
    - Process data files and generate intermediate results

3. **Protocol Buffers** (`protofiles/services.proto`)
    - Defines gRPC service interfaces
    - Message structures for communication

## Directory Structure

```
project/
├── client/
│   ├── main.go                    # Worker implementation
│   └── current_port_server.txt    # Port tracking file
├── server/
│   └── main.go                    # Master server implementation
├── protofiles/
│   └── services.proto             # gRPC service definitions
└── dataset/
     ├── file1.txt                  # Input data files
     ├── file2.txt
     └── ...
```

## Protocol Buffer Services

### MasterService

- `RegisterWorker(RegisterRequest) → RegisterResponse`
  - Allows workers to register with the master

### WorkerService

- `InvertedIndex(Task) → TaskResponse`
  - Executes inverted index map/reduce tasks
- `WordCount(Task) → TaskResponse`
  - Executes word count map/reduce tasks

### Message Types

```protobuf
message RegisterRequest {
     string address = 1;  // Worker IP address
     string port = 2;     // Worker port
}

message Task {
     int32 task = 1;      // Task type (1=map, 2=reduce, 3=exit)
     string filename = 2;  // Input filename
     int32 index = 3;     // Task index
}

message TaskResponse {
     bool success = 1;    // Task completion status
}
```

## Execution Flow

### 1. Master Server Startup

```bash
go run server/main.go <num_mappers> <num_reducers>
```
- Starts gRPC server on port 8080
- Launches `DivideTaskToWorkers()` goroutine
- Waits for worker registrations

### 2. Worker Node Startup

```bash
go run client/main.go <num_mappers> <num_reducers>
```
- Creates multiple worker goroutines (max of mappers/reducers)
- Each worker:
  - Gets unique port from `current_port_server.txt`
  - Registers with master server
  - Starts gRPC server to receive tasks

### 3. Task Distribution Sequence

#### Phase 1: Inverted Index

1. **Map Phase**: Master assigns map tasks to workers
    - Each mapper processes one input file
    - Extracts words and creates intermediate files
    - Uses SHA256 hashing for partitioning

2. **Reduce Phase**: Master assigns reduce tasks
    - Each reducer processes intermediate files from all mappers
    - Aggregates word-to-file mappings
    - Writes final results to `inverted_index_task.txt`

#### Phase 2: Word Count

1. **Map Phase**: Similar to inverted index mapping
    - Extracts words and counts occurrences
    - Creates intermediate files with word:count pairs

2. **Reduce Phase**: Aggregates word counts
    - Sums up counts for each word across all files
    - Writes results to `word_count_task.txt`

3. **Cleanup Phase**: Sends exit signals to all workers

## Key Features

### Load Balancing

- Uses SHA256 hashing to distribute words evenly across reducers
- Ensures balanced workload distribution

### Concurrency Control

- Mutex locks (`mu1`) protect shared file writes
- Synchronization prevents race conditions during output generation

### Fault Tolerance

- gRPC provides reliable communication
- Error handling for file operations and network calls

### Dynamic Port Management

- Workers automatically assign unique ports
- Port tracking through shared file system

## File Operations

### Input Files

- Located in `./dataset/` directory
- Named as `file1.txt`, `file2.txt`, etc.
- One file per mapper task

### Intermediate Files

#### Word Count

- Format: `reduce_wordcnt_{reducer_id}_{mapper_id}.txt`
- Content: `word:count` pairs

#### Inverted Index

- Format: `inverted_index_{reducer_id}_{mapper_id}.txt`
- Content: `word:filename` pairs

### Output Files

#### Word Count Results

- File: `word_count_task.txt`
- Format: `word:total_count`

#### Inverted Index Results

- File: `inverted_index_task.txt`
- Format: `word:file1 file2 file3`

## Requirements

### System Requirements

- Go 1.16 or higher
- Protocol Buffers compiler (`protoc`)
- gRPC Go libraries

### Dependencies

```go
import (
     "google.golang.org/grpc"
     "google.golang.org/grpc/credentials/insecure"
)
```

### File Structure Requirements

- `dataset/` directory with numbered input files
- `client/current_port_server.txt` for port management
- Write permissions in working directory

## Usage Instructions

### 1. Setup

```bash
# Install dependencies
go mod init mapreduce-project
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Generate protocol buffer files
protoc --go_out=. --go-grpc_out=. protofiles/services.proto
```

### 2. Prepare Data

```bash
# Create dataset directory
mkdir dataset

# Add input files
echo "sample text content" > dataset/file1.txt
echo "more sample data" > dataset/file2.txt
```

### 3. Initialize Port File

```bash
mkdir client
echo "9000" > client/current_port_server.txt
```

### 4. Run the System

```bash
# Terminal 1: Start master server
go run server/main.go 3 2  # 3 mappers, 2 reducers

# Terminal 2: Start workers
go run client/main.go 3 2  # Must match master parameters
```

### 5. Monitor Results

- Check `word_count_task.txt` for word frequency results
- Check `inverted_index_task.txt` for word-to-file mappings

## Configuration Parameters

- **Number of Mappers**: Should match number of input files
- **Number of Reducers**: Controls parallelism in reduce phase
- **Port Range**: Starting from value in `current_port_server.txt`

## Performance Considerations

- **I/O Bottleneck**: File operations can limit performance
- **Memory Usage**: Large files may require streaming approaches
- **Network Overhead**: gRPC calls add communication latency
- **Synchronization**: Mutex locks can create contention points