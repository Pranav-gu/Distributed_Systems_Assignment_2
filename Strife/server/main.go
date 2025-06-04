package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	stripe "Q3/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type bank struct {
	stripe.UnimplementedBankServersServer
	mu             sync.Mutex
	pendingTx      map[string]*PendingTransaction
	pendingTxMutex sync.RWMutex
}

// PendingTransaction represents a transaction in the prepare phase
type PendingTransaction struct {
	TransactionID string
	Username      string
	Amount        int32
	Operation     string // "debit" or "credit"
	Timestamp     int64
}

var (
	crt = "./certificate/client_cert.pem"
	key = "./certificate/client_key.pem"
	ca  = "./certificate/ca_cert.pem"
)

var crt_server string
var key_server string
var ca_server string

// var users map[string]int = make(map[string]int)
var server_id int
var bankname string

type pair struct {
	balance  int
	password string
	bank     string
}

var users map[string]pair = make(map[string]pair)

func UsersDataLoad(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}
	for i := 1; i < len(records); i++ {
		balance, err := strconv.Atoi(records[i][1])
		if err != nil {
			panic(err)
		}
		users[records[i][0]] = pair{balance: balance, password: records[i][2], bank: records[i][3]}
	}
	return records
}

func UnaryLoggingInterceptor1(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("Incoming request: Method=%s", info.FullMethod)
	res, err := handler(ctx, req)
	if err != nil {
		log.Printf("Error: %v", status.Convert(err).Message())
	} else {
		log.Printf("Response: %v", res)
	}
	return res, err
}

func getMyIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func sendRegisterRequest(client stripe.PaymentGatewayServiceClient, port string) {
	ip := getMyIP()
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}
	fmt.Println("Register request to Master with ip = ", ip, " and port = ", port)
	fmt.Println("Enter name of bank to authenticate with the Payment Gateway: ")
	fmt.Scanln(&bankname)
	req := &stripe.RegisterRequestServer{Address: ip, Port: port, Bankname: bankname, Serverid: int32(server_id)}
	res, err := client.RegisterServer(context.Background(), req)
	if err != nil {
		log.Fatalf("Error registering worker: %v", err)
	}
	log.Printf("Response from RegisterServer port = %s: %v", port, res)
}

func (s *bank) AuthenticateUser(ctx context.Context, req *stripe.AuthReq) (*stripe.AuthRes, error) {
	var username string = req.GetUsername()
	var password string = req.GetPassword()

	_, ok := users[username]
	if ok && password == users[username].password && bankname == users[username].bank {
		return &stripe.AuthRes{Success: true, Message: "User Authentication successful"}, nil
	}
	if !ok || bankname != users[username].bank {
		fmt.Println("User not present in Database")
		return &stripe.AuthRes{Success: false, Message: "User not present in Database"}, nil
	}
	fmt.Println("Login insuccessful")
	return &stripe.AuthRes{Success: false, Message: "Login insuccessful"}, nil
}

func (s *bank) Transaction(ctx context.Context, req *stripe.PaymentReqGateway) (*stripe.PaymentResponse, error) {
	var username string = req.GetUsername()
	var amount int = int(req.GetAmount())

	val, ok := users[username]
	if !ok {
		fmt.Printf("User %s does not exist\n", username)
		return &stripe.PaymentResponse{Success: false, Message: "User does not exist"}, nil
	}

	if val.balance < amount {
		fmt.Printf("Insufficient bank balance of user %s", username)
		return &stripe.PaymentResponse{Success: false, Message: "Insufficient bank balance"}, nil
	}

	val.balance -= amount
	users[username] = val

	for key, value := range users {
		fmt.Println(key, value)
	}
	return &stripe.PaymentResponse{Success: false, Message: "Amount Deducted"}, nil
}

func (s *bank) SendBalanceReq(ctx context.Context, req *stripe.AccountDetailsReqGateway) (*stripe.AccountDetailsRespGateway, error) {
	var username string = req.GetUsername()
	val, ok := users[username]
	if !ok {
		fmt.Printf("User %s does not exist\n", username)
		return &stripe.AccountDetailsRespGateway{Success: false, Balance: -1}, nil
	}
	return &stripe.AccountDetailsRespGateway{Success: true, Balance: int32(val.balance)}, nil
}

