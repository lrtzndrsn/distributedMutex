package main

import (
	"bufio"
	"context"
	distributed_mutex "distributed_mutex/grpc"
	pb "distributed_mutex/grpc"
	"os"

	"flag"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	pb.UnimplementedResourceAccessServer
	address               string
	port                  string
	initial_port          string
	client_array_size_int int
	coordinator           string
	IsCoordinator         bool
	IsResourceAvailable   bool
}

var client_array_size = flag.String("clientArraySize", "", "Max size of the client array")

func main() {

	flag.Parse()
	client_array_size_int, _ := strconv.Atoi(*client_array_size)
	client_IP := GetOutboundIP()
	initial_port := 5000
	port := FindAvailablePort(GetOutboundIP(), client_array_size_int, initial_port)

	client := &Client{
		address:               client_IP,
		port:                  port,
		initial_port:          strconv.Itoa(initial_port),
		client_array_size_int: client_array_size_int,
		IsCoordinator:         false,
		IsResourceAvailable:   false,
	}

	grpcServer := grpc.NewServer()
	listener, _ := net.Listen("tcp", client.address+":"+client.port)
	pb.RegisterResourceAccessServer(grpcServer, client)

	// Starting listening
	log.Print("Listening at: " + client.address + ":" + client.port)
	go grpcServer.Serve(listener)

	// Starting acting
	go RunProgram(client)
	time.Sleep(1 * time.Minute)

}

func RunProgram(client *Client) {
	log.Println("Calling election, coordinator before election: " + client.coordinator)
	client.CallElection(context.Background(), &pb.CallElectionMessage{})
	log.Println("Coordinator after election: " + client.coordinator)
	go UserResourceAccess(client)
}

// The user can give some input, and if access is granted by coordinator, print said input
func UserResourceAccess(client *Client) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		server_address := client.address + ":" + client.coordinator
		connection, _ := grpc.Dial(server_address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		grpc_client := pb.NewResourceAccessClient(connection)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		response, err := grpc_client.RequestResourceAccess(ctx, &pb.ResourceRequestMessage{
			Port: client.port,
		})

		if err != nil {
			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Unavailable { // If the gRPC server is unavailable, call a new election (leader might be dead)
				client.CallElection(context.Background(), &pb.CallElectionMessage{})
				continue
			} else {
				log.Fatalf("Error: %v", err)
			}
		}

		// If granted (Coordinator responds with true), the client will be allowed to print the user input (Just to simulate something)
		if response.IsRequestGranted {
			log.Printf("Access to print input granted. Input: %v", scanner.Text())
		} else {
			log.Print("Awww man, aint no access to resource for me")
		}
	}
}

func (client *Client) CallElection(context context.Context, call_election_message *pb.CallElectionMessage) (*pb.CallElectionResponseMessage, error) {
	log.Print("Election called!")

	address := client.address
	client_port, _ := strconv.Atoi(client.port)
	initial_port, _ := strconv.Atoi(client.initial_port)
	max_port := initial_port + client.client_array_size_int
	response_received := false

	if max_port == client_port {
		log.Print("I'm the coordinator, cause I'm the the biggest guy/girl in here!!!!")
		for i := client_port; i > initial_port; i-- {
			port := strconv.Itoa(initial_port + i)
			SendCoordinator(client, address, port)
		}
	}

	for i := client_port + 1; i <= max_port; i++ {
		port := strconv.Itoa(i)
		if SendElection(address, port) {
			response_received = true
		}
	}

	if !response_received {
		log.Print("No response means I'm the leader muhahaha")
		for i := client_port; i > initial_port; i-- {
			port := strconv.Itoa(i)
			SendCoordinator(client, address, port)
		}
	}

	return &pb.CallElectionResponseMessage{}, nil
}

func (client *Client) AssertCoordinator(context context.Context, message *distributed_mutex.AssertCoordinatorMessage) (*distributed_mutex.AssertCoordinatorResponseMessage, error) {
	client.coordinator = message.Port
	if message.Port == client.port {
		client.IsCoordinator = true
		go SetResourceAvailable(client)
	} else {
		client.IsCoordinator = false
	}
	log.Print("Someone thinks they are coordinator, this guy eh: " + message.Port)
	return &pb.AssertCoordinatorResponseMessage{
		Port: client.port,
	}, nil
}

func SetResourceAvailable(client *Client) {
	time.Sleep(5 * time.Second)
	client.IsResourceAvailable = true
}

func (client *Client) RequestResourceAccess(context context.Context, resource_request_message *pb.ResourceRequestMessage) (*pb.ResourceRequestResponse, error) {
	if client.IsCoordinator {
		if client.IsResourceAvailable {
			go GrantAccessToResource(client)
			return &pb.ResourceRequestResponse{
				IsRequestGranted: true,
			}, nil
		} else {
			return &pb.ResourceRequestResponse{
				IsRequestGranted: false,
			}, nil
		}
	} else {
		log.Fatal("I'm not coordinator you fool")
	}
	return &pb.ResourceRequestResponse{
		IsRequestGranted: false,
	}, nil
}

func GrantAccessToResource(client *Client) {
	client.IsResourceAvailable = false
	time.Sleep(5 * time.Second)
	client.IsResourceAvailable = true
}

func SendCoordinator(client *Client, address string, port string) {
	client.IsCoordinator = true
	log.Printf("Telling my subjects I'm the boss around here, subject: " + port)
	server_address := address + ":" + port
	connection, _ := grpc.Dial(server_address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	grpc_client := pb.NewResourceAccessClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	grpc_client.AssertCoordinator(ctx, &pb.AssertCoordinatorMessage{
		Port: client.port,
	})
}

func SendElection(address string, port string) bool {
	server_address := address + ":" + port
	connection, _ := grpc.Dial(server_address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := pb.NewResourceAccessClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := client.CallElection(ctx, &pb.CallElectionMessage{})
	if err == nil {
		return true
	}
	return false
}

func FindAvailablePort(address string, system_size int, initial_port int) string {
	for i := 0; i < system_size; i++ {
		port := strconv.Itoa(initial_port + i)
		timeout := time.Duration(1 * time.Second)
		_, err := net.DialTimeout("tcp", address+":"+port, timeout)
		if err != nil {
			log.Printf("Hey I'm at port: %v", port)
			return port
		}
	}
	log.Fatalf("No space left")
	return "Dosn't happen"
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
