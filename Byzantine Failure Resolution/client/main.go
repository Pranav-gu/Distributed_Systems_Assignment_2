package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	bft "Q4/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type general struct {
	bft.UnimplementedGeneralServiceServer
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

var n int
var t int
var order_global string
var min_port int
var max_port int
var dishonest_generals map[int]int = make(map[int]int)

func sendRegisterRequestCommander(client bft.CommanderServiceClient, port string) {
	ip := getMyIP()
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}

	fmt.Println("Register request to Master with ip = ", ip, " and port = ", port)
	req := &bft.RegisterRequest{Address: ip, Port: port}
	res, err := client.RegisterGeneral(context.Background(), req)
	if err != nil {
		log.Fatalf("Error registering worker: %v", err)
	}
	log.Printf("Response from RegisterServer port = %s: %v", port, res)
}

func (s *general) CommanderBroadcast(ctx context.Context, req *bft.Order) (*bft.OrderResponse, error) {
	var order string = req.GetOrder()
	var port string = req.GetPort()
	min_port = int(req.GetMinPort())
	max_port = int(req.GetMaxPort())
	res := true
	fmt.Println("order = ", order, "port = ", port)
	log.Println("Round 1: client port = ", port, " order = ", order, " min port = ", min_port, " max port = ", max_port)
	order_global = order
	go consensus(port)
	return &bft.OrderResponse{Success: res}, nil
}

func (s *general) CrossVerification(ctx context.Context, req *bft.Enquire) (*bft.EnquireResponse, error) {
	var recv_port string = req.GetRecvPort()
	res := true
	port_int, err := strconv.Atoi(recv_port)
	if err != nil {
		panic(err)
	}
	_, ok := dishonest_generals[port_int]
	if ok {
		var command int = rand.IntN(2)
		if command == 0 {
			return &bft.EnquireResponse{Success: res, Order: "Attack"}, nil
		} else {
			return &bft.EnquireResponse{Success: res, Order: "Retreat"}, nil
		}
	}
	return &bft.EnquireResponse{Success: res, Order: order_global}, nil
}

func consensus(port string) {
	time.Sleep(2 * time.Second)
	// form clients for all the nodes in advance and store them
	mp := make(map[int]bft.GeneralServiceClient)

	ip := getMyIP()
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}

	port_int, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	for i := min_port; i <= max_port; i++ {
		if i == port_int {
			continue
		}
		i_str := strconv.Itoa(i)
		conn, err := grpc.NewClient("localhost:"+i_str, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		client := bft.NewGeneralServiceClient(conn)
		mp[i] = client
	}
	for j := 0; j < t; j++ {
		count_attack := 0
		count_retreat := 0
		for i := min_port; i <= max_port; i++ {
			if i == port_int {
				continue
			}
			req := &bft.Enquire{Address: ip, SendPort: port, RecvPort: strconv.Itoa(i)}
			res, err := mp[i].CrossVerification(context.Background(), req)
			if err != nil {
				log.Fatalf("Error sending message: %v", err)
			}
			fmt.Printf("Round %s: Response received from Port %s by Port %s: %v\n", strconv.Itoa(j+2), strconv.Itoa(i), strconv.Itoa(port_int), res)
			if res.Order == "Attack" {
				count_attack = count_attack + 1
			} else if res.Order == "Retreat" {
				count_retreat = count_retreat + 1
			}
		}
		if count_attack >= count_retreat {
			order_global = "Attack"
		} else {
			order_global = "Retreat"
		}
		time.Sleep(time.Second*2)
	}
	fmt.Printf("Port = %s, Final Order = %s\n", port, order_global)
	time.Sleep(time.Second)
	os.Exit(0)
}

func main() {
	n1, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	t1, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	for i := 0; i < t1; i++ {
		x, err := strconv.Atoi(os.Args[3+i])
		if err != nil {
			panic(err)
		}
		if x == 0 { // x = 0 stands for dishonest commander
			continue
		}
		data, err := os.ReadFile("./client/current_port_server.txt")
		if err != nil {
			panic(err)
		}
		port, _ := strconv.Atoi(string(data))
		dishonest_generals[x-1+port]++
	}
	n = n1
	t = t1
	fmt.Println(t)

	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(n - 1)
	for i := 0; i < n-1; i++ {
		go func(i int) {
			defer wg.Done() // Signal that this goroutine is finished
			updatePortFile(&mu)
		}(i)
	}
	wg.Wait() // Wait for all goroutines to finish before exiting
}

// Function to update the port file safely and host gRPC Servers
func updatePortFile(mu *sync.Mutex) {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials())) // Commander Client
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := bft.NewCommanderServiceClient(conn)

	mu.Lock()
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
	mu.Unlock()

	sendRegisterRequestCommander(client, string(data))

	listener, err := net.Listen("tcp", ":"+string(data))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	bft.RegisterGeneralServiceServer(grpcServer, &general{})
	fmt.Println("Worker Server is running on port: ", string(data))

	// Serve requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