// Prepare is the first phase of 2PC protocol
func (s *bank) Prepare(ctx context.Context, req *stripe.PrepareRequest) (*stripe.PrepareResponse, error) {
	username := req.GetUsername()
	amount := req.GetAmount()
	txID := req.GetTransactionId()
	operation := req.GetOperation()

	log.Printf("Prepare request received for transaction %s: %s operation of %d for user %s",
		txID, operation, amount, username)

	s.pendingTxMutex.Lock()
	defer s.pendingTxMutex.Unlock()

	// Initialize pendingTx map if it's nil
	if s.pendingTx == nil {
		s.pendingTx = make(map[string]*PendingTransaction)
	}

	// Check if user exists
	val, exists := users[username]
	if !exists {
		return &stripe.PrepareResponse{
			Ready:   false,
			Message: fmt.Sprintf("User %s does not exist", username),
		}, nil
	}

	// For debit operations, check if user has sufficient balance
	if operation == "debit" && int(amount) > val.balance {
		return &stripe.PrepareResponse{
			Ready:   false,
			Message: fmt.Sprintf("Insufficient balance for user %s", username),
		}, nil
	}

	// Store the pending transaction
	s.pendingTx[txID] = &PendingTransaction{
		TransactionID: txID,
		Username:      username,
		Amount:        amount,
		Operation:     operation,
		Timestamp:     time.Now().Unix(),
	}

	// Persist pending transactions to disk
	s.persistPendingTransactions()

	return &stripe.PrepareResponse{
		Ready:   true,
		Message: "Prepared successfully",
	}, nil
}

// Commit is the second phase of 2PC protocol - finalize the transaction
func (s *bank) Commit(ctx context.Context, req *stripe.CommitRequest) (*stripe.CommitResponse, error) {
	sender_bank := req.GetSenderBank()
	recv_bank := req.GetRecvBank()
	txID := req.GetTransactionId()
	step := req.GetStep()
	log.Printf("Commit request received for transaction %s", txID)

	s.pendingTxMutex.Lock()
	defer s.pendingTxMutex.Unlock()

	// Check if the transaction exists in the pending transactions
	pendingTx, exists := s.pendingTx[txID]
	fmt.Println("sender = ", sender_bank, " bank = ", bankname)
	// if !exists && sender_bank != bankname { // for the case when sender and receiver are from the same bank
	// 	return &stripe.CommitResponse{
	// 		Success: false,
	// 		Message: fmt.Sprintf("Transaction %s not found or already processed", txID),
	// 	}, nil
	// }

	if !exists { // for the case when sender and receiver are from the same bank
		return &stripe.CommitResponse{
			Success: false,
			Message: fmt.Sprintf("Transaction %s not found or already processed", txID),
		}, nil
	}

	// if !exists && sender_bank == bankname {
	// 	val, _ := users[pendingTx.Username]
	// 	val.balance += int(pendingTx.Amount)
	// 	users[pendingTx.Username] = val
	// 	log.Printf("Committed credit: %d to user %s, new balance: %d",
	// 		pendingTx.Amount, pendingTx.Username, users[pendingTx.Username].balance)
	// 	// Persist updated user balances
	// 	s.persistUserBalances()

	// 	return &stripe.CommitResponse{
	// 		Success: true,
	// 		Message: "Transaction committed successfully",
	// 	}, nil
	// }
	fmt.Println(pendingTx)
	// Apply the transaction based on the operation
	if pendingTx.Operation == "debit" {
		// users[pendingTx.Username] -= int(pendingTx.Amount)
		val, _ := users[pendingTx.Username]
		val.balance -= int(pendingTx.Amount)
		users[pendingTx.Username] = val
		log.Printf("Committed debit: %d from user %s, new balance: %d",
			pendingTx.Amount, pendingTx.Username, users[pendingTx.Username].balance)
	} else if pendingTx.Operation == "credit" {
		// users[pendingTx.Username] += int(pendingTx.Amount)
		val, _ := users[pendingTx.Username]
		val.balance += int(pendingTx.Amount)
		users[pendingTx.Username] = val
		log.Printf("Committed credit: %d to user %s, new balance: %d",
			pendingTx.Amount, pendingTx.Username, users[pendingTx.Username].balance)
	} else {
		return &stripe.CommitResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown operation: %s", pendingTx.Operation),
		}, nil
	}

	if sender_bank != recv_bank || (sender_bank == recv_bank && step == 2) {
		// Remove the transaction from pending
		delete(s.pendingTx, txID)
		// Persist the updated pending transactions
		s.persistPendingTransactions()
		// Persist updated user balances
	}
	s.persistUserBalances(pendingTx.Username)

	return &stripe.CommitResponse{
		Success: true,
		Message: "Transaction committed successfully",
	}, nil
}

// Abort is the second phase of 2PC protocol - cancel the transaction
func (s *bank) Abort(ctx context.Context, req *stripe.AbortRequest) (*stripe.AbortResponse, error) {
	txID := req.GetTransactionId()
	log.Printf("Abort request received for transaction %s", txID)

	s.pendingTxMutex.Lock()
	defer s.pendingTxMutex.Unlock()

	// Check if the transaction exists in the pending transactions
	_, exists := s.pendingTx[txID]
	if !exists {
		return &stripe.AbortResponse{
			Success: false,
			Message: fmt.Sprintf("Transaction %s not found or already processed", txID),
		}, nil
	}

	// Remove the transaction from pending
	delete(s.pendingTx, txID)

	// Persist the updated pending transactions
	s.persistPendingTransactions()

	return &stripe.AbortResponse{
		Success: true,
		Message: "Transaction aborted successfully",
	}, nil
}

