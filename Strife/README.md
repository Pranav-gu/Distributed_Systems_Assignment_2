# Distributed Payment System Analysis

## System Architecture

This distributed payment system consists of three main components:

- **Payment Gateway** (`payment_gateway.go`): Acts as the coordinator service.
- **Bank Server** (`server.go`): Represents individual bank services.
- **Client** (`client.go`): Provides the user interface for transactions.

---

## Requirements

### Software Dependencies

- Go programming language
- gRPC and Protocol Buffers
- TLS certificates for secure communication
- CSV file handling capabilities

### System Requirements

- Network connectivity between all components
- File system access for persistence
- Port availability (8080 for gateway, dynamic ports for banks/clients - can be changed easily in current_port_server.txt files)

### Required Files/Directories

```
./certificate/
├── server_cert.pem
├── server_key.pem
├── ca_cert.pem
├── client_cert.pem
├── client_key.pem
└── script.sh (certificate generation script)

./payment_gateway/
├── transactions.txt
└── transaction_logs.json

./server/
├── dummydata.csv
├── current_port_server.txt
└── pending_transactions.json

./client/
├── current_port_server.txt
└── offline_payments.json

Q3/protofiles/ (gRPC service definitions)
```

---

## Instructions to Run Programs

### 1. Setup Phase

```bash
# Ensure all certificate files are in ./certificate/ directory
# Ensure CSV data file exists at ./server/dummydata.csv with format:
# username,balance,password,bank

# Initialize port tracking files
echo "8080" > ./server/current_port_server.txt
echo "9000" > ./client/current_port_server.txt
```

### 2. Start Payment Gateway

```bash
go run payment_gateway.go
# Runs on port 8080
# Acts as the transaction coordinator
```

### 3. Start Bank Servers

```bash
# For each bank server (multiple instances possible)
go run server.go <server_id>
# Example: go run server.go 1
# Server will prompt for bank name during registration
```

### 4. Start Clients

```bash
go run client.go
# Each client gets a unique port automatically
# Provides interactive menu for user operations
```

---

## Functionalities

### Payment Gateway (`payment_gateway.go`)

#### Core Services

- **RegisterClient**: Authenticates and registers client connections.
- **RegisterServer**: Registers bank servers with the gateway.
- **MakePayment**: Executes 2-phase commit transactions.
- **ViewBalance**: Retrieves account balance from banks.

#### Transaction Management

- **2-Phase Commit Protocol**:
    - Phase 1: Prepare both source and destination banks.
    - Phase 2: Commit or abort based on preparation results.
- **Transaction Logging**: Persistent storage of transaction states.
- **Idempotency Control**: Prevents duplicate transaction processing.
- **Recovery Mechanisms**: Handles partial commits and system failures.

#### Security Features

- TLS encryption for all communications.
- Certificate-based authentication.
- Mutual TLS verification.

---

### Bank Server (`server.go`)

#### User Management

- **AuthenticateUser**: Validates user credentials.
- **SendBalanceReq**: Returns user account balance.
- **Transaction**: Processes payment requests (legacy method).

#### 2-Phase Commit Participation

- **Prepare**: Validates and reserves resources for transactions.
- **Commit**: Finalizes approved transactions.
- **Abort**: Cancels prepared transactions.

#### Data Persistence

- **User Data**: CSV-based storage of account information.
- **Pending Transactions**: JSON storage for recovery.
- **Balance Updates**: Real-time CSV updates.

#### Maintenance Features

- **Stale Transaction Cleanup**: Removes expired pending transactions.
- **Recovery Support**: Reloads pending transactions on startup.
- **Concurrent Access Control**: Thread-safe operations.

---

### Client (`client.go`)

#### User Interface

- Interactive menu system.
- Login/authentication flow.
- Balance inquiry functionality.
- Payment initiation.

#### Offline Capabilities

- **Offline Payment Queuing**: Stores payments when gateway is unavailable.
- **Automatic Retry**: Processes queued payments when connectivity returns (on the client side).
- **Connection Monitoring**: Continuously checks gateway availability.

#### Transaction Features

- **Unique Transaction IDs**: SHA-256 hash generation for idempotency.
- **Payment Details Input**: User-friendly payment form.
- **Status Reporting**: Real-time transaction feedback.

---

## Key Technical Features

### Distributed Transaction Management

- Two-phase commit protocol ensures ACID properties.
- Coordinator (gateway) manages transaction lifecycle.
- Participants (banks) maintain transaction state.

### Fault Tolerance

- Transaction logging for recovery.
- Timeout handling for stale transactions.
- Partial commit detection and handling.

### Security

- End-to-end TLS encryption.
- Certificate-based mutual authentication.
- Secure credential handling.

### Scalability

- Multiple bank server support.
- Concurrent client connections.
- Dynamic port allocation.

### Data Consistency

- Atomic transaction processing.
- Persistent state management.
- Recovery from system failures.

---

## Transaction Flow Example

1. Client initiates payment request.
2. Gateway validates client authentication.
3. Gateway starts 2PC:
        - Sends prepare requests to source and destination banks.
        - Banks validate and reserve resources.
        - Banks respond with prepare confirmation.
4. If both banks prepared successfully:
        - Gateway sends commit requests.
        - Banks execute the transaction.
        - Gateway logs successful completion.
5. If any bank fails to prepare:
        - Gateway sends abort requests.
        - Banks release reserved resources.
        - Transaction is rolled back.

---


## Summary

- **Client:** Initiates requests and handles responses.
- **Server:** Implements business logic and communicates with the payment gateway.
- **Payment Gateway:** Processes payments and interacts with external providers.
- **services.proto:** Defines the API contract for gRPC communication.

This system provides a robust, secure, and fault-tolerant distributed payment processing platform with support for offline operations and guaranteed transaction consistency.