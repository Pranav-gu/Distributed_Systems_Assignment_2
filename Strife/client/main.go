package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	stripe "Q3/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	crt = "./certificate/client_cert.pem"
	key = "./certificate/client_key.pem"
	ca  = "./certificate/ca_cert.pem"
)
var username string
var password string
var bank string
var loggedIn bool = false

type transaction struct {
	Username         string
	Bank             string
	Amount           int32
	Receiverusername string
	Receiverbank     string
	Idempotent       string
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

	fmt.Printf("Enter Username for authentication with the Payment Gateway: ")
	fmt.Scanln(&username)
	fmt.Printf("Enter Password for authentication with the Payment Gateway: ")
	fmt.Scanln(&password)
	fmt.Printf("Enter Bankname for authentication with the Payment Gateway: ")
	fmt.Scanln(&bank)

	req := &stripe.RegisterRequestClient{Address: ip, Port: port, Username: username, Password: password, Bank: bank}
	res, err := client.RegisterClient(context.Background(), req)
	if err != nil {
		log.Fatalf("Error registering worker: %v", err)
	}
	log.Printf("Response from RegisterServer port = %s: %v", port, res)
	if !res.Success {
		log.Printf("User Credentials wrong!!!!")
	} else {
		loggedIn = true
		fmt.Println("Login Sucessful")
	}
}

func sendViewBalanceRequest(client stripe.PaymentGatewayServiceClient, port string) {
	ip := getMyIP()
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}

	var input string
	fmt.Printf("Enter the Bank of the User: \n")
	fmt.Scanln(&input)
	req := &stripe.AccountDetailsRequest{Address: ip, Port: port, Username: username, Bank: input}
	res, err := client.ViewBalance(context.Background(), req)
	if err != nil {
		log.Fatalf("Error sending View Balance request: %v", err)
	}
	log.Printf("Response from View Balance port: %v", res)
	if res.Message != "Ok" {
		log.Printf("Balance could not be viewed: %s", res.Message)
	}
	fmt.Printf("Balance: %d\n", res.Balance)
}

func checkOnline(client stripe.PaymentGatewayServiceClient, port string) {
	for {
		time.Sleep(10 * time.Second)
		if loggedIn {
			file, err := os.OpenFile("offline_payments.json", os.O_RDONLY, 0644)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			var remaining_transactions []transaction
			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				var payment transaction
				if err1 := json.Unmarshal(scanner.Bytes(), &payment); err1 != nil {
					fmt.Println("Error parsing JSON:", err)
					continue
				}
				if payment.Username != username {
					remaining_transactions = append(remaining_transactions, payment)
					continue
				}
				if !checkPortAlive("localhost", "8080", 2*time.Second) { // payment gateway down, queue payments in the offline_payment.json
					fmt.Println("Payment Gateway Down, Offline Payment scenario:")
					remaining_transactions = append(remaining_transactions, payment)
					continue
				}
				req := &stripe.PaymentDetails{Address: getMyIP(), Port: port, Username: payment.Username, Bank: payment.Bank,
					Amount: payment.Amount, Receiverusername: payment.Receiverusername, Receiverbank: payment.Receiverbank, Idemptotent: payment.Idempotent}
				res, err := client.MakePayment(context.Background(), req)
				if err != nil {
					log.Fatalf("Error making payment: %v", err)
				}
				log.Printf("Response from PaymentDetails rpc = %s: %v", port, res)
			}
			file.Close()
			file, _ = os.OpenFile("offline_payments.json", os.O_WRONLY|os.O_TRUNC, 0644)
			for i := 0; i < len(remaining_transactions); i++ {
				jsonData, _ := json.Marshal(remaining_transactions[i])
				file.Write(jsonData)
				file.WriteString("\n")
			}
			file.Close()
		}
	}
}

// checkPortAlive checks if a server is alive at the given address (host:port)
func checkPortAlive(host string, port string, timeout time.Duration) bool {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false // Port is not reachable
	}
	conn.Close()
	return true // Port is alive
}

