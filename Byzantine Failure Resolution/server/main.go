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

	bft "Q4/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type commander struct {
	bft.UnimplementedCommanderServiceServer
	mu sync.Mutex
}

type generals_info struct {
	address string
	port    string
}

var conns []bft.GeneralServiceClient
var generals []generals_info

var n int
var t int
var max_port int = ^int(0) << (32 - 1)
var min_port int = int(^uint(0) >> 1)
var dishonest bool

func (s *commander) RegisterGeneral(ctx context.Context, req *bft.RegisterRequest) (*bft.RegisterResponse, error) {
	var ip string = req.GetAddress()
	var port string = req.GetPort()
	res := true

	fmt.Println("ip = ", ip, " port = ", port)
	s.mu.Lock()
	conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client := bft.NewGeneralServiceClient(conn)
	conns = append(conns, client)
	generals = append(generals, generals_info{address: ip, port: port})
	s.mu.Unlock()

	port_int, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	min_port = min(min_port, port_int)
	max_port = max(max_port, port_int)

	fmt.Printf("Socket address at which client is receiving requests is %s:%s\n", ip, port)
	return &bft.RegisterResponse{Success: res}, nil
}

func OrderGenerals() {
	for {
		if n-1 == len(conns) {
			break
		}
	}
	mp := make(map[int]string)
	mp[0] = "Attack"
	mp[1] = "Retreat"
	var command int = rand.IntN(2)
	for i := 0; i < n-1; i++ {
		fmt.Println("i = ", i)
		if dishonest {
			command = rand.IntN(2)
		}
		req := &bft.Order{Order: mp[command], Port: generals[i].port, MinPort: int32(min_port), MaxPort: int32(max_port)}
		res, err := conns[i].CommanderBroadcast(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in ordering generals task: %v", err)
		}
		log.Printf("Response from Order to General: %v", res)
	}
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
			dishonest = true
		}
	}

	n = n1
	t = t1
	fmt.Println(t)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	bft.RegisterCommanderServiceServer(grpcServer, &commander{})
	log.Printf("Server is running on port: 8080")

	go OrderGenerals()
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
