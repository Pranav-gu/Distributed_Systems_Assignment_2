package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	net1 "net"
	"os"
	"strconv"
	"time"

	loadbalance "Q1/protofiles"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type backendserver struct {
	loadbalance.UnimplementedBackendServer
}

var load float32 = 0

func getProcessCPUByPort(port int) float32 {
	// Get all network connections
	conns, err := net.Connections("inet")
	if err != nil {
		return 0
	}

	// Find process using the specified port
	var pid int32
	for _, conn := range conns {
		if int(conn.Laddr.Port) == port {
			pid = conn.Pid
			break
		}
	}

	if pid == 0 {
		return 0
	}

	// Get process info
	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0
	}

	// Get CPU usage over an interval
	ctx := context.Background()
	cpuBefore, err := proc.CPUPercentWithContext(ctx)
	if err != nil {
		return 0
	}

	// Wait for a short interval
	time.Sleep(1 * time.Second)

	// Get CPU usage after the interval
	cpuAfter, err := proc.CPUPercentWithContext(ctx)
	if err != nil {
		return 0
	}

	// Ensure the CPU usage is non-negative
	cpuUsage := cpuAfter - cpuBefore
	if cpuUsage < 0 {
		cpuUsage = 0
	}

	return float32(cpuUsage)
}

func sendRegisterRequest(server loadbalance.LoadBalancerClient, port string) {
	ip := getMyIP()
	fmt.Printf("Sending Register Request %s", ip)
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}

	fmt.Println("Register request to LB Server")
	req := &loadbalance.ServerInfo{Address: ip, Port: port}
	res, err := server.RegisterServer(context.Background(), req)
	if err != nil {
		log.Fatalf("Error registering server: %v", err)
	}
	log.Printf("Response from RegisterServer: %v", res)
}

func sendLoadStatus(server loadbalance.LoadBalancerClient, port string) {
	ip := getMyIP()
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}
	fmt.Printf("Sending Load Status Request %s", ip)
	fmt.Println("Load Status update sent to LB Server")
	port_int, _ := strconv.Atoi(port)
	load = getProcessCPUByPort(port_int)
	req := &loadbalance.ServerLoad{
		Address:     ip,
		Port:        port,
		CurrentLoad: load,
	}
	res, err := server.UpdateLoad(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling UpdateLoad: %v", err)
	}
	log.Printf("Response from UpdateLoad: %v", res)
}

func updateLoadLoop(client loadbalance.LoadBalancerClient, port string) {
	for {
		time.Sleep(5 * time.Second)
		sendLoadStatus(client, port)
	}
}

func getMyIP() string {
	addrs, err := net1.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net1.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

func (s *backendserver) ProcessRequest(ctx context.Context, req *loadbalance.ComputeTask) (*loadbalance.ComputeResponse, error) {
	var a int = int(req.GetA())
	fmt.Println(a)
	var sum int64 = 0
	for i := 1; i < int(math.Pow10(a)); i++ {
		sum += int64(i)
	}
	return &loadbalance.ComputeResponse{Result: sum}, nil
}

func main() {
	// Add command-line flag for port
	portFlag := flag.Int("port", 0, "The port to use for this server")
	flag.Parse()

	var port string
	if *portFlag != 0 {
		// Use the explicitly specified port
		port = strconv.Itoa(*portFlag)
	} else {
		// Fallback to the old behavior of reading from file
		data, err := os.ReadFile("./server/curr_port.txt")
		if err != nil {
			panic(err)
		}
		port = string(data)

		// Increment the port for the next server
		portNum, _ := strconv.Atoi(port)
		portNum++

		file, err := os.Create("./server/curr_port.txt")
		if err != nil {
			panic(err)
		}
		_, err = file.WriteString(strconv.Itoa(portNum))
		if err != nil {
			panic(err)
		}
		file.Close()
	}

	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := loadbalance.NewLoadBalancerClient(conn)
	sendRegisterRequest(client, port)

	go updateLoadLoop(client, port) // Run load update in a goroutine

	listener, err := net1.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	loadbalance.RegisterBackendServer(grpcServer, &backendserver{})
	log.Printf("Server is running on port: %s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