func main() {
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

	data, err := os.ReadFile("./client/current_port_server.txt")
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(string(data))
	port++

	file, err := os.Create("./client/current_port_server.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	// sendRegisterRequest(client, string(data))

	for {
		var input string
		fmt.Printf("Do you want to exit (Y/N): ")
		fmt.Scanln(&input)
		go checkOnline(client, string(data))
		if input == "Y" {
			break
		}
		if !loggedIn {
			for {
				fmt.Printf("Do you want to send register request or make an offline payment (1/2): ")
				fmt.Scanln(&input)
				if input == "1" {
					sendRegisterRequest(client, string(data))
					if !loggedIn {
						continue
					}
					break
				} else if input == "2" {
					break
				} else {
					fmt.Printf("Please Enter a valid input.\n\n")
				}
			}
			if loggedIn {
				continue
			}
		} else {
			fmt.Printf("Do you want to view your Bank Balance(Y/N): ")
			fmt.Scanln(&input)
			if input == "Y" {
				sendViewBalanceRequest(client, string(data))
				continue
			}
		}

		var receiver_user string
		var receiver_bank string
		var amount int
		if !loggedIn {
			fmt.Printf("Enter your username: ")
			fmt.Scanln(&username)
		}
		fmt.Printf("Enter the bank via which you want to make the payment: ")
		fmt.Scanln(&bank)
		fmt.Printf("Enter the username of the receiver: ")
		fmt.Scanln(&receiver_user)
		fmt.Printf("Enter the bank of the receiver:  ")
		fmt.Scanln(&receiver_bank)
		fmt.Printf("Enter the amount to be made as the payment to the receiver: ")
		fmt.Scanln(&amount)

		id := string(data) + username + bank + strconv.Itoa(amount) + receiver_user + receiver_bank + strconv.Itoa(int(time.Now().Unix()))
		unique_id := sha256.Sum256([]byte(id))
		fmt.Println(unique_id)

		if !loggedIn {
			idempotent := hex.EncodeToString(unique_id[:])
			payment := transaction{
				Username:         username,
				Bank:             bank,
				Amount:           int32(amount),
				Receiverusername: receiver_user,
				Receiverbank:     receiver_bank,
				Idempotent:       idempotent,
			}
			file, err := os.OpenFile("offline_payments.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				panic(err)
			}
			fmt.Println(payment)
			jsonData, _ := json.Marshal(payment)
			file.Write(jsonData)
			file.WriteString("\n")
			file.Close()
			continue
		}

		// check if the payment gateway is down or not
		if !checkPortAlive("localhost", "8080", 2*time.Second) { // payment gateway down, queue payments in the pending_transactions.json
			fmt.Println("Payment Gateway Down, Offline Payment scenario:")

			idempotent := hex.EncodeToString(unique_id[:])
			payment := transaction{
				Username:         username,
				Bank:             bank,
				Amount:           int32(amount),
				Receiverusername: receiver_user,
				Receiverbank:     receiver_bank,
				Idempotent:       idempotent,
			}

			fmt.Println(payment)
			file, _ = os.OpenFile("offline_payments.json", os.O_WRONLY, 0644)
			jsonData, _ := json.Marshal(payment)
			file.Write(jsonData)
			file.WriteString("\n")
			file.Close()
			continue
		}

		// payment gateway is up and running
		req := &stripe.PaymentDetails{Address: getMyIP(), Port: string(data), Username: username, Bank: bank, Amount: int32(amount), Receiverusername: receiver_user, Receiverbank: receiver_bank, Idemptotent: hex.EncodeToString(unique_id[:])}
		res, err := client.MakePayment(context.Background(), req)
		if err != nil {
			log.Fatalf("Error making payment: %v", err)
		}
		log.Printf("Response from PaymentDetails rpc %v", res)
	}
}
