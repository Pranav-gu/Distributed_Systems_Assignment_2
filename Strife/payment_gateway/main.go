package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	stripe "Q3/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Transaction status constants
const (
	StatusPrepared  = "prepared"
	StatusCommitted = "committed"
	StatusAborted   = "aborted"
)

// TransactionLog represents a record of a distributed transaction
type TransactionLog struct {
	TransactionID  string
	SourceUser     string
	SourceBank     string
	SourcePrepared bool
	DestUser       string
	DestBank       string
	DestPrepared   bool
	Amount         int32
	Status         string
	Timestamp      time.Time
}

type gateway struct {
	stripe.UnimplementedPaymentGatewayServiceServer
	mu              sync.Mutex
	mu1             sync.Mutex
	transactionLogs map[string]*TransactionLog
	txLogMu         sync.RWMutex
}

var (
	crt = "./certificate/server_cert.pem"
	key = "./certificate/server_key.pem"
	ca  = "./certificate/ca_cert.pem"
)

// type pair struct {
// 	password string
// 	bank     string
// }

// var users map[string]pair = make(map[string]pair)
var users_loggedin map[string]int = make(map[string]int)
var bankservers_conns map[string]stripe.BankServersClient = make(map[string]stripe.BankServersClient)

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func UnaryLoggingInterceptor(
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

func (s *gateway) RegisterClient(ctx context.Context, req *stripe.RegisterRequestClient) (*stripe.RegisterResponse, error) {
	var ip string = req.GetAddress()
	var port string = req.GetPort()
	var username string = req.GetUsername()
	var password string = req.GetPassword()
	var bank string = req.GetBank()

	req1 := &stripe.AuthReq{Username: username, Password: password}
	res, err := bankservers_conns[bank].AuthenticateUser(context.Background(), req1)
	if err != nil {
		log.Fatalf("Error sending payment request to Bank Server: %v", err)
	}
	log.Printf("Response from Bank Server = %s: %v", port, res)

	if res.Success {
		users_loggedin[port] = 1
		fmt.Println("ip = ", ip, " port = ", port, " username = ", username, " password = ", password)
		return &stripe.RegisterResponse{Success: true}, nil
	}
	if res.Message == "User not present in Database" {
		fmt.Println("User not present in Database")
		return &stripe.RegisterResponse{Success: false}, nil
	}
	fmt.Println("Login insuccessful")
	return &stripe.RegisterResponse{Success: false}, nil
}

func (s *gateway) RegisterServer(ctx context.Context, req *stripe.RegisterRequestServer) (*stripe.RegisterResponse, error) {
	s.mu.Lock()

	var ip string = req.GetAddress()
	var port string = req.GetPort()
	var bankname string = req.GetBankname()
	var server_id int = int(req.GetServerid())

	app := "./certificate/script.sh"
	arg0 := port
	arg1 := strconv.Itoa(server_id)
	cmd := exec.Command(app, arg0, arg1)
	fmt.Println(cmd)
	stdout, err := cmd.Output()

	if err != nil {
		panic(err)
	}
	fmt.Println(string(stdout))

	items, _ := os.ReadDir(".")
	for _, item := range items {
		if !item.IsDir() && !(item.Name()[0] == 'g' && item.Name()[1] == 'o') && item.Name() != "pending_transactions.json" {
			app = "mv"
			arg0 = item.Name()
			arg1 = "./certificate"
			cmd = exec.Command(app, arg0, arg1)
			fmt.Println(cmd)
			stdout, err := cmd.Output()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(stdout))
		}
	}

	var (
		crt1 = "./certificate/client_cert_server" + strconv.Itoa(server_id) + ".pem"
		key1 = "./certificate/client_key_server" + strconv.Itoa(server_id) + ".pem"
		ca1  = "./certificate/ca_cert_server" + strconv.Itoa(server_id) + ".pem"
	)

	// Load the client certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt1, key1)
	if err != nil {
		panic(err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(ca1)
	if err != nil {
		panic(err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic(ok)
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost:" + port,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	// conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials())) // Bank Servers Client
	conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(creds)) // Bank Servers Client
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client := stripe.NewBankServersClient(conn)
	bankservers_conns[bankname] = client
	s.mu.Unlock()

	res := true
	fmt.Println("ip = ", ip, " port = ", port, " bank = ", bankname)
	return &stripe.RegisterResponse{Success: res}, nil
}

// Implement the 2PC version of MakePayment
func (s *gateway) MakePayment(ctx context.Context, req *stripe.PaymentDetails) (*stripe.PaymentResponse, error) {
	var port string = req.GetPort()
	var username string = req.GetUsername()
	var amount int = int(req.GetAmount())
	var bank string = req.GetBank()
	var recv_username string = req.GetReceiverusername()
	var recv_bank string = req.GetReceiverbank()
	var unique_id string = req.GetIdemptotent()

	_, ok := users_loggedin[port]
	if !ok {
		fmt.Println("User not registered with the Payment Gateway")
		return &stripe.PaymentResponse{Success: false, Message: "User not Authenticated"}, nil
	}

	// Check for idempotency
	s.mu1.Lock()
	file, err := os.OpenFile("./payment_gateway/transactions.txt", os.O_RDONLY, 0777)
	if err != nil {
		fmt.Println("1")
		s.mu1.Unlock()
		panic(err)
	}

	r := bufio.NewReader(file)
	line, err1 := Readln(r)
	fmt.Println("err1 = ", err1)
	for err1 == nil {
		fmt.Println(line, unique_id)
		if line == unique_id {
			file.Close()
			s.mu1.Unlock()
			return &stripe.PaymentResponse{Success: false, Message: "Failed due to maintenance of Idempotency"}, nil
		}
		line, err1 = Readln(r)
	}
	file.Close()

	file, err = os.OpenFile("./payment_gateway/transactions.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("2")
		s.mu1.Unlock()
		panic(err)
	}

	_, err = file.WriteString(unique_id + "\n")
	if err != nil {
		fmt.Println("3")
		file.Close()
		s.mu1.Unlock()
		panic(err)
	}
	file.Close()
	s.mu1.Unlock()

	// Create a new transaction log entry
	txLog := &TransactionLog{
		TransactionID:  unique_id,
		SourceUser:     username,
		SourceBank:     bank,
		SourcePrepared: false,
		DestUser:       recv_username,
		DestBank:       recv_bank,
		DestPrepared:   false,
		Amount:         int32(amount),
		Status:         "preparing",
		Timestamp:      time.Now(),
	}

	// Store the transaction log
	s.txLogMu.Lock()
	s.transactionLogs[unique_id] = txLog
	s.txLogMu.Unlock()

	// Log the start of 2PC
	log.Printf("Starting 2PC for transaction %s", unique_id)

	// PHASE 1: PREPARE
	// Check if both banks are registered
	sourceClient, sourceOk := bankservers_conns[bank]
	if !sourceOk {
		return &stripe.PaymentResponse{Success: false, Message: "Source bank not registered"}, nil
	}

	destClient, destOk := bankservers_conns[recv_bank]
	if !destOk {
		return &stripe.PaymentResponse{Success: false, Message: "Destination bank not registered"}, nil
	}

	// Prepare the source bank (debit)
	prepareSourceReq := &stripe.PrepareRequest{
		Username:      username,
		Amount:        int32(amount),
		TransactionId: unique_id,
		Operation:     "debit",
	}
	prepareSourceResp, err := sourceClient.Prepare(ctx, prepareSourceReq)
	if err != nil {
		log.Printf("Error preparing source bank: %v", err)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: "Error preparing source bank"}, nil
	}

	if !prepareSourceResp.Ready {
		log.Printf("Source bank not ready: %s", prepareSourceResp.Message)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: prepareSourceResp.Message}, nil
	}

	// Update transaction log
	s.txLogMu.Lock()
	s.transactionLogs[unique_id].SourcePrepared = true
	s.txLogMu.Unlock()

	// Prepare the destination bank (credit)
	prepareDestReq := &stripe.PrepareRequest{
		Username:      recv_username,
		Amount:        int32(amount),
		TransactionId: unique_id,
		Operation:     "credit",
	}
	prepareDestResp, err := destClient.Prepare(ctx, prepareDestReq)
	if err != nil {
		log.Printf("Error preparing destination bank: %v", err)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: "Error preparing destination bank"}, nil
	}

	if !prepareDestResp.Ready {
		log.Printf("Destination bank not ready: %s", prepareDestResp.Message)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: prepareDestResp.Message}, nil
	}

	// Update transaction log
	s.txLogMu.Lock()
	s.transactionLogs[unique_id].DestPrepared = true
	s.transactionLogs[unique_id].Status = StatusPrepared
	s.txLogMu.Unlock()

	// PHASE 2: COMMIT
	// Both banks are prepared, proceed to commit
	commitSourceReq := &stripe.CommitRequest{
		SenderBank:    bank,
		RecvBank:      recv_bank,
		TransactionId: unique_id,
		Step:          1,
	}
	commitSourceResp, err := sourceClient.Commit(ctx, commitSourceReq)
	if err != nil {
		log.Printf("Error committing to source bank: %v", err)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: "Error committing to source bank"}, nil
	}

	if !commitSourceResp.Success {
		log.Printf("Source bank commit failed: %s", commitSourceResp.Message)
		s.abortTransaction(ctx, unique_id)
		return &stripe.PaymentResponse{Success: false, Message: commitSourceResp.Message}, nil
	}

	commitDestReq := &stripe.CommitRequest{
		SenderBank:    bank,
		RecvBank:      recv_bank,
		TransactionId: unique_id,
		Step:          2,
	}
	commitDestResp, err := destClient.Commit(ctx, commitDestReq)
	if err != nil {
		log.Printf("Error committing to destination bank: %v", err)
		// At this point, the source has committed but destination failed
		// This is a critical error that needs manual resolution
		// In a real system, this should be handled with a recovery mechanism
		s.txLogMu.Lock()
		s.transactionLogs[unique_id].Status = "partial-commit"
		s.txLogMu.Unlock()

		// Persist the transaction logs to disk for recovery
		s.persistTransactionLogs()

		return &stripe.PaymentResponse{Success: false, Message: "Critical error: Partial commit occurred"}, nil
	}

	if !commitDestResp.Success {
		log.Printf("Destination bank commit failed: %s", commitDestResp.Message)
		// Similar critical error as above
		s.txLogMu.Lock()
		s.transactionLogs[unique_id].Status = "partial-commit"
		s.txLogMu.Unlock()

		// Persist the transaction logs to disk for recovery
		s.persistTransactionLogs()

		return &stripe.PaymentResponse{Success: false, Message: "Critical error: Partial commit occurred"}, nil
	}

	// Update transaction log - everything successful
	s.txLogMu.Lock()
	s.transactionLogs[unique_id].Status = StatusCommitted
	s.txLogMu.Unlock()

	// Persist the transaction logs
	s.persistTransactionLogs()

	log.Printf("Successfully completed 2PC transaction %s", unique_id)
	return &stripe.PaymentResponse{Success: true, Message: "Payment successful"}, nil
}

