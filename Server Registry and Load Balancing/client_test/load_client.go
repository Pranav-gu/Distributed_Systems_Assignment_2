package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
	"time"

	loadbalance "Q1/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestMetric struct {
	ClientID      int
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	TargetServer  string
	Success       bool
	ResponseValue int64
}

func sendRequestBackendServer(client loadbalance.BackendClient, clientID int) RequestMetric {
	metric := RequestMetric{
		ClientID:  clientID,
		StartTime: time.Now(),
		Success:   false,
	}

	a := rand.IntN(9) + 1 // Generate a random number between 1-9
	req := &loadbalance.ComputeTask{A: int32(a)}

	res, err := client.ProcessRequest(context.Background(), req)
	metric.EndTime = time.Now()
	metric.Duration = metric.EndTime.Sub(metric.StartTime)

	if err != nil {
		log.Printf("Client %d: Error while calling ProcessRequest: %v", clientID, err)
		return metric
	}

	metric.Success = true
	metric.ResponseValue = res.GetResult()
	return metric
}

type ServerInfo struct {
	Address string
	Port    string
}

func sendRequestLBServer(client loadbalance.LoadBalancerClient, clientID int) (ServerInfo, error) {
	req := &loadbalance.ClientRequest{}
	res, err := client.GetBestServer(context.Background(), req)
	if err != nil {
		log.Printf("Client %d: Error getting best server: %v", clientID, err)
		return ServerInfo{}, err
	}

	return ServerInfo{Address: res.Address, Port: res.Port}, nil
}

func runClient(wg *sync.WaitGroup, clientID int, numRequests int, lbAddress string, resultChan chan<- RequestMetric) {
	defer wg.Done()

	// Connect to load balancer
	conn, err := grpc.Dial(lbAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Client %d: Failed to connect to load balancer: %v", clientID, err)
		return
	}
	defer conn.Close()

	lbClient := loadbalance.NewLoadBalancerClient(conn)

	// Send multiple requests
	for i := 0; i < numRequests; i++ {
		// Get best server from load balancer
		server, err := sendRequestLBServer(lbClient, clientID)
		if err != nil {
			continue
		}

		// Connect to the backend server
		serverAddress := server.Address + ":" + server.Port
		if server.Address == "localhost" || server.Address == "127.0.0.1" {
			serverAddress = "localhost:" + server.Port
		}

		backendConn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Client %d: Failed to connect to backend server %s: %v", clientID, serverAddress, err)
			continue
		}

		backendClient := loadbalance.NewBackendClient(backendConn)

		// Send request and record metrics
		metric := sendRequestBackendServer(backendClient, clientID)
		metric.TargetServer = serverAddress

		// Send metrics to channel
		resultChan <- metric

		backendConn.Close()

		// Add a small random delay to simulate real-world scenarios
		time.Sleep(time.Duration(rand.IntN(50)) * time.Millisecond)
	}
}

func writeResultsToCSV(metrics []RequestMetric, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ClientID", "StartTime", "EndTime", "Duration_ms", "TargetServer", "Success", "ResponseValue"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, metric := range metrics {
		row := []string{
			strconv.Itoa(metric.ClientID),
			metric.StartTime.Format(time.RFC3339Nano),
			metric.EndTime.Format(time.RFC3339Nano),
			strconv.FormatFloat(float64(metric.Duration.Milliseconds()), 'f', 2, 64),
			metric.TargetServer,
			strconv.FormatBool(metric.Success),
			strconv.FormatInt(metric.ResponseValue, 10),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Configuration parameters
	var (
		numClients  = 100 // Number of concurrent clients
		numRequests = 10  // Requests per client
		lbAddress   = "localhost:8080"
		outputFile  = "load_test_results.csv"
	)

	// Allow command-line parameter overrides
	if len(os.Args) > 1 {
		if val, err := strconv.Atoi(os.Args[1]); err == nil {
			numClients = val
		}
	}
	if len(os.Args) > 2 {
		if val, err := strconv.Atoi(os.Args[2]); err == nil {
			numRequests = val
		}
	}
	if len(os.Args) > 3 {
		lbAddress = os.Args[3]
	}

	// Channel to collect metrics
	resultChan := make(chan RequestMetric, numClients*numRequests)

	// Start clients
	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Printf("Starting load test with %d clients, each making %d requests...\n", numClients, numRequests)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go runClient(&wg, i, numRequests, lbAddress, resultChan)
	}

	// Wait in a goroutine so we can show progress
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var metrics []RequestMetric
	for metric := range resultChan {
		metrics = append(metrics, metric)
	}

	// Calculate test duration
	testDuration := time.Since(startTime)

	// Show summary
	fmt.Printf("\nTest completed in %.2f seconds\n", testDuration.Seconds())
	fmt.Printf("Total requests: %d\n", len(metrics))

	// Count successful requests
	successCount := 0
	for _, m := range metrics {
		if m.Success {
			successCount++
		}
	}

	fmt.Printf("Successful requests: %d (%.2f%%)\n", successCount, float64(successCount)/float64(len(metrics))*100)

	// Save results to CSV
	if err := writeResultsToCSV(metrics, outputFile); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}

	fmt.Printf("Results saved to %s\n", outputFile)
	fmt.Println("Run the analysis tool to generate graphs and detailed metrics.")
}
