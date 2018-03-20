package main

import (
	"net"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "bitbucket.org/udt/wizefs/grpc/wizefsservice"
)

var (
	serverAddrTest = "127.0.0.1:10000"
)

func startServer(t *testing.T) {
	lis, err := net.Listen("tcp", serverAddrTest)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWizeFsServiceServer(grpcServer, pb.NewServer())
	grpcServer.Serve(lis)
}

func getConnection(t *testing.T) *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(serverAddrTest, opts...)
	if err != nil {
		t.Fatalf("Fail to dial: %v", err)
	}
	return conn
}

func testCreateInvalidOrigin(t *testing.T) {
	// TODO: HACK - start server for testing
	t.Logf("Starting server on %s", serverAddrTest)
	go startServer(t)
	time.Sleep(1 * time.Second)

	// start testing
	conn := getConnection(t)
	defer conn.Close()
	client := pb.NewWizeFsServiceClient(conn)

	origin := "image.jpg"

	// Create
	t.Logf("Request: Create. Origin: %s", origin)
	resp, err := client.Create(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Create method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Create method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)
}

func TestFullCircle(t *testing.T) {
	// TODO: HACK - start server for testing
	t.Logf("Starting server on %s", serverAddrTest)
	go startServer(t)
	time.Sleep(1 * time.Second)

	// start testing
	conn := getConnection(t)
	defer conn.Close()
	client := pb.NewWizeFsServiceClient(conn)

	origin := "GRPCTest"

	// Create
	t.Logf("Request: Create. Origin: %s", origin)
	resp, err := client.Create(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Create method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Create method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)
	time.Sleep(1 * time.Second)

	// TODO: check if origin dir was created

	//origin = "GRPCFail"

	// Mount
	t.Logf("Request: Mount. Origin: %s", origin)
	resp, err = client.Mount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Mount method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Mount method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)
	time.Sleep(1 * time.Second)

	// TODO: check if origin dir was mounted

	// TODO: add Put/Get tests
	// Put
	// TODO: HACK - just for local testing
	filepath := "test.txt"
	content, err := readFile(filepath)
	if err == nil {
		t.Logf("Request content: \n%s\n", content)

		t.Logf("Request: Put. Origin: %s", origin)
		respPut, err := client.Put(context.Background(),
			&pb.PutRequest{
				Filename: "test.txt",
				Content:  content,
				Origin:   origin,
			})
		if err != nil {
			t.Fatalf("Fail to execute Put method: %v", err)
		}
		if !respPut.Executed {
			t.Fatalf("Bad response from Put method: %s", respPut.Message)
		}
		t.Logf("Response message: %s.", respPut.Message)
	} else {
		t.Fatalf("We have problem with read file: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Get
	t.Logf("Request: Get. Origin: %s", origin)
	respGet, err := client.Get(context.Background(),
		&pb.GetRequest{
			Filename: "test.txt",
			Origin:   origin,
		})
	if err != nil {
		t.Fatalf("Fail to execute Get method: %v", err)
	}
	if !respGet.Executed {
		t.Fatalf("Bad response from Get method: %s", respGet.Message)
	}
	t.Logf("Response message: %s.", respGet.Message)
	t.Logf("Response content: \n%s\n", respGet.Content)

	time.Sleep(500 * time.Millisecond)

	// Unmount
	t.Logf("Request: Unmount. Origin: %s", origin)
	resp, err = client.Unmount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Unmount method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Unmount method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)

	time.Sleep(500 * time.Millisecond)

	// TODO: check if origin dir was unmounted

	// Delete
	t.Logf("Request: Delete. Origin: %s", origin)
	resp, err = client.Delete(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Delete method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Delete method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)

	time.Sleep(500 * time.Millisecond)

	// TODO: check if origin dir was deleted
}

func testCreate(t *testing.T, client pb.WizeFsServiceClient, origin string) {
	// Create
	t.Logf("Request: Create. Origin: %s", origin)
	resp, err := client.Create(context.Background(), &pb.FilesystemRequest{Origin: origin})
	if err != nil {
		t.Fatalf("Fail to execute Create method: %v", err)
	}
	if !resp.Executed {
		t.Fatalf("Bad response from Create method: %s", resp.Message)
	}
	t.Logf("Response message: %s.", resp.Message)
}

//func TestClient(t *testing.T) {
//	// TODO: HACK - start server for testing
//	t.Logf("Starting server on %s", serverAddrTest)
//	go startServer(t)
//	time.Sleep(1 * time.Second)

//	// start testing
//	conn := getConnection(t)
//	defer conn.Close()
//	client := pb.NewWizeFsServiceClient(conn)

//	origin := "image.tar"

//	testCreate(t, client, origin)
//}