// Abort a transaction that's in progress
func (s *gateway) abortTransaction(ctx context.Context, txID string) {
	log.Printf("Aborting transaction %s", txID)

	s.txLogMu.RLock()
	txLog, exists := s.transactionLogs[txID]
	s.txLogMu.RUnlock()

	if !exists {
		log.Printf("Transaction %s not found for abort", txID)
		return
	}

	// Abort at source bank if it was prepared
	if txLog.SourcePrepared {
		sourceClient, ok := bankservers_conns[txLog.SourceBank]
		if ok {
			abortReq := &stripe.AbortRequest{
				TransactionId: txID,
			}
			_, err := sourceClient.Abort(ctx, abortReq)
			if err != nil {
				log.Printf("Error aborting at source bank: %v", err)
				// Continue anyway to try aborting at destination
			}
		}
	}

	// Abort at destination bank if it was prepared
	if txLog.DestPrepared {
		destClient, ok := bankservers_conns[txLog.DestBank]
		if ok {
			abortReq := &stripe.AbortRequest{
				TransactionId: txID,
			}
			_, err := destClient.Abort(ctx, abortReq)
			if err != nil {
				log.Printf("Error aborting at destination bank: %v", err)
			}
		}
	}

	// Update transaction log
	s.txLogMu.Lock()
	s.transactionLogs[txID].Status = StatusAborted
	s.txLogMu.Unlock()

	// Persist transaction logs for recovery
	s.persistTransactionLogs()

	log.Printf("Transaction %s aborted", txID)
}

