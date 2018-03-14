## Setup


go version go1.9.2


### CLI application


wizefs_cli application is located here:
`$GOPATH/src/bitbucket.org/udt/wizefs/cmd/wizefs_cli`

You should go to this directory and run `go build`.

Also you can run from root directory `$GOPATH/src/bitbucket.org/udt/wizefs` this command:

`go build -o ./cmd/wizefs_cli/wizefs_cli -v ./cmd/wizefs_cli`

or you can build it right in the root directory:

`go build -v ./cmd/wizefs_cli`


### gRPC Server and Client


gRPC applications are located at the **grpc** directory:
`$GOPATH/src/bitbucket.org/udt/wizefs/grpc`

You should build **wizefs_mount application** before building gRPC Server and gRPC Client.

wizefs_mount application is located here:
`$GOPATH/src/bitbucket.org/udt/wizefs/cmd/wizefs_mount`

You should go to this directory and run `go build`.

Also you can run from root directory `$GOPATH/src/bitbucket.org/udt/wizefs` this command:

`go build -o ./cmd/wizefs_mount/wizefs_mount -v ./cmd/wizefs_mount`

Then you should build 2 commands independently by going to the appropriate folder in advance: `grpc/server` and `grpc/client`.

### REST Service

REST Service is located at the **rest** directory:
`$GOPATH/src/bitbucket.org/udt/wizefs/rest`

You should go to this directory and run `go build`.

Also you should build **wizefs_mount application** before REST Service. See more details in topic **gRPC Server and Client**.

### WizeFS Docker node (with REST Service running inside)

You can start WizeFS Docker node with `./start.sh`

### GUI application

See [GUI README](cmd/wizefs_ui/README.md)


## Flags


`--debug`

Enable debug output. Optional.

`--fg, -f`

Stay in the foreground. Optional.

`--notifypid`

Send USR1 to the specified process after successful mount. 
It used internally for daemonization.


## Common Info


`ORIGIN`

ORIGIN is single name of bucket. It is not directory with full path, just name or label for one of the many WizeFS buckets.

`Bucket`

Bucket for WizeFS is a synonym of FUSE-based filesystem. Currently there are 3 basic types of filesystems (and buckets):

1.  Loopback Filesystem (or simply LoopbackFS)
2.  Zipped Filesystem (or simply ZipFS) (read-only, in-memory)
3.  Loopback Zipped Filesystem (or simply LZFS).


## API (Command-line interface)


`create ORIGIN`

Create a new bucket. 
Now this command only checks if ORIGIN directory exists and create it if it is not exist. Also this command create config file for this bucket (wizefs.conf) and add this bucket to `created` map of common config (wizedb.conf).

### create Issues

