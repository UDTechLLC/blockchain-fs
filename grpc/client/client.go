package main

import (
	"flag"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "bitbucket.org/udt/wizefs/grpc/wizefsservice"
	"bitbucket.org/udt/wizefs/internal/tlog"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

func readFile(filename string) (content []byte, err error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		tlog.Fatal.Println(err)
		return nil, err
	}

	content, err = ioutil.ReadAll(file)
	if err != nil {
		tlog.Fatal.Println(err)
		return nil, err
	}

	return content, nil
}

func main() {
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		tlog.Fatal.Printf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewWizeFsServiceClient(conn)

	origin := "GRPC1"

	// Create
	tlog.Info.Printf("Request: Create. Origin: %s", origin)
	resp, err := client.Create(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)

	time.Sleep(1 * time.Second)

	// Mount
	tlog.Info.Printf("Request: Mount. Origin: %s", origin)
	resp, err = client.Mount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)

	time.Sleep(1 * time.Second)

	// Put
	// TODO: HACK - just for local testing
	filepath := "test.txt"
	content, err := readFile(filepath)
	if err == nil {
		tlog.Info.Printf("Request content: \n%s\n", content)

		tlog.Info.Printf("Request: Put. Origin: %s", origin)
		respPut, err := client.Put(context.Background(),
			&pb.PutRequest{
				Filename: "test.txt",
				Content:  content,
				Origin:   origin,
			})
		tlog.Info.Printf("Response: %v. Error: %v", respPut, err)
	}

	time.Sleep(1 * time.Second)

	// Get
	if err == nil {
		tlog.Info.Printf("Request: Get. Origin: %s", origin)
		respGet, err := client.Get(context.Background(),
			&pb.GetRequest{
				Filename: "test.txt",
				Origin:   origin,
			})
		tlog.Info.Printf("Error: %v", err)
		tlog.Info.Printf("Response message: %s.", respGet.Message)
		tlog.Info.Printf("Response contentb: \n%s\n", respGet.Content)
	}

	time.Sleep(1 * time.Second)

	// Remove
	if err == nil {
		tlog.Info.Printf("Request: Remove. Origin: %s", origin)
		respRemove, err := client.Remove(context.Background(),
			&pb.RemoveRequest{
				Filename: "test.txt",
				Origin:   origin,
			})
		tlog.Info.Printf("Error: %v", err)
		tlog.Info.Printf("Response message: %s.", respRemove.Message)
	}

	time.Sleep(1 * time.Second)

	// Unmount
	tlog.Info.Printf("Request: Unmount. Origin: %s", origin)
	resp, err = client.Unmount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)

	time.Sleep(1 * time.Second)

	// Delete
	tlog.Info.Printf("Request: Delete. Origin: %s", origin)
	resp, err = client.Delete(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)
}