// Persist transaction logs to disk for recovery
func (s *gateway) persistTransactionLogs() {
	s.txLogMu.RLock()
	defer s.txLogMu.RUnlock()

	// Create a copy of the transaction logs for serialization
	logsCopy := make(map[string]TransactionLog)
	for id, log := range s.transactionLogs {
		logsCopy[id] = *log
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(logsCopy, "", "  ")
	if err != nil {
		log.Printf("Error serializing transaction logs: %v", err)
		return
	}

	// Write to file
	err = os.WriteFile("./payment_gateway/transaction_logs.json", data, 0644)
	if err != nil {
		log.Printf("Error writing transaction logs to file: %v", err)
	}
}

// Load transaction logs from disk during startup
func (s *gateway) loadTransactionLogs() {
	data, err := os.ReadFile("./payment_gateway/transaction_logs.json")
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's fine
			return
		}
		log.Printf("Error reading transaction logs file: %v", err)
		return
	}

	var logs map[string]TransactionLog
	err = json.Unmarshal(data, &logs)
	if err != nil {
		log.Printf("Error deserializing transaction logs: %v", err)
		return
	}

	// Convert to pointers and update the transaction logs map
	s.txLogMu.Lock()
	for id, log := range logs {
		logCopy := log // Copy to avoid reference issues
		s.transactionLogs[id] = &logCopy
	}
	s.txLogMu.Unlock()

	log.Printf("Loaded %d transaction logs from disk", len(logs))
}

