package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"

	net1 "net"
	"os"
	"strings"
	"time"

	loadbalance "Q1/protofiles"

	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc"
)

type servers_info struct {
	address string
	port    string
}

var (
	servers []servers_info                // For pick_first & round_robin
	rrIndex int                           // Round-robin counter
	policy  string         = "least_load" // Default policy
)

var last_update map[string]int = make(map[string]int)

type loadbalanceserver struct {
	loadbalance.UnimplementedLoadBalancerServer
}

var etcdClient *clientv3.Client

func initEtcd() {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // Change if etcd is hosted elsewhere
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	data, err := etcdClient.Delete(context.Background(), "servers/", clientv3.WithPrefix())
	fmt.Println(data)
	if err != nil {
		log.Fatalf("Failed to delete entries from etcd: %v", err)
	}
}

func (s *loadbalanceserver) GetBestServer(ctx context.Context, req *loadbalance.ClientRequest) (*loadbalance.ServerResponse, error) {
	resp, err := etcdClient.Get(context.Background(), "servers/", clientv3.WithPrefix())
	if err != nil {
		log.Printf("Failed to fetch servers from etcd: %v", err)
		return nil, err
	}

	var serverList []string
	serverLoad := make(map[string]float64) // To store load values for least_load

	for _, kv := range resp.Kvs {
		addr := string(kv.Key[len("servers/"):])
		load, _ := strconv.ParseFloat(string(kv.Value), 64)
		serverList = append(serverList, addr)
		serverLoad[addr] = load
	}

	if len(serverList) == 0 {
		return nil, fmt.Errorf("no available servers")
	}

	switch policy {
	case "pick_first":
		parts := strings.Split(serverList[0], ":")
		return &loadbalance.ServerResponse{Address: parts[0], Port: parts[1]}, nil

	case "round_robin":
		server := serverList[rrIndex%len(serverList)]
		rrIndex = (rrIndex + 1) % len(serverList)
		fmt.Println(rrIndex)
		fmt.Printf("Server selected %s\n", server)
		parts := strings.Split(serverList[rrIndex], ":")
		return &loadbalance.ServerResponse{Address: parts[0], Port: parts[1]}, nil

	case "least_load":
		var bestServer string
		minLoad := math.MaxFloat64

		for server, load := range serverLoad {
			fmt.Println("server: ", server, " load: ", load)
			if load < minLoad {
				minLoad = load
				bestServer = server
			}
		}
		if bestServer == "" {
			return nil, fmt.Errorf("no available servers")
		}
		parts := strings.Split(bestServer, ":")
		return &loadbalance.ServerResponse{Address: parts[0], Port: parts[1]}, nil

	default:
		return nil, fmt.Errorf("invalid load balancing policy")
	}
}

func (s *loadbalanceserver) RegisterServer(ctx context.Context, req *loadbalance.ServerInfo) (*loadbalance.RegisterResponse, error) {
	ip_addr := req.GetAddress()
	port := req.GetPort()

	// Store in memory for pick_first & round_robin
	servers = append(servers, servers_info{address: ip_addr, port: port})

	// Store server details in etcd
	var store string = "servers/" + ip_addr + ":" + port
	_, err := etcdClient.Put(context.Background(), store, "0") // "0" represents initial load

	if err != nil {
		log.Printf("Failed to register server in etcd: %v", err)
		return &loadbalance.RegisterResponse{Success: false}, err
	}

	fmt.Printf("Server registered: %s:%s\n", ip_addr, port)
	return &loadbalance.RegisterResponse{Success: true}, nil
}

func (s *loadbalanceserver) UpdateLoad(ctx context.Context, req *loadbalance.ServerLoad) (*loadbalance.UpdateResponse, error) {
	ip_addr := req.GetAddress()
	port := req.GetPort()
	load := req.GetCurrentLoad()

	// Update load in etcd
	last_update[ip_addr+":"+port] = int(time.Now().Unix())
	_, err := etcdClient.Put(context.Background(), "servers/"+ip_addr+":"+port, fmt.Sprintf("%f", load))
	if err != nil {
		log.Printf("Failed to update load in etcd: %v", err)
		return &loadbalance.UpdateResponse{Success: false}, err
	}

	log.Printf("Load updated for %s:%s = %d", ip_addr, port, load)
	return &loadbalance.UpdateResponse{Success: true}, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func printAliveServers() {
	for {
		time.Sleep(5 * time.Second)
		resp, err := etcdClient.Get(context.Background(), "servers/", clientv3.WithPrefix())
		if err != nil {
			log.Printf("Failed to fetch servers from etcd: %v", err)
		}
		var curr_time int = int(time.Now().Unix())
		for _, kv := range resp.Kvs {
			addr := string(kv.Key[len("servers/"):])
			if val, exists := last_update[addr]; exists {
				if abs(curr_time-val) >= 8 {
					log.Printf("Server %s is unresponsive for 8+ seconds, removing...", addr)
					delete(last_update, addr)
					// Remove from etcd
					_, err := etcdClient.Delete(context.Background(), "servers/"+addr)
					if err != nil {
						log.Printf("Failed to remove server %s from etcd: %v", addr, err)
					}
				}
			}
			fmt.Println(addr)
		}
	}
}

func main() {
	policy = os.Args[1] // Set policy from command-line argument

	listener, err := net1.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	initEtcd()
	go printAliveServers()
	grpcServer := grpc.NewServer()
	loadbalance.RegisterLoadBalancerServer(grpcServer, &loadbalanceserver{})

	log.Printf("Server is running on port: 8080")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
