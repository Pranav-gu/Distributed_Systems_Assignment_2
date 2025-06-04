package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"

	loadbalance "Q1/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func sendRequestBackendServer(client loadbalance.BackendClient) {
	fmt.Println("Sending request to Backend server")
	var a int = rand.IntN(9)
	fmt.Println(a)
	req := &loadbalance.ComputeTask{A: int32(a)}
	res, err := client.ProcessRequest(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling ProcessRequest: %v", err)
	}
	log.Printf("Response from ProcessRequest: %v", res)
}

type ret struct {
	address string
	port    string
}

func sendRequestLBServer(client loadbalance.LoadBalancerClient) ret {
	req := &loadbalance.ClientRequest{}
	res, err := client.GetBestServer(context.Background(), req)
	if err != nil {
		log.Fatalf("Error getting best server: %v", err)
	}
	log.Printf("Best backend server: %v:%v", res.Address, res.Port)
	return ret{address: res.Address, port: res.Port}
}

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	lbClient := loadbalance.NewLoadBalancerClient(conn)

	for {
		var first string
		fmt.Printf("Do you want to continue (y/n): ")
		fmt.Scanln(&first)
		if first == "n" {
			break
		}

		server := sendRequestLBServer(lbClient)
		fmt.Println(server.address, server.port)

		conn1, err := grpc.NewClient("localhost:"+server.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to backend: %v", err)
		}
		defer conn1.Close()

		client := loadbalance.NewBackendClient(conn1)
		sendRequestBackendServer(client)
	}
}