// Recovery process for partial commits (would run during startup)
func (s *gateway) recoverPartialCommits(ctx context.Context) {
	s.txLogMu.RLock()
	var partialCommits []string
	for id, log := range s.transactionLogs {
		if log.Status == "partial-commit" {
			partialCommits = append(partialCommits, id)
		}
	}
	s.txLogMu.RUnlock()

	for _, id := range partialCommits {
		log.Printf("Attempting recovery for partial commit %s", id)
		// Implementation would depend on your specific recovery policy
		// Options include:
		// 1. Manual intervention (admin UI)
		// 2. Automatic rollback if possible
		// 3. Automatic completion if possible
	}
}

// Implement the rest of the methods
func (s *gateway) ViewBalance(ctx context.Context, req *stripe.AccountDetailsRequest) (*stripe.AccountDetailsResponse, error) {
	var ip string = req.GetAddress()
	var port string = req.GetPort()
	var username string = req.GetUsername()
	var bankname string = req.GetBank()
	_, ok := users_loggedin[port]
	if !ok {
		fmt.Printf("User with ip %s not registered with the Payment Gateway\n", ip)
		return &stripe.AccountDetailsResponse{Success: false, Message: "User not Authenticated", Balance: -1}, nil
	}
	req1 := &stripe.AccountDetailsReqGateway{Username: username}
	res, err := bankservers_conns[bankname].SendBalanceReq(context.Background(), req1)
	if err != nil {
		log.Fatalf("Error sending view balance request: %v", err)
	}
	log.Printf("Response from SendBalanceReq: %v", res)
	if !res.Success {
		return &stripe.AccountDetailsResponse{Success: false, Message: "User does not exist in the specified bank", Balance: -1}, nil
	}
	return &stripe.AccountDetailsResponse{Success: false, Message: "Ok", Balance: res.Balance}, nil
}

func main() {
	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		panic(err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ca)
	if err != nil {
		panic(err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic(ok)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// Initialize gateway with transaction logs map
	gw := &gateway{
		transactionLogs: make(map[string]*TransactionLog),
	}

	// Load transaction logs from disk
	gw.loadTransactionLogs()

	// Start recovery process for partial commits
	go gw.recoverPartialCommits(context.Background())

	grpcServer := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(UnaryLoggingInterceptor))
	stripe.RegisterPaymentGatewayServiceServer(grpcServer, gw)
	log.Printf("Server is running on port: 8080")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
