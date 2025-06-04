package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mapreduce "Q2/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type worker struct {
	mapreduce.UnimplementedWorkerServiceServer
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

var mu1 sync.Mutex
var numReducers int
var numMappers int

func sendRegisterRequest(client mapreduce.MasterServiceClient, port string) {
	ip := getMyIP()
	fmt.Printf("Sending Register Request %s\n", port)
	if ip == "" {
		log.Fatal("Could not determine IP address")
	}

	fmt.Println("Register request to Master with ip = ", ip, " and port = ", port)
	req := &mapreduce.RegisterRequest{Address: ip, Port: port}
	res, err := client.RegisterWorker(context.Background(), req)
	if err != nil {
		log.Fatalf("Error registering worker: %v", err)
	}
	log.Printf("Response from RegisterServer port = %s: %v", port, res)
}

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

func (s *worker) InvertedIndex(ctx context.Context, req *mapreduce.Task) (*mapreduce.TaskResponse, error) { // for Inverted Index Task
	var task int = int(req.GetTask())
	var filename string = req.GetFilename()
	var index int = int(req.GetIndex())
	fmt.Printf("Inverted Index filename = %s, task = %d, index = %d\n", filename, task, index)
	if task == 1 { // mapper
		file, err := os.Open("./dataset/file" + strconv.Itoa(index) + ".txt")
		var result []string
		if err != nil {
			log.Fatal(err)
		}
		Scanner := bufio.NewScanner(file)
		Scanner.Split(bufio.ScanWords)

		for Scanner.Scan() {
			result = append(result, Scanner.Text())
		}

		mp := make(map[string]string)

		var files []*os.File
		for j := 0; j < numReducers; j++ {
			file, err := os.OpenFile("./inverted_index_"+strconv.Itoa(j)+"_"+strconv.Itoa(index-1)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			files = append(files, file)
		}

		for j := 0; j < len(result); j++ {
			mp[result[j]] = filename
			h := sha256.Sum256([]byte(result[j]))
			hashInt := int(binary.BigEndian.Uint32(h[:4])) % numReducers
			_, err := files[hashInt].WriteString(result[j] + ":" + filename + "\n")
			if err != nil {
				panic(err)
			}
		}
	} else if task == 2 { // reducer
		type inter struct {
			files map[string]int
		}
		mp := make(map[string]inter)

		for j := 0; j < numMappers; j++ {
			file, err := os.Open("./inverted_index_" + strconv.Itoa(index-1) + "_" + strconv.Itoa(j) + ".txt")
			if err != nil {
				log.Fatal(err)
			}
			r := bufio.NewReader(file)
			s, err := Readln(r)
			for err == nil {
				res := strings.SplitN(s, ":", -1)
				val, exists := mp[res[0]]
				if !exists {
					val = inter{}
				}
				val1, ok := val.files[res[1]]
				if !ok {
					val1 = 1
					if val.files == nil {
						val.files = make(map[string]int)
					}
				}
				val.files[res[1]] = val1
				mp[res[0]] = val
				s, err = Readln(r)
			}
		}

		mu1.Lock()
		file, err := os.OpenFile("inverted_index_task.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		var str_file string = ""
		for key, value := range mp {
			str_file = str_file + key + ":"
			for key1, _ := range value.files {
				str_file = str_file + key1 + "\t"
			}
			str_file = str_file + "\n"
		}
		length, err := file.WriteString(str_file)
		if err != nil {
			panic(err)
		}
		mu1.Unlock()
		fmt.Println(length)
	}
	return &mapreduce.TaskResponse{Success: true}, nil
}

func (s *worker) WordCount(ctx context.Context, req *mapreduce.Task) (*mapreduce.TaskResponse, error) { // for Word Count Task
	var task int = int(req.GetTask())
	var filename string = req.GetFilename()
	var index int = int(req.GetIndex())
	fmt.Printf("Word Counting filename = %s, task = %d, index = %d\n", filename, task, index)
	if task == 1 { // mapper
		file, err := os.Open("./dataset/file" + strconv.Itoa(index) + ".txt")
		var result []string
		if err != nil {
			log.Fatal(err)
		}
		Scanner := bufio.NewScanner(file)
		Scanner.Split(bufio.ScanWords)

		for Scanner.Scan() {
			result = append(result, Scanner.Text())
		}

		var files []*os.File
		for j := 0; j < numReducers; j++ {
			file, err := os.OpenFile("reduce_wordcnt_"+strconv.Itoa(j)+"_"+strconv.Itoa(index-1)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			files = append(files, file)
		}

		for j := 0; j < len(result); j++ {
			h := sha256.Sum256([]byte(result[j]))
			hashInt := int(binary.BigEndian.Uint32(h[:4])) % numReducers
			_, err := files[hashInt].WriteString(result[j] + ":1" + "\n")
			if err != nil {
				panic(err)
			}
		}

	} else if task == 2 { // reducer
		mp := make(map[string]int)
		for j := 0; j < numMappers; j++ {
			file, err := os.Open("./reduce_wordcnt_" + strconv.Itoa(index-1) + "_" + strconv.Itoa(j) + ".txt")
			if err != nil {
				log.Fatal(err)
			}
			r := bufio.NewReader(file)
			s, err := Readln(r)
			for err == nil {
				parts := strings.Split(s, ":")
				word := strings.Join(parts[:len(parts)-1], "")
				count := parts[len(parts)-1]
				if word == "" {
					s, err = Readln(r)
					continue
				}
				val, exists := mp[word]
				if !exists {
					mp[word] = 0
				}
				freq, err := strconv.Atoi(count)
				if err != nil {
					fmt.Println(word, count)
					panic(err)
				}
				mp[word] = val + freq
				s, err = Readln(r)
			}
		}

		mu1.Lock()
		file, err := os.OpenFile("word_count_task.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		var str_file string = ""
		for key, value := range mp {
			str_value := strconv.Itoa(value)
			str_file = str_file + key + ":" + str_value + "\n"
		}
		length, err := file.WriteString(str_file)
		if err != nil {
			panic(err)
		}
		fmt.Println(length)
		mu1.Unlock()

	} else if task == 3 {
		fmt.Println("Task done, Exiting ...")
		go finish()
		return &mapreduce.TaskResponse{Success: true}, nil
	}
	return &mapreduce.TaskResponse{Success: true}, nil
}

func finish() {
	time.Sleep(1 * time.Second)
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

	numReducers = reducers
	numMappers = mappers
	numWorkers := max(mappers, reducers)
	fmt.Println(numWorkers)

	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(i int) {
			defer wg.Done() // Signal that this goroutine is finished
			updatePortFile(&mu, i)
		}(i)
	}
	wg.Wait() // Wait for all goroutines to finish before exiting
}

// Function to update the port file safely and host gRPC Servers
func updatePortFile(mu *sync.Mutex, i int) {
	fmt.Println("Inside update port file function i = ", i)

	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := mapreduce.NewMasterServiceClient(conn)

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
	fmt.Println("Updated port file port = ", string(data))
	mu.Unlock()

	sendRegisterRequest(client, string(data))

	listener, err := net.Listen("tcp", ":"+string(data))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	mapreduce.RegisterWorkerServiceServer(grpcServer, &worker{})
	fmt.Println("Worker Server is running on port: ", string(data))

	// Serve requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
