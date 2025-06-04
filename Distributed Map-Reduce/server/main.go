package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	mapreduce "Q2/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type master struct {
	mapreduce.UnimplementedMasterServiceServer
	mu sync.Mutex
}

var conns []mapreduce.WorkerServiceClient
var numMappers int
var numReducers int

func (s *master) RegisterWorker(ctx context.Context, req *mapreduce.RegisterRequest) (*mapreduce.RegisterResponse, error) {
	var ip string = req.GetAddress()
	var port string = req.GetPort()
	// workers = append(workers, store{address: ip, port: port})
	res := true

	fmt.Println("ip = ", ip, " port = ", port)
	s.mu.Lock()
	conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client := mapreduce.NewWorkerServiceClient(conn)
	conns = append(conns, client)
	s.mu.Unlock()
	fmt.Printf("Socket address at which client is receiving requests is %s:%s\n", ip, port)
	return &mapreduce.RegisterResponse{Success: res}, nil
}

func DivideTaskToWorkers() {
	for {
		if len(conns) != 0 {
			break
		}
	}
	for i := 0; i < numMappers; i++ {
		for {
			if i < len(conns) {
				break
			}
		}
		req := &mapreduce.Task{Task: 1, Filename: "./dataset/file" + strconv.Itoa(i+1) + ".txt", Index: int32(i + 1)}
		res, err := conns[i].InvertedIndex(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in map phase inverted index task: %v", err)
		}
		log.Printf("Response from Map Phase Inverted index: %v", res)
	}

	for i := 0; i < numReducers; i++ {
		req := &mapreduce.Task{Task: 2, Filename: "./dataset/file" + strconv.Itoa(i+1) + ".txt", Index: int32(i + 1)}
		res, err := conns[i].InvertedIndex(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in reduce phase inverted index task: %v", err)
		}
		log.Printf("Response from Reduce Phase Inverted index: %v", res)
	}

	for i := 0; i < numMappers; i++ {
		req := &mapreduce.Task{Task: 1, Filename: "./dataset/file" + strconv.Itoa(i+1) + ".txt", Index: int32(i + 1)}
		res, err := conns[i].WordCount(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in map phase word count task: %v", err)
		}
		log.Printf("Response from Map Phase Word count : %v", res)
	}

	for i := 0; i < numReducers; i++ {
		req := &mapreduce.Task{Task: 2, Filename: "./dataset/file" + strconv.Itoa(i+1) + ".txt", Index: int32(i + 1)}
		res, err := conns[i].WordCount(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in reduce phase inverted index task: %v", err)
		}
		log.Printf("Response from Reduce Phase Word count : %v", res)
	}
	for i := 0; i < max(numReducers, numMappers); i++ {
		req := &mapreduce.Task{Task: 3, Filename: "./dataset/file" + strconv.Itoa(i+1) + ".txt", Index: int32(i + 1)}
		res, err := conns[i].WordCount(context.Background(), req)
		if err != nil {
			log.Fatalf("Error in reduce phase inverted index task: %v", err)
		}
		log.Printf("Worker = %v exited successfully : %v", i, res)
	}
	os.Exit(0)
}

func main() {
	mappers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	reducers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	numMappers = mappers
	numReducers = reducers

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	mapreduce.RegisterMasterServiceServer(grpcServer, &master{})
	log.Printf("Server is running on port: 8080")

	go DivideTaskToWorkers()
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