* Check if bucket is (isn't) mounted. Perhaps should add flag for auto-mounting after creating

`delete ORIGIN`

Delete an existing bucket.
Now this command only checks if ORIGIN directory exists and delete it in this case with config file. Also this command delete bucket from `created` map of common config.

### delete Issues

* Check if bucket is mounted. Perhaps should add flag for auto-unmounting before deleting

`mount ORIGIN`

Mount an existing ORIGIN (directory or zip file) into MOUNTPOINT (this directory now is creating by application in the WizeFS root directory).
Also this command add bucket (with all needed data) to `mounted` map of common config (wizedb.conf).

`unmount ORIGIN`

Unmount an existing ORIGIN (application can search MOUNTPOINT by ORIGIN).
Also this command delete bucket from `mounted` map of common config.

`put FILE ORIGIN`

Upload FILE (you can use full path to the file here) to existing and mounted bucket with name (label) ORIGIN. Now it work only with directory-based bucket, but also you can experiment with LZFS bucket (zipped directory, with ORIGIN like archive.zip, currently only zip archive supported).

`get FILE ORIGIN`

Download FILE (you should use only filename now) from existing and mounted bucket with name (label) ORIGIN to the current directory. Now it work only with directory-based bucket, but also you can experiment with LZFS bucket (zipped directory, with ORIGIN like archive.zip, currently only zip archive supported).

`remove FILE ORIGIN`

Remove FILE (you should use only filename) from existing and mounted bucket with name (label) ORIGIN. Now it work only with directory-based bucket, but also you can experiment with LZFS bucket (zipped directory, with ORIGIN like archive.zip, currently only zip archive supported).


### API Commands Issues

* Add some other Filesystems API, like `check`, `list`
* Add Files API:  `search`
* Add Internal API: `verify`, `integrity`


## gRPC API methods


gRPC methods are identical to CLI commands, just first symbol is capitalized.

You can see gRPC Client code with all 6 commands working step by step.

```go
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

	time.Sleep(3 * time.Second)

	// Mount
	tlog.Info.Printf("Request: Mount. Origin: %s", origin)
	resp, err = client.Mount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)

	time.Sleep(3 * time.Second)

	// Put
	// TODO: HACK - just for local testing
	filepath := globals.OriginDirPath + "TESTDIR1/test.txt"
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

	time.Sleep(3 * time.Second)

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

	time.Sleep(3 * time.Second)

	// Unmount
	tlog.Info.Printf("Request: Unmount. Origin: %s", origin)
	resp, err = client.Unmount(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)

	time.Sleep(3 * time.Second)

	// Delete
	tlog.Info.Printf("Request: Delete. Origin: %s", origin)
	resp, err = client.Delete(context.Background(), &pb.FilesystemRequest{Origin: origin})
	tlog.Info.Printf("Response: %v. Error: %v", resp, err)
}
```

Other examples are located in the client_*_test.go test files.


### Create, Delete, Mount and Unmount methods


All methods with filesystem send with simple FilesystemRequest struct with only Origin value and receive simple FilesystemResponse struct with Executed boolean value and Message value.

```go
type FilesystemRequest struct {
	Origin string `protobuf:"bytes,1,opt,name=origin" json:"origin,omitempty"`
}

type FilesystemResponse struct {
	Executed bool   `protobuf:"varint,1,opt,name=executed" json:"executed,omitempty"`
	Message  string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}
```


### Put method


Put method sends PutRequest struct with Filename, Origin values and file Content as byte slice and receives PutResponse struct with Executed boolean value and Message value.

```go
type PutRequest struct {
	Filename string `protobuf:"bytes,1,opt,name=filename" json:"filename,omitempty"`
	Content  []byte `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	Origin   string `protobuf:"bytes,3,opt,name=origin" json:"origin,omitempty"`
}

type PutResponse struct {
	Executed bool   `protobuf:"varint,1,opt,name=executed" json:"executed,omitempty"`
	Message  string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}
```


### Get method


Get method sends GetRequest struct with Filename and Origin values and receives GetResponse struct with Executed boolean value, Message value and file Content as byte slice.

```go
type GetRequest struct {
	Filename string `protobuf:"bytes,1,opt,name=filename" json:"filename,omitempty"`
	Origin   string `protobuf:"bytes,2,opt,name=origin" json:"origin,omitempty"`
}

type GetResponse struct {
	Executed bool   `protobuf:"varint,1,opt,name=executed" json:"executed,omitempty"`
	Message  string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	Content  []byte `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}
```



## REST API



REST Service is listen on port 13000: `localhost:13000`

### Create bucket ORIGIN

```
curl -X POST localhost:13000/buckets -d '{"data":{"origin":"ORIGIN"}}'
```

### Delete bucket ORIGIN

```
curl -X DELETE localhost:13000/buckets/ORIGIN
```

### Mount bucket ORIGIN

```
curl -X POST localhost:13000/buckets/ORIGIN/mount
```

### Unmount bucket ORIGIN

```
curl -X POST localhost:13000/buckets/ORIGIN/unmount
```

### Put file /PATH/FILE to bucket ORIGIN

```
curl -F "filename=@/PATH/FILE" -X POST localhost:13000/buckets/ORIGIN/putfile
```

### Get file FILE from bucket ORIGIN

```
curl -X GET localhost:13000/buckets/ORIGIN/files/FILE --output /PATH/FILE
```

### Remove file FILE from bucket ORIGIN

```
curl -X DELETE localhost:13000/buckets/ORIGIN/files/FILE
```


## Next Issues

* Write Bash tests, Unit tests
* Write stress tests
* Develop third filesystem type (3) to combine ZipFS and LoopbackFS ideas
* Develop filesystem design for future versions
* etc