// Persist pending transactions to disk for recovery"./server/dummyData.csv"
func (s *bank) persistPendingTransactions() {
	data, err := json.MarshalIndent(s.pendingTx, "", "  ")
	if err != nil {
		log.Printf("Error serializing pending transactions: %v", err)
		return
	}

	err = os.WriteFile("./server/pending_transactions.json", data, 0644)
	if err != nil {
		log.Printf("Error writing pending transactions to file: %v", err)
	}
}

// Load pending transactions from disk during startup
func (s *bank) loadPendingTransactions() {
	data, err := os.ReadFile("./server/pending_transactions.json")
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's fine
			s.pendingTx = make(map[string]*PendingTransaction)
			return
		}
		log.Printf("Error reading pending transactions file: %v", err)
		s.pendingTx = make(map[string]*PendingTransaction)
		return
	}

	err = json.Unmarshal(data, &s.pendingTx)
	if err != nil {
		log.Printf("Error deserializing pending transactions: %v", err)
		s.pendingTx = make(map[string]*PendingTransaction)
		return
	}

	log.Printf("Loaded %d pending transactions from disk", len(s.pendingTx))
}

// Persist user balances to CSV file
func (s *bank) persistUserBalances(username string) {
	f, err := os.Open("./server/dummydata.csv")
	filepath := "./server/dummydata.csv"
	if err != nil {
		log.Fatal("Unable to read input file "+filepath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filepath, err)
	}
	for i := 1; i < len(records); i++ {
		if records[i][0] == username {
			continue
		}
		balance, err := strconv.Atoi(records[i][1])
		if err != nil {
			panic(err)
		}
		users[records[i][0]] = pair{balance: balance, password: records[i][2], bank: records[i][3]}
	}

	file, err := os.Create("./server/dummydata.csv")
	if err != nil {
		log.Printf("Error creating user balances file: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"username", "balance", "password", "bank"}); err != nil {
		log.Printf("Error writing header to CSV: %v", err)
		return
	}

	// Write user data
	for username, balance := range users {
		fmt.Println("writing into csv: ", username, " ", balance.balance, " ", balance.password, " ", balance.bank)
		if err := writer.Write([]string{username, strconv.Itoa(balance.balance), balance.password, balance.bank}); err != nil {
			log.Printf("Error writing user data to CSV: %v", err)
			return
		}
	}
}

// Cleanup stale pending transactions
func (s *bank) cleanupStalePendingTransactions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.pendingTxMutex.Lock()
		now := time.Now().Unix()
		staleTimeout := int64(30 * 60) // 30 minutes

		for id, tx := range s.pendingTx {
			if now-tx.Timestamp > staleTimeout {
				log.Printf("Removing stale transaction %s (age: %d seconds)", id, now-tx.Timestamp)
				delete(s.pendingTx, id)
			}
		}

		s.persistPendingTransactions()
		s.pendingTxMutex.Unlock()
	}
}

func main() {
	id, err := strconv.Atoi(os.Args[1]) // basically used for identification of certificates
	if err != nil {
		panic(err)
	}
	server_id = id

	crt_server = "./certificate/server_cert_server" + strconv.Itoa(server_id) + ".pem"
	key_server = "./certificate/server_key_server" + strconv.Itoa(server_id) + ".pem"
	ca_server = "./certificate/ca_cert_server" + strconv.Itoa(server_id) + ".pem"

	// Load the client certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		panic(err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(ca)
	if err != nil {
		panic(err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic(ok)
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost:8080",
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(creds)) // Payment Gateway Client
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := stripe.NewPaymentGatewayServiceClient(conn)

	data, err := os.ReadFile("./server/current_port_server.txt")
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(string(data))
	port++

	file, err := os.Create("./server/current_port_server.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	var filePath string = "./server/dummydata.csv"
	UsersDataLoad(filePath)

	sendRegisterRequest(client, string(data))

	listener, err := net.Listen("tcp", ":"+string(data))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Load the client certificates from disk
	certificate1, err := tls.LoadX509KeyPair(crt_server, key_server)
	if err != nil {
		panic(err)
	}

	// Create a certificate pool from the certificate authority
	certPool1 := x509.NewCertPool()
	ca1, err := os.ReadFile(ca_server)
	if err != nil {
		panic(err)
	}

	// Append the certificates from the CA
	if ok := certPool1.AppendCertsFromPEM(ca1); !ok {
		panic(ok)
	}

	// Create the TLS credentials
	creds1 := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate1},
		ClientCAs:    certPool1,
	})

	// Initialize the bank server with 2PC support
	bankServer := &bank{
		pendingTx: make(map[string]*PendingTransaction),
	}

	// Load any pending transactions from disk (for recovery)
	bankServer.loadPendingTransactions()

	// Start a background goroutine to clean up stale transactions
	go bankServer.cleanupStalePendingTransactions()

	grpcServer := grpc.NewServer(grpc.Creds(creds1), grpc.UnaryInterceptor(UnaryLoggingInterceptor1))
	stripe.RegisterBankServersServer(grpcServer, bankServer)
	fmt.Printf("Server is running on port: %s\n", string(data))

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
